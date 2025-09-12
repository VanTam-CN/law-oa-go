@echo off
REM 律师事务所管理系统 Windows 启动脚本

setlocal enabledelayedexpansion

REM 项目路径
set PROJECT_ROOT=%~dp0..
cd /d "%PROJECT_ROOT%"

REM 日志函数
:log_info
echo [INFO] %~1
goto :eof

:log_success
echo [SUCCESS] %~1
goto :eof

:log_warning
echo [WARNING] %~1
goto :eof

:log_error
echo [ERROR] %~1
goto :eof

REM 检查依赖
call :check_dependencies
if %ERRORLEVEL% neq 0 (
    exit /b 1
)

REM 检查配置文件
call :check_config
if %ERRORLEVEL% neq 0 (
    exit /b 1
)

REM 创建必要目录
call :create_directories

REM 下载依赖
call :download_dependencies
if %ERRORLEVEL% neq 0 (
    exit /b 1
)

REM 检查数据库连接
call :check_database
if %ERRORLEVEL% neq 0 (
    exit /b 1
)

REM 检查 Redis 连接
call :check_redis
if %ERRORLEVEL% neq 0 (
    exit /b 1
)

REM 运行数据库迁移
call :run_migrations

REM 启动应用
call :start_application

goto :eof

REM 检查依赖函数
:check_dependencies
call :log_info "检查依赖..."

REM 检查 Go
go version >nul 2>&1
if %ERRORLEVEL% neq 0 (
    call :log_error "Go 未安装，请先安装 Go 1.19 或更高版本"
    exit /b 1
)

call :log_success "Go 环境检查通过"

REM 检查 MySQL
mysql --version >nul 2>&1
if %ERRORLEVEL% neq 0 (
    call :log_warning "MySQL 客户端未安装，请确保数据库服务可用"
)

call :log_success "依赖检查完成"
goto :eof

REM 检查配置文件函数
:check_config
call :log_info "检查配置文件..."

if not exist ".env" (
    call :log_warning ".env 文件不存在，复制 .env.example"
    copy ".env.example" ".env" >nul
    call :log_warning "请编辑 .env 文件配置您的环境变量"
)

if not exist "config\config.yaml" (
    call :log_error "配置文件 config\config.yaml 不存在"
    exit /b 1
)

call :log_success "配置文件检查通过"
goto :eof

REM 创建必要目录函数
:create_directories
call :log_info "创建必要目录..."

if not exist "logs" mkdir logs
if not exist "uploads" mkdir uploads
if not exist "uploads\contract" mkdir uploads\contract
if not exist "uploads\evidence" mkdir uploads\evidence
if not exist "uploads\letter" mkdir uploads\letter
if not exist "uploads\other" mkdir uploads\other

call :log_success "目录创建完成"
goto :eof

REM 下载依赖函数
:download_dependencies
call :log_info "下载 Go 依赖..."

go mod tidy
if %ERRORLEVEL% neq 0 (
    call :log_error "依赖下载失败"
    exit /b 1
)

go mod download
if %ERRORLEVEL% neq 0 (
    call :log_error "依赖下载失败"
    exit /b 1
)

call :log_success "依赖下载完成"
goto :eof

REM 检查数据库连接函数
:check_database
call :log_info "检查数据库连接..."

if exist ".env" (
    REM 读取环境变量
    for /f "tokens=1,2 delims==" %%a in (.env) do (
        if "%%a"=="DB_HOST" set DB_HOST=%%b
        if "%%a"=="DB_PORT" set DB_PORT=%%b
        if "%%a"=="DB_USER" set DB_USER=%%b
        if "%%a"=="DB_PASSWORD" set DB_PASSWORD=%%b
        if "%%a"=="DB_NAME" set DB_NAME=%%b
    )
    
    if defined DB_HOST if defined DB_USER if defined DB_PASSWORD if defined DB_NAME (
        call :log_info "等待数据库启动..."
        timeout /t 5 /nobreak >nul
        
        mysql -h%DB_HOST% -P%DB_PORT% -u%DB_USER% -p%DB_PASSWORD% -e "USE %DB_NAME%" >nul 2>&1
        if %ERRORLEVEL% neq 0 (
            call :log_error "数据库连接失败，请检查配置"
            exit /b 1
        )
        
        call :log_success "数据库连接成功"
    ) else (
        call :log_warning "数据库配置不完整，跳过连接检查"
    )
) else (
    call :log_warning "未找到 .env 文件，跳过数据库检查"
)

goto :eof

REM 检查 Redis 连接函数
:check_redis
call :log_info "检查 Redis 连接..."

if exist ".env" (
    REM 读取环境变量
    for /f "tokens=1,2 delims==" %%a in (.env) do (
        if "%%a"=="REDIS_HOST" set REDIS_HOST=%%b
        if "%%a"=="REDIS_PORT" set REDIS_PORT=%%b
    )
    
    if defined REDIS_HOST (
        redis-cli -h %REDIS_HOST% -p %REDIS_PORT% ping >nul 2>&1
        if %ERRORLEVEL% neq 0 (
            call :log_error "Redis 连接失败，请检查配置"
            exit /b 1
        )
        
        call :log_success "Redis 连接成功"
    ) else (
        call :log_warning "Redis 配置不完整，跳过连接检查"
    )
) else (
    call :log_warning "未找到 .env 文件，跳过 Redis 检查"
)

goto :eof

REM 运行数据库迁移函数
:run_migrations
call :log_info "运行数据库迁移..."

if exist "migrate\main.go" (
    go run migrate\main.go up
    if %ERRORLEVEL% neq 0 (
        call :log_error "数据库迁移失败"
        exit /b 1
    )
    call :log_success "数据库迁移完成"
) else (
    call :log_warning "迁移工具不存在，跳过迁移"
)

goto :eof

REM 启动应用函数
:start_application
call :log_info "启动应用程序..."

REM 加载环境变量
if exist ".env" (
    for /f "tokens=1,2 delims==" %%a in (.env) do (
        set %%a=%%b
    )
)

REM 启动应用
go run main.go
goto :eof