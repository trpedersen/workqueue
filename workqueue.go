package workqueue

type Job interface {
	Execute() error   // execute the job
	Report(err error) // report the error status of Execute()
}

type worker struct {
	workerPool chan chan Job
	jobQueue   chan Job
	quit       chan struct{}
}

func newWorker(quit chan struct{}, workerPool chan chan Job) *worker {
	worker := &worker{
		workerPool: workerPool,
		jobQueue:   make(chan Job),
	}
	go worker.run(quit)
	return worker
}

func (worker *worker) run(quit chan struct{}) {
	for {
		worker.workerPool <- worker.jobQueue
		select {
		case job := <-worker.jobQueue:
			job.Report(job.Execute())
		case <-quit:
			return
		}
	}
}

type WorkQueue interface {
	Enqueue(job Job) error
	Halt() error
}

type workQueue struct {
	jobQueue   chan Job
	workerPool chan chan Job
	maxWorkers int
	quit       chan struct{}
}

func NewWorkQueue(maxWorkers int) (WorkQueue, error) {
	workQueue := &workQueue{
		jobQueue:   make(chan Job),
		workerPool: make(chan chan Job, maxWorkers),
		maxWorkers: maxWorkers,
		quit:       make(chan struct{}),
	}
	go workQueue.run()
	return workQueue, nil
}

func (workQueue *workQueue) run() {
	for i := 0; i < workQueue.maxWorkers; i++ {
		newWorker(workQueue.quit, workQueue.workerPool)
	}
	go func() {
		for {
			select {
			case job := <-workQueue.jobQueue:
				go func(job Job) {
					jobChannel := <-workQueue.workerPool
					jobChannel <- job
				}(job)
			case <-workQueue.quit:
				return
			}
		}
	}()
}

func (workQueue *workQueue) Enqueue(job Job) error {
	go func(job Job) {
		workQueue.jobQueue <- job
	}(job)

	return nil
}

func (workQueue *workQueue) Halt() error {
	// Halt the dispatcher. Idempotent.
	select {
	case <-workQueue.quit: // already closed
		return nil
	default:
		workQueue.quit <- struct{}{}
	}
	<-workQueue.quit
	return nil
}
