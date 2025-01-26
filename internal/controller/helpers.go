package controller

import (
	"context"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha1 "github.com/ruslanguns/metadata-injector-operator/api/v1alpha1"
)

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

func getGroupVersionResource(group, version, resource string) schema.GroupVersionResource {
	if group == "" && version == "v1" {
		return schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: resource,
		}
	}
	return schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
}

func getNamespaces(namespaces []string) []string {
	if len(namespaces) == 0 {
		return []string{""}
	}
	return namespaces
}

func shouldProcessResource(name string, targetNames []string) bool {
	if len(targetNames) == 0 {
		return true
	}
	for _, targetName := range targetNames {
		if name == targetName {
			return true
		}
	}
	return false
}

func updateMetadata(item *unstructured.Unstructured, labels, annotations map[string]string) {
	if len(labels) > 0 {
		currentLabels := item.GetLabels()
		if currentLabels == nil {
			currentLabels = make(map[string]string)
		}
		for k, v := range labels {
			currentLabels[k] = v
		}
		item.SetLabels(currentLabels)
	}

	if len(annotations) > 0 {
		currentAnnotations := item.GetAnnotations()
		if currentAnnotations == nil {
			currentAnnotations = make(map[string]string)
		}
		for k, v := range annotations {
			currentAnnotations[k] = v
		}
		item.SetAnnotations(currentAnnotations)
	}
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
