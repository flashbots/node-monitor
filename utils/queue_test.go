package utils_test

import (
	"testing"

	"github.com/flashbots/node-monitor/utils"
	"gotest.tools/assert"
)

func TestSortedStringQueue(t *testing.T) {
	q := utils.NewSortedStringQueue(4)
	assert.Equal(t, "", q.InsertAndPop("4"))
	assert.Equal(t, "", q.InsertAndPop("3"))
	assert.Equal(t, "", q.InsertAndPop("5"))
	assert.Equal(t, "", q.InsertAndPop("2"))
	assert.Equal(t, "2", q.InsertAndPop("6"))
	assert.Equal(t, "1", q.InsertAndPop("1"))
	assert.Equal(t, "3", q.InsertAndPop("7"))
	assert.Equal(t, "0", q.InsertAndPop("0"))
	assert.Equal(t, "4", q.InsertAndPop("8"))
}
