package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/golingon/lingon/pkg/testutil"
)

func TestPipeline(t *testing.T) {
	sf := make([]Step, 0)
	for range 10 {
		sf = append(sf, StepFunc(func(ctx context.Context, r *Request) (*Request, error) {
			r.State.Counter++
			return r, nil
		}))
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	p := NewPipeline(LoggerMiddleware(logger))
	p.Steps = []Step{
		StepFunc(func(ctx context.Context, r *Request) (*Request, error) {
			r.State.Messages = append(r.State.Messages, "first step")
			return r, nil
		}),
		p.Series(
			p.Parallel(Merge, sf...),
			StepFunc(func(ctx context.Context, r *Request) (*Request, error) {
				r.State.Messages = append(r.State.Messages, "extra serial step")
				return r, nil
			}),
		),
		StepFunc(func(ctx context.Context, r *Request) (*Request, error) {
			f := StepFunc(func(ctx context.Context, r *Request) (*Request, error) {
				r.State.Messages = append(r.State.Messages, "extra inner step")
				r.State.Err = errors.New("oops")
				return r, nil
			})
			resp, err := f.Run(ctx, r)

			sid, err := GetStepID(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if sid.String() != "00000000-0000-0000-0000-000000000028" {
				t.Log(sid.String())
			}
			r.State.Messages = append(r.State.Messages, "last step")
			return resp, err
		}),
	}

	ctx := context.Background()
	got, err := p.Run(ctx, &Request{State: &State{}})
	if err != nil {
		t.Fatal(err)
	}
	want := &Request{
		State: &State{
			Err:     errors.New("oops"),
			Counter: 10,
			Messages: []string{
				"first step",
				"extra serial step",
				"extra inner step",
				"last step",
			},
		},
	}
	if diff := testutil.Diff(got, want); diff != "" {
		t.Fatal(diff)
	}
}
