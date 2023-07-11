package broadcast

import (
	"context"
	"sync"
)

// Broadcast allows for concurrent subscriber notification.
type Broadcast[T any] struct {
	lock          sync.RWMutex
	idCounter     uint64
	subscriptions map[uint64]chan<- T
}

func New[T any]() *Broadcast[T] {
	return &Broadcast[T]{
		subscriptions: make(map[uint64]chan<- T),
	}
}

// Len returns the number of subscriptions.
func (s *Broadcast[T]) Len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return len(s.subscriptions)
}

// Notify sends t to all subscription channels in a blocking manner.
// Notify acquires a shared lock and is therefore safe to be called concurrently.
func (s *Broadcast[T]) Notify(ctx context.Context, t T) error {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, s := range s.subscriptions {
		s <- t
	}
	return nil
}

// Subscribe registers c to receive notifications on.
// The channel is closed and unregistered when ctx is canceled.
// Subscribe acquires an exlusive lock and is
// therefore safe to be called concurrently.
func (s *Broadcast[T]) Subscribe(ctx context.Context, c chan<- T) error {
	s.lock.Lock()
	s.idCounter++
	id := s.idCounter
	s.subscriptions[id] = c
	s.lock.Unlock()
	go func() {
		<-ctx.Done()
		// Subscription canceled
		s.lock.Lock()
		close(s.subscriptions[id])
		delete(s.subscriptions, id)
		s.lock.Unlock()
	}()
	return nil
}
