package metrics

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	log "github.com/DjSni/go-log"

	"pulse2mqtt/include/pulse"
	"pulse2mqtt/include/settings"
)

func getLegacyMetrics() Metrics {
	return fetchMetrics(pulse.LegacyMetricsPath(), func(body []byte, result *Metrics) error {
		return json.Unmarshal(body, result)
	})
}

func fetchMetrics(path string, decode func([]byte, *Metrics) error) Metrics {
	var result Metrics
	baseURL := "http://" + settings.Load.Service.Pulse.IP
	query := "?node_id=" + strconv.Itoa(settings.Load.Service.Pulse.Node)
	req, err := http.NewRequest(http.MethodGet, baseURL+path+query, nil)
	if err != nil {
		log.Error("Can not create metrics request:", err)
		return result
	}
	req.SetBasicAuth(settings.Load.Service.Pulse.User, settings.Load.Service.Pulse.Password)
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		log.Error("No response from metrics request:", err)
		return result
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Error("Metrics request returned:", resp.Status)
		return result
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Can not read metrics response:", err)
		return result
	}
	if err := decode(body, &result); err != nil {
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
