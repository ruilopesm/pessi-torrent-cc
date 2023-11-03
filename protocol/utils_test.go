package protocol

import (
	"testing"
)

func TestSetBitsAndGetBit(t *testing.T) {
	bitfield := make([]uint8, 10)

	setBit(bitfield, 0)
	if getBit(bitfield, 0) == false {
		t.Error("Bit in position 0 not set to 1")
	}

	setBit(bitfield, 5)
	if getBit(bitfield, 5) == false {
		t.Error("Bit in position 5 not set to 1")
	}

	setBit(bitfield, 10)
	if getBit(bitfield, 10) == false {
		t.Error("Bit in position 10 not set to 1")
	}
}
