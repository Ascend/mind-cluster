# Resource Allocation and Release Under the DRA Mechanism<a name="ZH-CN_TOPIC_000000_dra_deploy_workflow"></a>

## Working Principle<a name="section_dra_principle"></a>

Ascend DRA manages NPU devices based on the Kubernetes Dynamic Resource Allocation (DRA) mechanism. For the complete lifecycle, see [Before You Start/Lifecycle](00_before_you_start.md#section_dra_life_cycle).

There are two schemes for requesting full-NPU resources. They are mutually exclusive, and only one of them can be used:

- **Scheme 1: Request resources through a ResourceClaimTemplate.** Explicitly create a `ResourceClaimTemplate`, and the Pod references it through `spec.resourceClaims`. This scheme applies to scenarios where resource templates need to be explicitly managed or where consistency with the existing `ResourceClaim` process is required. It can be used to request an unsliced full NPU, and it can also request static vNPUs or dynamic vNPUs by replacing `deviceClassName`.
- **Scheme 2: Request resources by using an extended resource name.** No `ResourceClaimTemplate` is created. The Pod only configures the extended resource name in `resources.requests` and `resources.limits`. This scheme applies to scenarios where only a full NPU needs to be requested and a simpler manifest is preferred, and it requires K8s 1.36 or later. It supports only unsliced physical NPUs. To request vNPUs, use Scheme 1.

## Prerequisites<a name="section_dra_prereq"></a>

- Volcano and Ascend DRA have been installed and deployed. For details, see [Installation and Deployment](../../03_installation_guide/02_installation/00_helm_installation.md). The Ascend DRA component injects NPU devices into service containers through the CDI mechanism, and it neither installs nor depends on Ascend Docker Runtime.
- The NPU driver and firmware have been installed on the NPU nodes, and the devices are running properly.
- When [Scheme 2](#section_scheme_extended_resource) is used, the K8s version must be 1.36 or later, the `DynamicResourceAllocation` and `DRAExtendedResource` feature gates must be enabled (they are enabled by default in K8s 1.36 and later), and kube-scheduler must be used as the scheduler.

## Scheme 1: Requesting Resources Through a ResourceClaimTemplate<a name="section_scheme_claim_template"></a>

Users define NPU resource requests through a `ResourceClaimTemplate`. By referencing this template, the Pod automatically creates an independent `ResourceClaim` for each replica.

### Configuring the YAML File<a name="section_claim_template_config"></a>

1. Create the `ascend-npu-claim-template.yaml` file.

    <pre codetype="yaml">
    apiVersion: resource.k8s.io/v1
    kind: ResourceClaimTemplate
    metadata:
      name: ascend-npu-template
      namespace: default
    spec:
      metadata: {}
      spec:
        devices:
          requests:
          - <strong>name: npu              # Request name, referenced in the Pod by this name</strong>
            <strong>exactly:</strong>
              <strong>deviceClassName: npu.huawei.com # Reference the DeviceClass created when the DRA component is deployed</strong>
              <strong>count: 1                        # Number of NPUs requested. Modify as needed</strong>
    </pre>

    > [!NOTE]
    > The value of `deviceClassName` is `npu.huawei.com`. This `DeviceClass` is automatically created when `ascend-dra-driver.yaml` is deployed, and its selector matches the unsliced physical NPU whose driver is `npu.huawei.com`. To request vNPUs, change the value to `static-vnpu.huawei.com` or `dynamic-vnpu.huawei.com`. If you only need to request a full NPU and do not want to create a template, see [Scheme 2](#section_scheme_extended_resource).

2. Create the template.

    ```bash
    kubectl apply -f ascend-npu-claim-template.yaml
    ```

    Command output:

    ```text
    resourceclaimtemplate.resource.k8s.io/ascend-npu-template created
    ```

### Delivering a Job<a name="section_claim_template_apply"></a>

Deliver a Deployment that references the `ResourceClaimTemplate`. Each Pod replica automatically creates an independent ResourceClaim. After the Pod is scheduled to the NPU node, kubelet automatically calls the DRA plugin to complete `Prepare`.

1. Create the `dra-test-deployment.yaml` file.

    <pre codetype="yaml">
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: dra-test
      namespace: default
      labels:
        app: dra-test
    spec:
      <strong>replicas: 1                               # Number of replicas. Modify as needed</strong>
      selector:
        matchLabels:
          app: dra-test
      template:
        metadata:
          labels:
            app: dra-test
        spec:
          <strong>resourceClaims:</strong>
          <strong>- name: npu                           # Consistent with the name referenced in the container resources.claims</strong>
            <strong>resourceClaimTemplateName: ascend-npu-template</strong>
          containers:
          - name: test
            image: ubuntu:22.04
            imagePullPolicy: IfNotPresent
            command:
            - /bin/bash
            - -c
            - sleep inf
            resources:
              claims:
              <strong>- name: npu                      # Reference the name in resourceClaims above</strong>
    </pre>

2. Deliver the job.

   ```bash
   kubectl apply -f dra-test-deployment.yaml
   ```

   Command output:

   ```text
   deployment.apps/dra-test created
   ```

3. Check the `Deployment` status and confirm that the Pod enters the `Running` state.

   ```bash
   kubectl get pod -l app=dra-test -o wide
   ```

   Command output:

   ```text
   NAME                       READY   STATUS    RESTARTS   AGE   IP            NODE          NOMINATED NODE   READINESS GATES
   dra-test-<random-suffix>   1/1     Running   0          30s   10.244.1.20   node-npu-01   <none>           <none>
   ```

4. Verify that `Prepare` is complete.

   Check the DRA component logs to confirm that `PrepareResourceClaims` is called and devices are allocated (the following logs are consistent with the `Infof` output in the code `plugin.go` and `state.go`):

   ```bash
   kubectl logs -n kube-system <dra-pod-name> | grep "PrepareResourceClaims\|devices prepared\|Returning newly prepared"
   ```

   Command output:

   ```text
   PrepareResourceClaims is called, number of claims: 1
   devices prepared for claim <claim-uid>, count=1, cdiIDs=[npu.huawei.com/ascend910=0]
   Returning newly prepared devices for claim '<claim-uid>', count=1
   ```

   Where `<claim-uid>` is the UID of `ResourceClaim`, `cdiIDs` is the CDI device ID list in the format `<driverName>/<class>=<index>`, and the value of `driverName` is `npu.huawei.com`.

5. View the generated Container Device Interface (CDI) spec file.

   During the `Prepare` phase, the DRA plugin generates a CDI spec file for each Claim, located in the `/var/run/cdi` directory on the node.

   1. On the node where the Pod resides, view the CDI spec file.

      ```bash
      ls /var/run/cdi/
      ```

      The following is the output example, where a CDI spec file named after the Claim UID exists:

      ```text
      <claim-uid>.json
      ```

   2. View the content of the CDI spec file.

      ```bash
      cat /var/run/cdi/<claim-uid>.json
      ```

      The following is the output example (taking the allocation of two NPUs as an example; the actual content varies with the node device model and driver version):

      ```yaml
      cdiVersion: 0.8.0
      kind: ascend.com/npu
      devices:
      - name: "0"
        containerEdits:
          deviceNodes:
          - path: /dev/davinci0
            hostPath: /dev/davinci0
            type: c
            major: 503
            fileMode: 8624
      - name: "1"
        containerEdits:
          deviceNodes:
          - path: /dev/davinci1
            hostPath: /dev/davinci1
            type: c
            major: 503
            fileMode: 8624
      containerEdits:
        env:
        - LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/common:/usr/local/Ascend/driver/lib64/driver:/usr/lib64
        deviceNodes:
        - path: /dev/davinci_manager
          hostPath: /dev/davinci_manager
          type: c
          major: 504
          fileMode: 8624
        - path: /dev/hisi_hdc
          hostPath: /dev/hisi_hdc
          ...
        mounts:
        - hostPath: /usr/local/dcmi
          containerPath: /usr/local/dcmi
          options:
          - rbind
          - rprivate
          - ro
          type: bind
        - hostPath: /usr/local/bin/npu-smi
          containerPath: /usr/local/bin/npu-smi
          ...
      ```

      Field description:
      - `cdiVersion`: CDI spec version, fixed to `0.8.0`.
      - `kind`: CDI driver identifier, fixed to `ascend.com/npu`.
      - `devices`: Edit items divided by device granularity. Each `name` corresponds to the `/dev/davinci<id>` device node of one NPU.
      - `containerEdits` (spec level): Edit items shared by all devices, including shared device nodes (`davinci_manager`, `hisi_hdc`, and so on), the `LD_LIBRARY_PATH` environment variable, and driver library/tool mounts.

6. Access the service Pod container to verify that the NPU device has been injected.

   ```bash
   kubectl exec -it <pod-name> -- ls /dev/davinci*
   ```

   The following is the output example, where NPU device nodes are visible in the container:

   ```text
   /dev/davinci0
   ```

   > [!NOTE]
   > The NPU device node path is `/dev/davinci<id>`, where `<id>` is the physical device ID, consistent with the `physicId` attribute of the device in the ResourceSlice. The specific device injection is completed by the CDI specification. The DRA plugin generates the CDI spec, and the container runtime performs mounting according to the CDI spec.

### Deleting the Job<a name="section_claim_template_delete"></a>

After the `Deployment` is deleted, when each Pod is deleted, kubelet automatically calls the `UnprepareResourceClaims` callback of the DRA plugin to release devices and clean up CDI spec files.

1. Delete the `Deployment`.

   ```bash
   kubectl delete -f dra-test-deployment.yaml
   ```

   Command output:

   ```text
   deployment.apps "dra-test" deleted
   ```

2. Verify that `Unprepare` is complete.

   Check the DRA component logs to confirm that `UnprepareResourceClaims` is called and devices are released (the following logs are consistent with the `Infof` output in the code `plugin.go`):

   ```bash
   kubectl logs -n kube-system <dra-pod-name> | grep "UnprepareResourceClaims\|devices unprepared"
   ```

   Command output:

   ```text
   UnprepareResourceClaims is called, number of claims: 1
   devices unprepared for claim <claim-uid>
   ```

3. Confirm that the CDI spec file has been cleaned up.

   Run the following command on the node where the original Pod resides to confirm that the CDI spec file has been deleted:

   ```bash
   ls /var/run/cdi/
   ```

   The command output no longer contains the spec file corresponding to the deleted Claim.

4. Delete the `ResourceClaimTemplate`.

   When the job is no longer needed, delete the template:

   ```bash
   kubectl delete -f ascend-npu-claim-template.yaml
   ```

## Scheme 2: Requesting Resources by Using an Extended Resource Name<a name="section_scheme_extended_resource"></a>

No `ResourceClaim` is created. Instead, the extended resource name is directly configured in the `resources.requests` and `resources.limits` of the Pod to request unsliced physical NPU. kube-scheduler matches devices according to the `extendedResourceName` of the `DeviceClass`, automatically synthesizes a `ResourceClaim` and completes the allocation, and the allocation result is recorded in `pod.status.extendedResourceClaimStatus`.

This method supports only unsliced physical NPUs. `resources.requests` and `resources.limits` must be configured at the same time with equal values, and a value of `1` indicates that one unsliced physical NPU is requested. `static-vnpu.huawei.com` and `dynamic-vnpu.huawei.com` do not have an `extendedResourceName`, so vNPUs cannot be requested in this way. Use [Scheme 1](#section_scheme_claim_template) instead.

### Configuring the YAML File<a name="section_extended_config"></a>

1. Confirm the mapping between resource names and chip generations. The resource name must be equal to the `extendedResourceName` of a `DeviceClass`.

    **Table 1** Mapping between extended resource names and DeviceClasses

    <a name="table_dra_extended_resource"></a>

    | Extended Resource Name | DeviceClass | Applicable Chips |
    |--|--|--|
    | `huawei.com/Ascend910` | ascend910.ext.huawei.com | Ascend910 series, `chipName` starts with `910` or `Ascend910` |
    | `huawei.com/npu` | npu.ext.huawei.com | Ascend950 series, `chipName` starts with `Ascend950` |

    > [!NOTE]
    > Do not deploy Ascend Device Plugin on the same node. The extended resource names reported by Device Plugin are the same as those in this table. If both exist at the same time, the same card is counted twice, and a resource name conflict is triggered.

2. Create the `dra-extended-npu.yaml` file. The service Pod only needs to configure `requests` and `limits`, and does not configure `spec.resourceClaims`.

    <pre codetype="yaml">
    apiVersion: v1
    kind: Pod
    metadata:
      name: dra-extended-npu
      namespace: default
    spec:
      containers:
      - name: test
        image: ubuntu:22.04
        imagePullPolicy: IfNotPresent
        command:
        - /bin/bash
        - -c
        - sleep inf
        resources:
          <strong>requests:</strong>
            <strong>huawei.com/Ascend910: 1     # Request one unsliced physical full card. Modify as needed</strong>
          <strong>limits:</strong>
            <strong>huawei.com/Ascend910: 1     # Must be equal to requests</strong>
    </pre>

    For Ascend950 NPUs, change both `huawei.com/Ascend910` to `huawei.com/npu`.

### Delivering a Job<a name="section_extended_apply"></a>

1. Deliver the job.

   ```bash
   kubectl apply -f dra-extended-npu.yaml
   ```

2. Verify that the Pod enters the `Running` state and that the scheduler has completed resource allocation.

   ```bash
   kubectl get pod dra-extended-npu
   kubectl get pod dra-extended-npu -o jsonpath='{.status.extendedResourceClaimStatus.resourceClaimName}'
   ```

   The following is the output example. `Running` in the first `STATUS` indicates that the container has started, and a non-empty `resourceClaimName` in the second output indicates that the scheduler has completed resource allocation:

   ```text
   NAME                READY   STATUS    RESTARTS   AGE
   dra-extended-npu    1/1     Running   0          25s

   dra-extended-npu-abcde
   ```

3. Access the service Pod container to verify that the NPU device has been injected.

   ```bash
   kubectl exec -it dra-extended-npu -- ls /dev/davinci*
   ```

   The following is the output example:

   ```text
   /dev/davinci0
   ```

### Deleting the Job<a name="section_extended_delete"></a>

After the Pod is deleted, kubelet automatically calls the `UnprepareResourceClaims` callback of the DRA plugin to release devices and clean up CDI spec files. The `ResourceClaim` synthesized in this scheme is deleted by the cluster together with the Pod, and does not need to be deleted manually.

1. Delete the Pod.

   ```bash
   kubectl delete -f dra-extended-npu.yaml
   ```

   Command output:

   ```text
   pod "dra-extended-npu" deleted
   ```

2. Verify that the device has been released.

   ```bash
   kubectl logs -n ascend-dra-driver <dra-pod-name> | grep "UnprepareResourceClaims\|devices unprepared"
   ```

   Command output:

   ```text
   UnprepareResourceClaims is called, number of claims: 1
   devices unprepared for claim <claim-uid>
   ```

3. Confirm that the CDI spec file has been cleaned up.

   Run the following command on the node where the original Pod resides:

   ```bash
   ls /var/run/cdi/
   ```

   The command output no longer contains the spec file corresponding to the deleted Claim.

## Key Notes<a name="section_dra_notes"></a>

- **Idempotence**: Both `Prepare` and `Unprepare` operations are idempotent. Calling `Prepare` again on an already prepared Claim returns the same allocation result; calling `Unprepare` again on an already unprepared Claim is a no-op and does not report an error.
- **Checkpoint mechanism**: The DRA plugin persists the allocation relationship between Claims and devices through a checkpoint file (`checkpoint.json`). After the plugin is restarted, the allocation state is restored from the checkpoint, ensuring that already prepared Claims are not affected.
- **Deployment cleanup**: After the service-side resources are deleted, to uninstall the component, run `kubectl delete -f ascend-dra-driver.yaml`.
- **DeviceClass**: DRA automatically creates a DeviceClass named `npu.huawei.com` during deployment, and its selector matches all devices whose driver is `npu.huawei.com`. Users do not need to manually create a DeviceClass.
- **Relationship with Ascend Device Plugin**: DRA and Ascend Device Plugin are two different device access modes. Their functions are mutually exclusive on the same node, and they do not need to be deployed at the same time.
