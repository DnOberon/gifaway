package split

import (
	"os/exec"
	"sync"
)

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

type Task struct {
	ID   int
	Name string

	Command  *exec.Cmd
	ErrorBag []error
}

func NewTask(id int, name string, command *exec.Cmd) *Task {
	return &Task{Command: command}
}

func (t *Task) Run(wg *sync.WaitGroup) {
	// was running into some issues with either all tasks executing, or only one
	// I'm sure I'm missing something obvious, but until that time this will solve the problem
	if err := t.Command.Start(); err != nil {
		t.ErrorBag = append(t.ErrorBag, err)
	}

	if err := t.Command.Wait(); err != nil {
		t.ErrorBag = append(t.ErrorBag, err)
	}

	wg.Done()
}
