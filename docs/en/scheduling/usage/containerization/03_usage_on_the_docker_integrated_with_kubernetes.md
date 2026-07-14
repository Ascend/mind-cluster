# Usage on the Docker integrated with Kubernetes

If Docker is integrated with Kubernetes, you need to install Ascend Docker Runtime.

- **NPU resources allocated**: When job YAML files are used to deliver training or inference jobs, Volcano and Ascend Device Plugin automatically allocate NPUs, and Ascend Docker Runtime automatically mounts NPUs and related files and directories. Example:

    ```Yaml
    apiVersion: mindxdl.gitee.com/v1
    kind: xxx
    ...
    spec:
    
    ...
            spec:
    ...
              containers:
              - name: ascend 
                image: pytorch-test:latest     # Modify the image name as required.
    ...
                resources:
                  limits:
                    huawei.com/Ascend910: 1     # Modify the resource name and quantity as required.
                  requests:
                    huawei.com/Ascend910: 1     # Modify the resource name and quantity as required.
    ...
    ```

- **NPU resources not allocated**: Add ASCEND_VISIBLE_DEVICES\=void. Example:

    ```Yaml
    apiVersion: mindxdl.gitee.com/v1
    kind: xxx
    ...
    spec:
    ...
            spec:
    ...
              containers:
              - name: ascend 
                image: pytorch-test:latest     # Modify the image name as required.
    ...
                env:
                - name: ASCEND_VISIBLE_DEVICES     # Add this configuration when NPU resources are not allocated by resources.
                   value: "void"
    ...
    ```
