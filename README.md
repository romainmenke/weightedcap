# Weightedcap

`$ go get -u github.com/romainmenke/weightedcap`


This pkg is useful if you like semaphores but you have heterogynous jobs.

A simple example is a process that uses a significant amount of memory but not the same amount each time. If an estimate of memory usage is possible you can use this pkg to give a weight to each job. `Consume(ctx, n)` will block until enough capacity is available or ctx is done.

-----

semaphore:

```go
//  two concurrent jobs, regardless of individual cost
sema := make(chan struct{}, 2)

// this will block until an empty struct can be passed through the channel
sema <- struct{}{}

// reading makes room
defer func() { <-sema }()
```

-----

weightedcap:

```go
// use "New" to create a new capacity "manager"
cap := weightedcap.New(3)

// works with context.Context for timeouts and cancels.
ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*5)
defer cancel()

// Request N capacity with "Consume"
err := cap.Consume(ctx, 1)
if err != nil {
	t.Fatal(err)
}
// Free up N capacity with "Release"
defer cap.Release(1)
```
