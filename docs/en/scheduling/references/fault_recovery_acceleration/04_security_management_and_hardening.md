# Security Management and Hardening

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T06:23:50.124Z pushedAt=2026-06-09T07:15:15.700Z -->

## Security Management

> [!NOTE]
> MindIO TFT currently does not support public cloud scenarios, multi-tenant scenarios, or direct access to the system over the public network.

**Routine Antivirus Software Check**

Regular antivirus scans are performed on the cluster. Routine antivirus checks help protect the cluster from viruses, malicious code, spyware, and malicious programs, reducing risks such as system paralysis and information security issues. Mainstream antivirus software in the industry can be used for antivirus checks.

**Log Management**

Log management requires attention to the following two points.

- Check whether the system can limit the size of a single log file.
- Check whether there is a mechanism to clean up logs after the log space is full.

**Vulnerability/Functional Issue Remediation**

To ensure the security of the production environment and reduce the risk of attacks, regularly review the following vulnerabilities/functional issues fixed by the open-source community.

- Operating system vulnerabilities/functional issues
- Vulnerabilities/functional issues in other related components

## Security Hardening

### Hardening Notes

The security hardening measures listed in this document are basic hardening recommendations. You should re-evaluate the network security hardening measures for the entire system based on your own business needs, and may refer to industry best hardening practices and security expert recommendations when necessary.

### Risk Warning

Checkpoint serialization uses the pickle component built into Python. It is essential to ensure that unauthorized users do not have write permissions on the storage directory and its parent directories; otherwise, it may lead to the risk of checkpoint tampering, which can cause pickle deserialization injection.

### Operating System Security Hardening

**Firewall Configuration**

After the operating system is installed, if a regular user is configured, you can add the `"ALWAYS_SET_PATH=yes"` configuration in the `"/etc/login.defs"` file to prevent unauthorized operations. In addition, to prevent privilege escalation caused by bringing the current user's environment variables into other environments when using the `su` command to switch users, use the `su - [user]` command for user switching, and add the configuration parameter `"ALWAYS_SET_PATH=yes` in the server configuration file `"/etc/default/su"` to prevent privilege escalation.

**Setting umask**

It is recommended that users set the server's `umask` to 027 ~ 777 to restrict file permissions.

Taking setting `umask` to `027` as an example, the specific operation is as follows.

1. Log in to the server as the root user and edit the `"/etc/profile"` file.

    ```bash
    vim /etc/profile
    ```

2. Add `umask 027` at the end of the `"/etc/profile`" file, save and exit.
3. Run the following command to apply the configuration.

    ```bash
    source /etc/profile
    ```

**Security Hardening for Files Without an Owner**

Run the `find / -nouser -nogroup` command to search for files without an owner in the container or on the physical machine. Create corresponding users and user groups based on the UID and GID of the files, or modify the UID of an existing user or the GID of a user group to adapt, and assign an owner to the files to prevent security risks caused by files without an owner.

**Port Scanning**

Pay attention to ports that are listening on the entire network and unnecessary ports. If unnecessary ports are found, they should be closed immediately. It is recommended that you disable insecure services, such as Telnet and FTP, to enhance system security. For specific operation methods, refer to the official documentation of the operating system in use.

**Anti-DoS Attack**

Protect the system against DoS attacks by limiting the rate of connections to the server based on IP addresses. Methods include but are not limited to using the built-in Iptables firewall of the Linux system for prevention and optimizing `sysctl` parameters. For specific usage methods, consult relevant materials.

**SSH Hardening**

Since the `root` user has the highest privileges, for security purposes, it is recommended to disable the `root` user's SSH remote login permission to the server to enhance system security. The specific operation steps are as follows:

1. Log in to the node where the MindIO TFT component is installed.
2. Open the `"/etc/ssh/sshd_config"` file.

    ```bash
    vim /etc/ssh/sshd_config
    ```

3. Press `i` to enter insert mode, locate the `PermitRootLogin` configuration item and set its value to `no`.

    ```text
    PermitRootLogin no
    ```

4. Press `Esc`, type `:wq!`, and press `Enter` to save and exit.
5. Run the command to make the configuration take effect.

    ```bash
    systemctl restart sshd
    ```

**Buffer Overflow Protection**

To prevent buffer overflow attacks, it is recommended to use ASLR (Address Space Layout Randomization). By randomizing the layout of linear regions such as the heap, stack, and shared library mappings, ASLR increases the difficulty for attackers to predict target addresses, thereby preventing them from directly locating attack code. This technique can be applied to the heap, stack, and memory mapping regions (mmap base address, shared libraries, vdso page).

How to enable:

```bash
echo 2 >/proc/sys/kernel/randomize_va_space
```

## Enabling TLS Authentication

- To secure the communication between the Controller and Processor within MindIO TFT and protect information from tampering and impersonation, it is recommended to enable TLS encryption.
- TLS encryption is used only for inter-module communication within MindIO TFT, and does not provide external TLS access or authentication functions.
- Because enabling security authentication depends on the OpenSSL component, you are advised to use a vulnerability-free version of OpenSSL, which requires the use of GLIBC 2.33 or a later version.

### Importing TLS Certificates

- Configure TLS key certificates and other settings through the `tft_start_controller` and `tft_init_processor` interfaces to establish a TLS secure connection. Security options are enabled by default, and it is recommended that users enable TLS encryption to ensure communication security. To disable encryption, use the example below to call the interface and disable it.
- After the system starts, it is recommended to delete sensitive information files such as local key certificates.
- During interface call, the file path passed in should avoid containing English semicolons, commas, or colons.
- The certificate check period and certificate expiration warning time can be configured using the environment variables `TTP_ACCLINK_CHECK_PERIOD_HOURS` and `TTP_ACCLINK_CERT_CHECK_AHEAD_DAYS`.

**Example of TLS Interface Call**

- When TLS is disabled (`enable_tls=False`), `tls_info` is invalid and does not need to be configured. This switch does not affect the functional features of MindIO TFT.

    ```python
    from mindio_ttp.framework_ttp import tft_start_controller, tft_init_processor

    tft_start_controller(bind_ip: str, port: int, enable_tls=False, tls_info='')
    tft_init_processor(rank: int, world_size: int, enable_local_copy: bool, enable_tls=False, tls_info='', enable_uce=True, enable_arf=False)
    ```

    > [!CAUTION]
    > - If TLS is disabled (i.e., when `enable_tls=False`), there will be a high network security risk.
    > - The enable_tls switch status of `tft_start_controller` and `tft_init_processor` must remain consistent. If the `enable_tls` switches of the two interfaces differ, the following issues will occur:
    >   - TLS link establishment between modules fails.
    >   - MindIO TFT cannot run normally, and the training job fails to start.

- When TLS is enabled (`enable_tls=True`), certificate-related information is used as the required parameter `tls_info` for the following interfaces:

    ```python
    from mindio_ttp.framework_ttp import tft_start_controller, tft_init_processor, tft_register_decrypt_handler

    # In tls_info, use ";" to separate different fields and "," to separate individual fileserent fields
    tls_info = r"(
    tlsCert: /etc/ssl/certs/cert.pem;
    tlsCrlPath: /etc/ssl/crl/;
    tlsCaPath: /etc/ssl/ca/;
    tlsCaFile: ca_cert_1.pem, ca_cert_2.pem;
    tlsCrlFile: crl_1.pem, crl_2.pem;
    tlsPk: private key;
    tlsPkPwd: private key pwd;
    packagePath: /etc/ssl/
    )"

    # If the tlsPkPwd password is cipher text, register a password decryption function
    tft_register_decrypt_handler(user_decrypt_callback)
    tft_start_controller(bind_ip: str, port: int, enable_tls=True, tls_info=tls_info)
    tft_init_processor(rank: int, world_size: int, enable_local_copy: bool, enable_tls=True, tls_info=tls_info, enable_uce=True, enable_arf=False)
    ```

**Meaning of each field in tls_info**

| Field | Meaning | Required |
|--|--|--|
| tlsCert | Server certificate. | Yes |
| tlsCaPath | CA certificate storage path. | Yes |
| tlsCaFile | CA certificate list. | Yes |
| tlsCrlPath | Certificate revocation list storage path. | No |
| tlsCrlFile | Certificate revocation list. | No |
| tlsPk | Private key. | Yes |
| tlsPkPwd | Private key password. | Yes |
| packagePath | OpenSSL library path. | Yes |

> [!CAUTION]
> Certificate security requirements:
>
> - Use industry-recognized, secure, and trusted asymmetric encryption algorithms, key exchange algorithms, key lengths, hash algorithms, certificate formats, etc.
> - Certificates must be within their validity period.

### (Optional) Checking Certificate Validity

If TLS authentication is enabled, the certificate validity period must be monitored. Properly plan the certificate validity period and renewal cycle, and update certificates before they expire to mitigate security risks. MindIO TFT provides a periodic certificate validity check feature, with a default check interval of 7 days and a default advance warning time of 30 days. If a certificate is at risk of expiring, a WARNING message will be printed to the log configured via the `TTP_LOG_PATH` environment variable. Please pay attention to these warnings and take action promptly.
