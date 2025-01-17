package main

import (
	"testing"
	"time"
)

func Test21(t *testing.T) {
	n21()
}

// ============== Concurrent recv on same channel ==============
// TP
func n21() {
	x := make(chan int)

	go func() {
		<-x
	}()

	go func() {
		<-x
	}()

	time.Sleep(100 * time.Millisecond)

	x <- 1
	x <- 1

	time.Sleep(300 * time.Millisecond)
}
