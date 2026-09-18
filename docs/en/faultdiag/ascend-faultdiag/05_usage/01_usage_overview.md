# Feature Overview

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:27:13.692Z pushedAt=2026-08-24T02:51:43.243Z -->

ascend-fd provides the following features. This chapter describes the usage and applicable scenarios of each feature.

## Usage Instructions

- Due to the Linux system's maximum file descriptor limit (1024 by default), the cluster scale is recommended not to exceed 128 servers (1024 cards) for the diagnostics feature. If this scale is exceeded, use the `ulimit -n <value>` command to adjust the file descriptor limit. For details, see [Diagnosis Fails due to Large-Scale Clusters](../07_references/02_faq.md#diagnosis-fails-in-large-scale-clusters).
- When using the `ascend-fd` command, avoid using pipe commands as much as possible, as they may affect user IP acquisition and log auditing.

## Task Specifications

ascend-fd provides fault diagnostics only for training or inference tasks that occupy **all cards on the server**. In non-full-card scenarios, diagnostics may result in incorrect or failed root cause localization.

## System Time Requirements

- Synchronize the system time of all training/inference servers. Inconsistent time may lead to inaccurate analysis results.
- Synchronize the host system time and device system time on each server.
- If tasks are executed in containers, synchronize the system time between the host machine and the containers.

## Log Version Compatibility

| Log File             | Corresponding Software                  | Software Version                                           | Description                          |
|----------------------|-----------------------------------------|-----------------------------------------------------------|--------------------------------------|
| CANN App log         | CANN                                    | 7.0.RC1 or later                                          | Host-side and device-side App log    |
| PyTorch framework log | PyTorch adapter plugin                 | 5.0.RC3 or later                                          | -                                    |
| MindSpore framework log | MindSpore                             | 2.1.0 or later                                            | Some faults have specific version requirements |
| Host OS log          | Operating system                        | CentOS 7.6, Debian 10.0, EulerOS 2.10/2.12, CTyunOS 22.06 | Recommend the log size to be within 512 MB    |
| Device-side log      | Ascend HDK                              | 23.0.RC3 or later                                         | -                                    |
| MindCluster component log | Ascend Device Plugin, NodeD, Volcano, etc. | 6.0.RC3 or later                                      | -                                    |
| MindIE component log | MindIE Server, LLM, SD, RT, etc.        | 6.0.0 or later                                            | -                                    |
| AMCT log   | AMCT model compression                  | 7.0.RC1 or later                                          | -                                    |
| PyMotor/vLLM log     | MindIE-PyMotor, vLLM                    | 26.1.0 or later                                           | -                                    |

## Feature List

| Feature                                                          | Description                            |
|------------------------------------------------------------------|----------------------------------------|
| [Log Collection](02_log_collection.md)                           | Collects log files related to training/inference tasks |
| [Log Parsing](03_log_parsing.md)                                 | Extracts key information from raw logs |
| [Fault Diagnosis](04_fault_diagnosis.md)                         | Analyzes root cause nodes and fault events |
| [Single-Server Fault Diagnosis](05_single_server_diagnosis.md)   | Quickly diagnoses faults on a single server   |
| [SuperPoD Fault Diagnosis](06_superpod_diagnosis.md)             | Diagnostics for Atlas A3 SuperPoD     |
| [Custom Fault Entities](07_custom_fault_entities.md)             | Adds user-defined fault detection rules |
| [Fault Log Masking](08_fault_log_masking.md)                     | Filters out log information that does not need attention |
| [Custom Configuration](09_custom_configuration.md)               | Customizes log parsing and diagnosis behavior |
| [Service Flow Parsing (SDK)](10_service_flow_parsing.md)         | Parses service logs through SDK interfaces |
| [Root Cause Parsing and Diagnosis (SDK)](11_root_cause_parsing_diagnosis.md) | Performs root cause analysis through SDK interfaces |
| [Fault Event Parsing and Diagnosis (SDK)](12_fault_event_parsing_diagnosis.md) | Performs fault event analysis through SDK interfaces |
