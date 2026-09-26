package data

import (
	"io"
	"net/http"
	"strconv"
	"time"

	log "github.com/DjSni/go-log"
	sml "github.com/DjSni/go-sml"

	"pulse2mqtt/include/pulse"
	"pulse2mqtt/include/settings"
)

func getLegacyData() Data {
	baseURL := "http://" + settings.Load.Service.Pulse.IP
	query := "?node_id=" + strconv.Itoa(settings.Load.Service.Pulse.Node)
	req, err := http.NewRequest(http.MethodGet, baseURL+pulse.LegacyDataPath()+query, nil)
	if err != nil {
		log.Error("Can not create legacy data request:", err)
		return Data{}
	}
	req.SetBasicAuth(settings.Load.Service.Pulse.User, settings.Load.Service.Pulse.Password)
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		log.Error("No response from legacy data request:", err)
		return Data{}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Error("Legacy data request returned:", resp.Status)
		return Data{}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) < 16 {
		log.Error("Invalid legacy data response:", err)
		return Data{}
	}
	messages, err := sml.FileParse(body[8 : len(body)-8])
	if err != nil {
		log.Error("Legacy SML parse error:", err)
		return Data{}
	}
	return dataFromMessages(messages)
}
