# 通过crictl命令行工具使用CRI-O<a name="ZH-CN_TOPIC_0000002511347204"></a>

## 使用说明<a name="section0966931165318"></a>

- Ascend Docker Runtime支持挂载物理芯片，同时支持挂载虚拟芯片。挂载虚拟芯片前需要参考[创建vNPU](../02_virtual_instance/00_virtual_instance_with_hdk/04_static_vnpu_scheduling/01_creating_vnpu.md)章节，对物理芯片进行虚拟化操作，支持对物理芯片进行静态虚拟化和动态虚拟化。
- 可通过<b>ls /dev/davinci\*</b>命令查询当前可用的物理芯片ID；通过<b>ls /dev/vdavinci\*</b>命令查询当前可用的虚拟芯片ID。
- 若用户不需要挂载Ascend Docker Runtime的默认配置文件“/etc/ascend-docker-runtime.d/base.list”中所有内容，可创建自定义配置文件（例如hostlog.list），减少挂载内容，具体操作请参考[（可选）配置自定义挂载内容](./01_configuring_custom_mounted_content.md)章节。
<!-- npu="950" id2 -->
- 在Atlas 950 SuperPoD超节点和Atlas 850超节点场景下，支持在Legacy模式下通过`ASCEND_RDMA_DEVICES`指定UB网卡在“/sys/class/infiniband/”下的完整HCA名称（可通过<b>ls /sys/class/infiniband</b>命令查询），将对应的RDMA字符设备与驱动文件挂载至容器；前提条件和默认驱动挂载说明请参见[使用前必读](./00_before_you_start.md)。
<!-- end id2 -->

## 使用示例<a name="section148905517123"></a>

crictl是CRI-O的标准客户端，启动容器需要先准备Pod Sandbox配置文件（如pod-config.json）和容器配置文件（如container-config.json），再通过`crictl run`命令基于这两个配置文件创建并启动容器。以下示例中Pod Sandbox配置可复用，内容如下：

```json
{
  "metadata": {
    "name": "npu-pod",
    "namespace": "default"
  },
  "log_directory": "/tmp",
  "linux": {
    "security_context": {
      "namespace_options": {
        "pid": 1
      }
    }
  }
}
```

> [!NOTE]
> 非K8s场景下，Pod Sandbox配置中`linux.security_context.namespace_options.pid`设置为`1`（CONTAINER）表示容器使用独立PID namespace，避免依赖pause容器；CRI-O默认开启`drop_infra_ctr`会丢弃pause容器，若`linux.security_context.namespace_options.pid`为缺省值`0`（POD，共享pause容器的PID namespace），可能会导致容器创建失败。关于CRI-O配置详情请参见[CRI-O官方文档](https://github.com/cri-o/cri-o/blob/main/docs/crio.conf.5.md)。

以下示例中的\{image-name:tag\}为镜像名称与标签，如“ascend-pytorch:pytorch\_TAG”。

- 示例1：启动容器时，挂载物理芯片ID为0的芯片。

    在container-config.json的`envs`中通过指定`ASCEND_VISIBLE_DEVICES`挂载0号NPU：

    ```json
    {
      "metadata": {
        "name": "npu-container"
      },
      "image": {
        "image": "{image-name:tag}"
      },
      "command": [
        "/bin/bash",
        "-c",
        "# Replace with your custom business program here"
      ],
      "envs": [
        {
          "key": "ASCEND_VISIBLE_DEVICES",
          "value": "0"
        }
      ],
      "linux": {
        "security_context": {
          "namespace_options": {
            "pid": 1
          }
        }
      }
    }
    ```

    拉取镜像并启动容器：

    ```shell
    crictl pull {image-name:tag}
    crictl run container-config.json pod-config.json
    ```

- 示例2：启动容器时，仅挂载NPU设备和管理设备，不挂载驱动相关目录。

    在`envs`中追加`ASCEND_RUNTIME_OPTIONS=NODRV`：

    ```json
    {
      "metadata": {
        "name": "npu-container"
      },
      "image": {
        "image": "{image-name:tag}"
      },
      "command": [
        "/bin/bash",
        "-c",
        "# Replace with your custom business program here"
      ],
      "envs": [
        {
          "key": "ASCEND_VISIBLE_DEVICES",
          "value": "0"
        },
        {
          "key": "ASCEND_RUNTIME_OPTIONS",
          "value": "NODRV"
        }
      ],
      "linux": {
        "security_context": {
          "namespace_options": {
            "pid": 1
          }
        }
      }
    }
    ```

    拉取镜像并启动容器：

    ```shell
    crictl pull {image-name:tag}
    crictl run container-config.json pod-config.json
    ```

- 示例3：启动容器时，挂载物理芯片ID为0的芯片，读取自定义配置文件hostlog中的挂载内容。

    在`envs`中追加`ASCEND_RUNTIME_MOUNTS=hostlog`：

    ```json
    {
      "metadata": {
        "name": "npu-container"
      },
      "image": {
        "image": "{image-name:tag}"
      },
      "command": [
        "/bin/bash",
        "-c",
        "# Replace with your custom business program here"
      ],
      "envs": [
        {
          "key": "ASCEND_VISIBLE_DEVICES",
          "value": "0"
        },
        {
          "key": "ASCEND_RUNTIME_MOUNTS",
          "value": "hostlog"
        }
      ],
      "linux": {
        "security_context": {
          "namespace_options": {
            "pid": 1
          }
        }
      }
    }
    ```

    拉取镜像并启动容器：

    ```shell
    crictl pull {image-name:tag}
    crictl run container-config.json pod-config.json
    ```

- 示例4：启动容器时，挂载虚拟芯片ID为100的芯片。

    静态虚拟化场景下，`ASCEND_VISIBLE_DEVICES`取值为虚拟芯片ID，且`ASCEND_RUNTIME_OPTIONS`取值必须为`VIRTUAL`：

    ```json
    {
      "metadata": {
        "name": "npu-container"
      },
      "image": {
        "image": "{image-name:tag}"
      },
      "command": [
        "/bin/bash",
        "-c",
        "# Replace with your custom business program here"
      ],
      "envs": [
        {
          "key": "ASCEND_VISIBLE_DEVICES",
          "value": "100"
        },
        {
          "key": "ASCEND_RUNTIME_OPTIONS",
          "value": "VIRTUAL"
        }
      ],
      "linux": {
        "security_context": {
          "namespace_options": {
            "pid": 1
          }
        }
      }
    }
    ```

    拉取镜像并启动容器：

    ```shell
    crictl pull {image-name:tag}
    crictl run container-config.json pod-config.json
    ```

- 示例5：启动容器时，从物理芯片ID为0的芯片上，切分出4个AICore作为虚拟设备并挂载至容器中。

    ```json
    {
      "metadata": {
        "name": "npu-container"
      },
      "image": {
        "image": "{image-name:tag}"
      },
      "command": [
        "/bin/bash",
        "-c",
        "# Replace with your custom business program here"
      ],
      "envs": [
        {
          "key": "ASCEND_VISIBLE_DEVICES",
          "value": "0"
        },
        {
          "key": "ASCEND_VNPU_SPECS",
          "value": "vir04"
        }
      ],
      "linux": {
        "security_context": {
          "namespace_options": {
            "pid": 1
          }
        }
      }
    }
    ```

    拉取镜像并启动容器：

    ```shell
    crictl pull {image-name:tag}
    crictl run container-config.json pod-config.json
    ```

<!-- npu="310b" id1 -->
- 示例6：启动容器时，挂载物理芯片ID为0的芯片，并且允许挂载的驱动文件中存在软链接（仅适用于Atlas 500 A2智能小站、Atlas 200I A2加速模块和Atlas 200I DK A2开发者套件场景）。

    ```json
    {
      "metadata": {
        "name": "npu-container"
      },
      "image": {
        "image": "{image-name:tag}"
      },
      "command": [
        "/bin/bash",
        "-c",
        "# Replace with your custom business program here"
      ],
      "envs": [
        {
          "key": "ASCEND_VISIBLE_DEVICES",
          "value": "0"
        },
        {
          "key": "ASCEND_ALLOW_LINK",
          "value": "True"
        }
      ],
      "linux": {
        "security_context": {
          "namespace_options": {
            "pid": 1
          }
        }
      }
    }
    ```

    拉取镜像并启动容器：

    ```shell
    crictl pull {image-name:tag}
    crictl run container-config.json pod-config.json
    ```

<!-- end id1 -->
容器配置环境变量说明如[表1](#table5134121862415)所示。

容器启动后，可执行以下命令检查相应设备和驱动是否挂载成功，每台机型具体的挂载目录请参见[Ascend Docker Runtime默认挂载内容](../../07_references/05_appendix.md#ascend-docker-runtime默认挂载内容)。命令示例如下：

```shell
ls /dev | grep davinci* && ls /dev | grep devmm_svm && ls /dev | grep hisi_hdc && ls /usr/local/Ascend/driver && ls /usr/local/ |grep dcmi && ls /usr/local/bin
```

可能的输出结果如下：

```text
davinci0
davinci_manager
devmm_svm
hisi_hdc
include lib64
dcmi
npu-smi
```

<!-- npu="950" id3 -->

### 挂载UB网卡设备及驱动<a name="section_rdma_crio_usage"></a>

在Atlas 950 SuperPoD超节点和Atlas 850超节点场景下，可通过`ASCEND_RDMA_DEVICES`挂载UB网卡的RDMA字符设备与驱动文件，HCA名称与NPU编号请根据实际环境修改：

- 示例1：启动容器时，同时挂载0号NPU设备与`hrn5_0`、`hrn5_1`对应的RDMA设备，并自动挂载NPU与RDMA驱动文件。

    在container-config.json的`envs`中同时指定`ASCEND_VISIBLE_DEVICES`与`ASCEND_RDMA_DEVICES`：

    ```json
    {
      "metadata": {
        "name": "npu-container"
      },
      "image": {
        "image": "{image-name:tag}"
      },
      "command": [
        "/bin/bash",
        "-c",
        "# Replace with your custom business program here"
      ],
      "envs": [
        {
          "key": "ASCEND_VISIBLE_DEVICES",
          "value": "0"
        },
        {
          "key": "ASCEND_RDMA_DEVICES",
          "value": "hrn5_0,hrn5_1"
        }
      ],
      "linux": {
        "security_context": {
          "namespace_options": {
            "pid": 1
          }
        }
      }
    }
    ```

    拉取镜像并启动容器：

    ```shell
    crictl pull {image-name:tag}
    crictl run container-config.json pod-config.json
    ```

- 示例2：启动容器时，仅挂载`hrn5_0`对应的RDMA设备及驱动文件，不挂载NPU设备及NPU驱动。

    在`envs`中配置`ASCEND_VISIBLE_DEVICES=void`与`ASCEND_RDMA_DEVICES=hrn5_0`：

    ```json
    {
      "metadata": {
        "name": "rdma-container"
      },
      "image": {
        "image": "{image-name:tag}"
      },
      "command": [
        "/bin/bash",
        "-c",
        "# Replace with your custom business program here"
      ],
      "envs": [
        {
          "key": "ASCEND_VISIBLE_DEVICES",
          "value": "void"
        },
        {
          "key": "ASCEND_RDMA_DEVICES",
          "value": "hrn5_0"
        }
      ],
      "linux": {
        "security_context": {
          "namespace_options": {
            "pid": 1
          }
        }
      }
    }
    ```

    拉取镜像并启动容器：

    ```shell
    crictl pull {image-name:tag}
    crictl run container-config.json pod-config.json
    ```

- 示例3：启动容器时，仅挂载NPU和RDMA字符设备，不自动挂载驱动文件（由容器镜像自带驱动，或通过挂载卷按需挂载所需驱动文件）。

    在`envs`中追加`ASCEND_RUNTIME_OPTIONS=NODRV`：

    ```json
    {
      "metadata": {
        "name": "npu-rdma-container"
      },
      "image": {
        "image": "{image-name:tag}"
      },
      "command": [
        "/bin/bash",
        "-c",
        "# Replace with your custom business program here"
      ],
      "envs": [
        {
          "key": "ASCEND_VISIBLE_DEVICES",
          "value": "0"
        },
        {
          "key": "ASCEND_RDMA_DEVICES",
          "value": "hrn5_0"
        },
        {
          "key": "ASCEND_RUNTIME_OPTIONS",
          "value": "NODRV"
        }
      ],
      "linux": {
        "security_context": {
          "namespace_options": {
            "pid": 1
          }
        }
      }
    }
    ```

    拉取镜像并启动容器：

    ```shell
    crictl pull {image-name:tag}
    crictl run container-config.json pod-config.json
    ```

    >[!NOTE]
    >
    >- 开启`NODRV`后，若需从宿主机按需挂载NPU驱动目录或`rdma_driver.list`中的RDMA驱动文件（不挂载整个“/usr/lib64”目录），可在container-config.json的`mounts`字段中配置`readonly: true`的挂载项。
    >- Kubernetes集成CRI-O场景下，可在Pod容器`env`中配置`ASCEND_RDMA_DEVICES`自动挂载RDMA字符设备与驱动文件；开启`NODRV`后可通过`volumeMounts`与`hostPath`按需只读挂载驱动目录或`rdma_driver.list`中的驱动文件。

容器启动后，可在容器内执行以下命令检查RDMA字符设备与UB网卡设备是否识别成功：

```shell
ls -l /dev/infiniband && ibv_devices
```

<!-- end id3 -->

>[!NOTE]
>用户在使用过程中，请勿重复定义或在容器镜像中固定ASCEND\_VISIBLE\_DEVICES、ASCEND\_RDMA\_DEVICES、ASCEND\_RUNTIME\_OPTIONS、ASCEND\_RUNTIME\_MOUNTS和ASCEND\_VNPU\_SPECS等环境变量。
>断点续训（checkpoint/restore）当前依赖containerd snapshotter导出rootfs差异，CRI-O场景暂不支持，规划在后续版本提供。

**表 1** Ascend Docker Runtime容器配置环境变量解释

<a name="table5134121862415"></a>

|参数|说明|举例|
|--|--|--|
|ASCEND_VISIBLE_DEVICES|<ul><li>如果任务不需要使用NPU设备，可以设置ASCEND_VISIBLE_DEVICES环境变量的取值为void或为空。</li><li>如果任务需要使用NPU设备，必须使用ASCEND_VISIBLE_DEVICES指定被挂载至容器中的NPU设备，否则挂载NPU设备失败。使用设备序号指定设备时，支持单个和范围指定且支持混用；使用芯片名称指定设备时，支持同时指定多个同类型的芯片名称。</li></ul>|<ul><li>ASCEND_VISIBLE_DEVICES=void表示不挂载NPU设备及NPU驱动。Legacy模式下，若未指定ASCEND_RDMA_DEVICES，不执行Hook挂载流程；若指定了ASCEND_RDMA_DEVICES，仅挂载所选RDMA设备及对应的RDMA驱动文件。若同时指定ASCEND_RUNTIME_OPTIONS=NODRV，跳过驱动文件的自动挂载，仅注入所选RDMA字符设备并校验宿主机RDMA依赖。</li><li>挂载物理芯片（NPU）<ul><li>ASCEND_VISIBLE_DEVICES=0时，表示将0号NPU设备（/dev/davinci0）挂载入容器中。</li><li>ASCEND_VISIBLE_DEVICES=1,3时，表示将1、3号NPU设备挂载入容器中。</li><li>ASCEND_VISIBLE_DEVICES=0-2时，表示将0号至2号NPU设备（包含0号和2号）挂载入容器中，效果同-e ASCEND_VISIBLE_DEVICES=0,1,2。</li><li>ASCEND_VISIBLE_DEVICES=0-2,4时，表示将0号至2号以及4号NPU设备挂载入容器，效果同-e ASCEND_VISIBLE_DEVICES=0,1,2,4。</li><li>ASCEND_VISIBLE_DEVICES=XXX-Y，其中XXX表示NPU设备，支持的取值为npu、Ascend910、Ascend310、Ascend310B和Ascend310P；Y表示物理NPU设备ID。<ul><li>ASCEND_VISIBLE_DEVICES=npu-1，表示把1号NPU设备挂载进容器。</li><li>ASCEND_VISIBLE_DEVICES=npu-1,npu-3，表示把1号NPU和3号NPU挂载进容器。</li></ul></li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>使用芯片名称指定设备时，建议统一取值npu。</li><li>不支持在一个参数里既指定设备序号又指定NPU名称，即不支持ASCEND_VISIBLE_DEVICES=0,npu-1。</li></ul></div></div></li><li>挂载虚拟芯片（vNPU）<ul><li>**静态虚拟化**：和物理芯片使用方式相同，只需要把物理芯片ID换成虚拟芯片ID（vNPU ID）即可。</li><li>**动态虚拟化**：ASCEND_VISIBLE_DEVICES=0表示从0号NPU设备中划分出一定数量的AICore。<div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>一条动态虚拟化的命令只能指定一个物理NPU的ID进行动态虚拟化。</li><li>必须搭配ASCEND_VNPU_SPECS，表示在指定的NPU上划分出的AICore数量。</li><li>可以搭配ASCEND_RUNTIME_OPTIONS，但是只能取值为NODRV，表示不挂载驱动相关目录。</li></ul></div></div></li></ul></li></ul>|
|ASCEND_ALLOW_LINK|是否允许挂载的文件或目录中存在软链接，在Atlas 500 A2智能小站、Atlas 200I A2 AI加速模块和Atlas 200I DK A2开发者套件场景下需要指定该参数。<p>其他设备如<term>Atlas训练系列产品</term>、<term>Atlas A2训练系列产品</term>和Atlas 200I SoC A1核心板等产品可以使用该参数，但因其默认挂载内容中不存在软链接，所以无需额外指定该参数。</p>|<ul><li>ASCEND_ALLOW_LINK=True，表示在Atlas 500 A2智能小站、Atlas 200I A2 AI加速模块、Atlas 200I DK A2开发者套件、Atlas 850E超节点、Atlas 850超节点、Atlas 650E服务器、Atlas 650服务器和Atlas 950 SuperPoD超节点场景下允许挂载带有软链接的驱动文件。</li><li>ASCEND_ALLOW_LINK=False或者不指定该参数，Atlas 500 A2智能小站、Atlas 200I A2 AI加速模块和Atlas 200I DK A2开发者套件将无法使用Ascend Docker Runtime。</li><li>挂载所选RDMA设备的驱动文件时，自动允许通过属主、权限和路径校验的软链接，无需指定ASCEND_ALLOW_LINK=True；设置ASCEND_ALLOW_LINK=False时，该流程仍使用True。</li></ul>|
|ASCEND_UB_DRV_MOUNT|是否挂载UnifiedBus驱动相关文件，在Atlas 850E超节点、Atlas 850超节点、Atlas 650E服务器、Atlas 650服务器和Atlas 950 SuperPoD超节点等需要支持UnifiedBus通信协议的场景下，通过指定该参数决定是否挂载宿主机的UnifiedBus驱动相关文件到容器中。<p>其他设备如<term>Atlas训练系列产品</term>、<term>Atlas A2训练系列产品</term>和<term>Atlas A3训练系列产品</term>等可以使用该参数，但因其不依赖UnifiedBus驱动相关文件，所以无需额外指定该参数。</p>|<ul><li>ASCEND_UB_DRV_MOUNT=False，表示在Atlas 850E超节点、Atlas 850超节点、Atlas 650E服务器、Atlas 650服务器和Atlas 950 SuperPoD超节点场景下不挂载UnifiedBus驱动相关文件，此时可通过在容器内安装UnifiedBus驱动用户态软件包的方式支持相关功能。</li><li>ASCEND_UB_DRV_MOUNT=True或者不指定该参数，Atlas 850E超节点、Atlas 850超节点、Atlas 650E服务器、Atlas 650服务器和Atlas 950 SuperPoD超节点将挂载UnifiedBus驱动相关文件。由于UnifiedBus驱动相关文件存在软链接，为了成功挂载，会自动跳过软链接校验，无需指定ASCEND_ALLOW_LINK=True环境变量。需要挂载ub_driver.list的场景，指定ASCEND_ALLOW_LINK=False会无效。</li></ul>|
|ASCEND_RUNTIME_OPTIONS|配置容器运行时的挂载与虚拟化选项：<ul><li>NODRV：不自动挂载NPU与RDMA驱动相关文件及目录（若指定了ASCEND_RDMA_DEVICES，仍注入RDMA字符设备并校验宿主机RDMA设备、设备表、清单及必需文件）。</li><li>VIRTUAL：表示挂载的是虚拟芯片。</li><li>NODRV,VIRTUAL：表示挂载的是虚拟芯片，并且不自动挂载NPU与RDMA驱动相关文件及目录。</li></ul>|<ul><li>ASCEND_RUNTIME_OPTIONS=NODRV</li><li>ASCEND_RUNTIME_OPTIONS=VIRTUAL</li><li>ASCEND_RUNTIME_OPTIONS=NODRV,VIRTUAL</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>静态虚拟化场景下，ASCEND_RUNTIME_OPTIONS为必选参数，且取值必须包含VIRTUAL。</li><li>动态虚拟化场景下，若使用ASCEND_RUNTIME_OPTIONS参数，则其取值不能包含VIRTUAL。</li></ul></div></div>|
|ASCEND_RUNTIME_MOUNTS|待挂载内容的配置文件名，该文件中可配置需要挂载到容器内的文件及目录。|<ul><li>ASCEND_RUNTIME_MOUNTS=base</li><li>ASCEND_RUNTIME_MOUNTS=hostlog</li><li>ASCEND_RUNTIME_MOUNTS=hostlog,hostlog1,hostlog2</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>默认读取/etc/ascend-docker-runtime.d/base.list配置文件。</li><li>hostlog.list请根据实际自定义配置文件名修改。</li><li>支持读取多个自定义配置文件。</li><li>文件名必须小写，不能包含大写字母。</li></ul></div></div>|
|ASCEND_VNPU_SPECS|从物理NPU设备中切分出一定数量的AICore，指定为虚拟设备。支持的取值请参见[虚拟化模板](../02_virtual_instance/00_virtual_instance_with_hdk/03_virtualization_templates.md)中表1的“虚拟化实例模板”列。<ul><li>只有支持动态虚拟化的产品形态，才能使用该参数。</li><li>需配合参数“ASCEND_VISIBLE_DEVICES”一起使用，参数“ASCEND_VISIBLE_DEVICES”指定用于虚拟化的物理NPU设备。</li></ul>|ASCEND_VNPU_SPECS=vir04表示切分4个AICore作为虚拟设备，挂载至容器中。|
|ASCEND_RDMA_DEVICES|指定待挂载的RDMA设备，参数取值为UB网卡在“/sys/class/infiniband/”下的完整HCA名称，在Atlas 950 SuperPoD超节点和Atlas 850超节点场景下使用，仅在Legacy注入模式下支持；CDI模式会忽略ASCEND_RDMA_DEVICES，不会根据该参数自动注入RDMA设备及挂载对应驱动文件。<ul><li>ASCEND_RDMA_DEVICES未设置、取值为空或void时，不读取RDMA设备表，不自动发现和注入RDMA设备及驱动。</li><li>仅支持使用英文逗号分隔的HCA设备全称，去重后最多支持512个。名称仅支持英文字母、数字和下划线，不支持纯数字或“0-7”等范围格式，且必须以英文字母开头。</li></ul><p>该参数与ASCEND_VISIBLE_DEVICES相互独立，前提条件和默认驱动挂载说明请参见[使用前必读](./00_before_you_start.md)。</p>|<ul><li>ASCEND_RDMA_DEVICES=void，表示不自动发现和注入RDMA设备及驱动；若仍申请NPU，ASCEND_RUNTIME_MOUNTS手动指定的集合仍按普通规则挂载。NPU和RDMA都未申请时，不执行Hook挂载流程。</li><li>ASCEND_RDMA_DEVICES=hrn5_0，表示挂载hrn5_0对应的“/dev/infiniband/uverbsX”字符设备（宿主机存在“/dev/infiniband/rdma_cm”时一并挂载）及匹配的RDMA驱动文件。</li><li>ASCEND_RDMA_DEVICES=hrn5_0,hrn5_1，表示同时挂载hrn5_0和hrn5_1对应的RDMA字符设备及驱动文件。</li></ul>|
