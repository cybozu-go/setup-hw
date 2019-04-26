package lib

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/cybozu-go/setup-hw/config"
	"github.com/cybozu-go/setup-hw/redfish"
)

func TestDetectRedfishVersion(t *testing.T) {
	t.Parallel()

	redfishVersions := []struct {
		version  string
		expected bool
	}{
		{
			version:  "1.0.2",
			expected: true,
		},
		{
			version:  "1.2.0",
			expected: true,
		},
		{
			version:  "1.4.0",
			expected: true,
		},
		{
			version:  "0.0.0",
			expected: false,
		},
	}

	for _, rv := range redfishVersions {
		rv := rv
		mux := http.NewServeMux()
		mux.HandleFunc("/redfish/v1", func(w http.ResponseWriter, r *http.Request) {
			data := `{
  "@odata.id": "/redfish/v1",
  "RedfishVersion": "` + rv.version + `"
}`
			w.Write([]byte(data))
		})
		ts := httptest.NewTLSServer(mux)

		u, err := url.Parse(ts.URL)
		if err != nil {
			t.Fatal(err)
		}
		hostAndPort := strings.Split(u.Host, ":")
		if len(hostAndPort) != 2 {
			t.Fatal(errors.New("httptest.NewTLSServer() returned URL with host and/or port omitted"))
		}

		endpoint, err := url.Parse(ts.URL)
		if err != nil {
			t.Fatal(err)
		}
		uc := &config.UserConfig{}
		version, err := DetectRedfishVersion(endpoint, uc)
		if err != nil {
			t.Fatal(err)
		}

		ruleFile := fmt.Sprintf("dell_redfish_%s.yml", version)
		_, ok := redfish.Rules[ruleFile]
		if rv.expected && !ok {
			t.Error("rule is not found", ruleFile)
		}

		ts.Close()
	}
}
