package state

import (
	"math/big"
	"sync"
	"time"
)

type ELEndpoint struct {
	name string

	highestBlock     *big.Int
	highestBlockTime time.Time

	mx sync.RWMutex
}

func newELEndpoint(name string) *ELEndpoint {
	return &ELEndpoint{
		name: name,

		highestBlock:     big.NewInt(0),
		highestBlockTime: time.Time{},
	}
}

func (e *ELEndpoint) RegisterBlock(
	block *big.Int,
	ts time.Time,
) {
	e.mx.RLock()
	defer e.mx.RUnlock()

	if cmp := e.highestBlock.Cmp(block); cmp == -1 {
		e.mx.RUnlock()
		e.mx.Lock()
		if cmp := e.highestBlock.Cmp(block); cmp == -1 {
			e.highestBlock = new(big.Int).Set(block)
			e.highestBlockTime = ts
		}
		e.mx.Unlock()
		e.mx.RLock()
	}
}

func (e *ELEndpoint) TimeSinceHighestBlock() (block int64, timeSince time.Duration) {
	e.mx.RLock()
	defer e.mx.RUnlock()

	b := e.highestBlock.Int64()
	t := time.Since(e.highestBlockTime)

	return b, t
}
