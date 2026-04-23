# 附录

## 环境变量

> [!NOTE]说明
> 加粗显示的环境变量为常用环境变量。

|参数名称|参数说明|取值范围|缺省值|
|--|--|--|--|
|**TTP_LOG_PATH**|MindIO TFT日志路径。禁止配置软链接，日志文件名补充为ttp_log.log，建议日志路径中包含日期时间，避免多次训练记录在同一个日志中，造成循环覆写。推荐在训练启动脚本中按如下方式配置日志路径： <br> `date_time=\$(date +%Y-%m-%d-%H_%M_%S)` <br> `export TTP_LOG_PATH=logs/\${date_time}` <br>当使用共享存储时，建议按照节点配置日志路径：<br>`export TTP_LOG_PATH=logs/\${nodeId}`|文件夹路径。|logs|
|**TTP_LOG_LEVEL**|MindIO TFT日志等级。<ul><li>DEBUG：细节信息，仅当诊断问题时适用。</li><li>INFO：确认程序按预期运行。</li><li>WARNING：表明有已经或即将发生的意外。程序仍按预期进行。</li><li>ERROR：由于严重的问题，程序的某些功能已经不能正常执行。</li></ul>|<ul><li>DEBUG</li><li>INFO</li><li>WARNING</li><li>ERROR|INFO</li></ul>|
|TTP_LOG_MODE|MindIO TFT日志模式。<ul><li>ONLY_ONE：所有MindIO TFT进程写一个日志。</li><li>PER_PROC：每个MindIO TFT进程写独立日志，日志文件路径为 {TTP_LOG_PATH}/ttp_log.log.{pid}。</li></ul>|<ul><li>ONLY_ONE</li><li>PER_PROC（若非指定ONLY_ONE，则默认为PER_PROC）</li></ul>|PER_PROC|
|TTP_LOG_STDOUT|MindIO TFT日志记录方式。<ul><li>0：将MindIO TFT运行日志记录到对应的日志文件中。</li><li>1：直接打印MindIO TFT运行日志，不在本地存储。</li></ul>|<ul><li>0</li><li>1</li></ul>|0|
|MASTER_ADDR|训练主节点IP地址或域名。|合法的IPv4、IPv6地址或域名。|-|
|MASTER_PORT|训练主节点通信端口，端口可配。|[1024, 65535]|-|
|TTP_RETRY_TIMES|Processor TCP（Transmission Control Protocol）建链尝试次数。|[1, 300]|10|
|MINDIO_WAIT_MINDX_TIME|Controller等待MindCluster响应的最大时间，单位：s。|[1, 3600]|30|
|TTP_ACCLINK_CHECK_PERIOD_HOURS|开启TLS认证后，MindIO TFT检查证书有效性的周期，单位：h。|[24, 720]|168|
|TTP_ACCLINK_CERT_CHECK_AHEAD_DAYS|开启TLS认证后，MindIO TFT检查证书过期日提前告警的时长，单位：天。需满足证书过期提前告警时长不小于巡检周期，保证及时发现证书过期风险并告警。|[7, 180]，且需满足TTP_ACCLINK_CERT_CHECK_AHEAD_DAYS * 24 ≥ TTP_ACCLINK_CHECK_PERIOD_HOURS。|30|
|TTP_NORMAL_ACTION_TIME_LIMIT|故障恢复流程中，执行rebuild/repair/rollback回调函数的超时时间，单位：s。|[30, 1800]|180|
|MINDIO_FOR_MINDSPORE|表示是否启用MindSpore开关，传入True（不区分大小写）或1时，开启MindSpore开关，其他值关闭MindSpore开关。|<ul><li>True（不区分大小写）或1：启用MindSpore。</li><li>其他：关闭MindSpore。</li></ul>|False|
|MINDX_TASK_ID|MindIO ARF特性使用，MindCluster任务ID，由ClusterD配置，无需用户干预。|字符串。|-|
|TORCHELASTIC_USE_AGENT_STORE|PyTorch环境变量，控制创建TCP Store Server还是Client，MindIO TFT在临终Checkpoint保存且Torch Agent TCP Store Server连接失败场景下使用。|<ul><li>True：创建Client。</li><li>False：创建Server。</li></ul>|-|
|TTP_STOP_CLEAN_BEFORE_DUMP|MindIO TFT特性使用，控制MindIO TTP在保存临终Checkpoint前是否做stop&clean操作。|<ul><li>0：关闭临终前stop&clean操作。</li><li>1：启用临终前stop&clean操作。</li></ul>|0|

## 设置用户有效期

为保证用户的安全性，应设置用户的有效期，使用系统命令 **chage** 来设置用户的有效期。

命令为：

```bash
chage [-m mindays] [-M maxdays] [-d lastday] [-I inactive] [-E expiredate] [-W warndays] user
```

相关参数请参见[表1](#table_tft_18)。

**表 1<a id="table_tft_18"></a>**  设置用户有效期

|参数|参数说明|
|--|--|
|-d<br>--lastday|上一次更改的日期。|
|-E<br>--expiredate|用户到期的日期。超过该日期，此用户将不可用。|
|-h<br>--help|显示命令帮助信息。|
|-i<br>--iso8601|更改用户密码的过期日期并以YYYY-MM-DD格式显示。|
|-I<br>--inactive|停滞时期。过期指定天数后，设定密码为失效状态。|
|-l<br>--list|列出当前的设置。由非特权用户来确定口令或账户何时过期。|
|-m<br>--mindays|口令可更改的最小天数。设置为“0”表示任何时候都可以更改口令。|
|-M<br>--maxdays|口令保持有效的最大天数。设置为“-1”表示可删除这项口令的检测。设置为“99999”，表示无限期。|
|-R<br>--root|将命令执行的根目录设置为指定目录。|
|-W<br>--warndays|用户口令到期前，提前收到警告信息的天数。|

> [!NOTE]说明
>
> - 日期格式为YYYY-MM-DD，如 **chage -E 2017-12-01 _test_** 表示用户 **_test_** 的口令在2017年12月1日过期。
> - user必须填写，填写时请替换为具体用户，默认为root用户。
> - 账号口令应该定期更新，否则容易导致安全风险。

举例说明：修改用户 **_test_** 的有效期为90天。

```bash
chage -M 90 test
```

## 口令复杂度要求

口令至少满足如下要求：

1. 口令长度至少8个字符。
2. 口令必须包含如下至少两种字符的组合：
    - 一个小写字母
    - 一个大写字母
    - 一个数字
    - 一个特殊字符：\`\~!@\#$%^&\*\(\)-\_=+\\|[\{\}];:'",<.\>/?和空格

3. 口令不能和账号一样。

## 账户一览表

|用户|描述|初始密码|密码修改方法|
|--|--|--|--|
| *{MindIO-install-user}* |MindIO TFT安装用户。|用户自定义。|使用 **passwd** 命令修改。|
