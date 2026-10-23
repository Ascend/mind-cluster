# Examples and Verification

## Examples

### (Optional) Configuring Components<a name="ZH-CN_TOPIC_0000002511346449"></a>

If you have already configured resumable training-related functions when installing Ascend Device Plugin and NodeD, you can skip this chapter. If not, you need to configure [Ascend Device Plugin](#section14208511958) and [NodeD](#section162092113510) accordingly.

**Configuring Ascend Device Plugin<a name="section14208511958"></a>**

Only image-based startup of Ascend Device Plugin is supported.

1. Modify the startup YAML of Ascend Device Plugin based on the fault handling mode used. The bold parts below are modifications.
    1. Rescheduling mode

        >[!NOTE]
        >In rescheduling mode, an exception in Ascend Device Plugin also triggers fault rescheduling.

        <pre codetype="yaml">
        ...
              containers:
              - image: ascend-k8sdeviceplugin:v{version}
                name: device-plugin-01
                resources:
                  requests:
                    memory: 500Mi
                    cpu: 500m
                  limits:
                    memory: 500Mi
                    cpu: 500m
                command: [ "/bin/bash", "-c", "--"]
                args: [ "device-plugin
                         <strong>-volcanoType=true                    # Volcano must be used in rescheduling scenarios</strong>
                         <strong>-autoStowing=true                    # This field has been deprecated. Whether to enable automatic management switch; defaults to true. Setting to false disables automatic management, meaning that when a chip's health status changes from unhealthy to healthy, it will not be automatically added back to the schedulable resource pool. With automatic management disabled, when the chip's parameter-plane network fault is recovered, it will not be automatically added back to the schedulable resource pool. This feature only applies to <term>Atlas training products</term></strong>
                         <strong>-listWatchPeriod=5                   # Sets the health check interval; range [3,1800]; unit is seconds</strong>
                         -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log
                         -logLevel=0" ]
                securityContext:
                  privileged: true
                  readOnlyRootFilesystem: true
        ...</pre>

    2. (Optional) <span style="color:#D80000;">[DEPRECATED]</span> Graceful fault tolerance mode: Based on the rescheduling configuration, add the `-hotReset` field.

        >[!CAUTION]
        > **This feature has been deprecated. Do not use it!**
        >
        >- The graceful fault tolerance feature has been sunset. It will not be supported in PyTorch versions beyond 7.2.RC1 and MindSpore versions beyond 7.1.RC1.
        >- The function corresponding to `-hotReset = 1` has been sunset.
        >- The following configuration example is for historical reference only and is not recommended for use in production environments.

        <pre codetype="yaml">
        ...
              containers:
              - image: ascend-k8sdeviceplugin:v{version}
                name: device-plugin-01
                resources:
                  requests:
                    memory: 500Mi
                    cpu: 500m
                  limits:
                    memory: 500Mi
                    cpu: 500m
                command: [ "/bin/bash", "-c", "--"]
                args: [ "device-plugin
                         -volcanoType=true                    # Volcano must be used in rescheduling scenarios
                         -autoStowing=true                    # This field has been deprecated. Whether to enable automatic management switch; defaults to true. Setting to false disables automatic management, meaning that when a chip's health status changes from unhealthy to healthy, it will not be automatically added back to the schedulable resource pool. With automatic management disabled, when the chip's parameter-plane network fault is recovered, it will not be automatically added back to the schedulable resource pool. This feature only applies to <term>Atlas training products</term>
                         <span style="color:#D80000;"><strong>-hotReset=1 # [DEPRECATED] Enables graceful fault tolerance mode; the system will attempt to automatically reset the faulty chip (the online hot reset feature with a value of 1 has been deprecated)</strong></span>
                         -listWatchPeriod=5                   # Sets the health check interval; range [3,1800]; unit is seconds
                         -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log
                         -logLevel=0" ]
                securityContext:
                  privileged: true
                  readOnlyRootFilesystem: true
        ...</pre>

2. Run the following command on the K8s management node to start Ascend Device Plugin.

    ```shell
   # Replace {version} with the actual version number, for example, 26.2.0
    kubectl apply -f device-plugin-volcano-v{version}.yaml
    ```

**Configuring NodeD<a name="section162092113510"></a>**

You can configure the interval for reporting node status by manually modifying the NodeD startup YAML.

1. Go to the component decompression directory and run the following command to open the startup YAML file of NodeD.

    ```shell
   # Replace {version} with the actual version number, for example, 26.2.0
    vi noded-v{version}.yaml
    ```

2. Modify the `-reportInterval` parameter in the `args` line of the YAML file, as shown below:

    <pre codetype="yaml">
    ...
              env:
                - name: NODE_NAME
                  valueFrom:
                    fieldRef:
                      fieldPath: spec.nodeName
              imagePullPolicy: Never
              command: [ "/bin/bash", "-c", "--"]
              args: [ "/usr/local/bin/noded -logFile=/var/log/mindx-dl/noded/noded.log -logLevel=0 <strong>-reportInterval=5</strong>" ]
              securityContext:
                readOnlyRootFilesystem: true
                allowPrivilegeEscalation: true
              volumeMounts:
                - name: log-noded
    ...</pre>

## Scenario Examples

- [Examples and Verification in PyTorch Scenario](./01_pytorch_examples_and_verification.md)
- [Examples and Verification in MindSpore Scenario](./02_mindspore_examples_and_verification.md)
