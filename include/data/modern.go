package data

import (
	"encoding/binary"
	"fmt"
	"strconv"

	log "github.com/DjSni/go-log"
	sml "github.com/DjSni/go-sml"

	"pulse2mqtt/include/pulse"
	"pulse2mqtt/include/settings"
)

func getModernData() Data {
	result, err := fetchModernData()
	if err != nil {
		log.Error("Modern data request failed:", err)
		return Data{}
	}
	return result
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
	messages, recovered, err := parseSMLWithRecovery(resp.Body)
	if err != nil {
		return Data{}, fmt.Errorf("modern SML parse error: %w", err)
	}
	result := dataFromMessages(messages)
	if recovered {
		log.Debug("Using readings from individually CRC-valid SML messages in a transport-CRC-invalid frame:",
			" consume=", result.NodeValue.Total.Consume,
			" feed=", result.NodeValue.Total.Feed,
			" current=", result.NodeValue.Current.Consume)
	}
	return result, nil
}

func parseSML(body []byte) ([]sml.Message, error) {
	messages, _, err := parseSMLWithRecovery(body)
	return messages, err
}

func parseSMLWithRecovery(body []byte) ([]sml.Message, bool, error) {
	messages, err := sml.TransportParse(body)
	if err == nil {
		return messages, false, nil
	}
	if err.Error() != "Transport CRC error" {
		return nil, false, err
	}

	// The transport CRC is invalid. Work only on a copy, correct only the outer
	// CRC, and retain only complete SML messages whose own CRCs validate. This
	// never alters the captured wire bytes and does not repair message contents.
	frame := append([]byte(nil), body...)
	if len(frame) < 2 {
		return nil, false, err
	}
	crcEnd := len(frame) - 2
	binary.BigEndian.PutUint16(frame[crcEnd:], sml.Crc16Calculate(frame[:crcEnd], crcEnd))
	payload, payloadErr := sml.TransportPayload(frame)
	if payloadErr != nil {
		return nil, false, err
	}
	recovered := individuallyValidListMessages(payload)
	if len(recovered) == 0 {
		return nil, false, err
	}
	return recovered, true, nil
}

func individuallyValidListMessages(payload []byte) []sml.Message {
	messages := make([]sml.Message, 0, 1)
	lastEnd := 0
	for offset := 0; offset < len(payload); offset++ {
		if payload[offset]&sml.TYPEFIELD != sml.TYPELIST {
			continue
		}
		buffer := &sml.Buffer{Bytes: payload, Cursor: offset}
		message, err := sml.MessageParse(buffer, true)
		if err != nil || buffer.Cursor <= offset || offset < lastEnd || message.MessageBody.Tag != sml.MESSAGEGETLISTRESPONSE {
			continue
		}
		messages = append(messages, message)
		lastEnd = buffer.Cursor
	}
	return messages
}
