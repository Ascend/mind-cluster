# ockio

### 简介
MindCluster MindIO Async CheckPoint Persistence（下文简称`MindIO ACP`）加速大模型CheckPoint功能主要针对大模型训练中的CheckPoint的保存及加载进行加速，CheckPoint的数据先写入训练服务器的内存系统中，再异步写入后端的可靠性存储设备中。

#### 软件架构
![architecture](./doc/architecture.JPG)

### 编译
1. 下载代码
```
git clone https://gitcode.com/Ascend/mind-cluster.git

cd mind-cluster/component/mindio/acp/build
```

2. 编译

支持直接执行如下脚本编译
```
bash build.sh

build.sh支持5种参数：build.sh [ -h | --help ] [ -t | --type <build_type> ] [ -b | --builddir <build_path> ] [ -f | --flags <cmake_flags> ]
help:显示使用指导
type:编译类型,可填[debug, release]。
builddir:指定编译目录
cmake_flags:传递给cmake的自定义标志（这些参数必须出现在所有其他参数之后）。
不填入参数情况下,默认执行build.sh -t release
```

3. ut运行

支持直接执行如下脚本编译并运行ut
```
bash script/run_ut.sh
```

### 源码目录结构
`MindIO ACP`源码的主要目录结构如下所示。
```
mind-cluster
└── component
    └── mindio
        └── acp
            ├── 3rdparty  --------------------  第三方依赖库
            ├── build
            │   └── build.sh  ----------------  编译脚本
            ├── CMakeLists.txt  --------------  cmake文件
            ├── cmake_modules  ---------------  cmake文件
            ├── configs  ---------------------  默认配置文件
            ├── doc  -------------------------  文档目录
            ├── python_whl  ------------------	python源代码
            ├── README.md  -------------------	README
            ├── scripts  ---------------------  ci脚本和安装部署脚本
            ├── src  -------------------------  源代码目录
            └── test  ------------------------  单元测试和fuzz测试
```

### 安装使用
`MindIO ACP`编译生成.whl安装包供用户使用，whl包格式为 ```mindio_acp-${mindio_acp_version}-py3-none-linux_${arch}.whl```

其中，mindio_acp_version表示`MindIO ACP`的版本；arch表示架构，如x86或aarch64
参考安装命令如下
```
pip3 install mindio_acp-${mindio_acp_version}-py3-none-linux_${arch}.whl --force-reinstall
```

### 用户指南
`MindIO ACP`提供给开发者的的资料如下：

[《MindIO ACP用户指南》](https://www.hiascend.com/document/detail/zh/mindx-dl/600/clusterscheduling/ref/mindioacp/mindioacp001.html)

请根据用户指南了解`MindIO ACP`相关约束限制，进行安装，使用，管理与加固。

### 源码内公网地址
 
| 类型   | 开源代码地址      | 文件名      | 公网IP地址/公网URL地址/域名/邮箱地址 | 用途说明            |
|------  |-----------------|-------------|---------------------               |-------------------|
| 代码仓地址  | https://gitee.com/openeuler/libboundscheck.git | .gitmodules | https://gitee.com/openeuler/libboundscheck.git | 依赖三方库 |
| 代码仓地址  | https://github.com/gabime/spdlog.git | .gitmodules | https://github.com/gabime/spdlog.git | 依赖三方库 |
| 代码仓地址  | https://gitee.com/openeuler/ubs-comm.git | .gitmodules | https://gitee.com/openeuler/ubs-comm.git | 依赖三方库 |
| license 地址 | 不涉及 | LICENSE | http://www.apache.org/licenses/LICENSE-2.0 | license文件 |