package vu

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
	vuv1 "github.com/way-platform/tachograph-go/proto/gen/go/wayplatform/connect/tachograph/vu/v1"
)

func recordArrayTestBytes(size, count uint16, payload int) []byte {
	b := make([]byte, 5+payload)
	b[0] = 1
	binary.BigEndian.PutUint16(b[1:3], size)
	binary.BigEndian.PutUint16(b[3:5], count)
	return b
}

func recordArrayTestParsers() []struct {
	name  string
	parse func([]byte, int) error
} {
	return []struct {
		name  string
		parse func([]byte, int) error
	}{
		{"sizeOfRecordArray", func(b []byte, o int) error { _, e := sizeOfRecordArray(b, o); return e }},
		{"parseTimeRealRecordArray", func(b []byte, o int) error { _, _, e := parseTimeRealRecordArray(b, o); return e }},
		{"parseOdometerValueMidnightRecordArray", func(b []byte, o int) error { _, _, e := parseOdometerValueMidnightRecordArray(b, o); return e }},
		{"parseVuCardIWRecordArrayG2", func(b []byte, o int) error { _, _, e := parseVuCardIWRecordArrayG2(b, o); return e }},
		{"parseVuActivityDailyRecordArray", func(b []byte, o int) error { _, _, e := parseVuActivityDailyRecordArray(b, o); return e }},
		{"parseVuPlaceDailyWorkPeriodRecordArrayG2", func(b []byte, o int) error { _, _, e := parseVuPlaceDailyWorkPeriodRecordArrayG2(b, o); return e }},
		{"parseVuGNSSADRecordArray", func(b []byte, o int) error { _, _, e := parseVuGNSSADRecordArray(b, o); return e }},
		{"parseVuSpecificConditionRecordArray", func(b []byte, o int) error { _, _, e := parseVuSpecificConditionRecordArray(b, o); return e }},
		{"parseVuPlaceDailyWorkPeriodRecordArrayG2V2", func(b []byte, o int) error { _, _, e := parseVuPlaceDailyWorkPeriodRecordArrayG2V2(b, o); return e }},
		{"parseVuGNSSADRecordArrayG2", func(b []byte, o int) error { _, _, e := parseVuGNSSADRecordArrayG2(b, o); return e }},
		{"parseVuBorderCrossingRecordArray", func(b []byte, o int) error { _, _, e := parseVuBorderCrossingRecordArray(b, o); return e }},
		{"parseVuLoadUnloadRecordArray", func(b []byte, o int) error { _, _, e := parseVuLoadUnloadRecordArray(b, o); return e }},
		{"parseVuDetailedSpeedBlockRecordArray", func(b []byte, o int) error { _, _, e := parseVuDetailedSpeedBlockRecordArray(b, o); return e }},
		{"parseVuFaultRecordArrayGen2V1", func(b []byte, o int) error { _, _, e := parseVuFaultRecordArrayGen2V1(b, o); return e }},
		{"parseVuEventRecordArrayGen2V1", func(b []byte, o int) error { _, _, e := parseVuEventRecordArrayGen2V1(b, o); return e }},
		{"parseVuOverspeedControlDataRecordArrayGen2", func(b []byte, o int) error { _, _, e := parseVuOverspeedControlDataRecordArrayGen2(b, o); return e }},
		{"parseVuOverspeedEventRecordArrayGen2V1", func(b []byte, o int) error { _, _, e := parseVuOverspeedEventRecordArrayGen2V1(b, o); return e }},
		{"parseVuTimeAdjustmentRecordArrayGen2V1", func(b []byte, o int) error { _, _, e := parseVuTimeAdjustmentRecordArrayGen2V1(b, o); return e }},
		{"parseVuOverspeedControlDataRecordArrayGen2V2", func(b []byte, o int) error { _, _, e := parseVuOverspeedControlDataRecordArrayGen2V2(b, o); return e }},
		{"parseVuFaultRecordArrayGen2V2", func(b []byte, o int) error { _, _, e := parseVuFaultRecordArrayGen2V2(b, o); return e }},
		{"parseVuEventRecordArrayGen2V2", func(b []byte, o int) error { _, _, e := parseVuEventRecordArrayGen2V2(b, o); return e }},
		{"parseVuOverspeedEventRecordArrayGen2V2", func(b []byte, o int) error { _, _, e := parseVuOverspeedEventRecordArrayGen2V2(b, o); return e }},
		{"parseVuTimeAdjustmentRecordArrayGen2V2", func(b []byte, o int) error { _, _, e := parseVuTimeAdjustmentRecordArrayGen2V2(b, o); return e }},
		{"parseCertificateRecordArray", func(b []byte, o int) error { _, _, e := parseCertificateRecordArray(b, o); return e }},
		{"parseVehicleIdentificationNumberRecordArray", func(b []byte, o int) error { _, _, e := parseVehicleIdentificationNumberRecordArray(b, o); return e }},
		{"parseVehicleRegistrationIdentificationRecordArray", func(b []byte, o int) error {
			_, _, e := parseVehicleRegistrationIdentificationRecordArray(b, o)
			return e
		}},
		{"parseCurrentDateTimeRecordArray", func(b []byte, o int) error { _, _, e := parseCurrentDateTimeRecordArray(b, o); return e }},
		{"parseDownloadablePeriodRecordArray", func(b []byte, o int) error { _, _, e := parseDownloadablePeriodRecordArray(b, o); return e }},
		{"parseCardSlotsStatusRecordArray", func(b []byte, o int) error { _, _, _, e := parseCardSlotsStatusRecordArray(b, o); return e }},
		{"parseDownloadActivityDataRecordArrayGen2V1", func(b []byte, o int) error { _, _, e := parseDownloadActivityDataRecordArrayGen2V1(b, o); return e }},
		{"parseCompanyLocksRecordArrayGen2V1", func(b []byte, o int) error { _, _, e := parseCompanyLocksRecordArrayGen2V1(b, o); return e }},
		{"parseControlActivityRecordArrayGen2V1", func(b []byte, o int) error { _, _, e := parseControlActivityRecordArrayGen2V1(b, o); return e }},
		{"parseVehicleRegistrationNumberRecordArray", func(b []byte, o int) error { _, _, e := parseVehicleRegistrationNumberRecordArray(b, o); return e }},
		{"parseDownloadActivityDataRecordArrayGen2V2", func(b []byte, o int) error { _, _, e := parseDownloadActivityDataRecordArrayGen2V2(b, o); return e }},
		{"parseCompanyLocksRecordArrayGen2V2", func(b []byte, o int) error { _, _, e := parseCompanyLocksRecordArrayGen2V2(b, o); return e }},
		{"parseControlActivityRecordArrayGen2V2", func(b []byte, o int) error { _, _, e := parseControlActivityRecordArrayGen2V2(b, o); return e }},
		{"parseVuIdentificationRecordArrayGen2", func(b []byte, o int) error { _, _, e := parseVuIdentificationRecordArrayGen2(b, o); return e }},
		{"parseSensorPairedRecordArrayGen2", func(b []byte, o int) error { _, _, e := parseSensorPairedRecordArrayGen2(b, o); return e }},
		{"parseCalibrationRecordArrayGen2V1", func(b []byte, o int) error { _, _, e := parseCalibrationRecordArrayGen2V1(b, o); return e }},
		{"parseSensorExternalGNSSCoupledRecordArrayGen2", func(b []byte, o int) error { _, _, e := parseSensorExternalGNSSCoupledRecordArrayGen2(b, o); return e }},
		{"parseCalibrationRecordArrayGen2V2", func(b []byte, o int) error { _, _, e := parseCalibrationRecordArrayGen2V2(b, o); return e }},
		{"parseItsConsentRecordArrayGen2V2", func(b []byte, o int) error { _, _, e := parseItsConsentRecordArrayGen2V2(b, o); return e }},
		{"parsePowerSupplyInterruptionRecordArrayGen2V2", func(b []byte, o int) error { _, _, e := parsePowerSupplyInterruptionRecordArrayGen2V2(b, o); return e }},
	}
}

func TestRecordArrayMalformed(t *testing.T) {
	for _, parser := range recordArrayTestParsers() {
		t.Run(parser.name, func(t *testing.T) {
			for _, tc := range []struct {
				name   string
				data   []byte
				offset int
			}{
				{"nil", nil, 0}, {"short_header", []byte{1, 0, 4, 0}, 0},
				{"negative_offset", make([]byte, 5), -1}, {"past_end", make([]byte, 5), 6},
				{"max_offset", make([]byte, 5), math.MaxInt}, {"min_offset", make([]byte, 5), math.MinInt},
				{"max_dimensions", recordArrayTestBytes(65535, 65535, 0), 0},
				{"truncated_payload", recordArrayTestBytes(4, 1, 3), 0},
				{"truncated_count", recordArrayTestBytes(64, 2, 64), 0},
				{"zero_size_nonzero_count", recordArrayTestBytes(0, 1, 0), 0},
			} {
				t.Run(tc.name, func(t *testing.T) {
					before := bytes.Clone(tc.data)
					if err := parser.parse(tc.data, tc.offset); err == nil {
						t.Error("expected malformed RecordArray error")
					}
					if diff := cmp.Diff(before, tc.data); diff != "" {
						t.Errorf("input mutated: %s", diff)
					}
				})
			}
		})
	}
}

func TestRecordArrayValidBounds(t *testing.T) {
	for _, tc := range []struct {
		name        string
		size, count uint16
		payload     int
	}{
		{"empty", 0, 0, 0}, {"empty_fixed_size", 64, 0, 0},
		{"single", 4, 1, 4}, {"multiple", 64, 2, 128},
		{"max_size", 65535, 1, 65535}, {"max_count", 1, 65535, 65535},
	} {
		t.Run(tc.name, func(t *testing.T) {
			array := recordArrayTestBytes(tc.size, tc.count, tc.payload)
			data := append([]byte{9, 9, 9}, array...)
			data = append(data, 7, 7) // A following record must not be consumed.
			got, err := sizeOfRecordArray(data, 3)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(len(array), got); diff != "" {
				t.Errorf("span (-want +got): %s", diff)
			}
			_, size, count, header, err := parseRecordArrayHeader(data, 3)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff([]int{int(tc.size), int(tc.count), 5}, []int{int(size), int(count), header}); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestRecordArrayRawFileTruncated(t *testing.T) {
	for _, tag := range []byte{0x21, 0x22, 0x23, 0x24, 0x25, 0x31, 0x32, 0x33, 0x35} {
		for _, header := range [][]byte{recordArrayTestBytes(65535, 65535, 0), recordArrayTestBytes(4, 1, 3), recordArrayTestBytes(0, 1, 0)} {
			data := append([]byte{0x76, tag}, header...)
			if _, err := (UnmarshalOptions{}).UnmarshalRawVehicleUnitFile(data); err == nil {
				t.Errorf("tag %02x: expected error", tag)
			}
		}
	}
}

func TestSplitTransferValueInvalidSize(t *testing.T) {
	for _, size := range []int32{-1, 4, math.MaxInt32} {
		rec := &vuv1.RawVehicleUnitFile_Record{}
		rec.SetValue([]byte{1, 2, 3})
		rec.SetSignatureSize(size)
		if _, _, err := splitTransferValue(rec); err == nil {
			t.Errorf("size %d: expected error", size)
		}
	}
}

func FuzzRecordArrayBounds(f *testing.F) {
	for _, b := range [][]byte{nil, recordArrayTestBytes(0, 0, 0), recordArrayTestBytes(0, 1, 0), recordArrayTestBytes(4, 1, 3), recordArrayTestBytes(4, 1, 4), recordArrayTestBytes(65535, 65535, 0)} {
		f.Add(b, 0)
	}
	f.Add(make([]byte, 5), math.MaxInt)
	f.Add(make([]byte, 5), -1)
	f.Fuzz(func(t *testing.T, b []byte, offset int) {
		if len(b) > 65536 {
			t.Skip()
		}
		before := bytes.Clone(b)
		for _, p := range recordArrayTestParsers() {
			_ = p.parse(b, offset)
		}
		if diff := cmp.Diff(before, b); diff != "" {
			t.Fatalf("input mutated: %s", diff)
		}
	})
}
