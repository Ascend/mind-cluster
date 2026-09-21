# Manual Upgrade

Manual upgrade supports two methods:

- **Full Upgrade**: Upgrades the binary image files of each component and supports modifying configuration files. Cross-version upgrades are supported (e.g., upgrading from 5.0.x to 7.0.x), but training/inference tasks must be stopped.
- **Image Upgrade**: Only upgrades the binary files of each component. Does not support modifying permissions or startup parameters. Only supports upgrades within the same version, with no need to stop tasks.

Before upgrading, please check whether there are any running tasks, and disable pingmesh UnifiedBus network detection as appropriate. Upgrade steps include environment checks, uninstalling the old version, installing the new version, restoring configurations, and others.

For details, see [Developer Guide - Manual Upgrade](../../05_developer_guide/00_installation_deployment/01_upgrade.md).
