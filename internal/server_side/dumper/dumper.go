package dumper

import (
	"scloud/internal"
	"scloud/internal/server_side/entity"
)

type dumper struct {
	file internal.File
}

type Dumper interface {
	WriteOnDisk(messages []*entity.Message) []int
}

func New(file internal.File) *dumper {
	return &dumper{
		file: file,
	}
}

func (dump *dumper) WriteOnDisk(messages []*entity.Message) []int {
	if len(messages) == 0 {
		return nil
	}

	// Буфер по 4096b для каждого сообщения + 4096b на \n
	dataToWrite := make([]byte, 0, 4096*len(messages)+4096)
	var saved []int

	for _, m := range messages {
		dataToWrite = append(dataToWrite, m.Value...)
		dataToWrite = append(dataToWrite, '\n')

		saved = append(saved, m.Key)
	}

	// Не рассматриваем проблемы с записью на диск
	_, _ = dump.file.Write(dataToWrite)
	// НЕ делаем Close, поэтому сразу нужно сделать сброс на диск.
	_ = dump.file.Sync()

	return saved
}
