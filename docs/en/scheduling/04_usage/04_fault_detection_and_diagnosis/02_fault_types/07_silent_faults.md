# Silent Faults

## Overview

A silent fault is a special public fault comprehensively determined by ClusterD based on software fault and hardware fault information when the Pods of a job actively exit abnormally repeatedly but no hardware fault is reported on the node. After a fault is determined as a silent fault, ClusterD only appends the fault to the fault list of the `cluster-info-device-xxx` ConfigMap, and does not change the chip health status, so rescheduling is not triggered. In addition, ClusterD writes the fault to the public fault cache (`statistic-fault-info`) and records a silent fault entry in `clusterd-manual-info-cm`.

## Determination Mechanism

ClusterD obtains the rescheduling records of a job by listening to the `job-reschedule-reason` ConfigMap, and determines a silent fault comprehensively in combination with the hardware fault information of the node. All of the following conditions must be met to determine a silent fault (the thresholds below are default values and can be adjusted through configuration; for parameter descriptions, see [Configuring Silent Fault Detection](../03_configuration/09_silent_faults.md)):

- The total number of NPUs of the rescheduled job is not less than 16, and the job is a full-node job.
- Only one `pod-failed` event occurs in a single rescheduling. Batch Pod rescheduling is not counted.
- Within 30 seconds before and after the rescheduling, no hardware fault is reported on the node. Hardware faults include chip-level, node-level, switch-level, and external public faults, excluding silent faults themselves.
- Counting by node, 3 valid first-error records of software faults occur consecutively within the most recent 3 hours.

## Required Components

To ensure that the silent fault detection function works properly, install the following components.

- Mandatory components: Volcano, Ascend Device Plugin, and ClusterD
- Optional component: NodeD

## Usage Constraints<a name="section_silent_faults_constraint"></a>

- Silent fault detection is not supported when the `fault-scheduling` value of a job is `external-force-pod-failed`. With this value, the job does not record the rescheduling reason, and ClusterD cannot obtain the rescheduling record of the job from the `job-reschedule-reason` ConfigMap, so a silent fault is not determined.

## Related Operations

- [Configuring Silent Fault Detection](../03_configuration/09_silent_faults.md)
- [Configuring Public Faults](../03_configuration/08_public_faults.md)
