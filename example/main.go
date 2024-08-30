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
	strB := resB.Get()

	fmt.Println(strA)
	fmt.Println(strB)

	b.Shutdown()
}

func processor(in string) string {
	return strings.ToUpper(in)
}
