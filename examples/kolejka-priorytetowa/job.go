package main

import (
	"os/exec"
	"time"
)

type Job struct {
	ID        int64
	Index     int64
	Name      string
	Timestamp time.Time
	Epsilon   time.Duration
	Command   *exec.Cmd
}

type Jobs []*Job

// Len implements sort Interface.
func (j Jobs) Len() int {
	return len(j)
}

// Less implements sort Interface.
func (j Jobs) Less(n, m int) bool {
	if j[n].Timestamp.Equal(j[m].Timestamp) {
		return j[n].Epsilon < j[m].Epsilon
	}
	return j[n].Timestamp.Before(j[m].Timestamp)
}

// Swap implements sort Interface.
func (j Jobs) Swap(n, m int) {
	j[n], j[m] = j[m], j[n]
	j[n].Index = int64(n)
	j[m].Index = int64(m)
}

// Push implements heap Interface.
func (j *Jobs) Push(x interface{}) {
	n := len(*j)
	item := x.(*Job)
	item.Index = int64(n)
	*j = append(*j, item)
}

// Pop implements heap Interface.
func (j *Jobs) Pop() interface{} {
	old := *j
	n := len(old)
	item := old[n-1]
	item.Index = -1
	*j = old[0 : n-1]
	return item
}
