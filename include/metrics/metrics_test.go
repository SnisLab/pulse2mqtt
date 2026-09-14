package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pulse2mqtt/include/settings"
	"strings"
	"testing"
)

func TestMetricsNodeStatusJSONFields(t *testing.T) {
	body := []byte(`{
		"node_status": {
			"battery_voltage": 2.98779,
			"temperature": 23.840141,
			"avg_rssi": -56.213547
		}
	}`)

	var got Metrics
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("could not unmarshal metrics: %v", err)
	}

	if got.NodeStatus.NodeBatteryVoltage != 2.98779 {
		t.Errorf("unexpected battery voltage: %f", got.NodeStatus.NodeBatteryVoltage)
	}
	if got.NodeStatus.NodeTemperature != 23.840141 {
		t.Errorf("unexpected temperature: %f", got.NodeStatus.NodeTemperature)
	}
	if got.NodeStatus.NodeAvgRssi != -56.213547 {
		t.Errorf("unexpected average RSSI: %f", got.NodeStatus.NodeAvgRssi)
	}
}

func TestEstimateBatteryLevel(t *testing.T) {
	tests := []struct {
		name    string
		voltage float64
		profile string
		want    int
		wantOK  bool
	}{
		{name: "alkaline empty", voltage: 2.0, profile: "alkaline", want: 0, wantOK: true},
		{name: "alkaline half", voltage: 2.6, profile: "alkaline", want: 50, wantOK: true},
		{name: "alkaline full", voltage: 3.2, profile: "alkaline", want: 100, wantOK: true},
		{name: "lfb empty", voltage: 2.7, profile: "lfb_aa", want: 0, wantOK: true},
		{name: "lfb half", voltage: 2.95, profile: "lfb_aa", want: 50, wantOK: true},
		{name: "lfb above full", voltage: 3.3, profile: "lfb_aa", want: 100, wantOK: true},
		{name: "nimh half", voltage: 2.4, profile: "nimh_1_2v", want: 50, wantOK: true},
		{name: "regulated active", voltage: 3.0, profile: "regulated_1_5v", want: 80, wantOK: true},
		{name: "regulated empty", voltage: 2.79, profile: "regulated_1_5v", want: 0, wantOK: true},
		{name: "unknown", voltage: 3.0, profile: "unknown", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := EstimateBatteryLevel(tt.voltage, tt.profile)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("EstimateBatteryLevel(%f, %q) = (%d, %t), want (%d, %t)", tt.voltage, tt.profile, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestGetMetricsUsesBasicAuth(t *testing.T) {
	originalSettings := settings.Load
	t.Cleanup(func() { settings.Load = originalSettings })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "pulse-user" || password != "pulse-password" {
			t.Errorf("unexpected basic auth: user=%q password=%q ok=%t", username, password, ok)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if got, want := r.URL.Query().Get("node_id"), "7"; got != want {
			t.Errorf("unexpected node ID: got %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"node_status":{"battery_voltage":2.98779,"temperature":23.840141,"avg_rssi":-56.213547}}`))
	}))
	defer server.Close()

	settings.Load.Service.Pulse.IP = strings.TrimPrefix(server.URL, "http://")
	settings.Load.Service.Pulse.User = "pulse-user"
	settings.Load.Service.Pulse.Password = "pulse-password"
	settings.Load.Service.Pulse.Node = 7

	result := GetMetrics()

	if got, want := result.NodeStatus.NodeBatteryVoltage, 2.98779; got != want {
		t.Errorf("unexpected battery voltage: got %f, want %f", got, want)
	}
}
