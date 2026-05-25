//go:build darwin

package audio

import (
	"encoding/binary"
	"testing"
)

func TestUnsigned8ToSigned16LE(t *testing.T) {
	got := unsigned8ToSigned16LE([]byte{0, 128, 255})

	samples := []int16{
		int16(binary.LittleEndian.Uint16(got[0:2])),
		int16(binary.LittleEndian.Uint16(got[2:4])),
		int16(binary.LittleEndian.Uint16(got[4:6])),
	}
	want := []int16{-32768, 0, 32512}

	for i := range want {
		if samples[i] != want[i] {
			t.Fatalf("sample %d: got %d, want %d", i, samples[i], want[i])
		}
	}
}
