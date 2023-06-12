package corgroup

import (
	"fmt"
	"sync"
)

type GroupError struct {
	Errors      []error
	ErrorsCount int
}

func (g *GroupError) Error() string {
	return fmt.Sprintf("group ended execution with %d errors in subtasks", len(g.Errors))
}

// A Group is a collection of goroutines working on subtasks that are part of
// the same overall task.
type Group struct {
	wg               sync.WaitGroup
	errorsCount      int
	errs             []error
	thrownErrorCount int
	mutex            sync.Mutex
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the errors from function calls
func (g *Group) Wait() error {
	g.wg.Wait()
	if g.thrownErrorCount != 0 {
		return &GroupError{Errors: g.errs, ErrorsCount: g.thrownErrorCount}
	}
	return nil
}

// Go calls the given function in a new goroutine.
func (g *Group) Go(f func() error) {
	errIndex := g.errorsCount
	g.errorsCount++
	g.mutex.Lock()
	g.errs = append(g.errs, nil)
	g.mutex.Unlock()
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()
		if err := f(); err != nil {
			g.thrownErrorCount++
			g.errs[errIndex] = err
		}
	}()
}
