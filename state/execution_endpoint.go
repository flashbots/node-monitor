package state

import (
	"math/big"
	"sync"
	"time"
)

type ExecutionEndpoint struct {
	id        string
	name      string
	namespace string

	highestBlock     *big.Int
	highestBlockTime time.Time

	mx sync.RWMutex
}

func newExecutionEndpoint(id, name, namespace string) *ExecutionEndpoint {
	return &ExecutionEndpoint{
		id:        id,
		name:      name,
		namespace: namespace,

		highestBlock:     big.NewInt(0),
		highestBlockTime: time.Time{},
	}
}

func (e *ExecutionEndpoint) Name() (name, namespace string) {
	e.mx.RLock()
	defer e.mx.RUnlock()

	return e.name, e.namespace
}

func (e *ExecutionEndpoint) HighestBlock() *big.Int {
	e.mx.RLock()
	defer e.mx.RUnlock()

	res := new(big.Int).Set(e.highestBlock)
	return res
}

func (e *ExecutionEndpoint) HighestBlockTime() time.Time {
	e.mx.RLock()
	defer e.mx.RUnlock()

	return e.highestBlockTime
}

func (e *ExecutionEndpoint) UpdateHighestBlockIfNeeded(
	block *big.Int,
	blockTime time.Time,
) {
	e.mx.RLock()
	defer e.mx.RUnlock()

	// update the highest block
	if cmp := e.highestBlock.Cmp(block); cmp == -1 {
		e.mx.RUnlock()
		e.mx.Lock()
		if cmp := e.highestBlock.Cmp(block); cmp == -1 {
			e.highestBlock = new(big.Int).Set(block)
			e.highestBlockTime = blockTime
		}
		e.mx.Unlock()
		e.mx.RLock()
	}
}
