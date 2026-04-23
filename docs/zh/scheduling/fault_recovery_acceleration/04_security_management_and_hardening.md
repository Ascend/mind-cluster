# 安全管理与加固

## 安全管理

> [!NOTE]说明 
> MindIO TFT暂不支持公有云场景、多租户场景使用，不支持公网直接访问系统。

**防病毒软件例行检查**

定期开展对集群的防病毒扫描，防病毒例行检查会帮助集群免受病毒、恶意代码、间谍软件以及恶意程序侵害，降低系统瘫痪、信息安全问题等风险。可以使用业界主流防病毒软件进行防病毒检查。

**日志管理**

日志管理需要关注以下两点。

- 检查系统是否可以限制单个日志文件的大小。
- 检查日志空间占满后，是否存在机制进行清理。

**漏洞/功能问题修复**

为保证生产环境的安全，降低被攻击的风险，需要定期查看开源社区修复的以下漏洞/功能问题。

- 操作系统漏洞/功能问题。
- 其他相关组件漏洞/功能问题。

## 安全加固

### 加固须知

本文中列出的安全加固措施为基本的加固建议项。用户应根据自身业务，重新审视整个系统的网络安全加固措施，必要时可参考业界优秀加固方案和安全专家的建议。

### 风险提示

Checkpoint序列化过程中使用了torch.load接口，该接口中使用了Python自带的pickle组件，必须确保非授权用户没有存储目录及上层目录的写权限，需保证Checkpoint为可信数据，否则可能造成Checkpoint被篡改引起pickle反序列化注入的风险。

### 操作系统安全加固

**防火墙配置**

操作系统安装后，若配置普通用户，可以通过在“/etc/login.defs”文件中新增“ALWAYS\_SET\_PATH=yes”配置，防止越权操作。此外，为了防止使用“su”命令切换用户时，将当前用户环境变量带入其他环境造成提权，请使用 **su - [user]** 命令进行用户切换，同时在服务器配置文件“/etc/default/su”中增加配置参数“ALWAYS\_SET\_PATH=yes”防止提权。

**设置umask**

建议用户将服务器的umask设置为027\~777以限制文件权限。

以设置umask为027为例，具体操作如下。

1. 以root用户登录服务器，编辑“/etc/profile”文件。

    ```bash
    vim /etc/profile
    ```

2. 在“/etc/profile”文件末尾加上 **umask 027**，保存并退出。
3. 执行如下命令使配置生效。

    ```bash
    source /etc/profile
    ```

**无属主文件安全加固**

用户可以执行 **find / -nouser -nogroup** 命令，查找容器内或物理机上的无属主文件。根据文件的UID和GID创建相应的用户和用户组，或者修改已有用户的UID、用户组的GID来适配，赋予文件属主，避免无属主文件给系统带来安全隐患。

**端口扫描**

用户需要关注全网侦听的端口和非必要端口，如有非必要端口请及时关闭。建议用户关闭不安全的服务，如Telnet、FTP等，以提升系统安全性。具体操作方法可参考所使用操作系统的官方文档。

**防DoS攻击**

用户可以根据IP地址限制与服务器的连接速率对系统进行防DoS攻击，方法包括但不限于利用Linux系统自带Iptables防火墙进行预防、优化sysctl参数等。具体使用方法，用户可自行查阅相关资料。

**SSH加固**

由于root用户拥有最高权限，出于安全目的，建议取消root用户SSH远程登录服务器的权限，以提升系统安全性。具体操作步骤如下：

1. 登录安装MindIO TFT组件的节点。
2. 打开“/etc/ssh/sshd\_config”文件。

    ```bash
    vim /etc/ssh/sshd_config
    ```

3. 按“i”进入编辑模式，找到“PermitRootLogin”配置项并将其值设置为“no”。

    ```text
    PermitRootLogin no
    ```

4. 按“Esc”键，输入 **:wq!**，按“Enter”保存并退出编辑。
5. 执行命令使配置生效。

    ```bash
    systemctl restart sshd
    ```

**缓冲区溢出安全保护**

为阻止缓冲区溢出攻击，建议使用ASLR（Address Space Layout Randomization，内存地址随机化机制）技术，通过对堆、栈、共享库映射等线性区布局的随机化，增加攻击者预测目的地址的难度，防止攻击者直接定位攻击代码位置。该技术可作用于堆、栈、内存映射区（mmap基址、shared libraries、vdso页）。

开启方式：

```bash
echo 2 >/proc/sys/kernel/randomize_va_space
```

## 开启TLS认证

- 为了保障MindIO TFT组件内部Controller和Processor之间的通信安全，保护信息不被篡改、仿冒，建议启用TLS加密。
- TLS加密仅用于MindIO TFT内部模块间通信，不对外提供TLS接入、认证功能。
- 因为开启安全认证依赖OpenSSL组件，所以建议用户使用OpenSSL无漏洞版本，需要配套使用GLIBC 2.33或更高版本。

### 导入TLS证书

- 通过接口tft\_start\_controller、tft\_init\_processor配置TLS密钥证书等，进行TLS安全连接，安全选项默认开启，建议用户开启TLS加密配置，以保证通信安全，如需关闭加密功能，可以使用下面示例，调用接口进行关闭。
- 系统启动后，建议删除本地密钥证书等敏感信息文件。
- 调用该接口时，传入的文件路径应避免包含英文分号、逗号、冒号。
- 支持通过环境变量 **TTP\_ACCLINK\_CHECK\_PERIOD\_HOURS** 和 **TTP\_ACCLINK\_CERT\_CHECK\_AHEAD\_DAYS** 配置证书检查周期与证书过期预警时间。

**配置TLS接口调用示例**

- TLS关闭（**enable\_tls**=False）时，**tls\_info**无效，无需配置。此开关不影响MindIO TFT特性功能。

    ```python
    from mindio_ttp.framework_ttp import tft_start_controller, tft_init_processor
    
    tft_start_controller(bind_ip: str, port: int, enable_tls=False, tls_info='')
    tft_init_processor(rank: int, world_size: int, enable_local_copy: bool, enable_tls=False, tls_info='', enable_uce=True, enable_arf=False)
    ```

    > [!CAUTION]注意
    > - 如果关闭TLS（即**enable\_tls**=False时），会存在较高的网络安全风险。
    > - **tft\_start\_controller** 和 **tft\_init\_processor** 的enable\_tls开关状态需要保持一致。若两个接口enable\_tls开关不同，会造成以下问题：
    >   - 模块间TLS建链失败。
    >   - MindIO TFT无法正常运行，训练任务启动失败。

- TLS开启（**enable\_tls**=True）时，证书相关信息，作为必选参数 **tls\_info** 用于如下接口：

    ```python
    from mindio_ttp.framework_ttp import tft_start_controller, tft_init_processor, tft_register_decrypt_handler
    
    # 在tls_info中，以“;”分隔不同字段,以“,”分隔各个文件
    tls_info = r"(
    tlsCert: /etc/ssl/certs/cert.pem;
    tlsCrlPath: /etc/ssl/crl/;
    tlsCaPath: /etc/ssl/ca/;
    tlsCaFile: ca_cert_1.pem, ca_cert_2.pem;
    tlsCrlFile: crl_1.pem, crl_2.pem;
    tlsPk: private key;
    tlsPkPwd: private key pwd;
    packagePath: /etc/ssl/
    )"
    
    # 若tlsPkPwd口令为密文，则需注册口令解密函数
    tft_register_decrypt_handler(user_decrypt_callback)
    tft_start_controller(bind_ip: str, port: int, enable_tls=True, tls_info=tls_info)
    tft_init_processor(rank: int, world_size: int, enable_local_copy: bool, enable_tls=True, tls_info=tls_info, enable_uce=True, enable_arf=False)
    ```

**tls\_info中各字段含义**

|字段|含义|Required|
|--|--|--|
|tlsCert|Server证书。|是|
|tlsCaPath|CA证书存储路径。|是|
|tlsCaFile|CA证书列表。|是|
|tlsCrlPath|证书吊销列表存储路径。|否|
|tlsCrlFile|证书吊销列表。|否|
|tlsPk|私钥。|是|
|tlsPkPwd|私钥口令。|是|
|packagePath|OpenSSL库路径|是|

> [!CAUTION]注意
> 证书安全要求：
>
> - 需使用业界公认安全可信的非对称加密算法、密钥交换算法、密钥长度、Hash算法、证书格式等。
> - 应处于有效期内。

### （可选）证书有效性校验

如果启用TLS认证，则需要关注证书有效期。请合理规划证书有效期和证书更新周期，并在证书过期前及时更新证书，防范安全风险。MindIO TFT提供证书有效期定期巡检功能，默认巡检周期为7天，默认提前告警时间为30天，若发现证书存在过期风险，则会在环境变量 **TTP\_LOG\_PATH** 配置的日志中打印WARNING告警信息，请及时关注并处理。
