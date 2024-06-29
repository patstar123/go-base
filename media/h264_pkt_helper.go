package media

import (
	"errors"
	"github.com/pion/rtp/codecs"
)

// copy from pion\rtp@v1.8.3\codecs\h264_packet.go
const (
	naluTypeBitmask = 0x1F
	fuaNALUType     = 28
	fubNALUType     = 29
	fuHeaderSize    = 2
	fuEndBitmask    = 0x40
)

type H264PacketHTChecker struct {
	codecs.H264Packet
}

func (c *H264PacketHTChecker) Unmarshal(packet []byte) ([]byte, error) {
	return nil, errors.New("not implement")
}

func (c *H264PacketHTChecker) IsPartitionHead(payload []byte) bool {
	return c.H264Packet.IsPartitionHead(payload)
}

func (c *H264PacketHTChecker) IsPartitionTail(marker bool, payload []byte) bool {
	if marker || len(payload) == 0 {
		return true
	}

	// NALU Types
	// https://tools.ietf.org/html/rfc6184#section-5.4
	naluType := payload[0] & naluTypeBitmask
	switch {
	// 1-23: single NAL unit
	// 24/25: STAP-A/B
	// 26/27: MTAP16/24
	// 28/29: FU-A/B
	case naluType > 0 && naluType < 27:
		return true
	case naluType == fuaNALUType || naluType == fubNALUType:
		if len(payload) < fuHeaderSize {
			return false
		}
		return payload[1]&fuEndBitmask != 0
	default:
		return false
	}
}
