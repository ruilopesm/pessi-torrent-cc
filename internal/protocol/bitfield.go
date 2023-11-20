package protocol

func EncodeBitField(bitfield []uint16) []byte {
	lastElement := bitfield[len(bitfield)-1]
	size := (int(lastElement) / 8) + 1

	binaryBitfield := make([]byte, size)

	for value := range bitfield {
		SetBit(binaryBitfield, int(bitfield[value]))
	}

	return binaryBitfield
}

func DecodeBitField(binaryBitfield []byte) []uint16 {
	var bitfield []uint16
	size := len(binaryBitfield)

	for i := 0; i < size*8; i++ {
		if GetBit(binaryBitfield, i) {
			bitfield = append(bitfield, uint16(i))
		}
	}

	return bitfield
}

// Sets the bit value at a given position in the bitfield to 1 (starting at 0)
func SetBit(bitfield []byte, position int) {
	offset := position / 8
	value := bitfield[offset]
	index := position - (8 * offset)

	mask := 1 << (7 - index)

	bitfield[offset] = value | uint8(mask)
}

// Returns true if the bit value at a given position in the bitfield is set to 1 (starting at 0)
func GetBit(bitfield []byte, position int) bool {
	offset := position / 8
	value := bitfield[offset]
	index := position - (8 * offset)

	mask := uint8(1 << (7 - index))

	return (value & mask) == mask
}
