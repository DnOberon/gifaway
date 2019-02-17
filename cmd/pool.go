package cmd

import "sync"

// Simple worker pool structure using channels
type Pool struct {
	Tasks []*Task

	concurrency int
	tasksChan   chan *Task
	wg          sync.WaitGroup
}

func NewPool(tasks []*Task, concurrency int) *Pool {
	return &Pool{
		Tasks:       tasks,
		concurrency: concurrency,
		tasksChan:   make(chan *Task),
	}
}

func (p *Pool) Run() {
	for i := 0; i < p.concurrency; i++ {
		go p.work()
	}

	p.wg.Add(len(p.Tasks))
	for _, task := range p.Tasks {
		p.tasksChan <- task
	}

	close(p.tasksChan)

	p.wg.Wait()
}

func (p *Pool) work() {
	for task := range p.tasksChan {
		task.Run(&p.wg)
	}
}

// Generic Task structure, allowing freedom of func definition
type Task struct {
	Err error
	f   func() error
}

func NewTask(f func() error) *Task {
	return &Task{f: f}
}

func (t *Task) Run(wg *sync.WaitGroup) {
	t.Err = t.f()
	wg.Done()
}
