package main

import (
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
	if err := client.PowerOff(drive); err != nil {
		panic(err)
	}
}
