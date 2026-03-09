# hccl.json文件说明<a name="ZH-CN_TOPIC_0000002511346379"></a>

Ascend Operator将在训练启动时，为训练任务生成集合通信所需的RankTable文件。集合通信根据RankTable文件中的设备ID以及IP构建集合通信域，完成集合通信的信息交换。

- 使用Ascend Operator ConfigMap挂载RankTable时，需要在创建任务时，同时在训练YAML中创建名称为rings-config-<任务名\>的ConfigMap，并将该ConfigMap挂载进训练容器的“/user/serverid/devindex/config”路径下。Ascend Operator将根据Ascend Device Plugin在任务Pod中写的Annotation信息，构建出任务的集合通信文件RankTable File，并将其内容写入ConfigMap中，在训练容器中映射为“/user/serverid/devindex/config/hccl.json”文件。
- 使用共享存储的方式挂载RankTable时，需要在创建任务时，同时在训练YAML中挂载共享存储或者本地存储的目录，并将该目录挂载进训练容器的“/user/serverid/devindex/config”路径下。Ascend Operator将根据Ascend Device Plugin或volcano-scheduler在任务Pod中写的Annotation信息，构建出任务的集合通信文件RankTable File，并将其内容写入“/共享存储或者本地存储目录/hccl.json”文件中，在训练容器中映射为“/user/serverid/devindex/config/hccl.json”文件。
- 不同产品型号的hccl.json有不同的文件内容，详细说明如下所示。

**Atlas 训练系列产品、Atlas A2 训练系列产品、Atlas 800I A2 推理服务器、A200I A2 Box 异构组件<a name="section19616113871318"></a>**

hccl.json文件示例如下：

```
hccl.json:
----
{
    "status": "completed",  // Ascend Operator是否写入完成
    "server_list": [{    // 节点列表
        "device": [{   // NPU列表
            "device_id": "0",  // NPU的设备ID
            "device_ip": "192.168.101.xx",   // NPU的设备IP
            "rank_id": "0" // NPU对应的训练rank ID
        }, {
            "device_id": "1",
            "device_ip": "192.168.102.xx",
            "rank_id": "1"
        }, {
            "device_id": "2",
            "device_ip": "192.168.103.xx",
            "rank_id": "2"
        }, {
...
        }],
        "server_id": "xx-xx-xx-xx",   // AI Server标识，全局唯一
        "host_ip": "xx.xx.xx.xx",      // AI Server的Host IP地址
        "container_ip": "192.168.149.xx",   // Pod IP
    "hardware_type":"800I-A2-32G"       // 产品型号
    }]
    "server_count": "1",   // 任务总服务器数量
    "version": "1.0"
}
```

**Atlas A3 训练系列产品<a name="section285395510348"></a>**

hccl.json文件示例如下：

```
hccl.json:
----
{
    "status": "completed",  // Ascend Operator是否写入完成
    "server_list": [    // 节点列表
        {
            "device": [
                {
                    "device_id": "0",     // NPU的设备ID
                    "device_ip": "xx.xx.xx.xx",  // NPU的设备IP
                    "super_device_id": "37748736",   //NPU的设备ID
                    "rank_id": "0"             // NPU对应的训练rank ID
                },
...
                {
                    "device_id": "7",
                    "device_ip": "xx.xx.xx.xx",
                    "super_device_id": "38600711",
                    "rank_id": "7"
                }
            ],
            "server_id": "xx-xx-xx-xx",  //AI Server标识，全局唯一
            "host_ip": "xx.xx.xx.xx",      // AI Server的Host IP地址
            "container_ip": "192.168.149.xx",   // Pod IP
     "hardware_type":"800I-A3-64G"       // 产品型号
        }
    ],
    "server_count": "1",
    "version": "1.2",
    "super_pod_list": [   //超节点列表
        {
            "super_pod_id": "0",  //逻辑超节点ID
            "server_list": [
                {
                    "server_id": "xx-xx-xx-xx"   //AI Server标识，全局唯一
                }
            ]
        }
    ]
}
```

**Atlas A5 系列产品<a name="section285395510348"></a>**
产品包括：Atlas 950 训练系列产品、Atlas 850 服务器系列产品、Atlas 350 标卡系列产品

hccl.json文件示例如下：

```
hccl.json:
----
{
  "status": "completed", // Ascend Operator是否写入完成
  "version": "2.0",
  "rank_count": 1,     // 参与训练的rank个数
  "rank_list": [       // rank信息列表
    {
      "rank_id": 0,    // 训练rank ID
      "local_id": 0,   // 与拓扑文件中的ID关联
      "device_id": 0,  // 物理ID
      "level_list": [
        {
          "net_layer": 0,   // 通信层级
          "net_instance_id": "xx",
          "net_type": "TOPO_FILE_DESC",     // 网络类型，值为TOPO_FILE_DESC和CLOS，TOPO_FILE_DESC代表从文件中查询网络类型，CLOS代表clos网络
          "net_attr": "",                   // 组网层级
          "rank_addr_list": [
            {
              "addr_type": "EID",           // 地址类型
              "addr": "....",               // 地址值
              "ports": ["x/x"],             // NPU端口列表
              "plane_id": "1"               // 网络平面
            },
            ...
            {
              "addr_type": "EID",
              "addr": "....",
              "ports": ["x/x"],
              "plane_id": "1"
            },

          ]
        }
      ]
    }
  ]
}
```
