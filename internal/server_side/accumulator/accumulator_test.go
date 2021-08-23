package accumulator

import (
	"fmt"
	"sync"
	"testing"
	"time"

	dumper "scloud/internal/server_side/dumper/mocks"
	"scloud/internal/server_side/entity"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestAccumulator_Accumulate_Write_On_Time_Trigger(t *testing.T) {
	var wg sync.WaitGroup

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	toQueue := entity.Message{
		Key:   999,
		Value: []byte("1-2-3"),
	}
	done := make(chan struct{}, 1)

	dump := dumper.NewMockDumper(ctl)
	dump.EXPECT().WriteOnDisk([]*entity.Message{&toQueue}).Return([]int{toQueue.Key}).Times(1)
	acc := New(dump)

	wg.Add(1)
	go func() {
		defer wg.Done()
		acc.AddToQueue(&toQueue)
	}()
	wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Ставим емкость бача 2, а запишем 1 значение, чтобы писать в файл по таймеру.
		acc.Accumulate(time.Nanosecond, 2, done)
	}()

	var saved []int
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			time.Sleep(time.Millisecond)
			saved = acc.GetSavedRange()
			fmt.Println(saved)
			if len(saved) > 0 {
				break
			}
		}

		done <- struct{}{}
	}()
	wg.Wait()

	require.Len(t, saved, 1)
	require.Equal(t, []int{toQueue.Key}, saved)
}

func TestAccumulator_Accumulate_Write_On_Capacity_Trigger(t *testing.T) {
	var wg sync.WaitGroup

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	toQueue := entity.Message{
		Key:   999,
		Value: []byte("1-2-3"),
	}
	done := make(chan struct{}, 1)

	dump := dumper.NewMockDumper(ctl)
	dump.EXPECT().WriteOnDisk([]*entity.Message{&toQueue}).Return([]int{toQueue.Key}).Times(1)
	acc := New(dump)

	wg.Add(1)
	go func() {
		defer wg.Done()
		acc.AddToQueue(&toQueue)
	}()
	wg.Wait()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Ставим ожидание час, чтобы пойти по пути заполнения до batchSize.
		acc.Accumulate(time.Hour, 1, done)
	}()

	var saved []int
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			time.Sleep(time.Millisecond)
			saved = acc.GetSavedRange()
			fmt.Println(saved)
			if len(saved) > 0 {
				break
			}
		}

		done <- struct{}{}
	}()
	wg.Wait()

	require.Len(t, saved, 1)
	require.Equal(t, []int{toQueue.Key}, saved)
}
