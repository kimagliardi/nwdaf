# free5gc-nwdaf Helm chart

This is a Helm chart for deploying the [free5GC](https://github.com/free5gc/free5gc) NWDAF (Network Data Analytics Function) on Kubernetes.

This chart is designed to be included in the [dependencies](https://github.com/free5gc/free5gc-helm/tree/main/charts/free5gc/charts) of the [main free5gc chart](https://github.com/free5gc/free5gc-helm/tree/main/charts/free5gc). Furthermore, it can be installed separately on a Kubernetes cluster at the same network with other Free5GC NFs.

## Prerequisites
 - A Kubernetes cluster ready to use. You can use [kubeadm](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/create-cluster-kubeadm/) to create it.
 - [Helm3](https://helm.sh/docs/intro/install/).
 - [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) (optional).
 - A deployed NRF that this NWDAF can register with.

## Quickstart guide

### Install the NWDAF
Run the following commands on a host that can communicate with the API server of your cluster.
```console
kubectl create ns <namespace>
helm -n <namespace> install <release-name> ./free5gc-nwdaf/
```

### Check the state of the created pod
```console
kubectl -n <namespace> get pods -l "nf=nwdaf"
```

### Uninstall the NWDAF
```console
helm -n <namespace> delete <release-name>
```
Or...
```console
helm -n <namespace> uninstall <release-name>
```

## Customized installation
This chart allows you to customize its installation. The table below shows the parameters that can be modified before installing the chart or when upgrading it as well as their default values.

### Global parameters

| Parameter | Description | Default value |
| --- | --- | --- |
| `global.projectName` | The name of the project. | `free5gc` |
| `global.sbi.scheme` | The SBI scheme for all control plane NFs. Possible values are `http` and `https` | `http` |
| `global.nrf.service.name` | The name of the service used to access the NRF SBI interface. | `nrf-nnrf` |
| `global.nrf.service.port` | The NRF SBI port number. | `8000` |
| `global.nrf.service.type` | The type of the NRF SBI service. | `ClusterIP` |

### Common parameters
| Parameter | Description | Default value |
| --- | --- | --- |
| `initcontainers.curl.registry` | The Docker image registry of the Init Container waiting for the NRF to be ready. | `towards5gs` |
| `initcontainers.curl.image` | The Docker image name of the Init Container waiting for the NRF to be ready. | `initcurl` |
| `initcontainers.curl.tag` | The Docker image tag of the Init Container waiting for the NRF to be ready. | `"1.0.0"` |

### NWDAF parameters

| Parameter | Description | Default value |
| --- | --- | --- |
| `nwdaf.name` | The Network Function name of NWDAF. | `nwdaf` |
| `nwdaf.replicaCount` | The number of NWDAF replicas. | `1` |
| `nwdaf.image.name` | The NWDAF Docker image name. | `towards5gs/free5gc-nwdaf` |
| `nwdaf.image.tag` | The NWDAF Docker image tag. | `defaults to chart AppVersion` |
| `nwdaf.service.name` | The name of the service used to expose the NWDAF SBI interface. | `nwdaf-nnwdaf` |
| `nwdaf.service.type` | The type of the NWDAF SBI service. | `ClusterIP` |
| `nwdaf.service.port` | The NWDAF SBI port number. | `80` |
| `nwdaf.service.targetPort` | The NWDAF container port number. | `8000` |
| `nwdaf.volume.mount` | The path to the folder where configuration files should be mounted. | `/free5gc/config/`|
| `nwdaf.podAnnotations` | Pod annotations. | `{}`|
| `nwdaf.imagePullSecrets` | Image pull secrets. | `[]`|
| `nwdaf.podSecurityContext` | Pod security context. | `{}`|
| `nwdaf.securityContext` | Security context. | `{}`|
| `nwdaf.resources` | CPU and memory requests and limits. | `see values.yaml`|
| `nwdaf.readinessProbe` | Readiness probe. | `see values.yaml`|
| `nwdaf.livenessProbe` | Liveness probe. | `see values.yaml`|
| `nwdaf.nodeSelector` | Node selector. | `{}`|
| `nwdaf.tolerations` | Tolerations. | `[]`|
| `nwdaf.affinity` | Affinity. | `{}`|
| `nwdaf.autoscaling` | HPA parameters (disabled by default). | `see values.yaml`|
| `nwdaf.ingress` | Ingress parameters (disabled by default). | `see values.yaml`|
| `nwdaf.metrics.enabled` | Enable Prometheus metrics endpoint. | `false`|
| `nwdaf.configuration.logger.level` | Logger level. | `info`|

## Reference
 - https://github.com/free5gc/free5gc
 - https://github.com/free5gc/free5gc-helm
