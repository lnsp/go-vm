package vm

func toUint16(b bool) uint16 {
	if b {
		return 1
	}
	return 0
}
