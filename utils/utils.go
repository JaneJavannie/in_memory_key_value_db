package utils

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

func GetRequestUUID() string {
	return uuid.New().String()
}

func WithRetries(ctx context.Context, retriesNumber int, initialDelayMs time.Duration, action func() error) error {
	if action == nil {
		return errors.New("incorrect action")
	}

	var err error
	for retry := 1; retry <= retriesNumber; retry++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err = action(); err == nil {
			break
		}

		time.Sleep(initialDelayMs * time.Duration(retry))
	}

	return err
}
