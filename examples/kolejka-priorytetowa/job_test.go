package main

import (
	"container/heap"
	"fmt"
	"os/exec"
	"time"
)

func ExampleJobs() {
	jobs := make(Jobs, 0, 100)
	zero := time.Now()
	heap.Init(&jobs)
	heap.Push(&jobs, &Job{
		ID:        1,
		Name:      "first",
		Timestamp: zero.Add(10 * time.Hour),
		Epsilon:   5 * time.Minute,
		Command:   exec.Command("ls", "-lha"),
	})
	heap.Push(&jobs, &Job{
		ID:        1,
		Name:      "second",
		Timestamp: zero.Add(10 * time.Hour),
		Epsilon:   4 * time.Minute,
		Command:   exec.Command("pwd"),
	})
	heap.Push(&jobs, &Job{
		ID:        1,
		Name:      "third",
		Timestamp: zero.Add(5 * time.Hour),
		Epsilon:   4 * time.Minute,
		Command:   exec.Command("ps", "aux"),
	})
	fmt.Println(jobs.Len())

	j1 := heap.Pop(&jobs).(*Job)
	j2 := heap.Pop(&jobs).(*Job)
	j3 := heap.Pop(&jobs).(*Job)

	fmt.Println(j1.Name)
	fmt.Println(j2.Name)
	fmt.Println(j3.Name)
	fmt.Println(jobs.Len())

	// Output:
	// 3
	// third
	// second
	// first
	// 0
}
