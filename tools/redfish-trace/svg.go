package main

import (
	"fmt"
	"html"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// reqInfo is a flattened view of a request span for rendering and statistics.
type reqInfo struct {
	span     sdktrace.ReadOnlySpan
	kind     string
	path     string
	query    string
	status   int
	bytes    int
	expand   bool
	refetch  bool
	fallback bool
	reused   bool
	tls      bool
	tlsMs    float64
	expanded int
	err      bool
	start    time.Time
	dur      time.Duration
}

type cycleInfo struct {
	span     sdktrace.ReadOnlySpan
	mode     string
	cycle    int
	reqs     []reqInfo
	start    time.Time
	dur      time.Duration
	interval time.Duration
}

// ReqPerHour is the steady-state request rate a monitor-hw loop with the given interval
// would put on the BMC: requests / (traversal duration + interval).
// BusyPercent is the share of wall time the BMC spends serving the traversal loop.
func (s cycleStats) BusyPercent(interval time.Duration) float64 {
	period := s.Duration + interval
	if period <= 0 {
		return 0
	}
	return 100 * float64(s.Duration) / float64(period)
}

func (s cycleStats) ReqPerHour(interval time.Duration) float64 {
	period := s.Duration + interval
	if period <= 0 {
		return 0
	}
	return float64(s.Requests) * 3600 / period.Seconds()
}

type cycleStats struct {
	Requests, OK, NotModified, Errors, Expand, Refetch, Fallback, TLS, Reused, Expanded int
	Bytes                                                                               int
	Duration                                                                            time.Duration
	Lat                                                                                 latency // all requests
	LatPlain                                                                            latency // GET without $expand
	LatExpand                                                                           latency // GET with $expand
	Slowest                                                                             []reqInfo
	ErrList                                                                             []reqInfo      // every failed request, in time order
	ErrByKind                                                                           map[string]int // "404", "timeout", "connection", ...
}

// errKind classifies a failed request for reporting.
func (r reqInfo) errKind() string {
	if !r.err {
		return ""
	}
	if r.status != 0 {
		return strconv.Itoa(r.status)
	}
	d := r.span.Status().Description
	switch {
	case strings.Contains(d, "Client.Timeout") || strings.Contains(d, "deadline exceeded"):
		return "timeout"
	case strings.Contains(d, "context canceled"):
		return "canceled"
	case strings.Contains(d, "connection refused") || strings.Contains(d, "connection reset") || strings.Contains(d, "EOF"):
		return "connection"
	case strings.Contains(d, "tls"):
		return "tls"
	default:
		return "error"
	}
}

func (s cycleStats) errSummary() string {
	if len(s.ErrByKind) == 0 {
		return "none"
	}
	keys := make([]string, 0, len(s.ErrByKind))
	for k := range s.ErrByKind {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s×%d", k, s.ErrByKind[k]))
	}
	return strings.Join(parts, " ")
}

// latency holds percentiles of per-request durations.
type latency struct {
	N                         int
	P50, P90, P99, Max, Total time.Duration
}

func percentiles(ds []time.Duration) latency {
	if len(ds) == 0 {
		return latency{}
	}
	sorted := append([]time.Duration(nil), ds...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	pick := func(q float64) time.Duration {
		i := int(q*float64(len(sorted)-1) + 0.5)
		return sorted[i]
	}
	var total time.Duration
	for _, d := range sorted {
		total += d
	}
	return latency{N: len(sorted), P50: pick(0.50), P90: pick(0.90), P99: pick(0.99), Max: sorted[len(sorted)-1], Total: total}
}

func (l latency) String() string {
	if l.N == 0 {
		return "-"
	}
	return fmt.Sprintf("p50=%s p90=%s p99=%s max=%s", fmtDur(l.P50), fmtDur(l.P90), fmtDur(l.P99), fmtDur(l.Max))
}

func (c *cycleInfo) stats() cycleStats {
	s := cycleStats{Duration: c.dur, ErrByKind: map[string]int{}}
	var all, plain, expand []time.Duration
	for _, r := range c.reqs {
		if r.err {
			s.ErrList = append(s.ErrList, r)
			s.ErrByKind[r.errKind()]++
		}
		if r.kind == "request" {
			all = append(all, r.dur)
			if r.expand {
				expand = append(expand, r.dur)
			} else {
				plain = append(plain, r.dur)
			}
		}
		s.Requests++
		s.Bytes += r.bytes
		switch {
		case r.err:
			s.Errors++
		case r.status == 304:
			s.NotModified++
		case r.status >= 200 && r.status < 300:
			s.OK++
		}
		if r.expand {
			s.Expand++
		}
		if r.refetch {
			s.Refetch++
		}
		if r.fallback {
			s.Fallback++
		}
		if r.tls {
			s.TLS++
		}
		if r.reused {
			s.Reused++
		}
		s.Expanded += r.expanded
	}
	s.Lat, s.LatPlain, s.LatExpand = percentiles(all), percentiles(plain), percentiles(expand)
	slow := append([]reqInfo(nil), c.reqs...)
	sort.Slice(slow, func(i, j int) bool { return slow[i].dur > slow[j].dur })
	if len(slow) > 5 {
		slow = slow[:5]
	}
	s.Slowest = slow
	return s
}

func attrs(s sdktrace.ReadOnlySpan) map[attribute.Key]attribute.Value {
	m := map[attribute.Key]attribute.Value{}
	for _, kv := range s.Attributes() {
		m[kv.Key] = kv.Value
	}
	return m
}

// groupCycles builds cycleInfo structures from the exported spans.
func groupCycles(spans []sdktrace.ReadOnlySpan) []*cycleInfo {
	byID := map[trace.SpanID]*cycleInfo{}
	var cycles []*cycleInfo
	for _, s := range spans {
		if !strings.HasPrefix(s.Name(), "cycle ") {
			continue
		}
		a := attrs(s)
		c := &cycleInfo{span: s, mode: a[attrMode].AsString(), cycle: int(a[attrCycle].AsInt64()),
			start: s.StartTime(), dur: s.EndTime().Sub(s.StartTime()),
			interval: time.Duration(a[attrInterval].AsFloat64() * float64(time.Second))}
		byID[s.SpanContext().SpanID()] = c
		cycles = append(cycles, c)
	}
	for _, s := range spans {
		c, ok := byID[s.Parent().SpanID()]
		if !ok {
			continue
		}
		a := attrs(s)
		r := reqInfo{
			span: s, kind: a[attrKind].AsString(), path: a[attrPath].AsString(), query: a[attrQuery].AsString(),
			status: int(a[attrStatus].AsInt64()), bytes: int(a[attrBodySize].AsInt64()),
			expand: a[attrExpand].AsBool(), refetch: a[attrRefetch].AsBool(), fallback: a[attrFallback].AsBool(), reused: a[attrConnReuse].AsBool(),
			tls: a[attrTLS].AsBool(), tlsMs: a[attrTLSMillis].AsFloat64(), expanded: int(a[attrExpanded].AsInt64()),
			err: s.Status().Code == codes.Error, start: s.StartTime(), dur: s.EndTime().Sub(s.StartTime()),
		}
		c.reqs = append(c.reqs, r)
	}
	for _, c := range cycles {
		sort.Slice(c.reqs, func(i, j int) bool { return c.reqs[i].start.Before(c.reqs[j].start) })
	}
	sort.Slice(cycles, func(i, j int) bool { return cycles[i].start.Before(cycles[j].start) })
	return cycles
}

func (r reqInfo) color() string {
	switch {
	case r.err && r.status == 0:
		return "#7f1d1d" // transport error / timeout
	case r.err:
		return "#dc2626" // non-2xx
	case r.kind == "login" || r.kind == "session-check" || r.kind == "logout":
		return "#f59e0b"
	case r.kind == "version":
		return "#a16207"
	case r.status == 304:
		return "#2563eb"
	case r.expand:
		return "#7c3aed"
	default:
		return "#16a34a"
	}
}

func fmtDur(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func fmtBytes(n int) string {
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1fMiB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1fKiB", float64(n)/(1<<10))
	}
	return fmt.Sprintf("%dB", n)
}

// renderSVG draws, per cycle: a summary line, a compact strip of all requests on one lane,
// and (optionally) a waterfall with one row per request. The time axis is shared so bar
// lengths are comparable across modes.
func renderSVG(cycles []*cycleInfo, waterfall bool, title string) string {
	const (
		width   = 1600.0
		left    = 12.0
		right   = 12.0
		rowH    = 7.0
		stripH  = 16.0
		headerH = 74.0
		gap     = 24.0
	)
	plotW := width - left - right
	var maxDur time.Duration
	for _, c := range cycles {
		if c.dur > maxDur {
			maxDur = c.dur
		}
	}
	if maxDur == 0 {
		maxDur = time.Second
	}
	scale := plotW / float64(maxDur)

	var b strings.Builder
	total := 0.0
	panelTops := make([]float64, len(cycles))
	y := 40.0
	for i, c := range cycles {
		panelTops[i] = y
		h := headerH + stripH + 8
		if waterfall {
			h += float64(len(c.reqs))*rowH + 8
		}
		y += h + gap
	}
	total = y

	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" font-family="ui-monospace, SFMono-Regular, Menlo, monospace" font-size="11">`+"\n", width, total)
	b.WriteString(`<style>rect:hover{stroke:#000;stroke-width:1}</style>` + "\n")
	fmt.Fprintf(&b, `<rect width="100%%" height="100%%" fill="#fafafa"/>`+"\n")
	fmt.Fprintf(&b, `<text x="%.0f" y="20" font-size="14" font-weight="bold">%s</text>`+"\n", left, html.EscapeString(title))
	legend := []struct{ c, l string }{{"#16a34a", "GET 200"}, {"#7c3aed", "GET $expand"}, {"#2563eb", "304"}, {"#dc2626", "non-2xx"}, {"#7f1d1d", "error/timeout"}, {"#f59e0b", "session"}, {"#a16207", "version"}, {"#000", "| = TLS handshake"}}
	lx := left
	for _, l := range legend {
		fmt.Fprintf(&b, `<rect x="%.0f" y="26" width="10" height="10" fill="%s"/><text x="%.0f" y="35">%s</text>`, lx, l.c, lx+13, html.EscapeString(l.l))
		lx += 13 + 7.5*float64(len(l.l)) + 14
	}
	b.WriteString("\n")

	for i, c := range cycles {
		top := panelTops[i]
		st := c.stats()
		fmt.Fprintf(&b, `<text x="%.0f" y="%.0f" font-size="13" font-weight="bold">mode=%s cycle=%d</text>`+"\n", left, top+14, html.EscapeString(c.mode), c.cycle)
		summary := fmt.Sprintf("duration=%s requests=%d ok=%d errors=%d expand=%d(refetch %d, fallback %d, inlined %d) bytes=%s tls_handshakes=%d conn_reused=%d/%d | interval=%s -> %.0f req/h, busy %.1f%%",
			fmtDur(c.dur), st.Requests, st.OK, st.Errors, st.Expand, st.Refetch, st.Fallback, st.Expanded, fmtBytes(st.Bytes), st.TLS, st.Reused, st.Requests, fmtDur(c.interval), st.ReqPerHour(c.interval), st.BusyPercent(c.interval))
		fmt.Fprintf(&b, `<text x="%.0f" y="%.0f" fill="#333">%s</text>`+"\n", left, top+30, html.EscapeString(summary))
		lat := fmt.Sprintf("latency all: %s | plain: %s | $expand: %s | errors: %s", st.Lat, st.LatPlain, st.LatExpand, st.errSummary())
		fmt.Fprintf(&b, `<text x="%.0f" y="%.0f" fill="#333">%s</text>`+"\n", left, top+46, html.EscapeString(lat))

		// time axis ticks
		axisY := top + headerH - 12
		fmt.Fprintf(&b, `<line x1="%.0f" y1="%.1f" x2="%.0f" y2="%.1f" stroke="#999"/>`+"\n", left, axisY, left+plotW, axisY)
		step := niceStep(maxDur)
		for t := time.Duration(0); t <= maxDur; t += step {
			x := left + float64(t)*scale
			fmt.Fprintf(&b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#999"/><text x="%.1f" y="%.1f" font-size="9" fill="#666">%s</text>`, x, axisY-3, x, axisY+3, x+2, axisY-5, fmtDur(t))
		}
		b.WriteString("\n")

		// compact strip
		stripY := top + headerH
		fmt.Fprintf(&b, `<rect x="%.0f" y="%.1f" width="%.1f" height="%.0f" fill="#eee"/>`+"\n", left, stripY, float64(c.dur)*scale, stripH)
		for _, r := range c.reqs {
			x := left + float64(r.start.Sub(c.start))*scale
			w := float64(r.dur) * scale
			if w < 0.5 {
				w = 0.5
			}
			fmt.Fprintf(&b, `<rect x="%.2f" y="%.1f" width="%.2f" height="%.0f" fill="%s">%s</rect>`, x, stripY, w, stripH, r.color(), tooltip(r, c))
		}
		b.WriteString("\n")

		if !waterfall {
			continue
		}
		wy := stripY + stripH + 8
		for j, r := range c.reqs {
			ry := wy + float64(j)*rowH
			x := left + float64(r.start.Sub(c.start))*scale
			w := float64(r.dur) * scale
			if w < 1 {
				w = 1
			}
			fmt.Fprintf(&b, `<rect x="%.2f" y="%.1f" width="%.2f" height="%.1f" fill="%s">%s</rect>`, x, ry, w, rowH-1, r.color(), tooltip(r, c))
			if r.tls {
				fmt.Fprintf(&b, `<rect x="%.2f" y="%.1f" width="1.5" height="%.1f" fill="#000"/>`, x, ry, rowH-1)
			}
			label := r.path
			if r.expand {
				label += "?$expand"
			}
			if r.status != 0 && r.status != 200 {
				label += fmt.Sprintf(" [%d]", r.status)
			}
			fmt.Fprintf(&b, `<text x="%.2f" y="%.1f" font-size="6" fill="#444">%s</text>`+"\n", x+w+3, ry+rowH-2, html.EscapeString(label))
		}
	}
	b.WriteString("</svg>\n")
	return b.String()
}

func tooltip(r reqInfo, c *cycleInfo) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s %s", r.span.Name(), r.query)
	fmt.Fprintf(&sb, "\nstatus=%d bytes=%s duration=%s at +%s", r.status, fmtBytes(r.bytes), fmtDur(r.dur), fmtDur(r.start.Sub(c.start)))
	fmt.Fprintf(&sb, "\nconn_reused=%v tls_handshake=%v", r.reused, r.tls)
	if r.tls {
		fmt.Fprintf(&sb, " (%.1fms)", r.tlsMs)
	}
	if r.expand {
		fmt.Fprintf(&sb, "\n$expand: members inlined=%d refetch=%v", r.expanded, r.refetch)
	}
	if r.err {
		fmt.Fprintf(&sb, "\nerror: %s", r.span.Status().Description)
	}
	return "<title>" + html.EscapeString(sb.String()) + "</title>"
}

func niceStep(d time.Duration) time.Duration {
	for _, s := range []time.Duration{100 * time.Millisecond, 250 * time.Millisecond, 500 * time.Millisecond, time.Second, 2 * time.Second, 5 * time.Second, 10 * time.Second, 15 * time.Second, 30 * time.Second, time.Minute, 2 * time.Minute, 5 * time.Minute} {
		if d/s <= 16 {
			return s
		}
	}
	return 10 * time.Minute
}
