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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MetadataInjectorSpec defines the desired state of MetadataInjector
type MetadataInjectorSpec struct {
	// Selectors defines the criteria for selecting resources
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Selectors []ResourceSelector `json:"selectors"`

	// Inject defines the metadata to inject into the selected resources
	// +kubebuilder:validation:Required
	Inject MetadataInjection `json:"inject"`
}

// ResourceSelector defines the resource selection criteria
type ResourceSelector struct {
	// Kind is the resource kind (e.g., Pod, Deployment)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Kind string `json:"kind"`

	// Group is the API group of the resource
	// +optional
	Group string `json:"group,omitempty"`

	// Version is the API version of the resource
	// +optional
	Version string `json:"version,omitempty"`

	// Namespaces is the list of namespaces to target
	// If empty, targets all namespaces
	// +optional
	Namespaces []string `json:"namespaces,omitempty"`

	// Names is the list of resource names to target
	// If empty, targets all resources of the specified kind
	// +optional
	Names []string `json:"names,omitempty"`
}

// MetadataInjection defines the metadata to inject
type MetadataInjection struct {
	// Annotations to inject into the resources
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Labels to inject into the resources
	// +optional
	Labels map[string]string `json:"labels,omitempty"`
}

// MetadataInjectorStatus defines the observed state of MetadataInjector
type MetadataInjectorStatus struct {
	// LastScheduledTime is the last time the reconciliation was scheduled
	// +optional
	LastScheduledTime *metav1.Time `json:"lastScheduledTime,omitempty"`

	// NextScheduledTime is the next time the reconciliation will run
	// +optional
	NextScheduledTime *metav1.Time `json:"nextScheduledTime,omitempty"`

	// LastSuccessfulTime is the last time the reconciliation completed successfully
	// +optional
	LastSuccessfulTime *metav1.Time `json:"lastSuccessfulTime,omitempty"`

	// Interval is the interval between reconciliations
	// +optional
	Interval string `json:"interval,omitempty"`

	// Conditions represent the latest available observations of an object's state
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="Interval",type="string",JSONPath=".status.interval"
//+kubebuilder:printcolumn:name="Last Success",type="string",JSONPath=".status.lastSuccessfulTime"
//+kubebuilder:printcolumn:name="Next Run",type="string",JSONPath=".status.nextScheduledTime"

// MetadataInjector is the Schema for the metadatainjectors API
type MetadataInjector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MetadataInjectorSpec   `json:"spec,omitempty"`
	Status MetadataInjectorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MetadataInjectorList contains a list of MetadataInjector
type MetadataInjectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MetadataInjector `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MetadataInjector{}, &MetadataInjectorList{})
}
