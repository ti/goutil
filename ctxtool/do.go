import "context"

// DoWithContext add context to function
func DoWithContext(ctx context.Context, do func(ctx context.Context) error, fallback func(err error)) (err error) {
	errorChannel := make(chan error)
	var contextHasBeenDone bool
	go func() {
		err := do(ctx)
		errorChannel <- err
		if contextHasBeenDone && fallback != nil {
			fallback(err)
		}
	}()
	select {
	case err = <-errorChannel:
		return err
	case <-ctx.Done():
		contextHasBeenDone = true
		return ctx.Err()
	}
}
