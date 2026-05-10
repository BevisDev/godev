package framework

import (
	"context"
	"fmt"
	"sync"
)

type State string

const (
	StateInit  State = "init"
	StateStart State = "start"
	StateStop  State = "stop"
)

type Phase struct {
	pre   []Hook
	post  []Hook
	mu    sync.Mutex
	done  bool
	state State
}

func NewPhase(state State) *Phase {
	return &Phase{
		state: state,
	}
}

// Before registers a hook to run before phase.
func (p *Phase) Before(fn Hook) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.done {
		return fmt.Errorf(prefixBootstrap+"state %s, already before executed", p.state)
	}
	p.pre = append(p.pre, fn)
	return nil
}

// After registers a hook to run after phase.
func (p *Phase) After(fn Hook) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.done {
		return fmt.Errorf(prefixBootstrap+"state %s, already after executed", p.state)
	}
	p.post = append(p.post, fn)
	return nil
}

func (p *Phase) runPre(ctx context.Context) error {
	for idx, fn := range p.pre {
		if err := fn(ctx); err != nil {
			return fmt.Errorf(prefixBootstrap+"%s pre-hook[%d]: %w", p.state, idx, err)
		}
	}
	return nil
}

func (p *Phase) runPost(ctx context.Context) error {
	for idx, fn := range p.post {
		if err := fn(ctx); err != nil {
			return fmt.Errorf(prefixBootstrap+"%s post-hook[%d]: %w", p.state, idx, err)
		}
	}
	return nil
}

func (p *Phase) markDone() {
	p.mu.Lock()
	p.done = true
	p.mu.Unlock()
}

func (p *Phase) isDone() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.done
}
