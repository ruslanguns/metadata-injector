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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/ruslanguns/metadata-injector-operator/api/v1alpha1"
)

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
