# Appendix

## Environment Variables

For MindIO TFT environment variables, see [MindIO TFT Environment Variables](../../06_api/13_environment_variable_description.md#section10131935156854).

## Setting User Validity Period

To ensure user security, set the user validity period using the system command `chage`.

Example:

```bash
chage [-m mindays] [-M maxdays] [-d lastday] [-I inactive] [-E expiredate] [-W warndays] user
```

For related parameters, see [Table 1](#table_tft_18).

**Table 1<a id="table_tft_18"></a>** Setting the user validity period

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

> [!NOTE]
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
    - One special character: \`\~!@\#$%^&\*\(\)-\_=+\\|[\{\}];:'",<.\>/? and space

3. The password cannot be the same as the account name.

## Account List

|User|Description|Initial Password|Password Change Method|
|--|--|--|--|
| *{MindIO-install-user}* |MindIO TFT installation user.|User-defined.|Use the `passwd` command to change.|
