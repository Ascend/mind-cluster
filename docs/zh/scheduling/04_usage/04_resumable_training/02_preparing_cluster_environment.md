# 准备集群环境<a name="ZH-CN_TOPIC_0000002479386542"></a>

断点续训特性是基于MindCluster集群调度组件的高阶特性，结合昇腾软硬件全栈实现训练故障恢复，使用断点续训特性前需要满足以下前置条件。

- 完成K8s集群基础性能调优，详情请参见[K8s集群基础性能调优](../../07_references/05_appendix.md#k8s集群基础性能调优)。

- 具备共享存储系统

    断点续训特性的部分流程依赖读取存储数据，如加载CKPT、启动训练和编译缓存加载等，存储性能会影响断点续训整体恢复时间。为避免训练恢复时间劣化，建议进行存储性能配置优化，以下提供的推荐配置以万卡规模集群为例。
    - 8k IO读IOPS：\>1024W
    - 8k IO写IOPS：\>128W
    - 大文件顺序读带宽：\>288GB/s
    - 大文件创建写带宽：\>173GB/s

- （可选）扩展共享内存
    使用断点续训功能，建议扩展内存，请按注释添加参数，示例如下。

    ```yaml
    ...
            volumeMounts:                             # 断点续训扩容
             - name: shm
               mountPath: /dev/shm
            volumes:
            - name: shm
              emptyDir:
                medium: Memory
                sizeLimit: 16Gi
    ...
    ```

- （可选）配置CPU和内存资源
     若需要配置CPU、Memory资源，请参见如下示例手动添加“cpu”和“memory”参数和对应的参数值，具体数值请根据实际情况配置。

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
