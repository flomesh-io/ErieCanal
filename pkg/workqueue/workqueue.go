/*
 * Copyright 2022 The flomesh.io Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package workqueue

import (
	"fmt"
	"k8s.io/klog/v2"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type ProcessFunc func(key, name, namespace string) (bool, error)

type Interface interface {
	Enqueue(obj interface{})
	NumRequeues(key string) int
	Run(stopCh <-chan struct{}, process ProcessFunc)
	ShutDown()
}

type queueType struct {
	workqueue.RateLimitingInterface

	name string
}

func New(name string) Interface {
	return &queueType{
		RateLimitingInterface: workqueue.NewNamedRateLimitingQueue(workqueue.NewMaxOfRateLimiter(
			// exponential per-item rate limiter
			workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 30*time.Second),
			// overall rate limiter (not per item)
			&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
		), name),
		name: name,
	}
}

func (q *queueType) Enqueue(obj interface{}) {
	var key string
	var err error
	if key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}

	klog.V(5).Infof("%s: enqueueing key %q for %T object", q.name, key, obj)
	q.AddRateLimited(key)
}

func (q *queueType) Run(stopCh <-chan struct{}, process ProcessFunc) {
	go wait.Until(func() {
		for q.processNextWorkItem(process) {
		}
	}, time.Second, stopCh)
}

func (q *queueType) processNextWorkItem(process ProcessFunc) bool {
	obj, shutdown := q.Get()
	if shutdown {
		return false
	}

	key, ok := obj.(string)
	if !ok {
		panic(fmt.Sprintf("Work queue %q received type %T instead of string", q.name, obj))
	}

	defer q.Done(key)

	requeue, err := func() (bool, error) {
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			panic(err)
		}

		return process(key, name, ns)
	}()
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("%s: Failed to process object with key %q: %w", q.name, key, err))
	}

	if requeue {
		q.AddRateLimited(key)
		klog.V(5).Infof("%s: enqueued %q for retry - # of times re-queued: %d", q.name, key, q.NumRequeues(key))
	} else {
		q.Forget(key)
	}

	return true
}

func (q *queueType) NumRequeues(key string) int {
	return q.RateLimitingInterface.NumRequeues(key)
}

const (
	// Size of the job queue per worker
	maxJobPerWorker = 4096
)

// worker context for a worker routine
type worker struct {
	id            int
	jobs          chan Job        // Job queue
	stop          chan struct{}   // Stop channel
	wg            *sync.WaitGroup // Pointer to WorkerPool wg
	jobsProcessed uint64          // Jobs processed by this worker
}

// WorkerPool object representation
type WorkerPool struct {
	wg            sync.WaitGroup // Sync group, to stop workers if needed
	workerContext []*worker      // Worker contexts
	nWorkers      uint64         // Number of workers. Uint64 for easier mod hash later
	rRobinCounter uint64         // Used only by the round robin api. Modified atomically on API.
}

// Job is a runnable interface to queue jobs on a WorkerPool
type Job interface {
	// JobName returns the name of the job.
	JobName() string

	// Hash returns a uint64 hash for a job.
	Hash() uint64

	// Run executes the job.
	Run()

	// GetDoneCh returns the channel, which when closed, indicates that the job was finished.
	GetDoneCh() <-chan struct{}
}

// NewWorkerPool creates a new work group.
// If nWorkers is 0, will poll goMaxProcs to get the number of routines to spawn.
// Reminder: routines are never pinned to system threads, it's up to the go scheduler to decide
// when and where these will be scheduled.
func NewWorkerPool(nWorkers int) *WorkerPool {
	if nWorkers == 0 {
		// read GOMAXPROCS, -1 to avoid changing it
		nWorkers = runtime.GOMAXPROCS(-1)
	}

	klog.Infof("New worker pool setting up %d workers", nWorkers)

	var workPool WorkerPool
	for i := 0; i < nWorkers; i++ {
		workPool.workerContext = append(workPool.workerContext,
			&worker{
				id:            i,
				jobs:          make(chan Job, maxJobPerWorker),
				stop:          make(chan struct{}, 1),
				wg:            &workPool.wg,
				jobsProcessed: 0,
			},
		)
		workPool.wg.Add(1)
		workPool.nWorkers++

		go (workPool.workerContext[i]).work()
	}

	return &workPool
}

// AddJob posts the job on a worker queue
// Uses Hash underneath to choose worker to post the job to
func (wp *WorkerPool) AddJob(job Job) <-chan struct{} {
	wp.workerContext[job.Hash()%wp.nWorkers].jobs <- job
	return job.GetDoneCh()
}

// AddJobRoundRobin adds a job in round robin to the queues
// Concurrent calls to AddJobRoundRobin are thread safe and fair
// between each other
func (wp *WorkerPool) AddJobRoundRobin(jobs Job) {
	added := atomic.AddUint64(&wp.rRobinCounter, 1)
	wp.workerContext[added%wp.nWorkers].jobs <- jobs
}

// GetWorkerNumber get number of queues/workers
func (wp *WorkerPool) GetWorkerNumber() int {
	return int(wp.nWorkers)
}

// Stop stops the workerpool
func (wp *WorkerPool) Stop() {
	for _, worker := range wp.workerContext {
		worker.stop <- struct{}{}
	}
	wp.wg.Wait()
}

func (workContext *worker) work() {
	defer workContext.wg.Done()

	klog.Infof("Worker %d running", workContext.id)
	for {
		select {
		case j := <-workContext.jobs:
			t := time.Now()
			klog.Infof("work[%d]: Starting %v", workContext.id, j.JobName())

			// Run current job
			j.Run()

			klog.Infof("work[%d][%s] : took %v", workContext.id, j.JobName(), time.Since(t))
			workContext.jobsProcessed++

		case <-workContext.stop:
			klog.Infof("work[%d]: Stopped", workContext.id)
			return
		}
	}
}
