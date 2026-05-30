package hca

import (
	"math"

	"github.com/eman1can/sound_decrypt/br"
	"github.com/eman1can/sound_decrypt/enum"
)

// HCA version constants used for version-dependent decode paths.
const (
	hcaVersionV200 uint = 0x0200
	hcaVersionV300 uint = 0x0300
)

// ---------------------------------------------------------------------------
// Lookup tables — initialised from IEEE-754 hex literals to match clhca.c
// ---------------------------------------------------------------------------

var dequantizerScalingTable = func() [64]float32 {
	hex := [64]uint32{
		0x342A8D26, 0x34633F89, 0x3497657D, 0x34C9B9BE, 0x35066491, 0x353311C4, 0x356E9910, 0x359EF532,
		0x35D3CCF1, 0x360D1ADF, 0x363C034A, 0x367A83B3, 0x36A6E595, 0x36DE60F5, 0x371426FF, 0x3745672A,
		0x37838359, 0x37AF3B79, 0x37E97C38, 0x381B8D3A, 0x384F4319, 0x388A14D5, 0x38B7FBF0, 0x38F5257D,
		0x3923520F, 0x39599D16, 0x3990FA4D, 0x39C12C4D, 0x3A00B1ED, 0x3A2B7A3A, 0x3A647B6D, 0x3A9837F0,
		0x3ACAD226, 0x3B071F62, 0x3B340AAF, 0x3B6FE4BA, 0x3B9FD228, 0x3BD4F35B, 0x3C0DDF04, 0x3C3D08A4,
		0x3C7BDFED, 0x3CA7CD94, 0x3CDF9613, 0x3D14F4F0, 0x3D467991, 0x3D843A29, 0x3DB02F0E, 0x3DEAC0C7,
		0x3E1C6573, 0x3E506334, 0x3E8AD4C6, 0x3EB8FBAF, 0x3EF67A41, 0x3F243516, 0x3F5ACB94, 0x3F91C3D3,
		0x3FC238D2, 0x400164D2, 0x402C6897, 0x4065B907, 0x40990B88, 0x40CBEC15, 0x4107DB35, 0x413504F3,
	}
	var t [64]float32
	for i, v := range hex {
		t[i] = math.Float32frombits(v)
	}
	return t
}()

var dequantizerRangeTable = func() [16]float32 {
	hex := [16]uint32{
		0x3F800000, 0x3F2AAAAB, 0x3ECCCCCD, 0x3E924925, 0x3E638E39, 0x3E3A2E8C, 0x3E1D89D9, 0x3E088889,
		0x3D842108, 0x3D020821, 0x3C810204, 0x3C008081, 0x3B804020, 0x3B002008, 0x3A801002, 0x3A000801,
	}
	var t [16]float32
	for i, v := range hex {
		t[i] = math.Float32frombits(v)
	}
	return t
}()

// curve/scale to quantized resolution
var invertTable = [66]byte{
	14, 14, 14, 14, 14, 14, 13, 13, 13, 13, 13, 13, 12, 12, 12, 12,
	12, 12, 11, 11, 11, 11, 11, 11, 10, 10, 10, 10, 10, 10, 10, 9,
	9, 9, 9, 9, 9, 8, 8, 8, 8, 8, 8, 7, 6, 6, 5, 4,
	4, 4, 3, 3, 3, 2, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1,
	1, 1,
}

// coded resolution to max bits read
var maxBitTable = [16]byte{
	0, 2, 3, 3, 4, 4, 4, 4, 5, 6, 7, 8, 9, 10, 11, 12,
}

// bits to skip after reading a code (prefix codebook adjustment)
var readBitTable = [128]byte{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 2, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	2, 2, 2, 2, 2, 2, 3, 3, 0, 0, 0, 0, 0, 0, 0, 0,
	2, 2, 3, 3, 3, 3, 3, 3, 0, 0, 0, 0, 0, 0, 0, 0,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 4, 4,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 4,
	3, 3, 3, 3, 3, 3, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	3, 3, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
}

// code to quantized spectrum value (plain float literals are exact)
var readValTable = [128]float32{
	+0, +0, +0, +0, +0, +0, +0, +0, +0, +0, +0, +0, +0, +0, +0, +0,
	+0, +0, +1, -1, +0, +0, +0, +0, +0, +0, +0, +0, +0, +0, +0, +0,
	+0, +0, +1, +1, -1, -1, +2, -2, +0, +0, +0, +0, +0, +0, +0, +0,
	+0, +0, +1, -1, +2, -2, +3, -3, +0, +0, +0, +0, +0, +0, +0, +0,
	+0, +0, +1, +1, -1, -1, +2, +2, -2, -2, +3, +3, -3, -3, +4, -4,
	+0, +0, +1, +1, -1, -1, +2, +2, -2, -2, +3, -3, +4, -4, +5, -5,
	+0, +0, +1, +1, -1, -1, +2, -2, +3, -3, +4, -4, +5, -5, +6, -6,
	+0, +0, +1, -1, +2, -2, +3, -3, +4, -4, +5, -5, +6, -6, +7, -7,
}

var scaleConversionTable = func() [128]float32 {
	hex := [128]uint32{
		0x00000000, 0x32A0B051, 0x32D61B5E, 0x330EA43A, 0x333E0F68, 0x337D3E0C, 0x33A8B6D5, 0x33E0CCDF,
		0x3415C3FF, 0x34478D75, 0x3484F1F6, 0x34B123F6, 0x34EC0719, 0x351D3EDA, 0x355184DF, 0x358B95C2,
		0x35B9FCD2, 0x35F7D0DF, 0x36251958, 0x365BFBB8, 0x36928E72, 0x36C346CD, 0x370218AF, 0x372D583F,
		0x3766F85B, 0x3799E046, 0x37CD078C, 0x3808980F, 0x38360094, 0x38728177, 0x38A18FAF, 0x38D744FD,
		0x390F6A81, 0x393F179A, 0x397E9E11, 0x39A9A15B, 0x39E2055B, 0x3A16942D, 0x3A48A2D8, 0x3A85AAC3,
		0x3AB21A32, 0x3AED4F30, 0x3B1E196E, 0x3B52A81E, 0x3B8C57CA, 0x3BBAFF5B, 0x3BF9295A, 0x3C25FED7,
		0x3C5D2D82, 0x3C935A2B, 0x3CC4563F, 0x3D02CD87, 0x3D2E4934, 0x3D68396A, 0x3D9AB62B, 0x3DCE248C,
		0x3E0955EE, 0x3E36FD92, 0x3E73D290, 0x3EA27043, 0x3ED87039, 0x3F1031DC, 0x3F40213B, 0x3F800000,
		0x3FAA8D26, 0x3FE33F89, 0x4017657D, 0x4049B9BE, 0x40866491, 0x40B311C4, 0x40EE9910, 0x411EF532,
		0x4153CCF1, 0x418D1ADF, 0x41BC034A, 0x41FA83B3, 0x4226E595, 0x425E60F5, 0x429426FF, 0x42C5672A,
		0x43038359, 0x432F3B79, 0x43697C38, 0x439B8D3A, 0x43CF4319, 0x440A14D5, 0x4437FBF0, 0x4475257D,
		0x44A3520F, 0x44D99D16, 0x4510FA4D, 0x45412C4D, 0x4580B1ED, 0x45AB7A3A, 0x45E47B6D, 0x461837F0,
		0x464AD226, 0x46871F62, 0x46B40AAF, 0x46EFE4BA, 0x471FD228, 0x4754F35B, 0x478DDF04, 0x47BD08A4,
		0x47FBDFED, 0x4827CD94, 0x485F9613, 0x4894F4F0, 0x48C67991, 0x49043A29, 0x49302F0E, 0x496AC0C7,
		0x499C6573, 0x49D06334, 0x4A0AD4C6, 0x4A38FBAF, 0x4A767A41, 0x4AA43516, 0x4ADACB94, 0x4B11C3D3,
		0x4B4238D2, 0x4B8164D2, 0x4BAC6897, 0x4BE5B907, 0x4C190B88, 0x4C4BEC15, 0x00000000, 0x00000000,
	}
	var t [128]float32
	for i, v := range hex {
		t[i] = math.Float32frombits(v)
	}
	return t
}()

var intensityRatioTable = func() [16]float32 {
	hex := [16]uint32{
		0x40000000, 0x3FEDB6DB, 0x3FDB6DB7, 0x3FC92492, 0x3FB6DB6E, 0x3FA49249, 0x3F924925, 0x3F800000,
		0x3F5B6DB7, 0x3F36DB6E, 0x3F124925, 0x3EDB6DB7, 0x3E924925, 0x3E124925, 0x00000000, 0x00000000,
	}
	var t [16]float32
	for i, v := range hex {
		t[i] = math.Float32frombits(v)
	}
	return t
}()

var imdctWindowTable = func() [128]float32 {
	hex := [128]uint32{
		0x3A3504F0, 0x3B0183B8, 0x3B70C538, 0x3BBB9268, 0x3C04A809, 0x3C308200, 0x3C61284C, 0x3C8B3F17,
		0x3CA83992, 0x3CC77FBD, 0x3CE91110, 0x3D0677CD, 0x3D198FC4, 0x3D2DD35C, 0x3D434643, 0x3D59ECC1,
		0x3D71CBA8, 0x3D85741E, 0x3D92A413, 0x3DA078B4, 0x3DAEF522, 0x3DBE1C9E, 0x3DCDF27B, 0x3DDE7A1D,
		0x3DEFB6ED, 0x3E00D62B, 0x3E0A2EDA, 0x3E13E72A, 0x3E1E00B1, 0x3E287CF2, 0x3E335D55, 0x3E3EA321,
		0x3E4A4F75, 0x3E56633F, 0x3E62DF37, 0x3E6FC3D1, 0x3E7D1138, 0x3E8563A2, 0x3E8C72B7, 0x3E93B561,
		0x3E9B2AEF, 0x3EA2D26F, 0x3EAAAAAB, 0x3EB2B222, 0x3EBAE706, 0x3EC34737, 0x3ECBD03D, 0x3ED47F46,
		0x3EDD5128, 0x3EE6425C, 0x3EEF4EFF, 0x3EF872D7, 0x3F00D4A9, 0x3F0576CA, 0x3F0A1D3B, 0x3F0EC548,
		0x3F136C25, 0x3F180EF2, 0x3F1CAAC2, 0x3F213CA2, 0x3F25C1A5, 0x3F2A36E7, 0x3F2E9998, 0x3F32E705,
		0xBF371C9E, 0xBF3B37FE, 0xBF3F36F2, 0xBF431780, 0xBF46D7E6, 0xBF4A76A4, 0xBF4DF27C, 0xBF514A6F,
		0xBF547DC5, 0xBF578C03, 0xBF5A74EE, 0xBF5D3887, 0xBF5FD707, 0xBF6250DA, 0xBF64A699, 0xBF66D908,
		0xBF68E90E, 0xBF6AD7B1, 0xBF6CA611, 0xBF6E5562, 0xBF6FE6E7, 0xBF715BEF, 0xBF72B5D1, 0xBF73F5E6,
		0xBF751D89, 0xBF762E13, 0xBF7728D7, 0xBF780F20, 0xBF78E234, 0xBF79A34C, 0xBF7A5397, 0xBF7AF439,
		0xBF7B8648, 0xBF7C0ACE, 0xBF7C82C8, 0xBF7CEF26, 0xBF7D50CB, 0xBF7DA88E, 0xBF7DF737, 0xBF7E3D86,
		0xBF7E7C2A, 0xBF7EB3CC, 0xBF7EE507, 0xBF7F106C, 0xBF7F3683, 0xBF7F57CA, 0xBF7F74B6, 0xBF7F8DB6,
		0xBF7FA32E, 0xBF7FB57B, 0xBF7FC4F6, 0xBF7FD1ED, 0xBF7FDCAD, 0xBF7FE579, 0xBF7FEC90, 0xBF7FF22E,
		0xBF7FF688, 0xBF7FF9D0, 0xBF7FFC32, 0xBF7FFDDA, 0xBF7FFEED, 0xBF7FFF8F, 0xBF7FFFDF, 0xBF7FFFFC,
	}
	var t [128]float32
	for i, v := range hex {
		t[i] = math.Float32frombits(v)
	}
	return t
}()

var sinTables = func() [7][64]float32 {
	hex := [7][64]uint32{
		{
			0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75,
			0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75,
			0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75,
			0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75,
			0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75,
			0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75,
			0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75,
			0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75, 0x3DA73D75,
		},
		{
			0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31,
			0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31,
			0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31,
			0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31,
			0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31,
			0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31,
			0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31,
			0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31, 0x3F7B14BE, 0x3F54DB31,
		},
		{
			0x3F7EC46D, 0x3F74FA0B, 0x3F61C598, 0x3F45E403, 0x3F7EC46D, 0x3F74FA0B, 0x3F61C598, 0x3F45E403,
			0x3F7EC46D, 0x3F74FA0B, 0x3F61C598, 0x3F45E403, 0x3F7EC46D, 0x3F74FA0B, 0x3F61C598, 0x3F45E403,
			0x3F7EC46D, 0x3F74FA0B, 0x3F61C598, 0x3F45E403, 0x3F7EC46D, 0x3F74FA0B, 0x3F61C598, 0x3F45E403,
			0x3F7EC46D, 0x3F74FA0B, 0x3F61C598, 0x3F45E403, 0x3F7EC46D, 0x3F74FA0B, 0x3F61C598, 0x3F45E403,
			0x3F7EC46D, 0x3F74FA0B, 0x3F61C598, 0x3F45E403, 0x3F7EC46D, 0x3F74FA0B, 0x3F61C598, 0x3F45E403,
			0x3F7EC46D, 0x3F74FA0B, 0x3F61C598, 0x3F45E403, 0x3F7EC46D, 0x3F74FA0B, 0x3F61C598, 0x3F45E403,
			0x3F7EC46D, 0x3F74FA0B, 0x3F61C598, 0x3F45E403, 0x3F7EC46D, 0x3F74FA0B, 0x3F61C598, 0x3F45E403,
			0x3F7EC46D, 0x3F74FA0B, 0x3F61C598, 0x3F45E403, 0x3F7EC46D, 0x3F74FA0B, 0x3F61C598, 0x3F45E403,
		},
		{
			0x3F7FB10F, 0x3F7D3AAC, 0x3F7853F8, 0x3F710908, 0x3F676BD8, 0x3F5B941A, 0x3F4D9F02, 0x3F3DAEF9,
			0x3F7FB10F, 0x3F7D3AAC, 0x3F7853F8, 0x3F710908, 0x3F676BD8, 0x3F5B941A, 0x3F4D9F02, 0x3F3DAEF9,
			0x3F7FB10F, 0x3F7D3AAC, 0x3F7853F8, 0x3F710908, 0x3F676BD8, 0x3F5B941A, 0x3F4D9F02, 0x3F3DAEF9,
			0x3F7FB10F, 0x3F7D3AAC, 0x3F7853F8, 0x3F710908, 0x3F676BD8, 0x3F5B941A, 0x3F4D9F02, 0x3F3DAEF9,
			0x3F7FB10F, 0x3F7D3AAC, 0x3F7853F8, 0x3F710908, 0x3F676BD8, 0x3F5B941A, 0x3F4D9F02, 0x3F3DAEF9,
			0x3F7FB10F, 0x3F7D3AAC, 0x3F7853F8, 0x3F710908, 0x3F676BD8, 0x3F5B941A, 0x3F4D9F02, 0x3F3DAEF9,
			0x3F7FB10F, 0x3F7D3AAC, 0x3F7853F8, 0x3F710908, 0x3F676BD8, 0x3F5B941A, 0x3F4D9F02, 0x3F3DAEF9,
			0x3F7FB10F, 0x3F7D3AAC, 0x3F7853F8, 0x3F710908, 0x3F676BD8, 0x3F5B941A, 0x3F4D9F02, 0x3F3DAEF9,
		},
		{
			0x3F7FEC43, 0x3F7F4E6D, 0x3F7E1324, 0x3F7C3B28, 0x3F79C79D, 0x3F76BA07, 0x3F731447, 0x3F6ED89E,
			0x3F6A09A7, 0x3F64AA59, 0x3F5EBE05, 0x3F584853, 0x3F514D3D, 0x3F49D112, 0x3F41D870, 0x3F396842,
			0x3F7FEC43, 0x3F7F4E6D, 0x3F7E1324, 0x3F7C3B28, 0x3F79C79D, 0x3F76BA07, 0x3F731447, 0x3F6ED89E,
			0x3F6A09A7, 0x3F64AA59, 0x3F5EBE05, 0x3F584853, 0x3F514D3D, 0x3F49D112, 0x3F41D870, 0x3F396842,
			0x3F7FEC43, 0x3F7F4E6D, 0x3F7E1324, 0x3F7C3B28, 0x3F79C79D, 0x3F76BA07, 0x3F731447, 0x3F6ED89E,
			0x3F6A09A7, 0x3F64AA59, 0x3F5EBE05, 0x3F584853, 0x3F514D3D, 0x3F49D112, 0x3F41D870, 0x3F396842,
			0x3F7FEC43, 0x3F7F4E6D, 0x3F7E1324, 0x3F7C3B28, 0x3F79C79D, 0x3F76BA07, 0x3F731447, 0x3F6ED89E,
			0x3F6A09A7, 0x3F64AA59, 0x3F5EBE05, 0x3F584853, 0x3F514D3D, 0x3F49D112, 0x3F41D870, 0x3F396842,
		},
		{
			0x3F7FFB11, 0x3F7FD397, 0x3F7F84AB, 0x3F7F0E58, 0x3F7E70B0, 0x3F7DABCC, 0x3F7CBFC9, 0x3F7BACCD,
			0x3F7A7302, 0x3F791298, 0x3F778BC5, 0x3F75DEC6, 0x3F740BDD, 0x3F721352, 0x3F6FF573, 0x3F6DB293,
			0x3F6B4B0C, 0x3F68BF3C, 0x3F660F88, 0x3F633C5A, 0x3F604621, 0x3F5D2D53, 0x3F59F26A, 0x3F5695E5,
			0x3F531849, 0x3F4F7A1F, 0x3F4BBBF8, 0x3F47DE65, 0x3F43E200, 0x3F3FC767, 0x3F3B8F3B, 0x3F373A23,
			0x3F7FFB11, 0x3F7FD397, 0x3F7F84AB, 0x3F7F0E58, 0x3F7E70B0, 0x3F7DABCC, 0x3F7CBFC9, 0x3F7BACCD,
			0x3F7A7302, 0x3F791298, 0x3F778BC5, 0x3F75DEC6, 0x3F740BDD, 0x3F721352, 0x3F6FF573, 0x3F6DB293,
			0x3F6B4B0C, 0x3F68BF3C, 0x3F660F88, 0x3F633C5A, 0x3F604621, 0x3F5D2D53, 0x3F59F26A, 0x3F5695E5,
			0x3F531849, 0x3F4F7A1F, 0x3F4BBBF8, 0x3F47DE65, 0x3F43E200, 0x3F3FC767, 0x3F3B8F3B, 0x3F373A23,
		},
		{
			0x3F7FFEC4, 0x3F7FF4E6, 0x3F7FE129, 0x3F7FC38F, 0x3F7F9C18, 0x3F7F6AC7, 0x3F7F2F9D, 0x3F7EEA9D,
			0x3F7E9BC9, 0x3F7E4323, 0x3F7DE0B1, 0x3F7D7474, 0x3F7CFE73, 0x3F7C7EB0, 0x3F7BF531, 0x3F7B61FC,
			0x3F7AC516, 0x3F7A1E84, 0x3F796E4E, 0x3F78B47B, 0x3F77F110, 0x3F772417, 0x3F764D97, 0x3F756D97,
			0x3F748422, 0x3F73913F, 0x3F7294F8, 0x3F718F57, 0x3F708066, 0x3F6F6830, 0x3F6E46BE, 0x3F6D1C1D,
			0x3F6BE858, 0x3F6AAB7B, 0x3F696591, 0x3F6816A8, 0x3F66BECC, 0x3F655E0B, 0x3F63F473, 0x3F628210,
			0x3F6106F2, 0x3F5F8327, 0x3F5DF6BE, 0x3F5C61C7, 0x3F5AC450, 0x3F591E6A, 0x3F577026, 0x3F55B993,
			0x3F53FAC3, 0x3F5233C6, 0x3F5064AF, 0x3F4E8D90, 0x3F4CAE79, 0x3F4AC77F, 0x3F48D8B3, 0x3F46E22A,
			0x3F44E3F5, 0x3F42DE29, 0x3F40D0DA, 0x3F3EBC1B, 0x3F3CA003, 0x3F3A7CA4, 0x3F385216, 0x3F36206C,
		},
	}
	var t [7][64]float32
	for i, row := range hex {
		for j, v := range row {
			t[i][j] = math.Float32frombits(v)
		}
	}
	return t
}()

var cosTables = func() [7][64]float32 {
	hex := [7][64]uint32{
		{
			0xBD0A8BD4, 0x3D0A8BD4, 0x3D0A8BD4, 0xBD0A8BD4, 0x3D0A8BD4, 0xBD0A8BD4, 0xBD0A8BD4, 0x3D0A8BD4,
			0x3D0A8BD4, 0xBD0A8BD4, 0xBD0A8BD4, 0x3D0A8BD4, 0xBD0A8BD4, 0x3D0A8BD4, 0x3D0A8BD4, 0xBD0A8BD4,
			0x3D0A8BD4, 0xBD0A8BD4, 0xBD0A8BD4, 0x3D0A8BD4, 0xBD0A8BD4, 0x3D0A8BD4, 0x3D0A8BD4, 0xBD0A8BD4,
			0xBD0A8BD4, 0x3D0A8BD4, 0x3D0A8BD4, 0xBD0A8BD4, 0x3D0A8BD4, 0xBD0A8BD4, 0xBD0A8BD4, 0x3D0A8BD4,
			0x3D0A8BD4, 0xBD0A8BD4, 0xBD0A8BD4, 0x3D0A8BD4, 0xBD0A8BD4, 0x3D0A8BD4, 0x3D0A8BD4, 0xBD0A8BD4,
			0xBD0A8BD4, 0x3D0A8BD4, 0x3D0A8BD4, 0xBD0A8BD4, 0x3D0A8BD4, 0xBD0A8BD4, 0xBD0A8BD4, 0x3D0A8BD4,
			0xBD0A8BD4, 0x3D0A8BD4, 0x3D0A8BD4, 0xBD0A8BD4, 0x3D0A8BD4, 0xBD0A8BD4, 0xBD0A8BD4, 0x3D0A8BD4,
			0x3D0A8BD4, 0xBD0A8BD4, 0xBD0A8BD4, 0x3D0A8BD4, 0xBD0A8BD4, 0x3D0A8BD4, 0x3D0A8BD4, 0xBD0A8BD4,
		},
		{
			0xBE47C5C2, 0xBF0E39DA, 0x3E47C5C2, 0x3F0E39DA, 0x3E47C5C2, 0x3F0E39DA, 0xBE47C5C2, 0xBF0E39DA,
			0x3E47C5C2, 0x3F0E39DA, 0xBE47C5C2, 0xBF0E39DA, 0xBE47C5C2, 0xBF0E39DA, 0x3E47C5C2, 0x3F0E39DA,
			0x3E47C5C2, 0x3F0E39DA, 0xBE47C5C2, 0xBF0E39DA, 0xBE47C5C2, 0xBF0E39DA, 0x3E47C5C2, 0x3F0E39DA,
			0xBE47C5C2, 0xBF0E39DA, 0x3E47C5C2, 0x3F0E39DA, 0x3E47C5C2, 0x3F0E39DA, 0xBE47C5C2, 0xBF0E39DA,
			0x3E47C5C2, 0x3F0E39DA, 0xBE47C5C2, 0xBF0E39DA, 0xBE47C5C2, 0xBF0E39DA, 0x3E47C5C2, 0x3F0E39DA,
			0xBE47C5C2, 0xBF0E39DA, 0x3E47C5C2, 0x3F0E39DA, 0x3E47C5C2, 0x3F0E39DA, 0xBE47C5C2, 0xBF0E39DA,
			0xBE47C5C2, 0xBF0E39DA, 0x3E47C5C2, 0x3F0E39DA, 0x3E47C5C2, 0x3F0E39DA, 0xBE47C5C2, 0xBF0E39DA,
			0x3E47C5C2, 0x3F0E39DA, 0xBE47C5C2, 0xBF0E39DA, 0xBE47C5C2, 0xBF0E39DA, 0x3E47C5C2, 0x3F0E39DA,
		},
		{
			0xBDC8BD36, 0xBE94A031, 0xBEF15AEA, 0xBF226799, 0x3DC8BD36, 0x3E94A031, 0x3EF15AEA, 0x3F226799,
			0x3DC8BD36, 0x3E94A031, 0x3EF15AEA, 0x3F226799, 0xBDC8BD36, 0xBE94A031, 0xBEF15AEA, 0xBF226799,
			0x3DC8BD36, 0x3E94A031, 0x3EF15AEA, 0x3F226799, 0xBDC8BD36, 0xBE94A031, 0xBEF15AEA, 0xBF226799,
			0xBDC8BD36, 0xBE94A031, 0xBEF15AEA, 0xBF226799, 0x3DC8BD36, 0x3E94A031, 0x3EF15AEA, 0x3F226799,
			0x3DC8BD36, 0x3E94A031, 0x3EF15AEA, 0x3F226799, 0xBDC8BD36, 0xBE94A031, 0xBEF15AEA, 0xBF226799,
			0xBDC8BD36, 0xBE94A031, 0xBEF15AEA, 0xBF226799, 0x3DC8BD36, 0x3E94A031, 0x3EF15AEA, 0x3F226799,
			0xBDC8BD36, 0xBE94A031, 0xBEF15AEA, 0xBF226799, 0x3DC8BD36, 0x3E94A031, 0x3EF15AEA, 0x3F226799,
			0x3DC8BD36, 0x3E94A031, 0x3EF15AEA, 0x3F226799, 0xBDC8BD36, 0xBE94A031, 0xBEF15AEA, 0xBF226799,
		},
		{
			0xBD48FB30, 0xBE164083, 0xBE78CFCC, 0xBEAC7CD4, 0xBEDAE880, 0xBF039C3D, 0xBF187FC0, 0xBF2BEB4A,
			0x3D48FB30, 0x3E164083, 0x3E78CFCC, 0x3EAC7CD4, 0x3EDAE880, 0x3F039C3D, 0x3F187FC0, 0x3F2BEB4A,
			0x3D48FB30, 0x3E164083, 0x3E78CFCC, 0x3EAC7CD4, 0x3EDAE880, 0x3F039C3D, 0x3F187FC0, 0x3F2BEB4A,
			0xBD48FB30, 0xBE164083, 0xBE78CFCC, 0xBEAC7CD4, 0xBEDAE880, 0xBF039C3D, 0xBF187FC0, 0xBF2BEB4A,
			0x3D48FB30, 0x3E164083, 0x3E78CFCC, 0x3EAC7CD4, 0x3EDAE880, 0x3F039C3D, 0x3F187FC0, 0x3F2BEB4A,
			0xBD48FB30, 0xBE164083, 0xBE78CFCC, 0xBEAC7CD4, 0xBEDAE880, 0xBF039C3D, 0xBF187FC0, 0xBF2BEB4A,
			0xBD48FB30, 0xBE164083, 0xBE78CFCC, 0xBEAC7CD4, 0xBEDAE880, 0xBF039C3D, 0xBF187FC0, 0xBF2BEB4A,
			0x3D48FB30, 0x3E164083, 0x3E78CFCC, 0x3EAC7CD4, 0x3EDAE880, 0x3F039C3D, 0x3F187FC0, 0x3F2BEB4A,
		},
		{
			0xBCC90AB0, 0xBD96A905, 0xBDFAB273, 0xBE2F10A2, 0xBE605C13, 0xBE888E93, 0xBEA09AE5, 0xBEB8442A,
			0xBECF7BCA, 0xBEE63375, 0xBEFC5D27, 0xBF08F59B, 0xBF13682A, 0xBF1D7FD1, 0xBF273656, 0xBF3085BB,
			0x3CC90AB0, 0x3D96A905, 0x3DFAB273, 0x3E2F10A2, 0x3E605C13, 0x3E888E93, 0x3EA09AE5, 0x3EB8442A,
			0x3ECF7BCA, 0x3EE63375, 0x3EFC5D27, 0x3F08F59B, 0x3F13682A, 0x3F1D7FD1, 0x3F273656, 0x3F3085BB,
			0x3CC90AB0, 0x3D96A905, 0x3DFAB273, 0x3E2F10A2, 0x3E605C13, 0x3E888E93, 0x3EA09AE5, 0x3EB8442A,
			0x3ECF7BCA, 0x3EE63375, 0x3EFC5D27, 0x3F08F59B, 0x3F13682A, 0x3F1D7FD1, 0x3F273656, 0x3F3085BB,
			0xBCC90AB0, 0xBD96A905, 0xBDFAB273, 0xBE2F10A2, 0xBE605C13, 0xBE888E93, 0xBEA09AE5, 0xBEB8442A,
			0xBECF7BCA, 0xBEE63375, 0xBEFC5D27, 0xBF08F59B, 0xBF13682A, 0xBF1D7FD1, 0xBF273656, 0xBF3085BB,
		},
		{
			0xBC490E90, 0xBD16C32C, 0xBD7B2B74, 0xBDAFB680, 0xBDE1BC2E, 0xBE09CF86, 0xBE22ABB6, 0xBE3B6ECF,
			0xBE541501, 0xBE6C9A7F, 0xBE827DC0, 0xBE8E9A22, 0xBE9AA086, 0xBEA68F12, 0xBEB263EF, 0xBEBE1D4A,
			0xBEC9B953, 0xBED53641, 0xBEE0924F, 0xBEEBCBBB, 0xBEF6E0CB, 0xBF00E7E4, 0xBF064B82, 0xBF0B9A6B,
			0xBF10D3CD, 0xBF15F6D9, 0xBF1B02C6, 0xBF1FF6CB, 0xBF24D225, 0xBF299415, 0xBF2E3BDE, 0xBF32C8C9,
			0x3C490E90, 0x3D16C32C, 0x3D7B2B74, 0x3DAFB680, 0x3DE1BC2E, 0x3E09CF86, 0x3E22ABB6, 0x3E3B6ECF,
			0x3E541501, 0x3E6C9A7F, 0x3E827DC0, 0x3E8E9A22, 0x3E9AA086, 0x3EA68F12, 0x3EB263EF, 0x3EBE1D4A,
			0x3EC9B953, 0x3ED53641, 0x3EE0924F, 0x3EEBCBBB, 0x3EF6E0CB, 0x3F00E7E4, 0x3F064B82, 0x3F0B9A6B,
			0x3F10D3CD, 0x3F15F6D9, 0x3F1B02C6, 0x3F1FF6CB, 0x3F24D225, 0x3F299415, 0x3F2E3BDE, 0x3F32C8C9,
		},
		{
			0xBBC90F88, 0xBC96C9B6, 0xBCFB49BA, 0xBD2FE007, 0xBD621469, 0xBD8A200A, 0xBDA3308C, 0xBDBC3AC3,
			0xBDD53DB9, 0xBDEE3876, 0xBE039502, 0xBE1008B7, 0xBE1C76DE, 0xBE28DEFC, 0xBE354098, 0xBE419B37,
			0xBE4DEE60, 0xBE5A3997, 0xBE667C66, 0xBE72B651, 0xBE7EE6E1, 0xBE8586CE, 0xBE8B9507, 0xBE919DDD,
			0xBE97A117, 0xBE9D9E78, 0xBEA395C5, 0xBEA986C4, 0xBEAF713A, 0xBEB554EC, 0xBEBB31A0, 0xBEC1071E,
			0xBEC6D529, 0xBECC9B8B, 0xBED25A09, 0xBED8106B, 0xBEDDBE79, 0xBEE363FA, 0xBEE900B7, 0xBEEE9479,
			0xBEF41F07, 0xBEF9A02D, 0xBEFF17B2, 0xBF0242B1, 0xBF04F484, 0xBF07A136, 0xBF0A48AD, 0xBF0CEAD0,
			0xBF0F8784, 0xBF121EB0, 0xBF14B039, 0xBF173C07, 0xBF19C200, 0xBF1C420C, 0xBF1EBC12, 0xBF212FF9,
			0xBF239DA9, 0xBF26050A, 0xBF286605, 0xBF2AC082, 0xBF2D1469, 0xBF2F61A5, 0xBF31A81D, 0xBF33E7BC,
		},
	}
	var t [7][64]float32
	for i, row := range hex {
		for j, v := range row {
			t[i][j] = math.Float32frombits(v)
		}
	}
	return t
}()

// ---------------------------------------------------------------------------
// Step 1 — Decrypt frame bytes using cipher substitution table
// ---------------------------------------------------------------------------

func decryptFrame(file *File, frameData []byte) {
	for i := range frameData {
		frameData[i] = file.CipherTable[frameData[i]]
	}
}

// ---------------------------------------------------------------------------
// Step 2a — Unpack scalefactors
// ---------------------------------------------------------------------------

func unpackScaleFactors(ch *StChannel, sf *br.BitReader, hfrGroupCount uint, version uint) bool {
	csCount := ch.CodedCount
	var extraCount uint

	if ch.Type == enum.StereoSecondary || hfrGroupCount == 0 || version <= hcaVersionV200 {
		extraCount = 0
	} else {
		extraCount = hfrGroupCount
		csCount += extraCount
		if csCount > SamplesPerSubframe {
			return false
		}
	}

	deltaBits := br.Read(sf, 3)

	if deltaBits >= 6 {
		// fixed scalefactors
		for i := uint(0); i < csCount; i++ {
			ch.ScaleFactors[i] = byte(br.Read(sf, 6))
		}
	} else if deltaBits > 0 {
		// delta scalefactors
		expectedDelta := byte((1 << deltaBits) - 1)
		value := byte(br.Read(sf, 6))
		ch.ScaleFactors[0] = value

		for i := uint(1); i < csCount; i++ {
			delta := byte(br.Read(sf, deltaBits))

			if delta == expectedDelta {
				value = byte(br.Read(sf, 6))
			} else {
				test := int(value) + int(delta) - int(expectedDelta>>1)
				if test < 0 || test >= 64 {
					return false
				}
				value = value - (expectedDelta >> 1) + delta
				value = value & 0x3F
			}
			ch.ScaleFactors[i] = value
		}
	} else {
		// no scalefactors
		for i := uint(0); i < SamplesPerSubframe; i++ {
			ch.ScaleFactors[i] = 0
		}
	}

	// set derived HFR scales for v3.0
	for i := uint(0); i < extraCount; i++ {
		ch.ScaleFactors[SamplesPerSubframe-1-i] = ch.ScaleFactors[csCount-1-i]
	}

	return true
}

// ---------------------------------------------------------------------------
// Step 2b — Unpack intensity (joint stereo R channel) or HFR scales
// ---------------------------------------------------------------------------

func unpackIntensity(ch *StChannel, sf *br.BitReader, hfrGroupCount uint, version uint) bool {
	if ch.Type == enum.StereoSecondary {
		if version <= hcaVersionV200 {
			// Peek first (non-consuming). Only consume and set if value < 15.
			// value==15 is a special marker meaning "keep previous intensity".
			value := byte(br.Peek(sf, 4))
			if value < 15 {
				br.Skip(sf, 4) // consume the peeked bits
				ch.Intensity[0] = value
				for i := uint(1); i < SubFrames; i++ {
					ch.Intensity[i] = byte(br.Read(sf, 4))
				}
			}
			// value==15: bit position unchanged, intensity keeps previous frame's values
		} else {
			// v3.0
			value := byte(br.Read(sf, 4)) // peek + skip
			if value < 15 {
				deltaBits := br.Read(sf, 2)
				ch.Intensity[0] = value

				if deltaBits == 3 {
					// fixed intensities
					for i := uint(1); i < SubFrames; i++ {
						ch.Intensity[i] = byte(br.Read(sf, 4))
					}
				} else {
					// delta intensities
					bmax := byte((2 << deltaBits) - 1)
					bits := deltaBits + 1
					for i := uint(1); i < SubFrames; i++ {
						delta := byte(br.Read(sf, bits))
						if delta == bmax {
							value = byte(br.Read(sf, 4))
						} else {
							newVal := int(value) - int(bmax>>1) + int(delta)
							if newVal < 0 || newVal > 15 {
								return false
							}
							value = byte(newVal)
						}
						ch.Intensity[i] = value
					}
				}
			} else {
				// value == 15: set all to 7
				for i := uint(0); i < SubFrames; i++ {
					ch.Intensity[i] = 7
				}
			}
		}
	} else {
		// read high frequency scalefactors (v3.0 uses derived values from unpackScaleFactors)
		if version <= hcaVersionV200 {
			hfrScaleOffset := SamplesPerSubframe - hfrGroupCount
			for i := uint(0); i < hfrGroupCount; i++ {
				ch.ScaleFactors[hfrScaleOffset+i] = byte(br.Read(sf, 6))
			}
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// Step 2c — Calculate resolution from scalefactors + noise level
// ---------------------------------------------------------------------------

func calculateResolution(ch *StChannel, packedNoiseLevel uint, athCurve []byte, minResolution uint, maxResolution uint) {
	crCount := ch.CodedCount
	noiseCount := uint(0)
	validCount := uint(0)

	for i := uint(0); i < crCount; i++ {
		var newResolution byte
		scalefactor := ch.ScaleFactors[i]

		if scalefactor > 0 {
			noiseLevel := int(athCurve[i]) + int((packedNoiseLevel+i)>>8)
			curvePosition := noiseLevel + 1 - int(5*uint(scalefactor)>>1)

			if curvePosition < 0 {
				newResolution = 15
			} else if curvePosition <= 65 {
				newResolution = invertTable[curvePosition]
			} else {
				newResolution = 0
			}

			if newResolution > byte(maxResolution) {
				newResolution = byte(maxResolution)
			} else if newResolution < byte(minResolution) {
				newResolution = byte(minResolution)
			}

			if newResolution < 1 {
				ch.Noises[noiseCount] = byte(i)
				noiseCount++
			} else {
				ch.Noises[SamplesPerSubframe-1-validCount] = byte(i)
				validCount++
			}
		}
		ch.Resolution[i] = newResolution
	}

	ch.NoiseCount = noiseCount
	ch.ValidCount = validCount

	for i := crCount; i < SamplesPerSubframe; i++ {
		ch.Resolution[i] = 0
	}
}

// ---------------------------------------------------------------------------
// Step 2d — Calculate gain (scaling factor per coefficient)
// ---------------------------------------------------------------------------

func calculateGain(ch *StChannel) {
	for i := uint(0); i < ch.CodedCount; i++ {
		ch.Gain[i] = dequantizerScalingTable[ch.ScaleFactors[i]] * dequantizerRangeTable[ch.Resolution[i]]
	}
}

// ---------------------------------------------------------------------------
// Step 3 — Dequantize spectral coefficients
// ---------------------------------------------------------------------------

func dequantizeCoefficients(ch *StChannel, sf *br.BitReader, subframe int) {
	ccCount := ch.CodedCount

	for i := uint(0); i < ccCount; i++ {
		resolution := ch.Resolution[i]
		bits := uint(maxBitTable[resolution])
		code := br.Read(sf, bits)

		var qc float32
		if resolution > 7 {
			// sign-magnitude form (lowest bit = sign)
			signedCode := int32(1-int32((code&1)<<1)) * int32(code>>1)
			if signedCode == 0 {
				// zero uses one fewer bit (no sign bit), undo the sign read
				br.Back(sf, 1)
			}
			qc = float32(signedCode)
		} else {
			// prefix codebook
			index := int(resolution<<4) + int(code)
			skip := int(readBitTable[index]) - int(bits)
			if skip >= 0 {
				br.Skip(sf, uint(skip))
			} else {
				br.Back(sf, uint(-skip))
			}
			qc = readValTable[index]
		}

		ch.Spectra[subframe][i] = ch.Gain[i] * qc
	}

	// zero rest of spectra
	for i := ccCount; i < SamplesPerSubframe; i++ {
		ch.Spectra[subframe][i] = 0
	}
}

// ---------------------------------------------------------------------------
// Step 4a — Reconstruct noise for resolution-0 coefficients
// ---------------------------------------------------------------------------

func reconstructNoise(ch *StChannel, minResolution uint, msStereo uint, random *uint, subframe int) {
	if minResolution > 0 {
		return
	}
	if ch.ValidCount == 0 || ch.NoiseCount == 0 {
		return
	}
	if msStereo != 0 && ch.Type != enum.StereoPrimary {
		return
	}

	for i := uint(0); i < ch.NoiseCount; i++ {
		*random = 0x343FD**random + 0x269EC3

		randomIndex := SamplesPerSubframe - ch.ValidCount + ((*random & 0x7FFF) * ch.ValidCount >> 15)

		noiseIndex := uint(ch.Noises[i])
		validIndex := uint(ch.Noises[randomIndex])

		sfNoise := int(ch.ScaleFactors[noiseIndex])
		sfValid := int(ch.ScaleFactors[validIndex])
		scIndex := sfNoise - sfValid + 62
		if scIndex < 0 {
			scIndex = 0
		}

		ch.Spectra[subframe][noiseIndex] = scaleConversionTable[scIndex] * ch.Spectra[subframe][validIndex]
	}
}

// ---------------------------------------------------------------------------
// Step 4b — Reconstruct high-frequency bands (spectral band replication)
// ---------------------------------------------------------------------------

func reconstructHighFrequency(ch *StChannel, hfrGroupCount uint, bandsPerHfrGroup uint, stereoBandCount uint, baseBandCount uint, totalBandCount uint, version uint, subframe int) {
	if bandsPerHfrGroup == 0 {
		return
	}
	if ch.Type == enum.StereoSecondary {
		return
	}

	startBand := int(stereoBandCount + baseBandCount)
	highband := startBand
	lowband := startBand - 1

	hfrScales := ch.ScaleFactors[SamplesPerSubframe-hfrGroupCount:]

	var groupLimit int
	if version <= hcaVersionV200 {
		groupLimit = int(hfrGroupCount)
	} else {
		groupLimit = int(hfrGroupCount) >> 1
	}

	for group := uint(0); group < hfrGroupCount; group++ {
		lowbandSub := 0
		if int(group) < groupLimit {
			lowbandSub = 1
		}

		for i := uint(0); i < bandsPerHfrGroup; i++ {
			if highband >= int(totalBandCount) || lowband < 0 {
				break
			}

			scIndex := int(hfrScales[group]) - int(ch.ScaleFactors[lowband]) + 63
			if scIndex < 0 {
				scIndex = 0
			}

			ch.Spectra[subframe][highband] = scaleConversionTable[scIndex] * ch.Spectra[subframe][lowband]

			highband++
			lowband -= lowbandSub
		}
	}

	// last spectrum coefficient is 0
	if highband > 0 {
		ch.Spectra[subframe][highband-1] = 0
	}
}

// ---------------------------------------------------------------------------
// Step 5a — Apply intensity stereo (restore L/R bands from primary + panning)
// ---------------------------------------------------------------------------

func applyIntensityStereo(chPair []StChannel, subframe int, baseBandCount uint, totalBandCount uint) {
	if chPair[0].Type != enum.StereoPrimary {
		return
	}

	ratioL := intensityRatioTable[chPair[1].Intensity[subframe]]
	ratioR := 2.0 - ratioL

	spL := chPair[0].Spectra[subframe]
	spR := chPair[1].Spectra[subframe]

	for band := baseBandCount; band < totalBandCount; band++ {
		coefL := spL[band] * ratioL
		coefR := spL[band] * ratioR
		spL[band] = coefL
		spR[band] = coefR
	}
}

// ---------------------------------------------------------------------------
// Step 5b — Apply mid-side stereo
// ---------------------------------------------------------------------------

func applyMsStereo(chPair []StChannel, msStereo uint, baseBandCount uint, totalBandCount uint, subframe int) {
	if msStereo == 0 {
		return
	}
	if chPair[0].Type != enum.StereoPrimary {
		return
	}

	const ratio = float32(0.70710676908493) // 0x3F3504F3

	spL := chPair[0].Spectra[subframe]
	spR := chPair[1].Spectra[subframe]

	for band := uint(0); band < baseBandCount; band++ {
		coefL := (spL[band] + spR[band]) * ratio
		coefR := (spL[band] - spR[band]) * ratio
		spL[band] = coefL
		spR[band] = coefR
	}
}

// ---------------------------------------------------------------------------
// Step 6 — IMDCT transform (DCT-IV + windowed overlap-add)
// ---------------------------------------------------------------------------

// imdctTransform applies a DCT-IV to ch.Spectra[subframe] and updates
// ch.Wave[subframe] via windowed overlap-add.
// Mirrors clhca.c HCAIMDCT_Transform exactly using index arithmetic.
func imdctTransform(ch *StChannel, subframe int) {
	const size = 128 // samplesPerSubframe
	const half = 64  // size / 2

	// We alternate between two buffers: spectra (A) and temp (B).
	// Use a pair of named slices and an "odd" flag to track which holds the result.
	a := ch.Spectra[subframe] // initially the input
	b := ch.Temp

	// ---------- Pre-rotation (butterfly sum/diff) ----------
	// count1: number of outer groups; count2: size of each group half
	count1 := 1
	count2 := half
	// src = a, dst = b initially
	srcIsA := true

	for i := 0; i < int(MdctBits); i++ {
		var src, dst []float32
		if srcIsA {
			src, dst = a, b
		} else {
			src, dst = b, a
		}

		srcIdx := 0
		for j := 0; j < count1; j++ {
			d1 := j * count2 * 2 // head of this group in dst
			d2 := d1 + count2    // tail half of this group
			for k := 0; k < count2; k++ {
				av := src[srcIdx]
				bv := src[srcIdx+1]
				srcIdx += 2
				dst[d1] = av + bv
				dst[d2] = av - bv
				d1++
				d2++
			}
		}

		srcIsA = !srcIsA
		count1 <<= 1
		count2 >>= 1
	}

	// After mdctBits=7 (odd) iterations, we swapped 7 times:
	// srcIsA started true, flipped 7 times → srcIsA = false → result in b (Temp).
	// Post-rotation reads from b, writes to a.

	// ---------- Post-rotation (twiddle butterfly) ----------
	count1 = half
	count2 = 1
	// After pre-rotation: srcIsA=false → src=b, dst=a
	// We don't flip srcIsA here; we just mirror the same alternation.
	postSrcIsA := false // result of pre-rotation is in b

	for i := 0; i < int(MdctBits); i++ {
		var src, dst []float32
		if postSrcIsA {
			src, dst = a, b
		} else {
			src, dst = b, a
		}

		sinTable := sinTables[i][:]
		cosTable := cosTables[i][:]
		tblIdx := 0

		for j := 0; j < count1; j++ {
			s1Base := j * count2 * 2
			s2Base := s1Base + count2
			d1Base := j * count2 * 2
			// d2 starts at end of this group's pair and decrements
			d2Base := d1Base + count2*2 - 1

			d1 := d1Base
			d2 := d2Base
			for k := 0; k < count2; k++ {
				av := src[s1Base+k]
				bv := src[s2Base+k]
				sin := sinTable[tblIdx]
				cos := cosTable[tblIdx]
				tblIdx++
				dst[d1] = av*sin - bv*cos
				dst[d2] = av*cos + bv*sin
				d1++
				d2--
			}
		}

		postSrcIsA = !postSrcIsA
		count1 >>= 1
		count2 <<= 1
	}

	// After mdctBits=7 (odd) post-rotation iterations starting from postSrcIsA=false:
	// flipped 7 times → postSrcIsA = true → result in a (Spectra[subframe]).
	// So the final DCT result is in ch.Spectra[subframe] = a.
	dct := a // == ch.Spectra[subframe]
	prev := ch.ImdctPrevious
	wave := ch.Wave[subframe]

	for i := 0; i < half; i++ {
		wave[i] = imdctWindowTable[i]*dct[i+half] + prev[i]
		wave[i+half] = imdctWindowTable[i+half]*dct[size-1-i] - prev[i+half]
		prev[i] = imdctWindowTable[size-1-i] * dct[half-i-1]
		prev[i+half] = imdctWindowTable[half-i-1] * dct[i]
	}

}

// DecodeFrame decodes a single HCA frame. frameData must be exactly file.FrameSize bytes.
// Returns false if the frame is invalid (bad sync or CRC).
func DecodeFrame(file *File, frameData []byte) bool {
	if len(frameData) < 2 {
		return false
	}

	// Check sync word (unencrypted, big-endian)
	sync := uint(frameData[0])<<8 | uint(frameData[1])
	if sync != 0xFFFF {
		return false
	}

	// CRC over entire frame (including the 2-byte checksum at the end; result should be 0)
	if checksum(frameData) {
		return false
	}

	// Decrypt in place
	decryptFrame(file, frameData)

	// Build a BitReader over the decrypted frame, skip 16-bit sync
	sf := br.InitBitReader(frameData)
	br.Skip(sf, 16)

	// Read frame-level noise parameters
	frameAcceptableNoiseLevel := br.Read(sf, 9)
	frameEvaluationBoundary := br.Read(sf, 7)
	packedNoiseLevel := (frameAcceptableNoiseLevel << 8) - frameEvaluationBoundary

	// Per-channel: unpack scalefactors, intensity, resolution, gain
	for ch := uint(0); ch < file.ChannelCount; ch++ {
		if !unpackScaleFactors(&file.Channel[ch], sf, file.HfrGroupCount, file.Version) {
			return false
		}
		if !unpackIntensity(&file.Channel[ch], sf, file.HfrGroupCount, file.Version) {
			return false
		}
		calculateResolution(&file.Channel[ch], packedNoiseLevel, file.AthCurve, file.MinResolution, file.MaxResolution)
		calculateGain(&file.Channel[ch])
	}

	// Per subframe: dequantize coefficients
	for subframe := 0; subframe < int(SubFrames); subframe++ {
		for ch := uint(0); ch < file.ChannelCount; ch++ {
			dequantizeCoefficients(&file.Channel[ch], sf, subframe)
		}
	}

	// Per subframe: transform pipeline
	for subframe := 0; subframe < int(SubFrames); subframe++ {
		for ch := uint(0); ch < file.ChannelCount; ch++ {
			reconstructNoise(&file.Channel[ch], file.MinResolution, file.MsStereo, &file.Random, subframe)
			reconstructHighFrequency(&file.Channel[ch], file.HfrGroupCount, file.BandsPerHfrGroup,
				file.StereoBandCount, file.BaseBandCount, file.TotalBandCount, file.Version, subframe)
		}

		for ch := uint(0); ch+1 < file.ChannelCount; ch++ {
			if file.StereoBandCount > 0 {
				applyIntensityStereo(file.Channel[ch:ch+2], subframe, file.BaseBandCount, file.TotalBandCount)
			}
			applyMsStereo(file.Channel[ch:ch+2], file.MsStereo, file.BaseBandCount, file.TotalBandCount, subframe)
		}

		for ch := uint(0); ch < file.ChannelCount; ch++ {
			imdctTransform(&file.Channel[ch], subframe)
		}
	}

	return true
}
