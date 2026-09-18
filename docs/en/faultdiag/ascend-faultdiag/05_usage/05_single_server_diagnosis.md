# Single-Server Fault Diagnosis

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:27:17.459Z pushedAt=2026-08-24T02:51:43.245Z -->

Single-server fault diagnosis can quickly complete log cleaning and fault diagnosis on a single server, without the need for multi-server log dumping.

It is suitable for quickly troubleshooting issues on a single server, without the need for cross-server analysis.

## Procedure

1. Collect logs according to [Log Collection](./02_log_collection.md).

2. Create an output directory for single-server diagnosis results.

    ```shell
    mkdir <single_server_diagnostic_output_directory>
    ```

3. Run the diagnosis command.

    ```shell
    ascend-fd single-diag -i <collection_directory> -o <single_server_diagnostic_output_directory>
    ```

    Single-server diagnosis returns fault event analysis results by default.

## Diagnostic Report and Results

For interpretation of the diagnostic report, see [Diagnostic Report Interpretation](./04_fault_diagnosis.md#diagnosis-report-interpretation).

For the diagnostic result file, see [Diagnostic Result File](./04_fault_diagnosis.md#diagnostic-result-file).

> [!NOTE]
> Single-server diagnosis scans fault events in all valid logs on a node. If an error occurs during diagnosis execution, you can view all exception information in the `diag_report.json` file.
>
