package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"scloud/internal/client_side"
	"scloud/utils/fake_json"
)

func main() {
	limit := flag.Int("l", 0, "Number of messages for log. Minimum value is 1.")
	flag.Parse()
	if *limit == 0 {
		fmt.Println("Flag -l required to define number of messages")
		os.Exit(1)
	}

	st := time.Now()
	// take 1 param from config
	sender := client_side.New(*limit, fake_json.Generator{})
	err := sender.Send()
	if err != nil {
		fmt.Println("sender err", err)
	}
	end := time.Now().Sub(st)
	fmt.Println(end)
}
