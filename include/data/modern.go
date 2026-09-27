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
	req, err := http.NewRequest(http.MethodGet, baseURL+pulse.ModernDataPath()+query, nil)
	if err != nil {
		return Data{}, fmt.Errorf("can not create modern data request: %w", err)
	}
	req.SetBasicAuth(settings.Load.Service.Pulse.User, settings.Load.Service.Pulse.Password)
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return Data{}, fmt.Errorf("no response from modern data request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Data{}, fmt.Errorf("modern data request returned: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil || len(body) < 16 {
		if err == nil {
			err = fmt.Errorf("response too short: %d bytes", len(body))
		}
		return Data{}, fmt.Errorf("invalid modern data response: %w", err)
	}
	messages, err := parseSML(body)
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
