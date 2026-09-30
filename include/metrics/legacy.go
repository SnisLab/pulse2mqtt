package metrics

import (
	"encoding/json"
	"strconv"

	log "github.com/DjSni/go-log"

	"pulse2mqtt/include/pulse"
	"pulse2mqtt/include/settings"
)

func getLegacyMetrics() Metrics {
	result := fetchMetrics(pulse.LegacyMetricsPath(), func(body []byte, result *Metrics) error {
		return json.Unmarshal(body, result)
	})
	result.Diagnostics.InvalidMeterReadingsCount = result.NodeStatus.InvalidMeterReadingsCount
	result.Diagnostics.ValidMeterReadingsCount = result.NodeStatus.ValidMeterReadingsCount
	result.Diagnostics.MeterUARTErrorCount9600 = result.NodeStatus.Baud9600.UARTErrorCount
	return result
}

func fetchMetrics(path string, decode func([]byte, *Metrics) error) Metrics {
	var result Metrics
	baseURL := "http://" + settings.Load.Service.Pulse.IP
	query := "?node_id=" + strconv.Itoa(settings.Load.Service.Pulse.Node)
	resp, err := pulse.Get(baseURL + path + query)
	if err != nil {
		log.Error("No response from metrics request:", err)
		return result
	}
	if resp.StatusCode != 200 {
		log.Error("Metrics request returned:", resp.Status)
		return result
	}
	if err := decode(resp.Body, &result); err != nil {
		log.Error("Can not unmarshal metrics JSON:", err)
		return Metrics{}
	}
	logMetrics(result)
	return result
}

func logMetrics(result Metrics) {
	log.Debug("NodeBatteryVoltage:", result.NodeStatus.NodeBatteryVoltage)
	log.Debug("NodeTemperature:", result.NodeStatus.NodeTemperature)
	log.Debug("NodeAvgRssi:", result.NodeStatus.NodeAvgRssi)
	log.Debug("MeterMsgCountSent:", result.NodeStatus.MeterMsgCountSent)
	log.Debug("MeterPkgCountSent:", result.NodeStatus.MeterPkgCountSent)
	log.Debug("NodeVersion:", result.HubAttachments.NodeVersion)
}
