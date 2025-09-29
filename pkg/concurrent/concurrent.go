package concurrent

import (
	"fmt"
	"sync"

	"lazy-lagoon/pkg/types"
)

// ForEach executes the provided function for each item in the slice in parallel,
// but ensures that results are processed in the original order.
// This is useful when you want to parallelize operations but maintain order in the output.
func ForEach[T any, R any](items []T, fn func(item T, index int) (R, *types.TransformError)) ([]R, *types.TransformError) {
	var wg sync.WaitGroup
	results := make([]R, len(items))
	errChan := make(chan *types.TransformError, len(items))
	
	// Process each item in parallel
	for i, item := range items {
		wg.Add(1)
		go func(i int, item T) {
			defer func() {
				if r := recover(); r != nil {
					errChan <- &types.TransformError{Message: fmt.Sprintf("panic: %v", r)}
				}
			}()
			defer wg.Done()
			result, err := fn(item, i)
			if err != nil {
				errChan <- err
				return
			}
			// Store result in the same position as the input
			results[i] = result
		}(i, item)
	}
	
	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)
	
	// Check if any errors occurred
	if len(errChan) > 0 {
		return nil, <-errChan
	}
	
	return results, nil
}

// ForEachVoid executes the provided function for each item in the slice in parallel,
// but does not return any results. This is useful for operations that only have side effects.
func ForEachVoid[T any](items []T, fn func(item T, index int) *types.TransformError) *types.TransformError {
	var wg sync.WaitGroup
	errChan := make(chan *types.TransformError, len(items))
	
	// Process each item in parallel
	for i, item := range items {
		wg.Add(1)
		go func(i int, item T) {
			defer func() {
				if r := recover(); r != nil {
					errChan <- &types.TransformError{Message: fmt.Sprintf("panic: %v", r)}
				}
			}()
			defer wg.Done()
			if err := fn(item, i); err != nil {
				errChan <- err
				return
			}
		}(i, item)
	}
	
	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)
	
	// Check if any errors occurred
	if len(errChan) > 0 {
		return <-errChan
	}
	
	return nil
}

