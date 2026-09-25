package pulse

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pulse2mqtt/include/settings"
)

func TestInitializeAutoSelectsModernEndpoint(t *testing.T) {
	originalSettings := settings.Load
	t.Cleanup(func() { settings.Load = originalSettings })

	var modernRequests, legacyRequests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/node_data.json":
			modernRequests++
			_, _ = w.Write([]byte("modern data"))
		case "/data.json":
			legacyRequests++
			http.NotFound(w, r)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	settings.Load.Service.Pulse.IP = strings.TrimPrefix(server.URL, "http://")
	settings.Load.Service.Pulse.Node = 1
	settings.Load.Service.Pulse.Version = ""

	if err := Initialize(); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}
	if DataPath() != "/node_data.json" || MetricsPath() != "/node_metrics.json" {
		t.Fatalf("modern endpoints not selected: data=%s metrics=%s", DataPath(), MetricsPath())
	}
	if modernRequests != 1 || legacyRequests != 0 {
		t.Fatalf("unexpected probe requests: modern=%d legacy=%d", modernRequests, legacyRequests)
	}
}

func TestInitializeAutoFallsBackToLegacyEndpoint(t *testing.T) {
	originalSettings := settings.Load
	t.Cleanup(func() { settings.Load = originalSettings })

	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		if r.URL.Path == "/node_data.json" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte("legacy data"))
	}))
	defer server.Close()

	settings.Load.Service.Pulse.IP = strings.TrimPrefix(server.URL, "http://")
	settings.Load.Service.Pulse.Node = 1
	settings.Load.Service.Pulse.Version = ""

	if err := Initialize(); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}
	if DataPath() != "/data.json" || MetricsPath() != "/metrics.json" {
		t.Fatalf("legacy endpoints not selected: data=%s metrics=%s", DataPath(), MetricsPath())
	}
	if len(paths) != 2 || paths[0] != "/node_data.json" || paths[1] != "/data.json" {
		t.Fatalf("unexpected probe order: %v", paths)
	}
}

func TestInitializeUsesConfiguredVersion(t *testing.T) {
	originalSettings := settings.Load
	t.Cleanup(func() { settings.Load = originalSettings })

	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	settings.Load.Service.Pulse.IP = strings.TrimPrefix(server.URL, "http://")
	settings.Load.Service.Pulse.Version = "modern"

	if err := Initialize(); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}
	if DataPath() != "/node_data.json" || MetricsPath() != "/node_metrics.json" {
		t.Fatalf("configured modern endpoints not selected")
	}
}

func TestInitializeFailsWhenNoEndpointIsAvailable(t *testing.T) {
	originalSettings := settings.Load
	t.Cleanup(func() { settings.Load = originalSettings })

	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	settings.Load.Service.Pulse.IP = strings.TrimPrefix(server.URL, "http://")
	settings.Load.Service.Pulse.Version = ""

	if err := Initialize(); err == nil {
		t.Fatal("Initialize() accepted unavailable endpoints")
	}
}
