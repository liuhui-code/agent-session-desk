package pet

import (
	"context"
	"fmt"
	"time"
)

type Options struct {
	Live     *Companion
	State    Kind
	Count    int
	Duration time.Duration
	Report   string
}

func Run(ctx context.Context, o Options) error {
	if o.Live == nil {
		if _, err := New(o.State, o.Count); err != nil {
			return err
		}
	}
	if o.Duration < 0 || o.Duration > 5*time.Minute {
		return fmt.Errorf("demo duration must be 0-5m")
	}
	return runNative(ctx, o)
}
