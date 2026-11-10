package dd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	ddv1 "github.com/way-platform/tachograph-go/proto/gen/go/wayplatform/connect/tachograph/dd/v1"
)

func TestGen1EventFaultProtocolValues(t *testing.T) {
	for _, tt := range []struct {
		wire    byte
		want    ddv1.EventFaultType
		unknown int32
	}{
		{0, ddv1.EventFaultType_GENERAL_NO_FURTHER_DETAILS, 0},
		{4, ddv1.EventFaultType_GENERAL_DRIVING_WITHOUT_APPROPRIATE_CARD, 0},
		{7, ddv1.EventFaultType_GENERAL_OVER_SPEEDING, 0},
		{0x10, ddv1.EventFaultType_VU_SEC_NO_FURTHER_DETAILS, 0},
		{0x20, ddv1.EventFaultType_SENSOR_SEC_NO_FURTHER_DETAILS, 0},
		{0x30, ddv1.EventFaultType_FAULT_REC_EQ_NO_FURTHER_DETAILS, 0},
		{0x40, ddv1.EventFaultType_FAULT_CARD_NO_FURTHER_DETAILS, 0},
		{0x1f, 0, 0x1f}, {0x41, 0, 0x41}, {0x80, 0, 0x80}, {0xff, 0, 0xff},
	} {
		t.Run(fmt.Sprintf("type_%02x", tt.wire), func(t *testing.T) {
			got, raw := (UnmarshalOptions{}).parseEventFaultType(tt.wire)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("type (-want +got): %s", diff)
			}
			if diff := cmp.Diff(tt.unknown, raw); diff != "" {
				t.Errorf("unknown (-want +got): %s", diff)
			}
			out, err := (MarshalOptions{}).marshalEventFaultType(got, raw)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.wire, out); diff != "" {
				t.Errorf("wire (-want +got): %s", diff)
			}
		})
	}
	for _, tt := range []struct {
		wire    byte
		want    ddv1.EventFaultRecordPurpose
		unknown int32
	}{
		{0, ddv1.EventFaultRecordPurpose_TEN_MOST_RECENT, 0},
		{1, ddv1.EventFaultRecordPurpose_LONGEST_IN_LAST_10_DAYS, 0},
		{6, ddv1.EventFaultRecordPurpose_FIRST_AFTER_LAST_CALIBRATION, 0},
		{7, ddv1.EventFaultRecordPurpose_ACTIVE_OR_ONGOING, 0},
		{8, 0, 8}, {0x80, 0, 0x80}, {0xff, 0, 0xff},
	} {
		t.Run(fmt.Sprintf("purpose_%02x", tt.wire), func(t *testing.T) {
			got, raw := (UnmarshalOptions{}).parseEventFaultRecordPurpose(tt.wire)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("purpose (-want +got): %s", diff)
			}
			if diff := cmp.Diff(tt.unknown, raw); diff != "" {
				t.Errorf("unknown (-want +got): %s", diff)
			}
			out, err := (MarshalOptions{}).marshalEventFaultRecordPurpose(got, raw)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.wire, out); diff != "" {
				t.Errorf("wire (-want +got): %s", diff)
			}
		})
	}
}

func TestGen1EventFaultRecordsAllBytes(t *testing.T) {
	for _, preserve := range []bool{false, true} {
		opts := UnmarshalOptions{PreserveRawData: preserve}
		for _, tc := range []struct {
			name                     string
			size, cardStart, cardEnd int
			roundTrip                func([]byte) ([]byte, error)
		}{
			{"fault", 82, 10, 82, func(b []byte) ([]byte, error) {
				r, e := opts.UnmarshalVuFaultRecord(b)
				if e != nil {
					return nil, e
				}
				return (MarshalOptions{}).MarshalVuFaultRecord(r)
			}},
			{"event", 83, 10, 82, func(b []byte) ([]byte, error) {
				r, e := opts.UnmarshalVuEventRecord(b)
				if e != nil {
					return nil, e
				}
				return (MarshalOptions{}).MarshalVuEventRecord(r)
			}},
			{"overspeed", 31, 12, 30, func(b []byte) ([]byte, error) {
				r, e := opts.UnmarshalVuOverspeedEventRecord(b)
				if e != nil {
					return nil, e
				}
				return (MarshalOptions{}).MarshalVuOverspeedEventRecord(r)
			}},
		} {
			t.Run(fmt.Sprintf("%s/raw=%t", tc.name, preserve), func(t *testing.T) {
				for i := 0; i < 256; i++ {
					input := make([]byte, tc.size)
					input[0], input[1] = byte(i), byte(255-i)
					copy(input[tc.cardStart:tc.cardEnd], bytes.Repeat([]byte{0xff}, tc.cardEnd-tc.cardStart))
					original := bytes.Clone(input)
					out, err := tc.roundTrip(input)
					if err != nil {
						t.Fatalf("byte %02x: %v", i, err)
					}
					if diff := cmp.Diff(original, out); diff != "" {
						t.Fatalf("byte %02x round-trip (-want +got): %s", i, diff)
					}
					if diff := cmp.Diff(original, input); diff != "" {
						t.Fatalf("input mutated: %s", diff)
					}
				}
			})
		}
	}
}

func TestGen1EventFaultInvalidEncoding(t *testing.T) {
	for _, value := range []int32{0, 1, -1, 999} {
		if _, err := (MarshalOptions{}).marshalEventFaultType(ddv1.EventFaultType(value), 0); err == nil {
			t.Errorf("type %d: expected error", value)
		}
		if _, err := (MarshalOptions{}).marshalEventFaultRecordPurpose(ddv1.EventFaultRecordPurpose(value), 0); err == nil {
			t.Errorf("purpose %d: expected error", value)
		}
	}
	for _, raw := range []int32{-1, 256} {
		if _, err := (MarshalOptions{}).marshalEventFaultType(0, raw); err == nil {
			t.Errorf("raw type %d: expected error", raw)
		}
		if _, err := (MarshalOptions{}).marshalEventFaultRecordPurpose(0, raw); err == nil {
			t.Errorf("raw purpose %d: expected error", raw)
		}
	}
}
