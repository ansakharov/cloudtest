package client_side

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"scloud/utils/fake_json"
)

type client struct {
	limit    int
	gen      fake_json.Generator
	storage  map[int][]byte
	messages [][]byte
}

type Client interface {
	Send() error
}

func New(
	limit int,
	gen fake_json.Generator,
) Client {
	return &client{
		limit:   limit,
		gen:     gen,
		storage: make(map[int][]byte, limit),
	}
}

func (c *client) Send() error {
	var err error
	var wg sync.WaitGroup

	fmt.Println("Creating messages, wait please")

	start := time.Now()
	c.prepareData()

	fmt.Printf("Messages created, spent %.3fs\n", time.Now().Sub(start).Seconds())
	start = time.Now()
	fmt.Println("Start sending messages")

	conn, err := net.Dial("tcp", "127.0.0.1:8081")
	if err != nil {
		fmt.Println("failed connect to server", err)

		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		c.produceEvents(conn)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		c.consumeAcks(conn)
	}()

	wg.Wait()
	fmt.Printf("All acks received, spent %.3fs\n", time.Now().Sub(start).Seconds())

	return nil
}

func (c *client) produceEvents(conn net.Conn) {
	buff := bytes.Buffer{}
	for k, msg := range c.messages {
		buff.Reset()
		_, _ = buff.WriteString(strconv.Itoa(k))
		_, _ = buff.WriteString("\r")
		_, _ = buff.Write(msg)
		_, _ = buff.WriteString("\n")

		_, _ = conn.Write(buff.Bytes())
	}
	return
}

func (c *client) consumeAcks(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("failed read from TCP socket", err)
			os.Exit(1)
		}
		keyStr := message[:len(message)-1]
		key, _ := strconv.Atoi(keyStr)

		//fmt.Println("Ack по ключу", key, " удалим из мапы")

		delete(c.storage, key)

		if len(c.storage) == 0 {
			_, err = conn.Write([]byte("ack\n"))
			if err != nil {
				fmt.Println("can't write to socket", err)
				os.Exit(1)
			}

			return
		}
	}
}

func (c *client) prepareData() {
	for i := 0; i < c.limit; i++ {
		msg := c.gen.Gen(4096)
		c.storage[i] = msg
		c.messages = append(c.messages, msg)
	}
}
