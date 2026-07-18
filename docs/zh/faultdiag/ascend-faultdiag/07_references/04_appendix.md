# 附录

## 版本号格式

ascend-fd 版本号格式为 `X.Y.Z`，例如 `26.1.0`。

## 默认路径汇总

| 路径                                                 | 说明                 |
|------------------------------------------------------|----------------------|
| `~/.ascend_faultdiag/`                               | 日志和配置文件根目录 |
| `~/.ascend_faultdiag/RUN_LOG/`                       | 运行日志目录         |
| `~/.ascend_faultdiag/ascend_faultdiag_operation.log` | 操作日志             |
| `~/.ascend_faultdiag/install.log`                    | 安装日志             |
| `~/.ascend_faultdiag/config.json`                    | 配置文件             |
| `~/.ascend_faultdiag/custom-ascend-kg-config.json`   | 自定义故障实体文件   |
| `~/.ascend_faultdiag/blacklist-config.json`          | 屏蔽规则配置文件     |

> [!NOTE]
>
> - 日志文件大小不超过 10MB，超过后自动转储。

## 组件错误码

| 状态码                 | 含义             |
|------------------------|------------------|
| 500 BaseError          | 基础错误         |
| 501 PathError          | 无效的输入路径   |
| 502 FileNotExistError  | 文件不存在       |
| 503 InfoNotFoundError  | 查找的信息未找到 |
| 504 InfoIncorrectError | 信息不正确       |
| 505 FileOpenError      | 打开文件失败     |
| 506 InnerError         | 服务内部错误     |
| 507 ParamError         | 参数错误         |
| 508 FileTooLarge       | 文件数量过多     |
| 200 SuccessRet         | 操作成功         |

## 已支持故障

请参考 [MindCluster 26.1.0 故障诊断类型](https://raw.gitcode.com/Ascend/mind-cluster/blobs/c49ed595d1b87b34eaf020591778c4536c34b2fe/%E6%95%85%E9%9A%9C%E8%AF%8A%E6%96%AD%E7%B1%BB%E5%9E%8B.xlsx)。
