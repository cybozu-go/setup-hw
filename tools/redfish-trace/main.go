// redfish-trace measures the BMC load of monitor-hw's Redfish traversal on a real BMC.
//
// It reuses the collection rules embedded in setup-hw (redfish.Rules), runs the traversal in
// several modes ("current" = exactly what monitor-hw does today, "improved" = proposal 1
// (extra excludes) + proposal 2 ($expand on collections)), records every HTTP request as an
// OpenTelemetry span, and renders the spans as an SVG waterfall. It also verifies that all
// modes yield the same set of Prometheus metrics.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cybozu-go/log"
	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/cybozu-go/setup-hw/redfish"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"sigs.k8s.io/yaml"
)

type stringList []string

func (s *stringList) String() string     { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error { *s = append(*s, v); return nil }

func main() {
	var (
		host        = flag.String("host", "", "BMC address (host or host:port). Overrides -address-file")
		addressFile = flag.String("address-file", "", "setup-hw bmc-address.json to read the BMC address from (e.g. /etc/neco/bmc-address.json)")
		userFile    = flag.String("user-file", "", "setup-hw bmc-user.json to read the 'support' password from (e.g. /etc/neco/bmc-user.json)")
		user        = flag.String("user", "support", "BMC user (monitor-hw uses 'support')")
		password    = flag.String("password", os.Getenv("REDFISH_PASSWORD"), "BMC password (or env REDFISH_PASSWORD)")
		ruleName    = flag.String("rule", "", "embedded rule name, e.g. dell_redfish_1.20.1.yml (default: auto from RedfishVersion)")
		ruleFile    = flag.String("rule-file", "", "load the collection rule from this YAML file instead")
		modes       = flag.String("modes", "current,expand", "comma-separated modes to run in order. Each is name[:opt+opt...][@interval]; names: current, exclude, expand, improved (=expand:oem); options: levels=N, all, dot, oem, nextlink. e.g. current@1m,expand:levels=2+all@5m")
		noExpand    = flag.String("no-expand", `^/redfish/v1/?$`, "regexp of paths never fetched with $expand in 'all' mode (the service root inlines JsonSchemas/Registries)")
		cycles      = flag.Int("cycles", 2, "traversal cycles per mode (cycle 2+ shows session/connection reuse and learned collections)")
		interval    = flag.Duration("interval", 5*time.Second, "pause between cycles (monitor-hw default is 60s)")
		timeout     = flag.Duration("timeout", 5*time.Second, "per-request timeout (monitor-hw uses 5s)")
		svgOut      = flag.String("svg", "redfish-trace.svg", "output SVG path")
		jsonOut     = flag.String("spans-json", "", "also write raw spans as JSON (stdouttrace format)")
		noWaterfall = flag.Bool("no-waterfall", false, "omit the per-request waterfall (strip + summary only)")
		inputFile   = flag.String("input-file", "", "offline mode: serve responses from this Redfish dump (JSON object keyed by path)")
		fakeLatency = flag.Duration("fake-latency", 20*time.Millisecond, "offline mode: latency per request")
		dumpDir     = flag.String("dump-dir", "", "write collected data per mode as <dir>/<mode>.json (usable as -input-file later)")
		verbose     = flag.Bool("v", false, "log each request")
		nextLink    = flag.Bool("follow-nextlink", false, "expand modes: also fetch paginated pages (Members@odata.nextLink); monitor-hw does not")
		noLogout    = flag.Bool("no-logout", false, "keep the Redfish session at exit (monitor-hw never deletes it)")
		extraExcl   stringList
	)
	flag.Var(&extraExcl, "extra-exclude", "extra exclude regexp for 'exclude'/'improved' modes (repeatable; default /Oem/)")
	flag.Parse()
	if len(extraExcl) == 0 {
		extraExcl = stringList{"/Oem/"}
	}

	if err := run(*host, *addressFile, *userFile, *user, *password, *ruleName, *ruleFile, *modes, *cycles, *interval, *timeout,
		*svgOut, *jsonOut, !*noWaterfall, *inputFile, *fakeLatency, *dumpDir, *verbose, !*noLogout, *nextLink, *noExpand, extraExcl); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(host, addressFile, userFile, user, password, ruleName, ruleFile, modeList string, cycles int, interval, timeout time.Duration,
	svgOut, jsonOut string, waterfall bool, inputFile string, fakeLatency time.Duration, dumpDir string, verbose, logout, nextLink bool, noExpand string, extraExcl []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// ---- endpoint & credentials ----
	offline := inputFile != ""
	if host == "" && addressFile != "" {
		var ac struct {
			IPv4 struct{ Address string } `json:"ipv4"`
		}
		if err := readJSON(addressFile, &ac); err != nil {
			return err
		}
		host = ac.IPv4.Address
	}
	if password == "" && userFile != "" {
		var uc struct {
			Support struct {
				Password struct{ Raw string } `json:"password"`
			} `json:"support"`
		}
		if err := readJSON(userFile, &uc); err != nil {
			return err
		}
		password = uc.Support.Password.Raw
	}
	if offline {
		if host == "" {
			host = "offline.invalid"
		}
		if password == "" {
			password = "offline"
		}
	}
	if host == "" {
		return errors.New("-host (or -address-file / -input-file) is required")
	}
	if password == "" {
		return errors.New("-password, REDFISH_PASSWORD or -user-file is required")
	}
	endpoint, err := url.Parse("https://" + host)
	if err != nil {
		return err
	}

	newRT := func() http.RoundTripper { return newTransport() }
	if offline {
		ot, err := newOfflineTransport(inputFile, fakeLatency)
		if err != nil {
			return err
		}
		newRT = func() http.RoundTripper { return ot }
	}

	tracer, mem, shutdown, err := setupTracing(jsonOut)
	if err != nil {
		return err
	}

	// ---- rule ----
	rule, ruleLabel, err := selectRule(ctx, endpoint, ruleName, ruleFile, newRT(), timeout, offline)
	if err != nil {
		return err
	}
	fmt.Printf("rule: %s\n", ruleLabel)

	// ---- modes ----
	var modeDefs []Mode
	for _, spec := range strings.Split(modeList, ",") {
		if strings.TrimSpace(spec) == "" {
			continue
		}
		md, err := parseMode(spec, extraExcl, nextLink)
		if err != nil {
			return err
		}
		modeDefs = append(modeDefs, md)
	}
	if len(modeDefs) == 0 {
		return errors.New("no modes given")
	}

	results := map[string]cycleResult{}
	for mi, mode := range modeDefs {
		tr, err := newTraverser(endpoint, user, password, newRT(), timeout, tracer, mode, rule, noExpand, interval, verbose)
		if err != nil {
			return err
		}
		desc := fmt.Sprintf("interval=%s", tr.interval)
		if len(mode.ExtraExcludes) > 0 {
			desc += fmt.Sprintf(" extra-excludes=%v", mode.ExtraExcludes)
		}
		if mode.Expand {
			scope := "collections"
			if mode.ExpandAll {
				scope = "all"
			}
			desc += fmt.Sprintf(" %s on %s", mode.expandQuery(), scope)
		}
		fmt.Printf("== mode %s (%s)\n", mode.Name, desc)
		for c := 1; c <= cycles; c++ {
			res := tr.runCycle(ctx, c)
			if res.Err != nil {
				fmt.Printf("  cycle %d: FAILED: %v\n", c, res.Err)
			} else {
				fmt.Printf("  cycle %d: %d resources in %s\n", c, len(res.Data), fmtDur(res.Duration))
				results[mode.Name] = res
			}
			if ctx.Err() != nil {
				break
			}
			if c < cycles {
				select {
				case <-time.After(tr.interval):
				case <-ctx.Done():
				}
			}
		}
		if logout {
			if err := tr.logout(context.Background()); err != nil {
				fmt.Printf("  logout: %v\n", err)
			}
		}
		if ctx.Err() != nil {
			break
		}
		if mi < len(modeDefs)-1 {
			select {
			case <-time.After(interval):
			case <-ctx.Done():
			}
		}
	}

	if err := shutdown(context.Background()); err != nil {
		return err
	}

	// ---- summary ----
	cyclesInfo := groupCycles(mem.sorted())
	fmt.Printf("\n%-22s %-6s %9s %9s %6s %6s %6s %8s %10s %6s %8s %9s %8s\n", "mode", "cycle", "duration", "requests", "ok", "err", "expand", "inlined", "bytes", "tls", "reused", "interval", "req/h")
	for _, c := range cyclesInfo {
		st := c.stats()
		fmt.Printf("%-22s %-6d %9s %9d %6d %6d %6d %8d %10s %6d %8d %9s %8.0f\n", c.mode, c.cycle, fmtDur(st.Duration), st.Requests, st.OK, st.Errors, st.Expand, st.Expanded, fmtBytes(st.Bytes), st.TLS, st.Reused, fmtDur(c.interval), st.ReqPerHour(c.interval))
	}
	fmt.Println("req/h = requests * 3600 / (duration + interval): steady-state BMC request rate of a monitor-hw loop with that interval")

	fmt.Printf("\n%-22s %-6s %-8s %6s %9s %9s %9s %9s\n", "mode", "cycle", "kind", "n", "p50", "p90", "p99", "max")
	for _, c := range cyclesInfo {
		st := c.stats()
		for _, row := range []struct {
			kind string
			l    latency
		}{{"all", st.Lat}, {"plain", st.LatPlain}, {"$expand", st.LatExpand}} {
			if row.l.N == 0 {
				continue
			}
			fmt.Printf("%-22s %-6d %-8s %6d %9s %9s %9s %9s\n", c.mode, c.cycle, row.kind, row.l.N, fmtDur(row.l.P50), fmtDur(row.l.P90), fmtDur(row.l.P99), fmtDur(row.l.Max))
		}
	}

	fmt.Println("\nerrors per cycle:")
	for _, c := range cyclesInfo {
		st := c.stats()
		fmt.Printf("  %-22s %-3d %s\n", c.mode, c.cycle, st.errSummary())
		for _, r := range st.ErrList {
			q := ""
			if r.expand {
				q = "?$expand"
			}
			desc := r.errKind()
			if r.status == 0 {
				desc = r.span.Status().Description
			}
			fmt.Printf("      +%-8s %9s %-10s %s%s  %s\n", fmtDur(r.start.Sub(c.start)), fmtDur(r.dur), r.kind, r.path, q, desc)
		}
	}

	fmt.Println("\nslowest requests per cycle:")
	for _, c := range cyclesInfo {
		st := c.stats()
		for _, r := range st.Slowest {
			q := ""
			if r.expand {
				q = "?$expand"
			}
			fmt.Printf("  %-22s %-3d %9s %4d %s%s\n", c.mode, c.cycle, fmtDur(r.dur), r.status, r.path, q)
		}
	}

	// ---- metric equivalence ----
	if len(results) > 1 {
		fmt.Println()
		compareMetrics(ctx, rule, modeDefs, results)
	}

	if dumpDir != "" {
		if err := os.MkdirAll(dumpDir, 0o755); err != nil {
			return err
		}
		for name, res := range results {
			out := map[string]interface{}{}
			for p, c := range res.Data {
				out[p] = c.Data()
			}
			raw, _ := json.MarshalIndent(out, "", "  ")
			if err := os.WriteFile(dumpDir+"/"+name+".json", raw, 0o644); err != nil {
				return err
			}
		}
	}

	title := fmt.Sprintf("redfish-trace %s rule=%s %s", host, ruleLabel, time.Now().Format(time.RFC3339))
	if err := os.WriteFile(svgOut, []byte(renderSVG(cyclesInfo, waterfall, title)), 0o644); err != nil {
		return err
	}
	fmt.Printf("\nwrote %s\n", svgOut)
	return nil
}

func readJSON(path string, v interface{}) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, v); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// selectRule picks the collection rule the same way monitor-hw does (dell_redfish_<RedfishVersion>.yml),
// falling back to the newest embedded rule not newer than the BMC's version.
func selectRule(ctx context.Context, endpoint *url.URL, ruleName, ruleFile string, rt http.RoundTripper, timeout time.Duration, offline bool) (*redfish.CollectRule, string, error) {
	if ruleFile != "" {
		raw, err := os.ReadFile(ruleFile)
		if err != nil {
			return nil, "", err
		}
		rule := new(redfish.CollectRule)
		if err := yaml.Unmarshal(raw, rule); err != nil {
			return nil, "", fmt.Errorf("%s: %w", ruleFile, err)
		}
		if err := rule.Validate(); err != nil {
			return nil, "", err
		}
		if err := rule.Compile(); err != nil {
			return nil, "", err
		}
		return rule, ruleFile, nil
	}
	if ruleName != "" {
		rule, ok := redfish.Rules[ruleName]
		if !ok {
			return nil, "", fmt.Errorf("unknown rule %q; available: %s", ruleName, strings.Join(ruleNames(), ", "))
		}
		return rule, ruleName, nil
	}

	// GET /redfish/v1/ is unauthenticated on iDRAC; ask for RedfishVersion.
	var version string
	if offline {
		version = offlineVersion(rt)
	} else {
		client := &http.Client{Transport: rt, Timeout: timeout}
		req, _ := http.NewRequestWithContext(ctx, "GET", endpoint.JoinPath("/redfish/v1/").String(), nil)
		req.Header.Set("Accept", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			return nil, "", fmt.Errorf("GET /redfish/v1/: %w", err)
		}
		defer resp.Body.Close()
		var root struct{ RedfishVersion string }
		if err := json.NewDecoder(resp.Body).Decode(&root); err != nil {
			return nil, "", fmt.Errorf("GET /redfish/v1/: %w", err)
		}
		version = root.RedfishVersion
	}
	exact := fmt.Sprintf("dell_redfish_%s.yml", version)
	if rule, ok := redfish.Rules[exact]; ok {
		return rule, exact, nil
	}
	best := ""
	for _, name := range ruleNames() {
		v := strings.TrimSuffix(strings.TrimPrefix(name, "dell_redfish_"), ".yml")
		if !strings.HasPrefix(name, "dell_redfish_") || compareVersions(v, version) > 0 {
			continue
		}
		if best == "" || compareVersions(v, strings.TrimSuffix(strings.TrimPrefix(best, "dell_redfish_"), ".yml")) > 0 {
			best = name
		}
	}
	if best == "" {
		return nil, "", fmt.Errorf("no rule for RedfishVersion %s (monitor-hw would fail too); use -rule", version)
	}
	fmt.Printf("warning: no exact rule for RedfishVersion %s, monitor-hw would fail; using %s\n", version, best)
	return redfish.Rules[best], best, nil
}

func offlineVersion(rt http.RoundTripper) string {
	ot, ok := rt.(*offlineTransport)
	if !ok {
		return ""
	}
	raw, ok := ot.lookup("/redfish/v1")
	if !ok {
		return ""
	}
	var root struct{ RedfishVersion string }
	json.Unmarshal(raw, &root)
	return root.RedfishVersion
}

func ruleNames() []string {
	var names []string
	for n := range redfish.Rules {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func compareVersions(a, b string) int {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		var x, y int
		if i < len(as) {
			x, _ = strconv.Atoi(as[i])
		}
		if i < len(bs) {
			y, _ = strconv.Atoi(bs[i])
		}
		if x != y {
			if x < y {
				return -1
			}
			return 1
		}
	}
	return 0
}

// ---- metric equivalence check via the real redfish.Collector ----

type staticClient struct{ data map[string]*gabs.Container }

func (s staticClient) Traverse(_ context.Context, rule *redfish.CollectRule) redfish.Collected {
	return *redfish.NewCollected(s.data, rule)
}
func (staticClient) GetVersion(context.Context) (string, error) { return "", nil }
func (staticClient) Login(context.Context) error                { return nil }

func metricKeys(ctx context.Context, rule *redfish.CollectRule, data map[string]*gabs.Container) ([]string, error) {
	log.DefaultLogger().SetThreshold(log.LvError) // silence Collector.Update's info logs
	col, err := redfish.NewCollector(func(context.Context) (*redfish.CollectRule, error) { return rule, nil }, staticClient{data})
	if err != nil {
		return nil, err
	}
	col.Update(ctx)
	ch := make(chan prometheus.Metric, 1024)
	go func() { col.Collect(ch); close(ch) }()
	var keys []string
	for m := range ch {
		desc := m.Desc().String()
		if strings.Contains(desc, "hw_last_update") {
			continue
		}
		var d dto.Metric
		if err := m.Write(&d); err != nil {
			return nil, err
		}
		var labels []string
		for _, lp := range d.Label {
			labels = append(labels, lp.GetName()+"="+lp.GetValue())
		}
		name := desc
		if i := strings.Index(desc, `fqName: "`); i >= 0 {
			name = desc[i+9:]
			name = name[:strings.Index(name, `"`)]
		}
		val := ""
		switch {
		case d.Gauge != nil:
			val = strconv.FormatFloat(d.Gauge.GetValue(), 'g', -1, 64)
		case d.Counter != nil:
			val = strconv.FormatFloat(d.Counter.GetValue(), 'g', -1, 64)
		}
		keys = append(keys, fmt.Sprintf("%s{%s} %s", name, strings.Join(labels, ","), val))
	}
	sort.Strings(keys)
	return keys, nil
}

func compareMetrics(ctx context.Context, rule *redfish.CollectRule, modes []Mode, results map[string]cycleResult) {
	base := modes[0].Name
	baseKeys, err := metricKeys(ctx, rule, results[base].Data)
	if err != nil {
		fmt.Println("metric comparison failed:", err)
		return
	}
	fmt.Printf("metrics: %s=%d", base, len(baseKeys))
	baseSet := toSet(baseKeys)
	for _, m := range modes[1:] {
		res, ok := results[m.Name]
		if !ok {
			continue
		}
		keys, err := metricKeys(ctx, rule, res.Data)
		if err != nil {
			fmt.Println("metric comparison failed:", err)
			return
		}
		set := toSet(keys)
		var missing, extra []string
		for k := range baseSet {
			if !set[k] {
				missing = append(missing, k)
			}
		}
		for k := range set {
			if !baseSet[k] {
				extra = append(extra, k)
			}
		}
		sort.Strings(missing)
		sort.Strings(extra)
		fmt.Printf(" %s=%d", m.Name, len(keys))
		if len(missing)+len(extra) > 0 {
			fmt.Printf("\n  %s vs %s: missing=%d extra=%d (value changes between cycles are expected for sensor readings)\n", m.Name, base, len(missing), len(extra))
			for _, k := range missing {
				fmt.Println("    -", k)
			}
			for _, k := range extra {
				fmt.Println("    +", k)
			}
		} else {
			fmt.Printf(" (identical to %s)", base)
		}
	}
	fmt.Println()
}

func toSet(keys []string) map[string]bool {
	s := make(map[string]bool, len(keys))
	for _, k := range keys {
		s[k] = true
	}
	return s
}
