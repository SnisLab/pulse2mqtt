package data

import (
	"bufio"
	"bytes"
	"fmt"
	"strconv"
	"time"

	log "github.com/DjSni/go-log"
	sml "github.com/DjSni/go-sml"

	"pulse2mqtt/include/pulse"
	"pulse2mqtt/include/settings"
)

func getModernData() Data {
	for attempt := 0; attempt < 2; attempt++ {
		result, err := fetchModernData()
		if err == nil {
			return result
		}
		if attempt == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		log.Error("Modern data request failed after retry:", err)
	}
	return Data{}
}

func fetchModernData() (Data, error) {
	baseURL := "http://" + settings.Load.Service.Pulse.IP
	query := "?node_id=" + strconv.Itoa(settings.Load.Service.Pulse.Node)
	resp, err := pulse.Get(baseURL + pulse.ModernDataPath() + query)
	if err != nil {
		return Data{}, fmt.Errorf("no response from modern data request: %w", err)
	}
	if resp.StatusCode != 200 {
		return Data{}, fmt.Errorf("modern data request returned: %s", resp.Status)
	}
	if len(resp.Body) < 16 {
		if err == nil {
			err = fmt.Errorf("response too short: %d bytes", len(resp.Body))
		}
		return Data{}, fmt.Errorf("invalid modern data response: %w", err)
	}
	messages, err := parseSML(resp.Body)
	if err != nil {
		return Data{}, fmt.Errorf("modern SML parse error: %w", err)
	}
	return dataFromMessages(messages), nil
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
