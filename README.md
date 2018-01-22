# Weightedcap

`$ go get -u github.com/romainmenke/weightedcap`

```go
// use new to make a new capacity "manager"
cap := weightedcap.New(3)

// works with context.Context for timeouts and cancels.
ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*5)
defer cancel()

// Request N capacity with "Push"
err := cap.Push(ctx, 1)
if err != nil {
	t.Fatal(err)
}
// Free up capacity with "Pop"
defer cap.Pop(3)
```

This pkg is useful if you like semaphores but you have heterogynous jobs.

A simple example is a process that uses a lot of memory but not the same amount for each job. If an estimate of memory usage is possible you can use this pkg to give a weight to each job. `Push` will block until enough capacity is available.
