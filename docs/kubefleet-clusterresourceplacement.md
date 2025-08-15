# KubeFleet ClusterResourcePlacement Support

Draft now supports generating KubeFleet ClusterResourcePlacement manifests through the `kubefleet-clusterresourceplacement` addon template.

## Prerequisites

1. Have an existing Draft project with deployment files (run `draft create` first)
2. Have the Draft CLI installed and built

## Usage

The ClusterResourcePlacement addon supports both PickAll and PickFixed placement types as described in the [KubeFleet documentation](https://kubefleet.dev/docs/concepts/crp/).

### PickAll Placement Type

For distributing resources to all matching clusters:

```bash
draft distribute \
  --variable CRP_NAME=demo-crp \
  --variable RESOURCE_SELECTOR_NAME=ns-demo \
  --variable PLACEMENT_TYPE=PickAll \
  --variable PARTOF=my-project
```

This generates:

```yaml
apiVersion: placement.kubernetes-fleet.io/v1
kind: ClusterResourcePlacement
metadata:
  name: demo-crp
  labels:
    app.kubernetes.io/name: demo-crp
    app.kubernetes.io/part-of: my-project
    kubernetes.azure.com/generator: draft
spec:
  resourceSelectors:
    - group: ""
      kind: Namespace
      name: ns-demo
      version: v1
  policy:
    placementType: PickAll
```

### PickFixed Placement Type

For distributing resources to specific clusters:

```bash
draft distribute \
  --variable CRP_NAME=ns-demo-crp \
  --variable RESOURCE_SELECTOR_NAME=fmad-demo \
  --variable PLACEMENT_TYPE=PickFixed \
  --variable CLUSTER_NAMES=cluster-name-01,cluster-name-02 \
  --variable PARTOF=my-project
```

This generates:

```yaml
apiVersion: placement.kubernetes-fleet.io/v1
kind: ClusterResourcePlacement
metadata:
  name: ns-demo-crp
  labels:
    app.kubernetes.io/name: ns-demo-crp
    app.kubernetes.io/part-of: my-project
    kubernetes.azure.com/generator: draft
spec:
  resourceSelectors:
    - group: ""
      kind: Namespace
      name: ns-demo
      version: v1
  policy:
    placementType: PickFixed
    clusterNames:
       - cluster-name-01
       - cluster-name-02
```

#### Example with Three Clusters

```bash
draft distribute \
  --variable CRP_NAME=multi-cluster-demo \
  --variable RESOURCE_SELECTOR_NAME=demo-namespace \
  --variable PLACEMENT_TYPE=PickFixed \
  --variable CLUSTER_NAMES=cluster-east,cluster-west,cluster-central \
  --variable PARTOF=my-project
```

This generates:

```yaml
apiVersion: placement.kubernetes-fleet.io/v1
kind: ClusterResourcePlacement
metadata:
  name: multi-cluster-demo
  labels:
    app.kubernetes.io/name: multi-cluster-demo
    app.kubernetes.io/part-of: my-project
    kubernetes.azure.com/generator: draft
spec:
  resourceSelectors:
    - group: ""
      kind: Namespace
      name: demo-namespace
      version: v1
  policy:
    placementType: PickFixed
    clusterNames:
       - cluster-east
       - cluster-west
       - cluster-central
```

## Template Variables

| Variable | Type | Description | Required | Default |
|----------|------|-------------|----------|---------|
| `CRP_NAME` | string | Name of the ClusterResourcePlacement | Yes | - |
| `RESOURCE_SELECTOR_NAME` | string | Name of the resource to select for placement | Yes | - |
| `PLACEMENT_TYPE` | string | Placement policy type (PickAll or PickFixed) | No | "PickAll" |
| `CLUSTER_NAMES` | string | Comma-separated list of cluster names (for PickFixed only) | No | "" |
| `PARTOF` | string | Label to identify which project the resource belongs to | Yes | - |
| `GENERATORLABEL` | string | Label to identify who generated the resource | No | "draft" |

Draft will prompt you for the required values.

## Output

The generated ClusterResourcePlacement manifest will be created at `manifests/clusterresourceplacement.yaml` in your project directory.
