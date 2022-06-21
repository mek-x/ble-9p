package main

import (
	"context"
	"log"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
	"github.com/go-ble/ble/linux/hci/cmd"
)

func StartBleScan() {
	scanParams := cmd.LESetScanParameters{
		LEScanType:           0x00,   // 0x00: passive, 0x01: active
		LEScanInterval:       0x0004, // 0x0004 - 0x4000; N * 0.625msec
		LEScanWindow:         0x0004, // 0x0004 - 0x4000; N * 0.625msec
		OwnAddressType:       0x00,   // 0x00: public, 0x01: random
		ScanningFilterPolicy: 0x00,   // 0x00: accept all, 0x01: ignore non-white-listed.
	}

	d, err := linux.NewDevice(ble.OptScanParams(scanParams))
	if err != nil {
		log.Fatalf("can't get device : %s", err)
	}
	ble.SetDefaultDevice(d)

	// Scan for specified duration, or until interrupted by user.
	ctx := ble.WithSigHandler(context.WithCancel(context.Background()))

	go ble.Scan(ctx, false, advHandler, nil)
}

func advHandler(a ble.Advertisement) {
	log.Printf("%s [%ddBm] %s",
		a.Addr().String(),
		a.RSSI(),
		a.LocalName())

	md := a.ManufacturerData()
	sd := a.ServiceData()

	t := time.Now()

	d := make(map[string][]byte)

	if len(md) > 0 {
		d["MD"] = md
	}

	for _, v := range sd {
		d[v.UUID.String()] = v.Data
	}

	UpdateDevice(t, a.Addr().String(), a.RSSI(), d)
}
