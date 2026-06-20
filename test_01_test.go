package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

var moneyChan = make(chan int)
var nameChan = make(chan string)
var doneChan = make(chan struct{})

func pay(name string, money int, wait *sync.WaitGroup) {
	fmt.Printf("%s start shopping\n", name)
	time.Sleep(1 * time.Second)
	fmt.Printf("%s end shopping\n", name)

	moneyChan <- money
	nameChan <- name
	wait.Done()
}

func TestOne(t *testing.T) {
	var wait sync.WaitGroup
	startTime := time.Now()
	wait.Add(4)
	go pay("Jake", 1, &wait)
	go pay("Bob", 2, &wait)
	go pay("Summer", 3, &wait)
	go pay("Mindy", 4, &wait)

	go func() {
		wait.Wait()
		// close(moneyChan)
		// close(nameChan)
		doneChan <- struct{}{}
	}()

	moneyList := make([]int, 0)
	nameList := make([]string, 0)

	var event = func() {
		for {
			select {
			case money := <-moneyChan:
				moneyList = append(moneyList, money)
			case name := <-nameChan:
				nameList = append(nameList, name)
			case <-doneChan:
				fmt.Println("done")
				return
			case <-time.After(3 * time.Second):
				fmt.Println("time out")
				return
			}
		}
	}
	event()

	fmt.Println("moneyChan", moneyList)
	fmt.Println("nameChan", nameList)

	fmt.Println("finished", time.Since(startTime))
}
