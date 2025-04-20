package work

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPool(t *testing.T) {
	tests := []struct {
		name          string
		numWorkers    int
		bufferSize    int
		expectedError bool
	}{
		{"Valid_Parameters", 10, 20, false},
		{"Zero_Workers", 0, 20, true},
		{"Negative_Workers", -1, 20, true},
		{"Negative_Buffer", 10, -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, err := NewPool(tt.numWorkers, tt.bufferSize)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, pool)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pool)
			}
		})
	}
}

func TestPoolExecution(t *testing.T) {
	pool, _ := NewPool(2, 4)
	ctx := context.Background()

	successJob := NewJobWithResult(func() (any, error) {
		return "success", nil
	}, func(err error) {})

	errorJob := NewJobWithResult(func() (any, error) {
		return nil, errors.New("test error")
	}, func(err error) {})

	pool.Start(ctx)

	go func() {
		pool.AddJob(successJob)
		pool.AddJob(errorJob)

		pool.WaitJobs()
		pool.Stop()
	}()

	var results []Result
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for result := range pool.Result() {
			results = append(results, result)
		}
	}()

	wg.Wait()

	assert.Equal(t, 2, len(results))

	var successCount, errorCount int
	for _, result := range results {
		switch {
		case result.Error == nil:
			successCount++
			assert.Equal(t, "success", result.Value)
		case result.Error.Error() == errors.New("test error").Error():
			errorCount++
		}
	}

	assert.Equal(t, 1, successCount, "Should have one successful job")
	assert.Equal(t, 1, errorCount, "Should have one error job")
}
