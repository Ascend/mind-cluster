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
setlocal
chcp 65001 >nul 2>&1

set SCRIPT_DIR=%~dp0
for %%I in ("%SCRIPT_DIR%\..\..\..\..") do set PROJECT_ROOT=%%~fI

set PORTABLE_PYTHON=%USERPROFILE%\python_portable\python.exe
set PTH_FILE=%USERPROFILE%\python_portable\python310._pth
set PYTHON_SCRIPT=%SCRIPT_DIR%..\auto_collect.py

:: -----------------------------
:: 优先使用免安装 Python
:: -----------------------------
if exist "%PORTABLE_PYTHON%" (

    if not exist "%PTH_FILE%" (
        echo 错误：未找到 _pth 文件：%PTH_FILE%，可能无法导入模块
        pause
        exit /b 1
    )

    :: 备份 _pth
    if exist "%PTH_FILE%" (
        if not exist "%PTH_FILE%.bak" (
            copy "%PTH_FILE%" "%PTH_FILE%.bak" >nul
        )
    )

    :: 写入 _pth（加入项目根目录）
    if exist "%PTH_FILE%" (
        (
            echo python310.zip
            echo .
            echo import site
            echo %PROJECT_ROOT%
        ) > "%PTH_FILE%"
    )

    set PY=%PORTABLE_PYTHON%

) else (

    :: -----------------------------
    :: 检测本地 python
    :: -----------------------------
    where.exe python >nul 2>nul
    if errorlevel 1 (
        echo python 不存在（或未配置 PATH）
        pause
        exit /b 1
    )

    set PY=python
    set PYTHONPATH=%PROJECT_ROOT%;%PYTHONPATH%
)

:: -----------------------------
:: 检查 Python 脚本
:: -----------------------------
if not exist "%PYTHON_SCRIPT%" (
    echo 错误：找不到对应的 Python 脚本 %PYTHON_SCRIPT%
    pause
    exit /b 1
)

:: -----------------------------
:: 执行
:: -----------------------------
"%PY%" "%PYTHON_SCRIPT%" %*

pause
exit /b 0
