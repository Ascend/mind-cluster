# 根因节点清洗及诊断（SDK）

通过 ascend-fd 的 Python SDK 接口，可以对日志进行根因节点清洗和诊断，识别引发故障的设备。

## 适用场景

- 用户在自己的 Python 程序中集成根因分析功能
- 需要对根因节点进行程序化诊断

## 操作步骤

建议使用流程：根因节点清洗 → 根因节点诊断

1. SDK 导入

    ```python
    from ascend_fd import parse_root_cluster, diag_root_cluster
    ```

2. 调用清洗接口

    该接口返回根因节点清洗结果与清洗过程发生的错误。

    ```python
    rc_parse_results, rc_parse_err_msg = parse_root_cluster(input_log_list)
    ```

3. 调用诊断接口

    将清洗得到的 `rc_parse_results` 作为诊断接口的输入，进行根因节点诊断。

    该接口返回根因节点诊断结果与诊断过程中发生的错误。

    ```python
    result, err_msg_list = diag_root_cluster(rc_parse_results)
    ```

## 入参和返回值说明

根因节点清洗入参和返回值，请阅读 [SDK 接口参考](../06_api/09_sdk_api.md)中的 [parse_root_cluster](../06_api/09_sdk_api.md#parse_root_cluster) 接口定义。

根因节点诊断入参和返回值，请阅读 [SDK 接口参考](../06_api/09_sdk_api.md)中的 [diag_root_cluster](../06_api/09_sdk_api.md#diag_root_cluster) 接口定义。

## 参考

- 完整 SDK 接口说明请参考 [SDK 接口参考](../06_api/09_sdk_api.md)。
- 命令行方式请参考 [故障诊断](./04_fault_diagnosis.md)。
