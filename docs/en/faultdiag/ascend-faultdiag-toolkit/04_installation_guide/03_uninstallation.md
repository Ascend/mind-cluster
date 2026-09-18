# Uninstallation

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:15:29.535Z pushedAt=2026-08-24T02:22:41.523Z -->

This section describes the uninstallation process of ascend-fd-tk.

1. Uninstall the Whl package.

    ```bash
    pip3 uninstall ascend-faultdiag-toolkit
    ```

    Example of a successful uninstallation:

    ```txt
    Successfully uninstalled ascend-faultdiag-toolkit-{version}
    ```

2. Clear the cache and runtime data.

   - Linux platform: Tool data is uniformly stored in the user's home directory. Run `rm -rf ~/.ascend-faultdiag-toolkit/`.
   - Windows platform: Tool data is stored in **the current working directory where the tool was started each time**. A `.ascend-faultdiag-toolkit` folder may be generated in each working directory. You can use PowerShell to find and delete them in batches.

     ```powershell
     # Find all residual directories in batches
     Get-ChildItem -Path C:\ -Directory -Recurse -Filter ".ascend-faultdiag-toolkit" -ErrorAction SilentlyContinue | Select-Object FullName

     # Delete in batches after confirmation
     Get-ChildItem -Path C:\ -Directory -Recurse -Filter ".ascend-faultdiag-toolkit" -ErrorAction SilentlyContinue | Remove-Item -Recurse -Force
     ```

     Alternatively, you can locate the `.ascend-faultdiag-toolkit` folder in each commonly used working directory and delete it manually.

3. Verify uninstallation.

    Run the following command to verify whether the uninstallation is successful:

   - Linux platform: Run `which ascend-fd-tk`. If the command does not exist, the uninstallation is successful.
   - Windows platform: Run `where ascend-fd-tk`. If the command does not exist, the uninstallation is successful.
