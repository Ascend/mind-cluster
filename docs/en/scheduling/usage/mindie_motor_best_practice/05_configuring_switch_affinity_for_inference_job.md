# Configuring Inference Job Switch Affinity

Currently, switch affinity can only be configured on Atlas 800I A2 inference servers. Enabling this feature helps avoid downstream traffic conflicts on Spine switches. For details about how this feature works, see [Switch Affinity Scheduling 1.0](../basic_scheduling/01_affinity_scheduling/04_node_based_affinity.md#switch-affinity-scheduling-10).

## Prerequisites

You have completed steps in [(Optional) Using Volcano Switch Affinity Scheduling](../../developer_guide/installation_deployment/manual_installation/05_volcano.md#optional-using-volcano-switch-affinity-scheduling).

## Procedure

Set the switch affinity `tor-affinity` to `normal-schema`. The following is a YAML example:

```Yaml
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: mindie-server-0
  namespace: mindie
  labels:
    framework: pytorch
    app: mindie-ms-server        # Indicates the role of MindIE Motor in Ascend Job. Do not modify it.
    jobID: mindie-ms-test        # The unique identifier of the current MindIE Motor inference job in the cluster. Configure it based on the actual situation.
    tor-affinity: normal-schema    # Enable switch affinity
    ring-controller.atlas: ascend-910b
```
