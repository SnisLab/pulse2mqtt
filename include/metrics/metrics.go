package metrics

import (
	"encoding/json"
	"math"
	"pulse2mqtt/include/pulse"
)

type Metrics struct {
	Type       string `json:"$type"`
	NodeStatus struct {
		ProductID                 int     `json:"product_id"`
		BootloaderVersion         int     `json:"bootloader_version"`
		MeterMode                 int     `json:"meter_mode"`
		NodeBatteryVoltage        float64 `json:"battery_voltage"`
		NodeTemperature           float64 `json:"temperature"`
		NodeAvgRssi               float64 `json:"avg_rssi"`
		NodeAvgLqi                float64 `json:"node_avg_lqi"`
		RadioTxPower              int     `json:"radio_tx_power"`
		NodeUptimeMs              int     `json:"node_uptime_ms"`
		MeterMsgCountSent         int     `json:"meter_msg_count_sent"`
		MeterPkgCountSent         int     `json:"meter_pkg_count_sent"`
		InvalidMeterReadingsCount *int    `json:"invalid_meter_readings_count"`
		ValidMeterReadingsCount   *int    `json:"valid_meter_readings_count"`
		Baud9600                  struct {
			UARTErrorCount *int `json:"uart_error_count"`
		} `json:"baud_9600"`
		TimeInEm0Ms         int `json:"time_in_em0_ms"`
		TimeInEm1Ms         int `json:"time_in_em1_ms"`
		TimeInEm2Ms         int `json:"time_in_em2_ms"`
		AcmpRxAutolevel300  int `json:"acmp_rx_autolevel_300"`
		AcmpRxAutolevel9600 int `json:"acmp_rx_autolevel_9600"`
	} `json:"node_status"`
	HubAttachments struct {
		MeterPkgCountRecv            int    `json:"meter_pkg_count_recv"`
		MeterReadingCountRecv        int    `json:"meter_reading_count_recv"`
		MeterCorruptReadingCountRecv *int   `json:"meter_corrupt_reading_count_recv"`
		NodeVersion                  string `json:"node_version"`
	} `json:"hub_attachments"`
	Diagnostics struct {
		InvalidMeterReadingsCount           *int     `json:"-"`
		ValidMeterReadingsCount             *int     `json:"-"`
		MeterUARTErrorCount9600             *int     `json:"-"`
		MeterMsgCountSentDelta              *int     `json:"-"`
		MeterPkgCountSentDelta              *int     `json:"-"`
		HubMeterMsgCountReceivedDelta       *int     `json:"-"`
		HubMeterPkgCountReceivedDelta       *int     `json:"-"`
		HubCorruptReadingCountReceivedDelta *int     `json:"-"`
		PacketDeliveryRate                  *float64 `json:"-"`
		NodeAvailable                       *bool    `json:"-"`
		LastDataAgeMs                       *int64   `json:"-"`
		WiFiRSSI                            *int     `json:"-"`
	}
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

func GetMetrics() Metrics {
	var result Metrics
	if pulse.CurrentMode() == pulse.ModeModern {
		result = getModernMetrics()
	} else {
		result = getLegacyMetrics()
	}
	getBridgeDiagnostics(&result)
	return result
}
