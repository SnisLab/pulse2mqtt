package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pulse2mqtt/include/pulse"
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

func TestModernMetricsJSONFields(t *testing.T) {
	body := []byte(`{
		"node": {
			"node_version": 42,
			"battery_voltage": 3.012,
			"temperature": 24.5,
			"avg_rssi": -49.0,
			"avg_lqi": 91.0,
			"node_uptime": 1234,
			"meter_msg_count_sent": 12
		},
		"ir": {"acmp_rx_autolevel_9600": 133},
		"hub": {"meter_msg_count_received": 95}
	}`)

	var got Metrics
	if err := decodeModernMetrics(body, &got); err != nil {
		t.Fatalf("could not unmarshal modern metrics: %v", err)
	}
	if got.NodeStatus.NodeBatteryVoltage != 3.012 || got.NodeStatus.NodeTemperature != 24.5 || got.NodeStatus.NodeAvgRssi != -49 || got.NodeStatus.NodeAvgLqi != 91 {
		t.Fatalf("modern node metrics were not mapped: %+v", got.NodeStatus)
	}
	if got.NodeStatus.MeterMsgCountSent != 12 || got.NodeStatus.NodeUptimeMs != 1234 || got.HubAttachments.NodeVersion != "42" || got.NodeStatus.AcmpRxAutolevel9600 != 133 {
		t.Fatalf("modern counters/version were not mapped: %+v / %q", got.NodeStatus, got.HubAttachments.NodeVersion)
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

func TestGetModernMetricsUsesModernEndpoint(t *testing.T) {
	originalSettings := settings.Load
	t.Cleanup(func() { settings.Load = originalSettings })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/node_metrics.json" {
			t.Errorf("unexpected endpoint: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"node":{"node_version":0,"battery_voltage":3.02,"temperature":20.35,"avg_rssi":-56.27,"avg_lqi":205.64,"node_uptime":39377562,"meter_msg_count_sent":12928,"meter_pkg_count_sent":15524},"ir":{"acmp_rx_autolevel_9600":133},"hub":{"meter_pkg_count_received":5628,"meter_msg_count_received":4639},"packet_delivery_rate":95.8}`))
	}))
	defer server.Close()

	settings.Load.Service.Pulse.IP = strings.TrimPrefix(server.URL, "http://")
	settings.Load.Service.Pulse.Node = 1
	settings.Load.Service.Pulse.Version = "modern"
	if err := pulse.Initialize(); err != nil {
		t.Fatal(err)
	}

	result := GetMetrics()
	if result.NodeStatus.NodeBatteryVoltage != 3.02 || result.NodeStatus.NodeTemperature != 20.35 || result.NodeStatus.NodeAvgRssi != -56.27 {
		t.Fatalf("modern metrics were not read: %+v", result.NodeStatus)
	}
	if result.NodeStatus.NodeUptimeMs != 39377562 || result.HubAttachments.MeterPkgCountRecv != 5628 {
		t.Fatalf("modern counters were not read: %+v / %+v", result.NodeStatus, result.HubAttachments)
	}
}
