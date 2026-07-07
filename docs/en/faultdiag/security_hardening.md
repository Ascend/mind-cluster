# Security Hardening<a name="ZH-CN_TOPIC_0000001595078762"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-04T12:35:37.104Z pushedAt=2026-06-05T01:40:45.122Z -->

## Overview<a name="ZH-CN_TOPIC_0000001632128594"></a>

The security hardening measures listed in this document are basic hardening recommendations. You should re-evaluate the overall network security hardening measures of the system based on your own business needs.

Perform relevant configurations according to the security policies of their organization, including but not limited to the following.

- Software version
- Permission configuration
- Firewall settings

When necessary, refer to industry-leading security hardening solutions and recommendations from security experts. You should follow the official recommendations of the operating system and software you are using for related hardening.

## Operating System Security Hardening<a name="ZH-CN_TOPIC_0000001680248561"></a>

- Apply security patches in a timely manner according to your organization's security policy and use software versions approved by your organization.
- Delete or disable unnecessary system accounts to reduce security risks.
- Check for accounts with empty passwords.
- Strengthen password complexity to reduce the possibility of being guessed.
- Restrict users from using the `su` command.

## Firewall Configuration<a name="ZH-CN_TOPIC_0000001631968694"></a>

- After an operating system is installed, if a regular user is configured, add the `ALWAYS_SET_PATH` field in the `/etc/login.defs` file and set it to `yes` to prevent unauthorized operations.
- To prevent regular users from inheriting environment variables through `su root` and thereby escalating privileges, set `ALWAYS_SET_PATH` to `yes` in the server configuration file `/etc/pam.d/su`.
- For other operations, refer to the relevant guidance for the OS you are using.

## Setting `umask`<a name="ZH-CN_TOPIC_0000001680408433"></a>

It is recommended that you set `umask` to 027 or higher on hosts (including physical machines) and in containers to improve security.

Take setting `umask = 027` as an example:

1. Log in to the server as the root user and edit the `/etc/profile` file.

    ```shell
    vim /etc/profile
    ```

2. Add `umask 027` to the end of the `/etc/profile` file, then execute `:wq` to save and exit.
3. Execute the following command to make the configuration take effect.

    ```shell
    source /etc/profile
    ```

## SSH Security Hardening<a name="ZH-CN_TOPIC_0000001680128293"></a>

- Enhance the security of SSH connections by modifying configuration files in the `/etc/ssh/` path or the `~/.ssh` path, such as `ssh_config` and `sshd_config`. After making changes, the SSH service must be restarted or reloaded, for example by executing the `systemctl restart sshd` (or `service sshd restart`) command, for the configuration to take effect. It is especially recommended to disable the SSH v1 protocol and encryption components using insecure communication protocols.
- Be aware that enabling root login poses security risks. For detailed information, refer to the relevant documentation for the operating system in use.
- Perform SSH authentication login using public-private key pairs. When using this method, ensure that the algorithm and key length meet the security requirements of your organization. A reference is that the key length under the RSA algorithm should not be less than 3,072 bits. Additionally, do not set a private key with an empty password, as this introduces security risks.
- The length and complexity of the private key password should meet the security requirements of your organization.

## Ownerless File Security Hardening<a name="ZH-CN_TOPIC_0000001676284886"></a>

Because official Docker images differ from the operating system on the physical machine, users in the system may not have a one-to-one correspondence, causing files generated during physical machine or container operation to become files without an owner.

You can run the `find / -nouser -o -nogroup` command to search for files without an owner in the container or on the physical machine. Create corresponding users and user groups based on the file's UID and GID, or modify the UID of an existing user or the GID of a user group to adapt, assign file ownership, and prevent files without an owner from posing security risks to the system.
