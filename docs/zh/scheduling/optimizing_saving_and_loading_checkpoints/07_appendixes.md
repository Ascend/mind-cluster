# 附录

## （可选）使用DPC文件访问存储，加速Checkpoint加载

检查是否满足如下条件：

- 是否使用DPC（Distributed Parallel Client，分布式并行客户端）文件系统访问存储。
- 是否成功安装NDS 1.0软件包（/opt/oceanstor/dataturbo/sdk/lib/libdpc\_nds.so）。
- 训练进程（如果在容器内）能否访问此so。

如果以上条件全部满足，则自动启用NDS 1.0直通读功能，加速加载Checkpoint。

成功加载NDS 1.0的判断依据是查看日志是否出现如下字样：

```text
"initial and open nds file driver success"
```

NDS 1.0更多信息请参见[《OceanStor DataTurbo 25.x.x DTFS用户指南》](https://support.huawei.com/enterprise/zh/doc/EDOC1100539415/3f076df0)。

> [!CAUTION]注意
> 如果使用DPC文件系统访问存储，成功安装NDS 1.0软件包，安装地址为“/opt/oceanstor/dataturbo/sdk/lib/libdpc\_nds.so”，权限设置为444即可保证功能正常，启动训练前，请用户谨慎设置此文件的权限。

## 环境变量

|参数名称|参数说明|取值范围|缺省值|
|--|--|--|--|
|MINDIO_AUTO_PATCH_MEGATRON|是否在import mindio_acp的时候自动patch Megatron框架的源代码中的Checkpoint相关函数。|<ul><li>true或者1：开启</li><li>其他值：关闭</li></ul>|false|
|HCOM_FILE_PATH_PREFIX|HCOM生成的文件路径的前缀，通过前缀保证文件只会在当前路径下（此路径需要已存在）创建和删除。|路径参数|${install_path}|

## 设置用户有效期

为保证用户的安全性，应设置用户的有效期，使用系统命令chage来设置用户的有效期。

命令为：

```bash
chage [-m mindays] [-M maxdays] [-d lastday] [-I inactive] [-E expiredate] [-W warndays] user
```

相关参数请参见[表1](#table_acp_04)。

**表 1<a id="table_acp_04"></a>**  设置用户有效期

|参数|参数说明|
|--|--|
|-d<br>--lastday|上一次更改的日期。|
|-E<br>--expiredate|用户到期的日期。超过该日期，此用户将不可用。|
|-h<br>--help|显示命令帮助信息。|
|-i<br>--iso8601|更改用户密码的过期日期并以YYYY-MM-DD格式显示。|
|-I<br>--inactive|停滞时期。超过指定天数后，设定密码为失效状态。|
|-l<br>--list|列出当前的设置。由非特权用户来确定口令或账户何时过期。|
|-m<br>--mindays|口令可更改的最小天数。设置为“0”表示任何时候都可以更改口令。|
|-M<br>--maxdays|口令保持有效的最大天数。设置为“-1”表示可删除这项口令的检测。设置为“99999”，表示无限期。|
|-R<br>--root|将命令执行的根目录设置为指定目录。|
|-W<br>--warndays|用户口令到期前，提前收到警告信息的天数。|

> [!NOTE]说明
>
> - 日期格式为YYYY-MM-DD，如 **chage -E 2017-12-01  _test_** 表示用户 **_test_** 的口令在2017年12月1日过期。
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
|*{MindIO-install-user}*|MindIO ACP安装用户。|用户自定义。|使用 **passwd** 命令修改。|

> [!CAUTION]注意
> 为了保护密码安全性建议用户定期修改密码。

## 公网网址说明

以下表格中列出了产品中包含的公网网址，没有安全风险。

|网址|说明|
|--|--|
|`http://www.apache.org/licenses/LICENSE-2.0`|该网站是开源许可证的发布地址，用于代码版权信息声明。由于系统中无对外交互场景，因此无安全风险。|
