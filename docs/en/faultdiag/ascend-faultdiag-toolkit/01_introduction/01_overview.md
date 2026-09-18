# Introduction

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:15:02.955Z pushedAt=2026-08-24T02:22:41.512Z -->

## Overview

MindCluster Ascend FaultDiag Toolkit (abbreviated as ascend-fd-tk) is a link fault diagnosis tool for Ascend AI clusters. The tool provides both interactive and non-interactive operation modes, and supports two data processing capabilities: online automatic data collection and offline log parsing. It can collect cluster device information, conduct automated inspection diagnosis, and locate link faults in a cluster through information of servers, L1/L2 UnifiedBus switches, RoCE switches, and BMC.

## Applicable Scenarios

ascend-fd-tk is applicable to the following scenarios:

- **Service fault locating**: When a training or inference task is abnormal, quickly collect logs and locate the link fault.
- **Unified diagnosis across multiple network planes**: Supports batch collection and aggregated unified diagnosis in multi-network-plane scenarios.
- **Customized inspection**: Customize inspection dimensions as needed to meet link inspection requirements in customer-specific scenarios (this feature is a beta feature and is not recommended for use in production environments).
- **Periodic cluster diagnosis**: Supports routine cluster fault diagnosis and identifies potential link anomalies in advance.

## Core Capabilities

| Capability | Description                                                                 |
|------|--------------------------------------------------------------------|
| Online collection and parsing| Automatically collects data over SSH and parses information based on the configured connection information (account, password/key/password-free) of cluster devices (servers, BMCs, and switches). |
| Offline log parsing | Imports collected in-band server logs, BMC dump logs, and switch diagnostic information logs for information parsing.    |
| One-click diagnosis | Automatically collects information and diagnoses faults, and generates a diagnostic report.                                               |
| Batched collection and unified diagnosis | Supports batched device information collection in multi-network-plane scenarios, followed by unified diagnosis after information aggregation.                                         |
| Customized inspection | Supports customized diagnosis rules for inspection based on customer types (beta feature; not recommended for use in production environments).                      |

## Supported Operating Systems

ascend-fd-tk supports the following operating systems:

| Operating System | Version Requirement                              | Architecture | Description                      |
|----------|------------------------------------------------|------|-------------------------|
| Linux | CentOS 7.6+/Ubuntu 18.04+/openEuler 20.03+ | x86_64/aarch64 | Recommended                      |
| Windows | Windows 10/Windows 11                        | x86_64 | This is a beta feature and is not recommended for use in production environments. |

## Usage Process

Complete fault diagnosis by following the link fault diagnosis procedure below.

1. **Install the tool**: Install the tool using the Whl package.
2. **Configure the configuration file** (optional): Provide additional configuration information such as the equipment room location.
3. **Configure the data source/Set the log path**: In online mode, configure the connection file; in offline mode, set the log directory.
4. **Parse logs**: In online mode, device logs are automatically collected and parsed; in offline mode, log information is directly parsed.
5. **Conduct diagnosis/inspection**: Perform fault diagnosis or inspection on the parsed data.
6. **View report**: After diagnosis/inspection is complete, the report is automatically generated in the [report subdirectory under the home directory](../05_usage/01_usage_overview.md).

>[!NOTE]
> Before each use of the tool, it is recommended to run `clear_cache` to clear the cache, to prevent residual data from the previous diagnosis task from affecting the current diagnosis result.

## Disclaimer

- This document may contain third-party information, products, services, software, components, data, or content (collectively referred to as "third-party content"). Huawei does not control and assumes no responsibility for any third-party content, including but not limited to its accuracy, compatibility, reliability, availability, legality, appropriateness, performance, non-infringement, update status, unless otherwise expressly stated in this document. Any mention or reference to third-party content in this document does not constitute endorsement or warranty by Huawei of such third-party content.
- This feature reads and processes the relevant raw logs and monitoring metric files collected by users from the input directory. Users must ensure that such files contain no sensitive information or personal data. Huawei does not control and assumes no responsibility for the content of the input data.
- If users require third-party licenses, they must obtain such licenses through lawful means, unless otherwise expressly stated in this document.
