package workflow

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
		return uuid.UUID{}, ErrMissingFromContext
	}
	return v, nil
}

// UUID

var gen IDGenerator = UUIDV7{}

type UUIDV7 struct{}

func (_ UUIDV7) ID() uuid.UUID { return uuid.Must(uuid.NewV7()) }

// SetIDGenerator allows to set a [StaticID] generator for testing purposes.
func SetIDGenerator(g IDGenerator) {
	gen = g
}

func GetIDGenerator() IDGenerator {
	return gen
}

type IDGenerator interface {
	ID() uuid.UUID
}

var cpt = 0

const u = "00000000-0000-0000-0000-000000000000"

type StaticID struct{}

// ID returns a [uuid.UUID] starting at 1 in the form of
// "00000000-0000-0000-0000-000000000001". Each following calls increment by 1.
//
// Note: It is meant for debugging and testing.
func (_ StaticID) ID() uuid.UUID {
	// not efficient, meant for tests
	cpt++
	switch {
	case cpt < 10:
		return parse(fmt.Sprintf("%s%d", u[:len(u)-1], cpt))
	case cpt < 100:
		return parse(fmt.Sprintf("%s%d", u[:len(u)-2], cpt))
	case cpt < 1000:
		return parse(fmt.Sprintf("%s%d", u[:len(u)-3], cpt))
	default:
		return parse(u)
	}
}

func parse(s string) uuid.UUID {
	x, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return x
}
