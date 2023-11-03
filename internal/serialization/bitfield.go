package serialization

func EncodeBitField(bitfield []uint8) []byte {
	lastElement := bitfield[len(bitfield)-1]
	var size int
	size = (int(lastElement) / 8) + 1

	binaryBitfield := make([]byte, size)

	for value := range bitfield {
		SetBit(binaryBitfield, int(bitfield[value]))
	}

	return binaryBitfield
}

func decodeBitField(binaryBitfield []byte) []uint8 {
	var bitfield []uint8
	size := len(binaryBitfield)

	for i := 0; i < size*8; i++ {
		if GetBit(binaryBitfield, i) {
			bitfield = append(bitfield, uint8(i))
		}
	}

	return bitfield
}

// // SetBit sets the bit at the specified position to 1 (starting at 0).
func SetBit(bitfield []byte, position int) {
	offset := int(position / 8)
	value := bitfield[offset]
	index := position - (8 * offset)

	mask := 1 << (7 - index)

	bitfield[offset] = value | uint8(mask)
}

// // GetBit returns true if the bit value at a given position in the bitfield is set to 1 (starting at 0)
func GetBit(bitfield []byte, position int) bool {
	offset := int(position / 8)
	value := bitfield[offset]
	index := position - (8 * offset)

	mask := uint8(1 << (7 - index))

	return (value & mask) == mask
}
