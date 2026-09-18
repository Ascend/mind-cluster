# Root Cause Node Parsing and Diagnosis (SDK)

<!-- md-trans-meta sourceCommit=0bdfcd26262ca7cc1165bb55c8f28c6a29b377ef translatedAt=2026-08-24T02:28:09.472Z pushedAt=2026-08-24T02:51:43.268Z -->

Through the Python SDK APIs of ascend-fd, you can parse and diagnose root cause nodes in logs to identify the device that triggered the fault.

## Applicable Scenarios

- Integrate root cause analysis into your own Python programs.
- Programmatic diagnosis of root cause nodes is required.

## Procedure

Recommended workflow: root cause node parsing → root cause node diagnosis.

1. Import the SDK.

    ```python
    from ascend_fd import parse_root_cluster, diag_root_cluster
    ```

2. Call the parsing API.

    This API returns the root cause node parsing results and any errors that occurred during parsing.

    ```python
    rc_parse_results, rc_parse_err_msg = parse_root_cluster(input_log_list)
    ```

3. Call the diagnosis API.

    Use the parsed `rc_parse_results` as the input to the diagnosis API to perform root cause node diagnosis.

    This API returns the root cause node diagnosis result and the errors that occurred during the diagnosis process.

    ```python
    result, err_msg_list = diag_root_cluster(rc_parse_results)
    ```

## Input Parameters and Return Values

For the input parameters and return values of root cause node parsing, see the [parse_root_cluster](../06_api/09_sdk_api.md#parse_root_cluster) definition in [SDK API Reference](../06_api/09_sdk_api.md).

For the input parameters and return values of root cause node diagnosis, see the [diag_root_cluster](../06_api/09_sdk_api.md#diag_root_cluster) definition in [SDK API Reference](../06_api/09_sdk_api.md).

## Reference

- For the complete SDK API description, see [SDK API Reference](../06_api/09_sdk_api.md).
- For the command-line method, see [Fault Diagnosis](./04_fault_diagnosis.md).
