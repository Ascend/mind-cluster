# API Reference

These tables are arranged in the default order of function parameters.

## tft_init_controller

**Function**

Initializes the MindIO TFT Controller module.

**Format**

```python
mindio_ttp.framework_ttp.tft_init_controller(rank: int, world_size: int, enable_local_copy: bool, enable_arf=False, enable_zit=False)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|rank|Mandatory|Rank ID of the NPU on which a training job is being executed.|int, [-1, world_size). When MindCluster starts the Controller in the Torch Agent process, the rank value is -1.|
|world_size|Mandatory|Number of NPUs that participate in training jobs in a cluster.|int, [1, 100000].|
|enable_local_copy|Mandatory|Whether to enable local copy. The optimizer needs to be backed up before it is updated.|<ul><li>False: disabled</li><li>True: enabled</li></ul>|
|enable_arf|Optional|Whether to enable MindIO ARF.|<ul><li>False: disabled</li><li>True: enabled</li></ul>The default value is False.|
|enable_zit|Optional|Whether to enable MindIO ZIT.|<ul><li>False: disabled</li><li>True: enabled</li></ul>The default value is False.|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_start_controller

**Function**

Starts the MindIO TFT Controller module service after the Controller module is successfully initialized.

**Format**

```python
mindio_ttp.framework_ttp.tft_start_controller(bind_ip: str, port: int, enable_tls=True, tls_info='')
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|bind_ip|Mandatory|IP address or domain name of the node where Controller resides.|An IPv4 address that complies with the IP address specifications is required. It must be in the IP address range of cluster nodes. An all-zero IP address is not allowed. Domain names are supported.|
|port|Mandatory|Listening Parameters number of Controller.|[1024, 65535]|
|enable_tls|Optional|Whether to enable TLS encrypted transmission.|<ul><li>False: disabled</li><li>True: enabled</li></ul>The default value is True.|
|tls_info|Optional|TLS certificate configuration.|By default, this parameter is left empty. When TLS authentication is enabled, you need to configure the certificate information. The related fields must be organized in key-value pairs. For details about the configuration, see section [Importing LS Certificates] (./04_security_management_and_hardening.md#Import-tls-certificates).|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_destroy_controller

**Function**

Disables the MindIO TFT Controller service after the training is complete.

**Format**

```python
mindio_ttp.framework_ttp.tft_destroy_controller()
```

**Parameters**

None

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_init_processor

**Function**

Initializes the MindIO TFT Processor module.

**Format**

```python
mindio_ttp.framework_ttp.tft_init_processor(rank: int, world_size: int, enable_local_copy: bool, enable_tls=True, tls_info='', enable_uce=True, enable_arf=False, enable_zit=False)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|rank|Mandatory|Rank ID of the NPU on which a training job is being executed.|int, [0, world_size).|
|world_size|Mandatory|Number of NPUs that participate in training jobs in a cluster.|int, [1, 100000].|
|enable_local_copy|Mandatory|Whether to enable local copy.|<ul><li>False: disabled</li><li>True: enabled</li></ul>|
|enable_tls|Optional|Whether to enable TLS encrypted transmission.|<ul><li>False: disabled</li><li>True: enabled</li></ul>The default value is True.|
|tls_info|Optional|TLS certificate configuration.|By default, this parameter is left empty. When TLS authentication is enabled, you need to configure the certificate information. The related fields must be organized in key-value pairs. For details about the configuration, see section [Importing LS Certificates] (./04_security_management_and_hardening.md#Import-tls-certificates).|
|enable_uce|Optional|Whether to enable MindIO UCE.|<ul><li>False: disabled</li><li>True: enabled</li></ul>The default value is True.|
|enable_arf|Optional|Whether to enable MindIO ARF.|<ul><li>False: disabled</li><li>True: enabled</li></ul>The default value is False.|
|enable_zit|Optional|Whether to enable MindIO ZIT.|<ul><li>False: disabled</li><li>True: enabled</li></ul>The default value is False.|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_start_processor

**Function**

Starts the MindIO TFT Processor service after the Processor module is successfully initialized.

**Format**

```python
mindio_ttp.framework_ttp.tft_start_processor(master_ip: str, port: int, local_ip='')
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|master_ip|Mandatory|IP address or domain name of the node where Controller resides.|An IPv4 address that complies with the IP address specifications is required. It must be in the IP address range of cluster nodes. An all-zero IP address is not allowed. Domain names are supported.|
|port|Mandatory|Listening Parameters number of Controller.|[1024, 65535]|
|local_ip|Optional|Service IP address or domain name of the node where Processor is located in Kubernetes.|An IPv4 address that complies with the IP address specifications is required. It must be in the IP address range of cluster nodes. An all-zero IP address is not allowed. Domain names are supported.|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_destroy_processor

**Function**

Disables the MindIO TFT processor service after training is complete.

**Format**

```python
mindio_ttp.framework_ttp.tft_destroy_processor()
```

**Parameters**

None

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_start_updating_os

**Function**

Updates the optimizer state to **Updating** before the update process begins.

**Format**

```python
mindio_ttp.framework_ttp.tft_start_updating_os(backup_step: int)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|backup_step|Mandatory|Backup step.|-1 or a natural number. The value range is [-1, 9223372036854775807). <ul><li>–1: backup step is not used. </li><li>Natural number: step corresponding to the backup optimizer state data before the optimizer is updated.</li></ul>|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_start_copy_os

**Function**

Instructs Processor to start copying the optimizer states.

**Format**

```python
mindio_ttp.framework_ttp.tft_start_copy_os()
```

**Parameters**

None

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_end_updating_os

**Function**

Updates the optimizer state to **Updated** after the update process ends.

**Format**

```python
mindio_ttp.framework_ttp.tft_end_updating_os(step: int)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|step|Mandatory|Current step.|Positive integer within the range of [1, 9223372036854775807).|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_set_optimizer_replica

**Function**

Sets the replica relationship of the optimizer state corresponding to a rank.

**Format**

```python
mindio_ttp.framework_ttp.tft_set_optimizer_replica(rank: int, replica_info: list)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|rank|Mandatory|Rank ID of the NPU on which a training job is being executed.|int, [0, 100000).|
|replica_info|Mandatory|List of replica relationships. Each element is a dictionary. The dictionary is arranged in the sequence of ATTENTION (0) and MOE (1).|[<br>{<br>"rank_list":list,    # List of ranks for a corresponding replica group. In PyTorch scenarios, this is the DP group rank list. In MindSpore scenarios, this is the list of all replica ranks corresponding to the current NPU<br>"replica_cnt":int,   # Number of replicas. In the PyTorch scenario, the value is the number of replicas. In the MindSpore scenario, the value is the length of the rank list.<br>"replica_shift":int,  # Valid in the PyTorch scenario.<br>},<br>]|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_exception_handler

**Function**

Decorator, which decorates the train method of MindSpeed-LLM, captures training state exceptions, and reports and processes the exceptions. This API is for reference only for other training frameworks.

**Format**

```python
mindio_ttp.framework_ttp.tft_exception_handler(func: Callable)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|func|Mandatory|The function is used as a parameter.|Train method of the framework.|

**Return Value**

**func** returned by the decorator.

## tft_set_step_args

**Function**

Parameter set of the training framework.

> [!NOTE]NOTE
> For MindSpeed-LLM, the setting function has been adapted by MindIO TFT, and this API does not need to be called.

**Format**

```python
mindio_ttp.framework_ttp.tft_set_step_args(args)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|args|Mandatory|Parameter set to be saved for the training framework. MindIO TFT calls the registered callback function in the stop, clean, repair, or rollback phase, and returns the parameter set. The framework completes the corresponding function based on the parameter set.|The parameter set is determined by the training framework. MindIO TFT does not access or modify the parameter set. In the stop, clean, repair, or rollback phase, the registered service callback is called to return the parameter set. The service callback is responsible for verifying the value range.|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_register_rename_handler

**Function**

Registers the rename callback function in the framework.

> [!NOTE]NOTE
> For the MindSpeed-LLM training framework, the callback function has been adapted by MindIO TFT. For other frameworks, you need to ensure the security of the callback function.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_rename_handler(func: Callable, ctx = None)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|func|Mandatory|rename function, which is used to rename the successfully saved dying gasp checkpoint. following the same naming rule as that of the inherent framework checkpoint.|Callback function, which cannot be empty. For details about the input parameters, see [Table 1](#table_tft_06) and [Table 2](#table_tft_07). This callback function does not return any value. If the execution fails, an exception is thrown.|
|ctx|Optional|Callback function context.|This parameter is left empty by default.|

**Table 1** <a id="table_tft_06"></a> MindSpore callback function parameters

|Parameter|Description|Value|
|--|--|--|
|step|Step for dumping optimizer data.|Positive integer|
|ctx|Callback function context|Determined by the registration party.|

**Table 2** <a id="table_tft_07"></a> Parameters of the non-MindSpore callback function

|Parameter|Description|Value|
|--|--|--|
|step|Step for dumping optimizer data.|Positive integer|
|args|Parameter set by **tft_set_step_args**.|Determined by the registration party.|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_register_save_ckpt_handler

**Function**

Registers the dump callback function in the framework.

> [!NOTE]NOTE
> For the MindSpeed-LLM training framework, the callback function has been adapted by MindIO TFT. For other frameworks, you need to ensure the security of the callback function.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_save_ckpt_handler(func: Callable, ctx = None)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|func|Mandatory|Dying gasp checkpoint saving function.|Callback function, which is not empty. For details about the input parameters o see [Table 1](#table_tft_08). This callback function has no return value. If the execution fails, an exception is thrown.|
|ctx|Optional|Callback function context.|This parameter is left empty by default.|

**Table 1<a id="table_tft_08"></a>** Callback function parameters

|Parameter|Description|Value|
|--|--|--|
|step|Step for dumping optimizer data.|Positive integer|
|save_info|Rank list generated when different optimizers participate in saving the dying gasp checkpoint. Each element is a dictionary. The dictionary is arranged in the sequence of ATTENTION (0) and MOE (1).|[<br>{<br>"type": int,   # Optimizer type<br>"ranks": list, # Rank list of the optimizers that participate in saving the dying gasp<br>},<br>]|
|args|Parameter set by **tft_set_step_args**.|Determined by the registration party.|
|ctx|Callback function context|Determined by the registration party.|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_register_exit_handler

**Function**

Registers a user-defined exit method with MindIO TFT.

> [!NOTE]NOTE
> Currently, the registration and exit callback function is provided only for the MindSpore framework. You need to ensure the security of the callback function. For other frameworks, MindIO TFT is responsible for the exit.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_exit_handler(func: Callable, ctx = None)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|func|Mandatory|Callback function that is exited.|Callback function, which cannot be empty. For details about the input parameter requirements, see [Table 1](#table_tft_09). The callback function does not return any value, and an exception is thrown if the execution fails.|
|ctx|Optional|Callback function context.|This parameter is left empty by default.|

**Table 1<a id="table_tft_09"></a>** Callback function parameters

|Parameter|Description|Value|
|--|--|--|
|ctx|Callback function context|Determined by the registration party.|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_register_stop_handler

**Function**

Registers the callback function for stopping training during recovery.

> [!NOTE]NOTE
> For the MindSpeed-LLM training framework, the callback function has been adapted by MindIO TFT. For other frameworks, you need to ensure the security of the callback function.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_stop_handler(func: Callable, ctx = None)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|func|Mandatory|Callback function for stopping training. After training is stopped, it throws the FORCE STOP exception and hands over the control of the main training thread to the decorator.|Callback function, which cannot be empty. For details about the input parameter requirements, see [Table 1](#table_tft_19). The callback function does not return any value, and an exception is thrown if the execution fails.|
|ctx|Optional|Callback function context.|This parameter is left empty by default.|

**Table 1<a id="table_tft_19"></a>** Callback function parameters

|Parameter|Description|Value|
|--|--|--|
|args|Parameter set by **tft_set_step_args**.|Determined by the registration party.|
|ctx|Callback function context|Determined by the registration party.|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_register_clean_handler

**Function**

Registers the callback function for clearing residual operator executions during recovery.

> [!NOTE]NOTE
> For the MindSpeed-LLM training framework, the callback function has been adapted by MindIO TFT. For other frameworks, you need to ensure the security of the callback function.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_clean_handler(func: Callable, ctx = None)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|func|Mandatory|Callback function for clearing residual operator executions and underlying faults.|Callback function, which cannot be empty. For details about the input parameters, see [Table 1](#table_tft_10). Return values: <ul><li>0: Success. </li><li>1: Failure. </li><li>2: UCE scenario, where the model optimizer does not need to be rebuilt.</li></ul>|
|ctx|Optional|Callback function context.|This parameter is left empty by default.|

**Table 1<a id="table_tft_10"></a>** Callback function parameters

|Parameter|Description|Value|
|--|--|--|
|is_uce_error|Indicates whether a UCE occurs on the NPU.|<ul><li>False: No UCE occurs. </li><li>True: A UCE fault occurs.</li></ul>|
|args|Parameter set by **tft_set_step_args**.|Determined by the registration party.|
|ctx|Callback function context|Determined by the registration party.|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_register_rebuild_group_handler

**Function**

Registers the callback function for re-establishing MindIO ARF groups.

> [!NOTE]NOTE
> For the MindSpeed-LLM training framework, the callback function has been adapted by MindIO TFT. For other frameworks, you need to ensure the security of the callback function.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_rebuild_group_handler(func: Callable, ctx = None)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|func|Mandatory|Callback function for MindIO ARF to re-establish communication groups. This function is used to clear the old communication groups and re-establish new communication groups for normal nodes and restarted nodes. The default timeout interval for executing the callback function is 180 seconds. If the execution times out, the process fails to be executed. You can use the environment variable **TTP_NORMAL_ACTION_TIME_LIMIT** to set the timeout interval.|Callback function, which cannot be empty. For details about the input parameter requirements, see [Table 1](#table_tft_11). The callback function does not return any value, and an exception is thrown if the execution fails.|
|ctx|Optional|Callback function context.|This parameter is left empty by default.|

**Table 1<a id="table_tft_11"></a>** Callback function parameters

|Parameter|Description|Value|
|--|--|--|
|fault_ranks|A collection of faulty NPUs.|list.|
|args|Parameter set by **tft_set_step_args**.|Determined by the registration party.|
|ctx|Callback function context|Determined by the registration party.|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_register_repair_handler

**Function**

Registers the repair callback function.

> [!NOTE]NOTE
>
> - For the MindSpeed-LLM training framework, the callback function has been adapted by MindIO TFT. For other frameworks, you need to ensure the security of the callback function.
> - MindIO TFT has rebuilt and overwritten the variables in the model optimizer in the callback function. You need to similarly rebuild and overwrite any other variables that are involved in calculations and customized within the framework, and do so in the repair function.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_repair_handler(func: Callable, ctx = None)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|func|Mandatory|Callback function of repair, which is used to repair data such as the optimizer data. The default timeout interval for executing the callback function is 180 seconds. If the execution times out, the process fails to be executed. You can use the environment variable **TTP_NORMAL_ACTION_TIME_LIMIT** to set the timeout interval.|Callback function, which cannot be empty. For details about the input parameters, see [Table 1](#table_tft_12). This callback function does not return any value. If the execution fails, an exception is thrown.|
|ctx|Optional|Callback function context.|This parameter is left empty by default.|

**Table 1<a id="table_tft_12"></a>** Callback function parameters

|Parameter|Description|Value|
|--|--|--|
|step|Step corresponding to the repair.|Positive integer.|
|need_rebuild|-|Whether the model and optimizer need to be rebuilt.|<ul><li>False: No rebuilding is required. </li><li>True: Rebuilding is required.</li></ul>|
|error_ranks|List of faulty NPUs to be repaired.|list.|
|repair_info|Repair policy dictionary. The optimizer type follows the relationship of ATTENTION (0) and MOE (1).|{<br>"type": int,   # Optimizer type<br>"repair_type": Enum,   # See [RepairType](#repairtype).<br>"src": list,   # List of source ranks for optimizer repair data<br>"dst": list, # List of destination ranks for optimizer repair data<br>"rank_list": list, # List of ranks required for communication group repair<br>}|
|args|Parameter set by **tft_set_step_args**.|Determined by the registration party.|
|ctx|Callback function context|Determined by the registration party.|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_register_rollback_handler

**Function**

Register the callback function of rollback.

> [!NOTE]NOTE
> For the MindSpeed-LLM training framework, the callback function has been adapted by MindIO TFT. For other frameworks, you need to ensure the security of the callback function.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_rollback_handler(func: Callable, ctx = None)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|func|Mandatory|Rollback callback function, which is used to perform reset operations such as dataset rollback. The default timeout interval for executing the callback function is 180 seconds. If the execution times out, the process fails to be executed. You can use the environment variable **TTP_NORMAL_ACTION_TIME_LIMIT** to set the timeout interval.|Callback function, which cannot be empty. For details about the input parameters, see [Table 1](#table_tft_13). This callback function does not return any value. If the execution fails, an exception is thrown.|
|ctx|Optional|Callback function context.|This parameter is left empty by default.|

**Table 1<a id="table_tft_13"></a>** Callback function parameters

|Parameter|Description|Value|
|--|--|--|
|step|Step to which it is rolled back.|Positive integer.|
|args|Parameter set by **tft_set_step_args**.|Determined by the registration party.|
|ctx|Callback function context|Determined by the registration party.|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_register_stream_sync_handler

**Function**

Registers a synchronous callback.

> [!NOTE]NOTE
> For the MindSpeed-LLM training framework, the callback function has been adapted by MindIO TFT. For other frameworks, you need to ensure the security of the callback function.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_stream_sync_handler(func: Callable, ctx=None)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|func|Mandatory|Synchronization callback function to synchronize the operation after training is paused, ensuring no residual operators remain in the operator queue.|This callback function cannot be empty. The callback function has no parameter or return value. If the callback function fails to be executed, an exception is thrown.|
|ctx|Optional|Callback function context.|Determined by the registration party.|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_register_zit_upgrade_rollback_handler

**Function**

Registers the upgrade rollback callback function with Processor.

> [!NOTE]NOTE
> For MindSpeed-LLM, the callback function has been adapted. For other frameworks, you need to ensure the security of the callback function.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_zit_upgrade_rollback_handler(func: Callable, ctx = None)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|func|Mandatory|Rollback callback function, which is used to perform reset operations such as dataset rollback. The default timeout interval for executing the callback function is 180 seconds. If the execution times out, the process fails to be executed. You can use the environment variable **TTP_NORMAL_ACTION_TIME_LIMIT** to set the timeout interval.|The callback function cannot be left empty and has no return value. If the callback function fails to be executed, an exception is thrown.|
|ctx|Optional|Callback function context.|This parameter is left empty by default.|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_register_zit_upgrade_repair_handler

**Function**

Registers the upgrade repair callback function with Processor.

> [!NOTE]NOTE
> For MindSpeed-LLM, the callback function has been adapted. For other frameworks, you need to ensure the security of the callback function.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_zit_upgrade_repair_handler(func: Callable, ctx = None)
```

**Parameters <a id="section34575883518"></a>**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|func|Mandatory|Callback function of repair, which is used to repair data such as the optimizer data. The default timeout interval for executing the callback function is 180 seconds. If the execution times out, the process fails to be executed. You can use the environment variable **TTP_NORMAL_ACTION_TIME_LIMIT** to set the timeout interval.|The callback function cannot be left empty and has no return value. If the callback function fails to be executed, an exception is thrown.|
|ctx|Optional|Callback function context.|This parameter is left empty by default.|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_register_zit_upgrade_rebuild_handler

**Function**
Registers the callback function for rebuilding communication groups during upgrade with Processor.

> [!NOTE]NOTE
> For MindSpeed-LLM, the callback function has been adapted. For other frameworks, you need to ensure the security of the callback function.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_zit_upgrade_rebuild_handler(func: Callable, ctx = None)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|func|Mandatory|Callback function of rebuild, which is used to rebuild communication groups during upgrade. The default timeout interval for executing the callback function is 180 seconds. If the execution times out, the process fails to be executed. You can use the environment variable **TTP_NORMAL_ACTION_TIME_LIMIT** to set the timeout interval.|The callback function cannot be left empty and has no return value. If the callback function fails to be executed, an exception is thrown.|
|ctx|Optional|Callback function context.|This parameter is left empty by default.|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_register_zit_downgrade_rebuild_handler

**Function**

Registers the callback function for repair rebuilding during downgrade with Processor.

> [!NOTE]NOTE
> For MindSpeed-LLM, the callback function has been adapted. For other frameworks, you need to ensure the security of the callback function.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_zit_downgrade_rebuild_handler(func: Callable, ctx = None)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|func|Mandatory|Callback function of rebuild, which is used to rebuild repairs during downgrade. The default timeout interval for executing the callback function is 180 seconds. If the execution times out, the process fails to be executed. You can use the environment variable **TTP_NORMAL_ACTION_TIME_LIMIT** to set the timeout interval.|The callback function cannot be left empty and has no return value. If the callback function fails to be executed, an exception is thrown.|
|ctx|Optional|Callback function context.|This parameter is left empty by default.|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_register_exception_handler

**Function**

Registers the exception handler.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_exception_handler(fault_pattern: str, fault_type: str, fault_handle: Callable)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|fault_pattern|Mandatory|Exception keyword. It is used to exactly match the exception type.|Keyword string in the exception information.|
|fault_type|Mandatory|Exception type. When an exception is captured, the exception information is reported to MindIO together with the return value of fault_handle.|String. The value range is as follows (for details, see [ReportState](#reportstate)): <ul><li>RS_NORMAL</li><li>RS_RETRY</li><li>RS_UCE</li><li>RS_UCE_CORRUPTED</li><li>RS_HCCL_FAILED</li><li>RS_INIT_FINISH</li><li>RS_PREREPAIR_FINISH</li><li>RS_STEP_FINISH</li><li>RS_UNKNOWN</li></ul>|
|fault_handle|Mandatory|Exception handling method. It is used to receive the exception string and return a string. This return value is used together with fault_type to report exception information.|Executable method which receives an exception string and returns a string.|

**Return Value**

No return value.

## tft_report_error

**Function**

Reports an error type.

**Format**

```python
mindio_ttp.framework_ttp.tft_report_error(error_type: ReportState)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|error_type|Mandatory|Error type, which is used to determine the subsequent rectification process.|Actual error type. See [ReportState](#reportstate).|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_wait_next_action

**Function**

During the repair, the main training thread calls this API in the decorator to wait for the secondary thread to complete service data repair.

> [!NOTE]NOTE
> This API is a blocking API. It will be blocked until next action is obtained.

**Format**

```python
mindio_ttp.framework_ttp.tft_wait_next_action()
```

**Parameters**

None

**Return Value**

- **0**: success
- **1**: failure

## tft_get_repair_step

**Function**

Queries the step value at the position where repair is performed.

**Format**

```python
mindio_ttp.framework_ttp.tft_get_repair_step()
```

**Parameters**

None

**Return Value**

Step used for repair. The value **0** indicates an invalid value.

## tft_get_repair_type

**Function**

This API is called by MindSpore to query the repair type in the callback of the stop, clean, or repair phase.

**Format**

```python
mindio_ttp.framework_ttp.tft_get_repair_type()
```

**Parameters**

None

**Return Value**

The value is of the string type.

- **retry**: UCE repair
- **recover**: ARF repair
- **dump**: dying gasp
- **unknown**: repair type not found

## tft_is_reboot_node

**Function**

In the MindIO ARF process, this interface is used to determine whether the current process is a node that is restarted after a fault occurs. This interface can be called only once immediately after the tft_start_processor interface is successfully called.

**Format**

```python
mindio_ttp.framework_ttp.tft_is_reboot_node()
```

**Parameters**

None

**Return Value**

Boolean value, indicating whether the process is the restarted node after a fault occurs.

## tft_get_reboot_type

**Function**

This interface is provided for MindSpore to obtain the node restart scenario type from MindIO TTP after a node is restarted due to a fault. After the process is started, this interface can be called only once.

**Format**

```python
mindio_ttp.framework_ttp.tft_get_reboot_type()
```

**Parameters**

None

**Return Value**

String

- **arf**: process rescheduling
- **hot switch**: hot switching

## tft_reset_limit_step

**Function**

Updates the `prelock` flag in the Processor to true and reset `limitStep_` to the maximum value.

**Format**

```python
mindio_ttp.framework_ttp.tft_reset_limit_step()
```

**Parameters**

None

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_set_dp_group_info

**Function**

Registers the DP group information with Processor.

**Format**

```python
mindio_ttp.controller_ttp.tft_set_dp_group_info(rank: int, dp_rank_list: list)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|rank|Mandatory|Current rank.|≥ 0|
|dp_rank_list|Mandatory|DP group information.|The value cannot be left empty.|

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_report_load_ckpt_step

**Function**

Reports the number of steps loaded from the checkpoint during recovery via periodic checkpoints.

**Format**

```python
mindio_ttp.framework_ttp.tft_report_load_ckpt_step(step: int)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|step|Mandatory|Number of steps loaded from the checkpoint.|0 or a positive integer|

**Return Values**

None

## tft_register_decrypt_handler

**Function**

Registers the function for private key password decryption if TLS encryption is enabled.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_decrypt_handler(decryptor: Callable)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|decryptor|Mandatory|User-defined function for decrypting private key password.|TLS encryption is configured using tft_start_controller and tft_init_processor. If the password is in ciphertext, the decryption function needs to be registered. For details about the configuration, see section [Importing LS Certificates](./04_security_management_and_hardening.md#importing-tls-certificates).|

**Parameters of the Callback Function**

|Parameter|Description|Value|
|--|--|--|
|cipherText|Private key password to be decrypted.|Determined by the registration party.|

The return value of the callback function is `plainText : str`, that is, the decrypted private key password.

**Return Value**

No return value. If an error occurs, an error log is recorded and an exception is thrown.

## tft_notify_controller_dump

**Function**

This API is called by MindCluster to instruct MindIO TFT to proactively stop training and exit training after dumping.

**Format**

```python
mindio_ttp.controller_ttp.tft_notify_controller_dump()
```

**Parameters**

None

**Return Value**

- **0**: success.
- **1**: failure.

## tft_notify_controller_stop_train

**Function**

This API is called by MindCluster to instruct MindIO TFT to proactively stop training and notify MindIO TFT of the information about the faulty rank.

**Format**

```python
mindio_ttp.controller_ttp.tft_notify_controller_stop_train(fault_ranks: dict, stop_type: str = "stop", timeout: int = None)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|fault_ranks|Mandatory|Information about the faulty NPU.|<int key, int errorType>Dictionary: <ul><li>key indicates the ID of the faulty rank. </li><li>errorType indicates the fault type. </li><ul><li>0: UCE. </li><li>1: non-UCE. </li></ul></ul>|
|stop_type|Optional|Mode of stopping training.|The value is a character string. The following two modes are supported: <ul><li>"stop": stops training in taskabort mode. </li><li>"pause": pauses training in non-taskabort mode.</li></ul>|
|timeout|Optional|Timeout interval for waiting for the next notification from MindCluster after training is paused, in seconds.|The value is a non-negative integer.|

**Return Value**

- **0**: API call succeeded.
- **1**: API call failed.

## tft_notify_controller_on_global_rank

**Function**

This API is called by MindCluster to notify MindIO TFT of the global information about the faulty rank.

**Format**

```python
mindio_ttp.controller_ttp.tft_notify_controller_on_global_rank(fault_ranks: dict,time:int=1)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|fault_ranks|Mandatory|Information about the faulty NPU.|<int key, int errorType> dictionary: <ul><li>key indicates the ID of the faulty rank. </li><li>errorType indicates the fault type. </li><ul><li>0: UCE fault. </li><li>1: non-UCE fault.</li></ul></ul>|
|time|Optional|Maximum time for interacting with the MindCluster repair policy, which is determined based on the environment variable.|The value is an integer in the range [1, 3600], with a default of **1** (unit: seconds).|

**Return Value**

- **0**: API call succeeded.
- **1**: API call failed.

## tft_notify_controller_prepare_action

**Function**

This API is called by MindCluster to notify the repair policy to be executed by MindIO TFT.

> [!NOTE]NOTE
> The repair policy must be within the range of optional repair policies negotiated by MindCluster and MindIO TFT.

**Format**

```python
mindio_ttp.controller_ttp.tft_notify_controller_prepare_action(action: str, fault_ranks: dict = None)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|action|Mandatory|Notifies MindIO TFT of hot switching.|String. The supported repair policies are as follows: <ul><li>hot switch</li><li>stop switch</li></ul>|
|fault_ranks|Optional|Information about the faulty NPU.|Dictionary. The key indicates the rank ID, ranging from 1 to 100000. The value indicates the error type, ranging from 0 to 2.|

**Return Value**

- **0**: API call succeeded.
- **1**: API call failed.

## tft_notify_controller_change_strategy

**Function**

This API is called by MindCluster to notify the repair policy to be executed by MindIO TFT.

> [!NOTE]NOTE
> The repair policy must be within the range of optional repair policies negotiated by MindCluster and MindIO TFT.

**Format**

```python
mindio_ttp.controller_ttp.tft_notify_controller_change_strategy(strategy: str, params: str = "")
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|strategy|Mandatory|Notifies the MindIO TFT of the repair policy.|String. The supported repair policies are as follows: <ul><li>retry</li><li>downgrade </li><li>upgrade</li><li>recover</li><li>dump</li><li>continue</li><li>migration</li><li>exit</li></ul>|
|params|<ul><li>Mandatory for downgrade training.</li><li>Optional for other scenarios.</li></ul>|Downgrade training parameter.|The value is of string type and the default value is an empty string ("").|

**Return Value**

- **0**: API call succeeded.
- **1**: API call failed.

## tft_register_mindx_callback

**Function**

This API is called by MindCluster to register the repair process callback function with MindIO TFT.

**Format**

```python
mindio_ttp.controller_ttp.tft_register_mindx_callback(action: str, func: Callable)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|action|Mandatory|Name of the action to be registered by the callback function.|String. The supported actions are: <ul><li>report_fault_ranks</li> <li>report_stop_complete</li><li>report_strategies</li><li>report_result</li></ul>|
|func|Mandatory|Function to be registered.|Callback function, which is not empty. For details about the input parameters, see [Table 1](#table_tft_14) to [Table 4](#table_tft_17).|

**Table 1** <a id="table_tft_14"></a> Parameters of the callback function when action = report_fault_ranks

|Parameter|Description|Value|
|--|--|--|
|error_rank_dict|Information about the faulty NPU.|<int key, int errorType> dictionary: <ul><li>key indicates the ID of the faulty rank. </li><li>errorType indicates the fault type. </li><ul><li>0: UCE. </li><li>1: non-UCE.</li></ul></ul>|

**Table 2** Parameters of the callback function when action = report_stop_complete

|Parameter|Description|Value|
|--|--|--|
|code|Action execution result.|<ul><li>0: success. </li><li>400: common error. </li><li>401: The MindCluster task ID does not exist. </li><li>402: model error. </li><li>403: incorrect sequence. </li><li>404: Not all Processors are ready.</li></ul>|
|msg|Message indicating whether the training stops.|String|
|error_rank_dict|Information about the faulty NPU.|<int key, int errorType> dictionary: <ul><li>key indicates the ID of the faulty rank. </li><li>errorType indicates the fault type. </li><ul><li>0: UCE. </li><li>1: non-UCE.</li></ul></ul>|

**Table 3<a id="table_tft_16"></a>** Parameters of the callback function when action = report_strategies

|Parameter|Description|Value|
|--|--|--|
|error_rank_dict|Information about the faulty NPU.|<int key, int errorType> dictionary: <ul><li>key indicates the ID of the faulty rank. </li><li>errorType indicates the fault type. </li><ul><li>0: UCE. </li><li>1: non-UCE.</li></ul></ul>|
|strategy_list|List of repair policies supported by MindIO TFT based on the available replica information.|List. The supported repair policies (string) are: <ul><li>retry: UCE repair </li><li>recover: ARF repair </li><li>dump: dying gasp </li><li>exit: exit.</li></ul>|

**Table 4<a id="table_tft_17"></a>** Parameters of the callback function when action = report_result

|Parameter|Description|Value|
|--|--|--|
|code|Action execution result.|<ul><li>0: The repair is successful. </li><li>405: The retry repair fails. The recover, dump, and exit repair policies can be performed. </li><li>406: The repair fails. The dump or exit repair policy can be performed. </li><li>499: The recovery fails. Only the exit policy is supported.</li></ul>|
|msg|Message indicating repair success or failure.|str|
|error_rank_dict|Information about the faulty NPU.|<int key, int errorType> dictionary: <ul><li>key indicates the ID of the faulty rank. </li><li>errorType indicates the fault type. </li><ul><li>0: UCE. </li> <li>1: non-UCE.</li></ul></ul>|
|curr_strategy|Current repair policy.|String. The supported repair policy is strategy_list in Table 3.|

**Return Value**

- **0**: API call succeeded.
- **1**: API call failed.

## tft_query_high_availability_switch

**Function**

This API is called by MindCluster to check whether HA is enabled in real time.

**Format**

```python
mindio_ttp.controller_ttp.tft_query_high_availability_switch()
```

**Parameters**

None

**Return Value**

Boolean value, indicating whether high availability is enabled.

## tft_can_do_uce_repair

**Function**

This API is called by MindSpore to determine whether the optimizer data is polluted in the time dimension based on the time when an UCE is triggered by the L2 cache and the time before and after the optimizer is updated, and then return the result of whether the fault can be rectified.

> [!NOTE]NOTE
> This API determines optimizer data corruption solely by evaluating the intersection of time ranges, instead of memory addresses.

**Format**

```python
mindio_ttp.framework_ttp.tft_can_do_uce_repair(hbm_error_time: int, start_time: int = None, end_time: int = None)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|hbm_error_time|Mandatory|Time when the L2 Cache triggers a UCE.|int|
|start_time|Optional|Time obtained from the device before the optimizer is updated locally.|int|
|end_time|Optional|Time obtained from the device after the optimizer is updated locally.|int|

**Return Value**

Boolean value, which indicates whether fast recovery upon UCEs can be performed based on the time intersection.

## tft_set_update_start_time

**Function**

Sets the start time of optimizer update, which is used to determine whether optimizer data is polluted in the time dimension and return the result of whether the data can be restored.

**Format**

```python
mindio_ttp.utils.tft_set_update_start_time(start_time: int = None)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|start_time|Optional|Time obtained from the device before the optimizer is updated locally.|int|

**Return Values**

None

## tft_set_update_end_time

**Function**

Sets the end time of optimizer update, which is used to determine whether optimizer data is polluted in the time dimension and return the result of whether the data can be restored.

**Format**

```python
mindio_ttp.utils.tft_set_update_end_time(end_time: int = None)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|end_time|Optional|Time obtained from the device after the optimizer is updated locally.|int|

**Return Values**

None

## tft_pause_train

**Function**

Pauses training at a step.

**Format**

```python
mindio_ttp.framework_ttp.tft_pause_train(cur_step: int)
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|cur_step|Mandatory|Current step executed by the current training framework|0 or a positive integer|

**Return Values**

None

## OptimizerType

**Function**

Defines optimizer types.

**Format**

```python
mindio_ttp.framework_ttp.OptimizerType
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|OptimizerType|Mandatory|Optimizer type. <ul><li>ATTENTION: attention mechanism </li><li>MOE: MoE scenario.</li></ul>|<ul><li>ATTENTION: 0</li><li>MOE: 1</li></ul>|

**Return Values**

None

## Action

**Function**

Enumerates the action types after the main thread reports an exception.

**Format**

```python
mindio_ttp.framework_ttp.Action
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|Action|Mandatory|Action types after the main thread reports an exception. The options are as follows: <ul><li>RETRY: The training continues after the fault is rectified. </li><li>EXIT: The training exits.</li></ul>|<ul><li>RETRY: 0</li><li>EXIT: 1</li></ul>|

**Return Values**

None

## ReportState

**Function**

Enumerates the training states reported by the decorator.

**Format**

```python
mindio_ttp.framework_ttp.ReportState
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|ReportState|Mandatory|Types of the reported training status. The options are as follows: <ul><li>RS_NORMAL: normal status. </li><li>RS_UCE: UCE. </li><li>RS_UCE_CORRUPTED: multi-bit ECC in the on-chip memory. </li><li>RS_HCCL_FAILED: HCCL recalculation failure. </li><li>RS_UNKNOWN: other errors. </li><li>RS_INIT_FINISH: exception raised in the MindSpore framework by a newly launched ARF node after its training process finishes initialization. </li><li>RS_PREREPAIR_FINISH: exception thrown by a newly launched ARF node </li><li>RS_STEP_FINISH: exception thrown when the step-level pause finishes during hot switching.</li></ul>|<ul><li>RS_NORMAL.value: ttp_c2python_api.ReportState_RS_NORMAL. </li><li>RS_UCE.value: ttp_c2python_api.ReportState_RS_UCE. </li><li>RS_UCE_CORRUPTED: ttp_c2python_api.ReportState_RS_UCE_CORRUPTED. </li><li>RS_HCCL_FAILED.value: ttp_c2python_api.ReportState_RS_HCCL_FAILED. </li><li>RS_UNKNOWN.value: ttp_c2python_api.ReportState_RS_UNKNOWN. </li><li>RS_INIT_FINISH: ttp_c2python_api.ReportState_RS_INIT_FINISH. </li><li>RS_PREREPAIR_FINISH.value: ttp_c2python_api.ReportState_RS_PREREPAIR_FINISH. </li><li>RS_STEP_FINISH: ttp_c2python_api.ReportState_RS_STEP_FINISH.</li></ul>|

**Return Values**

None

## RepairType

**Function**

Defines repair types.

**Format**

```python
mindio_ttp.framework_ttp.RepairType
```

**Parameters**

|Parameter|Mandatory|Description|Value|
|--|--|--|--|
|RepairType|Mandatory|Repair type. <ul><li>RT_SEND: The backup rank sends data. </li><li>RT_UCE_HIGHLEVEL: The optimizer and model need to be rebuilt for the faulty rank. </li><li>RT_UCE_LOWLEVEL: The optimizer and model reconstruction are not required for the faulty rank. </li><li>RT_ROLLBACK: dataset rollback. </li><li>RT_RECV_REPAIR: rank started bo ARF to receive data. </li><li>RT_LOAD_CKPT: periodic checkpoint repair. </li><li>RT_LOAD_REBUILD: rebuilds periodic checkpoint repair for the model optimizer.</li></ul>|<ul><li>RT_SEND.value: ttp_c2python_api.RepairType_RT_SEND. </li><li>RT_UCE_HIGHLEVEL.value: ttp_c2python_api.RepairType_RT_UCE_HIGHLEVEL. </li><li>RT_UCE_LOWLEVEL.value: ttp_c2python_api.RepairType_RT_UCE_LOWLEVEL. </li><li>RT_ROLLBACK.value: ttp_c2python_api.RepairType_RT_ROLLBACK. </li><li>RT_RECV_REPAIR.value: ttp_c2python_api.RepairType_RT_RECV_REPAIR. </li><li>RT_LOAD_CKPT.value: ttp_c2python_api.RepairType_RT_LOAD_CKPT. </li><li>RT_LOAD_REBUILD.value: ttp_c2python_api.RepairType_RT_LOAD_REBUILD.</li></ul>|

**Return Values**

None
