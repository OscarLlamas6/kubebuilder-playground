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

// AuthorReconciler reconciles a Author object
type AuthorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=catalog.bookstore.io,resources=authors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=catalog.bookstore.io,resources=authors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=catalog.bookstore.io,resources=authors/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Author object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *AuthorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the Author instance
	var author catalogv1alpha1.Author
	if err := r.Get(ctx, req.NamespacedName, &author); err != nil {
		if apierrors.IsNotFound(err) {
			// Author was deleted, nothing to do
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Author")
		return ctrl.Result{}, err
	}

	logger.Info("Reconciling Author", "name", author.Name, "authorName", author.Spec.Name)

	// Count books by this author
	var bookList catalogv1alpha1.BookList
	if err := r.List(ctx, &bookList); err != nil {
		logger.Error(err, "Failed to list Books")
		return ctrl.Result{}, err
	}

	// Count books that reference this author
	// We check both the old 'author' field (string) and the new 'authorRef' field
	bookCount := 0
	for _, book := range bookList.Items {
		// Check old-style author field (deprecated)
		if book.Spec.Author == author.Spec.Name {
			bookCount++
		}
		// Check new-style authorRef field
		if book.Spec.AuthorRef != nil && book.Spec.AuthorRef.Name == author.Name {
			bookCount++
		}
	}

	// Update status
	author.Status.BookCount = bookCount

	// Set Ready condition
	readyCondition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "AuthorRegistered",
		Message:            "Author has been successfully registered in the catalog",
		ObservedGeneration: author.Generation,
	}

	if apimeta.SetStatusCondition(&author.Status.Conditions, readyCondition) {
		logger.Info("Updating Author status", "name", author.Name, "bookCount", bookCount)
		if err := r.Status().Update(ctx, &author); err != nil {
			logger.Error(err, "Failed to update Author status")
			return ctrl.Result{}, err
		}
		logger.Info("Author status updated successfully")
	}

	return ctrl.Result{}, nil
}

// findAuthorsForBook returns a list of reconcile requests for the Author referenced by the given Book.
// This function is used as a mapper for the Book watch.
func (r *AuthorReconciler) findAuthorsForBook(ctx context.Context, book client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)

	// Type assert to Book
	bookObj, ok := book.(*catalogv1alpha1.Book)
	if !ok {
		logger.Error(nil, "Failed to cast object to Book")
		return []reconcile.Request{}
	}

	// If Book has an AuthorRef, queue the Author for reconciliation
	if bookObj.Spec.AuthorRef != nil {
		logger.Info("Queueing Author for reconciliation due to Book change",
			"author", bookObj.Spec.AuthorRef.Name, "book", bookObj.Name)
		return []reconcile.Request{
			{
				NamespacedName: types.NamespacedName{
					Name:      bookObj.Spec.AuthorRef.Name,
					Namespace: bookObj.Namespace,
				},
			},
		}
	}

	return []reconcile.Request{}
}

// SetupWithManager sets up the controller with the Manager.
func (r *AuthorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&catalogv1alpha1.Author{}).
		// Watch Books and reconcile Authors when Books change
		Watches(
			&catalogv1alpha1.Book{},
			handler.EnqueueRequestsFromMapFunc(r.findAuthorsForBook),
		).
		Named("author").
		Complete(r)
}
