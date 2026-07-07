# Introduction

## Overview

Ascend-faultdiag-toolkit, a link diagnostic tool, analyzes device link faults through online collection or offline log analysis, covering servers, switching devices (L1/L2 UnifiedBus switches, RoCE switches), and BMC management.

### Online Collection

The user provides connection information (username, password/key/password-free authentication) for the target device, and the tool accesses the device to collect data.

### Offline Analysis

The user imports in-band logs from the collected server, BMC dump logs, and diagnostic information logs from switching devices to analyze key information.

### Fault Analysis

Combining online/offline collected information and inheriting fault patterns, the tool performs fault analysis.

## Instructions

For details, see [Usage Guide](../../../../component/ascend-faultdiag/toolkit_src/README.md).
