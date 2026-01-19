# 📚 BookStore Platform - Kubebuilder Learning Guide

## 📑 Table of Contents

- [Introduction](#introduction)
- [Prerequisites](#prerequisites)
- [Kind Cluster Configuration](#kind-cluster-configuration)
  - [Configuration Options](#configuration-options)
  - [Create Cluster with Configuration](#create-cluster-with-configuration)
- [Project Architecture](#project-architecture)
- [Phase 1: Book CRD and Basic Controller](#phase-1-book-crd-and-basic-controller)
  - [Project Structure](#project-structure)
  - [CRD Anatomy](#crd-anatomy)
  - [Controller Anatomy](#controller-anatomy)
  - [Kubebuilder Annotations](#kubebuilder-annotations)
- [Local Development vs Production](#local-development-vs-production)
  - [Local Development Mode](#local-development-mode)
  - [Production Cluster Mode](#production-cluster-mode)
- [Quick Commands with Taskfile](#quick-commands-with-taskfile)
- [Key Concepts](#key-concepts)
- [Next Steps](#next-steps)

---

## Introduction

This project is a bookstore management platform built with **Kubebuilder** to learn how to create Custom Resource Definitions (CRDs) and Kubernetes controllers from scratch.

**Goal:** Learn Kubebuilder by building a medium-complexity project that demonstrates real-world patterns and best practices for Kubernetes operators.

**Resources we'll build:**
- ✅ **Book** - Basic book resource (Phase 1 - Completed)
- ✅ **Author** - Authors with relationships to Books (Phase 2 - Completed)
- 🔜 **BookStore** - Physical bookstores (Phase 3)
- 🔜 **Inventory** - Book inventory per bookstore (Phase 3)
- 🔜 **BookReservation** - Reservations with webhooks (Phase 4)
- 🔜 **Review** - Book reviews (Phase 5)

---

## Prerequisites

- **Go** 1.23+
- **Docker** 17.03+
- **kubectl** v1.11.3+
- **kind** (Kubernetes in Docker)
- **kubebuilder** v4.5.1+
- **task** (Task runner)

### Verify installation:
```bash
go version
docker --version
kubectl version --client
kind version
kubebuilder version
task --version
```

---

## Kind Cluster Configuration

Kind (Kubernetes in Docker) allows you to create local Kubernetes clusters using Docker containers as nodes. It's ideal for development and testing.

### Configuration Options

The `kind/cluster-config.yaml` file defines the cluster topology and features:

#### **1. Nodes**

You can configure multiple nodes with different roles:

```yaml
nodes:
  - role: control-plane  # Master node
  - role: worker         # Worker node 1
  - role: worker         # Worker node 2
```

**Use cases:**
- **1 control-plane**: Simple development
- **3 control-plane + 2 workers**: Simulate HA (High Availability)
- **1 control-plane + N workers**: Scheduling testing

#### **2. Port Mappings**

Expose cluster ports to your local machine:

```yaml
extraPortMappings:
  - containerPort: 30080  # Port on the node
    hostPort: 8080        # Port on localhost
    protocol: TCP
```

**Example:** Access a NodePort service at `localhost:8080`

#### **3. Extra Mounts**

Mount local directories into nodes:

```yaml
extraMounts:
  - hostPath: /tmp/kind-data    # Path on your machine
    containerPath: /data         # Path in the node
    readOnly: false
```

**Use cases:**
- Share configuration files
- Use local persistent storage
- Debugging with shared logs

#### **4. Networking**

Configure cluster subnets:

```yaml
networking:
  podSubnet: "10.244.0.0/16"      # Network for Pods
  serviceSubnet: "10.96.0.0/12"   # Network for Services
  apiServerPort: 6443              # API Server port
```

**Important:** Useful if you have conflicts with other local networks.

#### **5. Kubeadm Config Patches**

Customize Kubernetes configuration:

```yaml
kubeadmConfigPatches:
  - |
    kind: ClusterConfiguration
    apiServer:
      extraArgs:
        audit-log-path: /var/log/kubernetes/audit.log
        enable-admission-plugins: NodeRestriction,PodSecurity
```

**Use cases:**
- Enable audit logging
- Configure admission controllers
- Adjust feature gates

#### **6. Node Labels**

Label nodes for specific scheduling:

```yaml
kubeadmConfigPatches:
  - |
    kind: JoinConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "workload=compute,zone=us-east-1a"
```

**Example:** Deploy specific workloads on labeled nodes.

#### **7. Kubernetes Version**

Specify K8s version:

```yaml
nodes:
  - role: control-plane
    image: kindest/node:v1.33.1  # Specific version
```

**Useful for:** Testing compatibility with different versions.

### Create Cluster with Configuration

```bash
# Using the configuration file
kind create cluster --config kind/cluster-config.yaml

# Or using Taskfile
task create-cluster
```

### Verify the Cluster

```bash
# View nodes
kubectl get nodes

# View cluster info
kubectl cluster-info

# View current context
kubectl config current-context
```

### Delete the Cluster

```bash
kind delete cluster --name bookstore-dev

# Or using Taskfile
task delete-cluster
```

---

## Project Architecture

```
kubebuilder-playground/
├── api/v1alpha1/              # Type definitions (CRDs)
│   ├── book_types.go          # Book Spec and Status
│   └── groupversion_info.go   # API group metadata
├── internal/controller/       # Reconciliation logic
│   └── book_controller.go     # Book controller
├── config/                    # Kubernetes manifests
│   ├── crd/                   # Generated CRDs
│   ├── rbac/                  # RBAC permissions
│   ├── manager/               # Controller deployment
│   └── samples/               # Resource examples
├── cmd/main.go               # Entry point
├── Taskfile.yaml             # Simplified commands
└── docs/                     # Documentation
```

---

## Phase 1: Book CRD and Basic Controller

### Project Structure

#### 1. **api/v1alpha1/book_types.go**

Defines the Book resource structure:

**BookSpec** (Desired state - what the user defines):
```go
type BookSpec struct {
    Title         string `json:"title"`           // Book title
    ISBN          string `json:"isbn"`            // Unique ISBN
    Author        string `json:"author"`          // Author name
    Publisher     string `json:"publisher"`       // Publisher
    PublishedYear int    `json:"publishedYear"`   // Publication year
    Genre         string `json:"genre"`           // Literary genre
    Pages         int    `json:"pages"`           // Number of pages
}
```

**BookStatus** (Observed state - what the controller reports):
```go
type BookStatus struct {
    Conditions      []metav1.Condition `json:"conditions"`      // Book status
    AvailableCopies int                `json:"availableCopies"` // Available copies
}
```

#### 2. **internal/controller/book_controller.go**

Contains the reconciliation logic:

```go
func (r *BookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. Get the Book from the cluster
    var book catalogv1alpha1.Book
    if err := r.Get(ctx, req.NamespacedName, &book); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // 2. Create Ready condition
    readyCondition := metav1.Condition{
        Type:   "Ready",
        Status: metav1.ConditionTrue,
        Reason: "BookRegistered",
        Message: "Book has been successfully registered in the catalog",
    }

    // 3. Update status if changed
    if apimeta.SetStatusCondition(&book.Status.Conditions, readyCondition) {
        if err := r.Status().Update(ctx, &book); err != nil {
            return ctrl.Result{}, err
        }
    }

    return ctrl.Result{}, nil
}
```

### CRD Anatomy

A Custom Resource Definition has three main parts:

#### **1. Spec (Specification)**
- Defines the **desired state** of the resource
- What the **user** configures
- **Immutable** once created (unless updated)

#### **2. Status**
- Defines the **observed state** of the resource
- What the **controller** reports
- Only the controller can update it (thanks to `+kubebuilder:subresource:status`)

#### **3. Metadata**
- Standard Kubernetes information: name, namespace, labels, annotations
- Managed by Kubernetes

### Controller Anatomy

#### **Main Components:**

1. **BookReconciler**: Structure containing the Kubernetes client
2. **Reconcile()**: Function executed when changes occur
3. **SetupWithManager()**: Registers the controller and configures watches

#### **Reconciliation Loop:**

```
User creates Book
    ↓
API Server saves to etcd
    ↓
Watch detects change
    ↓
Reconcile() executes
    ↓
Reads Book, processes logic
    ↓
Updates Status
    ↓
API Server saves status
```

### Kubebuilder Annotations

`+kubebuilder:` annotations automatically generate code and manifests:

#### **Validations:**
```go
// +kubebuilder:validation:Required
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:Pattern=`^[0-9-]+$`
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=2100
// +kubebuilder:validation:Enum=Fiction;NonFiction;Mystery
```

#### **Print Columns** (for `kubectl get books`):
```go
// +kubebuilder:printcolumn:name="Title",type=string,JSONPath=`.spec.title`
// +kubebuilder:printcolumn:name="Author",type=string,JSONPath=`.spec.author`
// +kubebuilder:printcolumn:name="Genre",type=string,JSONPath=`.spec.genre`
```

#### **RBAC** (controller permissions):
```go
// +kubebuilder:rbac:groups=catalog.bookstore.io,resources=books,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=catalog.bookstore.io,resources=books/status,verbs=get;update;patch
```

---

## Local Development vs Production

### Local Development Mode

**Command:** `task run-local` or `make run`

```
Your Laptop                        Kind Cluster
┌──────────────┐                  ┌─────────┐
│ Controller   │ ←── kubectl ───→ │  CRDs   │
│ (Go process) │      watch        │  Books  │
└──────────────┘                  └─────────┘
```

**Features:**
- ✅ Controller runs on your machine as a Go process
- ✅ Connects to cluster using `~/.kube/config`
- ✅ Direct logs in your terminal
- ✅ Fast iteration and debugging
- ❌ Only works while your laptop is on

**Steps:**
```bash
# 1. Install CRDs in the cluster
task install-crds

# 2. Run controller locally
task run-local
```

### Production Cluster Mode

**Command:** `task deploy-all`

```
Kind Cluster
┌────────────────────────────────────┐
│  Namespace: system                 │
│  ┌──────────────────────────────┐  │
│  │ Pod: controller-manager      │  │
│  │   Container: manager         │  │
│  │     - Binary: /manager       │  │
│  │     - Image: controller:v1   │  │
│  └──────────────────────────────┘  │
│           ↓ watch                  │
│  ┌──────────────────────────────┐  │
│  │ Namespace: default           │  │
│  │   Books: book-1984, hobbit   │  │
│  └──────────────────────────────┘  │
└────────────────────────────────────┘
```

**Features:**
- ✅ Controller runs as a Pod inside the cluster
- ✅ High availability (with multiple replicas)
- ✅ Runs 24/7
- ✅ Exactly like real production

**Steps:**
```bash
# 1. Build Docker image
task build

# 2. Load image into Kind
task load-image

# 3. Deploy to cluster
task deploy

# Or all in one command:
task deploy-all
```

**Resources created in cluster:**
- Namespace: `kubebuilder-playground-system`
- ServiceAccount: `controller-manager`
- ClusterRoles and RoleBindings (RBAC)
- Deployment: `controller-manager`
- Service: For Prometheus metrics

---

## Quick Commands with Taskfile

### Code Generation
```bash
task generate              # Generate CRDs and RBAC
```

### Local Development
```bash
task install-crds          # Install CRDs in cluster
task run-local             # Run controller locally
task dev                   # install-crds + run-local (quick start)
```

### Cluster Deployment
```bash
task build                 # Build Docker image
task load-image            # Load image into Kind
task deploy                # Deploy to cluster
task deploy-all            # build + load-image + deploy
task prod                  # generate + deploy-all (quick deploy)
```

### Testing
```bash
task create-sample         # Create sample Book
task list-books            # List all Books
task get-book BOOK=book-1984  # View Book details
```

### Debugging
```bash
task logs                  # View controller logs
task pods                  # List controller pods
```

### Cluster Management
```bash
task create-cluster        # Create Kind cluster
task delete-cluster        # Delete Kind cluster
```

### Cleanup
```bash
task undeploy              # Remove controller from cluster
task uninstall-crds        # Uninstall CRDs
task clean-all             # undeploy + uninstall-crds
```

---

## Key Concepts

### 1. **Custom Resource Definition (CRD)**
Extends the Kubernetes API with new resource types. In our case, we added the `Book` type.

### 2. **Controller**
Process that watches resources and executes logic to maintain desired state. It's a continuous reconciliation loop.

### 3. **Reconcile Loop**
Fundamental Kubernetes pattern:
- Observe current state
- Compare with desired state
- Take actions to converge to desired state

### 4. **Spec vs Status**
- **Spec**: What the user wants (desired state)
- **Status**: What's actually happening (observed state)

### 5. **Conditions**
Standard way to report status in Kubernetes:
- `Type`: Condition name (e.g., "Ready")
- `Status`: True, False, Unknown
- `Reason`: Programmatic code
- `Message`: Human-readable message

### 6. **RBAC (Role-Based Access Control)**
Kubernetes permission system. The controller needs permissions to read/write resources.

### 7. **Watch**
Kubernetes mechanism to observe resource changes in real-time.

### 8. **Leader Election**
When there are multiple controller replicas, only one actively reconciles (the "leader").

---

## Next Steps

### Phase 2: Resource Relationships
- Create **Author** CRD
- Relate Books with Authors through references
- Implement cross-resource watches
- Automatically update status when Author changes

### Phase 3: Advanced Resources
- Create **BookStore** and **Inventory**
- Implement many-to-many relationships
- Use finalizers for cleanup
- Owner references

### Phase 4: Webhooks
- Validation webhook for Books
- Mutation webhook (automatic defaults)
- Complex validations

### Phase 5: Advanced Features
- **BookReservation** with state machine
- Granular RBAC per user
- Subresources (scale)
- Custom metrics

---

## Additional Resources

- [Kubebuilder Book](https://book.kubebuilder.io/)
- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- [Controller Runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [Kind Documentation](https://kind.sigs.k8s.io/)

---

**Version:** Phase 2 Completed  
**Last updated:** January 2026  
**Author:** Oscar Llamas

---

## Phase 2: Author CRD and Resource Relationships

### Introduction to Phase 2

In Phase 1, we created a simple CRD (Book) with a basic controller. In Phase 2, we'll learn more advanced concepts:

- **Resource relationships**: How one resource can reference another
- **Cross-resource watches**: How a controller can observe multiple resource types
- **Status propagation**: How to cache information from related resources
- **Bidirectional updates**: How to keep related resources synchronized

### Step 1: Create the Author CRD

```bash
# Create API and controller for Author
kubebuilder create api --group catalog --version v1alpha1 --kind Author --resource --controller
```

**What does this command generate?**
- `api/v1alpha1/author_types.go` - Author definition
- `internal/controller/author_controller.go` - Author controller
- `config/crd/bases/catalog.bookstore.io_authors.yaml` - CRD manifest
- Updates `cmd/main.go` to register the new controller

### Step 2: Define the AuthorSpec

Edit `api/v1alpha1/author_types.go`:

```go
type AuthorSpec struct {
    // Name is the author's full name
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:MinLength=1
    // +kubebuilder:validation:MaxLength=100
    Name string `json:"name"`

    // Biography is a brief description of the author
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:MaxLength=1000
    Biography string `json:"biography,omitempty"`

    // BirthYear is the year of birth
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=1000
    // +kubebuilder:validation:Maximum=2100
    BirthYear int `json:"birthYear,omitempty"`

    // DeathYear is the year of death (if applicable)
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
```

**Explanation of new validations:**
- `MaxLength`: Maximum string length
- `Pattern=^https?://.*`: URL must start with http:// or https://

### Step 3: Define the AuthorStatus

```go
type AuthorStatus struct {
    // Conditions represents the author's status
    // +kubebuilder:validation:Optional
    Conditions []metav1.Condition `json:"conditions,omitempty"`

    // BookCount is the number of books by this author in the catalog
    // Automatically calculated by the controller
    // +kubebuilder:validation:Optional
    BookCount int `json:"bookCount,omitempty"`
}
```

**Why BookCount in status?**
- Status is calculated by the controller, not by the user
- BookCount updates automatically when Books are created/deleted
- It's derived information, not part of the desired state

### Step 4: Add Print Columns to Author

```go
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.spec.name`
// +kubebuilder:printcolumn:name="Nationality",type=string,JSONPath=`.spec.nationality`
// +kubebuilder:printcolumn:name="Birth Year",type=integer,JSONPath=`.spec.birthYear`
// +kubebuilder:printcolumn:name="Books",type=integer,JSONPath=`.status.bookCount`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`
type Author struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec   AuthorSpec   `json:"spec,omitempty"`
    Status AuthorStatus `json:"status,omitempty"`
}
```

### Step 5: Update Book to Use References

Edit `api/v1alpha1/book_types.go`:

**Add new types:**

```go
// AuthorReference contains a reference to an Author resource
type AuthorReference struct {
    // Name is the name of the Author resource
    // +kubebuilder:validation:Required
    Name string `json:"name"`
}

// AuthorInfo contains cached author information
type AuthorInfo struct {
    // Name is the author's full name
    // +kubebuilder:validation:Optional
    Name string `json:"name,omitempty"`

    // Nationality is the author's nationality
    // +kubebuilder:validation:Optional
    Nationality string `json:"nationality,omitempty"`

    // BirthYear is the year of birth
    // +kubebuilder:validation:Optional
    BirthYear int `json:"birthYear,omitempty"`
}
```

**Why create these types?**
- `AuthorReference`: Defines how a Book references an Author
- `AuthorInfo`: Caches Author information in the Book for quick queries

**Update BookSpec:**

```go
type BookSpec struct {
    Title string `json:"title"`
    ISBN  string `json:"isbn"`
    
    // Author (deprecated) - Keep for backward compatibility
    // +kubebuilder:validation:Optional
    Author string `json:"author,omitempty"`

    // AuthorRef is the reference to the Author resource
    // +kubebuilder:validation:Optional
    AuthorRef *AuthorReference `json:"authorRef,omitempty"`
    
    // ... rest of fields
}
```

**Why keep the `author` field?**
- Backward compatibility with existing Books
- Allows gradual migration
- Controller supports both formats

**Update BookStatus:**

```go
type BookStatus struct {
    Conditions      []metav1.Condition `json:"conditions,omitempty"`
    AvailableCopies int                `json:"availableCopies,omitempty"`
    
    // AuthorInfo contains cached information from the referenced Author
    // +kubebuilder:validation:Optional
    AuthorInfo *AuthorInfo `json:"authorInfo,omitempty"`
}
```

### Step 6: Implement the Author Controller

Edit `internal/controller/author_controller.go`:

```go
func (r *AuthorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    // 1. Get the Author
    var author catalogv1alpha1.Author
    if err := r.Get(ctx, req.NamespacedName, &author); err != nil {
        if apierrors.IsNotFound(err) {
            return ctrl.Result{}, nil
        }
        return ctrl.Result{}, err
    }

    logger.Info("Reconciling Author", "name", author.Name, "authorName", author.Spec.Name)

    // 2. List all Books
    var bookList catalogv1alpha1.BookList
    if err := r.List(ctx, &bookList); err != nil {
        logger.Error(err, "Failed to list Books")
        return ctrl.Result{}, err
    }

    // 3. Count books that reference this author
    bookCount := 0
    for _, book := range bookList.Items {
        // Check author field (string) - old format
        if book.Spec.Author == author.Spec.Name {
            bookCount++
        }
        // Check authorRef - new format
        if book.Spec.AuthorRef != nil && book.Spec.AuthorRef.Name == author.Name {
            bookCount++
        }
    }

    // 4. Update status
    author.Status.BookCount = bookCount

    // 5. Set Ready condition
    readyCondition := metav1.Condition{
        Type:               "Ready",
        Status:             metav1.ConditionTrue,
        Reason:             "AuthorRegistered",
        Message:            "Author has been successfully registered in the catalog",
        ObservedGeneration: author.Generation,
    }

    // 6. Update status if changed
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
```

**Code explanation:**

**Lines 1-10**: Get the Author from the cluster (same as Book)

**Lines 12-16**: List all Books in the cluster
- `r.List(ctx, &bookList)`: Lists all Books
- We don't filter by namespace because we want to count all

**Lines 18-27**: Count books by this author
- Iterate over all Books
- Check both formats: `author` (string) and `authorRef`
- Increment counter if it matches

**Lines 29-30**: Update the `bookCount` field in status

**Lines 32-38**: Create Ready condition (standard pattern)

**Lines 40-48**: Update status only if there were changes
- `apimeta.SetStatusCondition` returns `true` if something changed
- Only do `Update` if necessary (optimization)

### Step 7: Update the Book Controller

Edit `internal/controller/book_controller.go`:

```go
func (r *BookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    // 1. Get the Book
    var book catalogv1alpha1.Book
    if err := r.Get(ctx, req.NamespacedName, &book); err != nil {
        if apierrors.IsNotFound(err) {
            return ctrl.Result{}, nil
        }
        return ctrl.Result{}, err
    }

    // 2. If Book has authorRef, fetch Author and cache its info
    var authorFoundCondition metav1.Condition
    if book.Spec.AuthorRef != nil {
        var author catalogv1alpha1.Author
        authorKey := client.ObjectKey{
            Name:      book.Spec.AuthorRef.Name,
            Namespace: book.Namespace,
        }

        if err := r.Get(ctx, authorKey, &author); err != nil {
            if apierrors.IsNotFound(err) {
                // Author doesn't exist
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
                // Error fetching Author
                logger.Error(err, "Failed to get Author")
                return ctrl.Result{}, err
            }
        } else {
            // Author found, cache information
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

    // 4. Update status
    logger.Info("Updating Book status", "title", book.Spec.Title)
    if err := r.Status().Update(ctx, &book); err != nil {
        logger.Error(err, "Failed to update Book status")
        return ctrl.Result{}, err
    }
    logger.Info("Book status updated successfully")

    return ctrl.Result{}, nil
}
```

**Detailed explanation:**

**Lines 14-15**: Check if the Book has an Author reference
- `book.Spec.AuthorRef != nil`: Only process if there's a reference

**Lines 16-20**: Prepare Author lookup
- `client.ObjectKey`: Structure with Name and Namespace
- Use the same namespace as the Book

**Lines 22-34**: Handle case where Author doesn't exist
- Create `AuthorFound` condition with status `False`
- Clear `AuthorInfo` from status (no longer valid)
- Don't return error, just report the state

**Lines 39-52**: Handle case where Author exists
- Cache Author information in `book.Status.AuthorInfo`
- Only save relevant fields (not all information)
- Create `AuthorFound` condition with status `True`

**Why cache information?**
- Faster queries: Don't need to GET the Author every time
- Less load on API Server
- Information available in `kubectl get book -o yaml`

### Step 8: Implement Cross-Resource Watches

**What are Cross-Resource Watches?**

They allow a controller to observe changes in different resource types. For example:
- Book controller watches Authors
- When an Author changes, it reconciles all Books that reference it

**Update Book Controller - SetupWithManager:**

```go
// findBooksForAuthor returns reconcile requests for all Books that reference the given Author
func (r *BookReconciler) findBooksForAuthor(ctx context.Context, author client.Object) []reconcile.Request {
    logger := log.FromContext(ctx)

    // List all Books in the same namespace
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
```

**Code explanation:**

**`findBooksForAuthor` - Mapper Function:**
- Receives an Author that changed
- Returns list of Books that should be reconciled
- Called automatically by the framework when an Author changes

**Lines 6-10**: List Books in the same namespace
- `client.InNamespace(author.GetNamespace())`: Filter by namespace
- Optimization: Don't search in all namespaces

**Lines 13-26**: Create reconciliation requests
- Iterate over Books
- If `authorRef.Name` matches, add to list
- Each request will cause `Reconcile()` to execute for that Book

**`SetupWithManager`:**
- `For(&catalogv1alpha1.Book{})`: Primary resource it manages
- `Watches(&catalogv1alpha1.Author{}, ...)`: Also watches Authors
- `handler.EnqueueRequestsFromMapFunc`: Uses our mapper function

**Update Author Controller - SetupWithManager:**

```go
// findAuthorsForBook returns reconcile request for the Author referenced by the Book
func (r *AuthorReconciler) findAuthorsForBook(ctx context.Context, book client.Object) []reconcile.Request {
    logger := log.FromContext(ctx)

    // Type assert to Book
    bookObj, ok := book.(*catalogv1alpha1.Book)
    if !ok {
        logger.Error(nil, "Failed to cast object to Book")
        return []reconcile.Request{}
    }

    // If Book has authorRef, queue the Author
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
```

**Why do we need bidirectional watches?**

**Book → Author:**
- When an Author changes (e.g., updates biography)
- We need to update `authorInfo` in all their Books

**Author → Book:**
- When a Book is created/deleted
- We need to update `bookCount` in the Author

### Step 9: Generate and Deploy

```bash
# Generate manifests
make manifests generate

# Deploy everything
task deploy-all
```

**What gets generated?**
- Author CRD in `config/crd/bases/`
- RBAC roles for Author in `config/rbac/`
- Updated Book CRD with new fields
- Updated RBAC permissions for cross-resource access

### Step 10: Create Authors and Books

**Create Authors:**

```bash
kubectl apply -f config/samples/authors-collection.yaml
```

```yaml
apiVersion: catalog.bookstore.io/v1alpha1
kind: Author
metadata:
  name: george-orwell
spec:
  name: "George Orwell"
  biography: "English novelist, essayist, journalist, and critic."
  birthYear: 1903
  deathYear: 1950
  nationality: "British"
```

**Create Book with authorRef:**

```bash
kubectl apply -f config/samples/book-with-authorref.yaml
```

```yaml
apiVersion: catalog.bookstore.io/v1alpha1
kind: Book
metadata:
  name: book-1984
spec:
  title: "1984"
  isbn: "978-0-452-28423-4"
  authorRef:
    name: george-orwell  # Reference to Author
  publisher: "Penguin Books"
  publishedYear: 1949
  genre: "Fiction"
  pages: 328
```

### Step 11: Verify Relationships

```bash
# View Authors with bookCount
kubectl get authors

# Output:
# NAME           NAME            NATIONALITY   BIRTH YEAR   BOOKS   AGE
# george-orwell  George Orwell   British       1903         1       5m

# View Book with authorInfo
kubectl get book book-1984 -o yaml
```

**Expected output:**

```yaml
status:
  authorInfo:
    name: George Orwell
    nationality: British
    birthYear: 1903
  conditions:
  - type: Ready
    status: "True"
    reason: BookRegistered
    message: "Book has been successfully registered in the catalog"
  - type: AuthorFound
    status: "True"
    reason: AuthorExists
    message: "Referenced Author found and info cached"
```

### Step 12: Test Cross-Resource Watches

**Update the Author:**

```bash
kubectl patch author george-orwell --type='json' -p='[{"op": "replace", "path": "/spec/biography", "value": "Updated biography"}]'
```

**What happens?**
1. Author controller reconciles the Author
2. Book controller detects the change (thanks to watch)
3. Book controller reconciles all Books by that Author
4. `authorInfo` updates automatically in the Books

**Verify:**

```bash
# View controller logs
kubectl logs -n kubebuilder-playground-system deployment/kubebuilder-playground-controller-manager --tail=50

# You should see:
# "Queueing Book for reconciliation due to Author change" book="book-1984" author="george-orwell"
# "Author found, updating Book status" author="George Orwell"
```

### Key Concepts of Phase 2

#### 1. **Resource References**

**What are they?**
- Way to relate resources in Kubernetes
- Similar to "foreign keys" in databases

**Types of references:**
- **Name-based**: Only store the name (what we use)
- **UID-based**: Store the unique UID
- **Owner References**: Parent-child relationship with garbage collection

**When to use each?**
- Name-based: Simple relationships, easy to read
- UID-based: When you need to guarantee unique identity
- Owner References: When child should be deleted with parent

#### 2. **Status Propagation**

**What is it?**
- Copying information from one resource to another
- Caching data for quick queries

**Advantages:**
- Fewer calls to API Server
- Information available in a single GET
- Better performance

**Disadvantages:**
- Data can become stale
- Need watches to keep synchronized
- More complexity in controller

#### 3. **Cross-Resource Watches**

**What are they?**
- Observing changes in different resource types
- Automatically reconcile when they change

**Mapper Function Pattern:**
```go
func (r *Reconciler) findRelatedResources(ctx context.Context, obj client.Object) []reconcile.Request {
    // 1. Identify which resources are related
    // 2. Return list of requests to reconcile them
}
```

**When to use:**
- Relationships between resources
- Derived information that must update
- Maintaining consistency between resources

#### 4. **Bidirectional Updates**

**What are they?**
- Two controllers that watch each other
- Changes in A affect B, and vice versa

**Beware of infinite loops:**
- Use `ObservedGeneration` to detect real changes
- Only update status if something actually changed
- `apimeta.SetStatusCondition` returns `false` if no changes

### Complete Flow Diagram

```
User creates Book with authorRef
         ↓
Book Controller reconciles
         ↓
Searches for referenced Author
         ↓
    Exists?
    ↙     ↘
  Yes      No
   ↓       ↓
Cache    AuthorFound=False
info      ↓
   ↓    Update
AuthorFound=True  Book Status
   ↓       ↓
Update ←┘
Book Status
   ↓
Author Controller detects change (watch)
   ↓
Reconciles Author
   ↓
Counts Books that reference it
   ↓
Updates bookCount
   ↓
Book Controller detects Author change (watch)
   ↓
Reconciles Author's Books
   ↓
Updates authorInfo
```

### Useful Debugging Commands

```bash
# View cluster events
kubectl get events --sort-by='.lastTimestamp'

# View logs with filter
kubectl logs -n kubebuilder-playground-system deployment/kubebuilder-playground-controller-manager | grep -i author

# Describe a Book
kubectl describe book book-1984

# View controller RBAC
kubectl get clusterrole kubebuilder-playground-manager-role -o yaml

# Force reconciliation
kubectl annotate book book-1984 reconcile=now --overwrite
```

