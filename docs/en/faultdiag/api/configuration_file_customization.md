# Custom Configuration File Interface<a name="ZH-CN_TOPIC_0000002447169561"></a>

## Prototype<a name="zh-cn_topic_0000001511538701_section124882040143613"></a>

```shell
ascend-fd config subcommand
```

## Function<a name="zh-cn_topic_0000001511538701_section12230185113815"></a>

Provides a custom configuration file, including adding or modifying and querying custom configuration, and verifying the custom fault entity file `custom-fd-config.json`.

## Parameters<a name="zh-cn_topic_0000001511538701_section122149111390"></a>

**Table 1**  Subcommand parameters

|Parameter|Abbreviation|Required|Value Type|Description|
|--|--|--|--|--|
|`--update`|`-u`|Required. `--update`, `--show`, and `--check` are mutually exclusive. You must specify exactly one of them.|String|Adds or modifies custom configuration in a JSON file. For details about the parameters in the JSON file, see [Table 1 Parameters](../user_guide/13_customizing_a_configuration_file.md).|
|`--show`|`-s`|Required. `--update`, `--show`, and `--check` are mutually exclusive. You must specify exactly one of them.|Bool|Queries user-defined configuration.|
|`--check`|`-c`|Required. `--update`, `--show`, and `--check` are mutually exclusive. You must specify exactly one of them.|Bool|Checks the validity of the `custom-fd-config.json` file, primarily verifying the validity of the field attributes of each custom configuration.|
|`--help`|`-h`|Optional|-|Queries usage instructions.|

## Returns<a name="zh-cn_topic_0000001511538701_section1714345618323"></a>

Example: Add a custom configuration file entity through a JSON file.

```ColdFusion
ascend-fd config -u custom-config.json
Updated entity successfully.
```
