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

// BookSpec defines the desired state of Book.
type BookSpec struct {
	// Title is the title of the book.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Title string `json:"title"`

	// ISBN is the International Standard Book Number.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[0-9]{3}-[0-9]{1,5}-[0-9]{1,7}-[0-9]{1,7}-[0-9X]$|^[0-9]{9}[0-9X]$`
	ISBN string `json:"isbn"`

	// Author is the name of the book's author.
	// +kubebuilder:validation:Required
	Author string `json:"author"`

	// Publisher is the name of the publishing company.
	// +kubebuilder:validation:Optional
	Publisher string `json:"publisher,omitempty"`

	// PublishedYear is the year the book was published.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1000
	// +kubebuilder:validation:Maximum=2100
	PublishedYear int `json:"publishedYear,omitempty"`

	// Genre is the literary genre of the book.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=Fiction;NonFiction;Mystery;SciFi;Fantasy;Romance;Thriller;Biography;History;SelfHelp;Other
	Genre string `json:"genre,omitempty"`

	// Pages is the number of pages in the book.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	Pages int `json:"pages,omitempty"`
}

// BookStatus defines the observed state of Book.
type BookStatus struct {
	// Conditions represent the current status of the book.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// AvailableCopies represents the number of available copies across all bookstores.
	// This will be calculated by the controller based on Inventory resources.
	// +kubebuilder:validation:Optional
	AvailableCopies int `json:"availableCopies,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Title",type=string,JSONPath=`.spec.title`
// +kubebuilder:printcolumn:name="Author",type=string,JSONPath=`.spec.author`
// +kubebuilder:printcolumn:name="ISBN",type=string,JSONPath=`.spec.isbn`
// +kubebuilder:printcolumn:name="Genre",type=string,JSONPath=`.spec.genre`
// +kubebuilder:printcolumn:name="Available",type=integer,JSONPath=`.status.availableCopies`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Book is the Schema for the books API.
type Book struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BookSpec   `json:"spec,omitempty"`
	Status BookStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BookList contains a list of Book.
type BookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Book `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Book{}, &BookList{})
}
