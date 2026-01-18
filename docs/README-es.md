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

