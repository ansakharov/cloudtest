package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"scloud/internal/client_side"
	"scloud/utils/fake_messages"
)

func main() {
	limit := flag.Int("l", 0, "Number of messages for log. Minimum value is 1.")
	randGen := flag.Bool("rand", false, "Will generate random bytes for messages. Otherwise will generate sequence of 'a' byte for higher speed.")
	flag.Parse()
	if *limit == 0 {
		fmt.Println("Flag -l required to define number of messages")
		os.Exit(1)
	}

	var generator fake_messages.BytesGenerator
	if !*randGen {
		generator = &fake_messages.SimpleGenerator{}
	} else {
		generator = &fake_messages.RandomGenerator{}
	}

	st := time.Now()
	port := 8081

	sender := client_side.New(port, *limit, generator)
	err := sender.Produce()
	if err != nil {
		fmt.Println("sender err", err)
	}

	fmt.Printf("Total time spent: %0.2fs\n", time.Now().Sub(st).Seconds())
}
