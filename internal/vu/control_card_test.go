package vu

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/way-platform/tachograph-go/internal/dd"
	ddv1 "github.com/way-platform/tachograph-go/proto/gen/go/wayplatform/connect/tachograph/dd/v1"
)

// These bytes are synthetic; no tachograph download is needed for the regressions.
func controlCardBytes() []byte {
	return append([]byte{3, 0x0f}, []byte("TEST000000001ABC")...)
}

func checkControlCard(t *testing.T, card *ddv1.FullCardNumber) {
	t.Helper()
	if diff := cmp.Diff(ddv1.EquipmentType_CONTROL_CARD, card.GetCardType()); diff != "" {
		t.Errorf("card type (-want +got): %s", diff)
	}
	id := card.GetOwnerIdentification()
	got := []string{id.GetOwnerIdentification().GetValue(), id.GetConsecutiveIndex().GetValue(), id.GetReplacementIndex().GetValue(), id.GetRenewalIndex().GetValue()}
	if diff := cmp.Diff([]string{"TEST000000001", "A", "B", "C"}, got); diff != "" {
		t.Errorf("owner (-want +got): %s", diff)
	}
}

func TestOverviewGen1ControlCard(t *testing.T) {
	// Two certificates, VIN, VRI, time, downloadable period, slots, download activity,
	// zero company locks, one control activity, and signature (Appendix 7 TREP 01).
	const idxDownload = 194 + 194 + 17 + 15 + 4 + 8 + 1
	const idxControls = idxDownload + 58 + 1
	input := make([]byte, idxControls+1+31+128)
	copy(input[idxDownload+4:], controlCardBytes())
	input[idxDownload+4] = 4 // Company card used for the download.
	input[idxControls] = 1
	copy(input[idxControls+1+5:], controlCardBytes())
	got, err := unmarshalOverviewGen1(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.GetControlActivities()) != 1 {
		t.Fatal("expected one control activity")
	}
	checkControlCard(t, got.GetControlActivities()[0].GetControlCardNumber())
	out, err := (MarshalOptions{}).MarshalOverviewGen1(got)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(input, out); diff != "" {
		t.Errorf("overview round-trip (-want +got): %s", diff)
	}
}

func TestOverviewGen2ControlCard(t *testing.T) {
	// Both Gen2 VU formats carry FullCardNumberAndGeneration; a Gen1 card can
	// occur in a Gen2 VU, so exercise both card generation protocol bytes.
	for _, generation := range []byte{1, 2} {
		record := make([]byte, 32)
		copy(record[5:], controlCardBytes())
		record[23] = generation
		input := append([]byte{0x0a, 0, 32, 0, 1}, record...)
		t.Run(string(rune('0'+generation)), func(t *testing.T) {
			t.Run("Gen2V1", func(t *testing.T) {
				got, consumed, err := parseControlActivityRecordArrayGen2V1(input, 0)
				if err != nil {
					t.Fatal(err)
				}
				if consumed != len(input) || len(got) != 1 {
					t.Fatal("incorrect record count or consumed length")
				}
				checkControlCard(t, got[0].GetControlCardNumberAndGeneration().GetFullCardNumber())
				out, err := marshalControlActivitiesGen2V1(dd.MarshalOptions{}, got)
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(record, out); diff != "" {
					t.Errorf("control round-trip (-want +got): %s", diff)
				}
			})
			t.Run("Gen2V2", func(t *testing.T) {
				got, consumed, err := parseControlActivityRecordArrayGen2V2(input, 0)
				if err != nil {
					t.Fatal(err)
				}
				if consumed != len(input) || len(got) != 1 {
					t.Fatal("incorrect record count or consumed length")
				}
				checkControlCard(t, got[0].GetControlCardNumberAndGeneration().GetFullCardNumber())
				out, err := marshalControlActivitiesGen2V2(dd.MarshalOptions{}, got)
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(record, out); diff != "" {
					t.Errorf("control round-trip (-want +got): %s", diff)
				}
			})
		})
	}
}
