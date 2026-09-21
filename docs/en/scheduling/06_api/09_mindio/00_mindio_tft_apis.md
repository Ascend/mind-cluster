# MindIO TFT APIs

> All interface parameter tables and callback function parameter tables are arranged in the order of function parameters by default.

## tft\_init\_controller

**Function**

Initializes the MindIO TFT Controller module.

**Format**

```python
mindio_ttp.framework_ttp.tft_init_controller(rank: int, world_size: int, enable_local_copy: bool, enable_arf=False, enable_zit=False)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|rank|Mandatory|NPU card number on which the training task is currently executed.|int, [-1, world_size). When MindCluster starts the Controller in the Torch Agent process, the rank value is -1.|
|world_size|Mandatory|Number of cards in the entire cluster participating in the training task.|int, [1, 100000].|
|enable_local_copy|Mandatory|Whether to enable local copy. Before the optimizer update, a backup of the optimizer is made first.|<ul><li>False: disabled</li><li>True: enabled</li></ul>|
|enable_arf|Optional|MindIO ARF feature flag.|<ul><li>False: disabled</li><li>True: enabled</li></ul>Defaults to False.|
|enable_zit|Optional|MindIO ZIT feature flag.|<ul><li>False: disabled</li><li>True: enabled</li></ul>Defaults to False.|

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft_start_controller

**Function**

After the Controller module is successfully initialized, call this interface to start the MindIO TFT Controller module service.

**Format**

```python
mindio_ttp.framework_ttp.tft_start_controller(bind_ip: str, port: int, enable_tls=True, tls_info='')
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|bind_ip|Mandatory|IP address or domain name of the node where the Controller resides.|An IPv4 address that conforms to the address specification, located in the cluster node IP addresses, all zeros prohibited, domain name supported.|
|port|Mandatory|Listening port number of the Controller.|[1024, 65535]|
|enable_tls|Optional|TLS encrypted transmission flag.|<ul><li>False: disabled</li><li>True: enabled</li></ul>Defaults to True.|
|tls_info|Optional|TLS certificate configuration.|Defaults to empty. When TLS authentication is enabled, certificate information must be configured, and specific fields should be organized as key-value pairs. For specific configuration guidance, see [Importing TLS Certificates](../../07_references/00_fault_recovery_acceleration/04_security_management_and_hardening.md#importing-tls-certificates).|

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft\_destroy\_controller

**Function**

After training is complete, call this interface to shut down the MindIO TFT Controller service.

**Format**

```python
mindio_ttp.framework_ttp.tft_destroy_controller()
```

**Parameters**

None

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft\_init\_processor

**Function**

Initializes the MindIO TFT Processor module.

**Format**

```python
mindio_ttp.framework_ttp.tft_init_processor(rank: int, world_size: int, enable_local_copy: bool, enable_tls=True, tls_info='', enable_uce=True, enable_arf=False, enable_zit=False)
```

**Parameters**

|Parameter|Mandatory|Description|Value Requirements|
|--|--|--|--|
|rank|Mandatory|NPU card number currently executing the training task.|int, [0, world_size).|
|world_size|Mandatory|Number of cluster cards participating in the training task.|int, [1, 100000].|
|enable_local_copy|Mandatory|Whether to enable local copy.|<ul><li>False: disabled</li><li>True: enabled</li></ul>|
|enable_tls|Optional|TLS encrypted transmission switch.|<ul><li>False: disabled</li><li>True: enabled</li></ul>Defaults to True.|
|tls_info|Optional|TLS certificate configuration.|Defaults to empty. When TLS authentication is enabled, certificate information must be configured, and specific fields should be organized as key-value pairs. For specific configuration guidance, see [Importing TLS Certificates](../../07_references/00_fault_recovery_acceleration/04_security_management_and_hardening.md#importing-tls-certificates).|
|enable_uce|Optional|MindIO UCE feature flag.|<ul><li>False: disabled</li><li>True: enabled</li></ul>Defaults to True.|
|enable_arf|Optional|MindIO ARF feature flag.|<ul><li>False: disabled</li><li>True: enabled</li></ul>Defaults to False.|
|enable_zit|Optional|MindIO ZIT feature flag.|<ul><li>False: disabled</li><li>True: enabled</li></ul>Defaults to False.|

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft\_start\_processor

**Function**

After the Processor module is successfully initialized, call this interface to start the MindIO TFT Processor module service.

**Format**

```python
mindio_ttp.framework_ttp.tft_start_processor(master_ip: str, port: int, local_ip='')
```

**Parameters**

|Parameter|Mandatory|Description|Value Requirements|
|--|--|--|--|
|master_ip|Mandatory|IP address or domain name of the node where the Controller resides.|An IPv4 address that conforms to the IP address specification, located in the cluster node IP addresses, all zeros prohibited, domain name supported.|
|port|Mandatory|Listening port number of the Controller.|[1024, 65535]|
|local_ip|Optional|Service IP address or domain name of the node where the Processor resides in K8s.|An IPv4 address that conforms to the IP address specification, located in the cluster node IP addresses, all zeros prohibited, domain name supported.|

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft\_destroy\_processor

**Function**

After training is complete, call this interface to shut down the MindIO TFT Processor service.

**Format**

```python
mindio_ttp.framework_ttp.tft_destroy_processor()
```

**Parameters**

None

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft\_start\_updating\_os

**Function**

Before the optimizer state is updated, call this interface to update the optimizer state to Updating.

**Format**

```python
mindio_ttp.framework_ttp.tft_start_updating_os(backup_step: int)
```

**Parameters**

|Parameter|Mandatory or optional|Description|Value requirements|
|--|--|--|--|
|backup_step|Mandatory|Step to back up.|-1 or a natural number, in the range [-1, 9223372036854775807).<ul><li>-1: indicates that no backup step is used.</li><li>Natural number: the step corresponding to the optimizer state data backed up before the optimizer update.</li></ul>|

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft\_start\_copy\_os

**Function**

Notifies the Processor to start copying optimizer data.

**Format**

```python
mindio_ttp.framework_ttp.tft_start_copy_os()
```

**Parameters**

None

**Return Value**

No return value. If an error occurs, an ERROR log is printed and an exception is thrown.

## tft\_end\_updating\_os

**Function**

After the optimizer state update is complete, call this interface to update the optimizer state to Updated.

**Format**

```python
mindio_ttp.framework_ttp.tft_end_updating_os(step: int)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|step|Mandatory|Current step.|Positive integer, range [1, 9223372036854775807).|

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft\_set\_optimizer\_replica

**Function**

Sets the replica relationship of the optimizer state data corresponding to the rank.

**Format**

```python
mindio_ttp.framework_ttp.tft_set_optimizer_replica(rank: int, replica_info: list)
```

**Parameters**

|Name|Mandatory|Description|Value requirements|
|--|--|--|--|
|rank|Mandatory|NPU card number currently executing the training task.|int, [0, 100000).|
|replica_info|Mandatory|Replica relationship list, where each element is a dictionary arranged in the index order of ATTENTION (0) and MOE (1).|[<br>{<br>"rank_list":list,   # The corresponding set of replica relationship rank lists. In the PyTorch scenario, it is the DP group rank list; in the MindSpore scenario, it is the list of all replica cards corresponding to this card. <br>"replica_cnt":int,   # Number of replicas. In the PyTorch scenario, it is the number of replicas; in the MindSpore scenario, it is the length of rank_list. <br>"replica_shift":int,  # Valid in the PyTorch scenario.<br>},<br>]|

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft\_exception\_handler

**Function**

A decorator that decorates the train method of MindSpeed-LLM to capture training state exceptions and report and handle them. For other training frameworks, this interface only provides a reference functionality only.

**Format**

```python
mindio_ttp.framework_ttp.tft_exception_handler(func: Callable)
```

**Parameters**

|Name|Mandatory|Description|Value requirements|
|--|--|--|--|
|func|Yes|Function as a parameter.|The train method of the framework.|

**Return Value**

The `func` returned by the decorator.

## tft\_set\_step\_args

**Function**

The parameter set configured by the training framework.

> [!NOTE]
> For the MindSpeed-LLM training framework, the setting function has already been adapted by MindIO TFT and does not need to be called.

**Format**

```python
mindio_ttp.framework_ttp.tft_set_step_args(args)
```

**Parameters**

|Parameter|Mandatory or optional|Description|Value requirements|
|--|--|--|--|
|args|Mandatory|The parameter set that the training framework configures to be saved. When MindIO TFT invokes the registered callback functions during the stop/clean/repair/rollback phases, it passes the parameter set back, and the framework completes the corresponding function based on the parameter set.|Determined by the training framework. MindIO TFT neither accesses nor modifies this parameter set. During the stop/clean/repair/rollback phases, it invokes the registered business callback to pass it back, and the business callback is responsible for validating the value range.|

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft\_register\_rename\_handler

**Function**

Registers the framework-side rename callback function.

> [!NOTE]
> For the MindSpeed-LLM training framework, the callback function has already been adapted by MindIO TFT; for other frameworks, users need to ensure the safety of the callback function themselves.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_rename_handler(func: Callable, ctx = None)
```

**Parameters**

|Name|Type|Mandatory|Description|
|--|--|--|--|
|func|Mandatory|The rename function, which renames the successfully saved terminal checkpoint in accordance with the native framework's checkpoint naming rules.|Callback function, must not be empty. For input parameter requirements of the callback function, see [Table 1](#table_tft_06) and [Table 2](#table_tft_07). It is agreed that this callback function has no return value and throws an exception upon execution failure.|
|ctx|Optional|Callback function context.|Defaults to empty.|

**Table 1<a id="table_tft_06"></a>** MindSpore Callback Function Parameters

|Parameter|Description|Value Requirements|
|--|--|--|
|step|The step corresponding to when the optimizer data is dumped.|Positive integer.|
|ctx|Callback function context.|Determined by the registering party.|

**Table 2<a id="table_tft_07"></a>** Non-MindSpore Callback Function Parameters

|Parameter|Description|Value Requirements|
|--|--|--|
|step|Step corresponding to the dump of optimizer data.|Positive integer.|
|args|Parameters set by tft_set_step_args.|Determined by the registering party.|

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft\_register\_save\_ckpt\_handler

**Function**

Registers the framework-side dump callback function.

> [!NOTE]
> For the MindSpeed-LLM training framework, the callback function has already been adapted by MindIO TFT; for other frameworks, users need to ensure the safety of the callback function themselves.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_save_ckpt_handler(func: Callable, ctx = None)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|func|Mandatory|Terminal Checkpoint saving function, which completes the function of saving the terminal Checkpoint.|Callback function, must not be empty. For input parameter requirements of the callback function, see [Table 1](#table_tft_08). It is agreed that this callback function has no return value and throws an exception on execution failure.|
|ctx|Optional|Callback function context.|Defaults to empty.|

**Table 1<a id="table_tft_08"></a>**  Callback Function Parameters

|Name|Description|Value Requirements|
|--|--|--|
|step|The step corresponding to when the optimizer data is dumped.|Positive integer.|
|save_info|The rank list when different optimizers participate in saving terminal last words, where each element is a dictionary, and the dictionaries are arranged in the index order of ATTENTION (0) and MOE (1).|[<br>{<br>"type": int,   # optimizer type <br>"ranks": list, # rank list of the corresponding optimizer participating in saving terminal last words<br>},<br>]|
|args|The parameters configured by tft_set_step_args.|Determined by the registering party.|
|ctx|Callback function context.|Determined by the registering party.|

**Return Value**

No return value. On error, it logs an ERROR and throws an exception.

## tft\_register\_exit\_handler

**Function**

Registers a user-defined exit method with MindIO TFT.

> [!NOTE]
> Currently, the exit callback registration feature is provided only for the MindSpore framework. Users must ensure the safety of the callback function themselves. For other frameworks, the exit is handled by MindIO TFT.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_exit_handler(func: Callable, ctx = None)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|func|Mandatory|Callback function for completing exit.|Callback function, must not be empty. For input parameter requirements, see [Table 1](#table_tft_09). It is agreed that this callback function has no return value, and throws an exception on execution failure.|
|ctx|Optional|Callback function context.|Defaults to empty.|

**Table 1<a id="table_tft_09"></a>** Callback Function Parameters

|Name|Description|Value Requirements|
|--|--|--|
|ctx|Callback function context.|Determined by the registering party.|

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft\_register\_stop\_handler

**Function**

Registers a callback function to stop training during the recovery process.

> [!NOTE]
> For the MindSpeed-LLM training framework, the callback function has already been adapted by MindIO TFT; for other frameworks, users need to ensure the safety of the callback function themselves.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_stop_handler(func: Callable, ctx = None)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|func|Mandatory|Callback function for stopping training. It implements the stop training function and throws a FORCE STOP exception to hand over control of the training main thread to the decorator.|Callback function, must not be empty. For input parameter requirements of the callback function, see [Table 1](#table_tft_19). It is agreed that this callback function has no return value, and throws an exception on execution failure.|
|ctx|Optional|Callback function context.|Defaults to empty.|

**Table 1<a id="table_tft_19"></a>**  Callback Function Parameters

|Name|Description|Value Requirements|
|--|--|--|
|args|Parameters set by tft_set_step_args.|Determined by the registering party.|
|ctx|Callback function context.|Determined by the registering party.|

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft\_register\_clean\_handler

**Function**

Registers a callback function for cleaning up residual operator execution during the recovery process.

> [!NOTE]
> For the MindSpeed-LLM training framework, the callback function has already been adapted by MindIO TFT; for other frameworks, users must ensure the safety of the callback function themselves.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_clean_handler(func: Callable, ctx = None)
```

**Parameters**

|Name|Mandatory|Description|Value requirements|
|--|--|--|--|
|func|Mandatory|Callback function that cleans up residual operators and performs the functions of cleaning up residual operators and underlying faults.|Callback function, must not be empty. For input parameter requirements of the callback function, see [Table 1](#table_tft_10). It is agreed that this callback function has the following return values: <ul><li>0: Success.</li><li>1: Failure.</li><li>2: UCE scenario where the model optimizer does not need to be rebuilt.</li></ul>|
|ctx|Optional|Callback function context.|Defaults to empty.|

**Table 1<a id="table_tft_10"></a>**  Callback function parameters

|Name|Description|Value requirements|
|--|--|--|
|is_uce_error|Indicates whether a UCE fault has occurred on this card.|<ul><li>False: No UCE fault has occurred.</li><li>True: A UCE fault has occurred.</li></ul>|
|args|Parameters set by tft_set_step_args.|Determined by the registering party.|
|ctx|Callback function context.|Determined by the registering party.|

**Return Value**

No return value. On error, it logs an ERROR and throws an exception.

## tft\_register\_rebuild\_group\_handler

**Function**

Registers the callback function for MindIO ARF group rebuilding.

> [!NOTE]
> For the MindSpeed-LLM training framework, the callback function has already been adapted by MindIO TFT; for other frameworks, users need to ensure the safety of the callback function themselves.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_rebuild_group_handler(func: Callable, ctx = None)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|func|Mandatory|Callback function for MindIO ARF group rebuilding, which clears the old communication groups and rebuilds new communication groups for normal nodes and restarted nodes. The default execution timeout of the callback function is 180 seconds. If the timeout is exceeded, it will cause process execution failure. Users can set the timeout via the environment variable TTP_NORMAL_ACTION_TIME_LIMIT.|Callback function, must not be empty. For input parameter requirements, see [Table 1](#table_tft_11). It is agreed that this callback function has no return value and throws an exception on execution failure.|
|ctx|Optional|Callback function context.|Defaults to empty.|

**Table 1<a id="table_tft_11"></a>**  Callback Function Parameters

|Name|Description|Value Requirements|
|--|--|--|
|fault_ranks|Set of faulty cards.|list.|
|args|Parameters configured by tft_set_step_args.|Determined by the registering party.|
|ctx|Callback function context.|Determined by the registering party.|

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft\_register\_repair\_handler

**Function**

Registers the repair callback function.

> [!NOTE]
>
> - For the MindSpeed-LLM training framework, the callback function has already been adapted by MindIO TFT; for other frameworks, users must ensure the safety of the callback function themselves.
> - MindIO TFT has already rebuilt and overwritten the variables in the model optimizer within the callback function. For other variables involved in computation that users customize in the framework, users must implement their rebuilding and overwriting in repair themselves.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_repair_handler(func: Callable, ctx = None)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|func|Mandatory|The repair callback function, which completes data repair functions such as optimizer repair. The default execution timeout of the callback function is 180 seconds. If the timeout is exceeded, it will cause process execution failure. Users can set the timeout via the environment variable TTP_NORMAL_ACTION_TIME_LIMIT.|A callback function, which must not be empty. For input parameter requirements of the callback function, see [Table 1](#table_tft_12). It is agreed that this callback function has no return value, and throws an exception on execution failure.|
|ctx|Optional|The callback context.|Defaults to empty.|

**Table 1<a id="table_tft_12"></a>**  Callback Function Parameters

|Name|Description|Value Requirements|
|--|--|--|
|step|The step corresponding to the repair.|A positive integer.|
|need_rebuild|Whether the repair requires rebuilding the model and optimizer.|<ul><li>False: No rebuild is required.</li><li>True: A rebuild is required.</li></ul>|
|error_ranks|The list of faulty cards that need to be repaired.|A list.|
|repair_info|The repair strategy dict, in which the optimizer type corresponds according to the relationship of ATTENTION (0) and MOE (1).|{<br>"type": int,   # Optimizer type <br>"repair_type": Enum,   # For enum values, see [RepairType](#repairtype) <br>"src": list,    # List of source cards for optimizer repair data <br>"dst": list,   # List of destination cards for optimizer repair data<br>"rank_list": list, # List of cards required for establishing the repair communication group<br>}|
|args|The parameters set by tft_set_step_args.|Determined by the registering party.|
|ctx|The callback function context.|Determined by the registering party.|

**Return Value**

No return value. On error, it prints an ERROR log and throws an exception.

## tft\_register\_rollback\_handler

**Function**

Registers the rollback function.

> [!NOTE]
> For the MindSpeed-LLM training framework, the callback function has already been adapted by MindIO TFT; for other frameworks, users need to ensure the safety of the callback function themselves.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_rollback_handler(func: Callable, ctx = None)
```

**Parameters**

|Name|Mandatory|Description|Value requirements|
|--|--|--|--|
|func|Mandatory|Rollback callback function, which completes reset operations such as dataset rollback. The callback function execution timeout defaults to 180 seconds. If a timeout occurs, it will cause process execution failure. Users can set the timeout via the environment variable TTP_NORMAL_ACTION_TIME_LIMIT.|Callback function, must not be empty. For input parameter requirements of the callback function, see [Table 1](#table_tft_13). It is agreed that this callback function has no return value, and throws an exception on execution failure.|
|ctx|Optional|Callback function context.|Defaults to empty.|

**Table 1<a id="table_tft_13"></a>**  Callback function parameters

|Name|Description|Value requirements|
|--|--|--|
|step|The step to roll back to.|Positive integer.|
|args|Parameters set by tft_set_step_args.|Determined by the registering party.|
|ctx|Callback function context.|Determined by the registering party.|

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft\_register\_stream\_sync\_handler

**Function**

Registers a synchronous callback function.

> [!NOTE]
> For the MindSpeed-LLM training framework, the callback function has already been adapted by MindIO TFT; for other frameworks, users need to ensure the safety of the callback function themselves.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_stream_sync_handler(func: Callable, ctx=None)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|func|Mandatory|Synchronous callback function that performs synchronization operations after training is paused. It prevents residual operators from remaining unexecuted in the operator queue after training is paused.|Callback function, must not be empty. The callback function has no parameters. It is agreed that this callback function has no return value and throws an exception on execution failure.|
|ctx|Optional|Callback function context.|Determined by the registering party.|

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft\_register\_zit\_upgrade\_rollback\_handler

**Function**

The training framework registers a callback function with the Processor for rolling back the upgrade process.

> [!NOTE]
> For the MindSpeed-LLM training framework, the callback function has completed adaptation; for other frameworks, users need to ensure the safety of the callback function themselves.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_zit_upgrade_rollback_handler(func: Callable, ctx = None)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|func|Mandatory|Rollback callback function that completes reset operations such as dataset rollback. The default execution timeout of the callback function is 180 seconds. If the timeout is exceeded, it will cause process execution failure. Users can set the timeout via the environment variable TTP_NORMAL_ACTION_TIME_LIMIT.|Callback function, must not be empty. It is agreed that this callback function has no return value and throws an exception on execution failure.|
|ctx|Optional|Callback function context.|Defaults to empty.|

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft\_register\_zit\_upgrade\_repair\_handler

**Function**

The training framework registers a callback function for repair during the upgrade process with the Processor.

> [!NOTE]
> For the MindSpeed-LLM training framework, the callback function has already completed adaptation; for other frameworks, users must ensure the safety of the callback function themselves.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_zit_upgrade_repair_handler(func: Callable, ctx = None)
```

**Parameters<a id="section34575883518"></a>**

|Name|Mandatory|Description|Value requirements|
|--|--|--|--|
|func|Mandatory|Repair callback function, which completes data repair operations such as optimizer repair. The default execution timeout of the callback function is 180 seconds. If the timeout is exceeded, it will cause process execution failure. Users can set the timeout via the environment variable TTP_NORMAL_ACTION_TIME_LIMIT.|Callback function, not empty. It is agreed that this callback function has no return value and throws an exception on execution failure.|
|ctx|Optional|Callback function context.|Defaults to empty.|

**Return Value**

No return value. On error, it logs an ERROR message and throws an exception.

## tft\_register\_zit\_upgrade\_rebuild\_handler

**Function**
The training framework registers with the Processor a callback function for rebuilding the communication group during the upgrade process.

> [!NOTE]
> For the MindSpeed-LLM training framework, the callback function has already completed adaptation; for other frameworks, users must ensure the safety of the callback function themselves.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_zit_upgrade_rebuild_handler(func: Callable, ctx = None)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|func|Mandatory|Rebuild callback function, which completes the repair operation of rebuilding the communication group during the upgrade process. The callback function execution timeout defaults to 180 seconds. If it times out, it will cause process execution failure. Users can set the timeout via the environment variable TTP_NORMAL_ACTION_TIME_LIMIT.|Callback function, not empty. It is agreed that this callback function has no return value, and throws an exception on execution failure.|
|ctx|Optional|Callback function context.|Defaults to empty.|

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft\_register\_zit\_downgrade\_rebuild\_handler

**Function**

The training framework registers a callback function with the Processor for rebuilding and repairing during the downgrade process.

> [!NOTE]
> For the MindSpeed-LLM training framework, the callback function has completed adaptation; for other frameworks, users need to ensure the safety of the callback function themselves.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_zit_downgrade_rebuild_handler(func: Callable, ctx = None)
```

**Parameters**

|Name|Mandatory|Description|Value requirements|
|--|--|--|--|
|func|Mandatory|Rebuild callback function, which completes the rebuild and repair operation in the downgrade process. The default execution timeout of the callback function is 180 seconds. If the timeout is exceeded, it will cause process execution failure. Users can set the timeout via the environment variable TTP_NORMAL_ACTION_TIME_LIMIT.|Callback function, not empty. It is agreed that this callback function has no return value, and throws an exception on execution failure.|
|ctx|Optional|Callback function context.|Defaults to empty.|

**Return Value**

No return value. On error, it prints an ERROR log and throws an exception.

## tft\_register\_exception\_handler

**Function**

Registers an exception handler.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_exception_handler(fault_pattern: str, fault_type: str, fault_handle: Callable)
```

**Parameters**

|Parameter|Mandatory|Description|Value Requirements|
|--|--|--|--|
|fault_pattern|Mandatory|Exception keyword. Used to precisely match the exception type.|A keyword string in the exception information.|
|fault_type|Mandatory|Exception type. Used together with the return value of fault_handle to report exception information in MindIO when the corresponding exception is caught.|String. The value range is as follows (for details, see [ReportState](#reportstate)):<ul><li>RS_NORMAL</li><li>RS_RETRY</li><li>RS_UCE</li><li>RS_UCE_CORRUPTED</li><li>RS_HCCL_FAILED</li><li>RS_INIT_FINISH</li><li>RS_PREREPAIR_FINISH</li><li>RS_STEP_FINISH</li><li>RS_UNKNOWN</li></ul>|
|fault_handle|Mandatory|Exception handling method. Used to receive the exception information string and return a string. This return value is used together with fault_type when reporting exception information.|An executable method that needs to receive the exception string and returns a string.|

**Return Value**

No return value.

## tft\_report\_error

**Function**

Reports the error type.

**Format**

```python
mindio_ttp.framework_ttp.tft_report_error(error_type: ReportState)
```

**Parameters**

|Name|Mandatory|Description|Value requirements|
|--|--|--|--|
|error_type|Mandatory|Reports the exception type, which determines the subsequent repair process.|Actual error type. For the value range, see [ReportState](#reportstate).|

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft\_wait\_next\_action

**Function**

During repair, the training main thread calls this interface in the decorator to wait for the worker thread to complete business data repair.

> [!NOTE]
> This interface is a blocking interface. It keeps blocking until the next action is obtained.

**Format**

```python
mindio_ttp.framework_ttp.tft_wait_next_action()
```

**Parameters**

None

**Return Value**

- 0: success
- 1: failure

## tft_get_repair_step

**Function**

Queries the step value of the repair position.

**Format**

```python
mindio_ttp.framework_ttp.tft_get_repair_step()
```

**Parameters**

None

**Return Value**

The step used for repair. A return value of 0 indicates an invalid value.

## tft_get_repair_type

**Function**

Provided for MindSpore to query the repair type in the callback during the stop/clean/repair phase.

**Format**

```python
mindio_ttp.framework_ttp.tft_get_repair_type()
```

**Parameters**

None

**Return Value**

A string.

- retry: perform UCE repair.
- recover: perform ARF repair.
- dump: Execute terminal last words.
- unknown: No repair type found.

## tft_is_reboot_node

**Function**

In the MindIO ARF function process, determines whether the current process is a node that has been restarted after a fault. It can only be called immediately after the tft_start_processor interface is successfully called, and can only be called once.

**Format**

```python
mindio_ttp.framework_ttp.tft_is_reboot_node()
```

**Parameters**

None

**Return Value**

A bool value indicating whether the node was restarted after a fault.

## tft\_get\_reboot\_type

**Function**

Provided for MindSpore to call. After a node is restarted following a fault, the training framework obtains the node restart scenario type from mindio\_ttp. This interface can be called only once after process startup.

**Format**

```python
mindio_ttp.framework_ttp.tft_get_reboot_type()
```

**Parameters**

None

**Return Value**

A string.

- arf: indicates process rescheduling.
- hot switch: indicates sub-health hot switch.

## tft\_reset\_limit\_step

**Function**

Updates the prelock flag in the Processor to true and reset limitStep\_ to the maximum value.

**Format**

```python
mindio_ttp.framework_ttp.tft_reset_limit_step()
```

**Parameters**

None

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft_set_dp_group_info

**Function**

The training framework registers DP group information with the Processor.

**Format**

```python
mindio_ttp.controller_ttp.tft_set_dp_group_info(rank: int, dp_rank_list: list)
```

**Parameters**

|Parameter|Mandatory|Description|Value Requirements|
|--|--|--|--|
|rank|Mandatory|Current rank.|Greater than or equal to 0.|
|dp_rank_list|Mandatory|DP group information.|Non-empty.|

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft\_report\_load\_ckpt\_step

**Function**

When using periodic Checkpoint repair, reports the step loaded from the checkpoint.

**Format**

```python
mindio_ttp.framework_ttp.tft_report_load_ckpt_step(step: int)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|step|Mandatory|Step loaded from the Checkpoint.|Non-negative integer.|

**Return Value**

None

## tft\_register\_decrypt\_handler

**Function**

If the user enables TLS encryption, this interface must be used to register the private key passphrase decryption function.

**Format**

```python
mindio_ttp.framework_ttp.tft_register_decrypt_handler(decryptor: Callable)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|decryptor|Mandatory|User-defined private key passphrase decryption function.|TLS encryption is configured through tft_start_controller and tft_init_processor, and if the passphrase is ciphertext, a decryption function must be registered. For specific configuration guidance, see [Importing TLS Certificates](../../07_references/00_fault_recovery_acceleration/04_security_management_and_hardening.md#importing-tls-certificates).|

**Callback Function Parameters**

|Name|Description|Value Requirements|
|--|--|--|
|cipherText|Private key passphrase to be decrypted.|Determined by the registering party.|

**Callback Function Return Value**: plainText, string, the decrypted private key passphrase.

**Return Value**

No return value. On error, an ERROR log is printed and an exception is thrown.

## tft_notify_controller_dump

**Function**

Provided for MindCluster to call, notifying MindIO TFT to proactively stop training, perform a dump, and then exit training.

**Format**

```python
mindio_ttp.controller_ttp.tft_notify_controller_dump()
```

**Parameters**

None

**Return Value**

- 0: call succeeded
- 1: call failed

## tft\_notify\_controller\_stop\_train

**Function**

Provided for MindCluster to call, notifying MindIO TFT to proactively stop training and informing MindIO TFT of the information about the faulty card.

**Format**

```python
mindio_ttp.controller_ttp.tft_notify_controller_stop_train(fault_ranks: dict, stop_type: str = "stop", timeout: int = None)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|fault_ranks|Mandatory|Information about the faulty card.|Dictionary of <int key, int errorType>:<ul><li>key is the rank number of the faulty card</li><li>errorType is the fault type:</li><ul><li>0: UCE fault</li><li>1: non-UCE fault</li></ul></ul>|
|stop_type|Optional|Type of stopping training.|String, supporting the following two modes:<ul><li>"stop": pause training, taskabort mode.</li><li>"pause": pause training, non-taskabort mode.</li></ul>|
|timeout|Optional|Timeout for waiting for the next notification from MindCluster after pausing training.|Non-negative integer, unit: s.|

**Return Value**

- 0: call succeeded
- 1: call failed

## tft\_notify\_controller\_on\_global\_rank

**Function**

Provided for MindCluster to call, notifying MindIO TFT of the global faulty card information.

**Format**

```python
mindio_ttp.controller_ttp.tft_notify_controller_on_global_rank(fault_ranks: dict,time:int=1)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|fault_ranks|Mandatory|Information about the faulty card.|Dictionary of <int key, int errorType>: <ul><li>key is the rank number of the faulty card</li><li>errorType is the fault type:</li><ul><li>0: UCE fault.</li><li>1: non-UCE fault.</li></ul></ul>|
|time|Optional|Determined by the environment variable, it decides the maximum time for interacting with the MindCluster repair policy.|int, value range: [1, 3600], default value: 1, unit: s.|

**Return Value**

- 0: call succeeded
- 1: call failed

## tft\_notify\_controller\_prepare\_action

**Function**

Provided for MindCluster to call, notifying MindIO TFT of the repair strategy to be executed.

> [!NOTE]
> The repair strategy must fall within the range of optional repair strategies negotiated between MindCluster and MindIO TFT.

**Format**

```python
mindio_ttp.controller_ttp.tft_notify_controller_prepare_action(action: str, fault_ranks: dict = None)
```

**Parameters**

|Name|Type|Mandatory|Description|
|--|--|--|--|
|action|Mandatory|Notifies MindIO TFT of the sub-health migration hot switch action.|str. The supported repair strategies are as follows: <ul><li>hot switch</li><li>stop switch</li></ul>|
|fault_ranks|Optional|Information about the faulty card.|dict. The key is the rank number, ranging from 0 to 100000, and the value is errtype, ranging from 0 to 2.|

**Return Value**

- 0: call succeeded
- 1: call failed

## tft_notify_controller_change_strategy

**Function**

Provided for MindCluster to call, notifying MindIO TFT of the repair strategy to be executed.

> [!NOTE]
> The repair strategy must fall within the range of optional repair strategies negotiated between MindCluster and MindIO TFT.

**Format**

```python
mindio_ttp.controller_ttp.tft_notify_controller_change_strategy(strategy: str, params: str = "")
```

**Parameters**

|Name|Mandatory|Description|Value requirements|
|--|--|--|--|
|strategy|Mandatory|Notifies MindIO TFT of the repair strategy.|str. The supported repair strategies are as follows: <ul><li>retry</li><li>downgrade </li><li>upgrade</li><li>recover</li><li>dump</li><li>continue</li><li>migration</li><li>exit</li></ul>|
|params|<ul><li>Mandatory for downgrade training</li><li>Optional for others</li></ul>|Downgrade training parameters.|str. Default value: "".|

**Return Value**

- 0: call succeeded
- 1: call failed

## tft\_register\_mindx\_callback

**Function**

Provided for MindCluster to call, this interface registers the repair process callback function with MindIO TFT.

**Format**

```python
mindio_ttp.controller_ttp.tft_register_mindx_callback(action: str, func: Callable)
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|action|Mandatory|The action name to be registered by the callback function.|str. Supported action names are as follows: <ul><li>report_fault_ranks</li> <li>report_stop_complete</li><li>report_strategies</li><li>report_result</li></ul>|
|func|Mandatory|The function to be registered.|Callback function, must not be empty. For details about the callback function input parameters, see [Table 1](#table_tft_14) to [Table 4](#table_tft_17).|

**Table 1<a id="table_tft_14"></a>** Callback function parameters when action is report\_fault\_ranks

|Name|Description|Value Requirements|
|--|--|--|
|error_rank_dict|Information about the faulty card.|Dictionary of <int key, int errorType>: <ul><li>key is the rank number of the faulty card.</li><li>errorType is the fault type:</li><ul><li>0: UCE fault.</li><li>1: non-UCE fault.</li></ul></ul>|

**Table 2<a id="table_tft_15"></a>** Callback function parameters when action is report\_stop\_complete

|Name|Description|Value Requirements|
|--|--|--|
|code|Execution result of the action.|<ul><li>0: success.</li><li>400: common error.</li><li>401: MindCluster task id does not exist.</li><li>402: model error.</li><li>403: sequence error.</li><li>404: not all Processors are ready.</li></ul>|
|msg|Message indicating whether training is stopped.|str.|
|error_rank_dict|Information about the faulty card.|Dictionary of <int key, int errorType>: <ul><li>key is the rank number of the faulty card.</li><li>errorType is the fault type:</li><ul><li>0: UCE fault.</li><li>1: non-UCE fault.</li></ul></ul>|

**Table 3<a id="table_tft_16"></a>**  Callback function parameters when action is report\_strategies

|Parameter|Description|Value Requirements|
|--|--|--|
|error_rank_dict|Information about the faulty card.|Dictionary of <int key, int errorType>: <ul><li>key is the rank number of the faulty card.</li><li>errorType is the fault type:</li><ul><li>0: UCE fault.</li><li>1: non-UCE fault.</li></ul></ul>|
|strategy_list|List of repair strategies supported by MindIO TFT based on the currently available replica information.|list. The supported repair strategy values are as follows (str): <ul><li>retry: execute UCE repair.</li><li>recover: execute ARF repair.</li><li>dump: execute terminal checkpoint.</li><li>exit: exit.</li></ul>|

**Table 4<a id="table_tft_17"></a>**  Callback function parameters when action is report\_result

|Parameter|Description|Value Requirements|
|--|--|--|
|code|Execution result of the action.|<ul><li>0: repair succeeded.</li><li>405: retry repair failed, and recover, dump, and exit repair strategies are supported.</li><li>406: repair failed, and dump or exit repair strategies are supported.</li><li>499: repair failed, and only the exit strategy is supported.</li></ul>|
|msg|Message indicating repair success or failure.|str|
|error_rank_dict|Information about the faulty card.|Dictionary of <int key, int errorType>: <ul><li>key is the rank number of the faulty card.</li><li>errorType is the fault type:</li><ul><li>0: UCE fault.</li> <li>1: non-UCE fault.</li></ul></ul>|
|curr_strategy|The repair strategy used this time.|str. The supported repair strategy values are the strategy_list in Table 3.|

**Return Value**

- 0: call succeeded
- 1: call failed

## tft\_query\_high\_availability\_switch

**Function**

Provided for MindCluster to call, to query in real time whether high availability is enabled.

**Format**

```python
mindio_ttp.controller_ttp.tft_query_high_availability_switch()
```

**Parameters**

None

**Return Value**

A bool value indicating whether high availability is enabled.

## tft\_can\_do\_uce\_repair

**Function**

Provided for MindSpore to call. Based on the UCE fault time triggered by the L2 Cache and the times before and after the optimizer update, it determines whether the optimizer data may be contaminated in the time dimension, and then returns the judgment result of whether repair is possible.

> [!NOTE]
> This interface only determines whether the optimizer data may be contaminated based on the intersection of time intervals, and cannot make the judgment based on memory addresses.

**Format**

```python
mindio_ttp.framework_ttp.tft_can_do_uce_repair(hbm_error_time: int, start_time: int = None, end_time: int = None)
```

**Parameters**

|Parameter|Mandatory or optional|Description|Value requirements|
|--|--|--|--|
|hbm_error_time|Mandatory|Time of the UCE fault triggered by the L2 Cache.|int|
|start_time|Optional|Time obtained from the device before the optimizer performs local update.|int|
|end_time|Optional|Time obtained from the device after the optimizer performs local update.|int|

**Return Value**

A bool value indicating whether UCE fast recovery can be performed, determined by the time intersection.

## tft\_set\_update\_start\_time

**Function**

Sets the optimizer update start time, which is used to determine whether the optimizer data may be contaminated in the time dimension, and then returns the judgment result of whether repair is possible.

**Format**

```python
mindio_ttp.utils.tft_set_update_start_time(start_time: int = None)
```

**Parameters**

|Name|Mandatory|Description|Value requirements|
|--|--|--|--|
|start_time|Optional|The time obtained from the device before the optimizer updates locally.|int|

**Return Value**

None

## tft\_set\_update\_end\_time

**Function**

Sets the optimizer update end time, which is used to determine whether the optimizer data may be contaminated in the time dimension, and then returns a judgment result indicating whether repair is possible.

**Format**

```python
mindio_ttp.utils.tft_set_update_end_time(end_time: int = None)
```

**Parameters**

|Parameter|Mandatory or Optional|Description|Value Requirements|
|--|--|--|--|
|end_time|Optional|Time obtained from the device after the optimizer updates locally.|int|

**Return Value**

None

## tft\_pause\_train

**Function**

Pauses training at a specific step.

**Format**

```python
mindio_ttp.framework_ttp.tft_pause_train(cur_step: int)
```

**Parameters**

|Parameter|Mandatory|Description|Value requirements|
|--|--|--|--|
|cur_step|Mandatory|The number of steps executed by the current training framework.|Non-negative integer.|

**Return Value**

None

## OptimizerType

**Function**

Defines the optimizer type enumeration.

**Format**

```python
mindio_ttp.framework_ttp.OptimizerType
```

**Parameters**

|Parameter|Mandatory|Description|Value Requirements|
|--|--|--|--|
|OptimizerType|Mandatory|Distinguishes the optimizer type: <ul><li>ATTENTION: attention mechanism type.</li><li>MOE: MOE scenario.</li></ul>|<ul><li>ATTENTION: 0</li><li>MOE: 1</li></ul>|

**Return Value**

None

## Action

**Function**

Enumeration of action types after the main thread reports an exception.

**Format**

```python
mindio_ttp.framework_ttp.Action
```

**Parameters**

|Parameter|Mandatory or Optional|Description|Value Requirements|
|--|--|--|--|
|Action|Mandatory|Distinguishes the action type after the main thread reports an exception, as follows:<ul><li>RETRY: Continue training after successful repair.</li><li>EXIT: Exit.</li></ul>|<ul><li>RETRY: 0</li><li>EXIT: 1</li></ul>|

**Return Value**

None

## ReportState

**Function**

Decorator for reporting the training state enumeration.

**Format**

```python
mindio_ttp.framework_ttp.ReportState
```

**Parameters**

|Name|Mandatory|Description|Value Requirements|
|--|--|--|--|
|ReportState|Mandatory|Distinguishes the type of training state reported:<ul><li>RS_NORMAL: normal state.</li><li>RS_UCE: UCE error.</li><li>RS_UCE_CORRUPTED: on-chip memory MULTI BIT ECC fault.</li><li>RS_HCCL_FAILED: HCCL recomputation failure.</li><li>RS_UNKNOWN: other errors.</li><li>RS_INIT_FINISH: in the MindSpore framework, the exception thrown by a newly started ARF node after the training process completes initialization.</li><li>RS_PREREPAIR_FINISH: the exception thrown by a newly started ARF node.</li><li>RS_STEP_FINISH: the exception thrown when the step-level pause in sub-health hot switch has completed.</li></ul>|<ul><li>RS_NORMAL.value: ttp_c2python_api.ReportState_RS_NORMAL.</li><li>RS_UCE.value: ttp_c2python_api.ReportState_RS_UCE.</li><li>RS_UCE_CORRUPTED: ttp_c2python_api.ReportState_RS_UCE_CORRUPTED.</li><li>RS_HCCL_FAILED.value: ttp_c2python_api.ReportState_RS_HCCL_FAILED.</li><li>RS_UNKNOWN.value: ttp_c2python_api.ReportState_RS_UNKNOWN.</li><li>RS_INIT_FINISH: ttp_c2python_api.ReportState_RS_INIT_FINISH.</li><li>RS_PREREPAIR_FINISH.value: ttp_c2python_api.ReportState_RS_PREREPAIR_FINISH.</li><li>RS_STEP_FINISH: ttp_c2python_api.ReportState_RS_STEP_FINISH.</li></ul>|

**Return Value**

None

## RepairType

**Function**

Defines the repair type enumeration.

**Format**

```python
mindio_ttp.framework_ttp.RepairType
```

**Parameters**

|Parameter|Mandatory or Optional|Description|Value Requirements|
|--|--|--|--|
|RepairType|Mandatory|Distinguishes the repair type:<ul><li>RT_SEND: The backup card sends data.</li><li>RT_UCE_HIGHLEVEL: The faulty card requires optimizer and model rebuilding.</li><li>RT_UCE_LOWLEVEL: The faulty card does not require optimizer and model rebuilding.</li><li>RT_ROLLBACK: Rolls back the dataset.</li><li>RT_RECV_REPAIR: The newly restarted ARF card receives data.</li><li>RT_LOAD_CKPT: Periodic Checkpoint data repair.</li><li>RT_LOAD_REBUILD: Rebuilds the model optimizer and repairs periodic Checkpoint data.</li></ul>|<ul><li>RT_SEND.value: ttp_c2python_api.RepairType_RT_SEND.</li><li>RT_UCE_HIGHLEVEL.value: ttp_c2python_api.RepairType_RT_UCE_HIGHLEVEL.</li><li>RT_UCE_LOWLEVEL.value: ttp_c2python_api.RepairType_RT_UCE_LOWLEVEL.</li><li>RT_ROLLBACK.value: ttp_c2python_api.RepairType_RT_ROLLBACK.</li><li>RT_RECV_REPAIR.value: ttp_c2python_api.RepairType_RT_RECV_REPAIR.</li><li>RT_LOAD_CKPT.value: ttp_c2python_api.RepairType_RT_LOAD_CKPT.</li><li>RT_LOAD_REBUILD.value: ttp_c2python_api.RepairType_RT_LOAD_REBUILD.</li></ul>|

**Return Value**

None
