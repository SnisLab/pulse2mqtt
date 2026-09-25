package data

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"

	log "github.com/DjSni/go-log"
	sml "github.com/DjSni/go-sml"

	"pulse2mqtt/include/pulse"
	"pulse2mqtt/include/settings"
)

type Data struct {
	NodeValue struct {
		Total struct {
			Consume string
			Feed    string
		}
		Current struct {
			Consume string
		}
	}
}

// PrintMessage prints the SML message.
func PrintMessage(msg sml.Message, result *Data) {
	list, ok := msg.MessageBody.Data.(sml.GetListResponse)
	if !ok {
		panic("Could not cast list reponse")
	}

	for _, elem := range list.ValList {
		PrintListEntry(elem, result)
	}
}

// Octet2Obis converts an octet string to an OBIS code.
func Octet2Obis(o sml.OctetString) string {
	if len(o) < 6 {
		return ""
	}
	return fmt.Sprintf("%d-%d:%d.%d.%d*%d", o[0], o[1], o[2], o[3], o[4], o[5])

}

// ListEntry2Float converts a list entry to a float value.
func ListEntry2Float(entry sml.ListEntry) float64 {
	scaler := 1
	if entry.Scaler != 0 {
		scaler = int(entry.Scaler)
	}

	return float64(entry.Value.DataInt) * math.Pow10(scaler)
}

// PrintListEntry prints a list entry.
func PrintListEntry(entry sml.ListEntry, result *Data) {
	obis := Octet2Obis(entry.ObjName)
	if obis == "" || result == nil {
		return
	}
	//fmt.Printf("%-22s", obis)

	if ((entry.Value.Typ & sml.TYPEFIELD) == sml.TYPEINTEGER) || ((entry.Value.Typ & sml.TYPEFIELD) == sml.TYPEUNSIGNED) {
		value := ListEntry2Float(entry)

		unit := ""
		switch entry.Unit {
		case 0x1B:
			unit = "W"
		case 0x1E:
			unit = "kWh"
		}

		switch obis {
		case "1-0:1.8.0*255":
			// TotalConsume
			value = value / 1000
			s := fmt.Sprintf("%.4f", value)
			result.NodeValue.Total.Consume = s + " " + unit
		case "1-0:2.8.0*255":
			// TotalFeed
			value = value / 1000
			s := fmt.Sprintf("%.4f", value)
			result.NodeValue.Total.Feed = s + " " + unit
		case "1-0:16.7.0*255":
			// CurrentConsume
			value = value / 10
			s := fmt.Sprintf("%.f", value)
			result.NodeValue.Current.Consume = s + " " + unit
		}
	}
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

// GetData retrieves the data from the specified URL and parses the response.
func GetData() Data {
	var result Data
	baseURL := "http://" + settings.Load.Service.Pulse.IP
	query := "?node_id=" + strconv.Itoa(settings.Load.Service.Pulse.Node)
	req, err := http.NewRequest(http.MethodGet, baseURL+pulse.DataPath()+query, nil)
	if err != nil {
		log.Error("Can not create data request:", err)
		return result
	}
	req.SetBasicAuth(settings.Load.Service.Pulse.User, settings.Load.Service.Pulse.Password)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("No response from request:", err)
		return result
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Error("Data request returned:", resp.Status)
		return result
	}
	// response body is []byte
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Get no body -> ", err)
		return result
	}

	if body == nil {
		log.Error("No body found")
		log.Error(" ")
		log.Error(" ")
	} else if len(body) < 16 {
		log.Error(" ")
		log.Error("Body len is", len(body))
		log.Error("Body:", body)
		log.Error(" ")
		log.Error(" ")
	} else {
		messages, err := parseSML(body)
		if err != nil {
			log.Error("Parse error:", err)
			return result
		}
		for _, msg := range messages {
			if msg.MessageBody.Tag == sml.MESSAGEGETLISTRESPONSE {
				PrintMessage(msg, &result)
			}
		}

		log.Debug("Stromverbrauch:", result.NodeValue.Total.Consume)
		log.Debug("Stromeinspeißung:", result.NodeValue.Total.Feed)
		log.Debug("Aktueller verbrauch:", result.NodeValue.Current.Consume)
	}
	return result
}
