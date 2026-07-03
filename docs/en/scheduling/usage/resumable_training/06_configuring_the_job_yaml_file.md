# Configuring the Job YAML<a name="ZH-CN_TOPIC_0000002479226518"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T08:00:42.747Z pushedAt=2026-06-09T09:02:55.486Z -->

## Job YAML Configuration Examples<a name="ZH-CN_TOPIC_0000002511346461"></a>

For acjob, understand the YAML parameters before configuring the YAML file. For details, see [acjob Job YAML Parameter Description](../../api/).

For reschedule mode and graceful fault tolerance mode, refer to the following configuration examples in [Procedure](#zh-cn_topic_0000002202737289_zh-cn_topic_0000001951258657_section18181655154219). When the value of `subHealthyStrategy` is `graceExit`, you need to refer to [Configuring Proactive Checkpoint Saving for Sub-health Policy](./05_configuring_training_recovery.md#zh-cn_topic_0000002202737289_zh-cn_topic_0000001951258657_section1048332432310) to complete the adaptation of the startup script and job YAML, ensuring that the checkpoint file can be saved normally before the job is rescheduled due to a sub-health fault.

**Prerequisites<a name="zh-cn_topic_0000002202737289_zh-cn_topic_0000001951258657_section7585519135117"></a>**

You have created a specific mount path for the [hccl.json](../../api/hccl.json_file_description.md) file. For detailed steps, see "Step 4" in [Ascend Operator](../../developer_guide/installation_deployment/manual_installation/08_ascend_operator.md).

**Procedure<a name="zh-cn_topic_0000002202737289_zh-cn_topic_0000001951258657_section18181655154219"></a>**

1. Upload the YAML file to any directory on the management node and modify the file content based on the actual situation.
    - Taking `a800_AscendJob_{xxx}>b.yaml` as an example, create a **distributed training** job on an Atlas 200T A2 Box16 heterogeneous subrack. The job uses 2*4 chips. The modification example is as follows.

        ```Yaml
        apiVersion: mindxdl.gitee.com/v1
        kind: AscendJob
        metadata:
          name: default-test-mindspore
          labels:
            framework: mindspore  # Training framework name
            fault-scheduling: "grace"     # Enable graceful deletion mode
            ring-controller.atlas: ascend-{xxx}b
            fault-retry-times: "3"            # Enable unconditional retry capability for service plane faults, and set the restartPolicy value to Never
            tor-affinity: "normal-schema" # This label indicates whether the job uses switch affinity scheduling. If it is null or not specified, this feature is not used. large-model-schema indicates a large model job or filler job, and normal-schema indicates a normal job
            pod-rescheduling: "on"     # Enable Pod-level rescheduling
            subHealthyStrategy: "ignore"  # Ignore nodes whose health status is sub-healthy, and subsequent jobs will not be scheduled to such nodes in affinity scheduling
        spec:
          schedulerName: volcano    # Takes effect when the startup parameter enableGangScheduling of Ascend Operator is set to true.
          runPolicy:
            backoffLimit: 3      # Number of job rescheduling times
            schedulingPolicy:
              minAvailable: 3       # Total number of job replicas
              queue: default     # Queue to which the job belongs
          successPolicy: AllWorkers  # Prerequisites for job success
          replicaSpecs:
            Scheduler:
              replicas: 1            # Can only be 1
              restartPolicy:  Never   # Container restart policy
              template:
                metadata:
                  labels:
                    ring-controller.atlas: ascend-{xxx}b  # Identifies the product type
                spec:
                  terminationGracePeriodSeconds: 360  # The time from when the container receives SIGTERM to when it is forcibly stopped by K8s
                  nodeSelector:
                    host-arch: huawei-x86          # Atlas 200T A2 Box16 heterogeneous subrack only has the x86_64 architecture
                    accelerator-type: module-{xxx}b-16   # Node type
                  containers:
                  - name: ascend     # Cannot be modified
        ...
                    ports:                     # Optional; distributed training collective communication port
                      - containerPort: 2222
                        name: ascendjob-port
                    volumeMounts:
        ...

            Worker:
              replicas: 2
              restartPolicy: Never  # Container restart policy
              template:
                metadata:
                  labels:
                    ring-controller.atlas: ascend-{xxx}b  # Identifies the product type
                spec:
                  terminationGracePeriodSeconds: 360   # Time elapsed from when a container receives SIGTERM to when it is forcibly stopped by K8s
                  affinity:
        ...
                  nodeSelector:
                    host-arch: huawei-x86      # Atlas 200T A2 Box16 heterogeneous subrack only has x86_64 architecture
                    accelerator-type: module-{xxx}b-16   # Node type
                  containers:
                  - name: ascend      # Cannot be modified
        ...
                    env:
                    - name: ASCEND_VISIBLE_DEVICES
                      valueFrom:
                        fieldRef:
                          fieldPath: metadata.annotations['huawei.com/Ascend910']         # Must be consistent with the resources and requests below
        ...

                    ports:        # Optional; distributed training collective communication port
                      - containerPort: 2222
                        name: ascendjob-port
                    resources:
                      limits:
                        huawei.com/Ascend910: 4      # The number of NPU chips required is 4
                      requests:
                        huawei.com/Ascend910: 4       # Consistent with the limits value
        ```

    - Taking `a800_vcjob.yaml` as an example, create a **single-node training** job on an Atlas 800 training server. The job uses 8 chips. The modification example is as follows.

        ```Yaml
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: rings-config-mindx-dls-test     # The name after rings-config- must be consistent with the job name
        ...
          labels:
            ring-controller.atlas: ascend-910  # Identifies the product type
        ...
        ---
        apiVersion: batch.volcano.sh/v1alpha1   # Cannot be modified. Must use the Volcano API.
        kind: Job                               # Currently, only the Job type is supported
        metadata:
          name: mindx-dls-test                  # Job name, customizable
          labels:
            ring-controller.atlas: ascend-910
            fault-scheduling: "grace"        # Enable graceful deletion mode
            fault-retry-times: "3"            # Enable unconditional retry capability for service plane failures. At the same time, set the restartPolicy value to Never; set the event of policies to PodFailed, and set the action to Ignore.
            tor-affinity: "normal-schema" # This label indicates whether the job uses switch affinity scheduling. If it is null or not specified, this feature is not used. large-model-schema indicates a large model job or a filler job, and normal-schema indicates a normal job.
            pod-rescheduling: "on"     # Enable Pod-level rescheduling.
            subHealthyStrategy: "ignore"     # Ignore nodes whose health status is sub-healthy. Subsequent jobs will not be scheduled to such nodes in affinity scheduling.
        ...
        spec:
          policies:  # When using the rescheduling feature, you do not need to modify the content of policies.
            - event: PodFailed
              action: Ignore
        ...
          minAvailable: 1                  # 1 for a single node
        ...
          maxRetry: 3              # Number of rescheduling times
        ...
          - name: "default-test"
              replicas: 1                  # 1 for a single node
              template:
                metadata:
        ...
                spec:
                  terminationGracePeriodSeconds: 360  # The time from when a container receives SIGTERM to when it is forcibly stopped by K8s
        ...
                    env:
        ...
                  - name: ASCEND_VISIBLE_DEVICES                       # Ascend Docker Runtime uses this field
                    valueFrom:
                      fieldRef:
                        fieldPath: metadata.annotations['huawei.com/Ascend910']               # Must be consistent with the resources and requests below
        ...
                    resources:
                      requests:
                        huawei.com/Ascend910: 8          # The number of NPU chips required is 8. You can add lines below to configure resources such as memory and CPU.
                      limits:
                        huawei.com/Ascend910: 8          # Currently must be consistent with requests above.
        ...
                    nodeSelector:
                      host-arch: huawei-arm               # Optional. Fill in based on the actual situation.
                      accelerator-type: module      # Schedule to Atlas 800 training server.
        ...
                restartPolicy: Never   # Container restart policy
        ```

2. Configure the communication address for MindIO. Add the following content to the code.

    ```Yaml
    ...
       Master:
    ...
                env:
                  - name: POD_IP
                    valueFrom:
                      fieldRef:
                        fieldPath: status.podIP             # Used for MindIO communication. If this parameter is not configured, the normal startup of the training job will be affected.
    ```

3. (Optional) If dying gasp is enabled, you need to add the port information for dying gasp communication in the training YAML. Taking `pytorch_multinodes_acjob_{xxx}b.yaml` as an example, add the following bold content.

    <pre codetype="yaml">
    ...
       Master:
    ...
              <strong>env:</strong>
                  <strong>- name: TTP_PORT</strong>
                    <strong>value: "8000"     # Used for dying gasp communication. Ensure consistency throughout the configuration.</strong>
    ...
                ports:
                    - containerPort: 2222
                      name: ascendjob-port
                    <strong>- containerPort: 8000     # Used for dying gasp communication. Ensure consistency between the upper and lower parts.</strong>
                      <strong>name: ttp-port</strong>
                    <strong>- containerPort: 9601     # Communication port between TaskD Pods</strong>
                      <strong>name: taskd-port</strong>
    ...
       Worker:
    ...
              <strong>env:</strong>
                  <strong>- name: TTP_PORT</strong>
                    <strong>value: "8000"            # Used for dying gasp communication. Ensure consistency throughout the configuration.</strong>
    ...
                ports:
                    - containerPort: 2222
                      name: ascendjob-port
                    <strong>- containerPort: 8000     # Used for dying gasp communication. Ensure consistency throughout the configuration.</strong>
                      <strong>name: ttp-port</strong>
                    <strong>- containerPort: 9601     # Communication port between TaskD Pods</strong>
                      <strong>name: taskd-port</strong>

    ...</pre>

4. (Optional) If using dying gasp and process-level recovery, you need to add the port information for dying gasp communication and the process-level recovery switch in the training YAML. Taking `pytorch_multinodes_acjob_{xxx\}b.yaml` as an example, add the following bold content.

    <pre codetype="yaml">
    ...
      labels:
           framework: pytorch
           ring-controller.atlas: ascend-{xxx}b
           <strong>fault-scheduling: "grace"</strong>
           <strong>fault-retry-times: "10"   // Enable unconditional retry</strong>
           <strong>pod-rescheduling: "on"   // Enable Pod-level rescheduling</strong>
           tor-affinity: "null" # This label specifies whether the job uses switch affinity scheduling. null or omitting this label means it is not applied. large-model-schema indicates a large model job or filler job, normal-schema indicates a normal job
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
                 <strong>value: "8000"  # Used for MindIO communication. Ensure consistency throughout.</strong>
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
               <strong>- containerPort: 8000    # Used for MindIO communication. Ensure consistency throughout the configuration.</strong>
                 <strong>name: ttp-port</strong>
               <strong>- containerPort: 9601    # TaskD inter-pod communication port</strong>
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
                <strong>value: "8000"  # Used for MindIO communication. Ensure consistency throughout the configuration.</strong>
            command:                           # Training command, which can be modified
              - /bin/bash
              - -c
            args:
              - |
                cd /job/code;
                chmod +x scripts/train_start.sh;
                bash scripts/train_start.sh
             ports:                          # Default value
               - containerPort: 2222
                 name: ascendjob-port
               <strong>- containerPort: 8000    # Used for MindIO communication. Ensure consistency throughout the configuratin.</strong>
                 <strong>name: ttp-port</strong>
               <strong>- containerPort: 9601    # Communication port between TaskD Pods</strong>
                 <strong>name: taskd-port</strong>
    ...</pre>

5. When using the resumable training feature, it is recommended to expand the memory. Add the parameters as indicated in the comments. An example is shown below.

    ```Yaml
    ...
              volumeMounts:                             # Memory Expansion for Resumable Training
             - name: shm
               mountPath: /dev/shm
            volumes:
            - name: shm
              emptyDir:
                medium: Memory
                sizeLimit: 16Gi
    ...
    ```

6. If you need to configure CPU and memory resources, manually add the `cpu` and `memory` parameters and their corresponding values as shown in the following example. Configure the specific values based on the actual situation.

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

7. Modify the mount paths for the training script and code.

    The base image pulled from the Ascend image repository does not contain files such as training scripts and code. During training, files such as training scripts and code are typically mapped into the container by mounting.

    ```Yaml
              volumeMounts:
              - name: ascend-910-config
                mountPath: /user/serverid/devindex/config
              - name: code
                mountPath: /job/code                     # Training script path in the container
              - name: data
                mountPath: /job/data                      # Training dataset path in the container
              - name: output
                mountPath: /job/output                    # Training output path in the container
    ```

8. (Optional) As shown below, the three parameters following the training command `bash train_start.sh` in the YAML are the training code directory inside the container, the output directory (which includes generated log redirection files and framework model files), and the path of the startup script relative to the code directory (PyTorch command parameters do not involve a startup script). The subsequent parameters starting with -- are required by the training script. For single-node and distributed training scripts and script parameters, refer to the model description at the model script source for modifications.

    >[!NOTE]
    >This step can be skipped if graceful fault tolerance mode is used.
    - **PyTorch command parameters**

        ```shell
        command:
        - "/bin/bash"
        - "-c"
        - "cd /job/code/scripts;chmod +x train_start.sh;bash train_start.sh /job/code/ /job/output/ main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --lr=1.6 --dist-backend='hccl' --multiprocessing-distributed --epochs=90 --batch-size=1024 --resume=true;"
        ...
        ```

    - Skip this step for models that use the MindSpore architecture, including the ResNet-50 and Pangu_alpha models.

9. Select a storage method.
    - (Optional) For NFS scenarios, you need to specify the NFS server address, training dataset path, script path, and training output path. Modify them according to the actual situation. If you do not use NFS, modify them according to the relevant K8s guidance.

        >[!NOTE]
        > Do not use ConfigMap to mount the RankTable file, as this may cause job rescheduling to fail.

        ```Yaml
        ...
                  volumeMounts:
                  - name: ascend-910-config
                    mountPath: /user/serverid/devindex/config
                  - name: code
                    mountPath: /job/code                     # Training script path in the container
                  - name: data
                    mountPath: /job/data                      # Training dataset path in the container
                  - name: output
                    mountPath: /job/output                    # Training output path in the container
        ...
                   # Optional: To use Ascend Operator to generate the RankTable file for the training job, you need to add the following fields to set the save path for the hccl.json file in the container. This path cannot be modified.
                  - name: ranktable
                    mountPath: /user/serverid/devindex/config
        ...
                volumes:
        ...
                - name: code
                  nfs:
                    server: 127.0.0.1        # NFS server IP address
                    path: "xxxxxx"           # Training script path
                - name: data
                  nfs:
                    server: 127.0.0.1
                    path: "xxxxxx"           # Training dataset path
                - name: output
                  nfs:
                    server: 127.0.0.1
                    path: "xxxxxx"           # Save path for script-related models
        ...
                   # Optional. To use the component to generate a RankTable file for the PyTorch framework, add the following fields to set the save path for the hccl.json file.
                - name: ranktable         # Do not modify the default value of this parameter. Ascend Operator uses it to check whether file mounting for hccl.json is enabled.
                  hostPath:                    # Use hostPath mounting or NFS mounting.
                    path: /user/mindx-dl/ranktable/default.default-test-pytorch   # Shared storage or local storage path. /user/mindx-dl/ranktable/ is the prefix path, which must be consistent with the RankTable root directory mounted by the Ascend Operator. default.default-test-pytorch is the suffix path, which is recommended to be changed to namespace.job-name.
        ...
        ```

    - (Optional) If you use the local storage mounting method, change the NFS method in the YAML to `hostPath`.

        ```Yaml
                  volumes:
                  - name: code
                    hostPath:                                                        # Modify to local storage
                      path: "/data/atlas_dls/code/resnet/"
                  - name: data
                    hostPath:                                                        # Modify to local storage
                      path: "/data/atlas_dls/public/dataset/"
                  - name: output
                    hostPath:                                                        # Modify to local storage
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
