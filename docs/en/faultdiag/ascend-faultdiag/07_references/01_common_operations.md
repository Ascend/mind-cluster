# Common Operations

## Querying NPU Information

Query the NPU driver and firmware versions:

```shell
npu-smi info
```

Example output:

```text
+--------------------------------------------------------------------------------------------------------+
| npu-smi xxx                                             Version: xxx                                   |
+-------------------------------+-----------------+------------------------------------------------------+
| NPU     Name                  | Health          | Power(W)     Temp(C)           Hugepages-Usage(page) |
| Chip    Device                | Bus-Id          | AICore(%)    Memory-Usage(MB)                        |
+===============================+=================+======================================================+
| 0       xxx                   | OK              | 10.4          66               0    / 0              |
| 0       0                     | NA              | 0            957   / 10810                           |
+===============================+=================+======================================================+
```

## Environment Variable Description

| Environment Variable                  | Description                                                                                          |
|-----------------------|------------------------------------------------------------------------------------------------------|
| `ASCEND_FD_HOME_PATH` | Storage path for runtime logs, operation logs, and configuration files. The default is `$HOME/.ascend_faultdiag`. Setting it to the `/tmp` directory is not supported. |
