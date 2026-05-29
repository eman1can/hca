package awb

import (
	"log"

	"eman1can/br"
	"eman1can/hca"
)

var (
	headerMagic = uint(0x41465332)
)

type File struct {
	Version         uint
	OffsetSize      uint
	WaveIdAlignment uint
	TotalSubsongs   uint
	OffsetAlignment uint
	Subfiles        map[uint]*hca.File
}

func LoadAWB(data []byte, keycode uint64) *File {
	sf := br.InitBitReader(data)

	if br.Peek(sf, 32) == headerMagic {
		br.Skip(sf, 32)
	} else {
		log.Panicln("Invalid Header for AWB file")
	}

	file := File{
		Subfiles: make(map[uint]*hca.File),
	}

	version := br.ReadA(sf, 8)
	if version != 2 {
		log.Panicln("Invalid Version for AWB file", version)
	}

	file.OffsetSize = br.ReadA(sf, 8)
	file.WaveIdAlignment = br.ReadA(sf, 16)
	file.TotalSubsongs = br.ReadA(sf, 32)
	file.OffsetAlignment = br.ReadA(sf, 16)

	subkey := uint64(br.ReadA(sf, 16))
	if subkey != 0 {
		keycode = keycode * (((subkey) << 16) | (^subkey + 2))
	}

	var waveIds []uint
	for ix := uint(0); ix < file.TotalSubsongs; ix++ {
		waveId := br.ReadA(sf, 16)
		waveIds = append(waveIds, waveId)
	}

	var subfileOffsets []uint
	for ix := uint(0); ix < file.TotalSubsongs+1; ix++ {
		var fileOffset uint
		switch file.OffsetSize {
		case 0x04:
			fileOffset = br.ReadA(sf, 32)
		case 0x02:
			fileOffset = br.ReadA(sf, 16)
		}
		if ix != file.TotalSubsongs && fileOffset%file.OffsetAlignment != 0 {
			fileOffset += file.OffsetAlignment - (fileOffset % file.OffsetAlignment)
		}
		subfileOffsets = append(subfileOffsets, fileOffset)
	}

	for i, subfileNext := range subfileOffsets[1:] {
		if subfileNext > uint(len(data)) {
			log.Panicln("Invalid Subfile offset", subfileNext)
		}

		subfileOffset := subfileOffsets[i]

		subfileData := data[subfileOffset:subfileNext]

		waveId := waveIds[i]
		hcaFile := hca.LoadHCA(subfileData, keycode)
		file.Subfiles[waveId] = hcaFile
	}

	return &file
}
