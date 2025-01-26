package controller

import (
	"context"
	"time"

	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/ruslanguns/metadata-injector-operator/api/v1alpha1"
)

func NewBatchScheduler(c client.Client, dc dynamic.Interface, batchInterval time.Duration, workers int) *BatchScheduler {
	return &BatchScheduler{
		client:        c,
		dynamicClient: dc,
		batchInterval: batchInterval,
		jobsChan:      make(chan ReconcileJob, 100),
		workers:       workers,
		stopChan:      make(chan struct{}),
	}
}

func (bs *BatchScheduler) Start() {
	for i := 0; i < bs.workers; i++ {
		bs.wg.Add(1)
		go bs.worker()
	}

	bs.wg.Add(1)
	go bs.processBatches()
}

func (bs *BatchScheduler) Stop() {
	close(bs.stopChan)
	bs.wg.Wait()
}

func (bs *BatchScheduler) processBatches() {
	defer bs.wg.Done()
	ticker := time.NewTicker(bs.batchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-bs.stopChan:
			return
		case <-ticker.C:
			ctx := context.Background()
			var injectors corev1alpha1.MetadataInjectorList
			if err := bs.client.List(ctx, &injectors); err != nil {
				continue
			}

			for i := range injectors.Items {
				injector := &injectors.Items[i]
				if shouldProcess(injector) {
					bs.jobsChan <- ReconcileJob{
						Injector: injector.DeepCopy(),
						NextRun:  calculateNextRun(injector),
					}
				}
			}
		}
	}
}

func (bs *BatchScheduler) worker() {
	defer bs.wg.Done()
	for {
		select {
		case <-bs.stopChan:
			return
		case job := <-bs.jobsChan:
			ctx := context.Background()
			if err := bs.processJob(ctx, job); err != nil {
				log.FromContext(ctx).Error(err, "failed to process job",
					"name", job.Injector.Name,
					"namespace", job.Injector.Namespace)
			}
		}
	}
}
