# Security Management and Hardening

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T06:28:18.200Z pushedAt=2026-06-09T07:15:15.740Z -->

## Security Management

> [!NOTE]
> MindIO ACP currently does not support public cloud scenarios, multi-tenant scenarios, or direct system access from the public network.

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

2. Add `umask 027` at the end of the `"/etc/profile"` file, save and exit.
3. Run the following command to make the configuration take effect.

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

1. Log in to the node where the MindIO ACP component is installed.
2. Open the `"/etc/ssh/sshd_config"` file.

    ```bash
    vim /etc/ssh/sshd_config
    ```

3. Press `i` to enter insert mode, locate the `PermitRootLogin` configuration item, and set its value to `no`.

    ```text
    PermitRootLogin no
    ```

4. Press `Esc`, type `:wq!`, and press `Enter` to save and exit.
5. Run the command to apply the configuration.

    ```bash
    systemctl restart sshd
    ```
