package card

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func cyclicTestRecord(previous int, payload ...byte) []byte {
	b := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint16(b, uint16(previous))
	binary.BigEndian.PutUint16(b[2:], uint16(len(b)))
	copy(b[4:], payload)
	return b
}

func TestCyclicIteratorTraversal(t *testing.T) {
	older := cyclicTestRecord(0, 0x11, 0x22)
	newer := cyclicTestRecord(6, 0x33, 0x44)
	linear := append(bytes.Clone(older), newer...)
	// Rotate the complete ring by 2: the older header crosses the ring boundary.
	rotated := append(bytes.Clone(linear[2:]), linear[:2]...)
	// Payload crosses the boundary while its header remains contiguous.
	payloadWrap := append(bytes.Clone(older[5:]), older[:5]...)
	cycle := append(cyclicTestRecord(6, 1, 2), cyclicTestRecord(6, 3, 4)...)
	for _, tt := range []struct {
		name          string
		buffer        []byte
		start, oldest int
		positions     []int
		records       [][]byte
	}{
		{"single", older, 0, 0, []int{0}, [][]byte{older}},
		{"newest_to_oldest", linear, 6, 0, []int{6, 0}, [][]byte{newer, older}},
		{"header_wrap", rotated, 4, 10, []int{4, 10}, [][]byte{newer, older}},
		{"payload_wrap", payloadWrap, 1, 1, []int{1}, [][]byte{older}},
		{"stop_at_oldest", linear, 6, 6, []int{6}, [][]byte{newer}},
		{"cycle", cycle, 6, 99, []int{6, 0}, [][]byte{cycle[6:], cycle[:6]}},
		{"nil", nil, 0, 0, nil, nil}, {"empty", []byte{}, 0, 0, nil, nil},
		{"negative_index", older, -1, 0, nil, nil}, {"past_end", older, len(older), 0, nil, nil},
		{"max_index", older, math.MaxInt, 0, nil, nil}, {"min_index", older, math.MinInt, 0, nil, nil},
		{"short_buffer", []byte{0, 0, 0}, 0, 0, nil, nil},
		{"zero_length", make([]byte, 8), 0, 0, nil, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			before := bytes.Clone(tt.buffer)
			it := NewCyclicRecordIterator(tt.buffer, tt.start, tt.oldest)
			var positions []int
			var records [][]byte
			for it.Next() {
				b, p, n := it.Record()
				if n != len(b) {
					t.Fatalf("reported length %d, bytes %d", n, len(b))
				}
				positions = append(positions, p)
				records = append(records, b)
				if len(records) > 10 {
					t.Fatal("iterator failed to terminate")
				}
			}
			if err := it.Err(); err != nil {
				t.Fatal(err)
			}
			if it.Next() {
				t.Fatal("iterator resumed after termination")
			}
			if diff := cmp.Diff(tt.positions, positions); diff != "" {
				t.Errorf("positions (-want +got): %s", diff)
			}
			if diff := cmp.Diff(tt.records, records); diff != "" {
				t.Errorf("records (-want +got): %s", diff)
			}
			if diff := cmp.Diff(before, tt.buffer); diff != "" {
				t.Errorf("buffer mutated: %s", diff)
			}
		})
	}
}

func TestCyclicIteratorInvalidLengths(t *testing.T) {
	for _, tt := range []struct {
		name              string
		previous, current uint16
		oldest            int
	}{
		{"short_record", 0, 3, 0}, {"record_exceeds_ring", 0, 9, 0},
		{"previous_exceeds_ring", 9, 4, 4},
	} {
		t.Run(tt.name, func(t *testing.T) {
			b := make([]byte, 8)
			binary.BigEndian.PutUint16(b, tt.previous)
			binary.BigEndian.PutUint16(b[2:], tt.current)
			it := NewCyclicRecordIterator(b, 0, tt.oldest)
			for it.Next() {
			}
			if it.Err() == nil {
				t.Fatal("expected invalid length error")
			}
		})
	}
}

func TestCyclicIteratorRecordLimit(t *testing.T) {
	for _, count := range []int{4999, 5000, 5001} {
		b := make([]byte, 4*count)
		for i := 0; i < count; i++ {
			copy(b[i*4:], cyclicTestRecord(4))
		}
		it := NewCyclicRecordIterator(b, (count-1)*4, 0)
		got := 0
		for it.Next() {
			got++
		}
		want := min(count, 5000)
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("count %d (-want +got): %s", count, diff)
		}
		if (it.Err() != nil) != (count > 5000) {
			t.Errorf("count %d: unexpected error %v", count, it.Err())
		}
	}
}

func TestCyclicActivityChronologicalOrder(t *testing.T) {
	// Two valid 14-byte daily records with distinct dates; rotate so a header wraps.
	day := func(previous int, seconds uint32) []byte {
		b := cyclicTestRecord(previous, make([]byte, 10)...)
		binary.BigEndian.PutUint32(b[4:8], seconds)
		b[9] = 1                                   // BCD presence count 0001.
		binary.BigEndian.PutUint16(b[12:], 0x1801) // Driving at minute 1.
		return b
	}
	older, newer := day(0, 1577836800), day(14, 1577923200)
	linear := append(older, newer...)
	ring := append(bytes.Clone(linear[2:]), linear[:2]...)
	input := binary.BigEndian.AppendUint16(nil, 26)
	input = binary.BigEndian.AppendUint16(input, 12)
	input = append(input, ring...)
	activity, err := (UnmarshalOptions{}).unmarshalDriverActivityData(input)
	if err != nil {
		t.Fatal(err)
	}
	output, err := (MarshalOptions{}).MarshalDriverActivity(activity)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(input, output); diff != "" {
		t.Errorf("wrapped activity round-trip (-want +got): %s", diff)
	}
	got, err := (UnmarshalOptions{}).parseActivityRecordsWithIterator(ring, 12, 26)
	if err != nil {
		t.Fatal(err)
	}
	var dates []int64
	for _, record := range got {
		if !record.GetValid() {
			t.Fatal("valid synthetic daily record rejected")
		}
		dates = append(dates, record.GetActivityRecordDate().GetSeconds())
	}
	if diff := cmp.Diff([]int64{1577836800, 1577923200}, dates); diff != "" {
		t.Errorf("chronology (-want +got): %s", diff)
	}
}

func FuzzCyclicIterator(f *testing.F) {
	f.Add(cyclicTestRecord(0, 1, 2), 0, 0)
	f.Add([]byte{}, 0, 0)
	f.Add(cyclicTestRecord(0), math.MaxInt, -1)
	f.Fuzz(func(t *testing.T, b []byte, start, oldest int) {
		if len(b) > 4096 {
			t.Skip()
		}
		before := bytes.Clone(b)
		it := NewCyclicRecordIterator(b, start, oldest)
		count := 0
		for it.Next() {
			count++
			if count > min(len(b), 5000) {
				t.Fatal("iterator failed to terminate")
			}
			record, pos, n := it.Record()
			if pos < 0 || pos >= len(b) || n < 4 || n > len(b) || len(record) != n {
				t.Fatal("invalid record bounds")
			}
		}
		if diff := cmp.Diff(before, b); diff != "" {
			t.Fatalf("input mutated: %s", diff)
		}
	})
}
