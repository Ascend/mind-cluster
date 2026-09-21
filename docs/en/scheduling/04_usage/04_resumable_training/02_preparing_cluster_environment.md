# Preparing the Cluster Environment<a name="ZH-CN_TOPIC_0000002479386542"></a>

<!-- md-trans-meta sourceCommit=6f7a6aaf48adf808f43bcc175f4a532891d986e4 translatedAt=2026-08-31T03:55:46.365Z pushedAt=2026-08-31T05:47:15.254Z -->

The resumable training feature is an advanced feature built on the MindCluster cluster scheduling component. It implements training fault recovery by leveraging the full Ascend software and hardware stack. Before using the resumable training feature, the following prerequisites must be met.

- Complete K8s cluster basic performance tuning. For details, see [Kubernetes Cluster Basics Performance Tuning](../../07_references/05_appendix.md#kubernetes-cluster-basic-performance-tuning).

- Have a shared storage system

    Some processes of the resumable training feature depend on reading storage data, such as loading checkpoint, starting training, and loading compilation caches. Storage performance affects the overall recovery time of resumable training. To avoid degradation of the training recovery time, it is recommended that you optimize the storage performance configuration. The recommended configuration provided below uses a 10,000-card cluster as an example.
    - 8k IO read IOPS: \>10.24 million
    - 8k IO write IOPS: \>1.28 million
    - Large file sequential read bandwidth: \>288GB/s
    - Large file create write bandwidth: \>173GB/s

- (Optional) Expand shared memory
    To use the resumable training feature, it is recommended to expand the memory. Add the parameters as indicated in the comments. An example is as follows.

    ```yaml
    ...
            volumeMounts:                             # Resumable training scale-out
             - name: shm
               mountPath: /dev/shm
            volumes:
            - name: shm
              emptyDir:
                medium: Memory
                sizeLimit: 16Gi
    ...
    ```

- (Optional) Configure CPU and memory resources
    If you need to configure CPU and memory resources, manually add the `cpu` and `memory` parameters and their corresponding values as shown in the following example. Configure the specific values based on the actual situation.

    ```yaml
    ...
              resources:
                requests:
                  huawei.com/Ascend910: 8
                  cpu: 1000m
                  memory: 100Gi
                limits:
                  huawei.com/Ascend910: 8
                  cpu: 1000m
                  memory: 100Gi
    ...
    ```
