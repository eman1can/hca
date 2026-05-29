package hca

import (
	"errors"
	"log"
	"math"

	"github.com/eman1can/sound_decrypt/br"
	"github.com/eman1can/sound_decrypt/enum"
)

var (
	headerMagic = uint(0x48434100)
	headerMask  = uint(0x7F7F7F7F)

	fmtHeaderMagic  = uint(0x666D7400)
	compHeaderMagic = uint(0x636F6D70)
	decHeaderMagic  = uint(0x64656300)
	vbrHeaderMagic  = uint(0x76627200)
	athHeaderMagic  = uint(0x61746800)
	loopHeaderMagic = uint(0x6C6F6F70)
	ciphHeaderMagic = uint(0x63697068)
	rvaHeaderMagic  = uint(0x72766100)
	commHeaderMagic = uint(0x636F6D6D)
	padHeaderMagic  = uint(0x70616400)

	crcMaskTable = []uint{
		0x0000, 0x8005, 0x800F, 0x000A, 0x801B, 0x001E, 0x0014, 0x8011, 0x8033, 0x0036, 0x003C, 0x8039, 0x0028, 0x802D, 0x8027, 0x0022,
		0x8063, 0x0066, 0x006C, 0x8069, 0x0078, 0x807D, 0x8077, 0x0072, 0x0050, 0x8055, 0x805F, 0x005A, 0x804B, 0x004E, 0x0044, 0x8041,
		0x80C3, 0x00C6, 0x00CC, 0x80C9, 0x00D8, 0x80DD, 0x80D7, 0x00D2, 0x00F0, 0x80F5, 0x80FF, 0x00FA, 0x80EB, 0x00EE, 0x00E4, 0x80E1,
		0x00A0, 0x80A5, 0x80AF, 0x00AA, 0x80BB, 0x00BE, 0x00B4, 0x80B1, 0x8093, 0x0096, 0x009C, 0x8099, 0x0088, 0x808D, 0x8087, 0x0082,
		0x8183, 0x0186, 0x018C, 0x8189, 0x0198, 0x819D, 0x8197, 0x0192, 0x01B0, 0x81B5, 0x81BF, 0x01BA, 0x81AB, 0x01AE, 0x01A4, 0x81A1,
		0x01E0, 0x81E5, 0x81EF, 0x01EA, 0x81FB, 0x01FE, 0x01F4, 0x81F1, 0x81D3, 0x01D6, 0x01DC, 0x81D9, 0x01C8, 0x81CD, 0x81C7, 0x01C2,
		0x0140, 0x8145, 0x814F, 0x014A, 0x815B, 0x015E, 0x0154, 0x8151, 0x8173, 0x0176, 0x017C, 0x8179, 0x0168, 0x816D, 0x8167, 0x0162,
		0x8123, 0x0126, 0x012C, 0x8129, 0x0138, 0x813D, 0x8137, 0x0132, 0x0110, 0x8115, 0x811F, 0x011A, 0x810B, 0x010E, 0x0104, 0x8101,
		0x8303, 0x0306, 0x030C, 0x8309, 0x0318, 0x831D, 0x8317, 0x0312, 0x0330, 0x8335, 0x833F, 0x033A, 0x832B, 0x032E, 0x0324, 0x8321,
		0x0360, 0x8365, 0x836F, 0x036A, 0x837B, 0x037E, 0x0374, 0x8371, 0x8353, 0x0356, 0x035C, 0x8359, 0x0348, 0x834D, 0x8347, 0x0342,
		0x03C0, 0x83C5, 0x83CF, 0x03CA, 0x83DB, 0x03DE, 0x03D4, 0x83D1, 0x83F3, 0x03F6, 0x03FC, 0x83F9, 0x03E8, 0x83ED, 0x83E7, 0x03E2,
		0x83A3, 0x03A6, 0x03AC, 0x83A9, 0x03B8, 0x83BD, 0x83B7, 0x03B2, 0x0390, 0x8395, 0x839F, 0x039A, 0x838B, 0x038E, 0x0384, 0x8381,
		0x0280, 0x8285, 0x828F, 0x028A, 0x829B, 0x029E, 0x0294, 0x8291, 0x82B3, 0x02B6, 0x02BC, 0x82B9, 0x02A8, 0x82AD, 0x82A7, 0x02A2,
		0x82E3, 0x02E6, 0x02EC, 0x82E9, 0x02F8, 0x82FD, 0x82F7, 0x02F2, 0x02D0, 0x82D5, 0x82DF, 0x02DA, 0x82CB, 0x02CE, 0x02C4, 0x82C1,
		0x8243, 0x0246, 0x024C, 0x8249, 0x0258, 0x825D, 0x8257, 0x0252, 0x0270, 0x8275, 0x827F, 0x027A, 0x826B, 0x026E, 0x0264, 0x8261,
		0x0220, 0x8225, 0x822F, 0x022A, 0x823B, 0x023E, 0x0234, 0x8231, 0x8213, 0x0216, 0x021C, 0x8219, 0x0208, 0x820D, 0x8207, 0x0202,
	}

	minChannels   = uint(1)
	maxChannels   = uint(16)
	minSampleRate = uint(1)
	maxSampleRate = uint(0x7FFFFF)
	minFrameSize  = uint(0x8)
	maxFrameSize  = uint(0xFFFF)

	defaultRandom = uint(1)

	SubFrames          = uint(8)
	SamplesPerSubframe = uint(128)
	SamplesPerFrame    = SubFrames * SamplesPerSubframe
	MdctBits           = uint(7)
)

type StChannel struct {
	Type       enum.ChannelType
	CodedCount uint

	Intensity    []byte
	ScaleFactors []byte
	Resolution   []byte
	Noises       []byte
	NoiseCount   uint
	ValidCount   uint

	Gain    []float32
	Spectra [][]float32

	Temp          []float32
	Dct           []float32
	ImdctPrevious []float32

	Wave [][]float32
}

type File struct {
	Version    uint
	HeaderSize uint

	ChannelCount   uint
	SampleRate     uint
	FrameCount     uint
	EncoderDelay   uint
	EncoderPadding uint

	FrameSize        uint
	MinResolution    uint
	MaxResolution    uint
	TrackCount       uint
	ChannelConfig    uint
	StereoType       uint
	TotalBandCount   uint
	BaseBandCount    uint
	StereoBandCount  uint
	BandsPerHfrGroup uint
	MsStereo         uint

	VbrMaxFrameSize uint
	VbrNoiseLevel   uint

	AthType uint

	LoopStartFrame uint
	LoopEndFrame   uint
	LoopStartDelay uint
	LoopEndPadding uint
	LoopEnabled    uint

	EncryptionEnabled bool
	CiphType          uint

	RvaVolume float32

	CommentLen uint
	Comment    string

	HfrGroupCount uint
	AthCurve      []byte
	CipherTable   []byte

	Random  uint
	Channel []StChannel

	SampleCount     uint
	LoopStartSample uint
	LoopEndSample   uint

	// Samples holds the fully decoded PCM data: Samples[channel][sampleIndex].
	// Populated by LoadHCA after all frames are decoded.
	Samples [][]float32
}

// Calculates a CRC-16 Checksum over the given data
func checksum(data []byte) bool {
	crc := uint(0)
	for ix := uint(0); ix < uint(len(data)); ix++ {
		crc = ((crc << 8) ^ crcMaskTable[(crc>>8)^uint(data[ix])]) & 0xFFFF
	}
	return crc != 0
}

func peekMagic(sf *br.BitReader) uint {
	if !sf.Aligned {
		log.Panicln("Attempted to read magic from unaligned bit reader")
	}

	p := sf.Bit / 8
	v1 := int32(sf.Data[p])
	v2 := int32(sf.Data[p+1])
	v3 := int32(sf.Data[p+2])
	v4 := int32(sf.Data[p+3])
	v := (v1 << 24) | (v2 << 16) | (v3 << 8) | v4

	return uint(v) & headerMask
}

func headerCeil2(a uint, b uint) uint {
	if b < 1 {
		return 0
	}
	if a%b == 0 {
		return a / b
	}
	return a/b + (a % b)
}

func LoadHCA(data []byte, keycode uint64) (*File, error) {
	size := uint(len(data))
	sf := br.InitBitReader(data)

	if peekMagic(sf) != headerMagic {
		return nil, errors.New("invalid header magic")
	}
	br.Skip(sf, 32)

	file := File{}

	file.Version = br.ReadABE(sf, 16)
	if file.Version != 0x0200 && file.Version != 0x0300 {
		return nil, errors.New("invalid header version")
	}

	file.HeaderSize = br.ReadABE(sf, 16)
	if size < file.HeaderSize {
		return nil, errors.New("invalid header size")
	}

	if checksum(data[:file.HeaderSize]) {
		return nil, errors.New("invalid header checksum")
	}

	size -= 0x08

	if size >= 0x10 && peekMagic(sf) == fmtHeaderMagic {
		br.Skip(sf, 32)

		file.ChannelCount = br.ReadABE(sf, 8)
		file.SampleRate = br.ReadABE(sf, 24)
		file.FrameCount = br.ReadABE(sf, 32)
		file.EncoderDelay = br.ReadABE(sf, 16)
		file.EncoderPadding = br.ReadABE(sf, 16)

		if file.ChannelCount < minChannels || file.ChannelCount > maxChannels {
			return nil, errors.New("invalid channel count")
		}

		if file.FrameCount == 0 {
			return nil, errors.New("invalid frame count")
		}

		if file.SampleRate < minSampleRate || file.SampleRate > maxSampleRate {
			return nil, errors.New("invalid sample rate")
		}

		size -= 0x10
	} else {
		return nil, errors.New("invalid FMT header")
	}

	if size >= 0x10 && peekMagic(sf) == compHeaderMagic {
		br.Skip(sf, 32)

		file.FrameSize = br.ReadABE(sf, 16)
		file.MinResolution = br.ReadABE(sf, 8)
		file.MaxResolution = br.ReadABE(sf, 8)
		file.TrackCount = br.ReadABE(sf, 8)
		file.ChannelConfig = br.ReadABE(sf, 8)
		file.TotalBandCount = br.ReadABE(sf, 8)
		file.BaseBandCount = br.ReadABE(sf, 8)
		file.StereoBandCount = br.ReadABE(sf, 8)
		file.BandsPerHfrGroup = br.ReadABE(sf, 8)
		file.MsStereo = br.ReadABE(sf, 8)

		br.Skip(sf, 8)

		size -= 0x10
	} else if size >= 0x0C && peekMagic(sf) == decHeaderMagic {
		br.Skip(sf, 32)

		file.FrameSize = br.ReadABE(sf, 16)
		file.MinResolution = br.ReadABE(sf, 8)
		file.MaxResolution = br.ReadABE(sf, 8)
		file.TotalBandCount = br.ReadABE(sf, 8) + 1
		file.BaseBandCount = br.ReadABE(sf, 8) + 1
		file.TrackCount = br.Read(sf, 4)
		file.ChannelConfig = br.Read(sf, 4)
		file.StereoType = br.ReadABE(sf, 8)

		if file.StereoType == enum.Discrete {
			file.BaseBandCount = file.TotalBandCount
		}

		file.StereoBandCount = file.TotalBandCount - file.BaseBandCount
		file.BandsPerHfrGroup = 0

		size -= 0x0C
	} else {
		return nil, errors.New("invalid COMP header")
	}

	if size >= 0x08 && peekMagic(sf) == vbrHeaderMagic {
		br.Skip(sf, 32)

		file.VbrMaxFrameSize = br.ReadABE(sf, 16)
		file.VbrNoiseLevel = br.ReadABE(sf, 16)

		if file.FrameSize > 0 || file.VbrMaxFrameSize <= 8 || file.VbrMaxFrameSize > 0x1FF {
			return nil, errors.New("invalid vbr max frame size")
		}

		size -= 0x08
	} else {
		file.VbrMaxFrameSize = 0
		file.VbrNoiseLevel = 0
	}

	if size >= 0x06 && peekMagic(sf) == athHeaderMagic {
		br.Skip(sf, 32)
		file.AthType = br.ReadABE(sf, 16)

		size -= 0x06
	} else if file.Version < 2 {
		file.AthType = 1
	} else {
		file.AthType = 0
	}

	if size >= 0x10 && peekMagic(sf) == loopHeaderMagic {
		br.Skip(sf, 32)

		file.LoopStartFrame = br.ReadABE(sf, 32)
		file.LoopEndFrame = br.ReadABE(sf, 32)
		file.LoopStartDelay = br.ReadABE(sf, 16)
		file.LoopEndPadding = br.ReadABE(sf, 16)

		file.LoopEnabled = 1

		if file.LoopStartFrame < 0 || file.LoopStartFrame > file.LoopEndFrame && file.LoopEndFrame >= file.FrameCount {
			return nil, errors.New("invalid loop start frame")
		}

		size -= 0x10
	} else {
		file.LoopStartFrame = 0
		file.LoopEndFrame = 0
		file.LoopStartDelay = 0
		file.LoopEndPadding = 0

		file.LoopEnabled = 0
	}

	if size >= 0x06 && peekMagic(sf) == ciphHeaderMagic {
		br.Skip(sf, 32)

		file.CiphType = br.ReadABE(sf, 16)

		if file.CiphType != 0 && file.CiphType != 1 && file.CiphType != 56 {
			return nil, errors.New("invalid ciph type")
		}

		size -= 0x06
	} else {
		file.CiphType = 0
	}

	file.EncryptionEnabled = file.CiphType == 56

	if size >= 0x08 && peekMagic(sf) == rvaHeaderMagic {
		br.Skip(sf, 32)

		file.RvaVolume = math.Float32frombits(uint32(br.ReadABE(sf, 32)))

		size -= 0x08
	} else {
		file.RvaVolume = 1.0
	}

	if size >= 0x05 && peekMagic(sf) == commHeaderMagic {
		br.Skip(sf, 32)

		file.CommentLen = br.ReadABE(sf, 8)

		if file.CommentLen > size-8 {
			return nil, errors.New("invalid comment length")
		}

		var str []byte
		for i := 0; i < int(file.CommentLen); i++ {
			b := br.ReadABE(sf, 8)
			str = append(str, byte(b))
		}
		file.Comment = string(str)

		size -= 0x05 + file.CommentLen
	} else {
		file.CommentLen = 0
		file.Comment = ""
	}

	if size >= 0x04 && peekMagic(sf) == padHeaderMagic {
		size -= size - 0x02
	}

	if file.FrameSize < minFrameSize || file.FrameSize > maxFrameSize {
		return nil, errors.New("invalid frame size")
	}

	if file.Version < 2 {
		if file.MinResolution != 1 || file.MaxResolution != 15 {
			return nil, errors.New("invalid min or max resolution")
		}
	} else {
		if file.MinResolution > file.MaxResolution || file.MaxResolution > 15 {
			return nil, errors.New("invalid min or max resolution")
		}
	}

	file.TrackCount = max(file.TrackCount, 1)
	if file.TrackCount > file.ChannelCount {
		return nil, errors.New("invalid track count")
	}

	if file.TotalBandCount > SamplesPerSubframe || file.BaseBandCount > SamplesPerSubframe || file.StereoBandCount > SamplesPerSubframe || file.BaseBandCount+file.StereoBandCount > SamplesPerSubframe || file.BandsPerHfrGroup > SamplesPerSubframe {
		return nil, errors.New("invalid frame count")
	}

	file.HfrGroupCount = headerCeil2(file.TotalBandCount-file.BaseBandCount-file.StereoBandCount, file.BandsPerHfrGroup)

	file.AthCurve = AthInit(int(file.AthType), file.SampleRate)
	file.CipherTable = CipherInit(int(file.CiphType), keycode)
	file.Channel = ChannelInit(file.ChannelCount, file.ChannelConfig, file.TrackCount, file.BaseBandCount, file.StereoBandCount)

	file.Random = defaultRandom

	file.SampleCount = file.FrameCount*SamplesPerFrame - file.EncoderDelay - file.EncoderPadding
	file.LoopStartSample = file.LoopStartFrame*SamplesPerFrame - file.EncoderDelay + file.LoopStartDelay
	file.LoopEndSample = file.LoopEndFrame*SamplesPerFrame - file.EncoderDelay + (SamplesPerFrame - file.LoopEndPadding)

	// Allocate decoded sample storage: one slice per channel.
	file.Samples = make([][]float32, file.ChannelCount)
	for ch := uint(0); ch < file.ChannelCount; ch++ {
		file.Samples[ch] = make([]float32, file.SampleCount)
	}

	frameBuffer := make([]byte, file.FrameSize)
	sampleIdx := uint(0)

frameLoop:
	for frame := uint(0); frame < file.FrameCount; frame++ {
		frameOffset := file.HeaderSize + frame*file.FrameSize
		if uint(len(data)) < frameOffset+file.FrameSize {
			return nil, errors.New("invalid frame size")
		}
		copy(frameBuffer, sf.Data[frameOffset:frameOffset+file.FrameSize])

		if !DecodeFrame(&file, frameBuffer) {
			// Skip bad frames; leave corresponding output samples as zero.
			continue
		}

		for subframe := uint(0); subframe < SubFrames; subframe++ {
			for sample := uint(0); sample < SamplesPerSubframe; sample++ {
				absIdx := frame*SamplesPerFrame + subframe*SamplesPerSubframe + sample
				if absIdx < file.EncoderDelay {
					continue
				}
				if absIdx >= file.EncoderDelay+file.SampleCount {
					break frameLoop
				}
				for ch := uint(0); ch < file.ChannelCount; ch++ {
					file.Samples[ch][sampleIdx] = file.Channel[ch].Wave[subframe][sample]
				}
				sampleIdx++
			}
		}
	}

	return &file, nil
}
