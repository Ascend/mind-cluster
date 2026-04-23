# API接口参考

所有接口参数表和回调函数参数表，默认按照函数参数顺序排列。

## tft\_init\_controller

**接口功能**

初始化MindIO TFT Controller模块。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_init_controller(rank: int, world_size: int, enable_local_copy: bool, enable_arf=False, enable_zit=False)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|rank|必选|当前执行训练任务的NPU卡号。|int，[-1, world_size)。MindCluster在Torch Agent进程拉起Controller时rank值取-1。|
|world_size|必选|整个集群参与训练任务的卡数。|int，[1, 100000]。|
|enable_local_copy|必选|表示是否启用local copy。优化器更新前，先对优化器做一次备份。|<ul><li>False：关闭</li><li>True：启用</li></ul>|
|enable_arf|可选|MindIO ARF特性开关。|<ul><li>False：关闭</li><li>True：启用</li></ul>默认为False。|
|enable_zit|可选|MindIO ZIT特性开关。|<ul><li>False：关闭</li><li>True：启用</li></ul>默认为False。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_start\_controller

**接口功能**

在初始化Controller模块成功后，调用该接口以启动MindIO TFT Controller模块服务。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_start_controller(bind_ip: str, port: int, enable_tls=True, tls_info='')
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|bind_ip|必选|Controller所在节点IP地址或域名。|符合IP地址规范的IPv4地址，位于集群节点IP地址中，禁止全零IP地址，支持域名。|
|port|必选|Controller侦听端口号。|[1024, 65535]|
|enable_tls|可选|TLS加密传输开关。|<ul><li>False：关闭</li><li>True：启用</li></ul>默认为True。|
|tls_info|可选|TLS的证书配置。|默认为空，当开启TLS认证时，需要配置证书信息，具体字段应以键值对形式组织。具体配置指导见[导入TLS证书](./04_security_management_and_hardening.md#导入tls证书)。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_destroy\_controller

**接口功能**

在训练完成后，调用该接口以关闭MindIO TFT Controller服务。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_destroy_controller()
```

**接口参数**

无

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_init\_processor

**接口功能**

初始化MindIO TFT Processor模块。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_init_processor(rank: int, world_size: int, enable_local_copy: bool, enable_tls=True, tls_info='', enable_uce=True, enable_arf=False, enable_zit=False)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|rank|必选|当前执行训练任务NPU卡号。|int，[0, world_size)。|
|world_size|必选|参与训练任务的集群卡数。|int，[1, 100000]。|
|enable_local_copy|必选|是否启用local copy。|<ul><li>False：关闭</li><li>True：启用</li></ul>|
|enable_tls|可选|TLS加密传输开关。|<ul><li>False：关闭</li><li>True：启用</li></ul>默认为True。|
|tls_info|可选|TLS的证书配置。|默认为空，当开启TLS认证时，需要配置证书信息，具体字段应以键值对形式组织。具体配置指导见[导入TLS证书](./04_security_management_and_hardening.md#导入tls证书)。|
|enable_uce|可选|MindIO UCE特性开关。|<ul><li>False：关闭</li><li>True：启用</li></ul>默认为True。|
|enable_arf|可选|MindIO ARF特性开关。|<ul><li>False：关闭</li><li>True：启用</li></ul>默认为False。|
|enable_zit|可选|MindIO ZIT特性开关。|<ul><li>False：关闭</li><li>True：启用</li></ul>默认为False。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_start\_processor

**接口功能**

在初始化Processor模块成功后，调用该接口以启动MindIO TFT Processor模块服务。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_start_processor(master_ip: str, port: int, local_ip='')
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|master_ip|必选|Controller所在节点IP地址或域名。|符合IP地址规范的IPv4地址，位于集群节点IP地址中，禁止全零IP地址，支持域名。|
|port|必选|Controller侦听端口号。|[1024, 65535]|
|local_ip|可选|K8s中Processor所在节点的Service IP地址或域名。|符合IP地址规范的IPv4地址，位于集群节点IP地址中，禁止全零IP地址，支持域名。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_destroy\_processor

**接口功能**

在训练完成后，调用该接口以关闭MindIO TFT Processor服务。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_destroy_processor()
```

**接口参数**

无

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_start\_updating\_os

**接口功能**

在优化器状态更新前，调用该接口以更新optimizer state为Updating。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_start_updating_os(backup_step: int)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|backup_step|必选|备份的step。|-1或自然数，范围[-1, 9223372036854775807)。<ul><li>-1：表示不使用备份step。</li><li>自然数：优化器更新前，备份的优化器状态数据对应的step。</li></ul>|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_start\_copy\_os

**接口功能**

通知Processor开始copy优化器状态。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_start_copy_os()
```

**接口参数**

无

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_end\_updating\_os

**接口功能**

在优化器状态更新完成后，调用该接口以更新optimizer state为Updated。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_end_updating_os(step: int)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|step|必选|当前的step。|正整数，范围[1, 9223372036854775807)。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_set\_optimizer\_replica

**接口功能**

设置rank对应的优化器状态数据副本关系。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_set_optimizer_replica(rank: int, replica_info: list)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|rank|必选|当前执行训练任务的NPU卡号。|int，[0, 100000)。|
|replica_info|必选|副本关系list，其中每个元素是一个字典，字典按照ATTENTION（0）、MOE（1）的索引顺序排列。|[<br>{<br>"rank_list":list,   # 对应的一组副本关系rank列表，PyTorch场景为DP组rank list,MindSpore场景为该卡对应的所有副本卡的list <br>"replica_cnt":int,   # 副本数，PyTorch场景为副本数，MindSpore场景为rank_list的长度 <br>"replica_shift":int,  # PyTorch场景有效<br>},<br>]|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_exception\_handler

**接口功能**

装饰器，对MindSpeed-LLM的train方法进行装饰，捕获训练状态异常以及上报处理，对于用户的其他训练框架，本接口仅提供参考示例功能。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_exception_handler(func: Callable)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|函数作为参数。|框架的train方法。|

**返回值**

装饰器返回的func。

## tft\_set\_step\_args

**接口功能**

训练框架设置的参数集合。

> [!NOTE]说明
> 对于MindSpeed-LLM训练框架，设置功能已经由MindIO TFT完成适配，不需要调用。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_set_step_args(args)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|args|必选|训练框架设置需要保存的参数集合。MindIO TFT在 stop/clean/repair/rollback 等阶段调用注册的回调函数时，将参数集合传回，框架根据参数集合完成相应功能。|由训练框架决定，MindIO TFT不访问也不修改该参数集合，在 stop/clean/repair/rollback 等阶段时调用注册的业务回调将其传回，业务回调负责对取值范围进行校验。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_register\_rename\_handler

**接口功能**

注册框架侧rename回调函数。

> [!NOTE]说明 
> 对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_rename_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|rename函数，将保存成功的临终Checkpoint重命名，与原生框架Checkpoint命名规则一致。|回调函数，不为空，回调函数的入参要求请参见[表 1](#table_tft_06)和[表 2](#table_tft_07)，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**表 1<a id="table_tft_06"></a>**  MindSpore回调函数参数

|参数|说明|取值要求|
|--|--|--|
|step|dump优化器数据时对应的step。|正整数。|
|ctx|回调函数上下文。|由注册方决定。|

**表 2<a id="table_tft_07"></a>**  非MindSpore回调函数参数

|参数|说明|取值要求|
|--|--|--|
|step|dump优化器数据时对应的step。|正整数。|
|args|tft_set_step_args设置的参数。|由注册方决定。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_register\_save\_ckpt\_handler

**接口功能**

注册框架侧dump回调函数。

> [!NOTE]说明
> 对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_save_ckpt_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|临终Checkpoint保存函数，完成保存临终Checkpoint的功能。|回调函数，不为空，回调函数的入参要求请参见[表1](#table_tft_08)，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**表 1<a id="table_tft_08"></a>**  回调函数参数

|参数|说明|取值要求|
|--|--|--|
|step|dump优化器数据时对应的step。|正整数。|
|save_info|不同优化器参与保存临终遗言时的rank list，其中每个元素是一个字典，字典按照ATTENTION（0）、MOE（1）的索引顺序排列。|[<br>{<br>"type": int,   # 优化器类型 <br>"ranks": list, # 参与对应优化器保存临终遗言时的rank列表<br>},<br>]|
|args|tft_set_step_args设置的参数。|由注册方决定。|
|ctx|回调函数上下文。|由注册方决定。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_register\_exit\_handler

**接口功能**

向MindIO TFT注册用户自定义退出方法。

> [!NOTE]说明 
> 目前仅针对MindSpore框架提供了注册退出回调的功能，用户需要自行确保回调函数的安全性；其他框架的退出则由MindIO TFT负责。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_exit_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|完成退出的回调函数。|回调函数，不为空，回调函数的入参要求请参见[表1](#table_tft_09)，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**表 1<a id="table_tft_09"></a>**  回调函数参数

|参数|说明|取值要求|
|--|--|--|
|ctx|回调函数上下文。|由注册方决定。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_register\_stop\_handler

**接口功能**

在恢复过程中注册停止训练的回调函数。

> [!NOTE]说明
> 对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_stop_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|停止训练的回调函数，实现停止训练的功能，并抛出FORCE STOP异常将训练主线程控制权交由装饰器接管。|回调函数，不为空，回调函数的入参要求请参见[表1](#table_tft_19)，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**表 1<a id="table_tft_19"></a>**  回调函数参数

|参数|说明|取值要求|
|--|--|--|
|args|tft_set_step_args设置的参数。|由注册方决定。|
|ctx|回调函数上下文。|由注册方决定。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_register\_clean\_handler

**接口功能**

在恢复过程中注册清理残留算子执行的回调函数。

> [!NOTE]说明 
> 对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式<**

```python
mindio_ttp.framework_ttp.tft_register_clean_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|清理残留算子执行的回调函数，完成清理残留算子、底层故障的功能。|回调函数，不为空，回调函数的入参要求请参见[表1](#table_tft_10)。约定该回调函数返回值： <ul><li>0：成功。</li><li>1：失败。</li><li>2：UCE场景且无需重建模型优化器。</li></ul>|
|ctx|可选|回调函数上下文。|默认为空。|

**表 1<a id="table_tft_10"></a>**  回调函数参数

|参数|说明|取值要求|
|--|--|--|
|is_uce_error|表示该卡是否发生UCE故障。|<ul><li>False：未发生UCE故障。</li><li>True：发生UCE故障。</li></ul>|
|args|tft_set_step_args设置的参数。|由注册方决定。|
|ctx|回调函数上下文。|由注册方决定。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_register\_rebuild\_group\_handler

**接口功能**

注册MindIO ARF重新建组的回调函数。

> [!NOTE]说明
> 对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_rebuild_group_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|MindIO ARF重新建组的回调函数，完成正常节点与重启节点清理旧通信组并重建新通信组的功能。回调函数执行超时时间默认180秒。若超时，会导致流程执行失败。用户可通过环境变量TTP_NORMAL_ACTION_TIME_LIMIT来设置超时时间。|回调函数，不为空，回调函数的入参要求请参见[表1](#table_tft_11)，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**表 1<a id="table_tft_11"></a>**  回调函数参数

|参数|说明|取值要求|
|--|--|--|
|fault_ranks|故障卡集合。|list。|
|args|tft_set_step_args设置的参数。|由注册方决定。|
|ctx|回调函数上下文。|由注册方决定。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_register\_repair\_handler

**接口功能**

注册repair回调函数。

> [!NOTE]说明
>
> - 对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。
> - MindIO TFT已在回调函数中对模型优化器中的变量进行重建与覆写，用户在框架中自定义的其他参与计算的变量，需在repair中自行实现对其的重建与覆写。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_repair_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|repair回调函数，完成优化器修复等数据修复功能。回调函数执行超时时间默认180秒。若超时，会导致流程执行失败。用户可通过环境变量TTP_NORMAL_ACTION_TIME_LIMIT来设置超时时间。|回调函数，不为空，回调函数的入参要求请参见[表1](#table_tft_12)，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调上下文。|默认为空。|

**表 1<a id="table_tft_12"></a>**  回调函数参数

|参数|说明|取值要求|
|--|--|--|
|step|修复时对应的step。|正整数。|
|need_rebuild|-|修复是否需要重建模型和优化器。|<ul><li>False：无需重建。</li><li>True：需要重建。</li></ul>|
|error_ranks|需要修复的故障卡list。|list。|
|repair_info|修复策略dict，其中优化器类型按照ATTENTION（0）、MOE（1）的关系对应。|{<br>"type": int,   # 优化器类型 <br>"repair_type": Enum,   # 枚举类型取值参见[RepairType](#repairtype) <br>"src": list,    # 优化器修复数据的来源卡列表 <br>"dst": list,   # 优化器修复数据的目的卡列表<br>"rank_list": list, # 修复通信组建立所需要的卡列表<br>}|
|args|tft_set_step_args设置的参数。|由注册方决定。|
|ctx|回调函数上下文。|由注册方决定。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_register\_rollback\_handler

**接口功能**

注册rollback回滚函数。

> [!NOTE]说明 
> 对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_rollback_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|rollback回调函数，完成数据集回滚等重置操作。回调函数执行超时时间默认180秒。若超时，会导致流程执行失败。用户可通过设置环境变量TTP_NORMAL_ACTION_TIME_LIMIT来设置超时时间。|回调函数，不为空，回调函数的入参要求请参见[表1](#table_tft_13)，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**表 1<a id="table_tft_13"></a>**  回调函数参数

|参数|说明|取值要求|
|--|--|--|
|step|回滚到的step。|正整数。|
|args|tft_set_step_args设置的参数。|由注册方决定。|
|ctx|回调函数上下文。|由注册方决定。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_register\_stream\_sync\_handler

**接口功能**

注册同步回调函数。

> [!NOTE]说明
> 对于MindSpeed-LLM训练框架，回调函数已经由MindIO TFT完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_stream_sync_handler(func: Callable, ctx=None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|同步回调函数，完成训练暂停后同步操作。避免在暂停训练后算子队列有残留算子未执行完。|回调函数，不为空。回调函数无参数，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|由注册方决定。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_register\_zit\_upgrade\_rollback\_handler

**接口功能**

训练框架向Processor注册升级流程回滚的回调函数。

> [!NOTE]说明
> 对于MindSpeed-LLM训练框架，回调函数已经完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_zit_upgrade_rollback_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|rollback回调函数，完成数据集回滚等重置操作。回调函数执行超时时间默认180秒。若超时，会导致流程执行失败。用户可通过环境变量TTP_NORMAL_ACTION_TIME_LIMIT来设置超时时间。|回调函数，不为空，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_register\_zit\_upgrade\_repair\_handler

**接口功能**

训练框架向Processor注册升级流程修复的回调函数。

> [!NOTE]说明
> 对于MindSpeed-LLM训练框架，回调函数已经完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_zit_upgrade_repair_handler(func: Callable, ctx = None)
```

**接口参数<a id="section34575883518"></a>**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|repair回调函数，完成优化器修复等数据修复操作。回调函数执行超时时间默认180秒。若超时，会导致流程执行失败。用户可通过环境变量TTP_NORMAL_ACTION_TIME_LIMIT来设置超时时间。|回调函数，不为空，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_register\_zit\_upgrade\_rebuild\_handler

**接口功能**
训练框架向Processor注册升级流程重建通信组的回调函数。

> [!NOTE]说明 
> 对于MindSpeed-LLM训练框架，回调函数已经完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_zit_upgrade_rebuild_handler(func: Callable, ctx = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|rebuild回调函数，完成升级流程重建通信组的修复操作。回调函数执行超时时间默认180秒。若超时，会导致流程执行失败。用户可通过环境变量TTP_NORMAL_ACTION_TIME_LIMIT来设置超时时间。|回调函数，不为空，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_register\_zit\_downgrade\_rebuild\_handler

**接口功能**

训练框架向Processor注册降级流程重建修复的回调函数。

> [!NOTE]说明
> 对于MindSpeed-LLM训练框架，回调函数已经完成适配；而对于其他框架，用户需要自行确保回调函数的安全性。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_zit_downgrade_rebuild_handler(func: Callable, ctx = None)
```

**接口参数=**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|func|必选|rebuild回调函数，完成降级流程重建修复操作。回调函数执行超时时间默认180秒。若超时，会导致流程执行失败。用户可通过环境变量TTP_NORMAL_ACTION_TIME_LIMIT来设置超时时间。|回调函数，不为空，约定该回调函数无返回值，执行失败抛出异常。|
|ctx|可选|回调函数上下文。|默认为空。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_register\_exception\_handler

**接口功能**

注册异常处理程序。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_exception_handler(fault_pattern: str, fault_type: str, fault_handle: Callable)
```

**接口参数=**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|fault_pattern|必选|异常关键字。用于精确匹配异常类型。|异常信息中的关键字字符串。|
|fault_type|必选|异常类型。用于在捕获对应的异常时，与fault_handle的返回值一起在MindIO上报异常信息|字符串，取值范围如下（详情请参见[ReportState](#reportstate)）:<ul><li>RS_NORMAL</li><li>RS_RETRY</li><li>RS_UCE</li><li>RS_UCE_CORRUPTED</li><li>RS_HCCL_FAILED</li><li>RS_INIT_FINISH</li><li>RS_PREREPAIR_FINISH</li><li>RS_STEP_FINISH</li><li>RS_UNKNOWN</li></ul>|
|fault_handle|必选|异常处理方法。用于接收异常信息字符串，并返回一个字符串。该返回值与fault_type一起在上报异常信息时使用|可执行方法，该方法需要接收异常字符串，并且返回值为字符串。|

**返回值**

无返回值。

## tft\_report\_error

**接口功能**

上报错误类型。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_report_error(error_type: ReportState)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|error_type|必选|上报异常类型，用以决定后续修复流程。|实际错误类型。取值范围请参见[ReportState](#reportstate)。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_wait\_next\_action

**接口功能**

修复期间，训练主线程在装饰器中调用该接口等待从线程完成业务数据修复。

> [!NOTE]说明
> 该接口为阻塞接口，在未获取到下一次action前，会一直阻塞。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_wait_next_action()
```

**接口参数**

无

**返回值**

- 0：成功
- 1：失败

## tft\_get\_repair\_step

**接口功能**

查询修复位置的step值。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_get_repair_step()
```

**接口参数**

无

**返回值**

修复使用的step，返回0表示无效值。

## tft\_get\_repair\_type

**接口功能**

提供给MindSpore调用，用于在stop/clean/repair阶段的回调中查询修复类型。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_get_repair_type()
```

**接口参数**

无

**返回值**

str类型。

- retry：执行UCE修复。
- recover：执行ARF修复。
- dump：执行临终遗言。
- unknown：未找到修复类型。

## tft\_is\_reboot\_node

**接口功能**

MindIO ARF功能流程中，判断当前进程是否为故障后重新拉起的节点，仅支持在tft\_start\_processor接口调用成功后立即调用，且仅支持调用一次。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_is_reboot_node()
```

**接口参数**

无

**返回值**

bool值，表示是否为故障后重新拉起的节点。

## tft\_get\_reboot\_type

**接口功能**

提供给MindSpore调用，在故障重新拉起节点后，训练框架从mindio\_ttp获取节点重启场景类型，进程启动后仅支持调用一次。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_get_reboot_type()
```

**接口参数**

无

**返回值**

str类型。

- arf：代表进程重调度。
- hot switch：代表亚健康热切。

## tft\_reset\_limit\_step

**接口功能**

更新Processor中prelock标记为true，并重置limitStep\_为最大值。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_reset_limit_step()
```

**接口参数**

无

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_set\_dp\_group\_info

**接口功能**

训练框架向Processor注册DP组信息。

**接口格式**

```python
mindio_ttp.controller_ttp.tft_set_dp_group_info(rank: int, dp_rank_list: list)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|rank|必选|当前rank。|大于或等于0。|
|dp_rank_list|必选|DP组信息。|非空。|

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_report\_load\_ckpt\_step

**接口功能**

使用周期Checkpoint修复时，上报从Checkpoint加载的步数。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_report_load_ckpt_step(step: int)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|step|必选|从Checkpoint加载的步数。|非负整数。|

**返回值**

无

## tft\_register\_decrypt\_handler

**接口功能**

如果用户开启TLS加密，则需要使用该接口注册私钥口令解密函数。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_register_decrypt_handler(decryptor: Callable)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|decryptor|必选|用户自定义的私钥口令解密函数。|通过 tft_start_controller 和 tft_init_processor 配置TLS加密，并且如果口令为密文，则需注册解密函数。具体配置指导见[导入TLS证书](./04_security_management_and_hardening.md#导入tls证书)。|

**回调函数参数**

|参数|说明|取值要求|
|--|--|--|
|cipherText|需要解密的私钥口令。|由注册方决定。|

**回调函数返回值**为plainText : str，即解密后的私钥口令。

**返回值**

无返回值。出错时会打印ERROR日志并抛出异常。

## tft\_notify\_controller\_dump

**接口功能**

提供给MindCluster调用，通知MindIO TFT主动停止训练，执行dump后退出训练。

**接口格式**

```python
mindio_ttp.controller_ttp.tft_notify_controller_dump()
```

**接口参数**

无

**返回值**

- 0：调用成功
- 1：调用失败

## tft\_notify\_controller\_stop\_train

**接口功能**

提供给MindCluster调用，通知MindIO TFT主动停止训练，并告知MindIO TFT发生故障的卡信息。

**接口格式**

```python
mindio_ttp.controller_ttp.tft_notify_controller_stop_train(fault_ranks: dict, stop_type: str = "stop", timeout: int = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|fault_ranks|必选|发生故障的卡信息。|<int key, int errorType>字典：<ul><li>key为故障卡的rank号</li><li>errorType为故障类型：</li><ul><li>0：UCE故障</li><li>1：非UCE故障</li></ul></ul>|
|stop_type|可选|停止训练的类型。|字符串，支持以下两种方式：<ul><li>"stop"：暂停训练，taskabort方式。</li><li>"pause"：暂停训练，非taskabort方式。</li></ul>|
|timeout|可选|暂停训练之后等待MindCluster做下一步通知的超时时间。|非负整数，单位：s。|

**返回值**

- 0：调用成功
- 1：调用失败

## tft\_notify\_controller\_on\_global\_rank

**接口功能**

提供给MindCluster调用，通知MindIO TFT全局的故障卡信息。

**接口格式**

```python
mindio_ttp.controller_ttp.tft_notify_controller_on_global_rank(fault_ranks: dict,time:int=1)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|fault_ranks|必选|发生故障的卡信息。|<int key, int errorType>字典：<ul><li>key为故障卡的rank号</li><li>errorType为故障类型：</li><ul><li>0：UCE故障。</li><li>1：非UCE故障。</li></ul></ul>|
|time|可选|根据环境变量设置，决定与MindCluster的修复策略交互的最大时间。|int，取值范围：[1, 3600]，默认值：1，单位：s。|

**返回值**

- 0：调用成功
- 1：调用失败

## tft\_notify\_controller\_prepare\_action

**接口功能**

提供给MindCluster调用，通知MindIO TFT要执行的修复策略。

> [!NOTE]说明
> 该修复策略必须在MindCluster和MindIO TFT协商的可选修复策略范围内。

**接口格式**

```python
mindio_ttp.controller_ttp.tft_notify_controller_prepare_action(action: str, fault_ranks: dict = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|action|必选|通知MindIO TFT亚健康迁移热切动作。|str，支持的修复策略如下：<ul><li>hot switch</li><li>stop switch</li></ul>|
|fault_ranks|可选|发生故障的卡信息。|dict，key为rank号，取值范围0\~100000，value为errtype，取值范围0\~2。|

**返回值**

- 0：调用成功
- 1：调用失败

## tft\_notify\_controller\_change\_strategy

**接口功能**

提供给MindCluster调用，通知MindIO TFT要执行的修复策略。

> [!NOTE]说明
> 该修复策略必须在MindCluster和MindIO TFT协商的可选修复策略范围内。

**接口格式**

```python
mindio_ttp.controller_ttp.tft_notify_controller_change_strategy(strategy: str, params: str = "")
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|strategy|必选|通知MindIO TFT修复策略。|str，支持的修复策略如下：<ul><li>retry</li><li>downgrade </li><li>upgrade</li><li>recover</li><li>dump</li><li>continue</li><li>migration</li><li>exit</li></ul>|
|params|<ul><li>降级训练必选</li><li>其他可选</li></ul>|降级训练参数。|str，默认值：""。|

**返回值**

- 0：调用成功
- 1：调用失败

## tft\_register\_mindx\_callback

**接口功能**

提供给MindCluster调用，向MindIO TFT注册修复流程回调函数接口。

**接口格式**

```python
mindio_ttp.controller_ttp.tft_register_mindx_callback(action: str, func: Callable)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|action|必选|回调函数要注册的动作名。|str，支持的动作名如下：<ul><li>report_fault_ranks</li> <li>report_stop_complete</li><li>report_strategies</li><li>report_result</li></ul>|
|func|必选|要注册的函数。|回调函数，不为空，回调函数入参详情请参见[表1](#table_tft_14) ~ [表4](#table_tft_17)。|

**表 1<a id="table_tft_14"></a>**  action为report\_fault\_ranks时回调函数参数

|参数|说明|取值要求|
|--|--|--|
|error_rank_dict|发生故障的卡信息。|<int key, int errorType>字典：<ul><li>key为故障卡的rank号。</li><li>errorType为故障类型：</li><ul><li>0：UCE故障。</li><li>1：非UCE故障。</li></ul></ul>|

**表 2<a id="table_tft_15"></a>**  action为report\_stop\_complete时回调函数参数

|参数|说明|取值要求|
|--|--|--|
|code|action执行结果。|<ul><li>0：成功。</li><li>400：普通错误。</li><li>401：MindCluster task id不存在。</li><li>402：模型错误。</li><li>403：顺序错误。</li><li>404：Processor未全部准备就绪。</li></ul>|
|msg|训练是否停止消息。|str。|
|error_rank_dict|发生故障的卡信息。|<int key, int errorType>字典：<ul><li>key为故障卡的rank号。</li><li>errorType为故障类型：</li><ul><li>0：UCE故障。</li><li>1：非UCE故障。</li></ul></ul>|

**表 3<a id="table_tft_16"></a>**  action为report\_strategies时回调函数参数

|参数|说明|取值要求|
|--|--|--|
|error_rank_dict|发生故障的卡信息。|<int key, int errorType>字典：<ul><li>key为故障卡的rank号。</li><li>errorType为故障类型：</li><ul><li>0：UCE故障。</li><li>1：非UCE故障。</li></ul></ul>|
|strategy_list|基于当前可用的副本信息，MindIO TFT支持的修复策略列表。|list，支持的修复策略可选值如下（str）：<ul><li>retry：执行UCE修复。</li><li>recover：执行ARF修复。</li><li>dump：执行临终遗言。</li><li>exit：退出。</li></ul>|

**表 4<a id="table_tft_17"></a>**  action为report\_result时回调函数参数

|参数|说明|取值要求|
|--|--|--|
|code|action的执行结果。|<ul><li>0：修复成功。</li><li>405：retry修复失败，支持做recover、dump、exit修复策略。</li><li>406：修复失败，支持做dump或exit修复策略。</li><li>499：修复失败，仅支持exit策略。</li></ul>|
|msg|修复成功或失败的消息。|str|
|error_rank_dict|发生故障的卡信息。|<int key, int errorType>字典：<ul><li>key为故障卡的rank号。</li><li>errorType为故障类型：</li><ul><li>0：UCE故障。</li> <li>1：非UCE故障。</li></ul></ul>|
|curr_strategy|本次修复策略。|str，支持的修复策略取值范围为表3中的strategy_list。|

**返回值**

- 0：调用成功
- 1：调用失败

## tft\_query\_high\_availability\_switch

**接口功能**

提供给MindCluster调用，实时查询是否开启高可用。

**接口格式**

```python
mindio_ttp.controller_ttp.tft_query_high_availability_switch()
```

**接口参数**

无

**返回值**

bool值，是否开启高可用。

## tft\_can\_do\_uce\_repair

**接口功能**

提供给MindSpore调用，根据L2 Cache触发的UCE故障时间和优化器更新前后时间，判断优化器数据在时间维度是否有被污染的可能，进而返回是否能修复的判断结果。

> [!NOTE]说明
> 该接口仅从时间区间交集上判断优化器数据是否有被污染可能，无法根据内存地址判断。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_can_do_uce_repair(hbm_error_time: int, start_time: int = None, end_time: int = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|hbm_error_time|必选|L2 Cache触发的UCE故障时间。|int|
|start_time|可选|优化器在本地更新前从device获取的时间。|int|
|end_time|可选|优化器在本地更新后从device获取的时间。|int|

**返回值**

bool值，根据时间交集判断是否可以进行UCE快恢的判断结果。

## tft\_set\_update\_start\_time

**接口功能**

设置优化器更新开始时间，用于判断优化器数据在时间维度是否有被污染可能，进而返回是否能修复的判断结果。

**接口格式**

```python
mindio_ttp.utils.tft_set_update_start_time(start_time: int = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|start_time|可选|优化器在本地更新前从device获取的时间。|int|

**返回值**

无

## tft\_set\_update\_end\_time

**接口功能**

设置优化器更新结束时间，用于判断优化器数据在时间维度是否有被污染可能，进而返回是否能修复的判断结果。

**接口格式**

```python
mindio_ttp.utils.tft_set_update_end_time(end_time: int = None)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|end_time|可选|优化器在本地更新后从device获取的时间。|int|

**返回值**

无

## tft\_pause\_train

**接口功能**

将训练暂停在某一个step。

**接口格式**

```python
mindio_ttp.framework_ttp.tft_pause_train(cur_step: int)
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|cur_step|必选|当前训练框架执行的步数。|非负整数。|

**返回值**

无

## OptimizerType

**接口功能**

定义优化器类型枚举。

**接口格式**

```python
mindio_ttp.framework_ttp.OptimizerType
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|OptimizerType|必选|区分优化器类型：<ul><li>ATTENTION：注意力机制类型。</li><li>MOE：MOE场景。</li></ul>|<ul><li>ATTENTION：0</li><li>MOE：1</li></ul>|

**返回值**

无

## Action

**接口功能**

主线程上报异常后的动作类型枚举。

**接口格式**

```python
mindio_ttp.framework_ttp.Action
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|Action|必选|区分主线程上报异常后的动作类型，具体如下：<ul><li>RETRY：修复成功后续训。</li><li>EXIT：退出。</li></ul>|<ul><li>RETRY：0</li><li>EXIT：1</li></ul>|

**返回值**

无

## ReportState

**接口功能**

装饰器上报训练状态枚举。

**接口格式**

```python
mindio_ttp.framework_ttp.ReportState
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|ReportState|必选|区分上报的训练状态类型：<ul><li>RS_NORMAL：正常状态。</li><li>RS_UCE：UCE错误。</li><li>RS_UCE_CORRUPTED：片上内存 MULTI BIT ECC故障。</li><li>RS_HCCL_FAILED：HCCL重计算失败。</li><li>RS_UNKNOWN：其他错误。</li><li>RS_INIT_FINISH：在MindSpore框架中，ARF新启动的节点在训练进程完成初始化后抛出的异常。</li><li>RS_PREREPAIR_FINISH：ARF新启动的节点抛出的异常。</li><li>RS_STEP_FINISH：亚健康热切中step级暂停已经完成抛出的异常。</li></ul>|<ul><li>RS_NORMAL.value：ttp_c2python_api.ReportState_RS_NORMAL。</li><li>RS_UCE.value：ttp_c2python_api.ReportState_RS_UCE。</li><li>RS_UCE_CORRUPTED：ttp_c2python_api.ReportState_RS_UCE_CORRUPTED。</li><li>RS_HCCL_FAILED.value: ttp_c2python_api.ReportState_RS_HCCL_FAILED。</li><li>RS_UNKNOWN.value：ttp_c2python_api.ReportState_RS_UNKNOWN。</li><li>RS_INIT_FINISH：ttp_c2python_api.ReportState_RS_INIT_FINISH。</li><li>RS_PREREPAIR_FINISH.value：ttp_c2python_api.ReportState_RS_PREREPAIR_FINISH。</li><li>RS_STEP_FINISH：ttp_c2python_api.ReportState_RS_STEP_FINISH。</li></ul>|

**返回值**

无

## RepairType

**接口功能**

定义修复类型枚举。

**接口格式**

```python
mindio_ttp.framework_ttp.RepairType
```

**接口参数**

|参数|是否必选|说明|取值要求|
|--|--|--|--|
|RepairType|必选|区分修复类型：<ul><li>RT_SEND：备份卡发送数据。</li><li>RT_UCE_HIGHLEVEL：故障卡需要优化器和模型重建。</li><li>RT_UCE_LOWLEVEL：故障卡不需要优化器和模型重建。</li><li>RT_ROLLBACK：回滚数据集。</li><li>RT_RECV_REPAIR：ARF新拉起卡接收数据。</li><li>RT_LOAD_CKPT：周期Checkpoint数据修复。</li><li>RT_LOAD_REBUILD：重建模型优化器周期Checkpoint数据修复。</li></ul>|<ul><li>RT_SEND.value：ttp_c2python_api.RepairType_RT_SEND。</li><li>RT_UCE_HIGHLEVEL.value：ttp_c2python_api.RepairType_RT_UCE_HIGHLEVEL。</li><li>RT_UCE_LOWLEVEL.value：ttp_c2python_api.RepairType_RT_UCE_LOWLEVEL。</li><li>RT_ROLLBACK.value：ttp_c2python_api.RepairType_RT_ROLLBACK。</li><li>RT_RECV_REPAIR.value：ttp_c2python_api.RepairType_RT_RECV_REPAIR。</li><li>RT_LOAD_CKPT.value：ttp_c2python_api.RepairType_RT_LOAD_CKPT。</li><li>RT_LOAD_REBUILD.value：ttp_c2python_api.RepairType_RT_LOAD_REBUILD。</li></ul>|

**返回值**

无
