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
		{name: "zero scaler uses unit scale", value: 12, scaler: 0, want: 12},
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
	original := DResult
	t.Cleanup(func() { DResult = original })

	PrintListEntry(sml.ListEntry{
		ObjName: sml.OctetString{1, 0, 1, 8, 0, 255},
		Unit:    0x1E,
		Scaler:  -3,
		Value:   sml.Value{Typ: sml.TYPEINTEGER, DataInt: 1234},
	})

	if got, want := DResult.NodeValue.Total.Consume, "0.0012 kWh"; got != want {
		t.Fatalf("unexpected total consumption: got %q, want %q", got, want)
	}
}

func TestGetDataUsesBasicAuthAndNodeID(t *testing.T) {
	originalSettings := settings.Load
	originalResult := DResult
	t.Cleanup(func() {
		settings.Load = originalSettings
		DResult = originalResult
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

func TestGetDataIgnoresFailedResponse(t *testing.T) {
	originalSettings := settings.Load
	originalResult := DResult
	t.Cleanup(func() {
		settings.Load = originalSettings
		DResult = originalResult
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	settings.Load.Service.Pulse.IP = strings.TrimPrefix(server.URL, "http://")
	DResult.NodeValue.Total.Consume = "unchanged"

	GetData()

	if got := DResult.NodeValue.Total.Consume; got != "unchanged" {
		t.Fatalf("failed response changed data result to %q", got)
	}
}

func TestGetDataIgnoresInvalidSML(t *testing.T) {
	originalSettings := settings.Load
	originalResult := DResult
	t.Cleanup(func() {
		settings.Load = originalSettings
		DResult = originalResult
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(append(make([]byte, 8), 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09))
	}))
	defer server.Close()

	settings.Load.Service.Pulse.IP = strings.TrimPrefix(server.URL, "http://")
	DResult.NodeValue.Total.Consume = "unchanged"

	GetData()

	if got := DResult.NodeValue.Total.Consume; got != "unchanged" {
		t.Fatalf("invalid SML changed data result to %q", got)
	}
}
