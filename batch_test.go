package microbatcher

import (
	"strings"
	"testing"
	"time"
)

// A large enough duration that basically prevents frequency
// based processing from occurring.
const FIVE_MINUTES = 5 * time.Minute

// A small enough duration to guarantee frequency based processing.
const ONE_MILLISECOND = 1 * time.Millisecond

func uppercaseString(in string) string {
	return strings.ToUpper(in)
}

func TestBatcherLifecycle(t *testing.T) {
	b := NewBatcher(uppercaseString, FIVE_MINUTES, 10)

	go b.Start()

	b.Shutdown()
}

func TestBatcherBatchSize(t *testing.T) {
	b := NewBatcher(uppercaseString, FIVE_MINUTES, 1)

	resA, err := b.AddJob(Job[string]{Id: 1, Data: "hello world"})
	if err != nil {
		t.Error("failed to add job A")
	}

	resB, err := b.AddJob(Job[string]{Id: 2, Data: "foobar"})
	if err != nil {
		t.Error("failed to add job B")
	}

	go b.Start()
	defer b.Shutdown()

	aData := resA.Get()
	if aData != "HELLO WORLD" {
		t.Error("failed to process job A correctly")
	}

	bData := resB.Get()
	if bData != "FOOBAR" {
		t.Error("failed to process job B correctly")
	}

	if len(b.jobs) != 0 {
		t.Error("non-zero jobs on the queue")
	}
}

func TestBatcherBatchSizeProcessesCorrectNumber(t *testing.T) {
	jobs := []Job[string]{
		{Id: 1, Data: "hello world"},
		{Id: 2, Data: "foobar"},
		{Id: 3, Data: "baz"},
	}
	b := NewBatcher(uppercaseString, FIVE_MINUTES, 2)

	go b.Start()
	defer b.Shutdown()

	for i, job := range jobs {
		_, err := b.AddJob(job)
		if err != nil {
			t.Errorf("failed to add job %d", i)
		}
	}

	// Allow time for processing to occur.
	time.Sleep(10 * time.Millisecond)

	if len(b.jobs) != 1 {
		t.Error("jobs processed prematurely")
	}
}

func TestBatcherTimout(t *testing.T) {
	b := NewBatcher(uppercaseString, ONE_MILLISECOND, 10)

	go b.Start()
	defer b.Shutdown()

	resA, err := b.AddJob(Job[string]{Id: 1, Data: "hello world"})
	if err != nil {
		t.Error("failed to add job A")
	}

	resB, err := b.AddJob(Job[string]{Id: 2, Data: "foobar"})
	if err != nil {
		t.Error("failed to add job B")
	}

	aStr := resA.Get()
	if aStr != "HELLO WORLD" {
		t.Errorf("failed to process job 1 correctly")
	}

	bStr := resB.Get()
	if bStr != "FOOBAR" {
		t.Errorf("failed to process job 1 correctly")
	}

	if len(b.jobs) != 0 {
		t.Error("unprocessed jobs on the queue")
	}
}

func TestBatcherTimeoutReset(t *testing.T) {
	b := NewBatcher(uppercaseString, 100*time.Millisecond, 10)

	_, err := b.AddJob(Job[string]{Id: 1, Data: "hello world"})
	if err != nil {
		t.Error("failed to add job A")
	}

	_, err = b.AddJob(Job[string]{Id: 2, Data: "foobar"})
	if err != nil {
		t.Error("failed to add job B")
	}

	go b.Start()
	defer b.Shutdown()

	// Wait to allow the ticker to fire.
	time.Sleep(150 * time.Millisecond)

	_, err = b.AddJob(Job[string]{Id: 3, Data: "baz"})
	if err != nil {
		t.Error("failed to add job C to the queue")
	}

	if len(b.jobs) != 1 {
		t.Error("incorrect number of jobs on the queue")
	}
}

func TestBatcherCannotAddJobWhenShuttingDown(t *testing.T) {
	b := NewBatcher(uppercaseString, FIVE_MINUTES, 10)

	go b.Start()
	b.Shutdown()

	_, err := b.AddJob(Job[string]{Id: 1, Data: "hello world"})
	if err == nil {
		t.Error("added job to queue of shutdown batcher.")
	}
}

func TestBatcherShutdownClearsQueue(t *testing.T) {
	jobs := []Job[string]{
		{Id: 1, Data: "hello world"},
		{Id: 2, Data: "foobar"},
		{Id: 3, Data: "baz"},
	}
	b := NewBatcher(uppercaseString, FIVE_MINUTES, 2)

	go b.Start()

	for i, job := range jobs {
		_, err := b.AddJob(job)
		if err != nil {
			t.Errorf("failed to add job %d", i)
		}
	}

	// Allow time for processing to occur.
	time.Sleep(10 * time.Millisecond)

	// Add another job before shutting down.
	resD, err := b.AddJob(Job[string]{Id: 4, Data: "job 4"})
	if err != nil {
		t.Error("failed to add job 4")
	}

	b.Shutdown()

	// Allow time for processing.
	time.Sleep(10 * time.Millisecond)

	strD := resD.Get()
	if strD != "JOB 4" {
		t.Error("failed to process final job properly")
	}

	if len(b.jobs) != 0 {
		t.Error("shutdown did not clear remaining jobs.")
	}
}

func TestBatcherJobResultReaccessingOutput(t *testing.T) {

	b := NewBatcher(uppercaseString, FIVE_MINUTES, 1)

	go b.Start()
	defer b.Shutdown()

	resA, err := b.AddJob(Job[string]{Id: 4, Data: "hello world"})
	if err != nil {
		t.Error("failed to add job 4")
	}

	strA := resA.Get()
	reaccess := resA.Get()

	if strA != reaccess {
		t.Error("reaccessing result output does not match")
	}
}
