package main

import (
	"fmt"
	"slices"

	"github.com/sandbankdisperser/go-udisks"
)

func main() {
	client, err := udisks.NewClient()
	if err != nil {
		panic(err)
	}
	dd := "CT2000P3-10SSD2-DD564198842D5"
	drive, err := client.DriveById(dd)
	if err != nil {
		panic(err)
	}
	blocks, err := client.BlockDevicesOnDrive(drive.Id)
	if err != nil {
		panic(err)
	}
	blocks = slices.DeleteFunc(blocks, func(b *udisks.BlockDevice) bool {
		return b.IdLabel == "" && b.IdUsage != "crypto"
	})
	for _, v := range blocks {
		fmt.Println(v.Id, "=>", v.UUID, "Device", v.Device, "type", v.IdUsage, "fs", v.Filesystems, "label", v.IdLabel)
	}
}
