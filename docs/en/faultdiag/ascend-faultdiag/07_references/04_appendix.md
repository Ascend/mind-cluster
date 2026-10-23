# Appendix

## Version Number Format

The ascend-fd version number follows the format `X.Y.Z`, for example, `26.2.0`.

## Default Paths

| Path                                                 | Description                          |
|------------------------------------------------------|--------------------------------------|
| `~/.ascend_faultdiag/`                               | Root directory for logs and configuration files |
| `~/.ascend_faultdiag/RUN_LOG/`                       | Runtime log directory                |
| `~/.ascend_faultdiag/ascend_faultdiag_operation.log` | Operation log                        |
| `~/.ascend_faultdiag/install.log`                    | Installation log                     |
| `~/.ascend_faultdiag/config.json`                    | Configuration file                   |
| `~/.ascend_faultdiag/custom-ascend-kg-config.json`   | Custom fault entity file             |
| `~/.ascend_faultdiag/blacklist-config.json`          | Masking rule configuration file     |

> [!NOTE]
>
> The log file size does not exceed 10 MB. When it exceeds this limit, the log file is automatically dumped.

## Component Error Codes

| Status Code            | Meaning                          |
|------------------------|----------------------------------|
| 500 BaseError          | Basic error                      |
| 501 PathError          | Invalid input path               |
| 502 FileNotExistError  | File not exist              |
| 503 InfoNotFoundError  | Requested information not found  |
| 504 InfoIncorrectError | Incorrect information            |
| 505 FileOpenError      | Failed to open the file          |
| 506 InnerError         | Internal service error           |
| 507 ParamError         | Parameter error                  |
| 508 FileTooLarge       | Too many files                   |
| 200 SuccessRet         | Operation succeeded              |

## Known Faults

Refer to [MindCluster 26.2.0 Fault Diagnosis Types](https://raw.gitcode.com/Ascend/mind-cluster/blobs/c49ed595d1b87b34eaf020591778c4536c34b2fe/%E6%95%85%E9%9A%9C%E8%AF%8A%E6%96%AD%E7%B1%BB%E5%9E%8B.xlsx).
