# workqueue
Work queue + worker pool

Handy library for pooling worker go-routines. A good learning example of using pooled resources, channels, go routines, etc.

usage:

  myQ = workqueue.NewWorkQueue
  myQ.Enqueue(myJob)
  
(myJob implements Job interface.)

