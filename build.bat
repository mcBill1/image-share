@echo off
chcp 936 >nul 2>&1
setlocal enabledelayedexpansion

set "PROJECT_ROOT=%~dp0"
set "FRONTEND_DIR=%PROJECT_ROOT%frontend"
set "BACKEND_DIR=%PROJECT_ROOT%backend"
set "BUILD_DIR=%PROJECT_ROOT%build"
set "EMBED_DIR=%BACKEND_DIR%\cmd\frontend"

:start

echo ========================================
echo        ImageShare Build Script
echo ========================================
echo.
echo 请选择编译选项 / Please select build option:
echo.
echo 1. Windows AMD64
echo 2. Linux AMD64
echo 3. Linux ARM64
echo 4. 退出 / EXIT
echo.
set /p "choice=请输入选项 / Enter option (1/2/3): "

if "%choice%"=="1" goto :build_windows
if "%choice%"=="2" goto :build_linux_amd64
if "%choice%"=="3" goto :build_linux_arm64
if "%choice%"=="4" goto :exitbash

echo 无效选项，请输入 1、2 或 3 / Invalid option, please enter 1, 2 or 3
pause
exit /b 1

:build_windows
set "GOOS=windows"
set "GOARCH=amd64"
set "OUTPUT=%BUILD_DIR%\imageshare-windows-amd64.exe"
goto :build_common

:build_linux_amd64
set "GOOS=linux"
set "GOARCH=amd64"
set "OUTPUT=%BUILD_DIR%\imageshare-linux-amd64"
goto :build_common

:build_linux_arm64
set "GOOS=linux"
set "GOARCH=arm64"
set "OUTPUT=%BUILD_DIR%\imageshare-linux-arm64"
goto :build_common

:exitbash
exit

:build_common
echo.
echo ========================================
echo 步骤 1/3：编译前端 / Step 1/3: Build Frontend
echo ========================================
cd /d "%FRONTEND_DIR%"

if not exist "node_modules" (
    echo [安装依赖] npm install / [Install Dependencies] npm install
    call npm install
    if !errorlevel! neq 0 (
        echo [错误] npm install 失败 / [Error] npm install failed
        pause
        exit /b 1
    )
)

echo [编译前端] npm run build / [Build Frontend] npm run build
call npm run build
if !errorlevel! neq 0 (
    echo [错误] 前端编译失败 / [Error] Frontend build failed
    pause
    exit /b 1
)

echo.
echo ========================================
echo 步骤 2/3：复制前端到后端 / Step 2/3: Copy Frontend to Backend
echo ========================================

if exist "%EMBED_DIR%" (
    echo [清理旧文件] rd /s /q "%EMBED_DIR%" / [Clean Old Files]
    rd /s /q "%EMBED_DIR%"
)

echo [复制文件] xcopy "%FRONTEND_DIR%\dist" "%EMBED_DIR%" /e /h /y
xcopy "%FRONTEND_DIR%\dist" "%EMBED_DIR%" /e /h /y
if !errorlevel! neq 0 (
    echo [错误] 复制前端文件失败 / [Error] Copy frontend files failed
    pause
    exit /b 1
)

echo.
echo ========================================
echo 步骤 3/3：编译后端 (%GOOS%/%GOARCH%) / Step 3/3: Build Backend
echo ========================================
cd /d "%BACKEND_DIR%"

if not exist "%BUILD_DIR%" (
    mkdir "%BUILD_DIR%"
)

echo [编译] %GOOS% %GOARCH% / [Build] %GOOS% %GOARCH%
go build -o "%OUTPUT%" ./cmd
if !errorlevel! neq 0 (
    echo [错误] %GOOS%/%GOARCH% 编译失败 / [Error] %GOOS%/%GOARCH% build failed
    pause
    exit /b 1
)

echo.
echo ========================================
echo 编译完成！/ Build Complete!
echo ========================================
echo.
echo 输出文件 / Output: %OUTPUT%
echo.
dir "%OUTPUT%" 2>nul
echo.
pause
goto :start