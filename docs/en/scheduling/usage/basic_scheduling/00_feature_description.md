# Feature Description<a name="ZH-CN_TOPIC_0000002511347091"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-22T09:45:42.901Z pushedAt=2026-06-22T09:46:04.501Z -->

Basic scheduling features:

- Training jobs: [Full-NPU Scheduling](../../introduction/02_feature_description.md#full-npu-scheduling), [Static vNPU Scheduling](../../introduction/02_feature_description.md#static-vnpu-scheduling), [Multi-Level Scheduling](../../introduction/02_feature_description.md#multi-level-scheduling), [Elastic Training](../../introduction/02_feature_description.md#elastic-training), and [Resumable Training](../../usage/resumable_training/00_feature_description.md).
- Inference jobs: [Full-NPU Scheduling](../../introduction/02_feature_description.md#full-npu-scheduling), [Static vNPU Scheduling](../../introduction/02_feature_description.md#static-vnpu-scheduling), [Dynamic vNPU Scheduling](../../introduction/02_feature_description.md#dynamic-vnpu-scheduling), [Soft Partitioning-based Scheduling](../../introduction/02_feature_description.md#soft-partitioning-based-scheduling), [Recovery of Inference Card Faults](../../introduction/02_feature_description.md#recovery-of-inference-card-faults), and [Rescheduling upon Inference Card Faults](../../introduction/02_feature_description.md#rescheduling-upon-inference-card-faults).

    Different features depend on different components. For details, see the [Basic Scheduling](../../introduction/02_feature_description.md#basic-scheduling) chapter.

This document demonstrates how to deploy and execute training or inference jobs using NPUs based on a certain model. There are differences between the production environment and the examples. The examples in this chapter are for reference only, and you need to modify them according to your production environment.

## Job Types<a name="section14151030191813"></a>

Ascend Operator provides the following two ways to configure resource information:

- Configure resource information through environment variables: Provide corresponding environment variables for distributed training jobs of different AI frameworks. See [Ascend Operator Environment Variable Description](../../api/environment_variable_description.md). Users who use this method can only create Ascend Job (hereinafter referred to as acjob) objects.
- Configure resource information through files: Training job collective communication configuration file (RankTable File, also known as [hccl.json](../../api/hccl.json_file_description.md)). Users who use this method can create the following three types of objects: Volcano Job (hereinafter referred to as vcjob), Ascend Job (hereinafter referred to as acjob), and Deployment (hereinafter referred to as deploy).
    - (Recommended) Ascend Job: It is a custom job type defined by MindCluster. Currently, it supports launching training or inference jobs through two methods: configuring resource information via environment variables and configuring resource information via files.

        Each acjob YAML contains some fixed fields, such as `apiVersion`, `kind`, etc. For detailed descriptions of these fields, please refer to [acjob Yaml Description](../../api).

    - Volcano Job: I is suitable for batch processing jobs with a completion status.
    - Deployment: It is suitable for background resident jobs without a completion status. It is selected when continuous training jobs, persistent resource occupation, debugging training jobs, or providing inference service interfaces are required.

        >[!NOTE]
        >The update operation of Deployment is not supported. If an update is needed, please delete the job first and then create it again.

## Scheduling Time Description<a name="section12177114564719"></a>

In multi-job or single-job scenarios, the reference scheduling time for acjob on the Atlas 800T A2 training server is described as follows. To achieve the following reference times, ensure that the CPU frequency is at least 2.60 GHz and the API Server latency does not exceed 80 milliseconds. The scheduling time refers to the period from when a job is submitted until the Pod status changes to Running..

- Multi-job scheduling time
    - The peak number of concurrently created single-server single-device jobs is 100, meaning that 100 jobs are created simultaneously using 100 job YAML files, and the scheduling time for these 100 jobs is 107 seconds.
    - The stable creation rate for single-server single-device tasks is 5 per second. After one minute of sustained creation, 300 such jobs will have been created. The scheduling time for these 300 jobs is 293 seconds.

- Single-job scheduling time

    **Table 1** Single-job multi-pod scheduling description

    <a name="table18378013481"></a>

    |Number of Cluster Nodes|Number of Pods|Scheduling Time|
    |--|--|--|
    |100|100|14 seconds|
    |500|500|57 seconds|
    |1000|1000|114 seconds|
    |2000|2000|228 seconds|
    |3000|3000|269 seconds|
    |4000|4000|300 seconds|
    |5000|5000|400 seconds|

    >[!NOTE]
    >- The single-job multi-pod scenario refers to creating multiple pods with one job YAML. For example, creating 100 pods with one job YAML and scheduling these 100 pods to 100 nodes takes 14 seconds.
    >- To achieve the optimized scheduling reference time for 4,000 or 5,000 nodes, you need to follow the scheduling time performance tuning steps in [Installing Volcano](../../installation_guide/03_installation/manual_installation/05_volcano.md) to make the corresponding modifications.
    >- Currently, the scheduling specification for vcjob supports a maximum of 1,000 nodes.
