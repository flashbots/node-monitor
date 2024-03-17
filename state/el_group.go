package state

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/flashbots/node-monitor/utils"
)

const (
	maxHistoryBlocks = 1024
)

type ELGroup struct {
	name string

	endpoints map[string]*ELEndpoint

	blocks     *utils.SortedStringQueue
	blockTimes map[string]time.Time

	highestBlock    *big.Int
	highestBlockStr string

	mx sync.RWMutex
}

func newELGroup(name string) *ELGroup {
	return &ELGroup{
		name: name,

		blocks:     utils.NewSortedStringQueue(maxHistoryBlocks),
		blockTimes: make(map[string]time.Time, maxHistoryBlocks+1),

		highestBlock:    big.NewInt(0),
		highestBlockStr: utils.Bigint2string(big.NewInt(0)),

		endpoints: make(map[string]*ELEndpoint),
	}
}

func (g *ELGroup) registerEndpoint(name string) error {
	id := utils.MakeELEndpointID(g.name, name)

	if _, exists := g.endpoints[name]; exists {
		return fmt.Errorf("%w: %s",
			ErrExecutionEndpointDuplicateID, id,
		)
	}
	g.endpoints[name] = newELEndpoint(id)

	return nil
}

func (g *ELGroup) Name() string {
	// constant, no need for mutex
	return g.name
}

func (g *ELGroup) Endpoint(name string) *ELEndpoint {
	g.mx.RLock()
	defer g.mx.RUnlock()

	return g.endpoints[name]
}

func (g *ELGroup) RegisterBlockAndGetLatency(block *big.Int, ts time.Time) time.Duration {
	g.mx.RLock()
	defer g.mx.RUnlock()

	blockStr := utils.Bigint2string(block)

	// update the highest block (if needed)
	if blockStr > g.highestBlockStr {
		g.mx.RUnlock()
		g.mx.Lock()
		if blockStr > g.highestBlockStr {
			delete(g.blockTimes, g.blocks.InsertAndPop(blockStr))
			g.highestBlock = big.NewInt(0).Set(block)
			g.highestBlockStr = blockStr
			g.blockTimes[blockStr] = ts
		}
		g.mx.Unlock()
		g.mx.RLock()
	}

	// fill in the gaps (in case of missed block)
	if prevTS, exists := g.blockTimes[blockStr]; !exists || ts.Before(prevTS) {
		g.mx.RUnlock()
		g.mx.Lock()
		if prevTS, exists := g.blockTimes[blockStr]; !exists || ts.Before(prevTS) {
			delete(g.blockTimes, g.blocks.InsertAndPop(blockStr))
			g.blockTimes[blockStr] = ts
		}
		g.mx.Unlock()
		g.mx.RLock()
	}

	return ts.Sub(g.blockTimes[blockStr])
}

func (g *ELGroup) TimeSinceHighestBlock() (block int64, timeSince time.Duration) {
	g.mx.RLock()
	defer g.mx.RUnlock()

	b := g.highestBlock.Int64()
	t := time.Since(g.blockTimes[g.highestBlockStr])

	return b, t
}

func (g *ELGroup) IterateEndpointsRO(
	do func(name string, e *ELEndpoint),
) {
	g.mx.RLock()
	defer g.mx.RUnlock()

	for name, e := range g.endpoints {
		do(name, e)
	}
}
