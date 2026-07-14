# Preparing Kubernetes and Shared Storage<a name="ZH-CN_TOPIC_0000002479386542"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T07:59:53.340Z pushedAt=2026-06-09T09:02:55.483Z -->

The resumable training feature is an advanced feature for MindCluster scheduling components. It leverages the full Ascend software and hardware stack to achieve training fault recovery. Before using this feature, ensure the following prerequisites must be met.

- Complete the basic performance tuning for the Kubernetes cluster. For details, see [Kubernetes Cluster Basics Performance Tuning](../../appendix.md#kubernetes-cluster-basic-performance-tuning).

- Have a shared storage system

    Some processes of resumable training depend on reading storage data, such as loading checkpoint, starting training, and loading compilation caches. Storage performance affects the overall recovery time of resumable training. To prevent degradation of training recovery time, it is recommended to optimize storage performance configuration. The recommended configuration provided below uses a cluster of 10,000-card scale as an example.

    - 8K IO read IOPS: > 10.24 million
    - 8K IO write IOPS:> 1.28 million
    - Large file sequential read bandwidth: > 288 GB/s
    - Write bandwidth for large file creation: > 173 GB/s
