package data

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sml "github.com/DjSni/go-sml"

	"pulse2mqtt/include/settings"
)

func TestOctet2Obis(t *testing.T) {
	got := Octet2Obis(sml.OctetString{1, 0, 1, 8, 0, 255})
	if got != "1-0:1.8.0*255" {
		t.Fatalf("unexpected OBIS code: got %q", got)
	}
}

func TestOctet2ObisRejectsShortInput(t *testing.T) {
	if got := Octet2Obis(sml.OctetString{1, 0, 1}); got != "" {
		t.Fatalf("short OBIS input returned %q", got)
	}
}

func TestListEntry2Float(t *testing.T) {
	tests := []struct {
		name   string
		value  int64
		scaler int8
		want   float64
	}{
		{name: "zero scaler uses application default", value: 12, scaler: 0, want: 120},
		{name: "positive scaler", value: 12, scaler: 2, want: 1200},
		{name: "negative scaler", value: 1234, scaler: -3, want: 1.234},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ListEntry2Float(sml.ListEntry{Scaler: tt.scaler, Value: sml.Value{DataInt: tt.value}}); got != tt.want {
				t.Fatalf("ListEntry2Float() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrintListEntryUpdatesKnownValues(t *testing.T) {
	var result Data

	PrintListEntry(sml.ListEntry{
		ObjName: sml.OctetString{1, 0, 1, 8, 0, 255},
		Unit:    0x1E,
		Scaler:  -3,
		Value:   sml.Value{Typ: sml.TYPEINTEGER, DataInt: 1234},
	}, &result)

	if got, want := result.NodeValue.Total.Consume, "0.0012 kWh"; got != want {
		t.Fatalf("unexpected total consumption: got %q, want %q", got, want)
	}

	PrintListEntry(sml.ListEntry{
		ObjName: sml.OctetString{1, 0, 16, 7, 0, 255},
		Unit:    0x1B,
		Value:   sml.Value{Typ: sml.TYPEINTEGER, DataInt: 123},
	}, &result)
	if got, want := result.NodeValue.Current.Consume, "123 W"; got != want {
		t.Fatalf("unexpected current consumption: got %q, want %q", got, want)
	}
}

func TestGetDataUsesBasicAuthAndNodeID(t *testing.T) {
	originalSettings := settings.Load
	t.Cleanup(func() {
		settings.Load = originalSettings
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "pulse-user" || password != "pulse-password" {
			t.Errorf("unexpected basic auth: user=%q password=%q ok=%t", username, password, ok)
		}
		if got, want := r.URL.Query().Get("node_id"), "7"; got != want {
			t.Errorf("unexpected node ID: got %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(make([]byte, 16))
	}))
	defer server.Close()

	settings.Load.Service.Pulse.IP = strings.TrimPrefix(server.URL, "http://")
	settings.Load.Service.Pulse.User = "pulse-user"
	settings.Load.Service.Pulse.Password = "pulse-password"
	settings.Load.Service.Pulse.Node = 7

	GetData()
}

func TestGetDataFallsBackToNodeDataEndpoint(t *testing.T) {
	originalSettings := settings.Load
	t.Cleanup(func() { settings.Load = originalSettings })

	var oldRequests, newRequests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/data.json" {
			oldRequests++
			http.NotFound(w, r)
			return
		}
		if r.URL.Path != "/node_data.json" {
			t.Fatalf("unexpected fallback path: %s", r.URL.Path)
		}
		newRequests++
		username, password, ok := r.BasicAuth()
		if !ok || username != "pulse-user" || password != "pulse-password" {
			t.Errorf("unexpected basic auth: user=%q password=%q ok=%t", username, password, ok)
		}
		if got, want := r.URL.Query().Get("node_id"), "7"; got != want {
			t.Errorf("unexpected node ID: got %q, want %q", got, want)
		}
		_, _ = w.Write(make([]byte, 16))
	}))
	defer server.Close()

	settings.Load.Service.Pulse.IP = strings.TrimPrefix(server.URL, "http://")
	settings.Load.Service.Pulse.User = "pulse-user"
	settings.Load.Service.Pulse.Password = "pulse-password"
	settings.Load.Service.Pulse.Node = 7

	GetData()
	GetData()
	if oldRequests != 1 || newRequests != 2 {
		t.Fatalf("unexpected endpoint requests: old=%d new=%d", oldRequests, newRequests)
	}
}

func TestGetDataIgnoresFailedResponse(t *testing.T) {
	originalSettings := settings.Load
	t.Cleanup(func() {
		settings.Load = originalSettings
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	settings.Load.Service.Pulse.IP = strings.TrimPrefix(server.URL, "http://")
	result := GetData()
	if got := result.NodeValue.Total.Consume; got != "" {
		t.Fatalf("failed response returned data result %q", got)
	}
}

func TestGetDataIgnoresInvalidSML(t *testing.T) {
	originalSettings := settings.Load
	t.Cleanup(func() {
		settings.Load = originalSettings
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(append(make([]byte, 8), 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09))
	}))
	defer server.Close()

	settings.Load.Service.Pulse.IP = strings.TrimPrefix(server.URL, "http://")
	result := GetData()
	if got := result.NodeValue.Total.Consume; got != "" {
		t.Fatalf("invalid SML returned data result %q", got)
	}
}
