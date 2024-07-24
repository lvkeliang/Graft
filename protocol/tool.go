package protocol

func IntToBytes(n int) []byte {
	return []byte{byte(n >> 8), byte(n)}
}

func BytesToInt(bytes []byte) int {
	return int(bytes[0])<<8 | int(bytes[1])
}
