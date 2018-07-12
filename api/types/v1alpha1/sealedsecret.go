package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type SealedSecretSpec struct {
	Replicas int    `json:"replicas"`
	Owner    string `json:"owner"`
}

type SealedSecret struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Data              map[string][]byte `json:"data"`
	Spec              SealedSecretSpec  `json:"spec"`
}

type SealedSecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []SealedSecret `json:"items"`
}
