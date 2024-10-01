package workflow_test

import (
	"context"
	"errors"
	"flag"
	"io"
	"log/slog"
	"os"
	"reflect"
	"testing"

	"dario.cat/mergo"
	"github.com/golingon/lingon/pkg/testutil"
	wf "github.com/golingon/lingon/pkg/workflow"
)

var lf = flag.Bool("log", false, "show the logs")

type Result struct {
	Err      error
	Messages []string
	State    State
}
type State struct{ Counter int }

func TestPipeline(t *testing.T) {
	wf.SetIDGenerator(wf.StaticID{})

	sf := make([]wf.Step[Result], 0)
	for range 10 {
		sf = append(sf, wf.StepFunc[Result](func(ctx context.Context, r *Result) (*Result, error) {
			r.State.Counter++
			return r, nil
		}))
	}
	var logger *slog.Logger
	if *lf {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	} else {
		logger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	}
	p := wf.NewPipeline(wf.LoggerMiddleware[Result](logger))
	p.Steps = []wf.Step[Result]{
		wf.StepFunc[Result](func(ctx context.Context, r *Result) (*Result, error) {
			r.Messages = append(r.Messages, "first step")
			return r, nil
		}),
		p.Series(
			p.Parallel(wf.MergeTransform[Result](mergo.WithTransformers(addInt{})), sf...),
			wf.StepFunc[Result](func(ctx context.Context, r *Result) (*Result, error) {
				r.Messages = append(r.Messages, "extra serial step")
				return r, nil
			}),
		),
		wf.StepFunc[Result](func(ctx context.Context, r *Result) (*Result, error) {
			f := wf.StepFunc[Result](func(ctx context.Context, r *Result) (*Result, error) {
				r.Messages = append(r.Messages, "extra inner step")
				r.Err = errors.New("oops")
				return r, nil
			})
			resp, err := f.Run(ctx, r)

			sid, err := wf.GetStepID(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if sid.String() != "00000000-0000-0000-0000-000000000028" {
				t.Fatalf("got %q, want %q",
					sid.String(),
					"00000000-0000-0000-0000-000000000028")
			}
			r.Messages = append(r.Messages, "last step")
			return resp, err
		}),
	}

	ctx := context.Background()
	got, err := p.Run(ctx, &Result{})
	if err != nil {
		t.Fatal(err)
	}
	want := &Result{
		Err:   errors.New("oops"),
		State: struct{ Counter int }{Counter: 10},
		Messages: []string{
			"first step",
			"extra serial step",
			"extra inner step",
			"last step",
		},
	}
	if diff := testutil.Diff(got, want); diff != "" {
		t.Fatal(diff)
	}
}

type addInt struct{}

func (t addInt) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	if typ == reflect.TypeOf(int(0)) {
		return func(dst, src reflect.Value) error {
			if dst.CanSet() {
				dst.Set(reflect.ValueOf(int(dst.Int() + src.Int())))
			}
			return nil
		}
	}
	return nil
}
