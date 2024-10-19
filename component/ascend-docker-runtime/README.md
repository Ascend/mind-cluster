# Ascend Docker Runtime.zh
-  **[免责声明](#免责声明)**
-  **[组件介绍](#组件介绍)**
-  **[支持的产品形态](#支持的产品形态)**
-  **[编译Ascend-Docker-Runtime](#编译Ascend-Docker-Runtime)**
-  **[组件安装](#组件安装)**
-  **[更新日志](#更新日志)**
-  **[版本配套说明](#版本配套说明)**

# 免责声明
- 本代码仓库中包含多个开发分支，这些分支可能包含未完成、实验性或未测试的功能。在正式发布之前，这些分支不应被用于任何生产环境或依赖关键业务的项目中。请务必仅使用我们的正式发行版本，以确保代码的稳定性和安全性。
使用开发分支所导致的任何问题、损失或数据损坏，本项目及其贡献者概不负责。
- 正式版本请参考：[Ascend Docker Runtime正式release版本](https://gitee.com/ascend/ascend-docker-runtime/releases)
# 组件介绍
容器引擎插件（Ascend Docker，又叫昇腾容器）是CANN的基础组件，为所有的AI训练/推理作业提供Ascend NPU（昇腾处理器）容器化支持，使用户AI作业能够以Docker容器的方式平滑运行在昇腾设备之上，如图1-1所示。

图1-1 Ascend Docker

![image](assets/20210329102949456.png)

# 支持的产品形态
- Atlas 训练系列产品
- Atlas A2 训练系列产品 
- Atlas A3 训练系列产品
- 推理服务器（插Atlas 300I 推理卡）
- Atlas 200/300/500 推理产品
- Atlas 200I/500 A2 推理产品
- Atlas 推理系列产品（Ascend 310P AI处理器）
- Atlas 800I A2 推理服务器

## 设计简介

Ascend Docker Runtime本质上是基于OCI标准实现的Docker Runtime，不修改Docker引擎，对Docker以插件方式提供Ascend NPU适配功能。
如图1-2所示，Ascend Docker通过OCI接口与原生Docker对接。在原生Docker的runc启动容器过程中，会调用prestart-hook对容器进行配置管理。

图1-2 Docker适配原理

![image](assets/20230118566.png)

其中，prestart-hook是OCI定义的容器生存状态，即created状态到running状态的一个中间过渡所设置的钩子函数。在这个过渡状态，容器的namespace已经被创建，但容器的作业还没有启动，因此可以对容器进行设备挂载，cgroup配置等操作。这样随后启动的作业便可以使用到这些配置。
Ascend Docker在prestart-hook这个钩子函数中，对容器做了以下配置操作：
1.根据ASCEND_VISIBLE_DEVICES，将对应的NPU设备挂载到容器的namespace。
2.在Host上配置该容器的device cgroup，确保该容器只可以使用指定的NPU，保证设备的隔离。
3.将Host上的CANN Runtime Library挂载到容器的namespace。

# 编译Ascend-Docker-Runtime
执行以下步骤进行编译

 1、下载master分支下的源码包，获得ascend-docker-runtime
 
示例：源码放在/home/test/ascend-docker-runtime目录下
```shell
git clone https://gitee.com/ascend/ascend-docker-runtime.git
```

 2、下载tag为v1.1.10的安全函数库
````shell
cd /home/test/ascend-docker-runtime/platform
git clone -b v1.1.10 https://gitee.com/openeuler/libboundscheck.git
````

3、下载makeself
```shell
cd ../opensource
git clone -b openEuler-22.03-LTS https://gitee.com/src-openeuler/makeself.git
tar -zxvf makeself/makeself-2.4.2.tar.gz
```
 4、编译
```shell
cd ../build
bash build.sh
```
编译完成后，会在output文件夹看到相应的二进制run包
```shell
root@#:/home/test/ascend-docker-runtime/output# ll
...
-rwxr-xr-x  ... Ascend-docker-runtime_x.x.x_linux-x86_64.run*
```

# 组件安装
请参考[《MindX DL集群调度安装指南》--安装部署](https://www.hiascend.com/document/detail/zh/mindx-dl/60rc2/clusterscheduling/clusterschedulingig/clusterschedulingig/dlug_installation_012.html)中“Ascend Docker Runtime”章节进行。

# 更新日志

|      版本       |    发布日期    |       修改说明        |
|:-------------:|:----------:|:-----------------:|
|    v3.0.0     | 2023-01-18 |       第一次发布       |
|  v5.0.0-RC1   | 2023-04-18 | 配套MindX 5.0.RC1版本 |
|  v5.0.0-RC2   | 2023-07-18 | 配套MindX 5.0.RC2版本 |
|  v5.0.0-RC3   | 2023-10-27 | 配套MindX 5.0.RC3版本 |
|    v5.0.0     | 2023-12-29 |  配套MindX 5.0.0版本  |
|  v6.0.0-RC1   | 2024-04-22 | 配套MindX 6.0.RC1版本 |
|    v5.0.1     | 2024-05-18 | 配套MindX 5.0.1补丁版本 |
|  v6.0.0-RC2   | 2024-07-16 | 配套MindX 6.0.RC2版本 |

# 版本配套说明
- 版本配套详情请参考：[版本配套详情](https://www.hiascend.com/developer/download/commercial)