# Kubefleet ClusterResourcePlacement Support

Draft now supports generating Kubefleet ClusterResourcePlacement manifests through the `kubefleet-clusterresourceplacement` addon template.

## Prerequisites

1. Have an existing Draft project with deployment files (run `draft create` first)
2. Have the Draft CLI installed and built

## Usage

The ClusterResourcePlacement addon supports both PickAll and PickFixed placement types as described in the [Kubefleet documentation](https://fleet.azure.com/).

### PickAll Placement Type

For distributing resources to all matching clusters:

```bash
draft update --addon kubefleet-clusterresourceplacement \
  --variable CRP_NAME=demo-crp \
  --variable RESOURCE_SELECTOR_NAME=fmad-demo \
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
      name: fmad-demo
      version: v1
  policy:
    placementType: PickAll
```

### PickFixed Placement Type

For distributing resources to specific clusters:

```bash
draft update --addon kubefleet-clusterresourceplacement \
  --variable CRP_NAME=fmad-demo-crp \
  --variable RESOURCE_SELECTOR_NAME=fmad-demo \
  --variable PLACEMENT_TYPE=PickFixed \
  --variable CLUSTER_NAME_1=cluster-name-01 \
  --variable CLUSTER_NAME_2=cluster-name-02 \
  --variable PARTOF=my-project
```

This generates:

```yaml
apiVersion: placement.kubernetes-fleet.io/v1
kind: ClusterResourcePlacement
metadata:
  name: fmad-demo-crp
  labels:
    app.kubernetes.io/name: fmad-demo-crp
    app.kubernetes.io/part-of: my-project
    kubernetes.azure.com/generator: draft
spec:
  resourceSelectors:
    - group: ""
      kind: Namespace
      name: fmad-demo
      version: v1
  policy:
    placementType: PickFixed
    clusterNames:
       - cluster-name-01
       - cluster-name-02
```

## Template Variables

| Variable | Type | Description | Required | Default |
|----------|------|-------------|----------|---------|
| `CRP_NAME` | string | Name of the ClusterResourcePlacement | Yes | - |
| `RESOURCE_SELECTOR_NAME` | string | Name of the resource to select for placement | Yes | - |
| `PLACEMENT_TYPE` | string | Placement policy type (PickAll or PickFixed) | No | "PickAll" |
| `CLUSTER_NAME_1` | string | First cluster name (for PickFixed only) | No | "" |
| `CLUSTER_NAME_2` | string | Second cluster name (for PickFixed only) | No | "" |
| `PARTOF` | string | Label to identify which project the resource belongs to | Yes | - |
| `GENERATORLABEL` | string | Label to identify who generated the resource | No | "draft" |

## Interactive Mode

You can also run the command interactively without specifying all variables:

```bash
draft update --addon kubefleet-clusterresourceplacement
```

Draft will prompt you for the required values.

## Non-interactive Mode

For automation and CI/CD pipelines, use `--interactive=false` and provide all required variables:

```bash
draft update --addon kubefleet-clusterresourceplacement \
  --interactive=false \
  --variable CRP_NAME=my-crp \
  --variable RESOURCE_SELECTOR_NAME=my-namespace \
  --variable PLACEMENT_TYPE=PickAll \
  --variable PARTOF=my-project
```

## Output

The generated ClusterResourcePlacement manifest will be created at `manifests/clusterresourceplacement.yaml` in your project directory.