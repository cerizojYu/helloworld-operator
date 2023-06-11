/*
Copyright 2023.

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
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// type CRStatus string

const (
	ActiveStatus string = "active"
	FailedStatus string = "failed"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HelloworldSpec defines the desired state of Helloworld
type HelloworldSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Minimum:=1
	Replicas int32 `json:"replicas"`
	// +kubebuilder:validation:Optional
	Ingress networkingv1.IngressSpec `json:"ingress"`
}

// HelloworldStatus defines the observed state of Helloworld
type HelloworldStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	DeployStatus  string `json:"deploy_status"`
	SerivceStatus string `json:"service_status"`
	IngressStatus string `json:"ingress_status"`
}

// +kubebuilder:printcolumn:name="DeployStatus",type=string,JSONPath=`.status.deploy_status`
// +kubebuilder:printcolumn:name="SerivceStatus",type=string,JSONPath=`.status.service_status`
// +kubebuilder:printcolumn:name="IngressStatus",type=string,JSONPath=`.status.ingress_status`

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Helloworld is the Schema for the helloworlds API
type Helloworld struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HelloworldSpec   `json:"spec,omitempty"`
	Status HelloworldStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HelloworldList contains a list of Helloworld
type HelloworldList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Helloworld `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Helloworld{}, &HelloworldList{})
}
