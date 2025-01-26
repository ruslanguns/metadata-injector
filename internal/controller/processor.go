package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/ruslanguns/metadata-injector-operator/api/v1alpha1"
)

func (bs *BatchScheduler) processJob(ctx context.Context, job ReconcileJob) error {
	log := log.FromContext(ctx)

	interval := defaultReconcileInterval
	if customInterval, ok := job.Injector.Annotations[annotationReconcileInterval]; ok {
		if parsed, err := time.ParseDuration(customInterval); err == nil {
			interval = parsed
		}
	}

	intervalStatus := interval.String()
	if disabled, _ := strconv.ParseBool(job.Injector.Annotations[annotationDisableAutoReconcile]); disabled {
		intervalStatus = "False"
	}

	for _, selector := range job.Injector.Spec.Selectors {
		log.Info("Processing selector", "selector", selector)

		resource := strings.ToLower(fmt.Sprintf("%ss", selector.Kind))
		gvr := getGroupVersionResource(selector.Group, selector.Version, resource)
		namespaces := getNamespaces(selector.Namespaces)

		for _, ns := range namespaces {
			if err := bs.processNamespace(ctx, job, selector, gvr, ns); err != nil {
				log.Error(err, "failed to process namespace", "namespace", ns)
				continue
			}
		}
	}

	return bs.updateStatus(ctx, job.Injector, intervalStatus)
}

func (bs *BatchScheduler) processNamespace(ctx context.Context, job ReconcileJob, selector corev1alpha1.ResourceSelector, gvr schema.GroupVersionResource, namespace string) error {
	list, err := bs.dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable to list resources: %w", err)
	}

	for _, item := range list.Items {
		if !shouldProcessResource(item.GetName(), selector.Names) {
			continue
		}

		updateMetadata(&item, job.Injector.Spec.Inject.Labels, job.Injector.Spec.Inject.Annotations)

		_, err := bs.dynamicClient.Resource(gvr).Namespace(item.GetNamespace()).Update(
			ctx,
			&item,
			metav1.UpdateOptions{},
		)
		if err != nil {
			log.FromContext(ctx).Error(err, "failed to update resource",
				"name", item.GetName(),
				"namespace", item.GetNamespace(),
			)
			continue
		}
	}

	return nil
}
