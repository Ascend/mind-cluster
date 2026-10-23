# API Overview

This document summarizes all commands provided by ascend-fd-tk and their functional categories for quick reference. For detailed descriptions of each command, see the corresponding sections.

## Command Categories

### Basic Commands

| Command | Brief Description | Parameter Required | Detailed Description |
|------|----------|--------------|----------|
| `help` | Displays help information for all available commands | No | [help](01_basic/01_help.md) |
| `exit` | Exits the program | No | [exit](01_basic/02_exit.md) |
| `clear` | Clears the screen | No | [clear](01_basic/03_clear.md) |
| `about` | Views tool version information | No | [about](01_basic/04_about.md) |
| `guide` | Obtain usage guide information | No | [guide](01_basic/05_guide.md) |

### Configuration Commands

| Command | Brief Description | Parameter Required | Detailed Description |
|------|----------|--------------|----------|
| `set_config_dir` | Sets the configuration file directory (scans `LLD.xlsx`) | Yes (Directory Path) | [set_config_dir](02_config/01_set_config_dir.md) |
| `set_conn_config` | Sets the device connection configuration (host/BMC/switch) | Yes (Configuration File Path) | [set_conn_config](02_config/02_set_conn_config.md) |
| `set_host_dump_log` | Sets the server export log directory (offline) | Yes (Directory Path) | [set_host_dump_log](02_config/03_set_host_dump_log.md) |
| `set_bmc_dump_log` | Sets the BMC log directory (offline) | Yes (Directory Path) | [set_bmc_dump_log](02_config/04_set_bmc_dump_log.md) |
| `set_switch_dump_log` | Sets the switch command output text directory (offline) | Yes (Directory Path) | [set_switch_dump_log](02_config/05_set_switch_dump_log.md) |

### Collection Commands

| Command | Brief Description | Parameter Required | Detailed Description |
|------|----------|--------------|----------|
| `collect_bmc_dump_info` | Collect BMC dump info logs online | No | [collect_bmc_dump_info](03_collect/collect_bmc_dump_info.md) |

### Parsing and Diagnostics Commands

| Command | Brief Description | Parameter Required | Detailed Description |
|------|----------|--------------|----------|
| `auto_collect` | Starts automatic information collection, supporting offline and online collection, and is suitable for batch collection across different network planes | No | [auto_collect](04_parse_diagnosis/01_auto_collect.md) |
| `auto_diag` | Starts automatic diagnostics (used together with batch collection) | No | [auto_diag](04_parse_diagnosis/02_auto_diag.md) |
| `auto_collect_diag` | One-click automatic collection + diagnostics | No | [auto_collect_diag](04_parse_diagnosis/03_auto_collect_diag.md) |
| `auto_single_diag` | Diagnoses a single link (device IP + port) | Yes (IP, port) | [auto_single_diag](04_parse_diagnosis/04_auto_single_diag.md) |

### Inspection Commands

| Command | Brief Description | Parameter Required | Detailed Description |
|------|----------|--------------|----------|
| `auto_inspection` | Starts inspection, applicable to customer-customized inspection scenarios | Optional (determined by the customer type) | [auto_inspection](05_inspection/auto_inspection.md) |

### Maintenance Commands

| Command | Brief Description | Parameter Required | Detailed Description |
|------|----------|--------------|----------|
| `clear_cache` | Clears the tool runtime cache | No | [clear_cache](06_maintenance/clear_cache.md) |

## Usage Guide

- For first-time use or to view capabilities: run `guide` → `help` in sequence.
- To view the detailed description and usage of a single command: append `?` to the command, for example, `set_conn_config ?`.
