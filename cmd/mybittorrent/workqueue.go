package main

import (
	"fmt"
	"sync"
)

type request int

type workqueue struct {
	queue chan int
}

func newWorkQueue(length int) *workqueue {
	return &workqueue{
		queue: make(chan int, length),
	}
}

func (w *workqueue) addItem(item int) {
	w.queue <- item
}

func (w *workqueue) processItem() (int, bool) {
	select {
	case x, ok := <-w.queue:
		return x, !ok
	default:
		return -1, true
	}
}

type worker struct {
	run func(int) error
}

type workerPool struct {
	workers   []*worker
	workqueue *workqueue
}

func newWorkerPool(work *workqueue, workers ...*worker) *workerPool {
	return &workerPool{
		workers:   workers,
		workqueue: work,
	}
}

func (w *workerPool) start() {
	var wg sync.WaitGroup
	for _, worker := range w.workers {
		wg.Add(1)
		go func() {
			for {
				item, empty := w.workqueue.processItem()
				if empty {
					// fmt.Println("worker returned", item)
					wg.Done()
					return
				}

				// fmt.Println("worker fetching piece idx", item)
				err := worker.run(item)
				if err != nil {
					fmt.Println("error from worker", err)
				}
			}
		}()
	}
	wg.Wait()
}
