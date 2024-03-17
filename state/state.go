package state

import (
	"errors"
	"sync"
)

type State struct {
	executionGroups map[string]*ELGroup

	mx sync.RWMutex
}

var (
	ErrExecutionEndpointDuplicateID = errors.New("duplicate execution endpoint id")
)

func New() *State {
	return &State{
		executionGroups: make(map[string]*ELGroup),
	}
}

func (s *State) RegisterExecutionEndpoint(group, name string) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	if _, exists := s.executionGroups[group]; !exists {
		s.executionGroups[group] = newELGroup(group)
	}

	if err := s.executionGroups[group].registerEndpoint(name); err != nil {
		return err
	}

	return nil
}

func (s *State) ExecutionGroup(group string) *ELGroup {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.executionGroups[group]
}

func (s *State) IterateELGroupsRO(
	do func(name string, g *ELGroup),
) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	for name, g := range s.executionGroups {
		do(name, g)
	}
}
