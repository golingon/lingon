package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var ErrMissingFromContext = errors.New("missing from context")

type ctxKey int

const (
	stepIDKey ctxKey = iota + 1
	stepHistory
)

func setStepID(ctx context.Context, stepID uuid.UUID) context.Context {
	return context.WithValue(ctx, stepIDKey, stepID)
}

func GetStepID(ctx context.Context) (uuid.UUID, error) {
	v, ok := ctx.Value(stepIDKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("step %w", ErrMissingFromContext)
	}
	return v, nil
}

// UUID

var cpt = 0

const u = "00000000-0000-0000-0000-000000000000"

func incuuid() uuid.UUID {
	cpt++
	switch {
	case cpt < 10:
		return pu(fmt.Sprintf("%s%d", u[:len(u)-1], cpt))
	case cpt < 100:
		return pu(fmt.Sprintf("%s%d", u[:len(u)-2], cpt))
	default:
		return pu(u)
	}
}

func pu(s string) uuid.UUID {
	x, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return x
}
