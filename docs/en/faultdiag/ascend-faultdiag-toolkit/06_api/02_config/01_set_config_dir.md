# set_config_dir

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:18:22.094Z pushedAt=2026-08-24T02:22:41.582Z -->

## Command Function

Sets the configuration file directory path. The tool automatically scans and loads the configuration files in this directory. Currently, it supports loading the equipment room configuration file `LLD.xlsx` (which contains the UnifiedBus L1/L2 network mapping relationships), used to associate location information such as the cabinets and equipment room in diagnostic reports.

## Command Format

| Command Format | Description |
|---------|------|
| `set_config_dir <directory_Path>` | Sets the configuration file directory path |
| `set_config_dir ?` | Views details |

## Parameter Description

| Name | Type | Mandatory | Description |
|------|-----|------|------|
| `<directory_path>` | String | Yes | Directory path where the configuration file resides. The directory must contain `LLD.xlsx`. |

## LLD.xlsx File Structure

You need to provide the `LLD.xlsx` file. For a sample file, see [LLD.xlsx](../../../../resource/LLD.xlsx). The file must contain two sheets:

| Sheet Name | Mandatory Fields | Purpose |
|----------|--------|------|
| UnifiedBus L1 Network Mapping | Server, Equipment Room Name, Cabinet Number, Host SN, L1 Name, L1_IP, L1_SN | Describes the mapping between hosts and L1 switches |
| UnifiedBus L2 Network Mapping | Device Name, Equipment Room Name, Cabinet Number, Management IP Configuration, SN | Describes the machine room location information of L2 switches |

>[!NOTE]
> `set_config_dir` is an optional configuration command. The tool still works when it is not set, but it will not perform correlation analysis on the equipment room location dimension.

## Output Description

- On success, the command returns `Configuration succeeded. Configuration directory: {dir_path}`.
- On failure, the command returns `The directory path is empty. Set it again.`,`The directory {dir_path} does not exist. Set it again.`, or `The path {dir_path} is not a directory. Set it again.`.

## Examples

Non-interactive mode (command and output):

```bash
ascend-fd-tk set_config_dir /home/user/config set_conn_config /home/user/conn.ini auto_collect_diag
Setting successful. Configuration directory: /home/user/config
# Other log output...
```

Interactive mode (command and output):

```bash
ascend-fd-tk
>>> set_config_dir /home/user/config
Setting successful. Configuration directory: /home/user/config
```
