package ctxtool

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDoWithContext(t *testing.T) {

	fnErr := errors.New("fn error")
	fn := func() error {
		time.Sleep(3 * time.Second)
		return fnErr
	}

	ctx, c := context.WithTimeout(context.Background(), 2 * time.Second)
	defer c()

	var fnCallbackError error
	err := DoWithContext(ctx, func(ctx context.Context) error {
		return fn()
	}, func(err error) {
		fnCallbackError = err
	})

	if err != context.DeadlineExceeded {
		t.Fatal()
	}
	// catch the fn callback error
	time.Sleep(2 * time.Second)
	if fnCallbackError != fnErr {
		t.Fatal()
	}


	ctx, c = context.WithTimeout(context.Background(), 10 * time.Second)
	defer c()

	err = DoWithContext(ctx, func(ctx context.Context) error {
		return fn()
	}, func(err error) {
		// this should not happen
		t.Fatal()
	})

	if err != fnErr {
		t.Fatal()
	}
}
