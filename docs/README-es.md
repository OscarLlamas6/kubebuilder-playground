# 📚 BookStore Platform - Guía Completa de Aprendizaje Kubebuilder

## 📑 Índice

- [Introducción](#introducción)
- [Requisitos Previos](#requisitos-previos)
- [Configuración del Cluster Kind](#configuración-del-cluster-kind)
- [Arquitectura del Proyecto](#arquitectura-del-proyecto)
- [Fase 1: CRD Book y Controlador Básico](#fase-1-crd-book-y-controlador-básico)
- [Fase 2: CRD Author y Relaciones entre Recursos](#fase-2-crd-author-y-relaciones-entre-recursos)
- [Desarrollo Local vs Producción](#desarrollo-local-vs-producción)
- [Comandos Rápidos con Taskfile](#comandos-rápidos-con-taskfile)
- [Conceptos Clave de Kubernetes](#conceptos-clave-de-kubernetes)
- [Próximos Pasos](#próximos-pasos)

---

## Introducción

Este proyecto es una **plataforma de gestión de librerías** construida con **Kubebuilder** para aprender desde cero cómo crear **Custom Resource Definitions (CRDs)** y **controladores de Kubernetes**.

### 🎯 Objetivo del Proyecto

Aprender Kubebuilder construyendo un proyecto de mediana complejidad que demuestre:
- Cómo extender la API de Kubernetes con recursos personalizados
- Cómo implementar controladores que mantengan el estado deseado
- Cómo establecer relaciones entre diferentes recursos
- Cómo deployar y operar controladores en producción

### 📚 Recursos que Construiremos

- ✅ **Book** - Recurso básico de libros (Fase 1 - Completada)
- ✅ **Author** - Autores con relaciones a Books (Fase 2 - Completada)
- 🔜 **BookStore** - Librerías físicas (Fase 3)
- 🔜 **Inventory** - Inventario de libros por librería (Fase 3)
- 🔜 **BookReservation** - Reservas con webhooks (Fase 4)
- 🔜 **Review** - Reseñas de libros (Fase 5)

### 💡 ¿Por Qué Este Proyecto?

Este proyecto sirve como un **"Kubebuilder for Dummies"** - una guía completa que cualquier persona con conocimientos básicos de Kubernetes y Go puede seguir para entender cómo funcionan los operadores de Kubernetes.

---

## Requisitos Previos

### Software Necesario

- **Go** 1.23+ - Lenguaje de programación
- **Docker** 17.03+ - Para construir imágenes de contenedores
- **kubectl** v1.11.3+ - Cliente de línea de comandos de Kubernetes
- **kind** - Kubernetes in Docker, para clusters locales
- **kubebuilder** v4.5.1+ - Framework para construir operadores
- **task** - Task runner para comandos simplificados

### Verificar Instalación

```bash
# Verificar Go
go version
# Debe mostrar: go version go1.23.x ...

# Verificar Docker
docker --version
# Debe mostrar: Docker version 17.03.x ...

# Verificar kubectl
kubectl version --client
# Debe mostrar: Client Version: v1.11.3 ...

# Verificar kind
kind version
# Debe mostrar: kind v0.x.x ...

# Verificar kubebuilder
kubebuilder version
# Debe mostrar: Version: 4.5.1 ...

# Verificar task
task --version
# Debe mostrar: Task version: vx.x.x
```

### Conocimientos Previos Recomendados

- **Kubernetes básico**: Pods, Deployments, Services
- **Go básico**: Structs, interfaces, funciones
- **YAML**: Sintaxis básica
- **Git**: Comandos básicos

---

## Configuración del Cluster Kind

### ¿Qué es Kind?

**Kind** (Kubernetes in Docker) es una herramienta que permite crear clusters de Kubernetes locales usando contenedores Docker como nodos. Es ideal para:
- Desarrollo local
- Testing de controladores
- CI/CD pipelines
- Aprendizaje de Kubernetes

### Archivo de Configuración

El archivo `kind/cluster-config.yaml` define cómo se creará nuestro cluster:

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: bookstore-dev

nodes:
  - role: control-plane
  - role: worker
```

### Opciones de Configuración Detalladas

#### 1. **Nodos (Nodes)**

Los nodos son las máquinas (contenedores Docker) que forman el cluster.

```yaml
nodes:
  - role: control-plane  # Nodo maestro que gestiona el cluster
  - role: worker         # Nodo trabajador que ejecuta workloads
```

**¿Qué hace cada rol?**
- `control-plane`: Ejecuta el API Server, etcd, scheduler y controller-manager
- `worker`: Ejecuta los Pods de aplicaciones

**Casos de uso:**
- **1 control-plane**: Desarrollo simple (lo que usamos)
- **3 control-plane + 2 workers**: Simular Alta Disponibilidad
- **1 control-plane + N workers**: Testing de scheduling y distribución

#### 2. **Port Mappings**

Permite exponer puertos del cluster a tu máquina local.

```yaml
extraPortMappings:
  - containerPort: 30080  # Puerto en el nodo Kind
    hostPort: 8080        # Puerto en tu localhost
    protocol: TCP
```

**¿Para qué sirve?**
- Acceder a servicios NodePort desde tu navegador
- Ejemplo: Si tienes un servicio en el puerto 30080, puedes accederlo en `localhost:8080`

#### 3. **Extra Mounts**

Monta directorios de tu máquina en los nodos del cluster.

```yaml
extraMounts:
  - hostPath: /tmp/kind-data    # Directorio en tu máquina
    containerPath: /data         # Directorio en el nodo
    readOnly: false              # Permitir escritura
```

**Casos de uso:**
- Compartir archivos de configuración
- Usar almacenamiento persistente local
- Debugging con logs compartidos

#### 4. **Networking**

Configura las redes internas del cluster.

```yaml
networking:
  podSubnet: "10.244.0.0/16"      # Red para Pods
  serviceSubnet: "10.96.0.0/12"   # Red para Services
  apiServerPort: 6443              # Puerto del API Server
```

**¿Qué significa cada subnet?**
- `podSubnet`: Rango de IPs que se asignarán a los Pods
- `serviceSubnet`: Rango de IPs que se asignarán a los Services
- Útil si tienes conflictos con otras redes locales

#### 5. **Kubeadm Config Patches**

Personaliza la configuración interna de Kubernetes.

```yaml
kubeadmConfigPatches:
  - |
    kind: ClusterConfiguration
    apiServer:
      extraArgs:
        audit-log-path: /var/log/kubernetes/audit.log
        audit-log-maxage: "30"
```

**¿Qué hace esto?**
- `audit-log-path`: Habilita logging de auditoría (quién hizo qué en el cluster)
- `audit-log-maxage`: Mantiene logs por 30 días
- Útil para debugging y compliance

### Comandos para Gestionar el Cluster

#### Crear Cluster

```bash
# Usando el archivo de configuración
kind create cluster --config kind/cluster-config.yaml
```

**¿Qué hace este comando?**
1. Lee el archivo `kind/cluster-config.yaml`
2. Descarga la imagen de Kubernetes (si no existe)
3. Crea contenedores Docker para cada nodo
4. Configura la red entre nodos
5. Instala Kubernetes en cada nodo
6. Configura kubectl para conectarse al cluster

**O usando Taskfile:**
```bash
task create-cluster
```

#### Verificar el Cluster

```bash
# Ver nodos del cluster
kubectl get nodes
# Muestra: NAME, STATUS, ROLES, AGE, VERSION

# Ver información detallada del cluster
kubectl cluster-info
# Muestra: URLs del API Server y servicios core

# Ver contexto actual (a qué cluster está conectado kubectl)
kubectl config current-context
# Muestra: kind-bookstore-dev
```

#### Eliminar el Cluster

```bash
# Eliminar cluster por nombre
kind delete cluster --name bookstore-dev
```

**¿Qué hace este comando?**
1. Detiene todos los contenedores del cluster
2. Elimina los contenedores
3. Limpia la configuración de kubectl

**O usando Taskfile:**
```bash
task delete-cluster
```

---

## Arquitectura del Proyecto

### Estructura de Directorios

```
kubebuilder-playground/
├── api/v1alpha1/              # Definiciones de tipos (CRDs)
│   ├── book_types.go          # Struct Book con Spec y Status
│   ├── author_types.go        # Struct Author con Spec y Status
│   └── groupversion_info.go   # Metadata del grupo API
├── internal/controller/       # Lógica de reconciliación
│   ├── book_controller.go     # Controlador del Book
│   └── author_controller.go   # Controlador del Author
├── config/                    # Manifiestos de Kubernetes
│   ├── crd/                   # CRDs generados automáticamente
│   ├── rbac/                  # Permisos RBAC
│   ├── manager/               # Deployment del controlador
│   └── samples/               # Ejemplos de recursos
├── cmd/main.go               # Punto de entrada del programa
├── kind/                     # Configuración de Kind
│   └── cluster-config.yaml   # Definición del cluster
├── Taskfile.yaml             # Comandos simplificados
└── docs/                     # Documentación
    ├── README-es.md          # Documentación en español
    └── README-en.md          # Documentación en inglés
```

### Flujo de Datos

```
Usuario                    API Server              Controller
  │                            │                        │
  │  kubectl apply book.yaml   │                        │
  ├──────────────────────────>│                        │
  │                            │                        │
  │                            │  Watch detecta cambio  │
  │                            ├───────────────────────>│
  │                            │                        │
  │                            │                        │  Reconcile()
  │                            │                        │  - Lee Book
  │                            │                        │  - Procesa lógica
  │                            │                        │  - Actualiza Status
  │                            │                        │
  │                            │  Update Status         │
  │                            │<───────────────────────┤
  │                            │                        │
  │  kubectl get book          │                        │
  ├──────────────────────────>│                        │
  │                            │                        │
  │  Book con Status           │                        │
  │<───────────────────────────┤                        │
```


---

## Fase 1: CRD Book y Controlador Básico

### Paso 1: Inicializar el Proyecto Kubebuilder

```bash
# Crear directorio del proyecto
mkdir kubebuilder-playground
cd kubebuilder-playground

# Inicializar proyecto Kubebuilder
kubebuilder init --domain bookstore.io --repo github.com/OscarLlamas6/kubebuilder-playground
```

**¿Qué hace `kubebuilder init`?**
- `--domain bookstore.io`: Define el dominio del grupo API (será `catalog.bookstore.io`)
- `--repo`: Define el módulo Go del proyecto
- Crea la estructura básica del proyecto
- Genera `go.mod`, `Makefile`, `Dockerfile`
- Configura el manager principal en `cmd/main.go`

### Paso 2: Crear el API Book

```bash
# Crear API y controlador para Book
kubebuilder create api --group catalog --version v1alpha1 --kind Book --resource --controller
```

**¿Qué hace `kubebuilder create api`?**
- `--group catalog`: Nombre del grupo API (resultado: `catalog.bookstore.io`)
- `--version v1alpha1`: Versión del API (alpha indica que está en desarrollo)
- `--kind Book`: Nombre del recurso (será `Book`)
- `--resource`: Genera el archivo de tipos (`book_types.go`)
- `--controller`: Genera el controlador (`book_controller.go`)

**Archivos generados:**
- `api/v1alpha1/book_types.go` - Definición del CRD
- `internal/controller/book_controller.go` - Lógica del controlador
- `config/crd/bases/catalog.bookstore.io_books.yaml` - CRD manifest
- `config/samples/catalog_v1alpha1_book.yaml` - Ejemplo de Book

### Paso 3: Definir el BookSpec

Editar `api/v1alpha1/book_types.go`:

```go
type BookSpec struct {
    // Title es el título del libro
    // +kubebuilder:validation:Required - Campo obligatorio
    // +kubebuilder:validation:MinLength=1 - Mínimo 1 carácter
    Title string `json:"title"`

    // ISBN es el International Standard Book Number
    // +kubebuilder:validation:Required - Campo obligatorio
    // +kubebuilder:validation:Pattern - Valida formato ISBN-13 o ISBN-10
    ISBN string `json:"isbn"`

    // Author es el nombre del autor (deprecated en Fase 2)
    // +kubebuilder:validation:Optional - Campo opcional
    Author string `json:"author,omitempty"`

    // Publisher es la editorial
    // +kubebuilder:validation:Optional
    Publisher string `json:"publisher,omitempty"`

    // PublishedYear es el año de publicación
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=1000 - Año mínimo válido
    // +kubebuilder:validation:Maximum=2100 - Año máximo válido
    PublishedYear int `json:"publishedYear,omitempty"`

    // Genre es el género literario
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Enum - Solo valores permitidos
    Genre string `json:"genre,omitempty"`

    // Pages es el número de páginas
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=1
    Pages int `json:"pages,omitempty"`
}
```

**Explicación de las anotaciones:**
- `+kubebuilder:validation:Required`: El campo debe tener un valor
- `+kubebuilder:validation:MinLength=1`: Longitud mínima de string
- `+kubebuilder:validation:Pattern`: Expresión regular para validar formato
- `+kubebuilder:validation:Minimum/Maximum`: Rango de valores numéricos
- `+kubebuilder:validation:Enum`: Lista de valores permitidos
- `+kubebuilder:validation:Optional`: Campo no obligatorio

### Paso 4: Definir el BookStatus

```go
type BookStatus struct {
    // Conditions representa el estado actual del libro
    // Usa el patrón estándar de Kubernetes para reportar estado
    // +kubebuilder:validation:Optional
    Conditions []metav1.Condition `json:"conditions,omitempty"`

    // AvailableCopies es el número de copias disponibles
    // Se calculará en fases futuras basado en Inventory
    // +kubebuilder:validation:Optional
    AvailableCopies int `json:"availableCopies,omitempty"`
}
```

**¿Qué son las Conditions?**
Las Conditions son el patrón estándar de Kubernetes para reportar estado:
- `Type`: Nombre de la condición (ej: "Ready", "Available")
- `Status`: True, False, o Unknown
- `Reason`: Código programático (ej: "BookRegistered")
- `Message`: Mensaje legible para humanos
- `LastTransitionTime`: Cuándo cambió el estado

### Paso 5: Agregar Print Columns

Agregar antes de `type Book struct`:

```go
// +kubebuilder:printcolumn:name="Title",type=string,JSONPath=`.spec.title`
// +kubebuilder:printcolumn:name="Author",type=string,JSONPath=`.spec.author`
// +kubebuilder:printcolumn:name="ISBN",type=string,JSONPath=`.spec.isbn`
// +kubebuilder:printcolumn:name="Genre",type=string,JSONPath=`.spec.genre`
// +kubebuilder:printcolumn:name="Available",type=integer,JSONPath=`.status.availableCopies`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
```

**¿Qué hacen los printcolumns?**
Definen qué columnas se muestran cuando ejecutas `kubectl get books`:
- `name`: Nombre de la columna
- `type`: Tipo de dato (string, integer, date)
- `JSONPath`: Ruta al campo en el objeto

### Paso 6: Implementar el Book Controller

Editar `internal/controller/book_controller.go`:

```go
func (r *BookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)
    
    // 1. Obtener el Book del cluster
    var book catalogv1alpha1.Book
    if err := r.Get(ctx, req.NamespacedName, &book); err != nil {
        if apierrors.IsNotFound(err) {
            // Book fue eliminado, no hacer nada
            return ctrl.Result{}, nil
        }
        // Error al leer el objeto
        return ctrl.Result{}, err
    }

    // 2. Crear condición Ready
    readyCondition := metav1.Condition{
        Type:               "Ready",
        Status:             metav1.ConditionTrue,
        Reason:             "BookRegistered",
        Message:            "Book has been successfully registered in the catalog",
        ObservedGeneration: book.Generation,
    }

    // 3. Actualizar status si la condición cambió
    if apimeta.SetStatusCondition(&book.Status.Conditions, readyCondition) {
        if err := r.Status().Update(ctx, &book); err != nil {
            return ctrl.Result{}, err
        }
    }

    return ctrl.Result{}, nil
}
```

**Explicación línea por línea:**

1. **`logger := log.FromContext(ctx)`**
   - Obtiene un logger del contexto para escribir logs estructurados

2. **`var book catalogv1alpha1.Book`**
   - Declara una variable para almacenar el Book

3. **`r.Get(ctx, req.NamespacedName, &book)`**
   - `r.Get`: Método del cliente de Kubernetes para leer recursos
   - `ctx`: Contexto con timeout y cancelación
   - `req.NamespacedName`: Nombre y namespace del Book a leer
   - `&book`: Puntero donde se guardará el resultado

4. **`apierrors.IsNotFound(err)`**
   - Verifica si el error es porque el recurso no existe
   - Si fue eliminado, retornamos sin error (reconciliación exitosa)

5. **`metav1.Condition{...}`**
   - Crea una condición siguiendo el patrón estándar de Kubernetes
   - `ObservedGeneration`: Versión del recurso que observamos

6. **`apimeta.SetStatusCondition(&book.Status.Conditions, readyCondition)`**
   - Agrega o actualiza la condición en el status
   - Retorna `true` si hubo cambios

7. **`r.Status().Update(ctx, &book)`**
   - Actualiza solo el subrecurso `/status` del Book
   - Esto es posible gracias a `+kubebuilder:subresource:status`

### Paso 7: Generar Manifiestos

```bash
# Generar CRDs y RBAC
make manifests
```

**¿Qué hace `make manifests`?**
- Ejecuta `controller-gen` para procesar las anotaciones `+kubebuilder:`
- Genera el CRD en `config/crd/bases/catalog.bookstore.io_books.yaml`
- Genera roles RBAC en `config/rbac/`
- Actualiza el esquema OpenAPI del CRD con las validaciones

### Paso 8: Probar Localmente

```bash
# 1. Crear cluster Kind
task create-cluster

# 2. Instalar CRDs
task install-crds

# 3. Ejecutar controlador localmente
task run-local
```

**¿Qué pasa en cada paso?**

**`task create-cluster`:**
- Ejecuta `kind create cluster --config kind/cluster-config.yaml`
- Crea contenedores Docker para los nodos
- Configura kubectl para conectarse al cluster

**`task install-crds`:**
- Ejecuta `make install`
- Aplica los CRDs al cluster con `kubectl apply`
- Ahora Kubernetes conoce el tipo `Book`

**`task run-local`:**
- Ejecuta `make run`
- Compila el controlador como binario Go
- Lo ejecuta en tu máquina
- Se conecta al cluster usando `~/.kube/config`
- Observa cambios en Books y reconcilia

### Paso 9: Crear un Book

En otra terminal:

```bash
# Crear un Book de ejemplo
kubectl apply -f config/samples/catalog_v1alpha1_book.yaml

# Ver el Book creado
kubectl get books

# Ver detalles del Book
kubectl get book book-sample-1 -o yaml
```

**¿Qué verás?**
```yaml
status:
  conditions:
  - type: Ready
    status: "True"
    reason: BookRegistered
    message: "Book has been successfully registered in the catalog"
    lastTransitionTime: "2026-01-18T..."
```

### Paso 10: Deployar en el Cluster

```bash
# Construir imagen, cargarla en Kind y deployar
task prod
```

**¿Qué hace `task prod`?**
1. **`make manifests`**: Regenera CRDs y RBAC
2. **`make docker-build`**: Construye imagen Docker del controlador
3. **`kind load docker-image`**: Carga la imagen en el cluster Kind
4. **`make deploy`**: Aplica todos los manifiestos al cluster

**Recursos creados:**
- Namespace: `kubebuilder-playground-system`
- ServiceAccount: `controller-manager`
- ClusterRole: Permisos para leer/escribir Books
- ClusterRoleBinding: Asocia el ServiceAccount con el ClusterRole
- Deployment: Pod con el controlador
- Service: Para métricas de Prometheus


---

## Fase 2: CRD Author y Relaciones entre Recursos

### Introducción a la Fase 2

En la Fase 1 creamos un CRD simple (Book) con un controlador básico. En la Fase 2 aprenderemos conceptos más avanzados:

- **Relaciones entre recursos**: Cómo un recurso puede referenciar a otro
- **Cross-resource watches**: Cómo un controlador puede observar múltiples tipos de recursos
- **Status propagation**: Cómo cachear información de recursos relacionados
- **Bidirectional updates**: Cómo mantener sincronizados recursos relacionados

### Paso 1: Crear el CRD Author

```bash
# Crear API y controlador para Author
kubebuilder create api --group catalog --version v1alpha1 --kind Author --resource --controller
```

**¿Qué genera este comando?**
- `api/v1alpha1/author_types.go` - Definición del Author
- `internal/controller/author_controller.go` - Controlador del Author
- `config/crd/bases/catalog.bookstore.io_authors.yaml` - CRD manifest
- Actualiza `cmd/main.go` para registrar el nuevo controlador

### Paso 2: Definir el AuthorSpec

Editar `api/v1alpha1/author_types.go`:

```go
type AuthorSpec struct {
    // Name es el nombre completo del autor
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:MinLength=1
    // +kubebuilder:validation:MaxLength=100
    Name string `json:"name"`

    // Biography es una breve descripción del autor
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:MaxLength=1000
    Biography string `json:"biography,omitempty"`

    // BirthYear es el año de nacimiento
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=1000
    // +kubebuilder:validation:Maximum=2100
    BirthYear int `json:"birthYear,omitempty"`

    // DeathYear es el año de fallecimiento (si aplica)
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=1000
    // +kubebuilder:validation:Maximum=2100
    DeathYear int `json:"deathYear,omitempty"`

    // Nationality es la nacionalidad del autor
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:MaxLength=50
    Nationality string `json:"nationality,omitempty"`

    // Website es el sitio web oficial del autor
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Pattern=`^https?://.*`
    Website string `json:"website,omitempty"`
}
```

**Explicación de validaciones nuevas:**
- `MaxLength`: Longitud máxima del string
- `Pattern=^https?://.*`: URL debe empezar con http:// o https://

### Paso 3: Definir el AuthorStatus

```go
type AuthorStatus struct {
    // Conditions representa el estado del autor
    // +kubebuilder:validation:Optional
    Conditions []metav1.Condition `json:"conditions,omitempty"`

    // BookCount es el número de libros de este autor en el catálogo
    // Se calcula automáticamente por el controlador
    // +kubebuilder:validation:Optional
    BookCount int `json:"bookCount,omitempty"`
}
```

**¿Por qué BookCount en el status?**
- El status es calculado por el controlador, no por el usuario
- BookCount se actualiza automáticamente cuando se crean/eliminan Books
- Es información derivada, no parte del estado deseado

### Paso 4: Agregar Print Columns al Author

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

### Paso 5: Actualizar Book para Usar Referencias

Editar `api/v1alpha1/book_types.go`:

**Agregar nuevos tipos:**

```go
// AuthorReference contiene una referencia a un recurso Author
type AuthorReference struct {
    // Name es el nombre del recurso Author
    // +kubebuilder:validation:Required
    Name string `json:"name"`
}

// AuthorInfo contiene información cacheada del autor
type AuthorInfo struct {
    // Name es el nombre completo del autor
    // +kubebuilder:validation:Optional
    Name string `json:"name,omitempty"`

    // Nationality es la nacionalidad del autor
    // +kubebuilder:validation:Optional
    Nationality string `json:"nationality,omitempty"`

    // BirthYear es el año de nacimiento
    // +kubebuilder:validation:Optional
    BirthYear int `json:"birthYear,omitempty"`
}
```

**¿Por qué crear estos tipos?**
- `AuthorReference`: Define cómo un Book referencia a un Author
- `AuthorInfo`: Cachea información del Author en el Book para consultas rápidas

**Actualizar BookSpec:**

```go
type BookSpec struct {
    Title string `json:"title"`
    ISBN  string `json:"isbn"`
    
    // Author (deprecated) - Mantener para compatibilidad
    // +kubebuilder:validation:Optional
    Author string `json:"author,omitempty"`

    // AuthorRef es la referencia al recurso Author
    // +kubebuilder:validation:Optional
    AuthorRef *AuthorReference `json:"authorRef,omitempty"`
    
    // ... resto de campos
}
```

**¿Por qué mantener el campo `author`?**
- Compatibilidad hacia atrás con Books existentes
- Permite migración gradual
- El controlador soporta ambos formatos

**Actualizar BookStatus:**

```go
type BookStatus struct {
    Conditions      []metav1.Condition `json:"conditions,omitempty"`
    AvailableCopies int                `json:"availableCopies,omitempty"`
    
    // AuthorInfo contiene información cacheada del Author referenciado
    // +kubebuilder:validation:Optional
    AuthorInfo *AuthorInfo `json:"authorInfo,omitempty"`
}
```

### Paso 6: Implementar el Author Controller

Editar `internal/controller/author_controller.go`:

```go
func (r *AuthorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    // 1. Obtener el Author
    var author catalogv1alpha1.Author
    if err := r.Get(ctx, req.NamespacedName, &author); err != nil {
        if apierrors.IsNotFound(err) {
            return ctrl.Result{}, nil
        }
        return ctrl.Result{}, err
    }

    logger.Info("Reconciling Author", "name", author.Name, "authorName", author.Spec.Name)

    // 2. Listar todos los Books
    var bookList catalogv1alpha1.BookList
    if err := r.List(ctx, &bookList); err != nil {
        logger.Error(err, "Failed to list Books")
        return ctrl.Result{}, err
    }

    // 3. Contar libros que referencian este autor
    bookCount := 0
    for _, book := range bookList.Items {
        // Verificar campo author (string) - formato antiguo
        if book.Spec.Author == author.Spec.Name {
            bookCount++
        }
        // Verificar authorRef - formato nuevo
        if book.Spec.AuthorRef != nil && book.Spec.AuthorRef.Name == author.Name {
            bookCount++
        }
    }

    // 4. Actualizar status
    author.Status.BookCount = bookCount

    // 5. Establecer condición Ready
    readyCondition := metav1.Condition{
        Type:               "Ready",
        Status:             metav1.ConditionTrue,
        Reason:             "AuthorRegistered",
        Message:            "Author has been successfully registered in the catalog",
        ObservedGeneration: author.Generation,
    }

    // 6. Actualizar status si cambió
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

**Explicación del código:**

**Líneas 1-10**: Obtener el Author del cluster (igual que en Book)

**Líneas 12-16**: Listar todos los Books del cluster
- `r.List(ctx, &bookList)`: Lista todos los Books
- No filtramos por namespace porque queremos contar todos

**Líneas 18-27**: Contar libros del autor
- Iteramos sobre todos los Books
- Verificamos ambos formatos: `author` (string) y `authorRef`
- Incrementamos contador si coincide

**Líneas 29-30**: Actualizar el campo `bookCount` en el status

**Líneas 32-38**: Crear condición Ready (patrón estándar)

**Líneas 40-48**: Actualizar status solo si hubo cambios
- `apimeta.SetStatusCondition` retorna `true` si cambió algo
- Solo hacemos `Update` si es necesario (optimización)

### Paso 7: Actualizar el Book Controller

Editar `internal/controller/book_controller.go`:

```go
func (r *BookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    logger := log.FromContext(ctx)

    // 1. Obtener el Book
    var book catalogv1alpha1.Book
    if err := r.Get(ctx, req.NamespacedName, &book); err != nil {
        if apierrors.IsNotFound(err) {
            return ctrl.Result{}, nil
        }
        return ctrl.Result{}, err
    }

    // 2. Si el Book tiene authorRef, buscar el Author y cachear su info
    var authorFoundCondition metav1.Condition
    if book.Spec.AuthorRef != nil {
        var author catalogv1alpha1.Author
        authorKey := client.ObjectKey{
            Name:      book.Spec.AuthorRef.Name,
            Namespace: book.Namespace,
        }

        if err := r.Get(ctx, authorKey, &author); err != nil {
            if apierrors.IsNotFound(err) {
                // Author no existe
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
                // Error al buscar Author
                logger.Error(err, "Failed to get Author")
                return ctrl.Result{}, err
            }
        } else {
            // Author encontrado, cachear información
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

    // 3. Establecer condición Ready
    readyCondition := metav1.Condition{
        Type:               "Ready",
        Status:             metav1.ConditionTrue,
        Reason:             "BookRegistered",
        Message:            "Book has been successfully registered in the catalog",
        ObservedGeneration: book.Generation,
    }
    apimeta.SetStatusCondition(&book.Status.Conditions, readyCondition)

    // 4. Actualizar status
    logger.Info("Updating Book status", "title", book.Spec.Title)
    if err := r.Status().Update(ctx, &book); err != nil {
        logger.Error(err, "Failed to update Book status")
        return ctrl.Result{}, err
    }
    logger.Info("Book status updated successfully")

    return ctrl.Result{}, nil
}
```

**Explicación detallada:**

**Líneas 14-15**: Verificar si el Book tiene una referencia a Author
- `book.Spec.AuthorRef != nil`: Solo procesamos si hay referencia

**Líneas 16-20**: Preparar la búsqueda del Author
- `client.ObjectKey`: Estructura con Name y Namespace
- Usamos el mismo namespace que el Book

**Líneas 22-34**: Manejar caso donde Author no existe
- Creamos condición `AuthorFound` con status `False`
- Limpiamos `AuthorInfo` del status (ya no es válido)
- No retornamos error, solo reportamos el estado

**Líneas 39-52**: Manejar caso donde Author existe
- Cacheamos información del Author en `book.Status.AuthorInfo`
- Solo guardamos campos relevantes (no toda la información)
- Creamos condición `AuthorFound` con status `True`

**¿Por qué cachear información?**
- Consultas más rápidas: No necesitas hacer GET del Author cada vez
- Menos carga en el API Server
- Información disponible en `kubectl get book -o yaml`

### Paso 8: Implementar Cross-Resource Watches

**¿Qué son los Cross-Resource Watches?**

Permiten que un controlador observe cambios en recursos de diferente tipo. Por ejemplo:
- Book controller observa Authors
- Cuando un Author cambia, reconcilia todos los Books que lo referencian

**Actualizar Book Controller - SetupWithManager:**

```go
// findBooksForAuthor retorna reconcile requests para todos los Books que referencian el Author dado
func (r *BookReconciler) findBooksForAuthor(ctx context.Context, author client.Object) []reconcile.Request {
    logger := log.FromContext(ctx)

    // Listar todos los Books en el mismo namespace
    var bookList catalogv1alpha1.BookList
    if err := r.List(ctx, &bookList, client.InNamespace(author.GetNamespace())); err != nil {
        logger.Error(err, "Failed to list Books for Author", "author", author.GetName())
        return []reconcile.Request{}
    }

    // Filtrar Books que referencian este Author
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
        // Observar Authors y reconciliar Books que los referencian
        Watches(
            &catalogv1alpha1.Author{},
            handler.EnqueueRequestsFromMapFunc(r.findBooksForAuthor),
        ).
        Named("book").
        Complete(r)
}
```

**Explicación del código:**

**`findBooksForAuthor` - Función Mapper:**
- Recibe un Author que cambió
- Retorna lista de Books que deben reconciliarse
- Es llamada automáticamente por el framework cuando un Author cambia

**Líneas 6-10**: Listar Books en el mismo namespace
- `client.InNamespace(author.GetNamespace())`: Filtra por namespace
- Optimización: No buscamos en todos los namespaces

**Líneas 13-26**: Crear requests de reconciliación
- Iteramos sobre Books
- Si `authorRef.Name` coincide, agregamos a la lista
- Cada request hará que `Reconcile()` se ejecute para ese Book

**`SetupWithManager`:**
- `For(&catalogv1alpha1.Book{})`: Recurso principal que maneja
- `Watches(&catalogv1alpha1.Author{}, ...)`: Observa Authors también
- `handler.EnqueueRequestsFromMapFunc`: Usa nuestra función mapper

**Actualizar Author Controller - SetupWithManager:**

```go
// findAuthorsForBook retorna reconcile request para el Author referenciado por el Book
func (r *AuthorReconciler) findAuthorsForBook(ctx context.Context, book client.Object) []reconcile.Request {
    logger := log.FromContext(ctx)

    // Type assert a Book
    bookObj, ok := book.(*catalogv1alpha1.Book)
    if !ok {
        logger.Error(nil, "Failed to cast object to Book")
        return []reconcile.Request{}
    }

    // Si el Book tiene authorRef, encolar el Author
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
        // Observar Books y reconciliar Authors cuando Books cambian
        Watches(
            &catalogv1alpha1.Book{},
            handler.EnqueueRequestsFromMapFunc(r.findAuthorsForBook),
        ).
        Named("author").
        Complete(r)
}
```

**¿Por qué necesitamos watches bidireccionales?**

**Book → Author:**
- Cuando un Author cambia (ej: actualiza biografía)
- Necesitamos actualizar `authorInfo` en todos sus Books

**Author → Book:**
- Cuando un Book se crea/elimina
- Necesitamos actualizar `bookCount` en el Author

### Paso 9: Generar y Deployar

```bash
# Generar manifiestos
make manifests generate

# Deployar todo
task deploy-all
```

**¿Qué se genera?**
- CRD de Author en `config/crd/bases/`
- Roles RBAC para Author en `config/rbac/`
- Actualización del CRD de Book con nuevos campos
- Actualización de permisos RBAC para cross-resource access

### Paso 10: Crear Authors y Books

**Crear Authors:**

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

**Crear Book con authorRef:**

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
    name: george-orwell  # Referencia al Author
  publisher: "Penguin Books"
  publishedYear: 1949
  genre: "Fiction"
  pages: 328
```

### Paso 11: Verificar Relaciones

```bash
# Ver Authors con bookCount
kubectl get authors

# Salida:
# NAME           NAME            NATIONALITY   BIRTH YEAR   BOOKS   AGE
# george-orwell  George Orwell   British       1903         1       5m

# Ver Book con authorInfo
kubectl get book book-1984 -o yaml
```

**Salida esperada:**

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

### Paso 12: Probar Cross-Resource Watches

**Actualizar el Author:**

```bash
kubectl patch author george-orwell --type='json' -p='[{"op": "replace", "path": "/spec/biography", "value": "Updated biography"}]'
```

**¿Qué sucede?**
1. Author controller reconcilia el Author
2. Book controller detecta el cambio (gracias al watch)
3. Book controller reconcilia todos los Books de ese Author
4. `authorInfo` se actualiza automáticamente en los Books

**Verificar:**

```bash
# Ver logs del controlador
kubectl logs -n kubebuilder-playground-system deployment/kubebuilder-playground-controller-manager --tail=50

# Deberías ver:
# "Queueing Book for reconciliation due to Author change" book="book-1984" author="george-orwell"
# "Author found, updating Book status" author="George Orwell"
```

### Conceptos Clave de la Fase 2

#### 1. **Resource References**

**¿Qué son?**
- Forma de relacionar recursos en Kubernetes
- Similar a "foreign keys" en bases de datos

**Tipos de referencias:**
- **Name-based**: Solo guardamos el nombre (lo que usamos)
- **UID-based**: Guardamos el UID único
- **Owner References**: Relación padre-hijo con garbage collection

**¿Cuándo usar cada una?**
- Name-based: Relaciones simples, fácil de leer
- UID-based: Cuando necesitas garantizar identidad única
- Owner References: Cuando el hijo debe eliminarse con el padre

#### 2. **Status Propagation**

**¿Qué es?**
- Copiar información de un recurso a otro
- Cachear datos para consultas rápidas

**Ventajas:**
- Menos llamadas al API Server
- Información disponible en un solo GET
- Mejor rendimiento

**Desventajas:**
- Datos pueden quedar desactualizados
- Necesitas watches para mantener sincronizado
- Más complejidad en el controlador

#### 3. **Cross-Resource Watches**

**¿Qué son?**
- Observar cambios en recursos de diferente tipo
- Reconciliar automáticamente cuando cambian

**Patrón Mapper Function:**
```go
func (r *Reconciler) findRelatedResources(ctx context.Context, obj client.Object) []reconcile.Request {
    // 1. Identificar qué recursos están relacionados
    // 2. Retornar lista de requests para reconciliarlos
}
```

**Cuándo usar:**
- Relaciones entre recursos
- Información derivada que debe actualizarse
- Mantener consistencia entre recursos

#### 4. **Bidirectional Updates**

**¿Qué son?**
- Dos controladores que se observan mutuamente
- Cambios en A afectan B, y viceversa

**Cuidado con loops infinitos:**
- Usar `ObservedGeneration` para detectar cambios reales
- Solo actualizar status si realmente cambió algo
- `apimeta.SetStatusCondition` retorna `false` si no hay cambios

### Diagrama de Flujo Completo

```
Usuario crea Book con authorRef
         ↓
Book Controller reconcilia
         ↓
Busca Author referenciado
         ↓
    ¿Existe?
    ↙     ↘
  Sí      No
   ↓       ↓
Cachea   AuthorFound=False
info      ↓
   ↓    Actualiza
AuthorFound=True  Book Status
   ↓       ↓
Actualiza ←┘
Book Status
   ↓
Author Controller detecta cambio (watch)
   ↓
Reconcilia Author
   ↓
Cuenta Books que lo referencian
   ↓
Actualiza bookCount
   ↓
Book Controller detecta cambio en Author (watch)
   ↓
Reconcilia Books del Author
   ↓
Actualiza authorInfo
```

### Comandos Útiles para Debugging

```bash
# Ver eventos del cluster
kubectl get events --sort-by='.lastTimestamp'

# Ver logs con filtro
kubectl logs -n kubebuilder-playground-system deployment/kubebuilder-playground-controller-manager | grep -i author

# Describir un Book
kubectl describe book book-1984

# Ver RBAC del controlador
kubectl get clusterrole kubebuilder-playground-manager-role -o yaml

# Forzar reconciliación
kubectl annotate book book-1984 reconcile=now --overwrite
```

---


## Desarrollo Local vs Producción

### Modo Desarrollo Local

**Comando:** `task dev` o `task run-local`

**Arquitectura:**
```
Tu Laptop                          Cluster Kind
┌──────────────┐                  ┌─────────────┐
│ Controlador  │ ←── kubectl ───→ │  API Server │
│ (proceso Go) │      watch        │     ↓       │
│   ↓          │                   │   etcd      │
│ Reconcile()  │                   │     ↓       │
│   ↓          │                   │   CRDs      │
│ Update       │ ───────────────→  │   Books     │
│  Status      │                   │   Authors   │
└──────────────┘                  └─────────────┘
```

**Características:**
- ✅ Controlador corre en tu máquina como proceso Go
- ✅ Se conecta al cluster usando `~/.kube/config`
- ✅ Logs directos en tu terminal
- ✅ Rápido para iterar y debuggear
- ✅ Hot reload con `go run`
- ❌ Solo funciona mientras tu laptop esté encendida
- ❌ No simula condiciones de producción

**Pasos:**
```bash
# 1. Generar manifiestos
task generate

# 2. Instalar CRDs en el cluster
task install-crds
# Ejecuta: kubectl apply -f config/crd/bases/

# 3. Ejecutar controlador localmente
task run-local
# Ejecuta: go run cmd/main.go
```

**¿Cuándo usar modo local?**
- Desarrollo activo de features
- Debugging de lógica del controlador
- Testing rápido de cambios
- Aprendizaje y experimentación

### Modo Producción en Cluster

**Comando:** `task prod`

**Arquitectura:**
```
Cluster Kind
┌────────────────────────────────────────┐
│  Namespace: kubebuilder-playground-    │
│             system                     │
│  ┌──────────────────────────────────┐  │
│  │ Deployment: controller-manager   │  │
│  │   Replicas: 1                    │  │
│  │   ┌────────────────────────────┐ │  │
│  │   │ Pod: controller-xyz        │ │  │
│  │   │   Container: manager       │ │  │
│  │   │     Image: controller:v1   │ │  │
│  │   │     ↓                      │ │  │
│  │   │   Reconcile Loop           │ │  │
│  │   └────────────────────────────┘ │  │
│  │         ↓ watch                  │  │
│  │   ServiceAccount + RBAC          │  │
│  └──────────────────────────────────┘  │
│           ↓                            │
│  ┌──────────────────────────────────┐  │
│  │ Namespace: default               │  │
│  │   Books, Authors                 │  │
│  └──────────────────────────────────┘  │
└────────────────────────────────────────┘
```

**Características:**
- ✅ Controlador corre como Pod dentro del cluster
- ✅ Alta disponibilidad (con múltiples replicas)
- ✅ Corre 24/7
- ✅ Exactamente como producción real
- ✅ RBAC configurado correctamente
- ✅ Métricas y health checks
- ❌ Más lento para iterar
- ❌ Logs requieren kubectl

**Pasos detallados:**

```bash
# 1. Generar manifiestos actualizados
task generate
# Ejecuta: make manifests
# - controller-gen procesa anotaciones +kubebuilder:
# - Genera CRDs en config/crd/bases/
# - Genera RBAC en config/rbac/

# 2. Construir imagen Docker
task build
# Ejecuta: make docker-build IMG=bookstore-controller:v1
# - Compila binario Go con CGO_ENABLED=0
# - Crea imagen Docker multi-stage
# - Imagen base: gcr.io/distroless/static:nonroot
# - Tamaño final: ~20MB

# 3. Cargar imagen en Kind
task load-image
# Ejecuta: kind load docker-image bookstore-controller:v1 --name bookstore-dev
# - Copia imagen de Docker local a nodos Kind
# - Evita necesidad de registry externo

# 4. Deployar al cluster
task deploy
# Ejecuta: make deploy IMG=bookstore-controller:v1
# - Aplica CRDs
# - Crea namespace system
# - Crea ServiceAccount
# - Aplica RBAC (ClusterRole + ClusterRoleBinding)
# - Crea Deployment del controlador
# - Crea Service para métricas

# O todo en un comando:
task prod
# Ejecuta: generate + build + load-image + deploy
```

**Recursos creados:**

1. **Namespace:** `kubebuilder-playground-system`
   - Aísla recursos del controlador

2. **ServiceAccount:** `controller-manager`
   - Identidad del Pod del controlador

3. **ClusterRole:** `manager-role`
   - Permisos para:
     - get, list, watch, create, update, patch, delete en books
     - get, list, watch, create, update, patch, delete en authors
     - get, update, patch en books/status y authors/status

4. **ClusterRoleBinding:** `manager-rolebinding`
   - Asocia ServiceAccount con ClusterRole

5. **Deployment:** `controller-manager`
   - 1 replica del controlador
   - Liveness probe en `/healthz`
   - Readiness probe en `/readyz`
   - Resource limits configurados

6. **Service:** `controller-manager-metrics-service`
   - Expone métricas de Prometheus en puerto 8443

**¿Cuándo usar modo producción?**
- Testing de deployment
- Validar RBAC
- Testing de alta disponibilidad
- Simular condiciones reales
- Antes de deployar a producción real

---

## Comandos Rápidos con Taskfile

### ¿Qué es Taskfile?

Taskfile es un task runner similar a Make pero más simple y con mejor sintaxis. Definimos comandos comunes en `Taskfile.yaml`.

### Generación de Código

```bash
task generate
# Ejecuta: make manifests
# Genera: CRDs y RBAC basados en anotaciones +kubebuilder:
```

**¿Cuándo ejecutar?**
- Después de modificar anotaciones en `*_types.go`
- Después de cambiar RBAC en controladores
- Antes de deployar cambios

### Desarrollo Local

```bash
task install-crds
# Ejecuta: make install
# Instala: CRDs en el cluster actual

task run-local
# Ejecuta: make run
# Corre: Controlador como proceso Go local

task dev
# Ejecuta: generate + install-crds + run-local
# Uso: Quick start para desarrollo
```

### Deployment en Cluster

```bash
task build
# Ejecuta: make docker-build IMG=bookstore-controller:v1
# Construye: Imagen Docker del controlador

task load-image
# Ejecuta: kind load docker-image bookstore-controller:v1
# Carga: Imagen en cluster Kind

task deploy
# Ejecuta: make deploy IMG=bookstore-controller:v1
# Deploya: Controlador en el cluster

task deploy-all
# Ejecuta: build + load-image + deploy
# Uso: Deploy completo en un comando

task prod
# Ejecuta: generate + deploy-all
# Uso: Quick deploy desde código hasta cluster
```

### Testing y Verificación

```bash
task list-books
# Ejecuta: kubectl get books
# Muestra: Todos los Books con print columns

task get-book BOOK=book-1984
# Ejecuta: kubectl get book book-1984 -o yaml
# Muestra: Detalles completos de un Book

kubectl get authors
# Muestra: Todos los Authors con bookCount
```

### Debugging

```bash
task logs
# Ejecuta: kubectl logs -n kubebuilder-playground-system deployment/controller-manager --tail=50 -f
# Muestra: Logs del controlador en tiempo real

task pods
# Ejecuta: kubectl get pods -n kubebuilder-playground-system
# Muestra: Estado de los Pods del controlador
```

### Gestión del Cluster

```bash
task create-cluster
# Ejecuta: kind create cluster --config kind/cluster-config.yaml
# Crea: Cluster Kind con configuración personalizada

task delete-cluster
# Ejecuta: kind delete cluster --name bookstore-dev
# Elimina: Cluster Kind completamente
```

### Limpieza

```bash
task undeploy
# Ejecuta: make undeploy
# Elimina: Deployment del controlador (mantiene CRDs)

task uninstall-crds
# Ejecuta: make uninstall
# Elimina: CRDs del cluster (elimina todos los Books/Authors)

task clean-all
# Ejecuta: undeploy + uninstall-crds
# Limpieza: Completa del cluster
```

---

## Conceptos Clave de Kubernetes

### 1. Custom Resource Definition (CRD)

**¿Qué es?**
Extiende la API de Kubernetes con nuevos tipos de recursos.

**¿Cómo funciona?**
- Defines el esquema en Go (BookSpec, BookStatus)
- Kubebuilder genera el CRD YAML
- Kubernetes valida recursos contra el esquema
- API Server almacena recursos en etcd

**Ventajas:**
- API nativa de Kubernetes
- Validación automática
- Versionado (v1alpha1, v1beta1, v1)
- Integración con kubectl

### 2. Controlador

**¿Qué es?**
Proceso que observa recursos y ejecuta lógica para mantener el estado deseado.

**Patrón Reconcile Loop:**
```
┌─────────────────────────────────────┐
│  Watch detecta cambio en recurso    │
└──────────────┬──────────────────────┘
               ↓
┌─────────────────────────────────────┐
│  Reconcile() se ejecuta             │
│  1. Lee estado actual               │
│  2. Compara con estado deseado      │
│  3. Toma acciones para converger    │
│  4. Actualiza status                │
└──────────────┬──────────────────────┘
               ↓
┌─────────────────────────────────────┐
│  ¿Necesita requeue?                 │
│  - Sí: Reconcile() se ejecuta otra  │
│        vez después de X tiempo      │
│  - No: Espera próximo cambio        │
└─────────────────────────────────────┘
```

**Características:**
- Idempotente: Puede ejecutarse múltiples veces
- Eventual consistency: Converge al estado deseado
- Error handling: Reintenta automáticamente

### 3. Spec vs Status

**Spec (Especificación):**
- Estado **deseado** del recurso
- Lo que el **usuario** configura
- Inmutable hasta que el usuario lo actualice
- Ejemplo: `book.Spec.Title = "1984"`

**Status:**
- Estado **observado** del recurso
- Lo que el **controlador** reporta
- Solo el controlador puede modificarlo
- Ejemplo: `book.Status.Conditions = [Ready]`

**¿Por qué separar Spec y Status?**
- Claridad: Separación de responsabilidades
- Seguridad: Usuarios no pueden falsificar status
- Observabilidad: Fácil ver si el estado deseado se logró

### 4. Conditions

**¿Qué son?**
Forma estándar de reportar estado en Kubernetes.

**Estructura:**
```go
type Condition struct {
    Type               string      // "Ready", "Available", "Progressing"
    Status             string      // "True", "False", "Unknown"
    Reason             string      // "BookRegistered", "AuthorNotFound"
    Message            string      // Mensaje legible para humanos
    LastTransitionTime metav1.Time // Cuándo cambió el status
    ObservedGeneration int64       // Versión del recurso observada
}
```

**Ejemplo:**
```yaml
conditions:
- type: Ready
  status: "True"
  reason: BookRegistered
  message: "Book has been successfully registered"
  lastTransitionTime: "2026-01-18T12:00:00Z"
  observedGeneration: 1
```

**Mejores prácticas:**
- Usar tipos estándar cuando sea posible (Ready, Available)
- Reason debe ser CamelCase sin espacios
- Message debe ser descriptivo y útil
- Actualizar LastTransitionTime solo cuando Status cambia

### 5. RBAC (Role-Based Access Control)

**¿Qué es?**
Sistema de permisos de Kubernetes.

**Componentes:**
- **ServiceAccount**: Identidad del Pod
- **Role/ClusterRole**: Conjunto de permisos
- **RoleBinding/ClusterRoleBinding**: Asocia identidad con permisos

**Ejemplo en nuestro proyecto:**
```yaml
# ServiceAccount
apiVersion: v1
kind: ServiceAccount
metadata:
  name: controller-manager
  namespace: system

---
# ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: manager-role
rules:
- apiGroups: ["catalog.bookstore.io"]
  resources: ["books", "authors"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
- apiGroups: ["catalog.bookstore.io"]
  resources: ["books/status", "authors/status"]
  verbs: ["get", "update", "patch"]

---
# ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: manager-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: manager-role
subjects:
- kind: ServiceAccount
  name: controller-manager
  namespace: system
```

**¿Por qué ClusterRole y no Role?**
- ClusterRole: Permisos en todo el cluster
- Role: Permisos en un namespace específico
- Usamos ClusterRole porque Books/Authors pueden estar en cualquier namespace

### 6. Watch

**¿Qué es?**
Mecanismo de Kubernetes para observar cambios en recursos en tiempo real.

**¿Cómo funciona?**
```
Controller                    API Server
    │                              │
    │  Watch books                 │
    ├─────────────────────────────>│
    │                              │
    │  HTTP connection abierta     │
    │<─────────────────────────────┤
    │                              │
    │  (esperando eventos...)      │
    │                              │
Usuario crea Book                 │
    │                              │
    │  Event: ADDED book-1984      │
    │<─────────────────────────────┤
    │                              │
    │  Reconcile(book-1984)        │
    │                              │
Usuario actualiza Book            │
    │                              │
    │  Event: MODIFIED book-1984   │
    │<─────────────────────────────┤
    │                              │
    │  Reconcile(book-1984)        │
```

**Tipos de eventos:**
- `ADDED`: Recurso creado
- `MODIFIED`: Recurso actualizado
- `DELETED`: Recurso eliminado

**Ventajas:**
- Tiempo real: No hay polling
- Eficiente: Solo notifica cambios
- Escalable: Múltiples controladores pueden watch

### 7. Leader Election

**¿Qué es?**
Cuando hay múltiples replicas del controlador, solo una reconcilia activamente.

**¿Por qué?**
- Evitar conflictos: Dos controladores actualizando el mismo recurso
- Eficiencia: No duplicar trabajo
- Alta disponibilidad: Si el leader falla, otro toma su lugar

**¿Cómo funciona?**
```
Pod 1 (Leader)          Pod 2 (Standby)         Pod 3 (Standby)
    │                        │                        │
    │  Adquiere lease        │                        │
    │  en ConfigMap          │                        │
    ├───────────────────────>│                        │
    │                        │                        │
    │  Reconcilia recursos   │  Espera                │  Espera
    │  ↓                     │  ↓                     │  ↓
    │  Watch events          │  Watch lease           │  Watch lease
    │                        │                        │
    │  (Pod 1 falla)         │                        │
    X                        │                        │
                             │                        │
                             │  Detecta lease         │
                             │  expirado              │
                             ├───────────────────────>│
                             │                        │
                             │  Adquiere lease        │
                             │  Ahora es leader       │
                             │                        │
                             │  Reconcilia recursos   │
```

### 8. Finalizers

**¿Qué son?**
Mecanismo para ejecutar cleanup antes de eliminar un recurso.

**¿Cómo funcionan?**
```go
// Agregar finalizer
book.Finalizers = append(book.Finalizers, "bookstore.io/cleanup")

// En Reconcile, verificar si está siendo eliminado
if !book.DeletionTimestamp.IsZero() {
    // Recurso está siendo eliminado
    if containsString(book.Finalizers, "bookstore.io/cleanup") {
        // Ejecutar cleanup
        if err := r.cleanupExternalResources(book); err != nil {
            return ctrl.Result{}, err
        }
        
        // Remover finalizer
        book.Finalizers = removeString(book.Finalizers, "bookstore.io/cleanup")
        if err := r.Update(ctx, &book); err != nil {
            return ctrl.Result{}, err
        }
    }
    return ctrl.Result{}, nil
}
```

**Casos de uso:**
- Eliminar recursos externos (S3 buckets, DNS records)
- Notificar a otros sistemas
- Cleanup de recursos relacionados

---

## Próximos Pasos

### Fase 3: Recursos Avanzados
- Crear **BookStore** y **Inventory**
- Implementar relaciones many-to-many
- Usar finalizers para cleanup
- Owner references para garbage collection

### Fase 4: Webhooks
- Webhook de validación para Books
- Webhook de mutación (defaults automáticos)
- Validaciones complejas entre campos
- Conversion webhooks para versionado

### Fase 5: Features Avanzadas
- **BookReservation** con máquina de estados
- RBAC granular por usuario
- Subresources personalizados (scale)
- Métricas personalizadas con Prometheus
- Eventos personalizados

---

## Recursos Adicionales

- [Kubebuilder Book](https://book.kubebuilder.io/) - Documentación oficial
- [Kubernetes API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md) - Mejores prácticas
- [Controller Runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime) - Documentación del framework
- [Kind Documentation](https://kind.sigs.k8s.io/) - Guía de Kind
- [Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) - Patrón de operadores

---

**Versión:** Fase 2 Completada  
**Última actualización:** Enero 2026  
**Autor:** Oscar Llamas

