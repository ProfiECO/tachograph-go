package dd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	ddv1 "github.com/way-platform/tachograph-go/proto/gen/go/wayplatform/connect/tachograph/dd/v1"
)

func TestFullCardNumberCardTypes(t *testing.T) {
	// Protocol card types are not protobuf enum numbers: CONTROL_CARD is 3 on wire, 5 in proto.
	for _, tt := range []struct {
		wire   byte
		kind   ddv1.EquipmentType
		driver bool
	}{
		{1, ddv1.EquipmentType_DRIVER_CARD, true},
		{2, ddv1.EquipmentType_WORKSHOP_CARD, false},
		{3, ddv1.EquipmentType_CONTROL_CARD, false},
		{4, ddv1.EquipmentType_COMPANY_CARD, false},
	} {
		for _, preserve := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/raw=%t", tt.kind, preserve), func(t *testing.T) {
				input := append([]byte{tt.wire, 0x0f}, []byte("TEST000000001ABC")...)
				got, err := (UnmarshalOptions{PreserveRawData: preserve}).UnmarshalFullCardNumber(input)
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(tt.kind, got.GetCardType()); diff != "" {
					t.Errorf("card type (-want +got): %s", diff)
				}
				var fields []string
				var want []string
				var display string
				if tt.driver {
					if got.GetOwnerIdentification() != nil {
						t.Fatal("unexpected owner identification")
					}
					id := got.GetDriverIdentification()
					fields = []string{id.GetDriverIdentificationNumber().GetValue(), id.GetCardReplacementIndex().GetValue(), id.GetCardRenewalIndex().GetValue()}
					want = []string{"TEST000000001A", "B", "C"}
					display = want[0]
				} else {
					if got.GetDriverIdentification() != nil {
						t.Fatal("unexpected driver identification")
					}
					id := got.GetOwnerIdentification()
					fields = []string{id.GetOwnerIdentification().GetValue(), id.GetConsecutiveIndex().GetValue(), id.GetReplacementIndex().GetValue(), id.GetRenewalIndex().GetValue()}
					want = []string{"TEST000000001", "A", "B", "C"}
					display = want[0]
				}
				if diff := cmp.Diff(want, fields); diff != "" {
					t.Errorf("identification (-want +got): %s", diff)
				}
				out, err := (MarshalOptions{}).MarshalFullCardNumber(got)
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(input, out); diff != "" {
					t.Errorf("round-trip (-want +got): %s", diff)
				}
				formatted, err := (MarshalOptions{}).MarshalFullCardNumberAsString(got, 16)
				if err != nil {
					t.Fatal(err)
				}
				if diff := cmp.Diff(display, string(formatted)); diff != "" {
					t.Errorf("format (-want +got): %s", diff)
				}
			})
		}
	}
}

func TestFullCardNumberControlCardPadding(t *testing.T) {
	for _, padding := range []byte{0, ' '} {
		input := append([]byte{3, 0x0f}, bytes.Repeat([]byte{padding}, 16)...)
		copy(input[2:], "TEST")
		copy(input[15:], "ABC")
		got, err := (UnmarshalOptions{PreserveRawData: true}).UnmarshalFullCardNumber(input)
		if err != nil {
			t.Fatal(err)
		}
		out, err := (MarshalOptions{}).MarshalFullCardNumber(got)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(input, out); diff != "" {
			t.Errorf("padding round-trip (-want +got): %s", diff)
		}
	}
}

func TestFullCardNumberInvalidAndAbsent(t *testing.T) {
	for _, size := range []int{0, 17, 19} {
		if _, err := (UnmarshalOptions{}).UnmarshalFullCardNumber(make([]byte, size)); err == nil {
			t.Errorf("length %d: expected error", size)
		}
	}
	for _, wire := range []byte{0, 0xff, 0xfe} {
		input := bytes.Repeat([]byte{wire}, 18)
		got, err := (UnmarshalOptions{PreserveRawData: true}).UnmarshalFullCardNumber(input)
		if err != nil {
			t.Fatal(err)
		}
		if got.HasDriverIdentification() || got.HasOwnerIdentification() {
			t.Errorf("absent/unknown type %x has identification", wire)
		}
		if diff := cmp.Diff(input, got.GetRawData()); diff != "" {
			t.Errorf("raw data (-want +got): %s", diff)
		}
	}
	// Recognized equipment that is not a card remains unsupported.
	if _, err := (UnmarshalOptions{}).UnmarshalFullCardNumber(append([]byte{6}, make([]byte, 17)...)); err == nil {
		t.Error("vehicle unit: expected error")
	}
}

func TestMarshalFullCardNumberControlCard(t *testing.T) {
	// Encode independently constructed fields, including over a blank raw canvas,
	// so a raw-data passthrough cannot hide a missing owner-identification branch.
	for _, raw := range [][]byte{nil, make([]byte, 18)} {
		card := &ddv1.FullCardNumber{}
		card.SetCardType(ddv1.EquipmentType_CONTROL_CARD)
		card.SetCardIssuingMemberState(ddv1.NationNumeric(0x0f)) // Preserve baseline nation encoding.
		if raw != nil {
			card.SetRawData(raw)
		}
		owner := &ddv1.OwnerIdentification{}
		owner.SetOwnerIdentification(NewIa5StringValue(13, "TEST000000001"))
		owner.SetConsecutiveIndex(NewIa5StringValue(1, "A"))
		owner.SetReplacementIndex(NewIa5StringValue(1, "B"))
		owner.SetRenewalIndex(NewIa5StringValue(1, "C"))
		card.SetOwnerIdentification(owner)
		got, err := (MarshalOptions{}).MarshalFullCardNumber(card)
		if err != nil {
			t.Fatal(err)
		}
		want := append([]byte{3, 0x0f}, []byte("TEST000000001ABC")...)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("encoded card (-want +got): %s", diff)
		}
		if raw != nil {
			if diff := cmp.Diff(make([]byte, 18), raw); diff != "" {
				t.Errorf("input canvas mutated (-want +got): %s", diff)
			}
		}
	}
}
