package workflow_test

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"dario.cat/mergo"
	"github.com/golingon/lingon/pkg/testutil"
	wf "github.com/golingon/lingon/pkg/workflow"
)

var lf = flag.Bool("log", false, "show the logs")

type Result struct {
	Err      error
	Messages []string
	State    State
	sync.Mutex
}
type State struct{ Counter int }

func TestEmptyPipeline(t *testing.T) {
	p := wf.NewPipeline[Result]()
	_, err := p.Run(context.Background(), &Result{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMiddleware(t *testing.T) {
	cpt := 0
	incr := func() {
		cpt++
	}
	mid := func(inc func()) wf.Middleware[Result] {
		return func(next wf.Step[Result]) wf.Step[Result] {
			return wf.MidFunc[Result](func(ctx context.Context, res *Result) (*Result, error) {
				inc()
				return next.Run(ctx, res)
			})
		}
	}
	p := wf.NewPipeline(mid(incr))
	_, err := p.Run(context.Background(), &Result{Messages: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	if cpt != 0 {
		t.Fatalf("cpt %d != 0", cpt)
	}

	p.Steps = append(p.Steps, wf.StepFunc[Result](func(ctx context.Context, res *Result) (*Result, error) {
		return res, nil
	}))
	_, err = p.Run(context.Background(), &Result{Messages: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	if cpt != 1 {
		t.Fatalf("cpt %d != 1", cpt)
	}
}

func TestPipeline(t *testing.T) {
	wf.SetIDGenerator(&wf.StaticID{})

	sf := make([]wf.Step[Result], 0)
	for range 10 {
		sf = append(sf, wf.StepFunc[Result](func(ctx context.Context, r *Result) (*Result, error) {
			r.Lock()
			defer r.Unlock()
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

	p := wf.NewPipeline(LoggerMiddleware[Result](logger))
	p.Steps = []wf.Step[Result]{
		wf.StepFunc[Result](func(ctx context.Context, r *Result) (*Result, error) {
			r.Messages = append(r.Messages, "first step")
			return r, nil
		}),
		p.Series(
			p.Parallel(wf.MergeTransform[Result](mergo.WithTransformers(addInt{})), sf...),
			wf.StepFunc[Result](func(ctx context.Context, r *Result) (*Result, error) {
				r.Messages = append(r.Messages, "extra serial step")
				r.Err = errors.Join(r.Err, errIgnoreMe)
				return r, nil
			}),
		),
		handleErr{l: logger},
		wf.StepFunc[Result](func(ctx context.Context, r *Result) (*Result, error) {
			f := wf.StepFunc[Result](func(ctx context.Context, r *Result) (*Result, error) {
				r.Messages = append(r.Messages, "extra inner step")
				r.Err = errors.Join(r.Err, errors.New("oops"))
				return r, nil
			})
			resp, err := f.Run(ctx, r)
			if err != nil {
				t.Fatal(err)
			}
			sid, err := wf.GetStepID(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if sid.String() != "00000000-0000-0000-0000-000000000029" {
				t.Fatalf("got %q, want %q",
					sid.String(),
					"00000000-0000-0000-0000-000000000029")
			}
			r.Messages = append(r.Messages, "last step")
			return resp, err
		}),
	}

	ctx := context.Background()
	got, err := p.Run(ctx, &Result{Messages: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	want := &Result{
		Err:   errors.Join(nil, errors.New("oops")),
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

func LoggerMiddleware[T any](l *slog.Logger) wf.Middleware[T] {
	return func(next wf.Step[T]) wf.Step[T] {
		return wf.MidFunc[T](func(ctx context.Context, res *T) (*T, error) {
			name := wf.Name(next)
			if name != "MidFunc" {
				id, _ := wf.GetStepID(ctx)
				l.Info("start", "Type", name, "id", id, "STEP", next)
			}
			resp, err := next.Run(ctx, res)
			if name != "MidFunc" {
				id, _ := wf.GetStepID(ctx)
				t, errctx := wf.GetStepStartTime(ctx)
				if errors.Is(errctx, wf.ErrMissingFromContext) {
					l.Info("done", "Type", name, "id", id, "Result", fmt.Sprintf("%v", resp))
				} else {
					l.Info("done", "Type", name, "id", id, "duration", time.Since(t),
						"Result", fmt.Sprintf("%v", resp))
				}
			}
			return resp, err
		})
	}
}

type handleErr struct{ l *slog.Logger }

var errIgnoreMe = errors.New("ignore me")

func (h handleErr) Run(ctx context.Context, r *Result) (*Result, error) {
	if errors.Is(r.Err, errIgnoreMe) {
		h.l.Error("ignoring error", "err", r.Err)
		r.Err = nil
		return r, r.Err
	}
	h.l.Error("handling error", "err", r.Err)
	return r, r.Err
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
