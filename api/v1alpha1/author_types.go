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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AuthorSpec defines the desired state of Author.
type AuthorSpec struct {
	// Name is the full name of the author
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=100
	Name string `json:"name"`

	// Biography is a brief description of the author
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MaxLength=1000
	Biography string `json:"biography,omitempty"`

	// BirthYear is the year the author was born
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1000
	// +kubebuilder:validation:Maximum=2100
	BirthYear int `json:"birthYear,omitempty"`

	// DeathYear is the year the author died (if applicable)
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1000
	// +kubebuilder:validation:Maximum=2100
	DeathYear int `json:"deathYear,omitempty"`

	// Nationality is the author's nationality
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:MaxLength=50
	Nationality string `json:"nationality,omitempty"`

	// Website is the author's official website
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern=`^https?://.*`
	Website string `json:"website,omitempty"`
}

// AuthorStatus defines the observed state of Author.
type AuthorStatus struct {
	// Conditions represent the latest available observations of the Author's state
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// BookCount is the number of books by this author in the catalog
	// +kubebuilder:validation:Optional
	BookCount int `json:"bookCount,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Nationality",type=string,JSONPath=`.spec.nationality`
// +kubebuilder:printcolumn:name="Birth Year",type=integer,JSONPath=`.spec.birthYear`
// +kubebuilder:printcolumn:name="Books",type=integer,JSONPath=`.status.bookCount`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`

// Author is the Schema for the authors API.
type Author struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AuthorSpec   `json:"spec,omitempty"`
	Status AuthorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AuthorList contains a list of Author.
type AuthorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Author `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Author{}, &AuthorList{})
}
