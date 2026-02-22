package workerpool

import (
	"context"
	"sync"
)

// task

type Task[P any, R any] struct {
	ID         int
	Func       func(ctx context.Context, params P) (R, error)
	Parameters P
}

type TaskResponse[R any] struct {
	TaskId int
	Result R
	Err    error
}

func (t *Task[P, R]) Process(ctx context.Context) *TaskResponse[R] {
	result, err := t.Func(ctx, t.Parameters)
	return &TaskResponse[R]{
		TaskId: t.ID,
		Result: result,
		Err:    err,
	}
}
func NewTask[P any, R any](f func(ctx context.Context, params P) (R, error), params P, id int) *Task[P, R] {
	return &Task[P, R]{
		ID:         id,
		Func:       f,
		Parameters: params,
	}
}

// worker

type Worker[P any, R any] struct {
	ID           int
	taskChan     <-chan *Task[P, R]
	taskRespChan chan<- *TaskResponse[R]
}

func NewWorker[P any, R any](tasksCh chan *Task[P, R], respCh chan *TaskResponse[R], ID int) *Worker[P, R] {
	return &Worker[P, R]{
		ID:           ID,
		taskChan:     tasksCh,
		taskRespChan: respCh,
	}
}

func (wr *Worker[P, R]) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		for task := range wr.taskChan {
			wr.taskRespChan <- task.Process(ctx)
		}
	}()
}

// pool

type Pool[P any, R any] struct {
	Tasks        []*Task[P, R]
	workerCount  int
	taskChan     chan *Task[P, R]
	taskRespChan chan *TaskResponse[R]
}

func NewPool[P any, R any](tasks []*Task[P, R], workerCount int) *Pool[P, R] {
	return &Pool[P, R]{
		Tasks:        tasks,
		workerCount:  workerCount,
		taskChan:     make(chan *Task[P, R], len(tasks)),
		taskRespChan: make(chan *TaskResponse[R], len(tasks)),
	}
}

func (p *Pool[P, R]) Run(ctx context.Context) []*TaskResponse[R] {
	var wg sync.WaitGroup

	// 1. Start Response collector
	responses := make([]*TaskResponse[R], 0, len(p.Tasks))
	var respChanSignal = make(chan struct{})
	go func() {
		defer close(respChanSignal)
		for resp := range p.taskRespChan {
			responses = append(responses, resp)
		}
	}()

	// 2. Start workers
	for i := 0; i < p.workerCount; i++ {
		worker := NewWorker(p.taskChan, p.taskRespChan, i)
		worker.Start(ctx, &wg)
	}

	// 3. Send tasks to workers
	for _, task := range p.Tasks {
		p.taskChan <- task
	}
	close(p.taskChan)

	// 4. Wait for workers&response collector to finish
	wg.Wait()
	close(p.taskRespChan)
	<-respChanSignal

	return responses
}

func (p *Pool[P, R]) RunStream(ctx context.Context) <-chan *TaskResponse[R] {
	var wg sync.WaitGroup

	// 1. Start workers
	for i := 0; i < p.workerCount; i++ {
		worker := NewWorker(p.taskChan, p.taskRespChan, i)
		worker.Start(ctx, &wg)
	}

	// 2. Send tasks to workers
	go func() {
		defer close(p.taskChan)
		for _, task := range p.Tasks {
			select {
			case <-ctx.Done():
				return
			case p.taskChan <- task:
			}
		}
	}()

	// 3. Wait for all workers to finish
	go func() {
		wg.Wait()
		close(p.taskRespChan)
	}()

	return p.taskRespChan
}
