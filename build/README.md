# 所有组件统一编译说明
-   **[编译](#编译)**
-   **[自动拉取源码失败](#自动拉取源码失败)**


# 编译

1.  拉取mindxdl整体源码放在/usr1目录下

2.  修改组件版本配置文件service_config.ini中mindxdlversion字段值为所需编译版本，默认值如下，

        mindxdlversion=6.0.RC3

3.  执行以下命令，进入/usr1/mindxdl/build目录，选择构建脚本执行

    **cd /usr1/mindxdl/build**

        dos2unix *.sh && chmod +x *.sh
        
        ./build_all.sh $GOPATH

4.  执行完成后进入$GOPATH/目录在各组件“output“目录下生成编译完成的文件,
    其中ascend-for-volcano组件编译完成文件在output目录中。


# 自动拉取源码失败

1.  参考以下命令，分别在/opt/buildtools/volcano_opensource/volcano_1.9/与
    /opt/buildtools/volcano_opensource/volcano_1.7/目录下手动拉取Volcano v1.9.0与v1.7.0版本官方开源代码。

    **cd** **/opt/buildtools/volcano_opensource/volcano_1.9/**
    **git clone -b release-1.9 https://github.com/volcano-sh/volcano.git**
2. 进入$GOPATH/mindxdl/ascend-docker-runtime目录，执行ascend-docker-runtime 组件readme
    中编译部分2,3命令手动拉取编译所需包，其中ascend-docker-runtime目录修改为当前目录


