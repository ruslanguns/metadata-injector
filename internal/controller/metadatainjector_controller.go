/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/ruslanguns/metadata-injector-operator/api/v1alpha1"
)

const (
	annotationDisableAutoReconcile = "metadata-injector.ruso.dev/disable-auto-reconcile"
	annotationReconcileInterval    = "metadata-injector.ruso.dev/reconcile-interval"
	defaultReconcileInterval       = 5 * time.Minute
	defaultBatchInterval           = 1 * time.Minute
	defaultWorkers                 = 5
)

// ReconcileJob represents a scheduled reconciliation job
type ReconcileJob struct {
	Injector *corev1alpha1.MetadataInjector
	NextRun  time.Time
}

// BatchScheduler handles batch processing of MetadataInjectors
type BatchScheduler struct {
	client        client.Client
	dynamicClient dynamic.Interface
	batchInterval time.Duration
	jobsChan      chan ReconcileJob
	workers       int
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// MetadataInjectorReconciler reconciles a MetadataInjector object
// +kubebuilder:rbac:groups=core.k8s.ruso.dev,resources=metadatainjectors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.k8s.ruso.dev,resources=metadatainjectors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.k8s.ruso.dev,resources=metadatainjectors/finalizers,verbs=update
// +kubebuilder:rbac:groups="*",resources="*",verbs=get;list;watch;update;patch
type MetadataInjectorReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	DynamicClient dynamic.Interface
	scheduler     *BatchScheduler
}

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
	// Start workers
	for i := 0; i < bs.workers; i++ {
		bs.wg.Add(1)
		go bs.worker()
	}

	// Start batch processor
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

func (bs *BatchScheduler) processJob(ctx context.Context, job ReconcileJob) error {
	log := log.FromContext(ctx)

	// Get current interval status
	interval := defaultReconcileInterval
	if customInterval, ok := job.Injector.Annotations[annotationReconcileInterval]; ok {
		if parsed, err := time.ParseDuration(customInterval); err == nil {
			interval = parsed
		}
	}

	// Set status based on auto-reconcile setting
	intervalStatus := interval.String()
	if disabled, _ := strconv.ParseBool(job.Injector.Annotations[annotationDisableAutoReconcile]); disabled {
		intervalStatus = "False"
	}

	for _, selector := range job.Injector.Spec.Selectors {
		log.Info("Processing selector", "selector", selector)

		resource := strings.ToLower(fmt.Sprintf("%ss", selector.Kind))

		gvr := schema.GroupVersionResource{
			Group:    selector.Group,
			Version:  selector.Version,
			Resource: resource,
		}

		if selector.Group == "" && selector.Version == "v1" {
			gvr = schema.GroupVersionResource{
				Group:    "",
				Version:  "v1",
				Resource: resource,
			}
		}

		namespaces := selector.Namespaces
		if len(namespaces) == 0 {
			namespaces = []string{""}
		}

		for _, ns := range namespaces {
			list, err := bs.dynamicClient.Resource(gvr).Namespace(ns).List(ctx, metav1.ListOptions{})
			if err != nil {
				log.Error(err, "unable to list resources", "gvr", gvr, "namespace", ns)
				continue
			}

			for _, item := range list.Items {
				if len(selector.Names) > 0 {
					found := false
					for _, name := range selector.Names {
						if item.GetName() == name {
							found = true
							break
						}
					}
					if !found {
						continue
					}
				}

				if len(job.Injector.Spec.Inject.Labels) > 0 {
					labels := item.GetLabels()
					if labels == nil {
						labels = make(map[string]string)
					}
					for k, v := range job.Injector.Spec.Inject.Labels {
						labels[k] = v
					}
					item.SetLabels(labels)
				}

				if len(job.Injector.Spec.Inject.Annotations) > 0 {
					annotations := item.GetAnnotations()
					if annotations == nil {
						annotations = make(map[string]string)
					}
					for k, v := range job.Injector.Spec.Inject.Annotations {
						annotations[k] = v
					}
					item.SetAnnotations(annotations)
				}

				_, err := bs.dynamicClient.Resource(gvr).Namespace(item.GetNamespace()).Update(
					ctx,
					&item,
					metav1.UpdateOptions{},
				)
				if err != nil {
					log.Error(err, "failed to update resource",
						"name", item.GetName(),
						"namespace", item.GetNamespace(),
					)
					continue
				}
			}
		}
	}

	return bs.updateStatus(ctx, job.Injector, intervalStatus)
}

func (bs *BatchScheduler) updateStatus(ctx context.Context, injector *corev1alpha1.MetadataInjector, intervalStatus string) error {
	now := metav1.Now()
	nextRun := calculateNextRun(injector)

	patch := client.MergeFrom(injector.DeepCopy())
	injector.Status.LastSuccessfulTime = &now
	injector.Status.NextScheduledTime = &metav1.Time{Time: nextRun}
	injector.Status.Interval = intervalStatus

	return bs.client.Status().Patch(ctx, injector, patch)
}

func shouldProcess(injector *corev1alpha1.MetadataInjector) bool {
	if disabled, _ := strconv.ParseBool(injector.Annotations[annotationDisableAutoReconcile]); disabled {
		return false
	}
	return true
}

func calculateNextRun(injector *corev1alpha1.MetadataInjector) time.Time {
	interval := defaultReconcileInterval
	if customInterval, ok := injector.Annotations[annotationReconcileInterval]; ok {
		if parsed, err := time.ParseDuration(customInterval); err == nil {
			interval = parsed
		}
	}
	return time.Now().Add(interval)
}

func (r *MetadataInjectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var injector corev1alpha1.MetadataInjector
	if err := r.Get(ctx, req.NamespacedName, &injector); err != nil {
		if errors.IsNotFound(err) {
			log.Info("MetadataInjector resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get MetadataInjector")
		return ctrl.Result{}, err
	}

	// Process immediately
	job := ReconcileJob{
		Injector: injector.DeepCopy(),
		NextRun:  calculateNextRun(&injector),
	}
	if err := r.scheduler.processJob(ctx, job); err != nil {
		log.Error(err, "Failed to process job")
		return ctrl.Result{}, err
	}

	// Requeue based on the next scheduled run
	return ctrl.Result{RequeueAfter: time.Until(job.NextRun)}, nil
}

func (r *MetadataInjectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	dynamicClient, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return err
	}

	r.DynamicClient = dynamicClient
	r.scheduler = NewBatchScheduler(r.Client, dynamicClient, defaultBatchInterval, defaultWorkers)
	r.scheduler.Start()

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.MetadataInjector{}).
		Complete(r)
}
