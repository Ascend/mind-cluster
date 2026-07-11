# Overview

The functional interfaces provided by MindCluster Ascend FaultDiag include command interfaces and SDK interfaces. You can implement related functions by calling these interfaces.

Command interfaces: Provide log cleaning, fault diagnosis, single-server fault diagnosis, custom configuration file, custom fault entity, fault log suppression, version query, and help information via direct command call.

SDK interfaces: Code-level interfaces that directly call functions and methods, providing interfaces for service flow parsing, root cause node parsing, root cause node diagnosis, fault event parsing, and fault event diagnosis.

**Table 1** Command interfaces

|Command|Function Description|
|--|--|
|`ascend-fd parse`|Log parsing command that starts a log parsing task to parse intermediate result data collected during training/inference.|
|`ascend-fd diag`|Fault diagnosis command that starts a fault analysis task, analyzes the fault root cause, and outputs an analysis report.|
|`ascend-fd single-diag`|Single-server fault diagnosis command that starts a single-server fault analysis task and outputs an analysis report.|
|`ascend-fd entity`|Custom fault entity command that allows you to customize fault entities. MindCluster Ascend FaultDiag supports functions such as log parsing, fault diagnosis, and fault log shielding for custom faults.|
|`ascend-fd blacklist`|Fault log masking command. Log information containing fault keywords will not be recorded in the log-cleaned output file.|
|`ascend-fd config`|Custom configuration command. You can customize configurations such as whether to support parsing ModelArts key logs, configuring the size of console logs to read, and configuring parsing of custom files.|
|`ascend-fd version`|Version information viewing command that queries component version information.|
|`ascend-fd -h`|Help information query command.|

**Table 2** SDK interfaces

|SDK|Function Description|
|--|--|
|`parse_fault_type`|Service flow parsing interface.|
|`parse_root_cluster`|Root cause node parsing interface.|
|`diag_root_cluster`|Root cause node diagnosis interface.|
|`parse_knowledge_graph`|Fault event parsing interface.|
|`diag_knowledge_graph`|Fault event diagnosis interface.|

When this component is running, operation logs and run logs are generated in the `${HOME}/.ascend_faultdiag` directory. The log directory structure is as follows.

```text
${HOME}/.ascend_faultdiag
└── ascend_faultdiag_operation.log    # Operation log
└── RUN_LOG                           # Run log
  └─ 20241104142355468743_6797877f-7143-443f-a9c6-361e33032c5c
```

The log saving mechanism is as follows: A single log file does not exceed 10 MB. When this size limit is exceeded, logs are automatically rolled over to another log file. The number of log files for the same PID does not exceed 10. Once this limit is reached, the earliest created log file is automatically overwritten.
