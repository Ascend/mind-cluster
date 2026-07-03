# Alarm Reference

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T06:28:47.151Z pushedAt=2026-06-09T07:15:15.742Z -->

## ALM-0x1001001  MindIO ACP Persistent Checkpoint Data Anomaly

**Alarm Description**

This alarm is generated when the backend storage system fails.

The alarm is cleared when the backend storage system recovers.

**Alarm Attributes**

|Alarm ID|Alarm Severity|Auto-clearable|
|--|--|--|
|0x1001001|Major|Yes|

**Alarm Parameters**

None

**Impact on the System**

MindIO ACP becomes unavailable, and the system falls back to directly operating the back-end storage on the user side.

**Possible Causes**

- Backend storage system failure.
- Insufficient permissions to operate backend storage files.

**Handling Procedure**

1. Check whether the backend storage status is normal.
    - If the status is normal, perform [Step 2](#step_acp_li007).
    - If the status is abnormal, perform [Step 3](#step_acp_li008).

2. <a id="step_acp_li007"></a>Check whether the username and group permissions of the files in the backend storage are consistent with the permissions of the client process.
    - If the permissions are consistent, the alarm will be automatically cleared.
    - If the permissions are inconsistent, perform [Step 3](#step_acp_li008).

3. <a id="step_acp_li008"></a>Collect fault or log information and contact technical support for handling.

**Reference**

N/A

**Alarm Clearing**

After this alarm is rectified, the system automatically clears it. Manual clearing is not required.
