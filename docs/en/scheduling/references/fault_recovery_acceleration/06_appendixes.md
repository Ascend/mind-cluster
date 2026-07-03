# Appendix

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T06:24:36.938Z pushedAt=2026-06-09T07:15:15.708Z -->

## Environment Variables

> [!NOTE]
> Environment variables displayed in bold are commonly used environment variables.

|Variable|Description|Value Range|Default Value|
|--|--|--|--|
|TTP_LOG_PATH|MindIO TFT log path. Configuring soft links is prohibited. The log file name is supplemented as `ttp_log.log`. It is recommended to include the date and time in the log path to avoid multiple training records being written to the same log, causing cyclic overwriting. It is recommended to configure the log path in the training startup script as follows: <br> `date_time=\$(date +%Y-%m-%d-%H_%M_%S)` <br> `export TTP_LOG_PATH=logs/\${date_time}` <br>When using shared storage, it is recommended to configure the log path per node:<br>`export TTP_LOG_PATH=logs/\${nodeId}`|Folder path.|logs|
|TTP_LOG_LEVEL|MindIO TFT log level.<ul><li>DEBUG: Detailed information, applicable only when diagnosing problems.</li><li>INFO: Confirms that the program is running as expected.</li><li>WARNING: Indicates that an unexpected event has occurred or is about to occur. The program continues to run as expected.</li><li>ERROR: Indicates that some functions of the program cannot be executed normally due to a serious problem.</li></ul>|<ul><li>DEBUG</li><li>INFO</li><li>WARNING</li><li>ERROR|INFO</li></ul>|
|TTP_LOG_MODE|MindIO TFT log mode.<ul><li>ONLY_ONE: All MindIO TFT processes write one log.</li><li>PER_PROC: Each MindIO TFT process writes an independent log. The log file path is `{TTP_LOG_PATH}/ttp_log.log.{pid}`.</li></ul>|<ul><li>ONLY_ONE</li><li>PER_PROC (If ONLY_ONE is not specified, the default is PER_PROC)</li></ul>|PER_PROC|
|TTP_LOG_STDOUT|MindIO TFT log recording method.<ul><li>0: Records MindIO TFT runtime logs to the corresponding log file.</li><li>1: Prints MindIO TFT runtime logs directly without local storage.</li></ul>|<ul><li>0</li><li>1</li></ul>|0|
|MASTER_ADDR|IP address or domain name of the training master node.|Valid IPv4, IPv6 address, or domain name.|-|
|MASTER_PORT|Communication port of the training master node. The port is configurable.|[1024, 65535]|-|
|TTP_RETRY_TIMES|Number of Processor TCP (Transmission Control Protocol) connection attempts.|[1, 300]|10|
|MINDIO_WAIT_MINDX_TIME|Maximum time for the Controller to wait for a MindCluster response, in seconds|[1, 3600]|30|
|TTP_ACCLINK_CHECK_PERIOD_HOURS| Period for MindIO TFT to check certificate validity after TLS authentication is enabled, in hours.|[24, 720]|168|
|TTP_ACCLINK_CERT_CHECK_AHEAD_DAYS|Duration for MindIO TFT to issue an early warning before the certificate expiration date after TLS authentication is enabled, in days. The certificate expiration warning duration must not be less than the check period to ensure timely detection and warning of certificate expiration risks.|[7, 180]; satisfy TTP_ACCLINK_CERT_CHECK_AHEAD_DAYS * 24 ≥ TTP_ACCLINK_CHECK_PERIOD_HOURS.|30|
|TTP_NORMAL_ACTION_TIME_LIMIT|Timeout for executing the `rebuild/repair/rollback` callback during the fault recovery process, in seconds.|[30, 1800]|180|
|MINDIO_FOR_MINDSPORE|Whether to enable the MindSpore switch. When `True` (case-insensitive) or `1` is passed, the MindSpore switch is enabled. Other values disable the MindSpore switch.|<ul><li>True (case-insensitive) or 1: Enable MindSpore.</li><li>Other: Disable MindSpore.</li></ul>|False|
|MINDX_TASK_ID|Used by the MindIO ARF feature; MindCluster task ID; configured by ClusterD; no user intervention required.|String.|-|
|TORCHELASTIC_USE_AGENT_STORE|PyTorch environment variable; controls whether to create a TCP Store Server or Client. Used by MindIO TFT in scenarios where a dying gasp checkpoint is saved and the Torch Agent TCP Store Server connection fails.|<ul><li>True: Create Client.</li><li>False: Create Server.</li></ul>|-|
|TTP_STOP_CLEAN_BEFORE_DUMP|Used by the MindIO TFT feature; controls whether MindIO TTP performs a stop&clean operation before saving the dying gasp checkpoint.|<ul><li>0: Disable.</li><li>1: Enable.</li></ul>|0|

## Setting User Validity Period

To ensure user security, set the user validity period using the system command `chage`.

Example:

```bash
chage [-m mindays] [-M maxdays] [-d lastday] [-I inactive] [-E expiredate] [-W warndays] user
```

For related parameters, see [Table 1](#table_tft_18).

**Table 1<a id="table_tft_18"></a>**  Setting the user validity period

|Parameter|Description|
|--|--|
|-d<br>--lastday|Date of the last change.|
|-E<br>--expiredate|Date on which the user account expires. After this date, the user will be unavailable.|
|-h<br>--help|Displays command help information.|
|-i<br>--iso8601|Changes the expiration date of the user password and displays it in "YYYY-MM-DD" format.|
|-I<br>--inactive|Inactivity period. Sets the password to inactive status after the specified number of days past expiration.|
|-l<br>--list|Lists current settings. Used by non-privileged users to determine when their password or account expires.|
|-m<br>--mindays|Minimum number of days before a password can be changed. Setting this to "0" means the password can be changed at any time.|
|-M<br>--maxdays|Maximum number of days a password remains valid. Setting this to "-1" removes this password check. Setting it to "99999" means unlimited.|
|-R<br>--root|Sets the root directory for command execution to the specified directory.|
|-W<br>--warndays|Number of days before password expiration that the user receives a warning message.|

> [!NOTE]**NOTE**
>
> - The date format is "YYYY-MM-DD". For example, `chage -E 2017-12-01 test` indicates that the password for user `test` expires on December 1, 2017.
> - Replace the user with the actual user. The  default user is `root`.
> - Account passwords should be updated regularly; otherwise, it may lead to security risks.

Example: Modify the validity period of user `test` to 90 days.

```bash
chage -M 90 test
```

## Password Complexity Requirements

Passwords must meet at least the following requirements:

1. The password must be at least 8 characters long.
2. The password must contain a combination of at least two of the following character types:
    - One lowercase letter
    - One uppercase letter
    - One digit
    - One special character: `~!@#$%^&*()-_=+\|[{}];:'",<.>/? and space

3. The password cannot be the same as the account name.

## Account List

|User|Description|Initial Password|Password Change Method|
|--|--|--|--|
| *{MindIO-install-user}* |MindIO TFT installation user.|User-defined.|Use the `passwd` command to change.|
