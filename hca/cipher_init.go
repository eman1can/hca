package hca

func CipherCreateTable(key int) []int {
	table := make([]int, 16)
	mul := ((key & 1) << 3) | 5
	add := (key & 0xE) | 1

	key >>= 4
	for ix := 0; ix < 16; ix++ {
		key = (key*mul + add) & 0xF
		table[ix] = key
	}

	return table
}

func CipherInit(cipherType int, keycode uint64) []byte {
	if cipherType == 56 && keycode == 0 {
		cipherType = 0
	}

	data := make([]byte, 256)

	if cipherType == 1 {
		mul := 13
		add := 11

		v := 0
		for ix := 0; ix < 256-1; ix++ {
			v = (v*mul + add) & 0xFF
			if v == 0 || v == 0xFF {
				v = (v*mul + add) & 0xFF
			}
			data[ix] = byte(v)
		}

		data[0] = 0
		data[0xFF] = 0xFF
	} else if cipherType == 56 {
		kc := make([]byte, 8)
		seed := make([]byte, 16)
		base := make([]int, 256)

		if keycode != 0 {
			keycode--
		}

		for r := 0; r < 7; r++ {
			kc[r] = byte(keycode & 0xFF)
			keycode >>= 8
		}

		seed[0x00] = kc[1]
		seed[0x01] = kc[1] ^ kc[6]
		seed[0x02] = kc[2] ^ kc[3]
		seed[0x03] = kc[2]
		seed[0x04] = kc[2] ^ kc[1]
		seed[0x05] = kc[3] ^ kc[4]
		seed[0x06] = kc[3]
		seed[0x07] = kc[3] ^ kc[2]
		seed[0x08] = kc[4] ^ kc[5]
		seed[0x09] = kc[4]
		seed[0x0A] = kc[4] ^ kc[3]
		seed[0x0B] = kc[5] ^ kc[6]
		seed[0x0C] = kc[5]
		seed[0x0D] = kc[5] ^ kc[4]
		seed[0x0E] = kc[6] ^ kc[1]
		seed[0x0F] = kc[6]

		baseR := CipherCreateTable(int(kc[0]))
		for r := 0; r < 16; r++ {
			baseC := CipherCreateTable(int(seed[r]))
			nb := baseR[r] << 4
			for c := 0; c < 16; c++ {
				base[r*16+c] = nb | baseC[c]
			}
		}

		x := 0
		pos := 1
		for ix := 0; ix < 256; ix++ {
			x = (x + 17) & 0xFF
			if base[x] != 0 && base[x] != 0xFF {
				data[pos] = byte(base[x])
				pos++
			}
		}
		data[0] = 0
		data[0xFF] = 0xFF
	}

	return data
}
