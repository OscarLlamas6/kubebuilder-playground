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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	catalogv1alpha1 "github.com/OscarLlamas6/kubebuilder-playground/api/v1alpha1"
)

// BookReconciler reconciles a Book object
type BookReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=catalog.bookstore.io,resources=books,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=catalog.bookstore.io,resources=books/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=catalog.bookstore.io,resources=books/finalizers,verbs=update
// +kubebuilder:rbac:groups=catalog.bookstore.io,resources=authors,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Book object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *BookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("Reconciling Book", "name", req.Name, "namespace", req.Namespace)

	// 1. Fetch the Book instance
	var book catalogv1alpha1.Book
	if err := r.Get(ctx, req.NamespacedName, &book); err != nil {
		if apierrors.IsNotFound(err) {
			// Book was deleted, nothing to do
			logger.Info("Book not found, probably deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request
		logger.Error(err, "Failed to get Book")
		return ctrl.Result{}, err
	}

	// 2. If Book has an AuthorRef, fetch Author info and populate status
	var authorFoundCondition metav1.Condition
	if book.Spec.AuthorRef != nil {
		var author catalogv1alpha1.Author
		authorKey := client.ObjectKey{
			Name:      book.Spec.AuthorRef.Name,
			Namespace: book.Namespace,
		}

		if err := r.Get(ctx, authorKey, &author); err != nil {
			if apierrors.IsNotFound(err) {
				logger.Info("Referenced Author not found", "author", book.Spec.AuthorRef.Name)
				authorFoundCondition = metav1.Condition{
					Type:               "AuthorFound",
					Status:             metav1.ConditionFalse,
					Reason:             "AuthorNotFound",
					Message:            "Referenced Author does not exist",
					ObservedGeneration: book.Generation,
				}
				book.Status.AuthorInfo = nil
			} else {
				logger.Error(err, "Failed to get Author")
				return ctrl.Result{}, err
			}
		} else {
			// Author found, populate AuthorInfo in Book status
			logger.Info("Author found, updating Book status", "author", author.Spec.Name)
			book.Status.AuthorInfo = &catalogv1alpha1.AuthorInfo{
				Name:        author.Spec.Name,
				Nationality: author.Spec.Nationality,
				BirthYear:   author.Spec.BirthYear,
			}
			authorFoundCondition = metav1.Condition{
				Type:               "AuthorFound",
				Status:             metav1.ConditionTrue,
				Reason:             "AuthorExists",
				Message:            "Referenced Author found and info cached",
				ObservedGeneration: book.Generation,
			}
		}
		apimeta.SetStatusCondition(&book.Status.Conditions, authorFoundCondition)
	}

	// 3. Set Ready condition
	readyCondition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "BookRegistered",
		Message:            "Book has been successfully registered in the catalog",
		ObservedGeneration: book.Generation,
	}
	apimeta.SetStatusCondition(&book.Status.Conditions, readyCondition)

	// 4. Update the status
	logger.Info("Updating Book status", "title", book.Spec.Title)
	if err := r.Status().Update(ctx, &book); err != nil {
		logger.Error(err, "Failed to update Book status")
		return ctrl.Result{}, err
	}
	logger.Info("Book status updated successfully")

	return ctrl.Result{}, nil
}

// findBooksForAuthor returns a list of reconcile requests for all Books that reference the given Author.
// This function is used as a mapper for the Author watch.
func (r *BookReconciler) findBooksForAuthor(ctx context.Context, author client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)

	// List all Books in the same namespace as the Author
	var bookList catalogv1alpha1.BookList
	if err := r.List(ctx, &bookList, client.InNamespace(author.GetNamespace())); err != nil {
		logger.Error(err, "Failed to list Books for Author", "author", author.GetName())
		return []reconcile.Request{}
	}

	// Filter Books that reference this Author
	var requests []reconcile.Request
	for _, book := range bookList.Items {
		if book.Spec.AuthorRef != nil && book.Spec.AuthorRef.Name == author.GetName() {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      book.Name,
					Namespace: book.Namespace,
				},
			})
			logger.Info("Queueing Book for reconciliation due to Author change",
				"book", book.Name, "author", author.GetName())
		}
	}

	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *BookReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.Book{}).
		// Watch Authors and reconcile Books that reference them
		Watches(
			&catalogv1alpha1.Author{},
			handler.EnqueueRequestsFromMapFunc(r.findBooksForAuthor),
		).
		Named("book").
		Complete(r)
}
