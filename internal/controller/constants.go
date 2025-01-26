package controller

import "time"

const (
	annotationDisableAutoReconcile = "metadata-injector.ruso.dev/disable-auto-reconcile"
	annotationReconcileInterval    = "metadata-injector.ruso.dev/reconcile-interval"
	defaultReconcileInterval       = 5 * time.Minute
	defaultBatchInterval           = 1 * time.Minute
	defaultWorkers                 = 5
)
