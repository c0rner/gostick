package gostick

import (
	"fmt"
	"log"
	"strings"
	"time"
)

func Example() {
	stick, err := New()
	if err != nil {
		log.Fatalf("Unable to open Tellstick device, %s.\n", err)
	}

	fmt.Printf("Found '%s' with serial number '%s'\n", stick.Model, stick.Serial)

	// Request device firmware version
	stick.SendRaw(gostick.MsgGetVersion)

	// Poll 600 times (60 seconds)
	for i := 0; i < 600; i++ {
		var msg []string
		msg, err = stick.Poll()
		if err != nil {
			log.Printf("Poll() failed, %s.\n", err)
			break
		}
		if len(msg) > 0 {
			fmt.Printf("Poll() returned %d message(s)\n", len(msg))
			for _, m := range msg {
				fmt.Printf("Message: %s\n", m)
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Close and release usb device
	stick.Close()
}
