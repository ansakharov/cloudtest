package server_side

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"scloud/internal/server_side/accumulator"
	"scloud/internal/server_side/entity"
)

type server struct {
	port          int
	acc           accumulator.Accumulator
	messages      map[int]string
	parentContext context.Context
	parentCancel  context.CancelFunc
}

type Server interface {
	ListenAndServe()
}

func New(
	port int,
	acc accumulator.Accumulator,
) Server {
	parentCtx := context.Background()
	parentCtx, parentCancel := context.WithCancel(parentCtx)
	return &server{
		port:          port,
		acc:           acc,
		parentContext: parentCtx,
		parentCancel:  parentCancel,
	}
}

func (s *server) ListenAndServe() {
	var wg sync.WaitGroup
	go func() {

		s.acc.Accumulate(time.Millisecond*200, 50000, s.parentContext)
	}()
	defer s.parentCancel()

	listener, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(s.port))
	if err != nil {
		log.Fatal("Can't start server", err)
	}
	for {
		ctx, cancel := context.WithCancel(s.parentContext)

		fmt.Println("Waiting for new connection")
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept conn err", err)
			continue
		}

		wg.Add(2)
		go func() {
			defer wg.Done()
			s.handleReq(conn, cancel)
		}()
		go func() {
			defer wg.Done()
			s.sendSavedIDs(conn, ctx)
		}()
		wg.Wait()
	}
}

func (s *server) handleReq(conn net.Conn, cancel context.CancelFunc) {
	defer func() {
		fmt.Println("connection processed")

		errCloseConn := conn.Close()
		if errCloseConn != nil {
			println(errCloseConn, "closeConnErr")
		}

		cancel()
	}()

	reader := bufio.NewReader(conn)
	for {
		rawInfo, _ := reader.ReadString('\n')

		// Клиент прислал финальный ack.
		if rawInfo == "ack\n" {
			return
		}

		key, msg := s.parseMsg(&rawInfo)
		s.acc.AddToQueue(&entity.Message{Key: key, Value: msg})
	}
}

func (s *server) sendSavedIDs(conn net.Conn, ctx context.Context) {
	ti := time.NewTicker(time.Millisecond * 200)

	for {
		select {
		case _ = <-ctx.Done():
			return
		case <-ti.C:
			// req ended
			saved := s.acc.GetSavedRange(ctx)
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
	key, err := strconv.Atoi(keyStr)
	if err != nil {
		fmt.Println(err)
	}
	msg = msg[:len(msg)-1]

	return key, []byte(msg)
}
