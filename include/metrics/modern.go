package metrics

import (
	"encoding/json"
	"strconv"

	"pulse2mqtt/include/pulse"
)

type modernMetrics struct {
	Node modernNode `json:"node"`
	IR   modernIR   `json:"ir"`
	Hub  modernHub  `json:"hub"`
}

type modernNode struct {
	ProductID         int     `json:"product_id"`
	NodeVersion       int     `json:"node_version"`
	BootloaderVersion int     `json:"bootloader_version"`
	MeterMode         int     `json:"meter_mode"`
	BatteryVoltage    float64 `json:"battery_voltage"`
	Temperature       float64 `json:"temperature"`
	AvgRssi           float64 `json:"avg_rssi"`
	AvgLqi            float64 `json:"avg_lqi"`
	RadioTxPower      int     `json:"radio_tx_power"`
	NodeUptime        int     `json:"node_uptime"`
	MeterMsgCountSent int     `json:"meter_msg_count_sent"`
	MeterPkgCountSent int     `json:"meter_pkg_count_sent"`
	TimeInEm0Ms       int     `json:"time_in_em0_ms"`
	TimeInEm1Ms       int     `json:"time_in_em1_ms"`
	TimeInEm2Ms       int     `json:"time_in_em2_ms"`
}

type modernIR struct {
	AcmpRxAutolevel300  int `json:"acmp_rx_autolevel_300"`
	AcmpRxAutolevel9600 int `json:"acmp_rx_autolevel_9600"`
}

type modernHub struct {
	MeterPkgCountRecv     int `json:"meter_pkg_count_received"`
	MeterReadingCountRecv int `json:"meter_msg_count_received"`
}

func getModernMetrics() Metrics {
	return fetchMetrics(pulse.ModernMetricsPath(), func(body []byte, result *Metrics) error {
		return decodeModernMetrics(body, result)
	})
}

func decodeModernMetrics(body []byte, result *Metrics) error {
	var modern modernMetrics
	if err := json.Unmarshal(body, &modern); err != nil {
		return err
	}
	result.NodeStatus.ProductID = modern.Node.ProductID
	result.NodeStatus.BootloaderVersion = modern.Node.BootloaderVersion
	result.NodeStatus.MeterMode = modern.Node.MeterMode
	result.NodeStatus.NodeBatteryVoltage = modern.Node.BatteryVoltage
	result.NodeStatus.NodeTemperature = modern.Node.Temperature
	result.NodeStatus.NodeAvgRssi = modern.Node.AvgRssi
	result.NodeStatus.NodeAvgLqi = modern.Node.AvgLqi
	result.NodeStatus.RadioTxPower = modern.Node.RadioTxPower
	result.NodeStatus.NodeUptimeMs = modern.Node.NodeUptime
	result.NodeStatus.MeterMsgCountSent = modern.Node.MeterMsgCountSent
	result.NodeStatus.MeterPkgCountSent = modern.Node.MeterPkgCountSent
	result.NodeStatus.TimeInEm0Ms = modern.Node.TimeInEm0Ms
	result.NodeStatus.TimeInEm1Ms = modern.Node.TimeInEm1Ms
	result.NodeStatus.TimeInEm2Ms = modern.Node.TimeInEm2Ms
	result.NodeStatus.AcmpRxAutolevel300 = modern.IR.AcmpRxAutolevel300
	result.NodeStatus.AcmpRxAutolevel9600 = modern.IR.AcmpRxAutolevel9600
	result.HubAttachments.MeterPkgCountRecv = modern.Hub.MeterPkgCountRecv
	result.HubAttachments.MeterReadingCountRecv = modern.Hub.MeterReadingCountRecv
	result.HubAttachments.NodeVersion = strconv.Itoa(modern.Node.NodeVersion)
	return nil
}
