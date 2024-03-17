package utils

import "sort"

type SortedStringQueue struct {
	size  int
	items []string
}

func NewSortedStringQueue(size int) *SortedStringQueue {
	return &SortedStringQueue{
		size:  size,
		items: make([]string, 0, size+1),
	}
}

func (q *SortedStringQueue) InsertAndPop(item string) string {
	i := sort.SearchStrings(q.items, item)
	q.items = append(q.items, "")
	copy(q.items[i+1:], q.items[i:])
	q.items[i] = item

	if len(q.items) <= q.size {
		return ""
	}

	pop := q.items[0]
	copy(q.items, q.items[1:])
	q.items = q.items[:q.size]
	return pop
}
