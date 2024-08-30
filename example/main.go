package main

import (
	"fmt"
	"strings"
	"time"

	microbatcher "github.com/callum-thomas/micro-batcher"
)

func main() {
	b := microbatcher.NewBatcher(processor, 5*time.Second, 2)

	go b.Start()

	resA, err := b.AddJob(microbatcher.Job[string]{Id: 1, Data: "input"})
	if err != nil {
		panic(err)
	}

	resB, err := b.AddJob(microbatcher.Job[string]{Id: 2, Data: "another input"})
	if err != nil {
		panic(err)
	}

	strA := resA.Get()
	if strA.err != nil {
		panic(err)
	}

	strB := resB.Get()
	if strA.err != nil {
		panic(err)
	}

	fmt.Println(strA.output)
	fmt.Println(strB.output)

	b.Shutdown()
}

type Result struct {
	output string
	err    error
}

func processor(in string) *Result {
	return &Result{output: strings.ToUpper(in), err: nil}
}
