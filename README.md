# Metadata Injector Operator

[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/metadata-injector-operator)](https://artifacthub.io/packages/search?repo=metadata-injector-operator)

A Kubernetes operator that automatically injects metadata (labels and annotations) into specified resources across your cluster.

## Description

The Metadata Injector Operator provides a way to automatically manage and inject metadata into Kubernetes resources. It allows you to:

- Define target resources by kind, group, version, and names
- Specify namespaces to include or exclude
- Inject custom labels and annotations
- Configure automatic reconciliation intervals
- Enable/disable automatic reconciliation

This is particularly useful for:
- Enforcing consistent metadata across resources
- Implementing governance policies
- Automating resource tagging
- Managing resource classifications

## Getting Started

### Prerequisites
- go version v1.22.0+
- docker version 17.03+
- kubectl version v1.11.3+
- Access to a Kubernetes v1.11.3+ cluster

### Installation Methods

#### Method 1: Using Make Commands

1. **Install the CRDs:**
```sh
make install
```

2. **Deploy the operator:**
```sh
make deploy IMG=<your-registry>/metadata-injector-operator:tag
```

#### Method 2: Using Helm

1. **Add the Helm repository:**
```sh
helm repo add metadata-injector https://ruslanguns.github.io/metadata-injector-operator
helm repo update
```

2. **Install the chart:**
```sh
helm install metadata-injector metadata-injector/metadata-injector-operator \
  --namespace metadata-injector-system \
  --create-namespace
```

### Usage

1. Create a MetadataInjector resource:

```yaml
apiVersion: core.k8s.ruso.dev/v1alpha1
kind: MetadataInjector
metadata:
  name: example-injector
  annotations:
    metadata-injector.ruso.dev/reconcile-interval: "5m"  # Optional: Custom reconciliation interval
    metadata-injector.ruso.dev/disable-auto-reconcile: "false"  # Optional: Disable automatic reconciliation
spec:
  selectors:
    - kind: Secret
      group: ""
      version: "v1"
      namespaces:
        - default
        - kube-system
      names:
        - secret-name-1
        - secret-name-2
  inject:
    labels:
      environment: production
      team: platform
    annotations:
      description: "Managed by metadata-injector"
```

2. Apply the configuration:
```sh
kubectl apply -f config/samples/
```

### Configuration

#### Operator Configuration

The following configuration options are available:

- **Reconciliation Interval**: Set using the `metadata-injector.ruso.dev/reconcile-interval` annotation
- **Auto Reconciliation**: Control using the `metadata-injector.ruso.dev/disable-auto-reconcile` annotation
- **Resource Selection**: Configure using spec.selectors to target specific resources
- **Metadata Injection**: Define labels and annotations to inject in spec.inject

#### Helm Chart Configuration

The following values can be customized in your Helm chart installation:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `nameOverride` | Override the name of the chart | `""` |
| `fullnameOverride` | Override the full name of the chart | `""` |
| `namespace.create` | Create the namespace | `true` |
| `namespace.name` | Namespace name | `"metadata-injector-system"` |
| `image.repository` | Operator image repository | `ruslanguns/metadata-injector-operator` |
| `image.tag` | Operator image tag | `v0.0.1` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.name` | Service account name | `""` |
| `serviceAccount.annotations` | Service account annotations | `{}` |
| `rbac.create` | Create RBAC resources | `true` |
| `resources.limits.cpu` | CPU resource limits | `500m` |
| `resources.limits.memory` | Memory resource limits | `128Mi` |
| `resources.requests.cpu` | CPU resource requests | `10m` |
| `resources.requests.memory` | Memory resource requests | `64Mi` |
| `metrics.enabled` | Enable metrics | `true` |
| `metrics.port` | Metrics port | `8443` |
| `metrics.service.type` | Metrics service type | `ClusterIP` |
| `probe.port` | Health probe port | `8081` |
| `probe.liveness.initialDelaySeconds` | Liveness probe initial delay | `15` |
| `probe.liveness.periodSeconds` | Liveness probe period | `20` |
| `probe.readiness.initialDelaySeconds` | Readiness probe initial delay | `5` |
| `probe.readiness.periodSeconds` | Readiness probe period | `10` |
| `podSecurityContext.runAsNonRoot` | Run as non-root | `true` |
| `replicaCount` | Number of operator replicas | `1` |
| `crds.create` | Create CRDs | `true` |
| `podAnnotations` | Additional pod annotations | `{}` |
| `nodeSelector` | Node selector configuration | `{}` |
| `tolerations` | Pod tolerations | `[]` |
| `affinity` | Pod affinity rules | `{}` |

Example custom values file:
```yaml
# custom-values.yaml
namespace:
  name: "custom-namespace"

image:
  repository: custom-registry/metadata-injector-operator
  tag: v1.0.0

resources:
  limits:
    cpu: "1"
    memory: "256Mi"
  requests:
    cpu: "100m"
    memory: "128Mi"

metrics:
  enabled: true
  service:
    annotations:
      prometheus.io/scrape: "true"
      prometheus.io/port: "8443"

replicaCount: 2
```

Apply custom values:
```sh
helm install metadata-injector metadata-injector/metadata-injector-operator \
  -f custom-values.yaml \
  --namespace metadata-injector-system \
  --create-namespace
```

### Monitoring

Monitor the status of your MetadataInjector:

```sh
kubectl get metadatainjectors
```

This shows:
- Age of the injector
- Last successful execution
- Next scheduled run
- Current reconciliation interval

### Uninstallation

#### Method 1: Using Make Commands

1. **Remove MetadataInjector resources:**
```sh
kubectl delete -k config/samples/
```

2. **Remove the operator:**
```sh
make undeploy
```

3. **Remove CRDs:**
```sh
make uninstall
```

#### Method 2: Using Helm

```sh
helm uninstall metadata-injector -n metadata-injector-system
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

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