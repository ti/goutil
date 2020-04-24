package ctxtool

import (
	"context"
)

// DoWithContext add context to function
func DoWithContext(ctx context.Context, do func(ctx context.Context) error, fallback func(err error)) (err error) {
	errorChannel := make(chan error)
	var contextHasBeenDone = false
	go func() {
		err := do(ctx)
		if contextHasBeenDone {
			if fallback != nil {
				fallback(err)
			}
			return
		}
		errorChannel <- err
	}()
	select {
	case err = <-errorChannel:
		return err
	case <-ctx.Done():
		contextHasBeenDone = true
		return ctx.Err()
	}
}
