package workflow

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

type Step[T any] interface {
	Run(context.Context, *T) (*T, error)
}

func Name[T any](s Step[T]) string {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var z [0]T // zero alloc
	return strings.Replace(t.Name(), reflect.TypeOf(z).Elem().PkgPath()+".", "", 1)
}

type typ struct{}

var (
	_ Step[typ] = (*Pipeline[typ])(nil)
	_ Step[typ] = (*MidFunc[typ])(nil)
	_ Step[typ] = (*series[typ])(nil)
	_ Step[typ] = (*parallel[typ])(nil)
	_ Step[typ] = (*ifElse[typ])(nil)
)

type Pipeline[T any] struct {
	Steps []Step[T]
	Mid[T]
}

func (p *Pipeline[T]) Run(ctx context.Context, req *T) (*T, error) {
	resp := req
	var err error
	for i := range p.Steps {
		for _, m := range slices.Backward(p.Mid) {
			p.Steps[i] = m(p.Steps[i])
		}
		ctx = setStepID(ctx, gen.ID())
		resp, err = p.Steps[i].Run(ctx, req)
		if err != nil {
			return nil, err
		}
		req = resp
	}
	return resp, nil
}

func NewPipeline[T any](mid ...Middleware[T]) *Pipeline[T] {
	return &Pipeline[T]{
		Mid:   mid,
		Steps: make([]Step[T], 0),
	}
}

// StepFunc

type StepFunc[T any] func(context.Context, *T) (*T, error)

func (f StepFunc[T]) Run(ctx context.Context, res *T) (*T, error) {
	return f(setStepID(ctx, gen.ID()), res)
}

func (f StepFunc[T]) String() string {
	var z T
	return fmt.Sprintf("StepFunc[%T]", z)
}

// Middleware

type MidFunc[T any] func(context.Context, *T) (*T, error)

func (f MidFunc[T]) Run(ctx context.Context, req *T) (*T, error) {
	return f(ctx, req)
}

type Middleware[T any] func(s Step[T]) Step[T]

type Mid[T any] []Middleware[T]

// Selector

type Selector[T any] func(context.Context, *T) bool

type ifElse[T any] struct {
	s        Selector[T]
	ifStep   Step[T]
	elseStep Step[T]
	Mid[T]
}

func (s ifElse[T]) String() string {
	var z T
	return fmt.Sprintf("IfElse Step[%T] { IF: %v, ELSE: %v}", z, s.ifStep, s.elseStep)
}

func (p *Pipeline[T]) IfElse(s Selector[T], ifStep, elseStep Step[T]) Step[T] {
	return &ifElse[T]{
		s:        s,
		ifStep:   ifStep,
		elseStep: elseStep,
		Mid:      p.Mid,
	}
}

func (s ifElse[T]) Run(ctx context.Context, r *T) (*T, error) {
	var step Step[T]
	if s.s(ctx, r) {
		step = s.ifStep
	}
	if s.elseStep != nil {
		step = s.elseStep
	}
	if step == nil {
		return nil, fmt.Errorf("selector chosed missing else branch: %v", r)
	}
	for _, m := range slices.Backward(s.Mid) {
		step = m(step)
	}
	return step.Run(setStepID(ctx, gen.ID()), r)
}

// Series

type series[T any] struct {
	Stages []Step[T]
	Mid[T]
}

func (s *series[T]) String() string {
	if s == nil {
		return "none"
	}
	tt := make([]string, 0, len(s.Stages))
	for _, t := range s.Stages {
		tt = append(tt, Name(t))
	}
	return fmt.Sprintf("Serie{Stages: [%s]}", strings.Join(tt, ","))
}

// Series executes a series of steps in sequential order.
func (p *Pipeline[T]) Series(steps ...Step[T]) *series[T] {
	return &series[T]{
		Stages: steps,
		Mid:    p.Mid,
	}
}

func (s *series[T]) Run(ctx context.Context, req *T) (*T, error) {
	var err error
	resp := req

	for i := range s.Stages {
		for _, m := range slices.Backward(s.Mid) {
			s.Stages[i] = m(s.Stages[i])
		}
		ctx = setStepID(ctx, gen.ID())
		resp, err = s.Stages[i].Run(ctx, req)
		if err != nil {
			return resp, err
		}
		req = resp
	}
	return resp, nil
}

// Parallel

type parallel[T any] struct {
	merge MergeRequest[T]
	Tasks []Step[T]
	Mid[T]
}

func (p *parallel[T]) String() string {
	if p == nil {
		return "none"
	}
	tt := make([]string, 0, len(p.Tasks))
	for _, t := range p.Tasks {
		tt = append(tt, Name(t))
	}
	return fmt.Sprintf("Parallel{Tasks: [%s]}", strings.Join(tt, ", "))
}

type MergeRequest[T any] func(context.Context, *T, ...*T) (*T, error)

// Parallel executes a list of steps in parallel.
// Once all the steps are done, the merge request [MergeRequest] will combine all the results into one struct T.
func (p *Pipeline[T]) Parallel(merge MergeRequest[T], steps ...Step[T]) *parallel[T] {
	return &parallel[T]{
		merge: merge,
		Tasks: steps,
		Mid:   p.Mid,
	}
}

func (p *parallel[T]) Run(ctx context.Context, req *T) (*T, error) {
	tasks := make([]Step[T], len(p.Tasks))
	for i, s := range p.Tasks {
		tasks[i] = s
		for _, m := range slices.Backward(p.Mid) {
			tasks[i] = m(tasks[i])
		}
	}
	g, groupCtx := errgroup.WithContext(ctx)
	resps := make([]*T, len(p.Tasks))
	for i := range tasks {
		g.Go(func() error {
			defer CapturePanic(groupCtx)

			copyReq := new(T)
			*copyReq = *req
			resp, err := tasks[i].Run(setStepID(groupCtx, gen.ID()), copyReq)
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

func MergeTransform[T any](t ...func(*mergo.Config)) MergeRequest[T] {
	return func(ctx context.Context, res *T, responses ...*T) (*T, error) {
		var err error
		for _, r := range responses {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("aborting: %w", ctx.Err())
			default:

				err = mergo.Merge(res, r, t...)
				if err != nil {
					return nil, err
				}
			}
		}
		return res, nil
	}
}

func Merge[T any](ctx context.Context, req *T, responses ...*T) (*T, error) {
	return MergeTransform[T]()(ctx, req, responses...)
}

func CapturePanic(ctx context.Context) {
	if r := recover(); r != nil {
		slog.Error("panic recover", r)
	}
}
