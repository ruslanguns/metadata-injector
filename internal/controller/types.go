package controller

import (
	"sync"
	"time"

	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha1 "github.com/ruslanguns/metadata-injector-operator/api/v1alpha1"
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
