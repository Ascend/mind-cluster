# API Overview

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:28:29.048Z pushedAt=2026-08-24T02:51:43.274Z -->

ascend-fd provides two types of interfaces:

- **Command-line interface**: used through the `ascend-fd` command, suitable for O&M personnel and general users.
- **SDK interface**: imported and used by Python programs as a third-party library, suitable for developers to integrate into their own systems.

## Command-Line Interface

### Command Format

```shell
ascend-fd <subcommand> [parameters]
```

### Subcommand List

| Subcommand                                 | Function               | Description                                        |
|--------------------------------------------|------------------------|----------------------------------------------------|
| [parse](./02_command_parse.md)             | Log parsing   | Extracts key information from raw logs             |
| [diag](./03_command_diag.md)               | Fault diagnosis        | Analyzes the root cause of single-server or multi-server faults |
| [single-diag](./04_command_single_diag.md) | Single-server fault diagnosis | Quickly diagnoses on a single server           |
| [entity](./05_command_entity.md)           | Custom fault entity    | Manages custom fault detection rules               |
| [blacklist](./06_command_blacklist.md)     | Fault log masking      | Manages log masking rules                          |
| [config](./07_command_config.md)           | View configuration file | Manages the path of the configuration file in use |
| [version](./08_command_version.md)         | View version           | Displays the current ascend-fd version             |

> [!NOTE]
>
> For parameter details, refer to the specific sections of each subcommand.

## SDK Interface

ascend-fd provides a Python SDK for easy integration into automated workflows.

### SDK Interfaces

| Interface                                                           | Description         |
|----------------------------------------------------------------|---------------------|
| [parse_fault_type](./09_sdk_api.md#parse_fault_type)           | Service log parsing |
| [parse_root_cluster](./09_sdk_api.md#parse_root_cluster)       | Root cause node parsing |
| [diag_root_cluster](./09_sdk_api.md#diag_root_cluster)         | Root cause node diagnosis |
| [parse_knowledge_graph](./09_sdk_api.md#parse_knowledge_graph) | Fault event parsing |
| [diag_knowledge_graph](./09_sdk_api.md#diag_knowledge_graph)   | Fault event diagnosis |
