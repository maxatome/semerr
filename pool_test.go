package semerr_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/maxatome/go-testdeep/td"

	"github.com/maxatome/semerr"
)

func TestPool(t *testing.T) {
	p := semerr.New(context.Background(), 10, 10*time.Second)
	var cnt int32
	for i := 0; i < 100; i++ {
		p.Go(func() error {
			atomic.AddInt32(&cnt, 1)
			time.Sleep(10 * time.Millisecond)
			return nil
		})
	}
	td.CmpNoError(t, p.Wait())
	td.CmpLax(t, cnt, 100)

	// No timeout
	p = semerr.New(context.Background(), 10, 0)
	cnt = 0
	for i := 0; i < 100; i++ {
		p.Go(func() error {
			atomic.AddInt32(&cnt, 1)
			time.Sleep(10 * time.Millisecond)
			return nil
		})
	}
	td.CmpNoError(t, p.Wait())
	td.CmpLax(t, cnt, 100)

	// Timeout
	p = semerr.New(context.Background(), 10, time.Millisecond)
	p.Go(func() error {
		time.Sleep(10 * time.Second)
		return nil
	})
	td.Cmp(t, p.Wait(), context.DeadlineExceeded)

	// Foreign cancel
	ctx, cancel := context.WithCancel(context.Background())
	p = semerr.New(ctx, 10, time.Hour)
	p.Go(func() error {
		time.Sleep(10 * time.Second)
		return nil
	})
	cancel()
	td.Cmp(t, p.Wait(), context.Canceled)
}
