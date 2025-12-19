# 📚 BookStore Platform - Guía de Aprendizaje Kubebuilder

## 📑 Índice

- [Introducción](#introducción)
- [Requisitos Previos](#requisitos-previos)
- [Configuración del Cluster Kind](#configuración-del-cluster-kind)
  - [Opciones de Configuración](#opciones-de-configuración)
  - [Crear Cluster con Configuración](#crear-cluster-con-configuración)
- [Arquitectura del Proyecto](#arquitectura-del-proyecto)
- [Fase 1: CRD Book y Controlador Básico](#fase-1-crd-book-y-controlador-básico)
  - [Estructura del Proyecto](#estructura-del-proyecto)
  - [Anatomía de un CRD](#anatomía-de-un-crd)
  - [Anatomía del Controlador](#anatomía-del-controlador)
  - [Anotaciones Kubebuilder](#anotaciones-kubebuilder)
- [Desarrollo Local vs Producción](#desarrollo-local-vs-producción)
  - [Modo Desarrollo Local](#modo-desarrollo-local)
  - [Modo Producción en Cluster](#modo-producción-en-cluster)
- [Comandos Rápidos con Taskfile](#comandos-rápidos-con-taskfile)
- [Conceptos Clave](#conceptos-clave)
- [Próximos Pasos](#próximos-pasos)

---

## Introducción

Este proyecto es una plataforma de gestión de librerías construida con **Kubebuilder** para aprender cómo crear Custom Resource Definitions (CRDs) y controladores de Kubernetes desde cero.

**Objetivo:** Entender cómo está construido el proyecto Milo mediante la creación de un proyecto similar de mediana complejidad.

**Recursos que construiremos:**
- ✅ **Book** - Recurso básico de libros (Fase 1 - Completada)
- 🔜 **Author** - Autores con relaciones a Books (Fase 2)
- 🔜 **BookStore** - Librerías físicas (Fase 3)
- 🔜 **Inventory** - Inventario de libros por librería (Fase 3)
- 🔜 **BookReservation** - Reservas con webhooks (Fase 4)
- 🔜 **Review** - Reseñas de libros (Fase 5)

---

## Requisitos Previos

- **Go** 1.23+
- **Docker** 17.03+
- **kubectl** v1.11.3+
- **kind** (Kubernetes in Docker)
- **kubebuilder** v4.5.1+
- **task** (Task runner)

### Verificar instalación:
```bash
go version
docker --version
kubectl version --client
kind version
kubebuilder version
task --version
```

---

## Configuración del Cluster Kind

Kind (Kubernetes in Docker) permite crear clusters de Kubernetes locales usando contenedores Docker como nodos. Es ideal para desarrollo y testing.

### Opciones de Configuración

El archivo `kind/cluster-config.yaml` define la topología y características del cluster:

#### **1. Nodos (Nodes)**

Puedes configurar múltiples nodos con diferentes roles:

```yaml
nodes:
  - role: control-plane  # Nodo maestro
  - role: worker         # Nodo trabajador 1
  - role: worker         # Nodo trabajador 2
```

**Casos de uso:**
- **1 control-plane**: Desarrollo simple
- **3 control-plane + 2 workers**: Simular HA (Alta Disponibilidad)
- **1 control-plane + N workers**: Testing de scheduling

#### **2. Port Mappings**

Expone puertos del cluster a tu máquina local:

```yaml
extraPortMappings:
  - containerPort: 30080  # Puerto en el nodo
    hostPort: 8080        # Puerto en localhost
    protocol: TCP
```

**Ejemplo:** Acceder a un servicio NodePort en `localhost:8080`

#### **3. Extra Mounts**

Monta directorios locales en los nodos:

```yaml
extraMounts:
  - hostPath: /tmp/kind-data    # Ruta en tu máquina
    containerPath: /data         # Ruta en el nodo
    readOnly: false
```

**Casos de uso:**
- Compartir archivos de configuración
- Usar almacenamiento local persistente
- Debugging con logs compartidos

#### **4. Networking**

Configura las subredes del cluster:

```yaml
networking:
  podSubnet: "10.244.0.0/16"      # Red para Pods
  serviceSubnet: "10.96.0.0/12"   # Red para Services
  apiServerPort: 6443              # Puerto del API Server
```

**Importante:** Útil si tienes conflictos con otras redes locales.

#### **5. Kubeadm Config Patches**

Personaliza la configuración de Kubernetes:

```yaml
kubeadmConfigPatches:
  - |
    kind: ClusterConfiguration
    apiServer:
      extraArgs:
        audit-log-path: /var/log/kubernetes/audit.log
        enable-admission-plugins: NodeRestriction,PodSecurity
```

**Casos de uso:**
- Habilitar audit logging
- Configurar admission controllers
- Ajustar feature gates

#### **6. Node Labels**

Etiqueta nodos para scheduling específico:

```yaml
kubeadmConfigPatches:
  - |
    kind: JoinConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "workload=compute,zone=us-east-1a"
```

**Ejemplo:** Deployar workloads específicos en nodos etiquetados.

#### **7. Versión de Kubernetes**

Especifica la versión de K8s:

```yaml
nodes:
  - role: control-plane
    image: kindest/node:v1.33.1  # Versión específica
```

**Útil para:** Testing de compatibilidad con diferentes versiones.

### Crear Cluster con Configuración

```bash
# Usando el archivo de configuración
kind create cluster --config kind/cluster-config.yaml

# O usando el Taskfile
task create-cluster
```

### Verificar el Cluster

```bash
# Ver nodos
kubectl get nodes

# Ver información del cluster
kubectl cluster-info

# Ver contexto actual
kubectl config current-context
```

### Eliminar el Cluster

```bash
kind delete cluster --name bookstore-dev

# O usando Taskfile
task delete-cluster
```

---

## Arquitectura del Proyecto

```
kubebuilder-playground/
├── api/v1alpha1/              # Definiciones de tipos (CRDs)
│   ├── book_types.go          # Spec y Status del Book
│   └── groupversion_info.go   # Metadata del grupo API
├── internal/controller/       # Lógica de reconciliación
│   └── book_controller.go     # Controlador del Book
├── config/                    # Manifiestos de Kubernetes
│   ├── crd/                   # CRDs generados
│   ├── rbac/                  # Permisos RBAC
│   ├── manager/               # Deployment del controlador
│   └── samples/               # Ejemplos de recursos
├── cmd/main.go               # Punto de entrada
├── Taskfile.yaml             # Comandos simplificados
└── docs/                     # Documentación
```

---

## Fase 1: CRD Book y Controlador Básico

### Estructura del Proyecto

#### 1. **api/v1alpha1/book_types.go**

Define la estructura del recurso Book:

**BookSpec** (Estado deseado - lo que el usuario define):
```go
type BookSpec struct {
    Title         string `json:"title"`           // Título del libro
    ISBN          string `json:"isbn"`            // ISBN único
    Author        string `json:"author"`          // Nombre del autor
    Publisher     string `json:"publisher"`       // Editorial
    PublishedYear int    `json:"publishedYear"`   // Año de publicación
    Genre         string `json:"genre"`           // Género literario
    Pages         int    `json:"pages"`           // Número de páginas
}
```

**BookStatus** (Estado observado - lo que el controlador reporta):
```go
type BookStatus struct {
    Conditions      []metav1.Condition `json:"conditions"`      // Estado del libro
    AvailableCopies int                `json:"availableCopies"` // Copias disponibles
}
```

#### 2. **internal/controller/book_controller.go**

Contiene la lógica de reconciliación:

```go
func (r *BookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // 1. Obtener el Book del cluster
    var book catalogv1alpha1.Book
    if err := r.Get(ctx, req.NamespacedName, &book); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }

    // 2. Crear condición Ready
    readyCondition := metav1.Condition{
        Type:   "Ready",
        Status: metav1.ConditionTrue,
        Reason: "BookRegistered",
        Message: "Book has been successfully registered in the catalog",
    }

    // 3. Actualizar status si cambió
    if apimeta.SetStatusCondition(&book.Status.Conditions, readyCondition) {
        if err := r.Status().Update(ctx, &book); err != nil {
            return ctrl.Result{}, err
        }
    }

    return ctrl.Result{}, nil
}
```

### Anatomía de un CRD

Un Custom Resource Definition tiene tres partes principales:

#### **1. Spec (Especificación)**
- Define el **estado deseado** del recurso
- Lo que el **usuario** configura
- Es **inmutable** una vez creado (a menos que se actualice)

#### **2. Status**
- Define el **estado observado** del recurso
- Lo que el **controlador** reporta
- Solo el controlador puede actualizarlo (gracias a `+kubebuilder:subresource:status`)

#### **3. Metadata**
- Información estándar de Kubernetes: name, namespace, labels, annotations
- Gestionado por Kubernetes

### Anatomía del Controlador

#### **Componentes Principales:**

1. **BookReconciler**: Estructura que contiene el cliente de Kubernetes
2. **Reconcile()**: Función que se ejecuta cuando hay cambios
3. **SetupWithManager()**: Registra el controlador y configura watches

#### **Ciclo de Reconciliación:**

```
Usuario crea Book
    ↓
API Server guarda en etcd
    ↓
Watch detecta cambio
    ↓
Reconcile() se ejecuta
    ↓
Lee Book, procesa lógica
    ↓
Actualiza Status
    ↓
API Server guarda status
```

### Anotaciones Kubebuilder

Las anotaciones `+kubebuilder:` generan código y manifiestos automáticamente:

#### **Validaciones:**
```go
// +kubebuilder:validation:Required
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:Pattern=`^[0-9-]+$`
// +kubebuilder:validation:Minimum=1
// +kubebuilder:validation:Maximum=2100
// +kubebuilder:validation:Enum=Fiction;NonFiction;Mystery
```

#### **Print Columns** (para `kubectl get books`):
```go
// +kubebuilder:printcolumn:name="Title",type=string,JSONPath=`.spec.title`
// +kubebuilder:printcolumn:name="Author",type=string,JSONPath=`.spec.author`
// +kubebuilder:printcolumn:name="Genre",type=string,JSONPath=`.spec.genre`
```

#### **RBAC** (permisos del controlador):
```go
// +kubebuilder:rbac:groups=catalog.bookstore.io,resources=books,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=catalog.bookstore.io,resources=books/status,verbs=get;update;patch
```

---

## Desarrollo Local vs Producción

### Modo Desarrollo Local

**Comando:** `task run-local` o `make run`

```
Tu Laptop                          Cluster Kind
┌──────────────┐                  ┌─────────┐
│ Controlador  │ ←── kubectl ───→ │  CRDs   │
│ (proceso Go) │      watch        │  Books  │
└──────────────┘                  └─────────┘
```

**Características:**
- ✅ Controlador corre en tu máquina como proceso Go
- ✅ Se conecta al cluster usando `~/.kube/config`
- ✅ Logs directos en tu terminal
- ✅ Rápido para iterar y debuggear
- ❌ Solo funciona mientras tu laptop esté encendida

**Pasos:**
```bash
# 1. Instalar CRDs en el cluster
task install-crds

# 2. Ejecutar controlador localmente
task run-local
```

### Modo Producción en Cluster

**Comando:** `task deploy-all`

```
Cluster Kind
┌────────────────────────────────────┐
│  Namespace: system                 │
│  ┌──────────────────────────────┐  │
│  │ Pod: controller-manager      │  │
│  │   Container: manager         │  │
│  │     - Binario: /manager      │  │
│  │     - Imagen: controller:v1  │  │
│  └──────────────────────────────┘  │
│           ↓ watch                  │
│  ┌──────────────────────────────┐  │
│  │ Namespace: default           │  │
│  │   Books: book-1984, hobbit   │  │
│  └──────────────────────────────┘  │
└────────────────────────────────────┘
```

**Características:**
- ✅ Controlador corre como Pod dentro del cluster
- ✅ Alta disponibilidad (con múltiples replicas)
- ✅ Corre 24/7
- ✅ Exactamente como producción real

**Pasos:**
```bash
# 1. Construir imagen Docker
task build

# 2. Cargar imagen en Kind
task load-image

# 3. Deployar al cluster
task deploy

# O todo en un comando:
task deploy-all
```

**Recursos creados en el cluster:**
- Namespace: `kubebuilder-playground-system`
- ServiceAccount: `controller-manager`
- ClusterRoles y RoleBindings (RBAC)
- Deployment: `controller-manager`
- Service: Para métricas de Prometheus

---

## Comandos Rápidos con Taskfile

### Generación de Código
```bash
task generate              # Generar CRDs y RBAC
```

### Desarrollo Local
```bash
task install-crds          # Instalar CRDs en el cluster
task run-local             # Ejecutar controlador localmente
task dev                   # install-crds + run-local (quick start)
```

### Deployment en Cluster
```bash
task build                 # Construir imagen Docker
task load-image            # Cargar imagen en Kind
task deploy                # Deployar al cluster
task deploy-all            # build + load-image + deploy
task prod                  # generate + deploy-all (quick deploy)
```

### Testing
```bash
task create-sample         # Crear Book de ejemplo
task list-books            # Listar todos los Books
task get-book BOOK=book-1984  # Ver detalles de un Book
```

### Debugging
```bash
task logs                  # Ver logs del controlador
task pods                  # Listar pods del controlador
```

### Gestión del Cluster
```bash
task create-cluster        # Crear cluster Kind
task delete-cluster        # Eliminar cluster Kind
```

### Limpieza
```bash
task undeploy              # Remover controlador del cluster
task uninstall-crds        # Desinstalar CRDs
task clean-all             # undeploy + uninstall-crds
```

---

## Conceptos Clave

### 1. **Custom Resource Definition (CRD)**
Extiende la API de Kubernetes con nuevos tipos de recursos. En nuestro caso, agregamos el tipo `Book`.

### 2. **Controlador**
Proceso que observa recursos y ejecuta lógica para mantener el estado deseado. Es un loop de reconciliación continuo.

### 3. **Reconcile Loop**
Patrón fundamental de Kubernetes:
- Observar el estado actual
- Comparar con el estado deseado
- Tomar acciones para converger al estado deseado

### 4. **Spec vs Status**
- **Spec**: Lo que el usuario quiere (estado deseado)
- **Status**: Lo que realmente está pasando (estado observado)

### 5. **Conditions**
Forma estándar de reportar estado en Kubernetes:
- `Type`: Nombre de la condición (ej: "Ready")
- `Status`: True, False, Unknown
- `Reason`: Código programático
- `Message`: Mensaje legible para humanos

### 6. **RBAC (Role-Based Access Control)**
Sistema de permisos de Kubernetes. El controlador necesita permisos para leer/escribir recursos.

### 7. **Watch**
Mecanismo de Kubernetes para observar cambios en recursos en tiempo real.

### 8. **Leader Election**
Cuando hay múltiples replicas del controlador, solo una reconcilia activamente (la "leader").

---

## Próximos Pasos

### Fase 2: Relaciones entre Recursos
- Crear CRD **Author**
- Relacionar Books con Authors mediante referencias
- Implementar watches cross-resource
- Actualizar status automáticamente cuando cambie el Author

### Fase 3: Recursos Avanzados
- Crear **BookStore** y **Inventory**
- Implementar relaciones many-to-many
- Usar finalizers para cleanup
- Owner references

### Fase 4: Webhooks
- Webhook de validación para Books
- Webhook de mutación (defaults automáticos)
- Validaciones complejas

### Fase 5: Features Avanzadas
- **BookReservation** con máquina de estados
- RBAC granular por usuario
- Subresources (scale)
- Métricas personalizadas

---

## Recursos Adicionales

- [Kubebuilder Book](https://book.kubebuilder.io/)
- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- [Controller Runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [Proyecto Milo](https://github.com/datum-cloud/milo) - Referencia de arquitectura

---

**Versión:** Fase 1 Completada  
**Última actualización:** Diciembre 2025
