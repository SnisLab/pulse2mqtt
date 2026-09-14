package data

import (
	"testing"

	sml "github.com/DjSni/go-sml"
)

func TestOctet2Obis(t *testing.T) {
	got := Octet2Obis(sml.OctetString{1, 0, 1, 8, 0, 255})
	if got != "1-0:1.8.0*255" {
		t.Fatalf("unexpected OBIS code: got %q", got)
	}
}

func TestListEntry2Float(t *testing.T) {
	tests := []struct {
		name   string
		value  int64
		scaler int8
		want   float64
	}{
		{name: "zero scaler uses unit scale", value: 12, scaler: 0, want: 12},
		{name: "positive scaler", value: 12, scaler: 2, want: 1200},
		{name: "negative scaler", value: 1234, scaler: -3, want: 1.234},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ListEntry2Float(sml.ListEntry{Scaler: tt.scaler, Value: sml.Value{DataInt: tt.value}}); got != tt.want {
				t.Fatalf("ListEntry2Float() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrintListEntryUpdatesKnownValues(t *testing.T) {
	original := DResult
	t.Cleanup(func() { DResult = original })

	PrintListEntry(sml.ListEntry{
		ObjName: sml.OctetString{1, 0, 1, 8, 0, 255},
		Unit:    0x1E,
		Scaler:  -3,
		Value:   sml.Value{Typ: sml.TYPEINTEGER, DataInt: 1234},
	})

	if got, want := DResult.NodeValue.Total.Consume, "0.0012 kWh"; got != want {
		t.Fatalf("unexpected total consumption: got %q, want %q", got, want)
	}
}
