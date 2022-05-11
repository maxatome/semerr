package semerr

import (
	"context"
	"time"

	"golang.org/x/sync/semaphore"
)

type Pool struct {
	fns  chan func() error
	done chan error
}

func New(ctx context.Context, maxWorkers int, timeout time.Duration) *Pool {
	pool := Pool{
		fns:  make(chan func() error, maxWorkers),
		done: make(chan error, 1),
	}
	go func() {
		var cancel context.CancelFunc
		if timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, timeout)
		} else {
			ctx, cancel = context.WithCancel(ctx)
		}
		defer cancel()

		sem := semaphore.NewWeighted(int64(maxWorkers))

		errCh := make(chan error, 1)
		var err error
	loop:
		for {
			select {
			case err = <-errCh:
				cancel()
				break loop

			case <-ctx.Done():
				err = ctx.Err()
				break loop

			case fn, ok := <-pool.fns:
				if !ok {
					break loop
				}
				if err = sem.Acquire(ctx, 1); err != nil {
					break loop
				}
				go func() {
					defer sem.Release(1)
					if err := fn(); err != nil {
						errCh <- err
					}
				}()
			}
		}

		if err == nil {
			if e2 := sem.Acquire(ctx, int64(maxWorkers)); e2 != nil {
				select {
				case err = <-errCh:
				default:
					err = e2
				}
			}
		}

		pool.done <- err
		close(pool.done)
	}()

	return &pool
}

func (p *Pool) Go(fn func() error) {
	p.fns <- fn
}

func (p *Pool) Wait() error {
	close(p.fns)
	return <-p.done
}
