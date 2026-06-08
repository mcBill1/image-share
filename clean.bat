@echo off
chcp 65001 >nul 2>&1
echo ========================================
echo   清理编译产物
echo ========================================
echo.

:: 前端 node_modules
if exist "frontend\node_modules" (
    echo [删除] frontend\node_modules
    rd /s /q "frontend\node_modules"
) else (
    echo [跳过] frontend\node_modules 不存在
)

:: 前端 dist 构建输出
if exist "frontend\dist" (
    echo [删除] frontend\dist
    rd /s /q "frontend\dist"
) else (
    echo [跳过] frontend\dist 不存在
)

:: 后端嵌入的前端目录
if exist "backend\cmd\frontend" (
    echo [删除] backend\cmd\frontend
    rd /s /q "backend\cmd\frontend"
) else (
    echo [跳过] backend\cmd\frontend 不存在
)

if exist "backend\frontend" (
    echo [删除] backend\frontend
    rd /s /q "backend\frontend"
) else (
    echo [跳过] backend\frontend 不存在
)

:: Go 构建缓存
echo [清理] Go 构建缓存
go clean -cache 2>nul

echo.
echo ========================================
echo   清理完成
echo   保留: 源代码 + build\ 目录
echo ========================================
pause
