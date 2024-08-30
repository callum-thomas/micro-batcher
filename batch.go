package microbatcher

import (
	"errors"
	"sync"
	"time"
)

// Job represents a job to be processed by the Batcher.
type Job[A any] struct {
	// Id for the job. This should be unique for each job.
	Id int
	// Data required to process the job.
	Data A
}

type JobResult[B any] struct {
	JobId int
	data  *B
	ch    chan B
}

// Get reads the result of the job from the channel and returns.
func (jr *JobResult[B]) Get() B {
	if jr.data != nil {
		return *jr.data
	}

	val := <-jr.ch
	jr.data = &val

	return val
}

// batchJob is an intermediate structure to hold the original Job and
// the channel to return a JobResult.
type batchJob[A any, B any] struct {
	job   *Job[A]
	retCh chan B
}

// Batcher represents a unit that receives jobs and processes them in
// configurable batches.
type Batcher[A any, B any] struct {
	// Function that processes the jobs in the batcher.
	processor func(A) B
	// Minimum size for a batch of jobs to be processed before timeout.
	batchSize int
	// The frequency with which job batches should be processed if
	// there are inadequate jobs in the queue.
	frequency time.Duration
	// Status of Batcher shutdown.
	shuttingDown bool
	// Queue of jobs to be processed.
	jobs []batchJob[A, B]
	// Ticker to control time-based batch processing.
	ticker *time.Ticker

	mu sync.Mutex
}

// NewBatcher constructs a new Batcher configured with the given processor,
// frequency and batch size.
func NewBatcher[A any, B any](processor func(A) B, frequency time.Duration, batchSize int) *Batcher[A, B] {
	return &Batcher[A, B]{
		processor:    processor,
		batchSize:    batchSize,
		frequency:    frequency,
		shuttingDown: false,
		jobs:         []batchJob[A, B]{},
		ticker:       time.NewTicker(frequency),
	}
}

// AddJob adds the submitted job to the queue of the Batcher to be processed.
// An error is returned if the Batcher is in the process of shutting down,
// and is thus not able to accept new jobs.
func (b *Batcher[A, B]) AddJob(job Job[A]) (*JobResult[B], error) {
	if b.shuttingDown {
		return nil, errors.New("failed to add job; batcher is shutting down")
	}

	ch := make(chan B, 1)
	newJob := batchJob[A, B]{job: &job, retCh: ch}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.jobs = append(b.jobs, newJob)

	return &JobResult[B]{JobId: job.Id, ch: ch, data: nil}, nil
}

// Start begins the processing of jobs by the Batcher, generally run as a
// goroutine.
func (b *Batcher[A, B]) Start() {
	// Start the ticker based processing.
	go b.startTicker()

	// Start batch size processing.
	for {
		switch {
		case b.shuttingDown:
			b.mu.Lock()
			defer b.mu.Unlock()

			// Process all remaining jobs on the queue if any exist.
			if len(b.jobs) > 0 {
				b.processBatch(b.jobs)

				b.jobs = []batchJob[A, B]{}
			}

			return
		case len(b.jobs) >= b.batchSize:
			b.mu.Lock()

			// Create slice of jobs to be processed and update
			// job queue.
			batchJobs := b.jobs[0:b.batchSize]
			b.jobs = b.jobs[b.batchSize:]

			// Process the first batchSize jobs in the queue.
			b.processBatch(batchJobs)

			// Reset the ticker.
			b.ticker.Reset(b.frequency)

			// Release the mutex lock.
			b.mu.Unlock()
		}
	}
}

// Shutdown triggers the graceful shutdown of the Batcher, flushing all remaining
// jobs from the queue before ceasing to process.
func (b *Batcher[A, B]) Shutdown() {
	b.shuttingDown = true
}

func (b *Batcher[A, B]) startTicker() {
	for {
		<-b.ticker.C
		b.mu.Lock()
		b.processBatch(b.jobs)
		b.jobs = []batchJob[A, B]{}
		b.mu.Unlock()
	}
}

func (b *Batcher[A, B]) processBatch(batch []batchJob[A, B]) {
	for _, job := range batch {
		go b.processJob(job)
	}
}

func (b *Batcher[A, B]) processJob(job batchJob[A, B]) {
	res := b.processor(job.job.Data)
	job.retCh <- res
}
