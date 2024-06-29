package media

import "lx/meeting/base/utils"

// AVCodec 音视频编码格式
type AVCodec int

const (
	CodecUnknown AVCodec = iota
	CodecH264
	CodecH265
	CodecVP8
	CodecAV1

	CodecPCMA
	CodecPCMU
	CodecAAC
	CodecOpus
)

// PktFormat 数据包封包格式
type PktFormat int

const (
	FmtRaw       PktFormat = iota // origin codec format
	FmtHeadSize4                  // fix 4 bytes for payload size
	FmtAnnexB                     // for h264/h265
	FmtAvcc                       // for h264/h265
	FmtIvf                        // for vp8/AV1
	FmtOgg                        // for opus
	FmtPs                         // for h264/g711/...
	FmtRtp                        // for all av
	FmtPpt                        // pktpassthrough format for all av
)

//	FmtAnnexB: start code and h264 NALU
//	FmtIvf: ivfHead/ivfFrameHead/ivfFrameData for vp8/vp9 codec which refer to `ivfwriter.go`
//	FmtOgg: ogg(ID/Comment)Headers/oggPagePayload for opus codec which refer to `oggwriter.go`
//	FmtPpt: PassHead-or-RawData from track which is rtp packet usually for others codec(webrtc/.../pktpassthrough/pkt_passthrough.go)

type AVPacket struct {
	Codec AVCodec   // 数据编码格式
	Fmt   PktFormat // 数据包封装格式

	Data          []byte // 包数据
	RawData       []byte // 包数据的原始数据(可能为nil, rawData切片包含`data`切片)
	DataSize      int    // 数据包大小
	BufReferenced bool   // `data`和`rawData`切片是否属于引用; true: 切片归属数据产生者,消费者若须在其他协程使用则须拷贝; false: 数据归属消费者

	IsHead bool // 是否为数据流的头(可能是一个头+一个数据包(比如g711); 也可能其实一个头+后续数据包(比如opus/vp8))

	IsVideo    bool // 是否为视频
	IsKeyFrame bool // 是否为视频I帧

	Ts uint32 // 时间戳; rtp-fmt: 单位时钟频率, 其他: 单位毫秒,
}

func (p *AVPacket) Dereference() *AVPacket {
	if p.BufReferenced {
		// copy data for referenced buffer
		data, rawData := utils.CopyBuffer(p.Data, p.RawData)
		return &AVPacket{
			Codec:         p.Codec,
			Fmt:           p.Fmt,
			Data:          data,
			RawData:       rawData,
			DataSize:      len(data),
			BufReferenced: false,
			IsHead:        p.IsHead,
			IsVideo:       p.IsVideo,
			IsKeyFrame:    p.IsKeyFrame,
			Ts:            p.Ts,
		}
	} else {
		return p
	}
}
