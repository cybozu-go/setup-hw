package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cybozu-go/setup-hw/gabs"
	"github.com/cybozu-go/setup-hw/redfish"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Mode describes one traversal strategy to be measured.
type Mode struct {
	Name           string
	ExtraExcludes  []string      // additional exclude regexps on top of the rule
	Expand         bool          // use $expand
	ExpandAll      bool          // expand every GET (default: only collections)
	ExpandLevels   int           // $levels=N (default 1)
	ExpandDot      bool          // $expand=.(...) (subordinate resources only) instead of *(...) (also Links)
	FollowNextLink bool          // also fetch paginated collection pages (Members@odata.nextLink)
	Interval       time.Duration // pause between cycles; 0 = the global -interval
}

// expandQuery returns the $expand query string for this mode.
func (m Mode) expandQuery() string {
	levels := m.ExpandLevels
	if levels < 1 {
		levels = 1
	}
	kind := "*"
	if m.ExpandDot {
		kind = "."
	}
	return fmt.Sprintf("$expand=%s($levels=%d)", kind, levels)
}

// parseMode parses "name[:opt+opt...][@interval]" (options joined with "+" because "," separates modes).
//
// Names: current, exclude, expand, improved (= expand:oem).
// Options: levels=N, all, dot, oem, nextlink. Example: "expand:levels=2+all@5m".
func parseMode(spec string, extraExcl []string, followNextLink bool) (Mode, error) {
	spec = strings.TrimSpace(spec)
	var m Mode
	if i := strings.Index(spec, "@"); i >= 0 {
		d, err := time.ParseDuration(spec[i+1:])
		if err != nil {
			return m, fmt.Errorf("mode %q: bad interval: %w", spec, err)
		}
		m.Interval = d
		spec = spec[:i]
	}
	name, opts, _ := strings.Cut(spec, ":")
	m.Name = spec
	m.FollowNextLink = followNextLink
	switch name {
	case "current":
	case "exclude":
		m.ExtraExcludes = extraExcl
	case "expand":
		m.Expand = true
	case "improved":
		m.Expand = true
		m.ExtraExcludes = extraExcl
	default:
		return m, fmt.Errorf("unknown mode %q (current, exclude, expand, improved)", name)
	}
	if opts != "" {
		for _, o := range strings.Split(opts, "+") {
			switch {
			case o == "all":
				m.ExpandAll = true
			case o == "dot":
				m.ExpandDot = true
			case o == "oem":
				m.ExtraExcludes = extraExcl
			case o == "nextlink":
				m.FollowNextLink = true
			case strings.HasPrefix(o, "levels="):
				n, err := strconv.Atoi(strings.TrimPrefix(o, "levels="))
				if err != nil || n < 1 {
					return m, fmt.Errorf("mode %q: bad levels %q", spec, o)
				}
				m.ExpandLevels = n
			default:
				return m, fmt.Errorf("mode %q: unknown option %q (levels=N, all, dot, oem, nextlink)", spec, o)
			}
		}
	}
	if (m.ExpandAll || m.ExpandLevels > 0 || m.ExpandDot) && !m.Expand {
		return m, fmt.Errorf("mode %q: expand options need the expand or improved mode", spec)
	}
	return m, nil
}

// Attribute keys used on request spans. The SVG renderer reads them back.
const (
	attrPath      = "url.path"
	attrQuery     = "url.query"
	attrStatus    = "http.response.status_code"
	attrBodySize  = "http.response.body.size"
	attrExpand    = "redfish.expand"
	attrRefetch   = "redfish.refetch"
	attrMembers   = "redfish.members"
	attrExpanded  = "redfish.expanded_members"
	attrConnReuse = "net.conn.reused"
	attrTLS       = "tls.handshake"
	attrTLSMillis = "tls.handshake.ms"
	attrKind      = "redfish.kind" // request | login | session-check | version | logout
	attrMode      = "redfish.mode"
	attrCycle     = "redfish.cycle"
	attrInterval  = "redfish.interval_s"
)

type traverser struct {
	endpoint       *url.URL
	user           string
	password       string
	httpClient     *http.Client
	token          string
	sessionID      string
	tracer         trace.Tracer
	mode           Mode
	rule           *redfish.CollectRule
	extraExclude   *regexp.Regexp
	noExpand       *regexp.Regexp   // paths never fetched with $expand (in all mode)
	interval       time.Duration    // effective interval, recorded on the cycle span
	collections    []*regexp.Regexp // collection paths derived from metric rule paths
	learned        map[string]bool  // collection paths discovered at run time
	notFound       map[string]bool  // predicted collections that turned out not to exist (e.g. Storage/X/Drives on older iDRAC)
	followNextLink bool
	verbose        bool
}

func newTraverser(endpoint *url.URL, user, password string, rt http.RoundTripper, timeout time.Duration, tracer trace.Tracer, mode Mode, rule *redfish.CollectRule, noExpand string, interval time.Duration, verbose bool) (*traverser, error) {
	t := &traverser{
		endpoint: endpoint,
		user:     user,
		password: password,
		httpClient: &http.Client{
			Transport: rt,
			Timeout:   timeout,
		},
		tracer:         tracer,
		mode:           mode,
		rule:           rule,
		learned:        map[string]bool{},
		notFound:       map[string]bool{},
		followNextLink: mode.FollowNextLink,
		verbose:        verbose,
	}
	if len(mode.ExtraExcludes) > 0 {
		r, err := regexp.Compile(strings.Join(mode.ExtraExcludes, "|"))
		if err != nil {
			return nil, fmt.Errorf("mode %s: bad extra exclude: %w", mode.Name, err)
		}
		t.extraExclude = r
	}
	if mode.Expand {
		t.collections = collectionPatterns(rule)
	}
	if noExpand != "" {
		r, err := regexp.Compile(noExpand)
		if err != nil {
			return nil, fmt.Errorf("bad -no-expand: %w", err)
		}
		t.noExpand = r
	}
	t.interval = interval
	if mode.Interval > 0 {
		t.interval = mode.Interval
	}
	return t, nil
}

// newTransport mirrors the transport used by monitor-hw (zero-value Transport + InsecureSkipVerify).
func newTransport() *http.Transport {
	return &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
}

// collectionPatterns derives collection resource paths from metric rule paths:
// "/redfish/v1/Chassis/{chassis}/PCIeDevices/{device}" implies that
// "/redfish/v1/Chassis/{chassis}/PCIeDevices" is a collection.
func collectionPatterns(rule *redfish.CollectRule) []*regexp.Regexp {
	seen := map[string]bool{}
	var pats []*regexp.Regexp
	placeholder := regexp.MustCompile(`\{[^}]+\}`)
	for _, mr := range rule.MetricRules {
		segs := strings.Split(mr.Path, "/")
		for i := len(segs) - 1; i > 0; i-- {
			if !placeholder.MatchString(segs[i]) {
				continue
			}
			parent := strings.Join(segs[:i], "/")
			if parent == "" || seen[parent] {
				continue
			}
			seen[parent] = true
			marked := placeholder.ReplaceAllString(parent, "\x00")
			expr := "^" + strings.ReplaceAll(regexp.QuoteMeta(marked), "\x00", "[^/]+") + "$"
			pats = append(pats, regexp.MustCompile(expr))
		}
	}
	return pats
}

func (t *traverser) needTraverse(path string) bool {
	if !t.rule.TraverseRule.NeedTraverse(path) {
		return false
	}
	if t.extraExclude != nil && t.extraExclude.MatchString(path) {
		return false
	}
	return true
}

// shouldExpand decides whether a GET carries $expand.
func (t *traverser) shouldExpand(path string) bool {
	if !t.mode.Expand {
		return false
	}
	if t.mode.ExpandAll {
		return t.noExpand == nil || !t.noExpand.MatchString(path)
	}
	return t.isCollection(path)
}

func (t *traverser) isCollection(path string) bool {
	if t.learned[path] {
		return true
	}
	for _, p := range t.collections {
		if p.MatchString(path) {
			return true
		}
	}
	return false
}

// cycleResult holds everything one traversal cycle produced.
type cycleResult struct {
	Data     map[string]*gabs.Container
	Duration time.Duration
	Err      error
}

// runCycle reproduces one monitor-hw Update(): session check (+login), version GET, traversal.
func (t *traverser) runCycle(ctx context.Context, cycle int) cycleResult {
	ctx, span := t.tracer.Start(ctx, "cycle "+t.mode.Name,
		trace.WithAttributes(attribute.String(attrMode, t.mode.Name), attribute.Int(attrCycle, cycle),
			attribute.Float64(attrInterval, t.interval.Seconds())))
	defer span.End()
	start := time.Now()

	if err := t.login(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return cycleResult{Err: err, Duration: time.Since(start)}
	}
	// monitor-hw's ruleGetter calls GetVersion (GET /redfish/v1/) on every Update.
	if _, vspan, err := t.fetch(ctx, "/redfish/v1/", false, "version"); vspan != nil {
		vspan.End()
		if err != nil {
			span.RecordError(err)
		}
	}

	data := map[string]*gabs.Container{}
	t.get(ctx, t.rule.TraverseRule.Root, data)
	span.SetAttributes(attribute.Int("redfish.resources", len(data)))
	return cycleResult{Data: data, Duration: time.Since(start)}
}

func (t *traverser) get(ctx context.Context, path string, data map[string]*gabs.Container) {
	if !t.needTraverse(path) {
		return
	}
	if _, ok := data[path]; ok {
		return
	}

	if t.mode.Expand {
		// Prefer fetching the parent collection expanded before GETting one of its members
		// individually (members are often reached first through cross-links, e.g. PCIeFunctions
		// -> PCIeDevice). This keeps the request count independent of traversal order.
		if parent := parentPath(path); parent != "" && t.isCollection(parent) && t.needTraverse(parent) && !t.notFound[parent] {
			if _, ok := data[parent]; !ok {
				t.get(ctx, parent, data)
				if _, ok := data[path]; ok {
					return
				}
				if _, ok := data[parent]; !ok {
					t.notFound[parent] = true
				}
			}
		}
	}

	expand := t.shouldExpand(path)
	parsed, span, err := t.fetch(ctx, path, expand, "request")
	if err != nil {
		if span != nil {
			span.End()
		}
		return
	}

	if t.mode.Expand && !t.mode.ExpandAll && !expand {
		// Not predicted as a collection but it is one: learn it and refetch expanded
		// when that saves requests (N members >= 2 -> 1 extra request instead of N).
		if n := memberCount(parsed); n >= 2 {
			t.learned[path] = true
			span.SetAttributes(attribute.Int(attrMembers, n))
			p2, span2, err2 := t.fetch(ctx, path, true, "request")
			if err2 == nil {
				span2.SetAttributes(attribute.Bool(attrRefetch, true))
				parsed = p2
				expand = true
				span.End()
				span = span2
			} else if span2 != nil {
				span2.End()
			}
		}
	}

	data[path] = parsed
	if expand {
		t.storeExpanded(ctx, path, parsed, data, span)
	}
	span.End() // end before following links so the span measures this request only
	t.follow(ctx, parsed, data)
}

// storeExpanded registers every resource the BMC inlined into an expanded response under
// its own @odata.id, so metric rules match exactly as in the per-resource traversal. With
// $levels>=2 this recurses into nested resources. Sub-objects addressed by a fragment
// ("/Power#/PowerSupplies/0") are not separate resources and are left alone.
//
// iDRAC pages collections at 50 members and announces the rest via Members@odata.nextLink.
// monitor-hw never follows nextLink (follow() only looks at @odata.id), so by default this
// tool does not either; -follow-nextlink (or the nextlink option) enables it.
func (t *traverser) storeExpanded(ctx context.Context, self string, parsed *gabs.Container, data map[string]*gabs.Container, span trace.Span) {
	stored := 0
	var walk func(c *gabs.Container, top bool)
	walk = func(c *gabs.Container, top bool) {
		if m, err := c.ChildrenMap(); err == nil {
			if !top {
				if id, ok := m["@odata.id"]; ok {
					if p, ok := id.Data().(string); ok && len(m) > 1 && !strings.Contains(p, "#") && p != self {
						if _, dup := data[p]; !dup && t.needTraverse(p) {
							data[p] = c
							stored++
						}
					}
				}
			}
			for _, v := range m {
				walk(v, false)
			}
			return
		}
		if arr, err := c.Children(); err == nil {
			for _, v := range arr {
				walk(v, false)
			}
		}
	}
	walk(parsed, true)

	// Paginated collection: fetch the remaining pages (opt-in).
	var extraPages []*gabs.Container
	visited := map[string]bool{}
	page := parsed
	for t.followNextLink {
		next, ok := page.S("Members@odata.nextLink").Data().(string)
		if !ok || next == "" || visited[next] {
			break
		}
		visited[next] = true
		if !strings.Contains(next, "$expand") {
			if strings.Contains(next, "?") {
				next += "&" + t.mode.expandQuery()
			} else {
				next += "?" + t.mode.expandQuery()
			}
		}
		np, nspan, err := t.fetchRaw(ctx, next, "request", true)
		if nspan != nil {
			nspan.End()
		}
		if err != nil {
			break
		}
		walk(np, true)
		extraPages = append(extraPages, np)
		page = np
	}
	if span != nil {
		span.SetAttributes(attribute.Int(attrExpanded, stored))
	}
	for _, np := range extraPages {
		t.follow(ctx, np, data)
	}
}

func parentPath(path string) string {
	i := strings.LastIndex(path, "/")
	if i <= 0 {
		return ""
	}
	return path[:i]
}

func memberCount(parsed *gabs.Container) int {
	members, err := parsed.S("Members").Children()
	if err != nil {
		return 0
	}
	return len(members)
}

func (t *traverser) follow(ctx context.Context, parsed *gabs.Container, data map[string]*gabs.Container) {
	if childrenMap, err := parsed.ChildrenMap(); err == nil {
		// monitor-hw iterates the map in Go's random order; sort here so runs are reproducible.
		keys := make([]string, 0, len(childrenMap))
		for k := range childrenMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := childrenMap[k]
			if k != "@odata.id" {
				t.follow(ctx, v, data)
			} else if path, ok := v.Data().(string); ok {
				t.get(ctx, path, data)
			}
		}
		return
	}
	if childrenSlice, err := parsed.Children(); err == nil {
		for _, v := range childrenSlice {
			t.follow(ctx, v, data)
		}
	}
}

// fetch GETs a resource path, optionally with $expand, and returns the parsed body and its
// (still open) span. The caller must call span.End().
func (t *traverser) fetch(ctx context.Context, path string, expand bool, kind string) (*gabs.Container, trace.Span, error) {
	target := path
	if expand {
		target = path + "?" + t.mode.expandQuery()
	}
	return t.fetchRaw(ctx, target, kind, expand)
}

func (t *traverser) fetchRaw(ctx context.Context, rawPathAndQuery string, kind string, expand bool) (*gabs.Container, trace.Span, error) {
	// monitor-hw runs with NoEscape=true for Dell: the path is parsed as-is.
	u, err := t.endpoint.Parse(rawPathAndQuery)
	if err != nil {
		return nil, nil, err
	}
	name := "GET " + u.Path
	ctx, span := t.tracer.Start(ctx, name, trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(
		attribute.String(attrKind, kind),
		attribute.String(attrPath, u.Path),
		attribute.String(attrQuery, u.RawQuery),
		attribute.Bool(attrExpand, expand),
	))
	// The span is returned un-ended so callers can attach post-processing attributes; they must End() it.

	req, err := http.NewRequestWithContext(withConnTrace(ctx, span), "GET", u.String(), nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, span, err
	}
	req.Header.Set("X-Auth-Token", t.token)
	req.Header.Set("Accept", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		if t.verbose {
			fmt.Printf("  ! GET %s: %v\n", u.RequestURI(), err)
		}
		return nil, span, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	span.SetAttributes(attribute.Int(attrStatus, resp.StatusCode), attribute.Int(attrBodySize, len(body)))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, span, err
	}
	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("%d: %s", resp.StatusCode, u.RequestURI())
		span.SetStatus(codes.Error, err.Error())
		if t.verbose {
			fmt.Printf("  ! GET %s: %s\n", u.RequestURI(), resp.Status)
		}
		return nil, span, err
	}
	parsed, err := gabs.ParseJSON(body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, span, err
	}
	if t.verbose {
		fmt.Printf("  GET %s -> %d (%d bytes)\n", u.RequestURI(), resp.StatusCode, len(body))
	}
	return parsed, span, nil
}

// withConnTrace attaches an httptrace that records connection reuse and TLS handshakes on the span.
func withConnTrace(ctx context.Context, span trace.Span) context.Context {
	var tlsStart time.Time
	ct := &httptrace.ClientTrace{
		GotConn: func(info httptrace.GotConnInfo) {
			span.SetAttributes(attribute.Bool(attrConnReuse, info.Reused))
		},
		ConnectStart: func(network, addr string) {
			span.AddEvent("connect.start")
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
			span.AddEvent("tls.start")
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, err error) {
			span.AddEvent("tls.done")
			span.SetAttributes(attribute.Bool(attrTLS, true),
				attribute.Float64(attrTLSMillis, float64(time.Since(tlsStart).Microseconds())/1000))
			if err != nil {
				span.RecordError(err)
			}
		},
	}
	return httptrace.WithClientTrace(ctx, ct)
}

// ---- session handling, same flow as monitor-hw's Login()/checkSession() ----

type sessionLoginRequest struct {
	Username string `json:"UserName"`
	Password string `json:"Password"`
}

func (t *traverser) login(ctx context.Context) error {
	if t.sessionID != "" && t.checkSession(ctx) == nil {
		return nil
	}
	ctx, span := t.tracer.Start(ctx, "POST /redfish/v1/SessionService/Sessions", trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.String(attrKind, "login"), attribute.String(attrPath, "/redfish/v1/SessionService/Sessions")))
	defer span.End()

	body, _ := json.Marshal(sessionLoginRequest{Username: t.user, Password: t.password})
	u := t.endpoint.JoinPath("/redfish/v1/SessionService/Sessions")
	req, err := http.NewRequestWithContext(withConnTrace(ctx, span), "POST", u.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("login: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	span.SetAttributes(attribute.Int(attrStatus, resp.StatusCode), attribute.Int(attrBodySize, len(raw)))
	if resp.StatusCode != http.StatusCreated {
		err := fmt.Errorf("login: status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	var ses struct{ Id string }
	if err := json.Unmarshal(raw, &ses); err != nil {
		return fmt.Errorf("login: %w", err)
	}
	t.sessionID = ses.Id
	t.token = resp.Header.Get("X-Auth-Token")
	if t.token == "" {
		return errors.New("login: no X-Auth-Token in response")
	}
	return nil
}

func (t *traverser) checkSession(ctx context.Context) error {
	path := "/redfish/v1/SessionService/Sessions/" + t.sessionID
	ctx, span := t.tracer.Start(ctx, "GET "+path, trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.String(attrKind, "session-check"), attribute.String(attrPath, path)))
	defer span.End()
	req, err := http.NewRequestWithContext(withConnTrace(ctx, span), "GET", t.endpoint.JoinPath(path).String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", t.token)
	resp, err := t.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	span.SetAttributes(attribute.Int(attrStatus, resp.StatusCode), attribute.Int(attrBodySize, len(raw)))
	if resp.StatusCode != http.StatusOK {
		span.SetStatus(codes.Error, resp.Status)
		return fmt.Errorf("session check: %d", resp.StatusCode)
	}
	return nil
}

// logout deletes the session so the BMC's session slots are not exhausted by repeated runs.
func (t *traverser) logout(ctx context.Context) error {
	if t.sessionID == "" {
		return nil
	}
	path := "/redfish/v1/SessionService/Sessions/" + t.sessionID
	ctx, span := t.tracer.Start(ctx, "DELETE "+path, trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.String(attrKind, "logout"), attribute.String(attrPath, path)))
	defer span.End()
	req, err := http.NewRequestWithContext(withConnTrace(ctx, span), "DELETE", t.endpoint.JoinPath(path).String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", t.token)
	resp, err := t.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	span.SetAttributes(attribute.Int(attrStatus, resp.StatusCode))
	t.sessionID, t.token = "", ""
	return nil
}
