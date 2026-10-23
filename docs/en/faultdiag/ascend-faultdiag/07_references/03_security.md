# Security Hardening

## File Permissions

During the installation and use of ascend-fd, it is recommended to set file permissions as follows.

| Directory/File                    | Recommended Permission | Description                              |
|-----------------------------------|------------------------|------------------------------------------|
| `~/.ascend_faultdiag`             | 700                    | Owner only can read, write, and execute  |
| `~/.ascend_faultdiag/*.log`       | 600                    | Owner only can read and write            |
| `~/.ascend_faultdiag/config.json` | 600                    | Configuration file; owner only can read and write |

Set permissions:

```shell
chmod 700 ~/.ascend_faultdiag
chmod 600 ~/.ascend_faultdiag/*.log
chmod 600 ~/.ascend_faultdiag/config.json
```

## System Security Configuration

It is recommended to set `umask` to `027` to improve security.

```shell
vim /etc/profile
# Add umask 027 at the end of the file.
source /etc/profile
```

## Log Security

- Do not expose log files to unauthorized users.
- Operation logs record user actions. Keep them properly stored.

## User Permissions

- It is recommended to use the same user for both installation and usage.
- If `root` must be used for installation and a regular user for usage, ensure that the regular user has the correct `PATH` configuration and file permissions.
