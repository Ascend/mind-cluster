# Custom Fault Entity Interface<a name="ZH-CN_TOPIC_0000001849651633"></a>

## Prototype<a name="zh-cn_topic_0000001511538701_section124882040143613"></a>

```shell
ascend-fd entity Subcommand
```

## Function<a name="zh-cn_topic_0000001511538701_section12230185113815"></a>

Provides functions related to custom fault entities, including adding, modifying, querying, or deleting custom fault entities, and validating the custom fault entity file `custom-ascend-kg-config.json`.

## Parameters<a name="zh-cn_topic_0000001511538701_section122149111390"></a>

**Table 1** Subcommand parameters

|Parameter|Abbreviation|Required|Value Type|Description|
|--|--|--|--|--|
|`--update`|`-u`|Required. `--update`, `--delete`, `--show`, and `--check` are mutually exclusive, meaning only one can and must be specified.|String|Adds or modifies custom fault entity information in JSON file format. For details about the parameters in the JSON file, see [Table 1 Parameters](../user_guide/04_customizing_fault_entities.md).|
|`--delete`|`-d`|Required. `--update`, `--delete`, `--show`, and `--check` are mutually exclusive, meaning only one can and must be specified.|String|Deletes the custom fault entity information corresponding to the specified fault codes. Separate multiple fault codes with spaces.|
|`--show`|`-s`|Required. `--update`, `--delete`, `--show`, and `--check` are mutually exclusive, meaning only one can and must be specified.|String|Views custom fault entity information by fault code, with multiple fault codes separated by spaces. If no fault code is specified, all custom fault entity information is queried.|
|`--check`|`-c`|Required. `--update`, `--delete`, `--show`, and `--check` are mutually exclusive, meaning only one can and must be specified.|String|Checks the validity of the `custom-ascend-kg-config.json` file, primarily verifying the validity of the field attributes of each custom fault entity.|
|`--item`|None|Optional|String|Views partial information of a custom fault entity. Separate multiple values with spaces. If no value is specified, the information of the following three values is displayed.<ul><li>`attribute`: attribute information</li><li>`rule`: fault chain</li><li>`regex`: fault keyword</li></ul><p>Must be used with the `--show` (or `-s`) parameter.</p>|
|`--force`|`-f`|Optional|Bool|When this parameter is specified to delete a custom fault entity, no confirmation prompt will appear on the interface.<p>Must be used with the `--delete` (or `-d`) parameter.</p>|
|`--help`|`-h`|Optional|-|Queries usage instructions.|

## Returns<a name="zh-cn_topic_0000001511538701_section1714345618323"></a>

Example: Add a custom fault entity through a JSON file.

```ColdFusion
ascend-fd entity -u test_base.json
Updated entity successfully.
```
