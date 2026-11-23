package dsk

// readFAT12/writeFAT12 mirror ReadFAT/WriteFAT: the classic FAT12
// packing of two 12-bit cluster pointers into 3 bytes (entry N lives
// at byte offset N*3/2 - two bytes read/written always, with the
// odd/even cluster number choosing which nibble holds which half).
func readFAT12(fat []byte, cluster int) uint16 {
	p := cluster * 3 / 2
	val := uint16(fat[p]) | uint16(fat[p+1])<<8
	if cluster&1 != 0 {
		return (val >> 4) & 0x0FFF
	}
	return val & 0x0FFF
}

func writeFAT12(fat []byte, cluster int, value uint16) {
	p := cluster * 3 / 2
	if cluster&1 != 0 {
		fat[p] = (fat[p] & 0x0F) | byte((value&0x0F)<<4)
		fat[p+1] = byte((value >> 4) & 0xFF)
	} else {
		fat[p] = byte(value & 0xFF)
		fat[p+1] = (fat[p+1] & 0xF0) | byte((value>>8)&0x0F)
	}
}
