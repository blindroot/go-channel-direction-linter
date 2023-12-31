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