package wav

import (
	"encoding/binary"
	"io"

	"eman1can/hca"
)

// WriteWAV writes the decoded samples from file as a standard 16-bit little-endian
// PCM WAV. All decoding has already been performed by LoadHCA; this function only
// converts the float32 samples in file.Samples to int16 PCM and wraps them in a
// RIFF/WAV container.
func WriteWAV(file *hca.File, w io.Writer) error {
	const bitsPerSample uint32 = 16
	channels := uint32(file.ChannelCount)
	sampleRate := uint32(file.SampleRate)
	blockAlign := uint16(uint32(channels) * bitsPerSample / 8)
	byteRate := sampleRate * uint32(blockAlign)
	dataSize := uint32(file.SampleCount) * uint32(blockAlign)

	// --- WAV header ---
	writeU32LE := func(v uint32) error { return binary.Write(w, binary.LittleEndian, v) }
	writeU16LE := func(v uint16) error { return binary.Write(w, binary.LittleEndian, v) }
	writeTag := func(s string) error { _, err := io.WriteString(w, s); return err }

	for _, step := range []error{
		writeTag("RIFF"),
		writeU32LE(36 + dataSize),
		writeTag("WAVEfmt "),
		writeU32LE(16),               // fmt chunk size
		writeU16LE(1),                // PCM format
		writeU16LE(uint16(channels)), // channel count
		writeU32LE(sampleRate),
		writeU32LE(byteRate),
		writeU16LE(blockAlign),
		writeU16LE(uint16(bitsPerSample)),
		writeTag("data"),
		writeU32LE(dataSize),
	} {
		if step != nil {
			return step
		}
	}

	// --- PCM samples ---
	// Samples are interleaved: [ch0_s0, ch1_s0, ch0_s1, ch1_s1, ...]
	for i := uint(0); i < file.SampleCount; i++ {
		for ch := uint(0); ch < file.ChannelCount; ch++ {
			s := file.Samples[ch][i] * file.RvaVolume
			if s > 1.0 {
				s = 1.0
			} else if s < -1.0 {
				s = -1.0
			}
			pcm := int16(s * 32767)
			if err := binary.Write(w, binary.LittleEndian, pcm); err != nil {
				return err
			}
		}
	}

	return nil
}
