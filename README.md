Channel direction linter
====

Ideally whenever it's possible, I always prefer to declare the direction of channel operations used in function, e.g.:
```go
func ReadsFromChan(chanName <-chan Type) { }

func WritesToChan(chanName chan<- Type) { }
```
Therefore, we end up having send- or receive-only channel and compiler is going to throw an error in case of unintended misuse.

Apparently I wasn't able to easily find a linter which would warn me about opportunities to declare channel direction, 
so I decided to write one as toy project and mess around with Go's AST. 

**This is definitely not a production-grade tool. At least not yet! :)** I took some shortcuts here and there, 
plus I've strongly favoured having working PoC rather than "state of the art" Go code.

### Example usage

Just to prove it's working nicely I've cloned [Vitess repo](https://github.com/vitessio/vitess) and ran my linter against it:
```shell
vitess % ./channel-direction-linter ./...
/Users/anon/vitess/go/test/stress/stress.go:296:38: Function `startStressClient` uses channel `resultCh` as send-only, consider `func startStressClient(resultCh chan<- T`
/Users/anon/vitess/go/vt/vtgate/executor_framework_test.go:422:18: Function `getQueryLog` uses channel `logChan` as receive-only, consider `func getQueryLog(logChan <-chan T`
```
Aparrently the suggestions were accurate at that point of time :) See 
[stress.go:296](https://github.com/vitessio/vitess/blob/94861104f74f265a6b006197de12d494bf844ecc/go/test/stress/stress.go#L296) 
and 
[executor_framework_test.go:422](https://github.com/vitessio/vitess/blob/94861104f74f265a6b006197de12d494bf844ecc/go/vt/vtgate/executor_framework_test.go#L648)

### Known issues/limitations

1. Currently this linter ignores ellipsis (a.k.a. parameter lists):
    ```go
    func FuncName(chanArr ...chan bool) { }
    ```
2. If the channel is passed as an argument to a different function, it won't be analysed - e.g:
   ```go
    func aFunc(chan1, chan2 chan struct{}) struct{} {
	    x := <-testChan2
		anotherFunc(chan1) // this makes `chan1` usage analysis ignored
	    return x
    }
   ```
#### Misc

* Go AST viewer I've used -> http://goast.yuroyoro.net, super helpful! 
* A [very neat blog post](https://disaev.me/p/writing-useful-go-analysis-linter/) on writing a linter, which helped me with kicking off work on this one 