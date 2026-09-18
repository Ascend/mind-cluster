# set_conn_config

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:18:31.328Z pushedAt=2026-08-24T02:22:41.584Z -->

## Command Function

Sets the path of the device connection configuration file for online analysis scenarios. The tool reads the host, BMC, and switch connection information configured in the file and stores it in encrypted form. After the configuration is set successfully, it is recommended to delete the source file containing plaintext passwords as soon as possible.

## Command Format

| Command Format | Description |
|---------|-----------|
| `set_conn_config <File_Path>` | Sets the connection file path. |
| `set_conn_config ?` | Views details. |

## Parameter Description

| Name | Type | Mandatory | Description |
|------|-----|------|------|
| `<File_Path>` | String | Yes | Path of the connection file. |

## Configuration File Structure

```conn.ini
[host]
# port specifies the port; default is 22 if not specified. username specifies the username, password specifies the password, private_key specifies the private key file
1.1.1.1 port="22" username="root" private_key="~/.ssh/your_private_key"
1.1.1.2 port="22" username="root" password="<your_password>"

[bmc]
1.1.1.3 username="Administrator" password="<your_password>"

[switch]
# Supports IP range format ip1-ip2 (ensuring identical username and password), with step setting the increment
1.1.1.4-1.1.1.10 step=2 username="root" password="<your_password>"

[config]
# Supports setting a global private key file
private_key="~/.ssh/your_private_key"
```

Configuration notes for cluster device connection:

- When using the tool, you need to connect to each device to perform operations such as command query and log collection. Configure the cluster connection information by referring to the `conn.ini` file format above.
- If password-free connection is supported, you can leave the password blank.
- Key-based login is supported. `private_key` specifies the private key path, which can be configured separately on each line or configured for all environments in the cluster in the `config` section.
- The `ip` field supports both a single IP address and an IP range configuration (configured via `step`: `step` defaults to `1`; ensure that the username and password are the same).
- Devices such as the 1620 front-end switch are also treated as switches; enter them under the `switch` section.

[!NOTICE]
> `conn.ini` contains device login credentials and is considered sensitive information. Recommendations:
>
> - After configuration is complete, promptly delete the configuration file that contains plaintext passwords (the tool prompts you after successful setup).
> - Prefer using private key passwordless login to avoid storing plaintext passwords in the configuration file.
> - Restrict access permissions on the configuration file (for example, `chmod 600 conn.ini`) so that only the current user can read it.

## Output Description

- On success, the command returns `Configuration succeeded. Delete the configuration file containing the plaintext password as soon as possible.`
- On failure, the command returns `Failed to set the path. Error: {err}`, `The path is empty. Set it again.`, `The path {file_path} does not exist. Set it again.`, or `The path {file_path} is not a file. Set it again.`

## Examples

Non-interactive mode (command and output):

```bash
ascend-fd-tk set_conn_config /home/user/conn.ini auto_collect_diag
Configuration succeeded. Delete the configuration file containing the plaintext password as soon as possible.
# Other log output...
```

Interactive mode (command and output):

```bash
ascend-fd-tk
>>> set_conn_config /home/user/conn.ini
Configuration succeeded. Delete the configuration file containing the plaintext password as soon as possible.
```
