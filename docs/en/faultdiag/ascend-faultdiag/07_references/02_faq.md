# FAQs

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:29:57.133Z pushedAt=2026-08-24T02:51:43.298Z -->

## Python Version Not Meeting the Requirement During Installation

ascend-fd requires Python 3.7 or later. To use the performance degradation feature, Python 3.8 or later is required. Check the Python version:

```shell
python3 --version
```

If the version does not meet the requirement, upgrade Python or use Conda to create an isolated environment.

## Insufficient Drive Space During Cleaning

The parsing output directory requires at least 5 GB of available drive space. Free up drive space or specify a different output directory.

## Regular Users Cannot Use ascend-fd After Installation by Root

You need to configure the `PATH` environment variable. Query the location of ascend-fd as the `root` user:

```shell
which ascend-fd
```

Add `PATH` as a regular user (assuming ascend-fd is installed in `/usr/local/python3.7.5/bin`):

```shell
export PATH=$PATH:/usr/local/python3.7.5/bin
```

## Diagnosis Fails in Large-Scale Clusters

The default maximum number of file descriptors on a Linux system is 1,024. When the cluster scale exceeds 128 servers (1024 cards), you need to adjust the file descriptor limit:

```shell
ulimit -n 65535
```

## Too Many Faulty Devices in the Diagnostic Report

The terminal displays only 16 faulty device entries by default. The complete information can be viewed in `diag_report.json`.

## "command not found" Is Displayed After ascend-fd Is Installed

- ascend-fd fails to be installed. Reinstall it.
- The current device may have multiple Python versions, and ascend-fd is installed under a non-default Python.

  When ascend-fd is installed under a non-default Python, a prompt similar to the following is displayed:

  ```bash
  WARNING: The script ascend-fd is installed in '/usr/local/python3.8/bin' which is not on PATH.
    Consider adding this directory to PATH or, if you prefer to suppress this warning, use --no-warn-script-location.
  Successfully installed ascend-faultdiag-26.1.0
  ```

  You can resolve this issue as follows:
  1. Locate ascend-fd.

      ```bash
      find / -name ascend-fd
      ```

      The output is similar to the following:

      ```bash
      /usr/local/python3.8/bin/ascend-fd
      ```

  2. Add the directory found in step 1 to `PATH`.

      ```bash
      export PATH=$PATH:/usr/local/python3.8/bin/
      ```

  3. Run the ascend-fd command again.
