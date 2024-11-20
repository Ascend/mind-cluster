# AscendCommon

# 组件介绍
提供公共代码给其他组件使用，组件包括NPU-Exporter等。

# 集成AscendCommon编译其他组件
编译流程以编译NPU-Exporter为例

1. 通过git拉取源码，获得ascend-common和ascend-npu-exporter。

   ascend-common与ascend-npu-exporter放在同一目录下。

   示例：

   Npu-Exporter源码放在 /home/test/ascend-npu-exporter目录下。

   AscendCommon源码放在/home/test/ascend-common目录下。
2. 执行以下命令，进入NPU-Exporter构建目录，执行构建脚本，在“output“目录下生成二进制npu-exporter、yaml文件和Dockerfile等文件。

   **cd** _/home/test/_**ascend-npu-exporter/build/**

   **chmod +x build.sh**

   **./build.sh**
3. 执行以下命令，查看**output**生成的软件列表。

   **ll** _/home/test/_**ascend-npu-exporter/output**

   ```
   drwxr-xr-x  2 root root     4096 Feb 23 07:10 .
   drwxr-xr-x 10 root root     4096 Feb 23 07:10 ..
   -r--------  1 root root      623 Feb 23 07:10 Dockerfile
   -r--------  1 root root      964 Feb 23 07:10 Dockerfile-310P-1usoc
   -r-x------  1 root root 15861352 Feb 23 07:10 npu-exporter
   -r--------  1 root root     4089 Feb 23 07:10 npu-exporter-310P-1usoc-v5.0.RC3.yaml
   -r--------  1 root root     3438 Feb 23 07:10 npu-exporter-v5.0.RC3.yaml
   -r-x------  1 root root     2554 Feb 23 07:10 run_for_310P_1usoc.sh
   ```

# 说明

1. 编译NPU-Exporter等组件时，ascend-common要放在同一目录下