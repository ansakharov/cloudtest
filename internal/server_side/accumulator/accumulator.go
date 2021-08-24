package accumulator

import (
	"context"
	"os"
	"time"

	"scloud/internal/server_side/dumper"
	"scloud/internal/server_side/entity"
)

type accumulator struct {
	logName string
	file    *os.File
	dumper  dumper.Dumper
	in      chan *entity.Message
	out     chan []int
}

type Accumulator interface {
	Accumulate(duration time.Duration, batchSize int, ctx context.Context)
	GetSavedRange(ctx context.Context) []int
	AddToQueue(message *entity.Message)
}

func New(dumper dumper.Dumper) Accumulator {
	return &accumulator{
		dumper: dumper,
		in:     make(chan *entity.Message, 10000),
		out:    make(chan []int, 1),
	}
}

func (acc *accumulator) AddToQueue(message *entity.Message) {
	acc.in <- message
	return
}

func (acc *accumulator) GetSavedRange(ctx context.Context) []int {
	select {
	case _ = <-ctx.Done():
		return nil
	case saved := <-acc.out:
		return saved
	}
}

func (acc *accumulator) Accumulate(tickFrequency time.Duration, batchSize int, ctx context.Context) {
	tick := time.NewTicker(tickFrequency)
	defer tick.Stop()

	buf := make([]*entity.Message, 0, batchSize)

	// Запись на диск каждые N ms или по достижении K сообщений.
	for {
		select {
		case <-tick.C:
			if len(buf) > 0 {
				written := acc.dumper.WriteOnDisk(buf)
				acc.out <- written
				// re-init
				buf = make([]*entity.Message, 0, batchSize)
			}
		case item := <-acc.in:
			buf = append(buf, item)

			if len(buf) >= batchSize {
				written := acc.dumper.WriteOnDisk(buf)
				acc.out <- written
				// re-init
				buf = make([]*entity.Message, 0, batchSize)
			}
		case _ = <-ctx.Done():
			return
		}
	}
}
