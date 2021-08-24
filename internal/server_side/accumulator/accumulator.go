package accumulator

import (
	"context"
	"os"
	"time"

	"scloud/internal/server_side/dumper"
	"scloud/internal/server_side/entity"
)

type accumulator struct {
	logName   string
	file      *os.File
	dumper    dumper.Dumper
	in        chan *entity.Message
	out       chan []int
	writtenCh chan struct{}
}

type Accumulator interface {
	Accumulate(duration time.Duration, batchSize int, ctx context.Context)
	GetSavedRange(ctx context.Context) []int
	AddToQueue(message *entity.Message)
}

func New(dumper dumper.Dumper) Accumulator {
	return &accumulator{
		dumper:    dumper,
		in:        make(chan *entity.Message, 10000),
		out:       make(chan []int, 1),
		writtenCh: make(chan struct{}, 1),
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
	acc.writtenCh <- struct{}{}

	// Мы хотим линейную запись без потенциальных репозиционирований и блоков,
	// поэтому batchSize должен быть большим. Бесконечным он быть не должен - при записи терабайтов логов
	// придется ждать, пока они придут по сети. Поэтому дампим на диск, когда достигнем batchSize, параллельно
	// аккумулируем новые записи.
	// При лагах сети буфер может долго не заполняться до batchSize, а при последней итерации не заполнится до batchSize
	// никогда, вовсе тогда дампим его по таймеру.
	for {
		select {
		case <-tick.C:
			if len(buf) > 0 {
				<-acc.writtenCh
				// дампим на диск асинхронно, параллельно наполняя новый буфер
				go func(locBuff []*entity.Message) {
					written := acc.dumper.WriteOnDisk(locBuff)
					acc.out <- written
					acc.writtenCh <- struct{}{}
				}(buf)

				// Сброс буфера
				buf = make([]*entity.Message, 0, batchSize)
			}
		case item := <-acc.in:
			buf = append(buf, item)

			if len(buf) >= batchSize {
				<-acc.writtenCh
				// дампим на диск асинхронно, параллельно наполняя новый буфер
				go func(locBuff []*entity.Message) {
					written := acc.dumper.WriteOnDisk(locBuff)
					acc.out <- written
					acc.writtenCh <- struct{}{}
				}(buf)

				// Сброс буфера
				buf = make([]*entity.Message, 0, batchSize)
			}
		case _ = <-ctx.Done():
			return
		}
	}
}
