package concurrent

import "sync"

// Concurrent for simple concurrency control
type Concurrent struct {
	ch chan struct{}
	wg *sync.WaitGroup
}

// New is used to initial a concurrent control object
func New(limit int) *Concurrent {
	return &Concurrent{
		wg: &sync.WaitGroup{},
		ch: make(chan struct{}, limit),
	}
}

// Add is used to add a task
func (c *Concurrent) Add(n int) {
	c.wg.Add(n)
	for n > 0 {
		n--
		c.ch <- struct{}{}
	}
}

// Done is used to accomplish a task
func (c *Concurrent) Done() {
	c.wg.Done()
	<-c.ch
}

// Wait is used to wg for all tasks to be completed
func (c *Concurrent) Wait() {
	c.wg.Wait()
}
