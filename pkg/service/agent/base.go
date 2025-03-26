package agent

import (
	"context"
)

// Agent interface to execute external tools
type Agent interface {
	Execute(ctx context.Context, action string, params map[string]interface{}, files []string) ([]string, error)
}

// Registry save all agents
type Registry map[string]Agent

// NewRegistry constructor
func NewRegistry() Registry {
	return make(Registry)
}

// Register new agent
func (r Registry) Register(name string, agent Agent) {
	r[name] = agent
}

// Get to get agent
func (r Registry) Get(name string) (Agent, bool) {
	agent, exists := r[name]
	return agent, exists
}
