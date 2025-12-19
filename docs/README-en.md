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

**Goal:** Understand how the Milo project is built by creating a similar medium-complexity project.

**Resources we'll build:**
- ✅ **Book** - Basic book resource (Phase 1 - Completed)
- 🔜 **Author** - Authors with relationships to Books (Phase 2)
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
- [Milo Project](https://github.com/datum-cloud/milo) - Architecture reference

---

**Version:** Phase 1 Completed  
**Last updated:** December 2025
