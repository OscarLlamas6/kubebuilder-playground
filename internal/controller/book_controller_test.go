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

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	catalogv1alpha1 "github.com/OscarLlamas6/kubebuilder-playground/api/v1alpha1"
)

var _ = Describe("Book Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		book := &catalogv1alpha1.Book{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind Book")
			err := k8sClient.Get(ctx, typeNamespacedName, book)
			if err != nil && errors.IsNotFound(err) {
				resource := &catalogv1alpha1.Book{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: catalogv1alpha1.BookSpec{
						Title:         "Test Book",
						ISBN:          "978-0-123-45678-9",
						Author:        "Test Author",
						Publisher:     "Test Publisher",
						PublishedYear: 2024,
						Genre:         "Fiction",
						Pages:         200,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &catalogv1alpha1.Book{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Book")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &BookReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking if the Book status was updated")
			err = k8sClient.Get(ctx, typeNamespacedName, book)
			Expect(err).NotTo(HaveOccurred())
			Expect(book.Status.Conditions).NotTo(BeEmpty())
			Expect(book.Status.Conditions[0].Type).To(Equal("Ready"))
			Expect(book.Status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))
			Expect(book.Status.Conditions[0].Reason).To(Equal("BookRegistered"))
		})
	})
})
