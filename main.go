package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eman1can/sound_decrypt/acb"
	"github.com/eman1can/sound_decrypt/awb"
	"github.com/eman1can/sound_decrypt/wav"
)

func main() {
	// data, err := os.ReadFile("rrilgl")
	// if err != nil {
	// 	fmt.Println("Error reading file:", err)
	// 	return
	// }
	//
	// awbOffset := 13120
	// awbSize := 133153
	// acbOffset := 6560
	// acbSize := 6560
	//
	// acbData := make([]byte, acbSize)
	// awbData := make([]byte, awbSize)
	// copy(acbData, data[acbOffset:acbOffset+acbSize])
	// copy(awbData, data[awbOffset:awbOffset+awbSize])
	//
	// err = os.WriteFile("rrilgl.acb", acbData, 0666)
	// if err != nil {
	// 	fmt.Println("Error writing file:", err)
	// 	return
	// }
	// err = os.WriteFile("rrilgl.awb", awbData, 0666)
	// if err != nil {
	// 	fmt.Println("Error writing file:", err)
	// 	return
	// }

	acbData, err := os.ReadFile(filepath.Join("E:\\CodeProjects\\SIFAS\\elichika\\static\\sounds", "acb", "vo_na_m0201_0133.acb"))
	awbData, err := os.ReadFile(filepath.Join("E:\\CodeProjects\\SIFAS\\elichika\\static\\sounds", "awb", "vo_na_m0201_0133.awb"))

	// AWB Keycode for Love Live! School idol festival ALL STARS (Android)
	keycode := uint64(6498535309877346413)

	acbFile, err := acb.LoadACB(acbData)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading ACB:", err)
		os.Exit(1)
	}

	awbFile, err := awb.LoadAWB(awbData, keycode)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading AWB:", err)
		os.Exit(1)
	}

	for waveID, hcaFile := range awbFile.Subfiles {
		name, ok := acbFile.Names[uint16(waveID)]
		if !ok || name == "" {
			name = fmt.Sprintf("wave_%d", waveID)
		}

		// Sanitize filename
		name = strings.Map(func(r rune) rune {
			if strings.ContainsRune(`\/:*?"<>|`, r) {
				return '_'
			}
			return r
		}, name)

		outName := name + ".wav"

		f, err := os.Create(outName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating %s: %v\n", outName, err)
			continue
		}

		if err := wav.WriteWAV(hcaFile, f); err != nil {
			fmt.Fprintf(os.Stderr, "error writing %s: %v\n", outName, err)
			f.Close()
			continue
		}

		f.Close()
		fmt.Printf("wrote %s\n", outName)
	}
}
