package protocol

type Bitfield = []uint8

func EncodeBitField(bitfield []bool) Bitfield {
	size := (int(len(bitfield)) / 8) + 1

	binaryBitfield := make([]uint8, size)

	for i, value := range bitfield {
		if value {
			SetBit(binaryBitfield, i)
		}
	}

	return binaryBitfield
}

func DecodeBitField(binaryBitfield Bitfield) []bool {
	var bitfield []bool
	size := len(binaryBitfield)

	for i := 0; i < size*8; i++ {
		bitfield = append(bitfield, GetBit(binaryBitfield, i))
	}

	return bitfield
}

func NewCheckedBitfield(size int) Bitfield {
	bools := make([]bool, size)
	for i := 0; i < size; i++ {
		bools[i] = true
	}

	return EncodeBitField(bools)
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
