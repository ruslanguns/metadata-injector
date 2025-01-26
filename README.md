# Metadata Injector Operator

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

### Installation

1. **Install the CRDs:**
```sh
make install
```

2. **Deploy the operator:**
```sh
make deploy IMG=<your-registry>/metadata-injector-operator:tag
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

### Configuration Options

- **Reconciliation Interval**: Set using the `metadata-injector.ruso.dev/reconcile-interval` annotation
- **Auto Reconciliation**: Control using the `metadata-injector.ruso.dev/disable-auto-reconcile` annotation
- **Resource Selection**: Configure using spec.selectors to target specific resources
- **Metadata Injection**: Define labels and annotations to inject in spec.inject

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
