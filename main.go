package main

import (
	"eman1can/hca"
	"fmt"
	"os"
)

func main() {
	data, err := os.ReadFile("filename.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	awbMetapackOffset := 13120
	awbMetapackSize := 133153
	acbMetapackOffset := 6560
	acbMetapackSize := 6560

	acbData := make([]byte, acbMetapackSize)
	awbData := make([]byte, awbMetapackSize)
	copy(acbData[acbMetapackOffset:], data)
	copy(awbData[awbMetapackOffset:], data)

	awb := hca.loadAWB(awbData)

}
