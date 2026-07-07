# API Description

## Version Information

The tool version information can be viewed using the `about` command.

## Default Configuration

When the following paths are not manually set, the tool automatically reads the default files or directories under the execution path:

- Connection configuration: `conn.ini`
- BMC log directory: `bmc_dump_log`
- Host log directory: `host_dump_log`
- Switch log directory: `switch_dump_log`

## API Call Flow

**Online Diagnosis Workflow**

1. Use `set_conn_config` to set the device connection configuration.
2. Use `auto_collect_diag` to start one-click diagnosis.
3. After the diagnosis is complete, use `clear_cache` to clear the cache.

**Offline Diagnosis Workflow**

1. Use `set_host_dump_log` to set the server log directory.
2. Use `set_bmc_dump_log` to set the BMC log directory.
3. Use `set_switch_dump_log` to set the switch log directory.
4. Use `auto_collect_diag` to start one-click diagnosis.
5. After the diagnosis is complete, use `clear_cache` to clear the cache.

**Batch Diagnosis Workflow**

1. Use configuration commands to set partial device information.
2. Use `auto_collect` to collect device information.
3. Repeat steps 1 and 2 to set and collect other device information.
4. Use `auto_diag` to start unified diagnosis.
5. After the diagnosis is complete, use `clear_cache` to clear the cache.

## Basic Commands

### help

**Function**

Displays help information.

**Format**

| Format | Description |
|---------|------|
| `help` | Displays help information for all available commands. |
| `help ?` | Views details.  |

### exit

**Function**

Exits the link diagnostic tool.

**Format**

| Format | Description |
|---------|------|
| `exit` | Exits the link diagnosis tool. |
| `exit ?` | Views details. |

### clear

**Function**

Clears the terminal screen.

**Format**

| Format | Description |
|---------|------|
| `clear` | Clears the terminal screen. |
| `clear ?` | Views details.  |

### about

**Function**

Views information about Ascend-faultdiag-toolkit.

**Format**

| Format | Description |
|---------|------|
| `about` | Displays the version and contact information of Ascend-faultdiag-toolkit. |
| `about ?` | Views details.       |

### guide

**Function**

Displays the usage guide of Ascend-faultdiag-toolkit.

**Format**

| Format | Description |
|---------|------|
| `guide` | Displays the usage guide of Ascend-faultdiag-toolkit. |
| `guide ?` | Views details.        |

## Configuration Commands

### set_conn_config

**Function**

Sets device connection configuration.

**Format**

| Format | Description |
|---------|------|
| `set_conn_config <file path>` | Sets the device connection configuration file. |
| `set_conn_config ?` | Views details of the configuration. |

**Parameter Description**

|Parameter|Description|
|---|---|
|`<file path>`|Path to the connection configuration file|

**Configuration File Structure**

```ini
[host]
# port specifies the port, defaults to 22 if not specified; username specifies the username; password specifies the password; private_key specifies the private key file
1.1.1.1 port="22" username="root" private_key="~/.ssh/your_private_key"
1.1.2.1 port="22" username="root" password="321"

[bmc]
1.1.1.2 username="Administrator" password="123"

[switch]
# Support IP range format ip1-ip2 (requires the same username/password), step sets the step size
1.1.1.3-1.1.1.10 step=1 username="root" password="123"

[config]
# Support setting a global private key file
private_key="~/.ssh/your_private_key"
```

### set_host_dump_log

**Function**

Sets the server dump log directory.

**Format**

| Format | Description |
|---------|------|
| `set_host_dump_log <directory>` | Sets the server dump log directory. |
| `set_host_dump_log ?` | Views details. |

**Parameter Description**

|Parameter|Description|
|---|---|
|`<directory>`|Server log directory path|

**Supported Log Types**

- Logs collected by `A3device_log_one-click_collection_script<version>.sh`
- Logs collected by `link_down_collect_<version>.sh`
- Logs collected by `tool_log_collection_out_version_all_<version>.sh`

### set_bmc_dump_log

**Function**

Sets the BMC log directory.

**Format**

| Format | Description |
|---------|------|
| `set_bmc_dump_log <directory>` | Sets the BMC log directory. |
| `set_bmc_dump_log ?` | Views details. |

**Parameter Description**

|Parameter|Description|
|---|---|
|`<directory>`|BMC log directory path|

**Supported Log Types**

- Logs manually downloaded via the "One-Click Collection" button on the BMC WebUI.
- Logs collected using the `ipmcget -d diaginfo` command.

### set_switch_dump_log

**Function**

Sets the directory for switch command output directory.

**Format**

| Format | Description |
|---------|------|
| `set_switch_dump_log <directory>` | Sets the directory for switch command output. |
| `set_switch_dump_log ?` | Views details. |

**Parameter Description**

|Parameter|Description|
|---|---|
|`<directory>`|Switch log directory path|

**Supported Log Types**

- Results exported using the `display diagnostic-information <filename>` command or shell text copied after querying key commands.
- Log zip packages exported using the `collect diagnostic-information` command.

## Collection Commands

### collect_bmc_dump_info

**Function**

Collects BMC dump info logs online.

**Format**

| Format | Description |
|---------|------|
| `collect_bmc_dump_info` | Collects BMC dump info logs online. |
| `collect_bmc_dump_info ?` | Views details.        |

**Output Description**

After collection is complete, the logs are located in the `CommonPath.TOOL_HOME_BMC_DUMP_CACHE_DIR` directory.

### auto_collect

**Function**

Starts automatic information collection.

- Supports offline and online collection.
- Suitable for batch collection across different network planes.

**Format**

| Format | Description |
|---------|------|
| `auto_collect` | Starts automatic information collection. |
| `auto_collect ?` | Views details. |

## Diagnostic Commands

### auto_inspection

**Function**

Starts inspection result diagnosis.

**Format**

| Format | Description |
|---------|------|
| `auto_inspection` | Starts diagnosis using the default customer type. |
| `auto_inspection <customer type>` | Starts diagnosis using the specified customer type. |
| `auto_inspection ?` | Views supported customer types. |

**Parameter Description**

|Parameter|Description|
|---|---|
|`<customer type>`|Enumeration value of supported customer types. Currently supports `default`.|

### auto_diag

**Function**

Starts automatic diagnosis, suitable for unified diagnosis after batch collection.

**Format**

| Format | Description |
|---------|------|
| `auto_diag` | Starts automatic diagnosis. |
| `auto_diag ?` | Views details. |

### auto_collect_diag

**Function**

Starts one-click automatic collection and diagnosis and automatically executes the collection (online device collection or offline log collection) and diagnosis process.

**Format**

| Format | Description |
|---------|------|
| `auto_collect_diag` | Starts one-click automatic collection and diagnosis. |
| `auto_collect_diag ?` | Views details.        |

## Maintenance Commands

### clear_cache

**Function**

Clears the cache.

- Clears cache files generated during tool execution.
- It is recommended to execute this command before running a new diagnostic task.
- If the cleanup does not take effect, open the tool in administrator mode.

**Format**

| Format | Description |
|---------|------|
| `clear_cache` | Clears the Ascend-faultdiag-toolkit cache. |
| `clear_cache ?` | Views details.        |
