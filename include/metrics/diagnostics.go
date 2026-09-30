package metrics

import (
	"encoding/json"
	"fmt"

	log "github.com/DjSni/go-log"

	"pulse2mqtt/include/pulse"
	"pulse2mqtt/include/settings"
)

type bridgeNode struct {
	NodeID    int    `json:"node_id"`
	Available *bool  `json:"available"`
	LastData  *int64 `json:"last_data_ms"`
}

type bridgeStatus struct {
	WiFi *struct {
		RSSI *int `json:"rssi"`
	} `json:"wifi_status"`
}

func getBridgeDiagnostics(result *Metrics) {
	defer func() {
		log.Debug("Pulse bridge diagnostics:",
			" invalid_meter_readings=", diagnosticValue(result.Diagnostics.InvalidMeterReadingsCount),
			" valid_meter_readings=", diagnosticValue(result.Diagnostics.ValidMeterReadingsCount),
			" uart_errors_9600=", diagnosticValue(result.Diagnostics.MeterUARTErrorCount9600),
			" meter_message_delta=", diagnosticValue(result.Diagnostics.MeterMsgCountSentDelta),
			" meter_package_delta=", diagnosticValue(result.Diagnostics.MeterPkgCountSentDelta),
			" hub_messages_received=", result.HubAttachments.MeterReadingCountRecv,
			" hub_packages_received=", result.HubAttachments.MeterPkgCountRecv,
			" hub_message_delta=", diagnosticValue(result.Diagnostics.HubMeterMsgCountReceivedDelta),
			" hub_package_delta=", diagnosticValue(result.Diagnostics.HubMeterPkgCountReceivedDelta),
			" hub_corrupt_total=", diagnosticValue(result.HubAttachments.MeterCorruptReadingCountRecv),
			" hub_corrupt_delta=", diagnosticValue(result.Diagnostics.HubCorruptReadingCountReceivedDelta),
			" packet_delivery_rate=", diagnosticValue(result.Diagnostics.PacketDeliveryRate),
			" node_available=", diagnosticValue(result.Diagnostics.NodeAvailable),
			" last_data_ms=", diagnosticValue(result.Diagnostics.LastDataAgeMs),
			" wifi_rssi=", diagnosticValue(result.Diagnostics.WiFiRSSI))
	}()
	host := "http://" + settings.Load.Service.Pulse.IP
	nodeURL := host + "/nodes.json"
	response, err := pulse.Get(nodeURL)
	if err != nil {
		log.Debug("Can not retrieve Pulse node diagnostics:", err)
	} else if response.StatusCode != 200 {
		log.Debug("Pulse node diagnostics returned:", response.Status)
	} else {
		var nodes []bridgeNode
		if err := json.Unmarshal(response.Body, &nodes); err != nil {
			var node bridgeNode
			if singleErr := json.Unmarshal(response.Body, &node); singleErr == nil {
				nodes = []bridgeNode{node}
			} else {
				log.Debug("Can not decode Pulse node diagnostics:", err)
			}
		}
		for _, node := range nodes {
			if node.NodeID != settings.Load.Service.Pulse.Node {
				continue
			}
			result.Diagnostics.NodeAvailable = node.Available
			result.Diagnostics.LastDataAgeMs = node.LastData
			break
		}
	}

	statusURL := host + "/status.json?timeout=0"
	response, err = pulse.Get(statusURL)
	if err != nil {
		log.Debug("Can not retrieve Pulse bridge status:", err)
		return
	}
	if response.StatusCode != 200 {
		log.Debug("Pulse bridge status returned:", response.Status)
		return
	}
	var status bridgeStatus
	if err := json.Unmarshal(response.Body, &status); err != nil {
		log.Debug("Can not decode Pulse bridge status:", fmt.Errorf("%w (%d bytes)", err, len(response.Body)))
		return
	}
	if status.WiFi != nil {
		result.Diagnostics.WiFiRSSI = status.WiFi.RSSI
	}
}

func diagnosticValue[T any](value *T) any {
	if value == nil {
		return nil
	}
	return *value
}
