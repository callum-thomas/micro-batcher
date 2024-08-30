# Micro-Batcher
Microbatcher is a library to do batch processing of small tasks.

## Usage
This library can be used by:
  1. Creating a processor function
  2. Creating a `Batcher` that receives the processor function and configuration for batch sizing and frequency
  3. Starting the `Batcher` in a seperate goroutine
  4. Adding `Job`s to the queue for processing
  5. Retreiving the results for each `Job`
  6. Shutting down the `Batcher` when all jobs have been submitted.

  ```golang
    package main

    import (
        "fmt"
        "time"
    )

    func processor(in string) string {
        return fmt.Printf("%s updated\n", in)
    }

    func main() {
        b := microbatcher.NewBatcher(processor, 5 * time.Second, 2)

        go b.Start()
        defer b.Shutdown()

        resultA, err := b.AddJob(microbatcher.Job[string]{Id: 1, Data: "foobar"})
        if err != nil {
            // Handle retry logic here if needed
        }

        outputA := resultA.Get()
        fmt.Println(outputA) // FOOBAR
    }
  ```

For convenience, an example implementation is included in the `example` directory.

## Testing
Tests have been provided and can be run with either:
  * `go test .`
  * `make test`

Optionally, test coverage can be viewed in your browser with `make test && make coverage`.