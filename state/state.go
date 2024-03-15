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

	highestBlock     map[string]*big.Int
	highestBlockTime map[string]time.Time

	mx sync.RWMutex
}

var (
	ErrExecutionEndpointDuplicateID = errors.New("duplicate execution endpoint id")
)

func New() *State {
	return &State{
		executionEndpoints: map[string]*ExecutionEndpoint{},

		highestBlock:     make(map[string]*big.Int),
		highestBlockTime: make(map[string]time.Time),
	}
}

func (s *State) RegisterExecutionEndpoint(id string) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	name, namespace, err := parseExecutionPointID(id)
	if err != nil {
		return err
	}

	if _, exists := s.executionEndpoints[id]; exists {
		return fmt.Errorf("%w: %s",
			ErrExecutionEndpointDuplicateID, id,
		)
	}

	if _, exists := s.highestBlock[namespace]; !exists {
		s.highestBlock[namespace] = big.NewInt(0)
	}
	if _, exists := s.highestBlockTime[namespace]; !exists {
		s.highestBlockTime[namespace] = time.Time{}
	}

	s.executionEndpoints[id] = newExecutionEndpoint(id, name, namespace)

	return nil
}

func (s *State) ExecutionEndpoint(id string) *ExecutionEndpoint {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.executionEndpoints[id]
}

func (s *State) IterateNamespaces(do func(namespace string, highestBlock *big.Int, highestBlockTime time.Time)) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	for namespace, highestBlock := range s.highestBlock {
		highestBlock = new(big.Int).Set(highestBlock)
		highestBlockTime := s.highestBlockTime[namespace]
		do(namespace, highestBlock, highestBlockTime)
	}
}

func (s *State) IterateExecutionEndpoints(do func(id string, ee *ExecutionEndpoint)) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	for id, ee := range s.executionEndpoints {
		do(id, ee)
	}
}

func (s *State) HighestBlock(namespace string) *big.Int {
	s.mx.RLock()
	defer s.mx.RUnlock()

	res := new(big.Int).Set(s.highestBlock[namespace])
	return res
}

func (s *State) HighestBlockTime(namespace string) time.Time {
	s.mx.RLock()
	defer s.mx.RUnlock()

	return s.highestBlockTime[namespace]
}

func (s *State) UpdateHighestBlockIfNeeded(
	namespace string,
	block *big.Int,
	blockTime time.Time,
) {
	s.mx.RLock()
	defer s.mx.RUnlock()

	// update the highest block
	if cmp := s.highestBlock[namespace].Cmp(block); cmp == -1 {
		s.mx.RUnlock()
		s.mx.Lock()
		if cmp := s.highestBlock[namespace].Cmp(block); cmp == -1 {
			s.highestBlock[namespace] = new(big.Int).Set(block)
			s.highestBlockTime[namespace] = blockTime
		}
		s.mx.Unlock()
		s.mx.RLock()
	}
}
