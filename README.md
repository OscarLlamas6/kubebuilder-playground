# 📚 BookStore Platform - Kubebuilder Learning Project

A hands-on Kubebuilder project for learning how to build Kubernetes Custom Resource Definitions (CRDs) and controllers from scratch. This project implements a bookstore management platform to understand the architecture and patterns used in production-grade Kubernetes operators.

## 🎯 Overview

This project demonstrates:
- Building CRDs and controllers with Kubebuilder
- Kubernetes reconciliation loops and status conditions
- Local development with Kind clusters
- Production deployment patterns
- Comparison with real-world projects like [Milo](https://github.com/datum-cloud/milo)

**Current Status:** Phase 1 Complete ✅
- Book CRD with validations
- Basic controller with reconciliation logic
- Kind cluster configuration
- Sample resources and documentation

## 📖 Documentation

Comprehensive guides available in multiple languages:

- **[📘 English Documentation](docs/README-en.md)** - Complete guide with architecture, deployment, and examples
- **[📗 Documentación en Español](docs/README-es.md)** - Guía completa con arquitectura, deployment y ejemplos

Both include:
- Kind cluster configuration
- Project architecture
- CRD and controller anatomy
- Local vs production deployment
- Taskfile command reference
- Key Kubernetes concepts

## 🚀 Quick Start

```bash
# Create Kind cluster
task create-cluster

# Deploy controller to cluster
task prod

# Create sample books
kubectl apply -f config/samples/books-collection.yaml

# View books
task list-books
```

**See full documentation for detailed setup and explanations.**

## 🛠️ Common Commands

```bash
task --list              # View all available commands
task dev                 # Quick local development setup
task logs                # View controller logs
task get-book BOOK=name  # Get book details
task clean-all           # Remove everything
```

## 📦 What's Included

- **Book CRD** - Custom resource with validations
- **Controller** - Reconciliation logic with status updates
- **Kind Config** - 2-node cluster setup
- **Taskfile** - Simplified command interface
- **Samples** - 6 example books across different genres
- **Docs** - Complete guides in English and Spanish

## 🔗 Resources

- [Kubebuilder Documentation](https://book.kubebuilder.io/)
- [Milo Project](https://github.com/datum-cloud/milo) - Production reference
- [Kind Documentation](https://kind.sigs.k8s.io/)

## License

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

