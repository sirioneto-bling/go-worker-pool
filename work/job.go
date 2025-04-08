package work

// job implements Executor interface from the worker pool
type job struct {
	execute            func() (any, error)
	errorHandler       func(error)
	shouldReturnResult bool
}

// NewJob create's a new job
func NewJob(execute func() error, errorHandler func(error)) *job {
	return &job{
		execute:            func() (any, error) { return nil, execute() },
		errorHandler:       errorHandler,
		shouldReturnResult: false,
	}
}

// NewJobWithResult create's a new job instance with executor function returning result
func NewJobWithResult(execute func() (any, error), errorHandler func(error)) *job {
	return &job{
		execute:            execute,
		errorHandler:       errorHandler,
		shouldReturnResult: true,
	}
}

func (t *job) Execute() (any, error) {
	return t.execute()
}

func (t *job) OnError(err error) {
	t.errorHandler(err)
}

func (t *job) ShouldReturnResult() bool {
	return t.shouldReturnResult
}
