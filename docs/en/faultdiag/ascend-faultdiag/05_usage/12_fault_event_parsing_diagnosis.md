# Fault Event Parsing and Diagnosis (SDK)

<!-- md-trans-meta sourceCommit=0bdfcd26262ca7cc1165bb55c8f28c6a29b377ef translatedAt=2026-08-24T02:28:17.674Z pushedAt=2026-08-24T02:51:43.270Z -->

Through the Python SDK interfaces of ascend-fd, all fault events in logs can be parsed and diagnosed based on the knowledge graph to identify specific fault types and causes.

## Applicable Scenarios

- Integrate fault event analysis capabilities into your own Python programs.
- Programmatic diagnosis of fault events is required.

## Procedure

Recommended workflow: fault event parsing → fault event diagnosis.

1. Import the SDK.

    ```python
    from ascend_fd import parse_knowledge_graph, diag_knowledge_graph
    ```

2. Call the parsing API.

    This API returns the fault event parsing results and errors that occurred during parsing.

    ```python
    kg_parse_results, kg_parse_err_msg = parse_knowledge_graph(input_log_list, custom_entity)
    ```

3. Call the diagnosis API.

    Use the parsed `kg_parse_results` as the input to the diagnosis API to diagnose fault events.

    This API returns the fault event diagnosis results and any errors that occurred during the diagnosis process.

    ```python
    result, err_msg_list = diag_knowledge_graph(kg_parse_results)
    ```

## Input Parameters and Return Values

For the input parameters and return values of fault event parsing, see the [parse_knowledge_graph](../06_api/09_sdk_api.md#parse_knowledge_graph) definition in [SDK API Reference](../06_api/09_sdk_api.md).

For the input parameters and return values of fault event diagnosis, see the [diag_knowledge_graph](../06_api/09_sdk_api.md#diag_knowledge_graph) definition in [SDK API Reference](../06_api/09_sdk_api.md).

## Reference

- For the complete SDK API description, see [SDK API Reference](../06_api/09_sdk_api.md).
- For the command-line method, see [Fault Diagnosis](./04_fault_diagnosis.md).
