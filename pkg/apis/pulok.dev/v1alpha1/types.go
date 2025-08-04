// +kubebuilder:object:generate=true
// +groupName=pulok.dev
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Kluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              KlusterSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen=true
type KlusterSpec struct {
	Replicas int32  `json:"replicas"`
	Image    string `json:"image"`
	Port     int32  `json:"port"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type KlusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kluster `json:"items"`
}
