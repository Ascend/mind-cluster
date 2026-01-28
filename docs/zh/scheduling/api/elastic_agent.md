# Elastic-Agent（断点续训相关接口）<a name="ZH-CN_TOPIC_0000002479386784"></a>

>[!NOTE] 说明 
>-   Elastic Agent组件已经日落，相关资料将于2026年的8.3.0版本删除。
>-   内部接口严禁调用。

## mindx\_elastic.\_\_version\_\_<a name="ZH-CN_TOPIC_0000002511346763"></a>

获取Elastic Agent版本号。

输入值：空

返回值：Elastic Agent版本号

使用样例如下：

```
import mindx_elastic
mindx_elastic.__version__
```


## mindx\_elastic.api.patch\_torch\_methods（内部接口，严禁调用）<a name="ZH-CN_TOPIC_0000002479226826"></a>

**功能说明<a name="section1222112260226"></a>**

在构建训练镜像安装Elastic Agent时，使用**sed -i '/import os/i import mindx\_elastic.api' $\(pip3.7 show torch | grep Location | awk -F ' ' '\{print $2\}'\)/torch/distributed/run.py**命令后，Elastic Agent组件会在导入Torch的Elastic模块时执行mindx\_elastic.api.patch\_torch\_methods接口，使patch自动生效。

Elastic Agent组件会对torch.distributed.elastic.agent.server.api.SimpleElasticAgent.\_invoke\_run、torch.distributed.launcher.api.launch\_agent、torch.distributed.elastic.agent.server.api.SimpleElasticAgent.\_initialize\_workers等方法打patch，额外提供昇腾NPU设备的故障检测与恢复功能。


## mindx\_elastic.recover\_manager.DLRecoverManager（内部接口，严禁调用）<a name="ZH-CN_TOPIC_0000002511346787"></a>

DLRecoverManager类提供进程级恢复和进程级在线恢复相关接口。客户端以Python包形式import到客户端代码中。

>[!NOTE] 说明 
>DLRecoverManager类提供的接口可能抛出Exception异常，调用方自行捕获异常、处理异常。

**\_\_init\_\_\(self, info: pb.ClientInfo, server\_addr: str, secure\_conn: bool = True, cert\_path: str = ""\)<a name="section93535281517"></a>**

构造DLRecoverManager，用于后续的通信。

**表 1**  参数说明

|参数|类型|说明|
|--|--|--|
|info|pb.ClientInfo|<p>info.ip：str类型，客户端IP（暂未使用，预留）。</p><p>info.port：str类型，客户端端口（暂未使用，预留）。</p><p>info.taskId：str类型，任务ID。</p><p>info.role：str类型，客户端角色。</p>|
|server_addr|str|服务端地址|
|secure_conn|bool|是否开启安全连接，默认为True。|
|cert_path|str|安全证书地址，默认为""。|


**register\(self, request: pb.ClientInfo\) -\> pb.Status<a name="section92911329181515"></a>**

注册客户端，服务端为request指定的任务做恢复前的初始化操作。

**表 2**  参数说明

|参数|类型|说明|
|--|--|--|
|request|pb.ClientInfo|<p>request.ip：str类型，客户端IP（暂未使用，预留）。</p><p>request.port：str类型，客户端端口（暂未使用，预留）。</p><p>request.taskId：str类型，任务ID。</p><p>request.role：str类型，客户端角色。</p>|


**表 3**  返回值说明

|返回值类型|说明|
|--|--|
|Status|<p>Status.info：str类型，返回信息描述</p><p>Status.code：int类型，0表示成功，其他值表示失败。关于返回码的详细说明请参见<a href="#返回码说明">返回码说明</a>。</p>|


**start\_subscribe\(self\)<a name="section5051271214"></a>**

客户端和服务端建立gRPC长链接，服务端将通过该长链接与客户端单向通信。比如发生故障时，服务端给客户端发送停止训练、全局故障rank信息等。

**init\_clusterd\(self\)<a name="section18270133519256"></a>**

客户端初始化ClusterD服务端状态，保证后续任务正常注册、建立链接。


## report\_stop\_complete\(code: int, msg: str, fault\_ranks: dict\) -\> int<a name="ZH-CN_TOPIC_0000002511426697"></a>

客户端给服务端上报任务的进程停止完成。一般是在客户端收到服务端的停止训练信号后，客户端停止训练任务的进程，然后给服务端上报任务的进程停止完成。

**表 1**  参数说明

|参数|类型|说明|
|--|--|--|
|code|int|状态码|
|msg|str|返回信息|
|fault_ranks|dict|故障进程Rank|


**表 2**  返回值说明

|返回值类型|说明|
|--|--|
|int|0表示成功，其他返回值表示失败。关于返回码的详细说明请参见<a href="#返回码说明">返回码说明</a>。|



## report\_recover\_strategy\(fault\_ranks: dict, strategy\_list: list\) -\> int<a name="ZH-CN_TOPIC_0000002511346757"></a>

客户端给服务端上报客户端支持的恢复策略，供服务端选择最佳恢复策略，服务端再通过start\_subscribe构建的长链接下发给客户端。

**表 1**  参数说明

|参数|类型|说明|
|--|--|--|
|fault_ranks|dict|故障进程Rank|
|strategy_list|list|恢复策略列表|


**表 2**  返回值说明

|返回值类型|说明|
|--|--|
|int|0表示成功，其他返回值表示失败。关于返回码的详细说明请参见<a href="#返回码说明">返回码说明</a>。|



## report\_recover\_status\(code: int, msg: str, fault\_ranks: dict, strategy: str\) -\> int<a name="ZH-CN_TOPIC_0000002511426757"></a>

客户端给服务端上报任务恢复状态。

**表 1**  参数说明

|参数|类型|说明|
|--|--|--|
|code|int类型|状态码|
|msg|str|返回信息|
|fault_ranks|dict|故障进程Rank|
|strategy|str|修复策略|


**表 2**  返回值说明

|返回值类型|说明|
|--|--|
|int|0表示成功，其他返回值表示失败。关于返回码的详细说明请参见<a href="#返回码说明">返回码说明</a>。|



## report\_process\_fault\(fault\_ranks: dict\) -\> int<a name="ZH-CN_TOPIC_0000002479386856"></a>

客户端上报任务进程业务面故障。客户端先发现故障时，给服务端上报业务面故障所在的rank的信息。

**表 1**  参数说明

|参数|类型|说明|
|--|--|--|
|fault_ranks|dict|故障进程Rank|


**表 2**  返回值说明

|返回值类型|说明|
|--|--|
|int|0表示成功，其他返回值表示失败。关于返回码的详细说明请参见<a href="#返回码说明">返回码说明</a>。|



## 返回码说明<a name="ZH-CN_TOPIC_0000002511426709"></a>

Elastic-Agent返回码如[表1](#table1248859202914)所示。

**表 1** Elastic-Agent返回码

<a name="table1248859202914"></a>
|返回码|值|含义|
|--|--|--|
|OK|0|接口调用正常。|
|UnRegistry|400|Job ID未注册。|
|OrderMix|401|请求不符合状态机顺序。|
|JobNotExist|402|Job ID不存在。|
|ProcessRescheduleOff|403|未打开进程级恢复开关。|
|ProcessNotReady|404|训练进程未拉起。|
|RecoverableRetryError|405|恢复失败，失败原因为clean device失败。|
|UnRecoverableRetryError|406|恢复失败，失败原因为stop device失败。|
|DumpError|407|临终遗言保存失败。|
|UnInit|408|未调用初始化。|
|ClientError|499|其他失败原因。|
|OutOfMaxServeJobs|500|超过最大服务任务数。|
|OperateConfigMapError|501|操作ConfigMap失败。|
|OperatePodGroupError|502|操作PodGroup失败。|
|ScheduleTimeout|503|Pod调度超时。|
|SignalQueueBusy|504|控制信号入队失败。|
|EventQueueBusy|505|状态机事件入队失败。|
|ControllerEventCancel|506|状态机已退出。|
|WaitReportTimeout|507|等待客户端调用接口超时。|
|WaitPlatStrategyTimeout|508|等待AI平台准备恢复策略超时。|
|WriteConfirmFaultOrWaitPlatResultFault|509|AI平台故障信息错误。|
|ServerInnerError|599|服务端内部错误。|



