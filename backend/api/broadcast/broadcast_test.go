package broadcast_test

import (
	"context"
	"testing"

	"github.com/romshark/taskhub/api/broadcast"
	"github.com/stretchr/testify/require"
)

func TestNotify(t *testing.T) {
	b := broadcast.New[int]()

	sub1, sub2 := make(chan int, 1), make(chan int, 1)
	err := b.Subscribe(context.Background(), sub1)
	require.NoError(t, err)
	err = b.Subscribe(context.Background(), sub2)
	require.NoError(t, err)
	require.Equal(t, 2, b.Len())

	err = b.Notify(context.Background(), 42)
	require.NoError(t, err)

	require.Equal(t, 42, <-sub1)
	require.Equal(t, 42, <-sub2)
}

func TestClose(t *testing.T) {
	b := broadcast.New[int]()

	ctx, cancel := context.WithCancel(context.Background())

	sub1, sub2 := make(chan int, 1), make(chan int, 1)
	err := b.Subscribe(ctx, sub1)
	require.NoError(t, err)
	err = b.Subscribe(ctx, sub2)
	require.NoError(t, err)

	cancel()

	err = b.Notify(context.Background(), 42)
	require.NoError(t, err)

	// Drain the channel
	if v, ok := <-sub1; ok {
		require.Equal(t, 42, v)
	}
	if v, ok := <-sub2; ok {
		require.Equal(t, 42, v)
	}

	_, ok := <-sub1
	require.False(t, ok)
	_, ok = <-sub2
	require.False(t, ok)

	require.Equal(t, 0, b.Len())
}
