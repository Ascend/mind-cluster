# Security Hardening<a name="ZH-CN_TOPIC_0000002493263486"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T06:27:21.707Z pushedAt=2026-06-09T07:15:15.730Z -->

## Hardening Notes<a name="ZH-CN_TOPIC_0000002511346367"></a>

The security hardening measures listed in this document are basic hardening recommendations. You should re-evaluate the network security hardening measures for the entire system based on your own business needs, and may refer to industry best hardening practices and security expert recommendations when necessary.

## Operating System Security Hardening<a name="ZH-CN_TOPIC_0000002511346385"></a>

### Firewall Configuration<a name="ZH-CN_TOPIC_0000002511426355"></a>

After an operating system is installed, if a regular user is configured, you can prevent unauthorized operations by adding the `ALWAYS_SET_PATH` field in the `/etc/login.defs` file and setting it to `yes`.

### Setting umask<a name="ZH-CN_TOPIC_0000002511426335"></a>

It is recommended to set `umask` on the host and in containers to `027` or higher to enhance file permissions.

The following uses setting `umask` to `027` as an example to describe the specific operations.

1. Log in to the server as the `root` user and edit the `"/etc/profile"` file.

    ```shell
    vim /etc/profile
    ```

2. Add `umask 027` at the end of the `"/etc/profile"` file, save and exit.
3. Run the following command to make the configuration take effect.

    ```shell
    source /etc/profile
    ```

### Security Hardening for Ownerless Files<a name="ZH-CN_TOPIC_0000002479386436"></a>

Because of differences between the official Docker image and the operating system on the physical machine, users in the system may not have a one-to-one correspondence, which may cause files generated during the running of the physical machine or container to become files without an owner.

You can run the `find / -nouser -o -nogroup` command to find files without an owner in the container or on the physical machine. Create corresponding users and user groups based on the UID and GID of the files, or modify the UID of existing users and the GID of user groups to adapt, assign an owner to the files, and prevent files without an owner from posing security risks to the system.

### Port Scanning<a name="ZH-CN_TOPIC_0000002511346359"></a>

Pay attention to all listening ports and unnecessary ports on the entire network. If there are any unnecessary ports, close them promptly. It is recommended that you disable insecure services, such as Telnet and FTP. For specific disabling methods, refer to the relevant documentation of the operating system in use.

### DoS Attack Prevention<a name="ZH-CN_TOPIC_0000002511426325"></a>

You can prevent DoS attacks on the system by limiting the connection rate to the server based on IP addresses. Methods include, but are not limited to, using the built-in Iptables firewall of the Linux system for prevention and optimizing sysctl parameters. For specific usage methods, consult relevant materials.

### sudo Configuration<a name="ZH-CN_TOPIC_0000002511346403"></a>

- Set the `targetpw` option in the *sudo command` to require the target user's password by default. This prevents all users from escalating to the root account and executing system commands without entering the root password after adding sudo rules, which could lead to unauthorized command execution by ordinary users. This option is not added by default; it is recommended to add it.

    Run the `cat /etc/sudoers | grep -E "^[^#]*Defaults[[:space:]]+targetpw"` command to check if the `"Defaults targetpw"` or `"Defaults rootpw"` configuration entry exists. If not exist, add the `"Defaults targetpw"` or `"Defaults rootpw"` configuration entry under the `"#Defaults specification"` section in the `"/etc/sudoers"` file.

- Prohibit ordinary users or groups from escalating to the root user through all commands.

    Run the `cat /etc/sudoers` command to check if there are any `(ALL) ALL` or `(ALL:ALL) ALL` entries for users or groups other than `root ALL=(ALL:ALL) ALL` and "`root ALL=(ALL) ALL`" in the `"/etc/sudoers"` file. If such entries exist, confirm whether they are needed based on actual business scenarios. If they are not needed, delete them, for example: `user ALL=(ALL) ALL`, `%admin ALL=(ALL) ALL`, or `%sudo ALL=(ALL:ALL) ALL`.

### Ability to Resist Vulnerabilities<a name="ZH-CN_TOPIC_0000002479386454"></a>

Use the Linux built-in ASLR (Address Space Layout Randomization) feature to strengthen protection against vulnerabilities.

Write 2` to the `"/proc/sys/kernel/randomize_va_space"` file.

## Security Hardening for Docker<a name="ZH-CN_TOPIC_0000002479386408"></a>

### Enable Auditing for Docker<a name="ZH-CN_TOPIC_0000002479226438"></a>

**Audit Content<a name="zh-cn_topic_0000001446964952_section1555621244310"></a>**

- The Docker daemon runs with root privileges on the host, granting it significant permissions. It is recommended to configure an auditing mechanism on the host for the operation and usage status of the Docker daemon. If the Docker daemon engages in privilege escalation attacks, the source of the attack can be traced. To enable auditing, see [Enabling Audit for Docker](#enable-auditing-for-docker).

- The following directories store important information related to containers. It is recommended to configure auditing for these directories and key files.

    - `/usr/bin/dockerd`
    - `/var/lib/docker`
    - `/etc/docker`
    - `/etc/default/docker`
    - `/etc/sysconfig/docker`
    - `/etc/docker/daemon.json`
    - `/usr/bin/docker-containerd`
    - `/usr/bin/containerd`
    - `/usr/bin/docker-runc`
    - `docker.service`
    - `docker.socket`

    The above directories are the default installation directories for Docker. If a separate partition has been created for Docker, the paths may change. To enable auditing, see [Enabling Audit for Docker](#enabling-audit-for-docker).

**Enabling Audit for Docker<a name="enabling-audit-for-docker"></a>**

Enabling the audit function will allow the system to collect logs and other related information. By default, the host does not have the audit function enabled. You can add audit rules using the following method.

>[!NOTE]
>Enabling the audit mechanism requires installing auditd. For example, on Ubuntu, you can use the `apt install -y auditd` command for installation.

1. Add rules to the `"/etc/audit/audit.rules"` file, with each rule on a separate line. The rule format is as follows.

    ```shell
    -w file_path -k docker
    ```

    **Table 1**  Parameters

    |Parameter|Description|
    |--|--|
    |`-w`|Filters by file path.|
    |`file_path`|The file path for which the audit rule is enabled, for example: <ul><li>When `file_path` is `/usr/bin/docker`, it indicates enabling auditing of the Docker daemon on the host.</li><li>When `file_path` is `/etc/docker`, it indicates enabling auditing of Docker-related directories and key files on the host.</li></ul>|
    |`-k`|Filters by a specified keyword string.|

    >[!NOTE]
    >If the `"/etc/audit/audit.rules"` file contains `"This file is automatically generated from /etc/audit/rules.d"`, modifying the `"/etc/audit/audit.rules"` file is ineffective; you need to modify the `"/etc/audit/rules.d/audit.rules"` file for changes to take effect. For example, on Ubuntu systems, you need to modify the `"/etc/audit/rules.d/audit.rules"` file.

2. After configuration, restart the log daemon.

    ```shell
    service auditd restart
    ```

### Setting Docker Configuration File Permissions<a name="ZH-CN_TOPIC_0000002479226446"></a>

**TLS CA Certificate Permission Configuration<a name="zh-cn_topic_0000001497205397_zh-cn_topic_0273637460_section776865651617"></a>**

- Set the owner and group of the TLS CA certificate file to `root:root`, and set the permissions to `400`.

    Protect the TLS CA certificate file (the path to the CA certificate file is specified using the `--tlscacert` parameter) to prevent tampering. The certificate file is used to authenticate the Docker server with the specified CA certificate. Therefore, its owner and group must be `root`, and its permissions must be `400` to ensure the integrity of the CA certificate.

    This can be configured as follows.

    1. Execute the following command to set the file's owner and group to `root`.

        ```shell
        chown root:root <path to TLS CA certificate file>
        ```

        >[!NOTE]
        >The *path to TLS CA certificate file* is typically `"/usr/local/share/ca-certificates"`.

    2. Set the file permissions to `400`.

        ```shell
        chmod 400 <path to TLS CA certificate file>
        ```

**File Permission Configuration for "/etc/docker/daemon.json"<a name="zh-cn_topic_0000001497205397_section2824195145813"></a>**

- Set the owner and group of the `"daemon.json"` file to `root:root`, and set the file permissions to `600`.

    The `"daemon.json"` file contains sensitive parameters that change the Docker daemon. It is an important global configuration file. Its owner and group must be `root`, and it must be writable only by `root` to ensure the integrity of the file. This file does not exist by default.

    - If the `"daemon.json"` file does not exist by default, it indicates that the product is not using this file for configuration. You can execute the following command to set the configuration file to empty in the startup parameters, so that this file is not used as the default configuration file, preventing attackers from maliciously creating and modifying the configuration.

        ```shell
        docker --config-file=""
        ```

    - If the `"daemon.json"` file exists in the product environment, it indicates that this file has been used for configuration operations. You need to set the corresponding permissions to prevent malicious modification.
        1. Run the following command to set the file owner and group to `root`.

            ```shell
            chown root:root /etc/docker/daemon.json
            ```

        2. Run the following command to set the file permissions to `600`.

            ```shell
            chmod 600 /etc/docker/daemon.json
            ```

**Docker-Related Directory and File Permission Control<a name="zh-cn_topic_0000001497205397_section1997971395714"></a>**

**Table 1** Docker-related directory and file permission control

|Directory|File Owner|File Permissions|
|--|--|--|
|/etc/default/docker|root:root|644 or stricter|
|/etc/sysconfig/docker|root:root|644 or stricter|
|docker.service|root:root|644|
|docker.sock|root:docker|660|
|/etc/docker|root:root|755 or stricter|
|docker.socket|root:root|644 or stricter|

> [!NOTE]
> If the file or directory does not exist, it can be ignored.

### Controlling the Running User of Docker Containers<a name="ZH-CN_TOPIC_0000002511426373"></a>

It is recommended that you run containers as a non-root user or as a non-privileged root user. when using Docker.

### Disabling Insecure Protocols in Containers<a name="ZH-CN_TOPIC_0000002511426313"></a>

To avoid security risks, it is recommended that you use secure protocols, such as SSHv2, TLS 1.2, TLS 1.3, IPsec, SFTP, and SNMPv3. If insecure protocols are used in containers, such as Telnet, FTP, SSH v1.x, TFTP, SNMPv1, SNMPv2c, SSL 2.0, SSL 3.0, and TLS 1.0, it is recommended to disable them or replace them with secure protocols without affecting normal business operations.

### Creating a Separate Partition for Docker<a name="ZH-CN_TOPIC_0000002511426363"></a>

After Docker is installed, the default directory is `"/var/lib/docker"`, which is used to store Docker-related files, including images, containers, etc. When this directory is full, Docker and the host may become unusable. Therefore, it is recommended to create a separate partition (logical volume) to store Docker files.

- For newly installed devices, create a separate partition to mount the `"/var/lib/docker"` directory. Refer to [Operating System Drive Partitioning](./installation_guide/01_environment_dependencies.md).
- For systems that have already been installed, use the Logical Volume Manager (LVM) to create a partition.

### Limiting Container File Handles and Fork Processes<a name="ZH-CN_TOPIC_0000002479386446"></a>

To prevent attackers from using commands inside a container to launch fork bombs, causing denial of service, it is recommended that you set a global default ulimit to restrict the number of file handles and processes created.

1. Open the configuration file.
    - For CentOS 7.6, the default is the `"/usr/lib/systemd/system/docker.service"` file.
    - For Ubuntu 22.04, the default is the `"/lib/systemd/system/docker.service"` file.

2. Modify the configuration file.

    Locate the line containing `"/usr/bin/dockerd"` in the configuration file, and add the `nofile` (created file handles) parameter and `nproc` (processes) parameter limits after that line.

    The following is a modification example. Please set the corresponding values based on the actual situation.

    ```shell
    ...
    # the default is not to use systemd for cgroups because the delegate issues still
    # exists and systemd currently does not support the cgroup feature set required
    # for containers run by docker
    /usr/bin/dockerd --default-ulimit nofile=20480:40960 --default-ulimit nproc=1024:2048
    ...
    ```

    `--default-ulimit nproc=1024:2048` indicates that the number of processes is limited to 1,024. This value can be modified within the process, but cannot exceed 2,048, and the first value must be less than or equal to the second value. The meaning of the `nofile` configuration is the same as `nproc`.

### Security Hardening for Container Images<a name="ZH-CN_TOPIC_0000002479226470"></a>

- It is recommended to create a non-root user in the base image, start the image and processes as that non-root user, and grant the user only the necessary capabilities to avoid security risks such as container escape caused by high-privilege users.
- Properly control the owner and permissions of files in the image to prevent security risks such as container escape caused by unnecessary privilege escalation.
- Promptly fix vulnerabilities in the base image.
- During image distribution, it is recommended to enable Docker's Content Trust feature.
- Avoid using the ADD instruction in the Dockerfile. Using ADD to handle files from unknown sources poses a security risk.
- Avoid storing sensitive information in the Dockerfile.
- Avoid using the update command alone.
- Add a health check mechanism for containers, and verify the validity of the script or command specified by the health check. Ensure that the script or command does not cause service or system exceptions.
- Avoid including files and directories with SUID and SGID permissions in containers.
- Set resource quotas for containers to prevent them from consuming excessive system resources. System resources include but are not limited to CPU and memory.

### Enabling Live Restore<a name="ZH-CN_TOPIC_0000002479386398"></a>

By default, when the Docker daemon terminates, it shuts down running containers. After enabling this feature, containers continue running when the daemon becomes unavailable. For specific configuration, refer to the [official Docker documentation](https://docs.docker.com/config/containers/live-restore/).

### Restricting Uncontrolled Inter-Container Network Communication<a name="ZH-CN_TOPIC_0000002479386420"></a>

By default, network communication between all containers on the same host is not restricted. Therefore, each container can read all packets on the container network of the same host, which may lead to unintentional information leakage to other containers. Therefore, it is recommended to restrict communication between containers.

Modify the Docker startup parameters and add the `"--icc=false"` parameter to disable communication between containers. An example is shown below.

```shell
……
[Service]
Type=notify
# the default is not to use systemd for cgroups because the delegate issues still
# exists and systemd currently does not support the cgroup feature set required
# for containers run by docker
ExecStart=/usr/bin/dockerd  --userland-proxy=false --icc=false -H fd:// --containerd=/run/containerd/containerd.sock
ExecReload=/bin/kill -s HUP $MAINPID
TimeoutSec=0
RestartSec=2
Restart=always
……
```

### Disabling Userland Proxy<a name="ZH-CN_TOPIC_0000002479226462"></a>

It is recommended to add the `"--userland-proxy=false"` parameter to disable the userland proxy at startup, reducing the device's attack surface. An example is shown below.

```shell
……
[Service]
Type=notify
# the default is not to use systemd for cgroups because the delegate issues still
# exists and systemd currently does not support the cgroup feature set required
# for containers run by docker
ExecStart=/usr/bin/dockerd  --userland-proxy=false --icc=false -H fd:// --containerd=/run/containerd/containerd.sock
ExecReload=/bin/kill -s HUP $MAINPID
TimeoutSec=0
RestartSec=2
Restart=always
……
```

## Container Security Hardening<a name="ZH-CN_TOPIC_0000002511426345"></a>

You are advised to perform the following operations in the production environment to harden images.

- Create a non-root user in the base image, start the image and processes as the non-root user, and grant the user only the necessary capabilities to avoid security risks such as container escape caused by high-privilege users.
- Properly control the owner and permissions of files in the image to avoid security risks such as container escape caused by unnecessary unauthorized access.
- Promptly fix vulnerabilities in the base image.
- During image distribution, it is recommended to enable Docker's Content Trust feature.

**Seccomp Configuration<a name="section1330962354513"></a>**

The seccomp configuration can constrain the system calls of a container, reducing the container's attack surface on the system. For details, see [seccomp.md](https://github.com/kubernetes/design-proposals-archive/blob/main/node/seccomp.md).

In Kubernetes versions earlier than 1.19, seccomp uses the `annotations[seccomp.security.alpha.kubernetes.io/pod]` method. In version 1.19 and later, the seccomp feature has reached GA. For version 1.19 and later, it is recommended to use `securityContext.seccompProfile` for configuration, and starting from version 1.27, the annotation method is no longer effective. For details, see: [Kubernetes Removals and Major Changes In v1.27](https://kubernetes.io/blog/2023/03/17/upcoming-changes-in-kubernetes-v1-27/#support-for-deprecated-seccomp-annotations). Therefore, modify the seccomp configuration based on different Kubernetes versions and container security requirements.

There are two configuration methods provided to modify thehe MindCluster component settings. The following is the seccomp configuration for Resilience Controller, and other components have reserved related configurations.

>[!NOTE]
>
>- All MindCluster components, except for Elastic Agent and TaskD, require modification of the startup configuration file.
>- For descriptions of each component's configuration file, see [Table 1](./installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md).

```Yaml
    metadata:
      labels:
        app: resilience-controller
      ##### For Kubernetes versions lower than 1.19, seccomp is used with annotations.
      annotations:
        seccomp.security.alpha.kubernetes.io/pod: runtime/default
    spec:
      ##### For Kubernetes versions 1.19 and above, seccomp is used with securityContext.
#      securityContext:
#        seccompProfile:
#          type: RuntimeDefault
...
```

## Kubernetes Security Hardening<a name="ZH-CN_TOPIC_0000002479386428"></a>

Kubernetes requires the following hardening:

- kube-apiserver:
    - Modify the startup parameter `--profiling` and set its value to `false` to prevent users from dynamically changing the kube-apiserver log level.
    - Modify or add the startup parameter `--tls-cipher-suites` and set its value as follows to avoid risks caused by using insecure TLS cipher suites.

        ```shell
        --tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
        ```

    - Modify or add the startup parameter `--audit-policy-file` to configure the Kubernetes audit policy. For specific configuration, refer to the [Kubernetes official documentation](https://kubernetes.io/docs/tasks/debug/debug-cluster/audit/).

- kubelet:
    - To prevent a single Pod from occupying too many processes, you can enable `SupportPodPidsLimit` and set `--pod-max-pids`. Add `--feature-gates=SupportPodPidsLimit=true --pod-max-pids=<max pid number>` to the `KUBELET_KUBEADM_ARGS` item in the kubelet configuration file. After modifying the configuration, restart to take effect. For details, refer to the [Kubernetes official documentation](https://kubernetes.io/docs/concepts/policy/pid-limiting/).
    - Modify or add the startup parameter `"--tls-cipher-suites"` and set its value as follows to avoid risks from insecure TLS cipher suites.

        ```shell
        --tls-cipher-suites=TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
        ```

        >[!NOTE]
        >Kubernetes v1.19 and later versions support TLS v1.3 cipher suites. It is recommended to add TLS v1.3 cipher suites when using a higher version of Kubernetes.

- If the OS kernel version used by the Kubernetes cluster is greater than or equal to 4.6, manually enable AppArmor or SELinux after installing Kubernetes.
- For other security hardening content, refer to the relevant content in the Kubernetes official documentation [Security](https://kubernetes.io/docs/concepts/security/), or refer to other excellent hardening solutions in the industry.
- Configure appropriate permissions for upper-layer business platforms in Kubernetes, such as restricting the API groups that accounts can access, to prevent upper-layer businesses from operating unnecessary Kubernetes resources. For details, refer to the [official documentation](https://kubernetes.io/docs/reference/access-authn-authz/).

## ClusterD Security Hardening<a name="ZH-CN_TOPIC_0000002479226456"></a>

After ClusterD starts, it launches a gRPC server to listen for messages from the gRPC client within the training container, enabling the resumable training feature. By default, ClusterD uses non-secure gRPC communication. You can adopt TLS/SSL encrypted communication to prevent attacks during the communication process.

The following uses mutual authentication between ClusterD and NodeD as an example to introduce the detailed steps for ClusterD security hardening. In this example, ClusterD acts as the server and NodeD acts as the client.

**Prerequisites<a name="section10669832105619"></a>**

Before performing mutual authentication, prepare the following certificate files.

- rootCA.crt
- client.crt
- client.key
- server.crt
- server.key

**Procedure<a name="section2636144475610"></a>**

1. Pull the nginx image.

    ```shell
    docker pull nginx
    ```

2. <a name="li188871536105918"></a>Create a folder named `cert` under path A, and place the certificate files `rootCA.crt`, `server.crt`, and `server.key` listed in the [Prerequisites](#section10669832105619) into the `cert` folder.
3. <a name="li463812451313"></a>Create a new folder named `conf` under path A, create a file named `nginx.conf` in that folder, and write the following content into the file:

    ```conf
       worker_processes 1;
       worker_cpu_affinity 0001;

       worker_rlimit_nofile 4096;
       events {
           worker_connections 4096;
       }

       http {
        port_in_redirect off;
        server_tokens off;
        autoindex off;

        access_log /var/log/nginx/access.log;
        error_log /var/log/nginx/error.log info;

        limit_req_zone global zone=req_zone:100m rate=20r/s;
        limit_conn_zone global zone=north_conn_zone:100m;

        server {
         listen <ClusterD的pod IP>:9500 ssl;  # The Pod IP address of ClusterD. The port should be consistent with the port in the ClusterD configuration file. If it is an IPv6 address, configure it as [<ClusterD pod IP>]:9500 ssl;
         http2 on;


         proxy_ssl_session_reuse off;

         add_header Referrer-Policy "no-referrer";
         add_header X-XSS-Protection "1; mode=block";
         add_header X-Frame-Options DENY;
         add_header X-Content-Type-Options nosniff;
         add_header Strict-Transport-Security " max-age=31536000; includeSubDomains ";
         add_header Content-Security-Policy "default-src 'self'";
         add_header Cache-control "no-cache, no-store, must-revalidate";
         add_header Pragma no-cache;
         add_header Expires 0;

         ssl_session_tickets off;

         ssl_certificate     /etc/nginx/conf.d/cert/server.crt;                     # Server certificate path (permissions 400)cate path (permissions 400)
         ssl_certificate_key /etc/nginx/conf.d/cert/server.key;              # Server private key path, which cannot be configured in plain text (permissions 400).
         ssl_client_certificate /etc/nginx/conf.d/cert/rootCA.crt;
         ssl_verify_client on;
         ssl_verify_depth 2;
         send_timeout 60;

         limit_req zone=req_zone burst=20 nodelay;
         limit_conn north_conn_zone 20;
         keepalive_timeout  60;
         proxy_read_timeout 900;
         proxy_connect_timeout   60;
         proxy_send_timeout      60;
         client_header_timeout   60;
         client_body_timeout 10;
         client_header_buffer_size  2k;
         large_client_header_buffers 4 8k;
         client_body_buffer_size 16K;
         client_max_body_size 20m;
         ssl_protocols TLSv1.2 TLSv1.3;
         ssl_ciphers "ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256 !aNULL !eNULL !LOW !3DES !MD5 !EXP !PSK !SRP !DSS !RC4";
         ssl_session_timeout 10s;
         ssl_session_cache shared:SSL:10m;

         location / {
          grpc_pass grpc://<ClusterD的pod IP>:8899;                    # The Pod IP address of ClusterD. If the `useProxy` startup parameter of ClusterD is enabled, the IP address here is 127.0.0.1. If it is an IPv6 address, configure it as `grpc://[<ClusterD pod IP>]:8899;`
           }
        }
       }
    ```

4. Modify or add the following bolded fields in the ClusterD startup YAML file.

    ```shell
    # Configure -useProxy=true in the ClusterD startup command to enable the local proxy
       args: [ "/usr/local/bin/clusterd -logFile=/var/log/mindx-dl/clusterd/clusterd.log -logLevel=0 <strong>-useProxy=true"</strong> ]
    # Add to the containers section in Deployment
               <strong>- name: nginx</strong>
                 <strong>image: nginx:latest</strong>
                 <strong>imagePullPolicy: Never</strong>
                 <strong>command: [ "/bin/bash", "-c", "--"]</strong>
                 <strong>args: [ "sleep infinity" ]</strong>
                 <strong>volumeMounts:</strong>
                   <strong>- name: nginx-cert</strong>
                     <strong>mountPath: /etc/nginx/conf.d/cert</strong>
                   <strong>- name: nginx-conf</strong>
                     <strong>mountPath: /etc/nginx/conf</strong>

    # Add to the volumes section in Deployment
               <strong>- name: nginx-cert</strong>
                 <strong>hostPath:</strong>
                   <strong>path: /{PathA}/cert           # x509 certificate and private key directory path. Replace PathA with the file path from Step 2</strong>
               <strong>- name: nginx-conf</strong>
                 <strong>hostPath:</strong>
                   <strong>path: /{PathA}/config       # nginx startup configuration file. Replace PathA with the file path from Step 2</strong>

    # Change the ports section in Service to the following
           <strong>- protocol: TCP</strong>
             <strong>port: 8899</strong>
             <strong>targetPort: 9500</strong></pre>
    ```

5. Run the following command to start the ClusterD service.

    ```shell
    kubectl apply -f clusterd-v{version}.yaml
    ```

6. Run the following command to view the Pod IP of ClusterD, and write the queried Pod IP into the `nginx.conf` file in [Step 3](#li463812451313).

    ```shell
    kubectl get pod -A -o wide | grep clusterd
    ```

7. Start nginx.

    ```shell
    ## Enter the nginx container
    kubectl exec -it -n mindx-dl clusterd-{xxx} -c nginx bash      #Replace {xxx} with the Pod ID randomly generated by K8s after the ClusterD Pod starts.
    ## Run the following command to start nginx and enter the key passphrase when prompted.
    nginx -c /etc/nginx/conf/nginx.conf
    ```

8. Start the NodeD service.
    1. Create a folder named `cert` in path B, and place the `rootCA.crt`, `client.crt`, and `client.key` files listed in the [Prerequisites](#section10669832105619) into the `cert` folder.
    2. Create a folder named `conf` in path B, create a file named `nginx.conf` in that folder, and write the following content into the file:

        ```conf
           worker_processes 1;
           worker_cpu_affinity 0001;

           worker_rlimit_nofile 4096;
           events {
               worker_connections 4096;
           }

           http {
            port_in_redirect off;
            server_tokens off;
            autoindex off;

            access_log /var/log/nginx/access.log;
            error_log /var/log/nginx/error.log;

            grpc_buffer_size 16M;

            limit_req_zone global zone=req_zone:100m rate=20r/s;
            limit_conn_zone global zone=north_conn_zone:100m;

            server {
             listen 127.0.0.1:8899;
             http2 on;

             ssl_session_tickets off;

             limit_req zone=req_zone burst=20 nodelay;
             limit_conn north_conn_zone 20;
             keepalive_timeout  60;
             proxy_read_timeout 900;
             proxy_connect_timeout   60;
             proxy_send_timeout      60;
             client_header_timeout   60;
             client_body_timeout 10;
             client_header_buffer_size  200k;
             large_client_header_buffers 4 800k;
             client_body_buffer_size 160K;
             client_max_body_size 20m;

             location / {
              grpc_pass grpcs://<ClusterD的service IP>:9500;                    # Service IP address of ClusterD, which can be queried using the following command: kubectl get svc -A | grep clusterd. If it is an IPv6 address, configure it as grpcs://[<ClusterD service IP>]:9500;
              grpc_ssl_verify on;
              grpc_ssl_trusted_certificate /etc/nginx/conf.d/cert/rootCA.crt;
              grpc_ssl_verify_depth 2;
              grpc_ssl_certificate /etc/nginx/conf.d/cert/client.crt;
              grpc_ssl_certificate_key /etc/nginx/conf.d/cert/client.key;
              grpc_ssl_protocols TLSv1.2 TLSv1.3;
              grpc_ssl_ciphers "ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256 !aNULL !eNULL !LOW !3DES !MD5 !EXP !PSK !SRP !DSS !RC4";
              grpc_ssl_name <SAN or CN of the service certificate>;     # SAN or CN of the service certificate
               }
            }
           }
        ```

    3. Add the following bold fields to the NodeD startup YAML file.

        ```shell
        # Add `sleep 150`
                args: [ "<strong>sleep 150;</strong> /usr/local/bin/noded -logFile=/var/log/mindx-dl/noded/noded.log -logLevel=0" ]

        # Add to containers section
                   <strong>- name: nginx</strong>
                     <strong>image: nginx:latest</strong>
                     <strong>imagePullPolicy: Never</strong>
                     <strong>command: [ "/bin/bash", "-c", "--"]</strong>
                     <strong>args: [ "sleep infinity" ]</strong>
                     <strong>volumeMounts:</strong>
                       <strong>- name: nginx-cert</strong>
                         <strong>mountPath: /etc/nginx/conf.d/cert</strong>
                       <strong>- name: nginx-conf</strong>
                         <strong>mountPath: /etc/nginx/conf</strong>

        # Add to volumes item
                   <strong>- name: nginx-cert</strong>
                     <strong>hostPath:</strong>
                       <strong>path: /{Path B}/cert          # x509 certificate, private key directory path</strong>
                   <strong>- name: nginx-conf</strong>
                     <strong>hostPath:</strong>
                       <strong>path: /{Path B}/config      # nginx startup configuration file</strong></pre>
        ```

    4. Run the following command to start NodeD.

        ```shell
        kubectl apply -f noded-v{version}.yaml
        ```

    5. Enter the NodeD container and add a domain name resolution rule.

        ```shell
        ## Enter the noded container
        kubectl exec -it -n <noded pod ns> <noded pod name> bash
        ## Add a new domain name mapping rule
        echo 127.0.0.1 clusterd-grpc-svc.mindx-dl.svc.cluster.local >> /etc/hosts
        ```

    6. Start nginx.

        ```shell
        ## Enter the nginx container
        kubectl exec -it -n mindx-dl noded-{xxx} -c nginx bash      # {xxx} represents the Pod id randomly generated by K8s after the NodeD Pod starts
        ## Start nginx
        nginx -c /etc/nginx/conf/nginx.conf
        ```

## TaskD Security Hardening<a name="ZH-CN_TOPIC_0000002511346377"></a>

After TaskD starts, it launches a gRPC client to communicate with ClusterD via gRPC. Meanwhile, gRPC communication also occurs among TaskD's internal components (Manager, Proxy, Agent, Worker). By default, TaskD uses insecure gRPC communication. You can adopt TLS/SSL encrypted communication to prevent the communication process from being attacked.

The following uses nginx as an example to guide you in encrypting and authenticating cross-node communication of TaskD through a local network proxy.

**Prerequisites<a name="section106698321056192"></a>**

Before performing mutual authentication, prepare the following certificate files.

- rootCA.crt
- client.crt
- client.key
- server.crt
- server.key

**Procedure<a name="section39311920145712"></a>**

1. Pull the nginx image.

    ```shell
    docker pull nginx
    ```

2. <a name="li115126401711"></a>Place all certificate files listed in the [Prerequisites](#section106698321056192) into path A.
3. Prepare the nginx proxy configuration file for the master pod. Create a new folder named `conf` under path A, create a file named `master_nginx.conf` in that folder, and write the following content into the file:

    ```conf
    worker_processes 1;
    worker_cpu_affinity 0001;

    worker_rlimit_nofile 4096;
    events {
     worker_connections 4096;
    }
    http {
     access_log /etc/nginx/access.log;
     error_log /etc/nginx/error.log;
     server {
      listen 127.0.0.1:8899;
                    http2 on;
      location / {
       grpc_pass grpcs://{ClusterD的Pod IP}:9500;      # If it is an IPv6 address, configure it as grpcs://[{Pod IP of ClusterD}]:9500;0;
       grpc_ssl_verify on;
       grpc_ssl_trusted_certificate /etc/nginx/rootCA.crt;
       grpc_ssl_certificate /etc/nginx/client.crt;
       grpc_ssl_certificate_key /etc/nginx/client.key;
       grpc_ssl_verify_depth 2;
       grpc_ssl_protocols TLSv1.2 TLSv1.3;
       grpc_ssl_ciphers "ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256 !aNULL !eNULL !LOW !3DES !MD5 !EXP !PSK !SRP !DSS !RC4";
       grpc_ssl_name {SAN or CN of the service certificate};
      }
     }

            # Single-pod tasks do not need to configure the following server
     server {
      listen {master Pod的IP}:9601 ssl;                # If it is an IPv6 address, configure it as [{IP of the master Pod}]:9601 ssl;
      proxy_ssl_session_reuse off;
      http2 on;
      ssl_certificate     /etc/nginx/server.crt;      # Server certificate path (permissions 400)
      ssl_certificate_key /etc/nginx/server.key;      # Path to the server private key. The private key cannot be configured in plain text (permissions 400).
      ssl_client_certificate /etc/nginx/rootCA.crt;   # Path to the root certificate
      ssl_verify_client on;
      ssl_verify_depth 2;
      ssl_protocols TLSv1.2 TLSv1.3;
      ssl_ciphers "ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256 !aNULL !eNULL !LOW !3DES !MD5 !EXP !PSK !SRP !DSS !RC4";
      location / {
       grpc_pass grpc://127.0.0.1:9601;
      }
     }
    }
    ```

4. (For single-pod tasks, skip this step.) Prepare the nginx proxy configuration file for the worker pod. Create a new folder named `conf` under Path A, create a new file named `worker_nginx.conf` in that folder, and write the following content into the file.

    ```conf
    worker_processes 1;
    worker_cpu_affinity 0001;
    worker_rlimit_nofile 4096;
    events {
     worker_connections 4096;
    }
    http {
     access_log /etc/nginx/access.log;
     error_log /etc/nginx/error.log;
     server {
      listen 127.0.0.1:9601;
                    http2 on;
      location / {
       grpc_pass grpcs://{master svc ip}:9601;  # The svc IP address can be queried using the command `kubectl get svc -A |grep {jobname}`. If it is an IPv6 address, configure it as `grpcs://[{master svc ip}]:9601`;
       grpc_ssl_verify on;
       grpc_ssl_trusted_certificate /etc/nginx/rootCA.crt;
       grpc_ssl_certificate /etc/nginx/client.crt;
       grpc_ssl_certificate_key /etc/nginx/client.key;
       grpc_ssl_verify_depth 2;
       grpc_ssl_protocols TLSv1.2 TLSv1.3;
       grpc_ssl_ciphers "ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-GCM-SHA256 !aNULL !eNULL !LOW !3DES !MD5 !EXP !PSK !SRP !DSS !RC4";
       grpc_ssl_name DomainCA.com;
      }
     }
    }
    ```

5. Inject the environment variable for using the local proxy in the task YAML.

    <pre codetype="yaml">
              env:
                  - name: TTP_PORT
                    value: "8000"
                  - name: <strong>LOCAL_PROXY_ENABLE</strong>
                    value: "<strong>on</strong>"          # Switch for using local proxy communication</pre>

6. Add the following fields to the task pod.

    ```Yaml
        # Add to the containers section in the Deployment
               - name: nginx
                 image: nginx:latest
                 imagePullPolicy: Never
                 command: [ "/bin/bash", "-c", "--"]
                 args: [ "sleep infinity" ]
                 volumeMounts:
                   - name: nginx-conf
                     mountPath: /etc/nginx

       # Add to the volumes section in the Deployment
               - name: nginx-conf
                 hostPath:
                   path: /{Path A}/       # Path where the nginx startup configuration file and certificate key file are located. Replace path A with the file path from step 2.

    ```

7. (Skip this step for single-pod tasks.) Access the task pod to start nginx, including the master pod nginx and worker pod nginx.

    ```shell
    ## Access the nginx container.
    kubectl exec -it -n {Task  namespace} {Task pod name} -c nginx bash
    ## Run the following command to start nginx and enter the key passphrase when prompted.
    nginx -c /etc/nginx/conf/{master or worker}_nginx.conf
    ```

## Elastic Agent Security Hardening<a name="ZH-CN_TOPIC_0000002511346397"></a>

> [!NOTE]
> Elastic Agent has reached its end of life, and related materials will be removed in the version released on December 30, 2026.

For security hardening of Elastic Agent, see the [TaskD Security Hardening](#taskd-security-hardening) section.

## Viewing Command Line Operation Records<a name="ZH-CN_TOPIC_0000002524473029"></a>

Command line operation logs are recorded in the system history.

**Viewing the History of Executed Commands<a name="section1220492120526"></a>**

When installing, upgrading, or uninstalling Container Manager, or when querying container recovery progress through Container Manager, the historical command records from history will be saved to the `"~/.bash_history"` file. Therefore, you can directly view the `/.bash_history` file to find the command line records.

Historical commands are first cached in memory and are only written to the `"~/.bash_history"` file when the terminal exits normally. Execute the following command to immediately write the historical records from memory to the `"~/.bash_history"` file:

```shell
history -a
```

**Modifying the Number of Saved Historical Records<a name="section56389529527"></a>**

On Linux systems, the `history` command generally saves the latest 1,000 commands by default. If you need to modify the number of saved commands, for example, to keep only 200 historical commands, you can modify the `HISTSIZE` environment variable in the `"/etc/profile"` file. The modification method is as follows:

- Use an editor (such as the vim editor) to modify.
- Use `sed` to modify directly:

    `sed -i 's/^HISTSIZE=_number/HISTSIZE=_newNumber_/' /etc/profile`, where *number* represents the number of commands before modification, and *newNumber* represents the number of commands after modification. For example, changing the number of saved commands from 1,000 to 200:

    ```shell
    sed -i 's/^HISTSIZE=1000/HISTSIZE=200/' /etc/profile
    ```

After the modification is complete, execute `source /etc/profile`* to make the environment variable take effect.

**Modify the Timestamp of the History Command File<a name="section18178420544"></a>**

If you need to have a timestamp record in the history command file (combined with custom information such as user and IP), you can add the following configuration to `"/etc/profile"`:

`export HISTTIMEFORMAT="%F %T $USER\_IP:\`whoami\` "`

After adding, you need to execute the `source /etc/profile` command to make the environment variable take effect. After adding the timestamp, the result of the history command is as follows:

```shell
2025-12-02 20:44:34 xxx.xxx.xxx.xxx:root systemctl start container-manager.service
2025-12-02 20:44:34 xxx.xxx.xxx.xxx:root systemctl restart container-manager.timer
2025-12-02 20:44:34 xxx.xxx.xxx.xxx:root systemctl restart container-manager.service
2025-12-02 20:44:34 xxx.xxx.xxx.xxx:root systemctl status container-manager.service
2025-12-02 20:44:34 xxx.xxx.xxx.xxx:root systemctl status container-manager.service
```

In addition, if you need to record history commands in a custom file, you can set the `HISTFILE` environment variable in `"/etc/profile"`. After setting, execute the `source /etc/profile` command to make the environment variable take effect. For example:

```shell
HISTDIR=~/log/container-manager   # Configure the history command record save file
HISTFILE="$HISTDIR/container-manager.log"
mkdir -p $HISTDIR
chmod 750 $HISTDIR
touch $HISTFILE
chmod 640 $HISTFILE
USER_IP=`who -u am i 2>/dev/null| awk '{print $NF}'|sed -e 's/[()]//g'`
if [ -z $USER_IP ]
then
  USER_IP=`hostname`
fi
export HISTTIMEFORMAT="%F %T $USER_IP:`whoami` "    # History command display format: time, IP, username, executed command
PROMPT_COMMAND=' { date "+%Y-%m-%d %T - $(history 1 | { read x cmd; echo "$cmd"; })"; } >> $HISTFILE'    # Write the history command to the configuration file in real time
```

The log file path is `"\~/log/container-manager"`. Ensure sufficient drive space and set the log file permissions to `640`.
