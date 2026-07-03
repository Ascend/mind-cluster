# hccl.json File Description<a name="ZH-CN_TOPIC_0000002511346379"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T01:42:00.716Z pushedAt=2026-06-09T02:05:50.625Z -->

When training starts, Ascend Operator generates the RankTable file required for collective communication for training jobs. The collective communication domains are built based on the device IDs and IPs in the RankTable file to complete information exchange for collective communication.

- When using Ascend Operator ConfigMap to mount the RankTable, you need to create a ConfigMap named `rings-config-<job name>` in the YAML when creating a training job, and mount the ConfigMap to the `/user/serverid/devindex/config` path in the training container. Ascend Operator constructs the collective communication RankTable file for the job based on the `Annotation` information written by Ascend Device Plugin in the job Pod, writes its content into the ConfigMap, and maps it as the `"/user/serverid/devindex/config/hccl.json"` file in the training container.
- When using shared storage to mount the RankTable, you need to mount a shared storage or local storage directory in the YAML when creating a training job, and mount the directory to the `"/user/serverid/devindex/config"` path in the training container. Ascend Operator constructs the collective communication RankTable file for the job based on the `Annotation` information written by Ascend Device Plugin or volcano-scheduler in the job Pod, writes its content into the `"/shared storage or local storage directory/hccl.json"` file, and maps it as the `"/user/serverid/devindex/config/hccl.json"` file in the training container.
- Different products have different `hccl.json` file contents, as detailed below.

## Atlas Training Series Products, Atlas A2 Training Series Products, Atlas 800I A2 Inference Server, A200I A2 Box Heterogeneous Subrack<a name="section19616113871318"></a>

The `hccl.json` file example is as follows:

```json
hccl.json:
----
{
    "status": "completed",  // Whether Ascend Operator has finished writing
    "server_list": [{    // Node List
        "device": [{   // NPU List
            "device_id": "0",  // NPU Device ID
            "device_ip": "192.168.101.xx",   // NPU Device IP
            "rank_id": "0" // Training Rank ID Corresponding to the NPU
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
        "server_id": "xx-xx-xx-xx",   // AI Server Identifier, Globally Unique
        "host_ip": "xx.xx.xx.xx",      // Host IP Address of the AI Server
        "container_ip": "192.168.149.xx",   // Pod IP
    "hardware_type":"800I-A2-32G"       // Product Model
    }]
    "server_count": "1",   // Total Number of Servers for the Task
    "version": "1.0"
}
```

## Atlas A3 Training Series Products<a name="section285395510348"></a>

The following is an example of the `hccl.json` file:

```json
hccl.json:
----
{
    "status": "completed",  // Whether Ascend Operator Has Finished Writing
    "server_list": [    // Node List
        {
            "device": [
                {
                    "device_id": "0",     // NPU Device ID
                    "device_ip": "xx.xx.xx.xx",  // NPU Device IP
                    "super_device_id": "37748736",   //NPU Device ID
                    "rank_id": "0"             // Training Rank ID Corresponding to the NPU
                },
...
                {
                    "device_id": "7",
                    "device_ip": "xx.xx.xx.xx",
                    "super_device_id": "38600711",
                    "rank_id": "7"
                }
            ],
            "server_id": "xx-xx-xx-xx",  //AI Server identifier, globally unique
            "host_ip": "xx.xx.xx.xx",      // Host IP address of the AI Server
            "container_ip": "192.168.149.xx",   // Pod IP
     "hardware_type":"800I-A3-64G"       // Product model
        }
    ],
    "server_count": "1",
    "version": "1.2",
    "super_pod_list": [   //Super node list
        {
            "super_pod_id": "0",  //Logical super node ID
            "server_list": [
                {
                    "server_id": "xx-xx-xx-xx"   //AI Server identifier, globally unique
                }
            ]
        }
    ]
}
```

## Atlas 350 PCIe card, Atlas 850 Series Hardware Products, Atlas 950 SuperPoD<a name="section285395510348"></a>

The `hccl.json` file example is as follows:

```json
hccl.json:
----
{
  "status": "completed", // Whether the Ascend Operator has finished writing
  "version": "2.0",
  "rank_count": 1,     // Number of ranks participating in training
  "rank_list": [       // Rank Information List
    {
      "rank_id": 0,    // Training Rank ID
      "local_id": 0,   // Associated with the ID in the topology file
      "device_id": 0,  // Physical ID
      "level_list": [
        {
          "net_layer": 0,   // Communication Level
          "net_instance_id": "xx",          // Network ID
          "net_type": "TOPO_FILE_DESC",     // Network type, with values TOPO_FILE_DESC and CLOS. TOPO_FILE_DESC indicates querying the network type from a file, and CLOS indicates a CLOS network.
          "net_attr": "",                   // Network Hierarchy
          "rank_addr_list": [
            {
              "addr_type": "EID",           // Address Type
              "addr": "....",               // Address Value
              "ports": ["x/x"],             // NPU Port List
              "plane_id": "1"               // Network Plane
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
