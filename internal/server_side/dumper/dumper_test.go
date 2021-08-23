package dumper

import (
	"testing"

	dumper_mock "scloud/internal/server_side/dumper/mocks"
	"scloud/internal/server_side/entity"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestDumper_WriteOnDisk_No_Writes(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	file := dumper_mock.NewMockFile(ctl)
	dump := New(file)

	t.Run("no_messages", func(t *testing.T) {
		saved := dump.WriteOnDisk(nil)
		require.Nil(t, saved)
	})
}

func TestDumper_WriteOnDisk_Successful_Writes(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	messages := []*entity.Message{
		{
			Key:   0,
			Value: []byte("Say 'Hello'"),
		},
		{
			Key:   1,
			Value: []byte("and die"),
		},
	}

	bytesExpected := append(messages[0].Value, '\n')
	bytesExpected = append(bytesExpected, messages[1].Value...)
	bytesExpected = append(bytesExpected, '\n')

	file := dumper_mock.NewMockFile(ctl)
	file.EXPECT().Write(bytesExpected).Return(0, nil).Times(1)
	dump := New(file)

	saved := dump.WriteOnDisk(messages)
	require.Equal(t, []int{messages[0].Key, messages[1].Key}, saved)
}
