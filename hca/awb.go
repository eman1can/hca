package hca

import (
	"bytes"
	"encoding/binary"
	"log"
)

var (
	AwbHeaderMagic = []byte{0x41, 0x46, 0x53, 0x32}
)

func readU8le(b []byte, off int32) int32 {
	return int32(b[off])
}

func readU16le(b []byte, off int32) int32 {
	return int32(binary.LittleEndian.Uint16(b[off : off+2]))
}

func readU32le(b []byte, off int32) int32 {
	return int32(binary.LittleEndian.Uint32(b[off : off+4]))
}

func loadAWB(data []byte) {
	if !bytes.Equal(data[0:4], AwbHeaderMagic) {
		log.Panicln("Invalid Header for AWB file")
	}

	offsetSize := readU8le(data, 0x05)
	waveIdAlignment := readU16le(data, 0x06)
	totalSubsongs := readU32le(data, 0x08)
	offsetAlignment := readU16le(data, 0x0C)
	subkey := readU16le(data, 0x0E)

	idTableOffset := int32(0x10)
	offsetTableOffset := idTableOffset + totalSubsongs*waveIdAlignment
	for ix := 0; ix < int(totalSubsongs); ix++ {
		waveIdOffset := idTableOffset + int32(ix)*waveIdAlignment
		waveId := readU16le(data, waveIdOffset)

		var subfileOffsets []int32
		for targetSubsong := range totalSubsongs {
			subfileOffset := offsetTableOffset + targetSubsong*offsetSize
			var fileOffset int32
			switch offsetSize {
			case 0x04:
				fileOffset = readU32le(data, subfileOffset)
			case 0x02:
				fileOffset = readU16le(data, subfileOffset)
			}

			subfileOffsets = append(subfileOffsets, fileOffset)
		}
	}

}
