package workflow

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

var ErrMissingFromContext = errors.New("missing from context")

type ctxKey int

const (
	stepIDKey ctxKey = iota + 1
	stepStartTime
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

func setStepStartTime(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, stepStartTime, t)
}

func GetStepStartTime(ctx context.Context) (time.Time, error) {
	v, ok := ctx.Value(stepStartTime).(time.Time)
	if !ok {
		return time.Time{}, ErrMissingFromContext
	}
	return v, nil
}

// UUID

// IDGenerator generates globally unique ID [uuid.UUID]. It uses UUID version 7 which time-sorted.
// See https://uuid7.com, for more information on UUIDv7.
// It means to assign an ID for each [Step] of a [Pipeline].
// It can be overwritten by [SetIDGenerator].
var gen IDGenerator = &UUIDv7{}

type UUIDv7 struct{ sync.Mutex }

func (g *UUIDv7) ID() uuid.UUID {
	g.Lock()
	defer g.Unlock()
	return uuid.Must(uuid.NewV7())
}

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

type StaticID struct{ sync.Mutex }

// ID returns a [uuid.UUID] starting at 1 in the form of
// "00000000-0000-0000-0000-000000000001". Each following calls increment by 1.
//
// Note: It is meant for debugging and testing.
func (g *StaticID) ID() uuid.UUID {
	g.Lock()
	defer g.Unlock()
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
