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
