package main

import (
	"fmt"
	"os"
	"time"

	"scloud/internal/server_side"
	"scloud/internal/server_side/accumulator"
	"scloud/internal/server_side/dumper"
)

func main() {
	port := 8081
	logName := fmt.Sprintf("log%d.txt", time.Now().UnixNano())
	_, err := os.Create(logName)
	if err != nil {
		fmt.Println("cant create log")
		os.Exit(1)
	}

	file, err := os.OpenFile(logName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.ModeAppend)
	if err != nil {
		fmt.Println("open err", err)
		os.Exit(1)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	dump := dumper.New(file)
	acc := accumulator.New(dump)

	server_side.New(port, acc).ListenAndServe()

	fmt.Println("Server went down")
}
