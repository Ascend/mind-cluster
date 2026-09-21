# Before You Start<a name="ZH-CN_TOPIC_0000002518340700"></a>

<!-- md-trans-meta sourceCommit=cc2549abde258f50a946dbab123b7ff2035ed91e translatedAt=2026-09-01T06:14:08.995Z pushedAt=2026-09-01T06:16:53.026Z -->

In scenarios without K8s, Container Manager can be used to detect and handle NPU hardware faults and automatically recover containers.

## Prerequisites<a name="section1632062465010"></a>

Before using this feature, ensure that the following components are installed. If they are not installed, refer to [Installation and Deployment](../../05_developer_guide/00_installation_deployment/00_manual_installation/00_obtaining_software_packages.md) for instructions.

- Container Manager
- Ascend Docker Runtime (optional)

## Usage Instructions<a name="section44381612353"></a>

- This feature applies to scenarios without K8s and does not depend on the K8s scheduler.
- This feature does not apply to compute power virtualization scenarios, and does not support the shared device feature or mixed insertion mode.
- Privileged containers are managed only when NPUs are explicitly mounted through device configuration or the `ASCEND_VISIBLE_DEVICES` environment variable.

## Supported Product Forms<a name="section169961844182917"></a>

The following products support fault management and automatic recovery of faulty containers:

- <term>Atlas training products</term>
- <term>Atlas A2 training products</term>
- <term>Atlas A3 training products</term>
- <term>Atlas inference products</term>
- <term>Atlas A2 inference products</term>
- <term>Atlas A3 inference products</term>
