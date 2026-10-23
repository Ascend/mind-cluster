# Usage on the CRI-O Client Using the crictl CLI<a name="ZH-CN_TOPIC_0000002511347204"></a>

## Usage Instructions<a name="section0966931165318"></a>

- Ascend Docker Runtime supports mounting physical chips and virtual chips. Before mounting virtual chips, refer to the [Creating vNPUs](../02_virtual_instance/00_virtual_instance_with_hdk/04_static_vnpu_scheduling/01_creating_vnpu.md) section to virtualize physical chips. Both static virtualization and dynamic virtualization of physical chips are supported.
- You can query the currently available physical chip IDs using the <b>ls /dev/davinci\*</b> command, and query the currently available virtual chip IDs using the <b>ls /dev/vdavinci\*</b> command.
- If you do not need to mount all content from the default Ascend Docker Runtime configuration file `/etc/ascend-docker-runtime.d/base.list`, you can create a custom configuration file (for example, `hostlog.list`) to reduce the mounted content. For details, refer to the [(Optional) Configuring Custom Mounted Content](./01_configuring_custom_mounted_content.md) section.

## Usage Examples<a name="section148905517123"></a>

crictl is the standard client of CRI-O. To start a container, you must first prepare the Pod Sandbox configuration file (for example, `pod-config.json`) and the container configuration file (for example, `container-config.json`), and then run the `crictl run` command to create and start the container based on the two configuration files. In the following examples, the Pod Sandbox configuration can be reused and its content is as follows:

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
> In non-K8s scenarios, setting `linux.security_context.namespace_options.pid` to `1` (CONTAINER) in the Pod Sandbox configuration means that the container uses an independent PID namespace, avoiding dependence on the pause container. `drop_infra_ctr` is enabled by default in CRI-O, which discards the pause container. If `linux.security_context.namespace_options.pid` is left at the default value `0` (POD, which shares the PID namespace of the pause container), container creation may fail. For details about CRI-O configuration, see the [official CRI-O documentation](https://github.com/cri-o/cri-o/blob/main/docs/crio.conf.5.md).

In the following examples, `{image-name:tag}` is the image name and tag, such as "ascend-pytorch:pytorch\_TAG".

- Example 1: Mount the physical chip with chip ID 0 when starting the container.

    Specify `ASCEND_VISIBLE_DEVICES` in `envs` of `container-config.json` to mount NPU 0:

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

    Pull the image and start the container:

    ```shell
    crictl pull {image-name:tag}
    crictl run container-config.json pod-config.json
    ```

- Example 2: Mount only the NPU device and management device when starting the container, without mounting driver-related directories.

    Append `ASCEND_RUNTIME_OPTIONS=NODRV` to `envs`:

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

    Pull the image and start the container:

    ```shell
    crictl pull {image-name:tag}
    crictl run container-config.json pod-config.json
    ```

- Example 3: Mount the physical chip with chip ID 0 when starting the container and read the mounted content from the custom configuration file `hostlog`.

    Append `ASCEND_RUNTIME_MOUNTS=hostlog` to `envs`:

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

    Pull the image and start the container:

    ```shell
    crictl pull {image-name:tag}
    crictl run container-config.json pod-config.json
    ```

- Example 4: Mount the virtual chip with chip ID 100 when starting the container.

    In static virtualization scenarios, the value of `ASCEND_VISIBLE_DEVICES` is the virtual chip ID, and the value of `ASCEND_RUNTIME_OPTIONS` must be `VIRTUAL`:

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

    Pull the image and start the container:

    ```shell
    crictl pull {image-name:tag}
    crictl run container-config.json pod-config.json
    ```

- Example 5: When starting the container, allocate 4 AICores from the physical chip with chip ID 0 as a virtual device and mount it into the container.

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

    Pull the image and start the container:

    ```shell
    crictl pull {image-name:tag}
    crictl run container-config.json pod-config.json
    ```

- Example 6: Mount the physical chip with chip ID 0 when starting the container and allow symbolic links in the mounted driver files (applicable only to Atlas 500 A2 intelligent station, Atlas 200I A2 accelerator module, and Atlas 200I DK A2 scenarios).

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

    Pull the image and start the container:

    ```shell
    crictl pull {image-name:tag}
    crictl run container-config.json pod-config.json
    ```

For details about the container configuration environment variables, see [Table 1](#table5134121862415).

After the container is started, run the following commands to check whether the corresponding devices and drivers are mounted successfully. For the specific mount directory of each product model, see [Content Mounted by Ascend Docker Runtime](../../07_references/05_appendix.md#content-mounted-by-ascend-docker-runtime). Example commands are as follows:

```shell
ls /dev | grep davinci* && ls /dev | grep devmm_svm && ls /dev | grep hisi_hdc && ls /usr/local/Ascend/driver && ls /usr/local/ |grep dcmi && ls /usr/local/bin
```

The possible output is as follows:

```text
davinci0
davinci_manager
devmm_svm
hisi_hdc
include lib64
dcmi
npu-smi
```

>[!NOTE]
>During use, do not redefine or fix environment variables such as `ASCEND_VISIBLE_DEVICES`, `ASCEND_RUNTIME_OPTIONS`, `ASCEND_RUNTIME_MOUNTS`, and `ASCEND_VNPU_SPECS` in the container image.
>Resumable training (checkpoint/restore) currently depends on the containerd snapshotter to export rootfs differences and is not supported in CRI-O scenarios. Support is planned for a later version.

**Table 1** Description of Ascend Docker Runtime container configuration environment variables

<a name="table5134121862415"></a>

|Parameter|Description|Example|
|--|--|--|
| `ASCEND_VISIBLE_DEVICES` | <ul><li>If the task does not require NPU devices, set `ASCEND_VISIBLE_DEVICES` to `void` or leave it empty.</li><li>If the task requires NPU devices, you must use `ASCEND_VISIBLE_DEVICES` to specify the NPU devices to be mounted into the container; otherwise, NPU device mounting will fail. When specifying devices by device ID, one or more devices can be specified (mixed specification is supported). When specifying devices by chip name, multiple chip names of the same type can be specified simultaneously.</li></ul> | <ul><li>`ASCEND_VISIBLE_DEVICES=void` means the Ascend Docker Runtime mounting feature is not used, and NPU devices, drivers, and file directories are not mounted. Related mount parameters will also become ineffective.</li>\<li>Mounting physical chips (NPUs)<ul><li>`ASCEND_VISIBLE_DEVICES=0` means NPU device 0 (`/dev/davinci0`) is mounted into the container.</li><li>`ASCEND_VISIBLE_DEVICES=1,3` means NPU devices 1 and 3 are mounted into the container.</li><li>`ASCEND_VISIBLE_DEVICES=0-2` means NPU devices 0 through 2 (inclusive) are mounted, equivalent to `-e ASCEND_VISIBLE_DEVICES=0,1,2`.</li><li>`ASCEND_VISIBLE_DEVICES=0-2,4` means NPU devices 0 through 2 and device 4 are mounted, equivalent to `-e ASCEND_VISIBLE_DEVICES=0,1,2,4`.</li><li>`ASCEND_VISIBLE_DEVICES=XXX-Y`, where `XXX` represents the NPU device type (supported values: `npu`, `Ascend910`, `Ascend310`, `Ascend310B`, `Ascend310P`) and `Y` represents the physical NPU device ID.<ul><li>`ASCEND_VISIBLE_DEVICES=npu-1` means NPU device 1 is mounted into the container.</li><li>`ASCEND_VISIBLE_DEVICES=npu-1,npu-3` means NPU devices 1 and 3 are mounted into the container.</li></ul></li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>When specifying devices by chip name, it is recommended to use `npu` uniformly.</li><li>Mixing device IDs and NPU names in the same parameter is not supported. For example, `ASCEND_VISIBLE_DEVICES=0,npu-1` is not supported.</li></ul></div></div><li>Mounting virtual chips (vNPU)<ul><li>**Static virtualization**: The usage is the same as for physical chips; simply replace the physical chip ID with the virtual chip ID (vNPU ID).</li><li>**Dynamic virtualization**: `ASCEND_VISIBLE_DEVICES=0` means allocating a certain number of AICores from NPU device 0.<div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>A single dynamic virtualization command can only specify one physical NPU ID for dynamic virtualization.</li><li>`ASCEND_VNPU_SPECS` must be used together with this parameter to specify the number of AICores to allocate from the specified NPU.</li><li>`ASCEND_RUNTIME_OPTIONS` can also be used with the value `NODRV`, meaning driver-related directories are not mounted.</li></ul></div></div></li></ul></li></ul> |
| `ASCEND_ALLOW_LINK` | Whether to allow symbolic links in mounted files or directories. This parameter must be specified for Atlas 500 A2 intelligent station, Atlas 200I A2 AI accelerator module, and Atlas 200I DK A2. <p>Other devices such as <term>Atlas training products</term>, <term>Atlas A2 training products</term>, and Atlas 200I SoC A1 core board can also use this parameter, but since their default mounted contents do not contain symbolic links, it is not necessary to specify this parameter.</p> | <ul><li>`ASCEND_ALLOW_LINK=True` allows mounting driver files with symbolic links on Atlas 500 A2 intelligent station, Atlas 200I A2 accelerator module, Atlas 200I DK A2, Atlas 850E, Atlas 850, Atlas 650E, Atlas 650, and Atlas 950 SuperPoD.</li><li>`ASCEND_ALLOW_LINK=False` or omitting this parameter means Atlas 500 A2 intelligent station, Atlas 200I A2 AI accelerator module, and Atlas 200I DK A2 cannot use Ascend Docker Runtime.</li></ul> |
|`ASCEND_UB_DRV_MOUNT`|Whether to mount UnifiedBus driver-related files. In scenarios that require support for the UnifiedBus communication protocol, such as Atlas 850E, Atlas 850, Atlas 650E, Atlas 650, and Atlas 950 SuperPoD, specifying this parameter determines whether to mount the host machine's UnifiedBus driver-related files into the container. <p>Other products, such as <term>Atlas training products</term>, <term>Atlas A2 training products</term>, and <term>Atlas A3 training products</term> can use this parameter, but because they do not depend on UnifiedBus driver-related files, there is no need to additionally specify this parameter. </p>|<ul><li>`ASCEND_UB_DRV_MOUNT=False` indicates that in Atlas 850E, Atlas 850, Atlas 650E, Atlas 650, and Atlas 950 SuperPoD scenarios, UnifiedBus driver-related files are not mounted. In this case, related functions can be supported by installing the UnifiedBus driver user-mode software package in the container.</li><li>`ASCEND_UB_DRV_MOUNT=True` or omitting this parameter means Atlas 850E, Atlas 850, Atlas 650E, Atlas 650, and Atlas 950 SuperPoD will mount UnifiedBus driver-related files. Because UnifiedBus driver-related files contain soft links, in order to mount successfully, soft link verification is automatically skipped, and there is no need to specify the `ASCEND_ALLOW_LINK=True` environment variable. In scenarios where `ub_driver.list` needs to be mounted, specifying `ASCEND_ALLOW_LINK=False` will be ineffective.</li></ul>|
| `ASCEND_RUNTIME_OPTIONS` | Restricts the chip ID specified in `ASCEND_VISIBLE_DEVICES`: <ul><li>`NODRV`: Driver-related directories are not mounted.</li><li>`VIRTUAL`: A virtual chip is mounted.</li><li>`NODRV,VIRTUAL`: A virtual chip is mounted, and driver-related directories are not mounted.</li></ul> | <ul><li>`ASCEND_RUNTIME_OPTIONS=NODRV`</li><li>`ASCEND_RUNTIME_OPTIONS=VIRTUAL`</li><li>`ASCEND_RUNTIME_OPTIONS=NODRV,VIRTUAL`</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>In static virtualization scenarios, `ASCEND_RUNTIME_OPTIONS` is mandatory and must include `VIRTUAL`.</li><li>In dynamic virtualization scenarios, if `ASCEND_RUNTIME_OPTIONS` is used, its value cannot include `VIRTUAL`.</li></ul></div></div> |
| `ASCEND_RUNTIME_MOUNTS` | The configuration file name(s) for content to be mounted, which lists files and directories to be mounted into the container. | <ul><li>`ASCEND_RUNTIME_MOUNTS=base`</li><li>`ASCEND_RUNTIME_MOUNTS=hostlog`</li><li>`ASCEND_RUNTIME_MOUNTS=hostlog,hostlog1,hostlog2`</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>By default, the configuration file `/etc/ascend-docker-runtime.d/base.list` is read.</li><li>Modify `hostlog.list` according to the actual custom configuration file name.</li><li>Multiple custom configuration files can be read.</li><li>File names must be lowercase and cannot contain uppercase letters.</li></ul></div></div> |
| `ASCEND_VNPU_SPECS` | Allocates a certain number of AICores from a physical NPU device and designates them as virtual devices. For supported values, see the "Virtualization Instance Template" column in Table 1 of [Virtualization Templates](../02_virtual_instance/00_virtual_instance_with_hdk/03_virtualization_templates.md). <ul><li>This parameter can only be used on product models that support dynamic virtualization.</li><li>Must be used together with `ASCEND_VISIBLE_DEVICES`, which specifies the physical NPU device for virtualization.</li></ul> | `ASCEND_VNPU_SPECS=vir04` means allocating 4 AICores as a virtual device and mounting it into the container. |
