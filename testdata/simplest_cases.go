package testdata

import "sync"

type Test struct {
	prop int
	mu   sync.Mutex
}

func fun1(c chan struct{}) { // want "Function `fun1` uses channel `c` as send-only"
	c <- struct{}{}
}

func fun2(c2, ditched chan struct{}) struct{} { // want "Function `fun2` uses channel `c2` as receive-only"
	fun1(ditched)
	x := <-c2
	return x
}

func fun3(testChan3 chan bool) bool {
	testChan3 <- true
	x := <-testChan3
	return x
}

func fun4(c chan bool) { // want "Function `fun4` uses channel `c` as send-only"
	t := Test{prop: 3}
	t.mu.Lock()
	c <- true
	t.mu.Unlock()
}
