# Log Parsing

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:27:22.858Z pushedAt=2026-08-24T02:51:43.247Z -->

Log parsing is the process of extracting key information from raw logs. After parsing is complete, the valid information in the raw logs is extracted for subsequent diagnosis.

## Usage

### Prerequisites

1. Ensure that logs have been collected as required in [Log Collection](./02_log_collection.md#log-collection).

2. Create the parsing output directory.

    ```shell
    mkdir <parsing_output_dir>
    ```

### Full Parsing (Recommended)

All module logs must be stored according to the requirements in [Log Collection Archive Directory Structure](./02_log_collection.md#log-collection-archive-directory-structure).

Run the parsing command:

```shell
ascend-fd parse -i <collection_directory> -o <parsing_output_directory>
```

If you need to parse performance degradation (device resources and network congestion) data at the same time, add the `-p` parameter:

```shell
ascend-fd parse -i <collection_directory> -o <parsing_output_directory> -p
```

The following output indicates that parsing is successful:

```text
The parse job starts. Please wait. Job id: [****], run log file is [****].
These job ['Module 1', 'Module 2'...] succeeded.
The parse job is complete.
```

### Module-Specific Parsing

Specify input directories by module log type:

```shell
ascend-fd parse \
    --host_log <host-side OS log directory> \
    --device_log <Device-side log directory> \
    --train_log <user training and inference log directory> \
    --process_log <CANN application log directory> \
    --env_check <NPU port/status/resource information directory> \
    --dl_log <MindCluster component log directory> \
    --mindie_log <MindIE component log directory> \
    --amct_log <AMCT component log directory> \
    --pymotor_vllm_log <PyMotor/vLLM log directory> \
    --bmc_log <BMC-side log directory> \
    --lcne_log <LCNE component log directory> \
    -o <parsing_output_directory>
```

> [!NOTE]
>
> When `-i` is used together with detailed log directory parameters, the values of the detailed log directory parameters are read first, and then the remaining log directories are read according to the `-i` parameter.

### Detailed Parsing Parameter Descriptions

See [Detailed Parameter Description of parse](../06_api/02_command_parse.md#parameter-description).

### Parsing Output Results

See [parse Output Description](../06_api/02_command_parse.md#parsing-output-result).

## Multi-Node Fault Diagnosis

For multi-node fault diagnosis, the parsing output of each node must be aggregated into the same directory.

After parsing is complete on each server, the parsing output of all servers must be aggregated onto the same server. The directory structure is as follows:

```text
Diagnostic input directory
    |-- Parsing Output Directory 1 (recommended to use the node identifier, e.g., host1-192.168.1.1)
    |-- Parsing Output Directory 2
    └── Parsing Output Directory N
```

## Notes

- The parsing output directory must have more than 5 GB of available drive space. Insufficient space may result in partial loss of parsing results.
- During parsing, ensure that the directory to be parsed contains logs from only a single server.
