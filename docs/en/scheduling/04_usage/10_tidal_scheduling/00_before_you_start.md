# Before You Start<a name="ZH-CN_TOPIC_000000_task_alternation_overview"></a>

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-27T10:56:14.399Z pushedAt=2026-08-27T10:56:40.863Z -->

## Overview<a name="section_overview_alternation"></a>

In the scenario where large model inference and training share an NPU cluster, inference jobs scale in and out dynamically based on business traffic, while training jobs occupy NPU resources over a long period. When more NPUs are needed during inference peak hours, resources can be reclaimed from training jobs through Volcano's Preempt or Reclaim mechanism. After resources are released during inference off-peak hours, training jobs resume running.

This best practice describes how to configure tidal scheduling for inference and training jobs, and leverages the **prefer returning to original node for rebuilt Pods** feature. When Pods of both job types are rebuilt, they preferentially return to the nodes where they previously ran, reusing the container images already cached on those nodes and reducing the startup time from 5 to 20 minutes down to seconds.

**Core Benefits**

| Scenario | Without Return-to-Original-Node Preference | Enable Return-to-Original-Node Preference |
|------|-----------------------|----------------|
| Training Pod recovery after being evicted by Preempt/Reclaim | Randomly assigned node, re-pulling 20 to 40 GB images | Preferentially returns to original node, reusing image cache |
| Inference Pod scale-out after scale-in | Randomly assigned node, re-pulling 15 to 30 GB images | Preferentially returns to original node, reusing image cache |
| Startup latency | Minute-level | Second-level |
| Image pull bandwidth | Re-pulled on every rebuild | Pulled only on first build |

## Prerequisites<a name="section_prerequisites_alternation"></a>

- Volcano, Ascend Device Plugin, Ascend Docker Runtime, Ascend Operator, ClusterD, and NodeD have been installed and deployed. For details, see [Installation and Deployment](../../03_installation_guide/02_installation/00_helm_installation.md).
- ascend-volcano-plugin version ≥ 26.1.0 (including the return-to-original-node preference feature).
- Image cache has been configured on the NPU cluster nodes.

## Differences Between Preempt and Reclaim<a name="section_preempt_vs_reclaim"></a>

| Feature | Preempt | Reclaim |
|------|---------------|----------------|
| Trigger condition | A high-priority job cannot find sufficient resources | A high-weight queue has insufficient resources, so resources are reclaimed from low-weight queues |
| Resource release granularity | Job level (by `PriorityClass`) | Queue level (by `Queue weight`) |
| Configuration method | `PriorityClass` + `gang.enablePreemptable` | `Queue.weight` + `Queue.reclaimable` |
| Applicable scenario | Inference SLO is strict, and training can tolerate interruption | Resource priorities are managed by service line |

For more details about Preempt and Reclaim, see [Volcano Documentation — Actions](https://volcano.sh/docs/Scheduler/Actions).

## Training Job Restart Mechanism<a name="section_restart"></a>

A training job configures `minAvailable` equal to `replicas` to ensure gang integrity. When a Pod is evicted by Preempt or Reclaim, the number of Pods falls below `minAvailable`, and the `fault-scheduling: grace` label triggers the rescheduling module to automatically cascade-clean the remaining Pods and restart the job. After the job restarts, the Pods re-enter scheduling and preferentially return to the original node through the return-to-original-node feature.

## Supported Product Forms<a name="section_products_alternation"></a>

- Atlas 800 training server
- Atlas 800I A2 inference server
- Atlas 900 A3 SuperPoD
- Atlas 9000 A3 SuperPoD cluster computing system
- <term>Atlas inference products</term>
- A200I A2 Box heterogeneous subrack
- Atlas 350 accelerator card
