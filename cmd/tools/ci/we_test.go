package main

import (
	"context"
	"log/slog"
	"testing"
)

func TestPipeline(t *testing.T) {
	sf := make([]Step, 0)
	for i := range 10 {
		sf = append(sf, StepFunc(func(ctx context.Context, r *Request) (*Request, error) {
			t.Log("stepfunc", i)
			return r, nil
		}))
	}
	p := NewPipeline(LoggerMiddleware(slog.Default()))
	p.Steps = []Step{
		StepFunc(func(ctx context.Context, r *Request) (*Request, error) {
			t.Log("first step")
			return r, nil
		}),
		Series(
			Parallel(sf...),
			StepFunc(func(ctx context.Context, r *Request) (*Request, error) {
				t.Log("step extra")
				return r, nil
			}),
		),
		StepFunc(func(ctx context.Context, r *Request) (*Request, error) {
			t.Log("last step")
			return r, nil
		}),
	}

	req, err := p.Run(context.Background(), &Request{})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("req: %s\n", req.String())
}
