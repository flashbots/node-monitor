package state

import (
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"
)

type State struct {
	executionEndpoints map[string]*ExecutionEndpoint

	highestBlock     *big.Int
	highestBlockTime time.Time

	mx sync.RWMutex
}

var (
	ErrExecutionEndpointDuplicateID = errors.New("duplicate execution endpoint id")
)

func New() *State {
	return &State{
		executionEndpoints: map[string]*ExecutionEndpoint{},

		highestBlock:     big.NewInt(0),
		highestBlockTime: time.Time{},
	}
}

func (s *State) RegisterExecutionEndpoint(id string) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	if _, exists := s.executionEndpoints[id]; exists {
		return fmt.Errorf("%w: %s",
			ErrExecutionEndpointDuplicateID, id,
		)
	}

	s.executionEndpoints[id] = newExecutionEndpoint(id)

	return nil
}

func (s *State) ExecutionEndpoint(id string) *ExecutionEndpoint {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.executionEndpoints[id]
}

func (s *State) IterateExecutionEndpoints(do func(id string, ee *ExecutionEndpoint)) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	for id, ee := range s.executionEndpoints {
		do(id, ee)
	}
}

func (s *State) HighestBlock() *big.Int {
	s.mx.RLock()
	defer s.mx.RUnlock()

	res := new(big.Int).Set(s.highestBlock)
	return res
}

func (s *State) HighestBlockTime() time.Time {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.highestBlockTime
}

func (s *State) UpdateHighestBlockIfNeeded(block *big.Int, blockTime time.Time) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	// update the highest block
	if cmp := s.highestBlock.Cmp(block); cmp == -1 {
		s.mx.RUnlock()
		s.mx.Lock()
		if cmp := s.highestBlock.Cmp(block); cmp == -1 {
			s.highestBlock = new(big.Int).Set(block)
			s.highestBlockTime = blockTime
		}
		s.mx.Unlock()
		s.mx.RLock()
	}
}
