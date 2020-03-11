// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// AzureSQLVNetRuleSpec defines the desired state of AzureSQLVNetRule
type AzureSQLVNetRuleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ResourceGroup                string `json:"resourceGroup"`
	Server                       string `json:"server"`
	VirtualNetworkName           string `json:"virtualNetworkName"`
	SubnetName                   string `json:"subnetName"`
	AddressPrefix                string `json:"addressPrefix"`
	IgnoreMissingServiceEndpoint bool   `json:"ignoreMissingServiceEndpoint,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// AzureSQLVNetRule is the Schema for the azuresqlvnetrules API
type AzureSQLVNetRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureSQLVNetRuleSpec `json:"spec,omitempty"`
	Status ASOStatus            `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AzureSQLVNetRuleList contains a list of AzureSQLVNetRule
type AzureSQLVNetRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureSQLVNetRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AzureSQLVNetRule{}, &AzureSQLVNetRuleList{})
}
