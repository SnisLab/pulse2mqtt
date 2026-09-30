package data

import (
	"bytes"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	sml "github.com/DjSni/go-sml"

	"pulse2mqtt/include/pulse"
	"pulse2mqtt/include/settings"
)

const nodeDataHex = "1B1B1B1B0101010176050DE7EB4C62006200726301017601010504A2A3C40B0A0149534B000516465C7262016504A2A20862016335990076050DE7EB4D620062007263070177010B0A0149534B000516465C070100620AFFFF7262016504A2A20875770701004032010101010101010449534B0177070100600100FF010101010B0A0149534B000516465C0177070100010800FF650008010401621E52FF6503FEC26A0177070100020800FF0101621E52FF62000177070100100700FF0101621B520053017F01010163BB0E0076050DE7EB4E6200620072630201710163F1A9001B1B1B1B1A00F61F"

func TestOctet2Obis(t *testing.T) {
	got := Octet2Obis(sml.OctetString{1, 0, 1, 8, 0, 255})
	if got != "1-0:1.8.0*255" {
		t.Fatalf("unexpected OBIS code: got %q", got)
	}
}

func TestOctet2ObisRejectsShortInput(t *testing.T) {
	if got := Octet2Obis(sml.OctetString{1, 0, 1}); got != "" {
		t.Fatalf("short OBIS input returned %q", got)
	}
}

func TestListEntry2Float(t *testing.T) {
	tests := []struct {
		name   string
		value  int64
		scaler int8
		typ    uint8
		want   float64
	}{
		{name: "zero scaler uses application default", value: 12, scaler: 0, typ: sml.TYPEINTEGER, want: 120},
		{name: "positive scaler", value: 12, scaler: 2, typ: sml.TYPEINTEGER, want: 1200},
		{name: "negative scaler", value: 1234, scaler: -3, typ: sml.TYPEINTEGER, want: 1.234},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ListEntry2Float(sml.ListEntry{Scaler: tt.scaler, Value: sml.Value{Typ: tt.typ, DataInt: tt.value}}); got != tt.want {
				t.Fatalf("ListEntry2Float() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListEntry2FloatSupportsUnsigned64(t *testing.T) {
	entry := sml.ListEntry{Scaler: -1, Value: sml.Value{Typ: sml.TYPEUNSIGNED, DataUnsigned: 12345}}
	if got, want := ListEntry2Float(entry), 1234.5; got != want {
		t.Fatalf("ListEntry2Float() = %v, want %v", got, want)
	}
}

func TestPrintListEntryUpdatesKnownValues(t *testing.T) {
	var result Data

	PrintListEntry(sml.ListEntry{
		ObjName: sml.OctetString{1, 0, 1, 8, 0, 255},
		Unit:    0x1E,
		Scaler:  -3,
		Value:   sml.Value{Typ: sml.TYPEINTEGER, DataInt: 1234},
	}, &result)

	if got, want := result.NodeValue.Total.Consume, "0.0012 kWh"; got != want {
		t.Fatalf("unexpected total consumption: got %q, want %q", got, want)
	}

	PrintListEntry(sml.ListEntry{
		ObjName: sml.OctetString{1, 0, 16, 7, 0, 255},
		Unit:    0x1B,
		Value:   sml.Value{Typ: sml.TYPEINTEGER, DataInt: 123},
	}, &result)
	if got, want := result.NodeValue.Current.Consume, "123 W"; got != want {
		t.Fatalf("unexpected current consumption: got %q, want %q", got, want)
	}
}

func TestGetDataUsesBasicAuthAndNodeID(t *testing.T) {
	originalSettings := settings.Load
	t.Cleanup(func() {
		settings.Load = originalSettings
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "pulse-user" || password != "pulse-password" {
			t.Errorf("unexpected basic auth: user=%q password=%q ok=%t", username, password, ok)
		}
		if got, want := r.URL.Query().Get("node_id"), "7"; got != want {
			t.Errorf("unexpected node ID: got %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(make([]byte, 16))
	}))
	defer server.Close()

	settings.Load.Service.Pulse.IP = strings.TrimPrefix(server.URL, "http://")
	settings.Load.Service.Pulse.User = "pulse-user"
	settings.Load.Service.Pulse.Password = "pulse-password"
	settings.Load.Service.Pulse.Node = 7

	GetData()
}

func TestGetDataIgnoresFailedResponse(t *testing.T) {
	originalSettings := settings.Load
	t.Cleanup(func() {
		settings.Load = originalSettings
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	settings.Load.Service.Pulse.IP = strings.TrimPrefix(server.URL, "http://")
	result := GetData()
	if got := result.NodeValue.Total.Consume; got != "" {
		t.Fatalf("failed response returned data result %q", got)
	}
}

func TestGetDataIgnoresInvalidSML(t *testing.T) {
	originalSettings := settings.Load
	t.Cleanup(func() {
		settings.Load = originalSettings
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(append(make([]byte, 8), 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09))
	}))
	defer server.Close()

	settings.Load.Service.Pulse.IP = strings.TrimPrefix(server.URL, "http://")
	result := GetData()
	if got := result.NodeValue.Total.Consume; got != "" {
		t.Fatalf("invalid SML returned data result %q", got)
	}
}

func TestModernPulseFrameUsesCommonParser(t *testing.T) {
	hexData, err := os.ReadFile("testdata/node_data_strict_valid.hex")
	if err != nil {
		t.Fatal(err)
	}
	frame, err := hex.DecodeString(strings.TrimSpace(string(hexData)))
	if err != nil {
		t.Fatal(err)
	}
	messages, err := parseSML(frame)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 3 {
		t.Fatalf("parsed %d messages, want 3", len(messages))
	}
	listResponses := 0
	for _, message := range messages {
		if message.MessageBody.Tag == sml.MESSAGEGETLISTRESPONSE {
			listResponses++
		}
	}
	if listResponses == 0 {
		t.Fatal("no MESSAGEGETLISTRESPONSE found")
	}

	originalSettings := settings.Load
	t.Cleanup(func() { settings.Load = originalSettings })
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/node_data.json" {
			t.Errorf("unexpected endpoint: %s", r.URL.Path)
		}
		_, _ = w.Write(frame)
	}))
	defer server.Close()
	settings.Load.Service.Pulse.IP = strings.TrimPrefix(server.URL, "http://")
	settings.Load.Service.Pulse.Node = 1
	settings.Load.Service.Pulse.Version = "modern"
	if err := pulse.Initialize(); err != nil {
		t.Fatal(err)
	}
	result := GetData()
	if result.NodeValue.Total.Consume == "" || result.NodeValue.Total.Feed == "" {
		t.Fatalf("modern frame produced empty values: consume=%q feed=%q", result.NodeValue.Total.Consume, result.NodeValue.Total.Feed)
	}
}

func TestModernPulseRejectsLegacyMalformedCapture(t *testing.T) {
	hexData, err := os.ReadFile("testdata/node_data.hex")
	if err != nil {
		t.Fatal(err)
	}
	frame, err := hex.DecodeString(strings.TrimSpace(string(hexData)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parseSML(frame); err == nil {
		t.Fatal("strict SML parser accepted the malformed historical Pulse capture")
	}
}

func TestModernPulseSalvagesOnlyCRCValidGetListFromDamagedTransport(t *testing.T) {
	hexData, err := os.ReadFile("testdata/node_data_strict_valid.hex")
	if err != nil {
		t.Fatal(err)
	}
	frame, err := hex.DecodeString(strings.TrimSpace(string(hexData)))
	if err != nil {
		t.Fatal(err)
	}

	// Model the observed bridge defect: bytes in the opening control message
	// are zeroed, while the later GetList message and its message CRC survive.
	damaged := append([]byte(nil), frame...)
	damaged[13], damaged[14], damaged[15] = 0, 0, 0
	before := append([]byte(nil), damaged...)

	if _, err := sml.TransportParse(damaged); err == nil {
		t.Fatal("strict parser accepted the damaged transport frame")
	}
	messages, recovered, err := parseSMLWithRecovery(damaged)
	if err != nil {
		t.Fatalf("CRC-valid GetList message was not recovered: %v", err)
	}
	if !recovered {
		t.Fatal("expected recovery path to be reported")
	}
	if !bytes.Equal(damaged, before) {
		t.Fatal("recovery modified the captured frame")
	}
	if len(messages) != 1 || messages[0].MessageBody.Tag != sml.MESSAGEGETLISTRESPONSE {
		t.Fatalf("recovery returned %d messages, expected only the CRC-valid GetList response", len(messages))
	}
	result := dataFromMessages(messages)
	if !result.Valid() {
		t.Fatalf("recovered GetList response produced no values: %+v", result.NodeValue)
	}
}

func TestModernDataRetriesAfterParseError(t *testing.T) {
	hexData, err := os.ReadFile("testdata/node_data_strict_valid.hex")
	if err != nil {
		t.Fatal(err)
	}
	frame, err := hex.DecodeString(strings.TrimSpace(string(hexData)))
	if err != nil {
		t.Fatal(err)
	}

	originalSettings := settings.Load
	t.Cleanup(func() { settings.Load = originalSettings })
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			_, _ = w.Write([]byte("invalid"))
			return
		}
		_, _ = w.Write(frame)
	}))
	defer server.Close()
	settings.Load.Service.Pulse.IP = strings.TrimPrefix(server.URL, "http://")
	settings.Load.Service.Pulse.Node = 1
	settings.Load.Service.Pulse.Version = "modern"
	if err := pulse.Initialize(); err != nil {
		t.Fatal(err)
	}

	result := GetData()
	if requests != 2 {
		t.Fatalf("made %d requests, want 2", requests)
	}
	if !result.Valid() {
		t.Fatal("retry did not return valid data")
	}
}

func TestDataValid(t *testing.T) {
	if (Data{}).Valid() {
		t.Fatal("empty data reported as valid")
	}
	var result Data
	result.NodeValue.Current.Consume = "1 W"
	if !result.Valid() {
		t.Fatal("data with a measurement reported as invalid")
	}
}
