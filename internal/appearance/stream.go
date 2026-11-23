package appearance

import (
	"context"
	"io"
	"sync"
)

type stream struct {
	events   chan Observation
	done     chan struct{}
	stop     context.CancelFunc
	once     sync.Once
	closeErr error // Written by the owner before closing done.
}

func (s *stream) Next(ctx context.Context) (Observation, error) {
	if e := ctx.Err(); e != nil {
		return Observation{}, e
	}
	select {
	case <-s.done:
		return Observation{}, io.EOF
	default:
	}
	select {
	case v := <-s.events:
		return v, nil
	case <-s.done:
		return Observation{}, io.EOF
	case <-ctx.Done():
		return Observation{}, ctx.Err()
	}
}

func (s *stream) Close() error { s.once.Do(s.stop); <-s.done; return s.closeErr }

// Only one owner publishes; slow UI rendering retains one latest observation.
func (s *stream) publish(v Observation) {
	select {
	case s.events <- v:
		return
	default:
	}
	select {
	case <-s.events:
	default:
	}
	select {
	case s.events <- v:
	default:
	}
}
