# 故障事件清洗及诊断（SDK）

通过 ascend-fd 的 Python SDK 接口，基于知识图谱可以完成日志中所有故障事件清洗和诊断，识别具体故障类型和原因。

## 适用场景

- 用户在自己的 Python 程序中集成故障事件分析功能。
- 需要对故障事件进行程序化诊断。

## 操作步骤

建议使用流程：故障事件清洗 → 故障事件诊断。

1. SDK 导入

    ```python
    from ascend_fd import parse_knowledge_graph, diag_knowledge_graph
    ```

2. 调用清洗接口

    该接口返回故障事件清洗结果与清洗过程发生的错误。

    ```python
    kg_parse_results, kg_parse_err_msg = parse_knowledge_graph(input_log_list, custom_entity)
    ```

3. 调用诊断接口

    将清洗得到的 `kg_parse_results` 作为诊断接口的输入，进行故障事件诊断。

    该接口返回故障事件诊断结果与诊断过程中发生的错误。

    ```python
    result, err_msg_list = diag_knowledge_graph(kg_parse_results)
    ```

## 入参和返回值说明

故障事件清洗入参和返回值，请阅读 [SDK 接口参考](../06_api/09_sdk_api.md)中的 [parse_knowledge_graph](../06_api/09_sdk_api.md#parse_knowledge_graph) 接口定义。

故障事件诊断入参和返回值，请阅读 [SDK 接口参考](../06_api/09_sdk_api.md)中的 [diag_knowledge_graph](../06_api/09_sdk_api.md#diag_knowledge_graph) 接口定义。

## 参考

- 完整 SDK 接口说明请参考 [SDK 接口参考](../06_api/09_sdk_api.md)。
- 命令行方式请参考 [故障诊断](./04_fault_diagnosis.md)。
