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
	randGen := flag.Bool("rand", false, "Will generate random bytes. Otherwise will generate sequence of 'a' byte for higher speed.")
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
	// take 1 param from config

	port := 8081
	sender := client_side.New(port, *limit, generator)
	err := sender.Send()
	if err != nil {
		fmt.Println("sender err", err)
	}
	end := time.Now().Sub(st)
	fmt.Println(end)
}
