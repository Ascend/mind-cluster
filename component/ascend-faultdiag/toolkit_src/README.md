# MindCluster 链路故障诊断工具

**MindCluster Ascend FaultDiag Toolkit** 提供昇腾 AI 集群链路故障诊断能力的轻量工具，支持从 PC 或单台服务器远程访问集群设备采集数据诊断。

## 📢 变更通知

- **2026-06-05**: ✨ 诊断报告表格展示优化
- **2026-06-05**: ✨ 新增链路精准定位
- **2026-03-06**: 🚀 提供链路故障诊断分析能力

## 详细文档

完整使用文档请查看：[docs/zh/faultdiag/ascend-faultdiag-toolkit](../../../docs/zh/faultdiag/ascend-faultdiag-toolkit/menu_ascend-faultdiag-toolkit.md)

## 目录结构

```text
toolkit_src/
├── ascend_fd_tk/                  # 工具主程序包
│   ├── cli.py                     # 工具入口（ascend-fd-tk 命令）
│   ├── conn.ini                   # 连接配置示例文件
│   ├── core/                      # 核心模块
│   │   ├── cli_module/            # CLI 模型与命令注册
│   │   ├── collect/               # 数据采集（SSH / 离线日志 / 解析器）
│   │   ├── common/                # 公共常量、枚举、路径、异常
│   │   ├── config/                # 阈值/位置/连接等配置文件
│   │   ├── context/               # 诊断上下文与采集注册
│   │   ├── crypto/                # 连接配置加密
│   │   ├── fault_analyzer/        # 故障分析器（主机/BMC/交换机/HCCS）
│   │   ├── inspection/            # 巡检规则与配置
│   │   ├── log_parser/            # 离线日志解析配置
│   │   ├── model/                 # 领域模型（Host/BMC/Switch/Optical）
│   │   ├── report/                # 报告生成（多 Sheet Excel）
│   │   ├── root_cause/            # 根因分析（滤波器、SNR 检查）
│   │   └── service/               # 业务服务（auto_diag/auto_collect 等）
│   ├── examples/                  # 常用脚本与示例
│   │   ├── auto_diag/             # linux/windows 一键式采集/诊断脚本
│   │   ├── cmd/                   # 采集命令示例
│   │   ├── inspection/            # 巡检示例
│   │   ├── loopback_diag/         # 光模块环回诊断示例
│   │   └── scripts/               # clear_cache / 辅助脚本
│   ├── utils/                     # 工具目录
│   └── test/                      # 单元测试
├── setup.py                       # Python 包打包入口
├── requirements.txt               # Python 依赖清单
├── build_cli_exe.bat              # 打包脚本
├── ascend_faultdiag_toolkit.spec  # 打包脚本
└── README.md                      # 工具说明
```

## 分支维护策略

| 状态 | 时间 | 说明 |
|------|------|------|
| 计划 | 1-3 个月 | 计划特性 |
| 开发 | 3 个月 | 开发新特性并修复问题，定期发布新版本 |
| 维护 | 3-12 个月 | 常规分支维护 3 个月，长期支持分支维护 12 个月。对重大 BUG 进行修复，不合入新特性 |
| 生命周期终止（EOL） | N/A | 分支不再接受任何修改 |

## 免责声明

- 本仓库代码中包含多个开发分支，这些分支可能包含未完成、实验性或未测试的功能。在正式发布前，这些分支不应被应用于任何生产环境或者依赖关键业务的项目中。请务必使用我们的正式发行版本，以确保代码的稳定性和安全性。
- 正式版本请参考 [release 版本](https://gitcode.com/ascend/mind-cluster/releases)。

## License

MindCluster 以 Apache 2.0 许可证许可，对应许可证文本可查阅 [MindCluster 根目录](https://gitcode.com/Ascend/mind-cluster/blob/master/LICENSE)。

## 建议与交流

欢迎大家为社区做贡献。如果有任何疑问或建议，请提交 [issue](https://gitcode.com/Ascend/mind-cluster/issues)，我们会尽快回复。感谢您的支持。
