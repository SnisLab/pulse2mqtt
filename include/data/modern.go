package data

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	log "github.com/DjSni/go-log"
	sml "github.com/DjSni/go-sml"

	"pulse2mqtt/include/pulse"
	"pulse2mqtt/include/settings"
)

func getModernData() Data {
	baseURL := "http://" + settings.Load.Service.Pulse.IP
	query := "?node_id=" + strconv.Itoa(settings.Load.Service.Pulse.Node)
	req, err := http.NewRequest(http.MethodGet, baseURL+pulse.ModernDataPath()+query, nil)
	if err != nil {
		log.Error("Can not create modern data request:", err)
		return Data{}
	}
	req.SetBasicAuth(settings.Load.Service.Pulse.User, settings.Load.Service.Pulse.Password)
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		log.Error("No response from modern data request:", err)
		return Data{}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Error("Modern data request returned:", resp.Status)
		return Data{}
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) < 16 {
		log.Error("Invalid modern data response:", err)
		return Data{}
	}
	messages, err := parseSML(body)
	if err != nil {
		log.Error("Modern SML parse error:", err)
		return Data{}
	}
	return dataFromMessages(messages)
}

func parseSML(body []byte) ([]sml.Message, error) {
	frame, err := sml.TransportRead(bufio.NewReader(bytes.NewReader(body)))
	if err != nil {
		return nil, err
	}
	if len(frame) < 16 {
		return nil, fmt.Errorf("SML transport frame too short: %d", len(frame))
	}
	return sml.FileParse(frame[8 : len(frame)-8])
}
