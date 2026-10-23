# Service Log Parsing (SDK)

You can use the Python SDK interfaces of ascend-fd to parse service logs and extract key information.

## Applicable Scenarios

- Integrate the log parsing capability into your own Python programs.
- Service logs need to be processed programmatically.

## Procedure

1. Import the SDK.

    ```python
    from ascend_fd import parse_fault_type
    ```

2. Call the parsing API.

    ```python
    result, err_msg_list = parse_fault_type(input_log_list)
    ```

## Input Parameters and Return Values

For the input parameters and return values of service log parsing, see the [parse_fault_type](../06_api/09_sdk_api.md#parse_fault_type)  definition in [SDK API Reference](../06_api/09_sdk_api.md).
