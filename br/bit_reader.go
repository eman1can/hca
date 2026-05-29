package br

import (
	"encoding/binary"
	"log"
)

type BitReader struct {
	Data    []byte
	Size    uint
	Bit     uint
	Aligned bool
}

func InitBitReader(data []byte) *BitReader {
	return &BitReader{
		Data:    data,
		Size:    uint(len(data) * 8),
		Bit:     0,
		Aligned: true,
	}
}

func Skip(sf *BitReader, size uint) {
	sf.Bit += size
	sf.Aligned = sf.Bit%8 == 0
}

func Seek(sf *BitReader, offset uint) {
	offset = max(offset, 0)
	offset = min(offset, sf.Size)

	sf.Bit = offset
	sf.Aligned = sf.Bit%8 == 0
}

func ReadA(sf *BitReader, size uint) uint {
	if !sf.Aligned {
		log.Panicln("readA called on unaligned BitReader")
	}

	pos := sf.Bit / 8

	v := uint(0)
	switch size {
	case 8:
		v = uint(sf.Data[pos])
	case 16:
		v = uint(binary.LittleEndian.Uint16(sf.Data[pos : pos+2]))
	case 24:
		v = uint(sf.Data[pos]) | uint(sf.Data[pos+1])<<8 | uint(sf.Data[pos+2])<<16
	case 32:
		v = uint(binary.LittleEndian.Uint32(sf.Data[pos : pos+4]))
	default:
		log.Panicln("readA called on invalid size")
	}

	sf.Bit += size

	return v
}

func peek(sf *BitReader, size uint) uint {
	if size == 0 {
		return 0
	}

	bitPos := sf.Bit
	bitRem := bitPos & 0x07
	bitSize := sf.Size

	v := uint(0)
	if bitPos+size > bitSize {
		return v
	}

	bitOffset := bitRem + size
	bitsLeft := bitSize - bitPos

	pos := bitPos >> 3
	v = uint(sf.Data[pos])

	if bitsLeft >= 32 && bitOffset >= 25 {
		mask := []uint{0xFFFFFFFF, 0x7FFFFFFF, 0x3FFFFFFF, 0x1FFFFFFF, 0x0FFFFFFF, 0x07FFFFFF, 0x03FFFFFF, 0x01FFFFFF}

		v = (v << 8) | uint(sf.Data[pos+1])
		v = (v << 8) | uint(sf.Data[pos+2])
		v = (v << 8) | uint(sf.Data[pos+3])

		v &= mask[bitRem]
		return v >> (32 - bitRem - size)
	}

	if bitsLeft >= 24 && bitOffset >= 17 {
		mask := []uint{0xFFFFFF, 0x7FFFFF, 0x3FFFFF, 0x1FFFFF, 0x0FFFFF, 0x07FFFF, 0x03FFFF, 0x01FFFF}

		v = (v << 8) | uint(sf.Data[pos+1])
		v = (v << 8) | uint(sf.Data[pos+2])

		v &= mask[bitRem]
		return v >> (24 - bitRem - size)
	}

	if bitsLeft >= 16 && bitOffset >= 9 {
		mask := []uint{0xFFFF, 0x7FFF, 0x3FFF, 0x1FFF, 0x0FFF, 0x07FF, 0x03FF, 0x01FF}

		v = (v << 8) | uint(sf.Data[pos+1])

		v &= mask[bitRem]
		return v >> (16 - bitRem - size)
	}

	mask := []uint{0xFF, 0x7F, 0x3F, 0x1F, 0x0F, 0x07, 0x03, 0x01}

	v &= mask[bitRem]
	return v >> (8 - bitRem - size)
}

func Read(sf *BitReader, size uint) uint {
	v := peek(sf, size)
	sf.Bit += size
	sf.Aligned = sf.Bit%8 == 0
	return v
}
