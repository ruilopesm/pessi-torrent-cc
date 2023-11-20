package protocol

func EncodeBitField(bitfield []uint16) []uint8 {
	lastElement := bitfield[len(bitfield)-1]
	size := (int(lastElement) / 8) + 1

	binaryBitfield := make([]uint8, size)

	for value := range bitfield {
		SetBit(binaryBitfield, int(bitfield[value]))
	}

	return binaryBitfield
}

func DecodeBitField(binaryBitfield []uint8) []uint16 {
	var bitfield []uint16
	size := len(binaryBitfield)

	for i := 0; i < size*8; i++ {
		if GetBit(binaryBitfield, i) {
			bitfield = append(bitfield, uint16(i))
		}
	}

	return bitfield
}

func NewCheckedBitfield(size int) []uint16 {
	chunksAvailable := make([]uint16, size)
	for i := 0; i < size; i++ {
		chunksAvailable[i] = uint16(i)
	}

	return chunksAvailable
}

// Sets the bit value at a given position in the bitfield to 1 (starting at 0)
func SetBit(bitfield []uint8, position int) {
	offset := position / 8
	value := bitfield[offset]
	index := position - (8 * offset)

	mask := 1 << (7 - index)

	bitfield[offset] = value | uint8(mask)
}

// Returns true if the bit value at a given position in the bitfield is set to 1 (starting at 0)
func GetBit(bitfield []uint8, position int) bool {
	offset := position / 8
	value := bitfield[offset]
	index := position - (8 * offset)

	mask := uint8(1 << (7 - index))

	return (value & mask) == mask
}
