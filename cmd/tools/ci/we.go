package main

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"strings"

	"dario.cat/mergo"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type Pipeline struct {
	Steps []Step
	Mid
}

func (p *Pipeline) Run(ctx context.Context, req *Request) (*Request, error) {
	// steps := make([]Step, len(p.Steps))
	// for i, s := range p.Steps {
	// 	steps[i] = s
	// 	for _, m := range slices.Backward(p.Mid) {
	// 		steps[i] = m(steps[i])
	// 	}
	// }
	//
	// resp := req
	// var err error
	// for _, s := range steps {
	// 	resp, err = s.Run(ctx, req)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	req = resp
	// }

	resp := req
	var err error
	for i := range p.Steps {
		for _, m := range slices.Backward(p.Mid) {
			p.Steps[i] = m(p.Steps[i])
		}

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
		Steps: make([]Step, 0),
		Mid:   append([]Middleware{initStateReq()}, mid...),
	}
}

type Step interface {
	Run(ctx context.Context, req *Request) (*Request, error)
}

func Name(s Step) string {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return strings.ToUpper(t.Name())
}

type Request struct {
	Current *State
	History []*State
}

func (r *Request) String() string {
	sb := strings.Builder{}
	sb.WriteString("Request{Current:")
	sb.WriteString(r.Current.String())
	sb.WriteString(",History:[]State{")
	for i, h := range r.History {
		sb.WriteString(strings.TrimPrefix(h.String(), "State"))
		if i == len(r.History)-2 {
			sb.WriteString(",")
		}
	}
	sb.WriteString("}}")
	return sb.String()
}

type State struct {
	ID uuid.UUID
}

func NewState() *State {
	id := uuid.Must(uuid.NewV7())
	return &State{ID: id}
}

func (s *State) String() string {
	if s == nil {
		return "nil"
	}
	return fmt.Sprintf("State{ID:%q}", s.ID)
}

type StepFunc func(context.Context, *Request) (*Request, error)

func (f StepFunc) Run(ctx context.Context, req *Request) (*Request, error) {
	return f(ctx, req)
}

type Middleware func(s Step) Step

type Mid []Middleware

func LoggerMiddleware(l *slog.Logger) Middleware {
	return func(next Step) Step {
		return StepFunc(func(ctx context.Context, req *Request) (*Request, error) {
			// before the step
			var id uuid.UUID
			if req != nil {
				id = req.Current.ID
			}
			l.Info("step start", "Type", Name(next), "ID", id.String(), "Request", req)

			resp, err := next.Run(ctx, req)

			// after the step
			l.Info("step done", "Type", Name(next), "ID", id.String(), "Response", resp)
			return resp, err
		})
	}
}

func initStateReq() Middleware {
	return func(next Step) Step {
		return StepFunc(func(ctx context.Context, req *Request) (*Request, error) {
			id := uuid.Must(uuid.NewV7())
			if req == nil {
				req = &Request{Current: &State{ID: id}}
			}
			if req != nil && req.Current == nil {
				req.Current = &State{ID: id}
			}
			o, err := next.Run(ctx, req)
			req.History = append(req.History, req.Current)
			req.Current = NewState()
			return o, err
		})
	}
}

type series struct {
	Stages []Step
	Mid
}

func Series(steps ...Step) *series {
	return &series{Stages: steps}
}

func (s *series) Run(ctx context.Context, req *Request) (*Request, error) {
	var err error
	resp := req

	for _, stage := range s.Stages {
		resp, err = stage.Run(ctx, req)
		if err != nil {
			return nil, err
		}
		req = resp
	}
	return resp, nil
}

type parallel struct {
	Tasks []Step
}

func Parallel(steps ...Step) *parallel {
	return &parallel{Tasks: steps}
}

func (p *parallel) Run(ctx context.Context, req *Request) (*Request, error) {
	resps := make([]*Request, len(p.Tasks))
	g, groupCtx := errgroup.WithContext(ctx)

	for i := range p.Tasks {
		g.Go(func() error {
			defer CapturePanic(groupCtx)

			copyReq := &Request{}
			*copyReq = *req
			resp, err := p.Tasks[i].Run(groupCtx, copyReq)
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

func (p *parallel) merge(ctx context.Context, req *Request, responses ...*Request) (*Request, error) {
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
		}
	}
	return req, nil
}

func CapturePanic(ctx context.Context) {
	if r := recover(); r != nil {
		slog.Error("panic recover", r)
	}
}
