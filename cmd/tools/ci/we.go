package main

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"strings"

	"dario.cat/mergo"
	"golang.org/x/sync/errgroup"
)

type Step interface {
	Run(ctx context.Context, req *Request) (*Request, error)
}

type Pipeline struct {
	Steps []Step
	Mid
}

func (p *Pipeline) Run(ctx context.Context, req *Request) (*Request, error) {
	resp := req
	var err error
	for i := range p.Steps {
		for _, m := range slices.Backward(p.Mid) {
			p.Steps[i] = m(p.Steps[i])
		}
		ctx = setStepID(ctx, incuuid())
		resp, err = p.Steps[i].Run(ctx, req)
		if err != nil {
			return nil, err
		}
		req = resp
	}
	return resp, nil
}

func NewPipeline(mid ...Middleware) *Pipeline {
	return &Pipeline{
		Mid: mid,
	}
}

func Name(s Step) string {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return strings.ToUpper(t.Name())
}

type Request struct {
	State *State
}

func (r *Request) String() string {
	return fmt.Sprintf("%#v", r.State)
}

type State struct {
	Err      error
	Messages []string
	Counter  int
}

// StepFunc

type StepFunc func(context.Context, *Request) (*Request, error)

func (f StepFunc) Run(ctx context.Context, req *Request) (*Request, error) {
	return f(setStepID(ctx, incuuid()), req)
}

// Middleware

type MidFunc func(context.Context, *Request) (*Request, error)

func (f MidFunc) Run(ctx context.Context, req *Request) (*Request, error) {
	return f(ctx, req)
}

type Middleware func(s Step) Step

type Mid []Middleware

func LoggerMiddleware(l *slog.Logger) Middleware {
	return func(next Step) Step {
		return MidFunc(func(ctx context.Context, req *Request) (*Request, error) {
			name := Name(next)
			if name != "MidFunc" {
				l.Info("step start 🏁", "Type", name, "Request", req)
			}
			resp, err := next.Run(ctx, req)

			if name != "MidFunc" {
				l.Info("step done  ✅", "Type", name, "Request", req)
			}
			return resp, err
		})
	}
}

// Series
type series struct {
	Stages []Step
	Mid
}

func (p *Pipeline) Series(steps ...Step) *series {
	return &series{
		Stages: steps,
		Mid:    p.Mid,
	}
}

func (s *series) Run(ctx context.Context, req *Request) (*Request, error) {
	var err error
	resp := req

	for i := range s.Stages {
		for _, m := range slices.Backward(s.Mid) {
			s.Stages[i] = m(s.Stages[i])
		}
		ctx = setStepID(ctx, incuuid())
		resp, err = s.Stages[i].Run(ctx, req)
		if err != nil {
			return nil, err
		}
		req = resp
	}
	return resp, nil
}

// Parallel

type parallel struct {
	merge MergeRequest
	Tasks []Step
	Mid
}

type MergeRequest func(context.Context, *Request, ...*Request) (*Request, error)

// TODO: how to configure the merge
func (p *Pipeline) Parallel(merge MergeRequest, steps ...Step) *parallel {
	return &parallel{
		merge: merge,
		Tasks: steps,
		Mid:   p.Mid,
	}
}

func (p *parallel) Run(ctx context.Context, req *Request) (*Request, error) {
	resps := make([]*Request, len(p.Tasks))
	g, groupCtx := errgroup.WithContext(ctx)

	tasks := make([]Step, len(p.Tasks))
	for i, s := range p.Tasks {
		tasks[i] = s
		for _, m := range slices.Backward(p.Mid) {
			tasks[i] = m(tasks[i])
		}
	}
	for i := range tasks {
		g.Go(func() error {
			defer CapturePanic(groupCtx)

			copyReq := &Request{}
			*copyReq = *req
			groupCtx = setStepID(groupCtx, incuuid())
			resp, err := tasks[i].Run(groupCtx, copyReq)
			if err != nil {
				return err
			}
			resps[i] = resp
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return p.merge(ctx, req, resps...)
}

func Merge(ctx context.Context, req *Request, responses ...*Request) (*Request, error) {
	var err error
	for _, r := range responses {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("aborting: %w", ctx.Err())
		default:
			err = mergo.Merge(req, r)
			if err != nil {
				return nil, err
			}
			// TODO: get step ID from ctx
		}
	}
	return req, nil
}

func CapturePanic(ctx context.Context) {
	if r := recover(); r != nil {
		slog.Error("panic recover", r)
	}
}
