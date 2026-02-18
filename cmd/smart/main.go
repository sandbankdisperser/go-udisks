package main

import (
	"fmt"

	"github.com/sandbankdisperser/go-udisks"
)

func main() {
	client, err := udisks.NewClient()
	if err != nil {
		panic(err)
	}
	drives, err := client.Drives()
	if err != nil {
		panic(err)
	}
	for _, v := range drives {
		if v.Ata != nil {
			fmt.Println(v.Id, "ATA SMART", v.Ata)
		}
		if v.NVMeController != nil {
			fmt.Println(v.Id, "NVME SMART", v.NVMeController)
		}

	}
}
