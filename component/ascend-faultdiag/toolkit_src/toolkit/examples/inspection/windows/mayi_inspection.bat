::!/usr/bin/env python3
:: -*- coding: utf-8 -*-
:: Copyright 2026 Huawei Technologies Co., Ltd
::
:: Licensed under the Apache License, Version 2.0 (the "License");
:: you may not use this file except in compliance with the License.
:: You may obtain a copy of the License at
::
:: http://www.apache.org/licenses/LICENSE-2.0
::
:: Unless required by applicable law or agreed to in writing, software
:: distributed under the License is distributed on an "AS IS" BASIS,
:: WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
:: See the License for the specific language governing permissions and
:: limitations under the License.
:: ==============================================================================
@echo off
setlocal

set SCRIPT_DIR=%~dp0
for %%I in ("%SCRIPT_DIR%\..\..\..\..") do set PROJECT_ROOT=%%~fI
set PORTABLE_PYTHON=%USERPROFILE%\python_portable\python.exe
set PTH_FILE=%USERPROFILE%\python_portable\python310._pth
set PYTHON_SCRIPT=%SCRIPT_DIR%..\mayi_inspection.py

:: 有免安装python则优先使用，无则使用原来python
if exist "%PORTABLE_PYTHON%" (

    if not exist "%PTH_FILE%" (
        echo 错误：未找到 _pth 文件：%PTH_FILE%，可能无法导入模块
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

    set PY="%PORTABLE_PYTHON%"
) else (
    set PY=python
    set PYTHONPATH=%PROJECT_ROOT%;%PYTHONPATH%
)

if not exist "%PYTHON_SCRIPT%" (
    echo 错误：找不到对应的 Python 脚本 %PYTHON_SCRIPT%
    exit /b 1
)

"%PY%" "%PYTHON_SCRIPT%" %*
pause