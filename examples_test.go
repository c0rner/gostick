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

	fmt.Printf("Found '%s' with serial '%s'\n", stick.Model, stick.Serial)

	// Request device firmware version
	stick.SendRaw(strings.NewReader("V+"))

	// Poll until we get a reply
	for {
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
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	stick.Close()
}
