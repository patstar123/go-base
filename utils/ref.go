package utils

func CopyBuffer(p []byte, rawData []byte) ([]byte /*p*/, []byte /*rawData*/) {
	buffer := make([]byte, 0, len(p))
	p2 := append(buffer, p...)
	rawData2 := rawData

	if rawData == nil {
		// do nothing
	} else if IsSliceAddressEqual(p, rawData) {
		rawData2 = p2
	} else {
		buffer2 := make([]byte, 0, len(rawData))
		rawData2 = append(buffer2, rawData...)
	}

	return p2, rawData2
}
