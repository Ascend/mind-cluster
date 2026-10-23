# Resource Allocation and Release Under the DRA Mechanism<a name="ZH-CN_TOPIC_000000_dra_deploy_workflow"></a>

## Working Principle<a name="section_dra_principle"></a>

Ascend DRA manages NPU devices based on the Kubernetes Dynamic Resource Allocation mechanism. For the complete lifecycle (deployment and discovery, user submission, scheduling, Prepare, running, and Unprepare), see [Before You Start/Lifecycle](00_before_you_start.md#section_dra_life_cycle).

## Prerequisites<a name="section_dra_prereq"></a>

- Volcano and Ascend Docker Runtime have been installed and deployed. For details, see [Installation and Deployment](../../03_installation_guide/02_installation/00_helm_installation.md).
- The NPU driver and firmware have been installed on the NPU nodes, and the devices are running properly.
- Ascend DRA has been installed and deployed.

## Procedure<a name="section_dra_steps"></a>

### Creating a ResourceClaimTemplate<a name="section_create_template"></a>

Define NPU resource requests through a ResourceClaimTemplate. By referencing this template, a Pod automatically creates an independent ResourceClaim for each replica.

Create the `ascend-npu-claim-template.yaml` file:

```yaml
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
      - name: npu              # Request name, referenced in the Pod by this name
        deviceClassName: npu.huawei.com   # Reference the DeviceClass created when the DRA component is deployed
        count: 1               # Number of NPUs requested. Modify as needed
```

Run the following command to create the template:

```bash
kubectl apply -f ascend-npu-claim-template.yaml
```

Command output:

```text
resourceclaimtemplate.resource.k8s.io/ascend-npu-template created
```

> [!NOTE]
> The value of `deviceClassName` is `npu.huawei.com`. This DeviceClass is automatically created when `ascend-dra-driver.yaml` is deployed, and its selector matches all devices whose driver is `npu.huawei.com`.

### Delivering a Job and Triggering Prepare<a name="section_prepare"></a>

Deliver a Deployment that references the ResourceClaimTemplate. Each Pod replica automatically creates an independent ResourceClaim. After the Pod is scheduled to the NPU node, kubelet automatically calls the DRA plugin to complete Prepare.

Create the `dra-test-deployment.yaml` file:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dra-test
  namespace: default
  labels:
    app: dra-test
spec:
  replicas: 1                               # Number of replicas. Modify as needed
  selector:
    matchLabels:
      app: dra-test
  template:
    metadata:
      labels:
        app: dra-test
    spec:
      resourceClaims:
      - name: npu                           # Consistent with the name referenced in the container resources.claims
        resourceClaimTemplateName: ascend-npu-template
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
          - name: npu                      # Reference the name in resourceClaims above
```

1. Deliver the job.

   ```bash
   kubectl apply -f dra-test-deployment.yaml
   ```

   Command output:

   ```text
   deployment.apps/dra-test created
   ```

2. Check the Deployment status and confirm that the Pod enters the `Running` state.

   ```bash
   kubectl get pod -l app=dra-test -o wide
   ```

   Command output:

   ```text
   NAME                       READY   STATUS    RESTARTS   AGE   IP            NODE          NOMINATED NODE   READINESS GATES
   dra-test-<random-suffix>   1/1     Running   0          30s   10.244.1.20   node-npu-01   <none>           <none>
   ```

3. Verify that Prepare is complete.

   Check DRA logs to confirm that `PrepareResourceClaims` is called and devices are allocated (the following logs are consistent with the `Infof` output in the code `plugin.go` and `state.go`):

   ```bash
   kubectl logs -n ascend-dra-driver <dra-pod-name> | grep "PrepareResourceClaims\|devices prepared\|Returning newly prepared"
   ```

   The following is the output example (`<claim-uid>` is the UID of the ResourceClaim, `cdiIDs` is the CDI device ID list in the format `<driverName>/<class>=<index>`, and the value of `driverName` is `npu.huawei.com`):

   ```text
   PrepareResourceClaims is called, number of claims: 1
   devices prepared for claim <claim-uid>, count=1, cdiIDs=[npu.huawei.com/ascend910=0]
   Returning newly prepared devices for claim '<claim-uid>', count=1
   ```

4. View the generated CDI spec file.

   During the Prepare phase, the DRA plugin generates a CDI spec file for each Claim, located in the `/var/run/cdi` directory on the node. Run the following command on the node where the Pod resides:

   ```bash
   ls /var/run/cdi/
   ```

   The following is the output example, where a CDI spec file named after the Claim UID exists:

   ```text
   <claim-uid>.json
   ```

   View the content of the CDI spec file:

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
   - `kind`: CDI driver identifier, fixed to `https://ascend.com/npu`.
   - `devices`: Edit items divided by device granularity. Each `name` corresponds to the `/dev/davinci<id>` device node of one NPU.
   - `containerEdits` (spec level): Edit items shared by all devices, including shared device nodes (`davinci_manager`, `hisi_hdc`, and so on), the `LD_LIBRARY_PATH` environment variable, and driver library/tool mounts.

5. Access the service Pod container to verify that the NPU device has been injected.

   ```bash
   kubectl exec -it <pod-name> -- ls /dev/davinci*
   ```

   The following is the output example, where NPU device nodes are visible in the container:

   ```text
   /dev/davinci0
   ```

   > [!NOTE]
   > The NPU device node path is `/dev/davinci<id>`, where `<id>` is the physical device ID, consistent with the `physicId` attribute of the device in the ResourceSlice. The specific device injection is completed by the CDI specification. The DRA plugin generates the CDI spec, and the container runtime performs mounting according to the CDI spec.

### Deleting the Job and Triggering Unprepare<a name="section_unprepare"></a>

After the Deployment is deleted, when each Pod is deleted, kubelet automatically calls the `UnprepareResourceClaims` callback of the DRA plugin to release devices and clean up CDI spec files.

1. Delete the Deployment.

   ```bash
   kubectl delete -f dra-test-deployment.yaml
   ```

   Command output:

   ```text
   deployment.apps "dra-test" deleted
   ```

2. Verify that Unprepare is complete.

   Check the DRA component logs to confirm that `UnprepareResourceClaims` is called and devices are released (the following logs are consistent with the `Infof` output in the code `plugin.go`):

   ```bash
   kubectl logs -n ascend-dra-driver <dra-pod-name> | grep "UnprepareResourceClaims\|devices unprepared"
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

   The echo no longer contains the spec file corresponding to the deleted Claim.

### Cleaning Up Resources<a name="section_cleanup"></a>

To clean up the DRA component and related resources, run the following commands.

1. Delete the ResourceClaimTemplate.

   ```bash
   kubectl delete -f ascend-npu-claim-template.yaml
   ```

2. (Optional) Delete the DRA component.

   ```bash
   kubectl delete -f ascend-dra-driver.yaml
   ```

## Key Notes<a name="section_dra_notes"></a>

- **Idempotence**: Both Prepare and Unprepare operations are idempotent. Preparing an already prepared Claim again returns the same allocation result. Unpreparing an already unprepared Claim is a no-op and does not report an error.
- **Checkpoint mechanism**: The DRA plugin persists the allocation relationship between Claims and devices through a checkpoint file (`checkpoint.json`). After the plugin is restarted, the allocation state is restored from the checkpoint, ensuring that already prepared Claims are not affected.
- **DeviceClass**: The DRA component automatically creates a DeviceClass named `npu.huawei.com` during deployment, and its selector matches all devices whose driver is `npu.huawei.com`. Users do not need to manually create a DeviceClass.
- **Relationship with Ascend Device Plugin**: DRA and Ascend Device Plugin are two different device access methods and are mutually exclusive. Nodes using the DRA method do not need to deploy Ascend Device Plugin.
