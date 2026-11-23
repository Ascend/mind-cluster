REM!/usr/bin/env python3
REM -*- coding: utf-8 -*-
REM Copyright 2026 Huawei Technologies Co., Ltd
REM
REM Licensed under the Apache License, Version 2.0 (the "License");
REM you may not use this file except in compliance with the License.
REM You may obtain a copy of the License at
REM
REM     http://www.apache.org/licenses/LICENSE-2.0
REM
REM Unless required by applicable law or agreed to in writing, software
REM distributed under the License is distributed on an "AS IS" BASIS,
REM WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
REM See the License for the specific language governing permissions and
REM limitations under the License.
REM ==============================================================================
@echo off
chcp 65001 >nul
setlocal

set SCRIPT_DIR=%~dp0
set CACHE_DIR=%SCRIPT_DIR%..\..\cache

if exist "%CACHE_DIR%" (
    echo 正在清理缓存目录: %CACHE_DIR%
    rd /s /q %CACHE_DIR% >nul 2>&1
    rem 重新创建bmc_dump_cache和parse_cache空文件夹
    mkdir "%CACHE_DIR%\host_dump_cache" 2>nul
    mkdir "%CACHE_DIR%\bmc_dump_cache" 2>nul
    mkdir "%CACHE_DIR%\switch_cli_output_cache" 2>nul


    echo 缓存目录清理完成，已重新创建空文件夹
) else (
    echo 缓存目录不存在: %CACHE_DIR%
)
pause
exit /b 0
