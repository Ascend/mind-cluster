# 常用操作<a name="ZH-CN_TOPIC_0000001829748104"></a>

## 自定义MindCluster Ascend FaultDiag家目录<a name="ZH-CN_TOPIC_0000001876627541"></a>

支持通过环境变量“ASCEND\_FD\_HOME\_PATH”设置MindCluster Ascend FaultDiag的家目录。该目录用于落盘MindCluster Ascend FaultDiag的运行日志、操作日志文件；屏蔽故障关键词文件和用户自定义的故障文件也存储在该目录下。

**操作步骤<a name="section78581144191111"></a>**

执行以下命令，在环境变量中，设置MindCluster Ascend FaultDiag家目录。

```
export ASCEND_FD_HOME_PATH=/自定义路径
```

>[!NOTE] 说明 
>-   当未设置该环境变量时，MindCluster Ascend FaultDiag家目录默认为“$HOME/.ascend\_faultdiag/”。
>-   当设置该环境变量时，所设置路径需要存在且为目录，且不支持指定“/tmp”路径为家目录。目录属主需要为root或程序执行者，且在该目录下程序执行者具有创建文件、读写文件的权限。


