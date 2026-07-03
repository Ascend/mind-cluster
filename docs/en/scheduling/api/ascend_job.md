# AscendJob<a name="ZH-CN_TOPIC_0000002479226878"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:41:25.322Z pushedAt=2026-06-09T02:05:50.618Z -->

AscendJob: Abbreviated as acjob, it is a custom job type defined by MindCluster. Currently, it supports launching training or inference jobs through two methods: configuring resource information via environment variables and configuring resource information via files.

## Supported AI Frameworks<a name="zh-cn_topic_0000002377698613_section1580601414413"></a>

- MindSpore
- PyTorch

## Example<a name="zh-cn_topic_0000002377698613_section7389161784012"></a>

The `pytorch_multinodes_acjob_910b.yaml` example is as follows.

```Yaml
apiVersion: mindxdl.gitee.com/v1
kind: AscendJob
metadata:
  name: default-test-pytorch
  labels:
    framework: pytorch
    ring-controller.atlas: ascend-910b
    tor-affinity: "null" #This label indicates whether the job uses switch affinity scheduling. If it is null or not specified, it is not applicable. large-model-schema indicates a large model job, and normal-schema indicates a normal job.
spec:
  schedulerName: volcano   # work when enableGangScheduling is true
  runPolicy:
    schedulingPolicy:      # work when enableGangScheduling is true
      minAvailable: 2
      queue: default
  successPolicy: AllWorkers
  replicaSpecs:
    Master:
      replicas: 1
      restartPolicy: Never
      template:
        metadata:
          labels:
            ring-controller.atlas: ascend-910b
        spec:
          affinity:
            podAntiAffinity:
              requiredDuringSchedulingIgnoredDuringExecution:
                - labelSelector:
                    matchExpressions:
                      - key: job-name
                        operator: In
                        values:
                          - default-test-pytorch
                  topologyKey: kubernetes.io/hostname
          nodeSelector:
            host-arch: huawei-arm
            accelerator-type: card-910b-2 # depend on your device model, 910bx8 is module-910b-8, 910bx16 is module-910b-16
          containers:
          - name: ascend # do not modify
            image: pytorch-test:latest         # training framework image， which can be modified
            imagePullPolicy: IfNotPresent
            env:
              - name: XDL_IP                                       # IP address of the physical node, which is used to identify the node where the pod is running
                valueFrom:
                  fieldRef:
                    fieldPath: status.hostIP
              # ASCEND_VISIBLE_DEVICES is used by ascend-docker-runtime when in the full-npu scheduling scene with volcano scheduler.
              # Please delete it in the following scenarios: static vNPU scheduling, dynamic vNPU scheduling, volcano without Ascend-volcano-plugin, volcano not used
              - name: ASCEND_VISIBLE_DEVICES
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.annotations['huawei.com/Ascend910']               # The value must be the same as resources.requests
            command:                           # training command, which can be modified
              - /bin/bash
              - -c
            args: [ "cd /job/code/scripts; chmod +x train_start.sh; bash train_start.sh /job/code /job/output main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --dist-backend='hccl' --multiprocessing-distributed --epochs=90 --batch-size=4096" ]
            ports:                          # default value containerPort: 2222 name: ascendjob-port if not set
              - containerPort: 2222         # determined by user
                name: ascendjob-port        # do not modify
            resources:
              limits:
                huawei.com/Ascend910: 2
              requests:
                huawei.com/Ascend910: 2
            volumeMounts:
            - name: code
              mountPath: /job/code
            - name: data
              mountPath: /job/data
            - name: output
              mountPath: /job/output
            - name: ascend-driver
              mountPath: /usr/local/Ascend/driver
            - name: ascend-add-ons
              mountPath: /usr/local/Ascend/add-ons
            - name: dshm
              mountPath: /dev/shm
            - name: localtime
              mountPath: /etc/localtime
          volumes:
          - name: code
            nfs:
              server: 127.0.0.1
              path: "/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/"
          - name: data
            nfs:
              server: 127.0.0.1
              path: "/data/atlas_dls/public/dataset/"
          - name: output
            nfs:
              server: 127.0.0.1
              path: "/data/atlas_dls/output/"
          - name: ascend-driver
            hostPath:
              path: /usr/local/Ascend/driver
          - name: ascend-add-ons
            hostPath:
              path: /usr/local/Ascend/add-ons
          - name: dshm
            emptyDir:
              medium: Memory
          - name: localtime
            hostPath:
              path: /etc/localtime
    Worker:
      replicas: 1
      restartPolicy: Never
      template:
        metadata:
          labels:
            ring-controller.atlas: ascend-910b
        spec:
          affinity:
            podAntiAffinity:
              requiredDuringSchedulingIgnoredDuringExecution:
                - labelSelector:
                    matchExpressions:
                      - key: job-name
                        operator: In
                        values:
                          - default-test-pytorch
                  topologyKey: kubernetes.io/hostname
          nodeSelector:
            host-arch: huawei-arm
            accelerator-type: card-910b-2 # depend on your device model, 910bx8 is module-910b-8, 910bx16 is module-910b-16
          containers:
          - name: ascend # do not modify
            image: pytorch-test:latest                # training framework image， which can be modified
            imagePullPolicy: IfNotPresent
            env:
              - name: XDL_IP                                       # IP address of the physical node, which is used to identify the node where the pod is running
                valueFrom:
                  fieldRef:
                    fieldPath: status.hostIP
              # ASCEND_VISIBLE_DEVICES is used by ascend-docker-runtime when in the full-npu scheduling scene with volcano scheduler.
          # Please delete it in the following scenarios: static vNPU scheduling, dynamic vNPU scheduling, volcano without Ascend-volcano-plugin, volcano not used
              - name: ASCEND_VISIBLE_DEVICES
                valueFrom:
                  fieldRef:
                    fieldPath: metadata.annotations['huawei.com/Ascend910']               # The value must be the same as resources.requests
            command:                                  # training command, which can be modified
              - /bin/bash
              - -c
            args: ["cd /job/code/scripts; chmod +x train_start.sh; bash train_start.sh /job/code /job/output main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --dist-backend='hccl' --multiprocessing-distributed --epochs=90 --batch-size=4096"]
            ports:                          # default value containerPort: 2222 name: ascendjob-port if not set
              - containerPort: 2222         # determined by user
                name: ascendjob-port        # do not modify
            resources:
              limits:
                huawei.com/Ascend910: 2
              requests:
                huawei.com/Ascend910: 2
            volumeMounts:
            - name: code
              mountPath: /job/code
            - name: data
              mountPath: /job/data
            - name: output
              mountPath: /job/output
            - name: ascend-driver
              mountPath: /usr/local/Ascend/driver
            - name: ascend-add-ons
              mountPath: /usr/local/Ascend/add-ons
            - name: dshm
              mountPath: /dev/shm
            - name: localtime
              mountPath: /etc/localtime
          volumes:
          - name: code
            nfs:
              server: 127.0.0.1
              path: "/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/"
          - name: data
            nfs:
              server: 127.0.0.1
              path: "/data/atlas_dls/public/dataset/"
          - name: output
            nfs:
              server: 127.0.0.1
              path: "/data/atlas_dls/output/"
          - name: ascend-driver
            hostPath:
              path: /usr/local/Ascend/driver
          - name: ascend-add-ons
            hostPath:
              path: /usr/local/Ascend/add-ons
          - name: dshm
            emptyDir:
              medium: Memory
          - name: localtime
            hostPath:
              path: /etc/localtime

```

## Job Status Description<a name="zh-cn_topic_0000002377698613_section177175313294"></a>

After starting a training Job, you can run the `kubectl get acjob` command to view the running status of the acjob. The current running statuses are as follows.

**Table 2**  Description of acjob statuses

|Status|Description|
|--|--|
|Created|The job has been created, but one or more of its sub-resources (Pod/Service) are not yet ready.|
|Running|All sub-resources (Pod/Service) of the job have been scheduled and started.|
|Restarting|One or more sub-resources (Pod/Service) of the job failed to run, but are being restarted according to the restart policy.|
|Succeeded|All sub-resources (Pod/Service) of the job are in the successful termination phase.|
|Failed|One or more sub-resources (Pod/Service) of the job failed to run.|

## Description of Job Exception Conditions<a name="zh-cn_topic_0000002377698613_section177175313295"></a>

When a job encounters an exception, the `status.conditions` field of Ascend Job records detailed exception information. Each condition contains the following fields:

|Field|Type|Description|
|--|--|--|
|type|string|Condition type, such as Failed, Restarting, Running, Succeeded, Created|
|status|string|Condition status: True, False, Unknown|
|lastTransitionTime|string|Time when the condition status transitioned (RFC3339 format)|
|lastUpdateTime|string|Final time after the condition was updated (RFC3339 format)|
|message|string|Detailed description of the condition|
|reason|string|Reason code for the condition transition|

## Description of Common Exception Reasons

|Code|Description|
|--|--|
|JobFailed|The job failed, usually because a Pod failed.|
|jobRestarting|The job is restarting, and the failed Pod is being restarted according to the restart policy.|
|SyncPodGroupFailed|Failed to synchronize the PodGroup.|
|PodGroupNotInitialized|The PodGroup is not initialized, usually because the volcano-scheduler is not running.|
|PodGroupPending|The PodGroup is in a pending state, usually due to insufficient cluster resources.|
|SyncServiceFailed|Failed to synchronize the Service.|
|PodCreateFailed|Failed to create the Pod.|
|JobValidFailed|Job validation failed.|

## Example of Exception Conditions

```yaml
status:
  conditions:
  - type: Failed
    status: "True"
    lastTransitionTime: "2024-01-01T10:00:00Z"
    lastUpdateTime: "2024-01-01T10:00:00Z"
    message: "Job default/test-job has failed because has pod failed."
    reason: "JobFailed"
  - type: Restarting
    status: "True"
    lastTransitionTime: "2024-01-01T10:00:00Z"
    lastUpdateTime: "2024-01-01T10:00:00Z"
    message: "Job default/test-job is unconditional retry job and remain retry times is <3>."
    reason: "jobRestarting"
  - type: Failed
    status: "True"
    lastTransitionTime: "2024-01-01T10:00:00Z"
    lastUpdateTime: "2024-01-01T10:00:00Z"
    message: "Job test-job has failed because it has reached the specified backoff limit"
    reason: "JobFailed"
```

## Viewing Job Exception Information

Use the following command to view the detailed status and exception information of the job:

```bash
# View AscendJob Status
kubectl get acjob -n <namespace> <job-name> -o yaml

# View AscendJob Status Summary
kubectl get acjob -n <namespace> <job-name> -o jsonpath={.status.conditions}

# View the Latest Status of AscendJob
kubectl get acjob -n <namespace> <job-name> -o jsonpath={.status.conditions[-1]}
```
