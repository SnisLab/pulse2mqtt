package metrics

import (
	"encoding/json"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	log "github.com/DjSni/go-log"

	"pulse2mqtt/include/pulse"
	"pulse2mqtt/include/settings"
)

type Metrics struct {
	Type       string `json:"$type"`
	NodeStatus struct {
		ProductID           int     `json:"product_id"`
		BootloaderVersion   int     `json:"bootloader_version"`
		MeterMode           int     `json:"meter_mode"`
		NodeBatteryVoltage  float64 `json:"battery_voltage"`
		NodeTemperature     float64 `json:"temperature"`
		NodeAvgRssi         float64 `json:"avg_rssi"`
		NodeAvgLqi          float64 `json:"node_avg_lqi"`
		RadioTxPower        int     `json:"radio_tx_power"`
		NodeUptimeMs        int     `json:"node_uptime_ms"`
		MeterMsgCountSent   int     `json:"meter_msg_count_sent"`
		MeterPkgCountSent   int     `json:"meter_pkg_count_sent"`
		TimeInEm0Ms         int     `json:"time_in_em0_ms"`
		TimeInEm1Ms         int     `json:"time_in_em1_ms"`
		TimeInEm2Ms         int     `json:"time_in_em2_ms"`
		AcmpRxAutolevel300  int     `json:"acmp_rx_autolevel_300"`
		AcmpRxAutolevel9600 int     `json:"acmp_rx_autolevel_9600"`
	} `json:"node_status"`
	HubAttachments struct {
		MeterPkgCountRecv     int    `json:"meter_pkg_count_recv"`
		MeterReadingCountRecv int    `json:"meter_reading_count_recv"`
		NodeVersion           string `json:"node_version"`
	} `json:"hub_attachments"`
}

func batteryVoltageRange(profile string) (emptyVoltage float64, fullVoltage float64, ok bool) {
	switch profile {
	case "alkaline":
		return 2.0, 3.2, true
	case "lfb_aa":
		return 2.7, 3.2, true
	case "nimh_1_2v":
		return 2.0, 2.8, true
	default:
		return 0, 0, false
	}
}

// EstimateBatteryLevel converts voltage to an approximate charge level for a
// supported profile. Regulated cells use a fixed status instead of a curve.
func EstimateBatteryLevel(voltage float64, profile string) (int, bool) {
	if profile == "regulated_1_5v" {
		if voltage < 2.8 {
			return 0, true
		}
		return 80, true
	}

	emptyVoltage, fullVoltage, ok := batteryVoltageRange(profile)
	if !ok {
		return 0, false
	}

	level := (voltage - emptyVoltage) / (fullVoltage - emptyVoltage) * 100
	if level < 0 {
		return 0, true
	}
	if level > 100 {
		return 100, true
	}
	return int(math.Round(level)), true
}

// PrettyPrint prints a struct in a readable format.
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

// GetMetrics retrieves the metrics from the specified URL and parses the response.
func GetMetrics() Metrics {
	var result Metrics
	baseURL := "http://" + settings.Load.Service.Pulse.IP
	query := "?node_id=" + strconv.Itoa(settings.Load.Service.Pulse.Node)
	req, err := http.NewRequest(http.MethodGet, baseURL+pulse.MetricsPath()+query, nil)
	if err != nil {
		log.Error("Can not create metrics request:", err)
		return result
	}
	req.SetBasicAuth(settings.Load.Service.Pulse.User, settings.Load.Service.Pulse.Password)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("No response from request:", err)
		return result
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Error("Metrics request returned:", resp.Status)
		return result
	}
	// response body is []byte
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Can not read metrics response:", err)
		return result
	}

	// Parse []byte to go struct pointer
	if err := json.Unmarshal(body, &result); err != nil {
		log.Error("Can not unmarshal metrics JSON:", err)
		return result
	}

	log.Debug("NodeBatteryVoltage:", result.NodeStatus.NodeBatteryVoltage)
	log.Debug("NodeTemperature:", result.NodeStatus.NodeTemperature)
	log.Debug("NodeAvgRssi:", result.NodeStatus.NodeAvgRssi)
	log.Debug("MeterMsgCountSent:", result.NodeStatus.MeterMsgCountSent)
	log.Debug("MeterPkgCountSent:", result.NodeStatus.MeterPkgCountSent)
	log.Debug("NodeVersion:", result.HubAttachments.NodeVersion)
	return result
}
