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
chcp 65001 >nul 2>&1
setlocal enabledelayedexpansion
title 自动部署免安装Python+依赖

:: ===================== 全局配置 =====================
set "PYTHON_VERSION=3.10.11"
set "PYTHON_ARCH=amd64"
set "PYTHON_ROOT=%USERPROFILE%\python_portable"
set "REQUIREMENTS_PATH=%~dp0..\..\..\requirements.txt"

set "PYTHON_URL=https://mirrors.huaweicloud.com/python/%PYTHON_VERSION%/python-%PYTHON_VERSION%-embed-%PYTHON_ARCH%.zip"
set "DOWNLOAD_ZIP=%PYTHON_ROOT%\python_embed.zip"

:: ===================== Step 1: 检查已有 Python =====================
if exist "%PYTHON_ROOT%\python.exe" (
    echo [INFO] 检测到已有embed Python：%PYTHON_ROOT%
    set "PY_EXE=%PYTHON_ROOT%\python.exe"
    goto check_pip
)

mkdir "%PYTHON_ROOT%" 2>nul

:: ===================== Step 2: 下载 Python =====================
set "RETRY=0"
:download_python
echo.
echo [INFO] 正在下载 Python Embed %PYTHON_VERSION% ...
powershell -Command "Invoke-WebRequest '%PYTHON_URL%' -OutFile '%DOWNLOAD_ZIP%' -UseBasicParsing" 2>nul

if exist "%DOWNLOAD_ZIP%" (
    echo [INFO] Python 下载成功
) else (
    set /a RETRY+=1
    if %RETRY% GEQ 3 (
        echo 错误：Python 下载失败三次，终止
        pause
        exit /b 1
    )
    echo [WARN] 下载失败，等待 2 秒后重试（%RETRY%/3）...
    timeout /t 2 >nul
    goto download_python
)

:: ===================== Step 3: 解压 =====================
echo [INFO] 解压 Python...
powershell -Command "Expand-Archive -Path '%DOWNLOAD_ZIP%' -DestinationPath '%PYTHON_ROOT%' -Force"
del "%DOWNLOAD_ZIP%" >nul 2>&1

set "PY_EXE=%PYTHON_ROOT%\python.exe"

:: ===================== Step 4: 修复 _pth =====================
echo [INFO] 修复 python310._pth ...
set "PTH_FILE=%PYTHON_ROOT%\python310._pth"

(
echo python310.zip
echo .
echo Lib/site-packages
echo import site
echo
) > "%PTH_FILE%"

:: 强制 ANSI 无 BOM
powershell -Command "$c=Get-Content '%PTH_FILE%' -Raw; Set-Content '%PTH_FILE%' -Value $c -Encoding ASCII"

:: 创建 site-packages
mkdir "%PYTHON_ROOT%\Lib\site-packages" 2>nul

:: ===================== Step 5: 检查 pip =====================
:check_pip
echo.
echo [INFO] 检查 pip...

"%PY_EXE%" -m pip -V >nul 2>&1
if %errorlevel%==0 (
    echo [INFO] 已检测到 pip
    goto load_requirements
)

echo [INFO] pip 不存在，正在安装...

set "PIP_BOOT=%PYTHON_ROOT%\get-pip.py"

powershell -Command "Invoke-WebRequest 'https://bootstrap.pypa.io/get-pip.py' -OutFile '%PIP_BOOT%' -UseBasicParsing" 2>nul
if not exist "%PIP_BOOT%" (
    echo 错误：get-pip.py 下载失败
    pause
    exit /b 1
)

"%PY_EXE%" "%PIP_BOOT%" --trusted-host mirrors.tools.huawei.com --no-warn-script-location
if %errorlevel% neq 0 (
    echo 错误：pip 安装失败
    pause
    exit /b 1
)

del "%PIP_BOOT%" >nul 2>&1
echo [INFO] pip 安装完成

:: ===================== Step 6: 检查 requirements =====================
:load_requirements
echo.
echo [INFO] 正在检查 requirements.txt ...

if exist "%REQUIREMENTS_PATH%" (
    echo [INFO] 找到本地 requirements.txt：%REQUIREMENTS_PATH%
) else (
    echo 错误：未找到 requirements.txt ：%REQUIREMENTS_PATH%
    pause
    exit /b 1
)

:: 可选：验证文件是否为空
powershell -Command "$content = Get-Content '%REQUIREMENTS_PATH%' -Raw; if ($content.Trim() -eq '') { exit 1 }"
if errorlevel 1 (
    echo 错误：requirements.txt 内容为空
    pause
    exit /b 1
)

echo [INFO] requirements.txt 已就绪

:: ===================== Step 7: 安装依赖 =====================
echo.
echo [INFO] 正在安装依赖到embed Python...
set "SITE_PACKAGES=%PYTHON_ROOT%\Lib\site-packages"

"%PY_EXE%" -m pip install -r %REQUIREMENTS_PATH% ^
    --target "%SITE_PACKAGES%" ^
    --no-cache-dir ^
    --retries 5 ^
    --timeout 300 ^
    -i https://pypi.tuna.tsinghua.edu.cn/simple

if %errorlevel% neq 0 (
    echo 错误：依赖安装失败，请检查 requirements.txt
    pause
    exit /b 1
)

echo.
echo 成功：依赖安装成功：%SITE_PACKAGES%

:: ===================== Step 8: 完成 =====================
echo.
echo ==================================================
echo 已成功部署免安装 Python
echo 路径：%PYTHON_ROOT%
echo ==================================================
pause
exit /b 0
