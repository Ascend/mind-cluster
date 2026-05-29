# 配置任务YAML<a name="ZH-CN_TOPIC_0000002479226518"></a>


## 任务YAML配置示例<a name="ZH-CN_TOPIC_0000002511346461"></a>

如果是acjob任务，在配置YAML前，请先了解相关YAML参数说明，详细说明如[acjob任务yaml参数说明](../../api/yaml_configuration.md#acjob)所示。

重调度模式和优雅容错模式可参见如下[操作步骤](#zh-cn_topic_0000002202737289_zh-cn_topic_0000001951258657_section18181655154219)配置示例。当**subHealthyStrategy**取值为graceExit时，需要参见[配置亚健康主动CKPT保存](./05_configuring_training_recovery.md#zh-cn_topic_0000002202737289_zh-cn_topic_0000001951258657_section1048332432310)完成启动脚本与任务YAML的适配，以确保任务因亚健康故障被重调度前能够正常保存CKPT文件。

**前提条件<a name="zh-cn_topic_0000002202737289_zh-cn_topic_0000001951258657_section7585519135117"></a>**

用户已创建[hccl.json](../../api/hccl.json_file_description.md)文件的具体挂载路径，详细操作步骤请参见[Ascend Operator](../../developer_guide/installation_deployment/manual_installation/08_ascend_operator.md)中的“步骤4”。

**操作步骤<a name="zh-cn_topic_0000002202737289_zh-cn_topic_0000001951258657_section18181655154219"></a>**

1. 将YAML文件上传至管理节点任意目录，并根据实际情况修改文件内容。
    - 以a800\_AscendJob\_<i>\{xxx\}</i>b.yaml为例，在一台Atlas 200T A2 Box16 异构子框节点创建**分布式训练**任务，任务使用2\*4个芯片，修改示例如下。

        ```Yaml
        apiVersion: mindxdl.gitee.com/v1
        kind: AscendJob
        metadata:
          name: default-test-mindspore
          labels:
            framework: mindspore  # 训练框架名称
            fault-scheduling: "grace"     # 开启优雅删除模式
            ring-controller.atlas: ascend-{xxx}b
            fault-retry-times: "3"            # 开启业务面故障无条件重试能力，同时需要将restartPolicy取值设置为Never
            tor-affinity: "normal-schema" #该标签为任务是否使用交换机亲和性调度标签，null或者不写该标签则不使用该特性。large-model-schema表示大模型任务或填充任务，normal-schema表示普通任务
            pod-rescheduling: "on"     # 开启Pod级别重调度
            subHealthyStrategy: "ignore"  # 忽略健康状态为亚健康的节点，后续任务在亲和性调度上不优先调度该节点
          annotations:
            huawei.com/schedule_policy: "chip8-node16" # 调度策略
        spec:
          schedulerName: volcano    # 当Ascend Operator组件的启动参数enableGangScheduling为true时生效
          runPolicy:
            backoffLimit: 3      # 任务重调度次数
            schedulingPolicy:
              minAvailable: 3       # 任务总副本数
              queue: default     # 任务所属队列
          successPolicy: AllWorkers  # 任务成功的前提
          replicaSpecs:
            Scheduler:
              replicas: 1            #只能为1
              restartPolicy:  Never   #容器重启策略
              template:
                metadata:
                  labels:
                    ring-controller.atlas: ascend-{xxx}b  # 标识产品类型
                spec:
                  terminationGracePeriodSeconds: 360  #容器收到SIGTERM到被K8s强制停止经历的时间
                  nodeSelector:
                    example-key: example-value    # 示例值，用户可根据调度意图自行配置nodeSelector
                  containers:
                  - name: ascend     # 不能修改
        ...
                    ports:                     # 可选，分布式训练集合通信端口
                      - containerPort: 2222
                        name: ascendjob-port
                    volumeMounts:
        ...

            Worker:
              replicas: 2
              restartPolicy: Never  # 容器重启策略
              template:
                metadata:
                  labels:
                    ring-controller.atlas: ascend-{xxx}b  # 标识产品类型
                spec:
                  terminationGracePeriodSeconds: 360   #容器收到SIGTERM到被K8s强制停止经历的时间
                  affinity:
        ...
                  nodeSelector:
                    example-key: example-value    # 示例值，用户可根据调度意图自行配置nodeSelector
                  containers:
                  - name: ascend      # 不能修改
        ...
                    env:
                    - name: ASCEND_VISIBLE_DEVICES
                      valueFrom:
                        fieldRef:
                          fieldPath: metadata.annotations['huawei.com/Ascend910']         # 需要和下面resources和requests保持一致
        ...

                    ports:        # 可选，分布式训练集合通信端口
                      - containerPort: 2222
                        name: ascendjob-port
                    resources:
                      limits:
                        huawei.com/Ascend910: 4      # 需要的NPU芯片个数为4
                      requests:
                        huawei.com/Ascend910: 4       # 与limits取值一致
        ```

    - 以a800\_vcjob.yaml为例，在一台Atlas 800 训练服务器节点创建**单机训练**任务，任务使用8个芯片，修改示例如下。

        ```Yaml
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: rings-config-mindx-dls-test     # rings-config-后的名字需要与任务名一致
        ...
          labels:
            ring-controller.atlas: ascend-910  # 标识产品类型
        ...
        ---
        apiVersion: batch.volcano.sh/v1alpha1   # 不可修改。必须使用Volcano的API。
        kind: Job                               # 目前只支持Job类型
        metadata:
          name: mindx-dls-test                  # 任务名，可自定义
          labels:
            ring-controller.atlas: ascend-910
            fault-scheduling: "grace"        # 开启优雅删除模式
            fault-retry-times: "3"            # 开启业务面故障无条件重试能力，同时需要将restartPolicy取值设置为Never；并将policies的event设置为PodFailed，action设置为Ignore
            tor-affinity: "normal-schema" #该标签为任务是否使用交换机亲和性调度标签，null或者不写该标签则不使用该特性。large-model-schema表示大模型任务或填充任务，normal-schema表示普通任务
            pod-rescheduling: "on"     # 开启Pod级别重调度
            subHealthyStrategy: "ignore"     # 忽略健康状态为亚健康的节点，后续任务在亲和性调度上不优先调度该节点
          annotations:
            huawei.com/schedule_policy: "chip4-node8"
        ...
        spec:
          policies:  # 使用重调度功能时，无需修改 policies 内容
            - event: PodFailed
              action: Ignore
        ...
          minAvailable: 1                  # 单机为1
        ...
          maxRetry: 3              # 重调度次数
        ...
          - name: "default-test"
              replicas: 1                  # 单机为1
              template:
                metadata:
        ...
                spec:
                  terminationGracePeriodSeconds: 360  #容器收到SIGTERM到被K8s强制停止经历的时间
        ...
                    env:
        ...
                  - name: ASCEND_VISIBLE_DEVICES                       # Ascend Docker Runtime使用该字段
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.annotations['huawei.com/Ascend910']               # 需要和下面resources和requests保持一致
        ...
                    resources:
                      requests:
                        huawei.com/Ascend910: 8          # 需要的NPU芯片个数为8。可在下方添加行，配置memory、cpu等资源
                      limits:
                        huawei.com/Ascend910: 8          # 目前需要和上面requests保持一致
        ...
                    nodeSelector:
                      example-key: example-value    # 可选值，用户可根据实际需求配置nodeSelector
        ...
                restartPolicy: Never   # 容器重启策略
        ```

2. 配置MindIO的通信地址。在代码中新增以下内容。

    ```Yaml
    ...
       Master:
    ...
                env:
                  - name: POD_IP
                    valueFrom:
                      fieldRef:
                        fieldPath: status.podIP             # 用于MindIO通信，如果不配置此参数会影响训练任务的正常拉起。
    ```

3. （可选）如果开启了临终遗言，需要在训练YAML中增加临终遗言通信的端口信息，以pytorch\_multinodes\_acjob\_<i>\{xxx\}</i>b.yaml为例，新增以下加粗内容。

    <pre codetype="yaml">
    ...
       Master:
    ...
              <strong>env:</strong>
                  <strong>- name: TTP_PORT</strong>
                    <strong>value: "8000"     # 用于临终遗言通信，请注意上下保持一致</strong>
    ...
                ports:
                    - containerPort: 2222
                      name: ascendjob-port
                    <strong>- containerPort: 8000     # 用于临终遗言通信，请注意上下保持一致</strong>
                      <strong>name: ttp-port</strong>
                    <strong>- containerPort: 9601     # TaskD Pod间通信端口</strong>
                      <strong>name: taskd-port</strong>
    ...
       Worker:
    ...
              <strong>env:</strong>
                  <strong>- name: TTP_PORT</strong>
                    <strong>value: "8000"            # 用于临终遗言通信，请注意上下保持一致</strong>
    ...
                ports:
                    - containerPort: 2222
                      name: ascendjob-port
                    <strong>- containerPort: 8000     # 用于临终遗言通信，请注意上下保持一致</strong>
                      <strong>name: ttp-port</strong>
                    <strong>- containerPort: 9601     # TaskD Pod间通信端口</strong>
                      <strong>name: taskd-port</strong>

    ...</pre>

4. （可选）如果使用临终遗言和进程级恢复，需要在训练YAML中增加临终遗言通信的端口信息和进程级恢复开关等信息，以pytorch\_multinodes\_acjob\_<i>\{xxx\}</i>b.yaml为例，新增以下加粗内容。

    <pre codetype="yaml">
    ...
      labels:
           framework: pytorch
           ring-controller.atlas: ascend-{xxx}b
           <strong>fault-scheduling: "grace"</strong>
           <strong>fault-retry-times: "10"   // 开启无条件重试</strong>
           <strong>pod-rescheduling: "on"   // 开启Pod级重调度</strong>
           tor-affinity: "null" # 该标签为任务是否使用交换机亲和性调度标签，null或者不写该标签则不适用。large-model-schema表示大模型任务或填充任务，normal-schema表示普通任务
    ...
      annotations:
         ...
         <strong>recover-strategy: "recover,dump"</strong>
      replicaSpecs:
          Master:
            replicas: 1
            <strong>restartPolicy: Never</strong>
            template:
                metadata:
    ...
               <strong>- name: TTP_PORT</strong>
                 <strong>value: "8000"  # 用于MindIO通信，请注意上下保持一致</strong>
            command:                           # training command, which can be modified
              - /bin/bash
              - -c
            args:
              - |
                cd /job/code;
                chmod +x scripts/train_start.sh;
                bash scripts/train_start.sh
             ports:                          # default value
               - containerPort: 2222
                 name: ascendjob-port
               <strong>- containerPort: 8000    # 用于MindIO通信，请注意上下保持一致</strong>
                 <strong>name: ttp-port</strong>
               <strong>- containerPort: 9601    # TaskD Pod间通信端口</strong>
                 <strong>name: taskd-port</strong>
    ...

    ...
      replicaSpecs:
          Worker:
            replicas: 1
            <strong>restartPolicy: Never</strong>
            template:
                metadata:
    ...
                <strong>- name: TTP_PORT</strong>
                <strong>value: "8000"  # 用于MindIO通信，请注意上下保持一致</strong>
            command:                           # training command, which can be modified
              - /bin/bash
              - -c
            args:
              - |
                cd /job/code;
                chmod +x scripts/train_start.sh;
                bash scripts/train_start.sh
             ports:                          # default value
               - containerPort: 2222
                 name: ascendjob-port
               <strong>- containerPort: 8000    # 用于MindIO通信，请注意上下保持一致</strong>
                 <strong>name: ttp-port</strong>
               <strong>- containerPort: 9601    # TaskD Pod间通信端口</strong>
                 <strong>name: taskd-port</strong>
    ...</pre>

5. 使用断点续训功能，建议扩展内存，请按注释添加参数，示例如下。

    ```Yaml
    ...
              volumeMounts:                             #断点续训扩容
             - name: shm
               mountPath: /dev/shm
            volumes:
            - name: shm
              emptyDir:
                medium: Memory
                sizeLimit: 16Gi
    ...
    ```

6. 若需要配置CPU、Memory资源，请参见如下示例手动添加“cpu”和“memory”参数和对应的参数值，具体数值请根据实际情况配置。

    ```Yaml
    ...
              resources:
                requests:
                  huawei.com/Ascend910: 8
                  cpu: 100m
                  memory: 100Gi
                limits:
                  huawei.com/Ascend910: 8
                  cpu: 100m
                  memory: 100Gi
    ...
    ```

7. 修改训练脚本、代码的挂载路径。

    从昇腾镜像仓库拉取的基础镜像中不包含训练脚本、代码等文件，训练时通常使用挂载的方式将训练脚本、代码等文件映射到容器内。

    ```Yaml
              volumeMounts:
              - name: ascend-910-config
                mountPath: /user/serverid/devindex/config
              - name: code
                mountPath: /job/code                     # 容器中训练脚本路径
              - name: data
                mountPath: /job/data                      # 容器中训练数据集路径
              - name: output
                mountPath: /job/output                    # 容器中训练输出路径
    ```

8. （可选）如下所示，YAML中训练命令**bash train\_start.sh**后跟的三个参数依次为容器内训练代码目录、输出目录（其中包括生成日志重定向文件以及框架模型文件）、启动脚本相对代码目录的路径（PyTorch命令参数不涉及启动脚本）。之后的以“--”开头的参数为训练脚本需要的参数。单机和分布式训练脚本、脚本参数可参考模型脚本来源处的模型说明修改。

    >[!NOTE]
    >使用**优雅容错模式**可跳过该步骤。
    - **PyTorch命令参数**

        ```shell
        command:
        - "/bin/bash"
        - "-c"
        - "cd /job/code/scripts;chmod +x train_start.sh;bash train_start.sh /job/code/ /job/output/ main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --lr=1.6 --dist-backend='hccl' --multiprocessing-distributed --epochs=90 --batch-size=1024 --resume=true;"
        ...
        ```

    - 使用**MindSpore架构**的模型，包括ResNet50模型和Pangu\_alpha模型需要跳过此步骤。

9. 选择存储方式。
    - （可选）NFS场景需要指定NFS服务器地址、训练数据集路径、脚本路径和训练输出路径，请根据实际修改。如果不使用NFS请根据K8s相关指导自行修改。

        >[!NOTE]
        >请勿使用ConfigMap挂载RankTable文件，否则可能会导致任务重调度失败。

        ```Yaml
        ...
                  volumeMounts:
                  - name: ascend-910-config
                    mountPath: /user/serverid/devindex/config
                  - name: code
                    mountPath: /job/code                     # 容器中训练脚本路径
                  - name: data
                    mountPath: /job/data                      # 容器中训练数据集路径
                  - name: output
                    mountPath: /job/output                    # 容器中训练输出路径
        ...
                   # 可选，使用Ascend Operator组件为训练任务生成RankTable文件，需要新增以下字段，设置容器中hccl.json文件保存路径，该路径不可修改。
                  - name: ranktable
                    mountPath: /user/serverid/devindex/config
        ...
                volumes:
        ...
                - name: code
                  nfs:
                    server: 127.0.0.1        # NFS服务器IP地址
                    path: "xxxxxx"           # 配置训练脚本路径
                - name: data
                  nfs:
                    server: 127.0.0.1
                    path: "xxxxxx"           # 配置训练集路径
                - name: output
                  nfs:
                    server: 127.0.0.1
                    path: "xxxxxx"           # 设置脚本相关模型的保存路径
        ...
                   # 可选，使用组件为PyTorch框架生成RankTable文件，需要新增以下字段，设置hccl.json文件保存路径
                - name: ranktable         #请勿修改此参数的默认值，Ascend Operator会用于检查是否开启文件挂载hccl.json。
                  hostPath:                    #请使用hostpath挂载或NFS挂载
                    path: /user/mindx-dl/ranktable/default.default-test-pytorch   # 共享存储或者本地存储路径，/user/mindx-dl/ranktable/为前缀路径，必须和Ascend Operator挂载的Ranktable根目录保持一致。default.default-test-pytorch为后缀路径，建议改为:namespace.job-name。
        ...
        ```

    - （可选）如果使用本地存储的挂载方式，需要将YAML中的NFS方式改为hostPath。

        ```Yaml
                  volumes:
                  - name: code
                    hostPath:                                                        # 修改为本地存储
                      path: "/data/atlas_dls/code/resnet/"
                  - name: data
                    hostPath:                                                        # 修改为本地存储
                      path: "/data/atlas_dls/public/dataset/"
                  - name: output
                    hostPath:                                                        # 修改为本地存储
                      path: "/data/atlas_dls/output/"
                  - name: ascend-driver
                    hostPath:
                      path: /usr/local/Ascend/driver
                  - name: dshm
                    emptyDir:
                      medium: Memory
                  - name: localtime
                    hostPath:
                      path: /etc/localtime
        ```
