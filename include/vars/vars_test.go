package vars

import (
	"strings"
	"testing"

	"pulse2mqtt/include/version"
)

func TestMQTTClientIDContainsVersion(t *testing.T) {
	want := "go_Pulse2MQTT_v" + version.Version
	if got := MQTTClientID(); got != want {
		t.Fatalf("unexpected default MQTT client ID: got %q, want %q", got, want)
	}
	if !strings.HasSuffix(MQTTClientID(), version.Version) {
		t.Fatalf("client ID does not contain application version: %q", MQTTClientID())
	}
}
