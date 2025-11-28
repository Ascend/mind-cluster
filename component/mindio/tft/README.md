# mindio_tft

### 简介
MindCluster MindIO Training Fault Tolerance（MindIO TFT）包括临终CheckPoint保存、进程级在线恢复、进程级别重调度等功能。MindIO TFT面向Megatron-LM分布式训练场景，提供了ZeRO-1级别大模型并行容错能力，通过监控运行时训练集群各卡状态，由故障发生类型选择合适的修复策略，实现模型参数、优化器状态快速恢复，将MTTR时延由小时降低至分钟级，从故障类别角度提供以下能力：
- MindCluster MindIO Try To Persist（MindIO TTP）功能，主要针对大模型训练过程中故障恢复加速，MindIO TTP特性通过在训练过程中发生故障后，校验中间状态数据的完整性和一致性，生成一次临终CheckPoint数据，恢复训练时能够通过该CheckPoint数据恢复，减少故障造成的训练迭代损失。
- MindCluster MindIO Uncorrectable Memory Error（MindIO UCE）功能，主要是针对大模型训练过程中片上内存的UCE故障检测，并完成在线修复，达到Step级重计算。
- MindCluster MindIO Air Refuelling（MindIO ARF）功能，训练发生异常后，不用重新拉起整个集群，只需以节点为单位进行重启或替换，对于部分故障仅需原地重启单进程，完成修复并继续训练。
- MindIO TFT搭配MindCluster使用，此外还支持网络故障快速恢复、亚健康故障热切换、在线压测/借轨回切。

#### 软件架构
![architecture](./doc/architecture.JPG)

- Controller模块：负责分布式任务的协同，内部维护状态机，状态机支持不同场景的流程控制；实时收集各个训练进程的训练状态，当训练发生异常后，结合异常类型，触发状态机运作，将状态机对应的Action发送到Processor模块执行。
- Processor模块：负责与训练框架交互，获取训练进程的训练状态，向Controller汇报，同时负责执行Controller模块下发的对应Action动作。
- [Adaptor模块](https://gitcode.com/Ascend/MindSpeed-LLM/tree/master/mindspeed_llm/core/high_availability)：负责完成训练框架对MindIO TTP、MindIO UCE、MindIO ARF特性的适配。目前MindIO TFT已完成对MindSpeed-LLM训练框架的适配。对于其他训练框架，需用户参考并[自行适配](https://www.hiascend.com/document/detail/zh/mindx-dl/600/clusterscheduling/ref/mindiottp/mindiotft018.html)。

#### 约束限制
众多大模型框架都支持ZeRO（Zero Redundancy Optimizer，零冗余优化器）以减少对显存的使用，当前MindIO TFT仅支持开启ZeRO-1，支持DP（Data Parallelism，数据并行） Size为偶数，同时使用不同的功能对DP Size有不同的限制：
- MindIO TTP功能
  - 为了保证故障发生后，有完整的优化器状态数据，要求DP Size能被副本数整除。
  - 开启MoE（Mixture of Experts，混合专家结构）前要求稠密层DP Size大于1；开启MoE后要求稠密层和稀疏层DP Size都大于1。
  - 针对分布式优化器，MindIO TFT在ZeRO-1功能的基础上，通过以算代传，在DP Group上重新切分优化器ZeRO-1范围，实现了优化器数据副本。
- MindIO UCE / MindIO ARF功能
  - 若要实现从当前Step恢复训练，对DP Size限制与MindIO TTP功能一致。
  - 对于显存有限，不做副本的情况，即DP Size = 1，此时若发生UCE或者节点故障，支持在线从周期性CheckPoint中加载模型权重和优化器参数恢复训练，损失当前Step到上次周期性CheckPoint的Step之间的训练成本。

### 编译

mindio_tft编译不依赖MindCluster.

1. 下载代码
```
git clone https://gitcode.com/Ascend/mind-cluster.git

cd mind-cluster/component/mindio/tft
```

2. 编译执行

支持直接执行如下脚本编译
```
bash build/build.sh

build.sh支持3个参数，按顺序分别是<build_mode> <build_dir> <build_ut> <build_whl> <CMAKE_FLAGS>
build_mode：编译类型，可填RELEASE或DEBUG
need_build_ut：是否编译uttest，可填ON或OFF
open_abi：编译时是否添加-D_GLIBCXX_USE_CXX11_ABI=1宏，可填ON或OFF
build_whl：是否编译python的whl包，可填ON或OFF
build_compiler：编译器选择，输入bisheng可手动指定编译器为bisheng
不填入参数情况下，默认执行build.sh RELEASE OFF ON ON gcc
```

3. ut运行

支持直接执行如下脚本编译并运行ut
```
bash test/run_dt.sh
```

### 安装使用

mindio_tft将所有特性集成到whl包中供用户使用，whl包格式为 ```mindio_ttp-{version}-py3-none-linux_{arch}.whl```

其中，versin表示mindio_tft版本；arch表示linux架构，如x86或aarch64

#### whl包编译

可以直接执行如下命令进行编译，在output目录下生成whl包
```
bash build/build.sh -t release
```

#### whl包依赖

whl包只能安装到npu环境上，且依赖于NPU固件驱动和CANN包，具体版本依赖详见下面的软件硬件配套说明

请在环境上提前安装NPU固件驱动和CANN包([环境安装参考链接](https://www.hiascend.com/document/detail/zh/CANNCommunityEdition/81RC1alpha002/softwareinst/instg/instg_0000.html))

安装完成后需要配置CANN环境变量([参考安装Toolkit开发套件包的第三步配置环境变量](https://www.hiascend.com/document/detail/zh/CANNCommunityEdition/81RC1alpha002/softwareinst/instg/instg_0008.html))

#### whl包安装

whl包的默认安装根路径为 /path/to/python3/site-packages/mindio_ttp

参考安装命令如下
```bash
pip3 install mindio_ttp-{version}-py3-none-linux_{arch}.whl --force-reinstall --no-index
```

安装完成后目录结构如下
```
${INSTALL_PATH}/
    |-- mindio_tft
        |-- controller_ttp
        |-- framework_ttp
        |-- mindspore_api
        |-- utils
        |-- __init__.py

default ${INSTALL_PATH} is /path/to/python3/site-packages/mindio_ttp
```

### 使用方法
[API接口参考](https://www.hiascend.com/document/detail/zh/mindx-dl/600/clusterscheduling/ref/mindiottp/mindiotft043.html)

### 软件硬件配套说明
- 硬件型号支持
  - Atlas 800T A2 系列产品
  - Atlas 900 A3 SuperPoD 超节点
- 平台：aarch64 / x86
- 配套软件：驱动固件 Ascend HDK 25.3.RC1、 CANN 8.3.RC1及之后版本
- python 3.7 ~ 3.11
- torch 2.7.1
- torch_npu 7.2.0

### 安全声明
[安全声明](./doc/SECURITYNOTE.md)

### 许可证
mindio_tft使用Apache License，详见[LICENSE](../../../LICENSE)文件。