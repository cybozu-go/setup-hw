package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// offlineTransport serves Redfish responses from a dump file (a JSON object keyed by
// resource path, as produced by `collector show`) so the tool can be exercised without a BMC.
// It emulates SessionService and $expand=*($levels=1) on collections.
type offlineTransport struct {
	data    map[string]json.RawMessage
	latency time.Duration
	token   string
}

func newOfflineTransport(path string, latency time.Duration) (*offlineTransport, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data map[string]json.RawMessage
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &offlineTransport{data: data, latency: latency, token: "offline-token"}, nil
}

func (o *offlineTransport) respond(req *http.Request, status int, body []byte, hdr http.Header) *http.Response {
	if o.latency > 0 {
		time.Sleep(o.latency)
	}
	if hdr == nil {
		hdr = http.Header{}
	}
	hdr.Set("Content-Type", "application/json")
	return &http.Response{
		StatusCode: status, Status: fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header: hdr, Body: io.NopCloser(bytes.NewReader(body)), ContentLength: int64(len(body)), Request: req,
	}
}

func (o *offlineTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	p := req.URL.Path
	switch {
	case req.Method == "POST" && p == "/redfish/v1/SessionService/Sessions":
		h := http.Header{}
		h.Set("X-Auth-Token", o.token)
		return o.respond(req, http.StatusCreated, []byte(`{"Id":"offline-session"}`), h), nil
	case strings.HasPrefix(p, "/redfish/v1/SessionService/Sessions/"):
		if req.Header.Get("X-Auth-Token") != o.token {
			return o.respond(req, http.StatusUnauthorized, []byte(`{}`), nil), nil
		}
		return o.respond(req, http.StatusOK, []byte(`{"Id":"offline-session"}`), nil), nil
	}
	if req.Header.Get("X-Auth-Token") != o.token {
		return o.respond(req, http.StatusUnauthorized, []byte(`{"error":"no session"}`), nil), nil
	}
	body, ok := o.lookup(p)
	if !ok {
		return o.respond(req, http.StatusNotFound, []byte(`{"error":"not found"}`), nil), nil
	}
	if skip := req.URL.Query().Get("$skip"); skip != "" {
		n, _ := strconv.Atoi(skip)
		body = o.page(p, body, n)
	}
	if strings.Contains(req.URL.RawQuery, "$expand") {
		body = o.expand(body)
	}
	return o.respond(req, http.StatusOK, body, nil), nil
}

func (o *offlineTransport) lookup(p string) (json.RawMessage, bool) {
	if b, ok := o.data[p]; ok {
		return b, true
	}
	if b, ok := o.data[strings.TrimSuffix(p, "/")]; ok {
		return b, true
	}
	return nil, false
}

// page emulates iDRAC's 50-member paging for a $skip request: Members[n:n+50] plus nextLink.
func (o *offlineTransport) page(path string, body json.RawMessage, n int) json.RawMessage {
	var obj map[string]json.RawMessage
	if json.Unmarshal(body, &obj) != nil {
		return body
	}
	var members []json.RawMessage
	if json.Unmarshal(obj["Members"], &members) != nil {
		return body
	}
	// The dump holds the first page only (50 of N); synthesize the remainder as bare links
	// pointing at the paths the dump knows under this collection.
	var all []json.RawMessage
	prefix := path + "/"
	for k := range o.data {
		if strings.HasPrefix(k, prefix) && !strings.Contains(strings.TrimPrefix(k, prefix), "/") {
			all = append(all, json.RawMessage(fmt.Sprintf(`{"@odata.id":%q}`, k)))
		}
	}
	if len(all) > len(members) {
		members = all
	}
	const pageSize = 50
	if n > len(members) {
		n = len(members)
	}
	end := n + pageSize
	if end >= len(members) {
		end = len(members)
		delete(obj, "Members@odata.nextLink")
	} else {
		obj["Members@odata.nextLink"], _ = json.Marshal(fmt.Sprintf("%s?$skip=%d", path, end))
	}
	obj["Members"], _ = json.Marshal(members[n:end])
	out, _ := json.Marshal(obj)
	return out
}

// expand inlines Members one level, like iDRAC's $expand=*($levels=1) on a collection.
func (o *offlineTransport) expand(body json.RawMessage) json.RawMessage {
	var obj map[string]json.RawMessage
	if json.Unmarshal(body, &obj) != nil {
		return body
	}
	var members []map[string]json.RawMessage
	if raw, ok := obj["Members"]; !ok || json.Unmarshal(raw, &members) != nil {
		return body
	}
	for i, m := range members {
		var id string
		if json.Unmarshal(m["@odata.id"], &id) != nil {
			continue
		}
		if full, ok := o.lookup(id); ok {
			var fm map[string]json.RawMessage
			if json.Unmarshal(full, &fm) == nil {
				members[i] = fm
			}
		}
	}
	mb, _ := json.Marshal(members)
	obj["Members"] = mb
	out, _ := json.Marshal(obj)
	return out
}
