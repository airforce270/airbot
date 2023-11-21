// Package cleanup provides a way to register functions for cleanup
// and run them.
package cleanup

import (
	"errors"
	"fmt"
)

// NewCleaner creates a new Cleaner.
func NewCleaner() Cleaner {
	return Cleaner{}
}

// Cleaner contains functions to be called to cleanup before program exit.
type Cleaner []Func

func (c *Cleaner) Register(f Func) {
	*c = append(*c, f)
}

func (c *Cleaner) Cleanup() error {
	var errs []error
	for _, f := range *c {
		if err := f.F(); err != nil {
			errs = append(errs, fmt.Errorf("cleanup function %s failed: %w", f.Name, err))
		}
	}
	return errors.Join(errs...)
}

// Func is a function that should be called before program exit.
type Func struct {
	// Name is the function's human-readable name.
	Name string
	// F is the function to be called.
	F func() error
}
