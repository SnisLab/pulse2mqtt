package vars

import (
	"strings"
	"testing"

	"pulse2mqtt/include/version"
)

func TestDefaultMQTTClientIDContainsVersion(t *testing.T) {
	want := "go_Pulse2MQTT_v" + version.Version
	if Mqtt_cID != want {
		t.Fatalf("unexpected default MQTT client ID: got %q, want %q", Mqtt_cID, want)
	}
	if !strings.HasSuffix(Mqtt_cID, version.Version) {
		t.Fatalf("client ID does not contain application version: %q", Mqtt_cID)
	}
}
