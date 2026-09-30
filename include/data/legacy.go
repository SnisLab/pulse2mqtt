package data

import (
	"strconv"

	log "github.com/DjSni/go-log"
	sml "github.com/DjSni/go-sml"

	"pulse2mqtt/include/pulse"
	"pulse2mqtt/include/settings"
)

func getLegacyData() Data {
	baseURL := "http://" + settings.Load.Service.Pulse.IP
	query := "?node_id=" + strconv.Itoa(settings.Load.Service.Pulse.Node)
	resp, err := pulse.Get(baseURL + pulse.LegacyDataPath() + query)
	if err != nil {
		log.Error("No response from legacy data request:", err)
		return Data{}
	}
	if resp.StatusCode != 200 {
		log.Error("Legacy data request returned:", resp.Status)
		return Data{}
	}
	if len(resp.Body) < 16 {
		log.Error("Invalid legacy data response:", err)
		return Data{}
	}
	messages, err := sml.TransportParse(resp.Body)
	if err != nil {
		log.Error("Legacy SML parse error:", err)
		return Data{}
	}
	return dataFromMessages(messages)
}
