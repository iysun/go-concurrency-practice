package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

var sum int = 0
var wait sync.WaitGroup
var lock sync.Mutex

func add() {
	lock.Lock()
	for i := 0; i < 100_000; i++ {
		sum++
	}
	lock.Unlock()
	wait.Done()
}

func sub() {
	lock.Lock()
	for i := 0; i < 100_000; i++ {
		sum--
	}
	lock.Unlock()
	wait.Done()
}

func TestTwo(t *testing.T) {
	startTime := time.Now()
	wait.Add(2)
	go add()
	go sub()
	wait.Wait()
	fmt.Println(sum)
	fmt.Println(time.Since(startTime))
}
