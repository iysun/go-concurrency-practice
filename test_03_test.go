package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestThree(t *testing.T) {
	maps := sync.Map{}
	go func() {
		for {
			maps.Store(1, "Summer")
		}
	}()

	go func() {
		for {
			val, ok := maps.Load(1)
			fmt.Println(val, ok)
		}
	}()
	<-time.After(3 * time.Second)
}
