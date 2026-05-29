package main

import (
	"fmt"
	"os"

	"eman1can/awb"
	"eman1can/br"
)

func main() {
	data, err := os.ReadFile("filename.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// AWB Keycode for Love Live! School Idol Festival 2 MIRACLE LIVE (Android)
	// keycode := uint64(5067530812966687744)

	// AWB Keycode for Love Live! School idol festival ALL STARS (Android)
	keycode := uint64(5067530812966687744)

	awbOffset := 13120
	awbSize := 133153
	acbOffset := 6560
	acbSize := 6560

	acbData := make([]byte, acbSize)
	awbData := make([]byte, awbSize)
	copy(acbData[acbOffset:], data)
	copy(awbData[awbOffset:], data)

	sf := br.InitBitReader(awbData)
	awb := awb.LoadAWB(sf, keycode)
}
