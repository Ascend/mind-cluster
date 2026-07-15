# 业务日志清洗（SDK）

通过 ascend-fd 的 Python SDK 接口，可以对业务日志进行清洗处理，提取关键信息。

## 适用场景

- 用户自己的 Python 程序中集成日志清洗功能
- 需要对业务日志进行程序化处理

## 操作步骤

1. 导入 SDK：

    ```python
    from ascend_fd import parse_fault_type
    ```

2. 调用清洗接口：

    ```python
    result, err_msg_list = parse_fault_type(input_log_list)
    ```

## 入参和返回值说明

业务日志清洗入参和返回值，请阅读 `SDK 接口参考` 中的 [parse_fault_type](../06_api/09_sdk_api.md#parse_fault_type) 接口定义。
