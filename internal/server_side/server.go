package server_side

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"scloud/internal/server_side/accumulator"
	"scloud/internal/server_side/entity"
)

type server struct {
	port     int
	acc      accumulator.Accumulator
	messages map[int]string
	done     chan struct{}
}

type Server interface {
	ListenAndServe()
}

func New(
	port int,
	acc accumulator.Accumulator,
) Server {
	return &server{
		port: port,
		acc:  acc,
		done: make(chan struct{}, 1),
	}
}

func (s *server) ListenAndServe() {
	go func() {
		s.acc.Accumulate(time.Millisecond*200, 50000, s.done)
	}()

	listener, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(s.port))
	if err != nil {
		log.Fatal("Can't start server", err)
	}
	finished := make(chan struct{}, 1)
	for {
		fmt.Println("Waiting for new connection")
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept conn err", err)
			continue
		}

		doneChan := make(chan struct{}, 1)
		go s.handleReq(conn, doneChan)
		go s.sendSavedIDs(conn, doneChan, finished)
		<-finished
	}
}

func (s *server) handleReq(conn net.Conn, doneChan chan<- struct{}) {
	defer func() {
		errCloseConn := conn.Close()
		if errCloseConn != nil {
			println(errCloseConn, "closeConnErr")
		}
		fmt.Println("connection processed")
	}()

	reader := bufio.NewReader(conn)
	for {
		rawInfo, _ := reader.ReadString('\n')

		// Клиент прислал финальный ack.
		if rawInfo == "ack\n" {
			doneChan <- struct{}{}

			return
		}

		key, msg := s.parseMsg(&rawInfo)
		s.acc.AddToQueue(&entity.Message{Key: key, Value: msg})
	}
}

func (s *server) sendSavedIDs(conn net.Conn, done <-chan struct{}, finished chan<- struct{}) {
	ti := time.NewTicker(time.Millisecond * 200)
	for {
		select {
		case _ = <-done:
			finished <- struct{}{}

			return
		case <-ti.C:
			saved := s.acc.GetSavedRange()
			if len(saved) != 0 {
				for _, num := range saved {
					// Ack message num
					numStr := strconv.Itoa(num)
					_, err := conn.Write([]byte(numStr + "\n"))
					if err != nil {
						fmt.Println(err, "writeAnswer")
					}
				}
			}
		}
	}
}

func (s *server) parseMsg(rawInfo *string) (int, []byte) {
	keyMsg := strings.Split(*rawInfo, "\r")
	keyStr, msg := keyMsg[0], keyMsg[1]
	key, _ := strconv.Atoi(keyStr)
	msg = msg[:len(msg)-1]

	return key, []byte(msg)
}
