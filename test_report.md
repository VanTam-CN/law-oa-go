# Law OA Go 测试报告

## 测试执行信息
- **执行时间**: 2025年 9月16日 星期二 10时17分30秒 CST
- **测试环境**: dev
- **测试类型**: all
- **并行执行**: false
- **并行任务数**: 4
- **超时设置**: 30m

## 测试结果概览
### 单元测试
- **状态**: 失败
- **覆盖率**: 0%
### 集成测试
- **状态**: 失败
- **覆盖率**: 0%
### 端到端测试
- **状态**: 失败
### 性能测试
- **状态**: 失败
### 安全测试
- **状态**: 通过

## 详细报告

### 文件清单
- [性能测试报告](test-reports/performance.txt)

## 执行日志

[0;34m[INFO][0m 2025-09-16 09:48:09 开始执行 dev 环境的 all 测试...
[0;34m[INFO][0m 2025-09-16 09:48:09 并行模式: false, 任务数: 4
[0;34m[INFO][0m 2025-09-16 09:48:09 构建应用...
[0;32m[SUCCESS][0m 2025-09-16 09:48:17 应用构建成功
[0;34m[INFO][0m 2025-09-16 09:48:17 运行单元测试...
?   	law-oa-go/internal/api	[no test files]
=== RUN   TestTokenManager_NewTokenManager
=== RUN   TestTokenManager_NewTokenManager/创建TokenManager
--- PASS: TestTokenManager_NewTokenManager (0.04s)
    --- PASS: TestTokenManager_NewTokenManager/创建TokenManager (0.04s)
=== RUN   TestTokenManager_CreateTokens
=== RUN   TestTokenManager_CreateTokens/创建访问令牌和刷新令牌
--- PASS: TestTokenManager_CreateTokens (0.07s)
    --- PASS: TestTokenManager_CreateTokens/创建访问令牌和刷新令牌 (0.07s)
=== RUN   TestTokenManager_VerifyToken
=== RUN   TestTokenManager_VerifyToken/验证有效令牌
=== RUN   TestTokenManager_VerifyToken/验证无效令牌
--- PASS: TestTokenManager_VerifyToken (0.02s)
    --- PASS: TestTokenManager_VerifyToken/验证有效令牌 (0.01s)
    --- PASS: TestTokenManager_VerifyToken/验证无效令牌 (0.01s)
PASS
ok  	law-oa-go/internal/auth	1.912s
testing: warning: no tests to run
PASS
ok  	law-oa-go/internal/benchmark	1.335s [no tests to run]
=== RUN   TestCacheService_NewCacheService
=== RUN   TestCacheService_NewCacheService/创建缓存服务
--- PASS: TestCacheService_NewCacheService (0.00s)
    --- PASS: TestCacheService_NewCacheService/创建缓存服务 (0.00s)
=== RUN   TestCacheService_getFullKey
=== RUN   TestCacheService_getFullKey/生成完整缓存键
=== RUN   TestCacheService_getFullKey/生成完整缓存键/带前缀
=== RUN   TestCacheService_getFullKey/生成完整缓存键/空前缀
=== RUN   TestCacheService_getFullKey/生成完整缓存键/只有前缀
=== RUN   TestCacheService_getFullKey/生成完整缓存键/空键
--- PASS: TestCacheService_getFullKey (0.00s)
    --- PASS: TestCacheService_getFullKey/生成完整缓存键 (0.00s)
        --- PASS: TestCacheService_getFullKey/生成完整缓存键/带前缀 (0.00s)
        --- PASS: TestCacheService_getFullKey/生成完整缓存键/空前缀 (0.00s)
        --- PASS: TestCacheService_getFullKey/生成完整缓存键/只有前缀 (0.00s)
        --- PASS: TestCacheService_getFullKey/生成完整缓存键/空键 (0.00s)
=== RUN   TestCacheService_Set
=== RUN   TestCacheService_Set/设置缓存失败
--- PASS: TestCacheService_Set (0.00s)
    --- PASS: TestCacheService_Set/设置缓存失败 (0.00s)
=== RUN   TestCacheService_Get
=== RUN   TestCacheService_Get/缓存未命中
    cache_test.go:67: 跳过需要Redis连接的测试
--- PASS: TestCacheService_Get (0.00s)
    --- SKIP: TestCacheService_Get/缓存未命中 (0.00s)
=== RUN   TestCacheService_Exists
=== RUN   TestCacheService_Exists/检查缓存存在
    cache_test.go:74: 跳过需要Redis连接的测试
--- PASS: TestCacheService_Exists (0.00s)
    --- SKIP: TestCacheService_Exists/检查缓存存在 (0.00s)
=== RUN   TestCacheService_Increment
=== RUN   TestCacheService_Increment/递增计数器
    cache_test.go:81: 跳过需要Redis连接的测试
--- PASS: TestCacheService_Increment (0.00s)
    --- SKIP: TestCacheService_Increment/递增计数器 (0.00s)
=== RUN   TestCacheService_Ping
=== RUN   TestCacheService_Ping/测试Redis连接
    cache_test.go:88: 跳过需要Redis连接的测试
--- PASS: TestCacheService_Ping (0.00s)
    --- SKIP: TestCacheService_Ping/测试Redis连接 (0.00s)
=== RUN   TestCacheKey
=== RUN   TestCacheKey/生成各种缓存键
--- PASS: TestCacheKey (0.00s)
    --- PASS: TestCacheKey/生成各种缓存键 (0.00s)
=== RUN   TestCacheKeyGenerator
=== RUN   TestCacheKeyGenerator/缓存键生成器
--- PASS: TestCacheKeyGenerator (0.00s)
    --- PASS: TestCacheKeyGenerator/缓存键生成器 (0.00s)
PASS
ok  	law-oa-go/internal/cache	1.661s
=== RUN   TestResponse_Success
=== RUN   TestResponse_Success/成功响应_-_返回数据
--- PASS: TestResponse_Success (0.01s)
    --- PASS: TestResponse_Success/成功响应_-_返回数据 (0.01s)
=== RUN   TestResponse_SuccessWithMessage
=== RUN   TestResponse_SuccessWithMessage/成功响应_-_自定义消息
--- PASS: TestResponse_SuccessWithMessage (0.00s)
    --- PASS: TestResponse_SuccessWithMessage/成功响应_-_自定义消息 (0.00s)
=== RUN   TestResponse_SuccessWithPage
=== RUN   TestResponse_SuccessWithPage/分页响应
--- PASS: TestResponse_SuccessWithPage (0.00s)
    --- PASS: TestResponse_SuccessWithPage/分页响应 (0.00s)
=== RUN   TestResponse_Error
=== RUN   TestResponse_Error/错误响应
--- PASS: TestResponse_Error (0.00s)
    --- PASS: TestResponse_Error/错误响应 (0.00s)
=== RUN   TestResponse_BadRequest
=== RUN   TestResponse_BadRequest/400错误
--- PASS: TestResponse_BadRequest (0.00s)
    --- PASS: TestResponse_BadRequest/400错误 (0.00s)
=== RUN   TestResponse_Unauthorized
=== RUN   TestResponse_Unauthorized/401错误
--- PASS: TestResponse_Unauthorized (0.00s)
    --- PASS: TestResponse_Unauthorized/401错误 (0.00s)
=== RUN   TestResponse_Forbidden
=== RUN   TestResponse_Forbidden/403错误
--- PASS: TestResponse_Forbidden (0.00s)
    --- PASS: TestResponse_Forbidden/403错误 (0.00s)
=== RUN   TestResponse_NotFound
=== RUN   TestResponse_NotFound/404错误
--- PASS: TestResponse_NotFound (0.00s)
    --- PASS: TestResponse_NotFound/404错误 (0.00s)
=== RUN   TestResponse_InternalServerError
=== RUN   TestResponse_InternalServerError/500错误
--- PASS: TestResponse_InternalServerError (0.00s)
    --- PASS: TestResponse_InternalServerError/500错误 (0.00s)
=== RUN   TestResponse_ValidationError
=== RUN   TestResponse_ValidationError/验证错误
--- PASS: TestResponse_ValidationError (0.00s)
    --- PASS: TestResponse_ValidationError/验证错误 (0.00s)
=== RUN   TestRequestBodyBuffer
=== RUN   TestRequestBodyBuffer/请求体缓冲区基本功能
=== RUN   TestRequestBodyBuffer/多次读取
=== RUN   TestRequestBodyBuffer/关闭缓冲区
--- PASS: TestRequestBodyBuffer (0.00s)
    --- PASS: TestRequestBodyBuffer/请求体缓冲区基本功能 (0.00s)
    --- PASS: TestRequestBodyBuffer/多次读取 (0.00s)
    --- PASS: TestRequestBodyBuffer/关闭缓冲区 (0.00s)
=== RUN   TestBusinessError
=== RUN   TestBusinessError/业务错误基本功能
=== RUN   TestBusinessError/无底层错误的业务错误
--- PASS: TestBusinessError (0.00s)
    --- PASS: TestBusinessError/业务错误基本功能 (0.00s)
    --- PASS: TestBusinessError/无底层错误的业务错误 (0.00s)
=== RUN   TestValidationError
=== RUN   TestValidationError/验证错误创建
--- PASS: TestValidationError (0.00s)
    --- PASS: TestValidationError/验证错误创建 (0.00s)
=== RUN   TestNotFoundError
=== RUN   TestNotFoundError/未找到错误创建
--- PASS: TestNotFoundError (0.00s)
    --- PASS: TestNotFoundError/未找到错误创建 (0.00s)
=== RUN   TestUnauthorizedError
=== RUN   TestUnauthorizedError/未授权错误创建
--- PASS: TestUnauthorizedError (0.00s)
    --- PASS: TestUnauthorizedError/未授权错误创建 (0.00s)
=== RUN   TestForbiddenError
=== RUN   TestForbiddenError/禁止访问错误创建
--- PASS: TestForbiddenError (0.00s)
    --- PASS: TestForbiddenError/禁止访问错误创建 (0.00s)
=== RUN   TestDatabaseError
=== RUN   TestDatabaseError/数据库错误创建
--- PASS: TestDatabaseError (0.00s)
    --- PASS: TestDatabaseError/数据库错误创建 (0.00s)
=== RUN   TestInternalError
=== RUN   TestInternalError/内部错误创建
--- PASS: TestInternalError (0.00s)
    --- PASS: TestInternalError/内部错误创建 (0.00s)
=== RUN   TestErrorCheckingFunctions
=== RUN   TestErrorCheckingFunctions/检查未找到错误
=== RUN   TestErrorCheckingFunctions/检查验证错误
=== RUN   TestErrorCheckingFunctions/检查冲突错误
=== RUN   TestErrorCheckingFunctions/检查未授权错误
=== RUN   TestErrorCheckingFunctions/检查禁止访问错误
=== RUN   TestErrorCheckingFunctions/检查数据库错误
--- PASS: TestErrorCheckingFunctions (0.00s)
    --- PASS: TestErrorCheckingFunctions/检查未找到错误 (0.00s)
    --- PASS: TestErrorCheckingFunctions/检查验证错误 (0.00s)
    --- PASS: TestErrorCheckingFunctions/检查冲突错误 (0.00s)
    --- PASS: TestErrorCheckingFunctions/检查未授权错误 (0.00s)
    --- PASS: TestErrorCheckingFunctions/检查禁止访问错误 (0.00s)
    --- PASS: TestErrorCheckingFunctions/检查数据库错误 (0.00s)
=== RUN   TestExtractBusinessError
=== RUN   TestExtractBusinessError/提取业务错误
=== RUN   TestExtractBusinessError/提取嵌套业务错误
=== RUN   TestExtractBusinessError/非业务错误
--- PASS: TestExtractBusinessError (0.00s)
    --- PASS: TestExtractBusinessError/提取业务错误 (0.00s)
    --- PASS: TestExtractBusinessError/提取嵌套业务错误 (0.00s)
    --- PASS: TestExtractBusinessError/非业务错误 (0.00s)
=== RUN   TestResponseWithMockFactory
=== RUN   TestResponseWithMockFactory/使用Mock工厂数据测试响应
--- PASS: TestResponseWithMockFactory (0.00s)
    --- PASS: TestResponseWithMockFactory/使用Mock工厂数据测试响应 (0.00s)
=== RUN   TestErrorHandlingIntegration
=== RUN   TestErrorHandlingIntegration/错误处理流程集成测试
--- PASS: TestErrorHandlingIntegration (0.00s)
    --- PASS: TestErrorHandlingIntegration/错误处理流程集成测试 (0.00s)
=== RUN   TestEnvFunctions
=== RUN   TestEnvFunctions/GetEnv_获取字符串环境变量
=== RUN   TestEnvFunctions/GetEnvInt_获取整数环境变量
=== RUN   TestEnvFunctions/GetEnvBool_获取布尔环境变量
--- PASS: TestEnvFunctions (0.00s)
    --- PASS: TestEnvFunctions/GetEnv_获取字符串环境变量 (0.00s)
    --- PASS: TestEnvFunctions/GetEnvInt_获取整数环境变量 (0.00s)
    --- PASS: TestEnvFunctions/GetEnvBool_获取布尔环境变量 (0.00s)
=== RUN   TestStreamResponse
=== RUN   TestStreamResponse/创建流式响应
=== RUN   TestStreamResponse/发送数据到流
=== RUN   TestStreamResponse/关闭流式响应
=== RUN   TestStreamResponse/向已关闭的流发送数据
--- PASS: TestStreamResponse (0.00s)
    --- PASS: TestStreamResponse/创建流式响应 (0.00s)
    --- PASS: TestStreamResponse/发送数据到流 (0.00s)
    --- PASS: TestStreamResponse/关闭流式响应 (0.00s)
    --- PASS: TestStreamResponse/向已关闭的流发送数据 (0.00s)
=== RUN   TestStreamResponse_Render
=== RUN   TestStreamResponse_Render/渲染流式响应
--- PASS: TestStreamResponse_Render (0.00s)
    --- PASS: TestStreamResponse_Render/渲染流式响应 (0.00s)
=== RUN   TestStreamPaginatedResponse
=== RUN   TestStreamPaginatedResponse/创建分页流式响应
=== RUN   TestStreamPaginatedResponse/分页流式响应内容类型
--- PASS: TestStreamPaginatedResponse (0.00s)
    --- PASS: TestStreamPaginatedResponse/创建分页流式响应 (0.00s)
    --- PASS: TestStreamPaginatedResponse/分页流式响应内容类型 (0.00s)
=== RUN   TestStreamSuccess
=== RUN   TestStreamSuccess/流式成功响应基本测试
--- PASS: TestStreamSuccess (0.00s)
    --- PASS: TestStreamSuccess/流式成功响应基本测试 (0.00s)
=== RUN   TestStreamPaginatedSuccess
=== RUN   TestStreamPaginatedSuccess/分页流式成功响应基本测试
--- PASS: TestStreamPaginatedSuccess (0.00s)
    --- PASS: TestStreamPaginatedSuccess/分页流式成功响应基本测试 (0.00s)
=== RUN   TestStreamResponse_WriteContentType
=== RUN   TestStreamResponse_WriteContentType/设置内容类型
--- PASS: TestStreamResponse_WriteContentType (0.00s)
    --- PASS: TestStreamResponse_WriteContentType/设置内容类型 (0.00s)
=== RUN   TestStreamResponse_Concurrent
=== RUN   TestStreamResponse_Concurrent/并发发送数据
--- PASS: TestStreamResponse_Concurrent (0.01s)
    --- PASS: TestStreamResponse_Concurrent/并发发送数据 (0.01s)
=== RUN   TestUnifiedAPIResponse
=== RUN   TestUnifiedAPIResponse/Test_API_Success_Response
=== RUN   TestUnifiedAPIResponse/Test_API_Error_Response
=== RUN   TestUnifiedAPIResponse/Test_API_Error_With_Context
=== RUN   TestUnifiedAPIResponse/Test_API_Error_With_Suggestions
=== RUN   TestUnifiedAPIResponse/Test_API_Pagination_Response
=== RUN   TestUnifiedAPIResponse/Test_Backward_Compatibility_-_Success
=== RUN   TestUnifiedAPIResponse/Test_Backward_Compatibility_-_Error
=== RUN   TestUnifiedAPIResponse/Test_Request_ID_Handling
=== RUN   TestUnifiedAPIResponse/Test_Response_Builder
--- PASS: TestUnifiedAPIResponse (0.01s)
    --- PASS: TestUnifiedAPIResponse/Test_API_Success_Response (0.00s)
    --- PASS: TestUnifiedAPIResponse/Test_API_Error_Response (0.00s)
    --- PASS: TestUnifiedAPIResponse/Test_API_Error_With_Context (0.00s)
    --- PASS: TestUnifiedAPIResponse/Test_API_Error_With_Suggestions (0.00s)
    --- PASS: TestUnifiedAPIResponse/Test_API_Pagination_Response (0.00s)
    --- PASS: TestUnifiedAPIResponse/Test_Backward_Compatibility_-_Success (0.00s)
    --- PASS: TestUnifiedAPIResponse/Test_Backward_Compatibility_-_Error (0.00s)
    --- PASS: TestUnifiedAPIResponse/Test_Request_ID_Handling (0.00s)
    --- PASS: TestUnifiedAPIResponse/Test_Response_Builder (0.00s)
=== RUN   TestPaginationCalculation
=== RUN   TestPaginationCalculation/Test_Single_Page
=== RUN   TestPaginationCalculation/Test_Multiple_Pages
=== RUN   TestPaginationCalculation/Test_Last_Page
--- PASS: TestPaginationCalculation (0.00s)
    --- PASS: TestPaginationCalculation/Test_Single_Page (0.00s)
    --- PASS: TestPaginationCalculation/Test_Multiple_Pages (0.00s)
    --- PASS: TestPaginationCalculation/Test_Last_Page (0.00s)
PASS
ok  	law-oa-go/internal/common	1.729s
=== RUN   TestNewConcurrentService
--- PASS: TestNewConcurrentService (0.00s)
=== RUN   TestConcurrentService_StartStop
--- PASS: TestConcurrentService_StartStop (0.00s)
=== RUN   TestConcurrentService_SubmitTask
Task test_task completed successfully
--- PASS: TestConcurrentService_SubmitTask (0.20s)
=== RUN   TestConcurrentService_SubmitTaskAndWait
Task wait_task completed successfully
--- PASS: TestConcurrentService_SubmitTaskAndWait (0.01s)
=== RUN   TestConcurrentService_SubmitBatchTasks
Task batch_1757987318596302000 completed successfully
Task batch_1757987318596302000 completed successfully
--- PASS: TestConcurrentService_SubmitBatchTasks (0.37s)
=== RUN   TestConcurrentService_SubmitDatabaseTask
Task db_1757987318962738000 completed successfully
Task db_1757987318962738000 completed successfully
--- PASS: TestConcurrentService_SubmitDatabaseTask (0.01s)
=== RUN   TestConcurrentService_SubmitFileTask
Task file_1757987318975456000 completed successfully
Task file_1757987318975456000 completed successfully
--- PASS: TestConcurrentService_SubmitFileTask (0.01s)
=== RUN   TestConcurrentService_SubmitAPITask
==================
WARNING: DATA RACE
Write at 0x00c0005a408f by goroutine 112:
  law-oa-go/internal/concurrency.TestConcurrentService_SubmitAPITask.func1()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/concurrent_service_test.go:208 +0x14e
  law-oa-go/internal/concurrency.(*APIRequestTask).Execute()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/concurrent_service.go:168 +0x16b
  law-oa-go/internal/concurrency.(*WorkerPool).processTask()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/worker_pool.go:266 +0x4d3
  law-oa-go/internal/concurrency.(*WorkerPool).worker()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/worker_pool.go:227 +0xcc
  law-oa-go/internal/concurrency.(*WorkerPool).Start.gowrap2()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/worker_pool.go:104 +0x44

Previous write at 0x00c0005a408f by goroutine 113:
  law-oa-go/internal/concurrency.TestConcurrentService_SubmitAPITask.func1()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/concurrent_service_test.go:208 +0x14e
  law-oa-go/internal/concurrency.(*APIRequestTask).Execute()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/concurrent_service.go:168 +0x16b
  law-oa-go/internal/concurrency.(*WorkerPool).processTask()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/worker_pool.go:266 +0x4d3
  law-oa-go/internal/concurrency.(*WorkerPool).worker()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/worker_pool.go:227 +0xcc
  law-oa-go/internal/concurrency.(*WorkerPool).Start.gowrap2()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/worker_pool.go:104 +0x44

Goroutine 112 (running) created at:
  law-oa-go/internal/concurrency.(*WorkerPool).Start()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/worker_pool.go:104 +0x117
  law-oa-go/internal/concurrency.(*ConcurrentService).Start()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/concurrent_service.go:222 +0x51
  law-oa-go/internal/concurrency.TestConcurrentService_SubmitAPITask()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/concurrent_service_test.go:192 +0x3d
  testing.tRunner()
      /usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1690 +0x226
  testing.(*T).Run.gowrap1()
      /usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1743 +0x44

Goroutine 113 (running) created at:
  law-oa-go/internal/concurrency.(*WorkerPool).Start()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/worker_pool.go:104 +0x117
  law-oa-go/internal/concurrency.(*ConcurrentService).Start()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/concurrent_service.go:222 +0x51
  law-oa-go/internal/concurrency.TestConcurrentService_SubmitAPITask()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/concurrent_service_test.go:192 +0x3d
  testing.tRunner()
      /usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1690 +0x226
  testing.(*T).Run.gowrap1()
      /usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1743 +0x44
==================
==================
WARNING: DATA RACE
Write at 0x00c0005c8128 by goroutine 112:
  law-oa-go/internal/concurrency.(*APIRequestTask).Execute()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/concurrent_service.go:173 +0x191
  law-oa-go/internal/concurrency.(*WorkerPool).processTask()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/worker_pool.go:266 +0x4d3
  law-oa-go/internal/concurrency.(*WorkerPool).worker()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/worker_pool.go:227 +0xcc
  law-oa-go/internal/concurrency.(*WorkerPool).Start.gowrap2()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/worker_pool.go:104 +0x44

Previous write at 0x00c0005c8128 by goroutine 113:
  law-oa-go/internal/concurrency.(*APIRequestTask).Execute()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/concurrent_service.go:173 +0x191
  law-oa-go/internal/concurrency.(*WorkerPool).processTask()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/worker_pool.go:266 +0x4d3
  law-oa-go/internal/concurrency.(*WorkerPool).worker()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/worker_pool.go:227 +0xcc
  law-oa-go/internal/concurrency.(*WorkerPool).Start.gowrap2()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/worker_pool.go:104 +0x44

Goroutine 112 (running) created at:
  law-oa-go/internal/concurrency.(*WorkerPool).Start()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/worker_pool.go:104 +0x117
  law-oa-go/internal/concurrency.(*ConcurrentService).Start()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/concurrent_service.go:222 +0x51
  law-oa-go/internal/concurrency.TestConcurrentService_SubmitAPITask()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/concurrent_service_test.go:192 +0x3d
  testing.tRunner()
      /usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1690 +0x226
  testing.(*T).Run.gowrap1()
      /usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1743 +0x44

Goroutine 113 (running) created at:
  law-oa-go/internal/concurrency.(*WorkerPool).Start()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/worker_pool.go:104 +0x117
  law-oa-go/internal/concurrency.(*ConcurrentService).Start()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/concurrent_service.go:222 +0x51
  law-oa-go/internal/concurrency.TestConcurrentService_SubmitAPITask()
      /Users/mac/Desktop/FT/law-oa-go/internal/concurrency/concurrent_service_test.go:192 +0x3d
  testing.tRunner()
      /usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1690 +0x226
  testing.(*T).Run.gowrap1()
      /usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1743 +0x44
==================
Task api_1757987318986777000 completed successfully
Task api_1757987318986777000 completed successfully
    testing.go:1399: race detected during execution of test
--- FAIL: TestConcurrentService_SubmitAPITask (0.01s)
=== RUN   TestConcurrentService_GetMetrics
Task metrics_task_1 completed successfully
Task metrics_task_2 completed successfully
Task metrics_task_0 completed successfully
--- PASS: TestConcurrentService_GetMetrics (0.20s)
=== RUN   TestConcurrentService_IsRunning
--- PASS: TestConcurrentService_IsRunning (0.00s)
=== RUN   TestConcurrentService_GetActiveTasksCount
Task long_task completed successfully
--- PASS: TestConcurrentService_GetActiveTasksCount (0.45s)
=== RUN   TestConcurrentService_UpdateConfig
--- PASS: TestConcurrentService_UpdateConfig (0.00s)
=== RUN   TestConcurrentService_GetConfig
--- PASS: TestConcurrentService_GetConfig (0.00s)
=== RUN   TestConcurrentService_ConcurrentTaskSubmission
Task concurrent_task_1 completed successfully
Task concurrent_task_0 completed successfully
Task concurrent_task_8 completed successfully
Task concurrent_task_9 completed successfully
Task concurrent_task_3 completed successfully
Task concurrent_task_6 completed successfully
Task concurrent_task_7 completed successfully
Task concurrent_task_5 completed successfully
Task concurrent_task_2 completed successfully
Task concurrent_task_4 completed successfully
--- PASS: TestConcurrentService_ConcurrentTaskSubmission (0.10s)
=== RUN   TestNewCircuitBreaker
--- PASS: TestNewCircuitBreaker (0.00s)
=== RUN   TestCircuitBreaker_ExecuteSuccess
--- PASS: TestCircuitBreaker_ExecuteSuccess (0.00s)
=== RUN   TestCircuitBreaker_ExecuteFailure
--- PASS: TestCircuitBreaker_ExecuteFailure (0.00s)
=== RUN   TestCircuitBreaker_TripToOpen
--- PASS: TestCircuitBreaker_TripToOpen (0.00s)
=== RUN   TestCircuitBreaker_ResetTimeout
--- PASS: TestCircuitBreaker_ResetTimeout (0.15s)
=== RUN   TestCircuitBreaker_HalfOpenFailure
--- PASS: TestCircuitBreaker_HalfOpenFailure (0.15s)
=== RUN   TestNewRateLimiter
--- PASS: TestNewRateLimiter (0.00s)
=== RUN   TestRateLimiter_Allow
--- PASS: TestRateLimiter_Allow (1.00s)
=== RUN   TestRateLimiter_Wait
--- PASS: TestRateLimiter_Wait (1.01s)
=== RUN   TestRateLimiter_WaitContextTimeout
--- PASS: TestRateLimiter_WaitContextTimeout (0.10s)
=== RUN   TestRateLimiter_TokenRefill
--- PASS: TestRateLimiter_TokenRefill (1.00s)
=== RUN   TestNewConcurrentSafe
--- PASS: TestNewConcurrentSafe (0.00s)
=== RUN   TestConcurrentSafe_ExecuteDisabled
--- PASS: TestConcurrentSafe_ExecuteDisabled (0.00s)
=== RUN   TestConcurrentSafe_ExecuteWithCircuitBreaker
--- PASS: TestConcurrentSafe_ExecuteWithCircuitBreaker (0.00s)
=== RUN   TestConcurrentSafe_ExecuteWithRateLimiter
--- PASS: TestConcurrentSafe_ExecuteWithRateLimiter (0.10s)
=== RUN   TestConcurrentSafe_ExecuteWithBoth
--- PASS: TestConcurrentSafe_ExecuteWithBoth (0.00s)
=== RUN   TestConcurrentSafe_ConcurrentExecution
--- PASS: TestConcurrentSafe_ConcurrentExecution (1.01s)
=== RUN   TestConcurrentSafe_GetStatusMethods
--- PASS: TestConcurrentSafe_GetStatusMethods (0.00s)
=== RUN   TestNewWorkerPool
--- PASS: TestNewWorkerPool (0.00s)
=== RUN   TestWorkerPool_StartStop
--- PASS: TestWorkerPool_StartStop (0.00s)
=== RUN   TestWorkerPool_SubmitTask
Task test_task completed successfully
--- PASS: TestWorkerPool_SubmitTask (0.20s)
=== RUN   TestWorkerPool_SubmitWithResult
Task test_result_task completed successfully
--- PASS: TestWorkerPool_SubmitWithResult (0.11s)
=== RUN   TestWorkerPool_ConcurrentTasks
Task concurrent_task_0 completed successfully
Task concurrent_task_4 completed successfully
Task concurrent_task_2 completed successfully
Task concurrent_task_3 completed successfully
Task concurrent_task_1 completed successfully
Task concurrent_task_5 completed successfully
Task concurrent_task_6 completed successfully
Task concurrent_task_9 completed successfully
Task concurrent_task_7 completed successfully
Task concurrent_task_8 completed successfully
--- PASS: TestWorkerPool_ConcurrentTasks (0.20s)
=== RUN   TestWorkerPool_TaskErrorHandling
Task error_task failed: task failed intentionally
--- PASS: TestWorkerPool_TaskErrorHandling (1.21s)
=== RUN   TestWorkerPool_ContextCancellation
--- PASS: TestWorkerPool_ContextCancellation (0.50s)
=== RUN   TestWorkerPool_RetryMechanism
Task retry_task completed successfully
--- PASS: TestWorkerPool_RetryMechanism (0.05s)
=== RUN   TestWorkerPool_Metrics
Task metrics_task_1 completed successfully
Task metrics_task_2 completed successfully
Task metrics_task_0 completed successfully
Task metrics_task_4 completed successfully
Task metrics_task_3 completed successfully
--- PASS: TestWorkerPool_Metrics (0.20s)
=== RUN   TestWorkerPool_FullQueue
Task queue_task_1 completed successfully
--- PASS: TestWorkerPool_FullQueue (0.20s)
FAIL
FAIL	law-oa-go/internal/concurrency	9.098s
=== RUN   TestConfig_Load_DefaultConfig
=== RUN   TestConfig_Load_DefaultConfig/加载默认配置
Warning: .env file not found, using environment variables
--- PASS: TestConfig_Load_DefaultConfig (0.02s)
    --- PASS: TestConfig_Load_DefaultConfig/加载默认配置 (0.02s)
=== RUN   TestConfig_Load_EnvironmentVariables
=== RUN   TestConfig_Load_EnvironmentVariables/从环境变量加载配置
Warning: .env file not found, using environment variables
--- PASS: TestConfig_Load_EnvironmentVariables (0.00s)
    --- PASS: TestConfig_Load_EnvironmentVariables/从环境变量加载配置 (0.00s)
=== RUN   TestConfig_Load_MissingJWTSecret
=== RUN   TestConfig_Load_MissingJWTSecret/缺少JWT密钥
Warning: .env file not found, using environment variables
--- PASS: TestConfig_Load_MissingJWTSecret (0.00s)
    --- PASS: TestConfig_Load_MissingJWTSecret/缺少JWT密钥 (0.00s)
=== RUN   TestConfig_Load_ShortJWTSecret
=== RUN   TestConfig_Load_ShortJWTSecret/JWT密钥太短
Warning: .env file not found, using environment variables
--- PASS: TestConfig_Load_ShortJWTSecret (0.00s)
    --- PASS: TestConfig_Load_ShortJWTSecret/JWT密钥太短 (0.00s)
=== RUN   TestConfig_Load_IncompleteDatabaseConfig
=== RUN   TestConfig_Load_IncompleteDatabaseConfig/数据库配置不完整
Warning: .env file not found, using environment variables
--- PASS: TestConfig_Load_IncompleteDatabaseConfig (0.01s)
    --- PASS: TestConfig_Load_IncompleteDatabaseConfig/数据库配置不完整 (0.01s)
=== RUN   TestConfig_GetDatabaseDSN
=== RUN   TestConfig_GetDatabaseDSN/获取数据库DSN
--- PASS: TestConfig_GetDatabaseDSN (0.00s)
    --- PASS: TestConfig_GetDatabaseDSN/获取数据库DSN (0.00s)
=== RUN   TestConfig_GetRedisAddr
=== RUN   TestConfig_GetRedisAddr/获取Redis地址
--- PASS: TestConfig_GetRedisAddr (0.00s)
    --- PASS: TestConfig_GetRedisAddr/获取Redis地址 (0.00s)
=== RUN   TestConfig_GetElasticsearchURL
=== RUN   TestConfig_GetElasticsearchURL/获取Elasticsearch_URL
--- PASS: TestConfig_GetElasticsearchURL (0.00s)
    --- PASS: TestConfig_GetElasticsearchURL/获取Elasticsearch_URL (0.00s)
=== RUN   TestConfig_IsProduction
=== RUN   TestConfig_IsProduction/判断生产环境
=== RUN   TestConfig_IsProduction/判断生产环境/生产环境
=== RUN   TestConfig_IsProduction/判断生产环境/开发环境
=== RUN   TestConfig_IsProduction/判断生产环境/测试环境
=== RUN   TestConfig_IsProduction/判断生产环境/未知环境
--- PASS: TestConfig_IsProduction (0.00s)
    --- PASS: TestConfig_IsProduction/判断生产环境 (0.00s)
        --- PASS: TestConfig_IsProduction/判断生产环境/生产环境 (0.00s)
        --- PASS: TestConfig_IsProduction/判断生产环境/开发环境 (0.00s)
        --- PASS: TestConfig_IsProduction/判断生产环境/测试环境 (0.00s)
        --- PASS: TestConfig_IsProduction/判断生产环境/未知环境 (0.00s)
=== RUN   TestConfig_IsDevelopment
=== RUN   TestConfig_IsDevelopment/判断开发环境
=== RUN   TestConfig_IsDevelopment/判断开发环境/开发环境
=== RUN   TestConfig_IsDevelopment/判断开发环境/生产环境
=== RUN   TestConfig_IsDevelopment/判断开发环境/测试环境
=== RUN   TestConfig_IsDevelopment/判断开发环境/未知环境
--- PASS: TestConfig_IsDevelopment (0.00s)
    --- PASS: TestConfig_IsDevelopment/判断开发环境 (0.00s)
        --- PASS: TestConfig_IsDevelopment/判断开发环境/开发环境 (0.00s)
        --- PASS: TestConfig_IsDevelopment/判断开发环境/生产环境 (0.00s)
        --- PASS: TestConfig_IsDevelopment/判断开发环境/测试环境 (0.00s)
        --- PASS: TestConfig_IsDevelopment/判断开发环境/未知环境 (0.00s)
=== RUN   TestConfig_GetPort
=== RUN   TestConfig_GetPort/获取端口
=== RUN   TestConfig_GetPort/获取端口/指定端口
=== RUN   TestConfig_GetPort/获取端口/空端口
=== RUN   TestConfig_GetPort/获取端口/默认端口
--- PASS: TestConfig_GetPort (0.00s)
    --- PASS: TestConfig_GetPort/获取端口 (0.00s)
        --- PASS: TestConfig_GetPort/获取端口/指定端口 (0.00s)
        --- PASS: TestConfig_GetPort/获取端口/空端口 (0.00s)
        --- PASS: TestConfig_GetPort/获取端口/默认端口 (0.00s)
=== RUN   TestConfig_Load_WithMockFactory
=== RUN   TestConfig_Load_WithMockFactory/使用Mock工厂测试配置加载
Warning: .env file not found, using environment variables
    config_test.go:393: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/config/config_test.go:393
        	Error:      	Not equal: 
        	            	expected: "test"
        	            	actual  : "dev"
        	            	
        	            	Diff:
        	            	--- Expected
        	            	+++ Actual
        	            	@@ -1 +1 @@
        	            	-test
        	            	+dev
        	Test:       	TestConfig_Load_WithMockFactory/使用Mock工厂测试配置加载
        	Messages:   	环境应该为test
    config_test.go:394: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/config/config_test.go:394
        	Error:      	Not equal: 
        	            	expected: "9090"
        	            	actual  : "8080"
        	            	
        	            	Diff:
        	            	--- Expected
        	            	+++ Actual
        	            	@@ -1 +1 @@
        	            	-9090
        	            	+8080
        	Test:       	TestConfig_Load_WithMockFactory/使用Mock工厂测试配置加载
        	Messages:   	端口应该为9090
    config_test.go:395: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/config/config_test.go:395
        	Error:      	Not equal: 
        	            	expected: "mock-host"
        	            	actual  : "localhost"
        	            	
        	            	Diff:
        	            	--- Expected
        	            	+++ Actual
        	            	@@ -1 +1 @@
        	            	-mock-host
        	            	+localhost
        	Test:       	TestConfig_Load_WithMockFactory/使用Mock工厂测试配置加载
        	Messages:   	数据库主机应该为mock-host
    config_test.go:396: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/config/config_test.go:396
        	Error:      	Not equal: 
        	            	expected: "mock-redis"
        	            	actual  : "localhost"
        	            	
        	            	Diff:
        	            	--- Expected
        	            	+++ Actual
        	            	@@ -1 +1 @@
        	            	-mock-redis
        	            	+localhost
        	Test:       	TestConfig_Load_WithMockFactory/使用Mock工厂测试配置加载
        	Messages:   	Redis主机应该为mock-redis
    config_test.go:397: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/config/config_test.go:397
        	Error:      	Not equal: 
        	            	expected: "mock-jwt-secret-that-is-at-least-32-characters-long"
        	            	actual  : "your-secret-key-change-this-in-production"
        	            	
        	            	Diff:
        	            	--- Expected
        	            	+++ Actual
        	            	@@ -1 +1 @@
        	            	-mock-jwt-secret-that-is-at-least-32-characters-long
        	            	+your-secret-key-change-this-in-production
        	Test:       	TestConfig_Load_WithMockFactory/使用Mock工厂测试配置加载
        	Messages:   	JWT密钥应该正确
--- FAIL: TestConfig_Load_WithMockFactory (0.02s)
    --- FAIL: TestConfig_Load_WithMockFactory/使用Mock工厂测试配置加载 (0.02s)
FAIL
FAIL	law-oa-go/internal/config	0.677s
?   	law-oa-go/internal/database	[no test files]
?   	law-oa-go/internal/infrastructure	[no test files]
?   	law-oa-go/internal/logger	[no test files]
?   	law-oa-go/internal/logging	[no test files]
?   	law-oa-go/internal/models	[no test files]
=== RUN   TestBaseError
--- PASS: TestBaseError (0.00s)
=== RUN   TestBusinessError
--- PASS: TestBusinessError (0.00s)
=== RUN   TestValidationError
--- PASS: TestValidationError (0.00s)
=== RUN   TestDatabaseError
--- PASS: TestDatabaseError (0.00s)
=== RUN   TestAuthorizationError
--- PASS: TestAuthorizationError (0.00s)
=== RUN   TestConcurrencyError
--- PASS: TestConcurrencyError (0.00s)
=== RUN   TestNetworkError
--- PASS: TestNetworkError (0.00s)
=== RUN   TestPanicError
--- PASS: TestPanicError (0.00s)
=== RUN   TestContextManagement
--- PASS: TestContextManagement (0.00s)
=== RUN   TestErrorTypeChecking
--- PASS: TestErrorTypeChecking (0.00s)
=== RUN   TestErrorUtilityFunctions
--- PASS: TestErrorUtilityFunctions (0.00s)
=== RUN   TestHelperFunctions
--- PASS: TestHelperFunctions (0.00s)
=== RUN   TestErrorWrapping
--- PASS: TestErrorWrapping (0.00s)
=== RUN   TestErrorSeverity
--- PASS: TestErrorSeverity (0.00s)
PASS
ok  	law-oa-go/internal/errors	1.576s
# law-oa-go/internal/repositories.test
ld: warning: '/private/var/folders/4p/bng36r_s65d26yqk0lfpw2rh0000gn/T/go-link-963396470/000023.o' has malformed LC_DYSYMTAB, expected 98 undefined symbols to start at index 884, found 95 undefined symbols starting at index 884
=== RUN   TestAuthHandler_Login
=== RUN   TestAuthHandler_Login/Login_Success
    assertions.go:120: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:66
        	Error:      	Expected value not to be nil.
        	Test:       	TestAuthHandler_Login/Login_Success
        	Messages:   	Expected success response - response should not be nil
    auth_handler_test.go:69: FAIL:	FindByID(string,uint)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:49]
    auth_handler_test.go:69: FAIL: 1 out of 2 expectation(s) were met.
        	The code you are testing needs to make 1 more call(s).
        	at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:69]
=== RUN   TestAuthHandler_Login/Login_Invalid_Credentials
    auth_handler_test.go:90: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:90
        	Error:      	Not equal: 
        	            	expected: 404
        	            	actual  : 200
        	Test:       	TestAuthHandler_Login/Login_Invalid_Credentials
    auth_handler_test.go:93: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:93
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestAuthHandler_Login/Login_Invalid_Credentials
    auth_handler_test.go:94: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:94
        	Error:      	map[string]interface {}(nil) does not contain "error"
        	Test:       	TestAuthHandler_Login/Login_Invalid_Credentials
    auth_handler_test.go:97: FAIL:	FindByID(string,uint)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:49]
    auth_handler_test.go:97: FAIL: 2 out of 3 expectation(s) were met.
        	The code you are testing needs to make 1 more call(s).
        	at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:97]
=== RUN   TestAuthHandler_Login/Login_Invalid_Request
    auth_handler_test.go:114: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:114
        	Error:      	Not equal: 
        	            	expected: 400
        	            	actual  : 200
        	Test:       	TestAuthHandler_Login/Login_Invalid_Request
    auth_handler_test.go:117: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:117
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestAuthHandler_Login/Login_Invalid_Request
    auth_handler_test.go:118: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:118
        	Error:      	map[string]interface {}(nil) does not contain "error"
        	Test:       	TestAuthHandler_Login/Login_Invalid_Request
=== RUN   TestAuthHandler_Login/Login_Inactive_User
    auth_handler_test.go:151: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:151
        	Error:      	Not equal: 
        	            	expected: 404
        	            	actual  : 200
        	Test:       	TestAuthHandler_Login/Login_Inactive_User
    auth_handler_test.go:154: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:154
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestAuthHandler_Login/Login_Inactive_User
    auth_handler_test.go:155: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:155
        	Error:      	map[string]interface {}(nil) does not contain "error"
        	Test:       	TestAuthHandler_Login/Login_Inactive_User
    auth_handler_test.go:158: FAIL:	FindByID(string,uint)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:49]
    auth_handler_test.go:158: FAIL: 3 out of 4 expectation(s) were met.
        	The code you are testing needs to make 1 more call(s).
        	at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:158]
--- FAIL: TestAuthHandler_Login (0.02s)
    --- FAIL: TestAuthHandler_Login/Login_Success (0.01s)
    --- FAIL: TestAuthHandler_Login/Login_Invalid_Credentials (0.00s)
    --- FAIL: TestAuthHandler_Login/Login_Invalid_Request (0.00s)
    --- FAIL: TestAuthHandler_Login/Login_Inactive_User (0.00s)
=== RUN   TestAuthHandler_Register
=== RUN   TestAuthHandler_Register/Register_Success
    auth_handler_test.go:211: FAIL:	FindByID(string,uint)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:188]
    auth_handler_test.go:211: FAIL: 2 out of 3 expectation(s) were met.
        	The code you are testing needs to make 1 more call(s).
        	at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:211]
=== RUN   TestAuthHandler_Register/Register_Email_Already_Exists
    auth_handler_test.go:240: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:240
        	Error:      	Not equal: 
        	            	expected: 409
        	            	actual  : 200
        	Test:       	TestAuthHandler_Register/Register_Email_Already_Exists
    auth_handler_test.go:243: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:243
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestAuthHandler_Register/Register_Email_Already_Exists
    auth_handler_test.go:244: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:244
        	Error:      	map[string]interface {}(nil) does not contain "error"
        	Test:       	TestAuthHandler_Register/Register_Email_Already_Exists
    auth_handler_test.go:247: FAIL:	FindByID(string,uint)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:188]
    auth_handler_test.go:247: FAIL: 3 out of 4 expectation(s) were met.
        	The code you are testing needs to make 1 more call(s).
        	at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:247]
=== RUN   TestAuthHandler_Register/Register_Invalid_Data
    auth_handler_test.go:265: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:265
        	Error:      	Not equal: 
        	            	expected: 400
        	            	actual  : 200
        	Test:       	TestAuthHandler_Register/Register_Invalid_Data
    auth_handler_test.go:268: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:268
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestAuthHandler_Register/Register_Invalid_Data
    auth_handler_test.go:269: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:269
        	Error:      	map[string]interface {}(nil) does not contain "error"
        	Test:       	TestAuthHandler_Register/Register_Invalid_Data
=== RUN   TestAuthHandler_Register/Register_Invalid_Role
    auth_handler_test.go:292: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:292
        	Error:      	Not equal: 
        	            	expected: 400
        	            	actual  : 200
        	Test:       	TestAuthHandler_Register/Register_Invalid_Role
    auth_handler_test.go:295: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:295
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestAuthHandler_Register/Register_Invalid_Role
    auth_handler_test.go:296: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:296
        	Error:      	map[string]interface {}(nil) does not contain "error"
        	Test:       	TestAuthHandler_Register/Register_Invalid_Role
    auth_handler_test.go:299: FAIL:	FindByID(string,uint)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:188]
    auth_handler_test.go:299: FAIL: 4 out of 5 expectation(s) were met.
        	The code you are testing needs to make 1 more call(s).
        	at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:299]
--- FAIL: TestAuthHandler_Register (1.10s)
    --- FAIL: TestAuthHandler_Register/Register_Success (1.10s)
    --- FAIL: TestAuthHandler_Register/Register_Email_Already_Exists (0.00s)
    --- FAIL: TestAuthHandler_Register/Register_Invalid_Data (0.00s)
    --- FAIL: TestAuthHandler_Register/Register_Invalid_Role (0.00s)
=== RUN   TestAuthHandler_GetProfile
=== RUN   TestAuthHandler_GetProfile/Get_Profile_Success
    assertions.go:120: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:339
        	Error:      	Expected value not to be nil.
        	Test:       	TestAuthHandler_GetProfile/Get_Profile_Success
        	Messages:   	Expected success response - response should not be nil
    auth_handler_test.go:342: FAIL:	FindByID(string,uint)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:327]
    auth_handler_test.go:342: FAIL: 0 out of 1 expectation(s) were met.
        	The code you are testing needs to make 1 more call(s).
        	at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:342]
=== RUN   TestAuthHandler_GetProfile/Get_Profile_Unauthorized
    auth_handler_test.go:354: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:354
        	Error:      	Not equal: 
        	            	expected: 401
        	            	actual  : 200
        	Test:       	TestAuthHandler_GetProfile/Get_Profile_Unauthorized
    auth_handler_test.go:357: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:357
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestAuthHandler_GetProfile/Get_Profile_Unauthorized
    auth_handler_test.go:358: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:358
        	Error:      	map[string]interface {}(nil) does not contain "error"
        	Test:       	TestAuthHandler_GetProfile/Get_Profile_Unauthorized
=== RUN   TestAuthHandler_GetProfile/Get_Profile_User_Not_Found
    auth_handler_test.go:374: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:374
        	Error:      	Not equal: 
        	            	expected: 404
        	            	actual  : 200
        	Test:       	TestAuthHandler_GetProfile/Get_Profile_User_Not_Found
    auth_handler_test.go:377: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:377
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestAuthHandler_GetProfile/Get_Profile_User_Not_Found
    auth_handler_test.go:378: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:378
        	Error:      	map[string]interface {}(nil) does not contain "error"
        	Test:       	TestAuthHandler_GetProfile/Get_Profile_User_Not_Found
    auth_handler_test.go:381: FAIL:	FindByID(string,uint)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:327]
    auth_handler_test.go:381: FAIL:	FindByID(string,uint)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:363]
    auth_handler_test.go:381: FAIL: 0 out of 2 expectation(s) were met.
        	The code you are testing needs to make 2 more call(s).
        	at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:381]
--- FAIL: TestAuthHandler_GetProfile (0.00s)
    --- FAIL: TestAuthHandler_GetProfile/Get_Profile_Success (0.00s)
    --- FAIL: TestAuthHandler_GetProfile/Get_Profile_Unauthorized (0.00s)
    --- FAIL: TestAuthHandler_GetProfile/Get_Profile_User_Not_Found (0.00s)
=== RUN   TestAuthHandler_UpdateProfile
=== RUN   TestAuthHandler_UpdateProfile/Update_Profile_Success
    assertions.go:120: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:431
        	Error:      	Expected value not to be nil.
        	Test:       	TestAuthHandler_UpdateProfile/Update_Profile_Success
        	Messages:   	Expected success response - response should not be nil
    auth_handler_test.go:434: FAIL:	FindByID(string,uint)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:409]
    auth_handler_test.go:434: FAIL:	FindByEmail(string,string)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:410]
    auth_handler_test.go:434: FAIL:	Update(string,mock.anythingOfTypeArgument)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:411]
    auth_handler_test.go:434: FAIL:	FindByID(string,uint)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:412]
    auth_handler_test.go:434: FAIL: 0 out of 4 expectation(s) were met.
        	The code you are testing needs to make 4 more call(s).
        	at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:434]
=== RUN   TestAuthHandler_UpdateProfile/Update_Profile_Invalid_Data
    auth_handler_test.go:453: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:453
        	Error:      	Not equal: 
        	            	expected: 400
        	            	actual  : 200
        	Test:       	TestAuthHandler_UpdateProfile/Update_Profile_Invalid_Data
    auth_handler_test.go:456: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:456
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestAuthHandler_UpdateProfile/Update_Profile_Invalid_Data
    auth_handler_test.go:457: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:457
        	Error:      	map[string]interface {}(nil) does not contain "error"
        	Test:       	TestAuthHandler_UpdateProfile/Update_Profile_Invalid_Data
    auth_handler_test.go:458: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:458
        	Error:      	Not equal: 
        	            	expected: float64(400)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestAuthHandler_UpdateProfile/Update_Profile_Invalid_Data
--- FAIL: TestAuthHandler_UpdateProfile (0.00s)
    --- FAIL: TestAuthHandler_UpdateProfile/Update_Profile_Success (0.00s)
    --- FAIL: TestAuthHandler_UpdateProfile/Update_Profile_Invalid_Data (0.00s)
=== RUN   TestAuthHandler_ChangePassword
=== RUN   TestAuthHandler_ChangePassword/Change_Password_Success
    assertions.go:120: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:505
        	Error:      	Expected value not to be nil.
        	Test:       	TestAuthHandler_ChangePassword/Change_Password_Success
        	Messages:   	Expected success response - response should not be nil
    auth_handler_test.go:508: FAIL:	FindByID(string,uint)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:486]
    auth_handler_test.go:508: FAIL:	Update(string,mock.anythingOfTypeArgument)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:487]
    auth_handler_test.go:508: FAIL: 0 out of 2 expectation(s) were met.
        	The code you are testing needs to make 2 more call(s).
        	at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:508]
=== RUN   TestAuthHandler_ChangePassword/Change_Password_Wrong_Current_Password
    auth_handler_test.go:542: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:542
        	Error:      	Not equal: 
        	            	expected: 400
        	            	actual  : 200
        	Test:       	TestAuthHandler_ChangePassword/Change_Password_Wrong_Current_Password
    auth_handler_test.go:545: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:545
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestAuthHandler_ChangePassword/Change_Password_Wrong_Current_Password
    auth_handler_test.go:546: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:546
        	Error:      	map[string]interface {}(nil) does not contain "error"
        	Test:       	TestAuthHandler_ChangePassword/Change_Password_Wrong_Current_Password
    auth_handler_test.go:547: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:547
        	Error:      	Not equal: 
        	            	expected: float64(400)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestAuthHandler_ChangePassword/Change_Password_Wrong_Current_Password
    auth_handler_test.go:550: FAIL:	FindByID(string,uint)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:486]
    auth_handler_test.go:550: FAIL:	Update(string,mock.anythingOfTypeArgument)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:487]
    auth_handler_test.go:550: FAIL:	FindByID(string,uint)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:525]
    auth_handler_test.go:550: FAIL: 0 out of 3 expectation(s) were met.
        	The code you are testing needs to make 3 more call(s).
        	at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:550]
=== RUN   TestAuthHandler_ChangePassword/Change_Password_Weak_New_Password
    auth_handler_test.go:584: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:584
        	Error:      	Not equal: 
        	            	expected: 400
        	            	actual  : 200
        	Test:       	TestAuthHandler_ChangePassword/Change_Password_Weak_New_Password
    auth_handler_test.go:587: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:587
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestAuthHandler_ChangePassword/Change_Password_Weak_New_Password
    auth_handler_test.go:588: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:588
        	Error:      	map[string]interface {}(nil) does not contain "error"
        	Test:       	TestAuthHandler_ChangePassword/Change_Password_Weak_New_Password
    auth_handler_test.go:589: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:589
        	Error:      	Not equal: 
        	            	expected: float64(400)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestAuthHandler_ChangePassword/Change_Password_Weak_New_Password
    auth_handler_test.go:592: FAIL:	FindByID(string,uint)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:486]
    auth_handler_test.go:592: FAIL:	Update(string,mock.anythingOfTypeArgument)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:487]
    auth_handler_test.go:592: FAIL:	FindByID(string,uint)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:525]
    auth_handler_test.go:592: FAIL:	FindByID(string,uint)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:567]
    auth_handler_test.go:592: FAIL: 0 out of 4 expectation(s) were met.
        	The code you are testing needs to make 4 more call(s).
        	at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:592]
--- FAIL: TestAuthHandler_ChangePassword (0.00s)
    --- FAIL: TestAuthHandler_ChangePassword/Change_Password_Success (0.00s)
    --- FAIL: TestAuthHandler_ChangePassword/Change_Password_Wrong_Current_Password (0.00s)
    --- FAIL: TestAuthHandler_ChangePassword/Change_Password_Weak_New_Password (0.00s)
=== RUN   TestAuthHandler_RefreshToken
=== RUN   TestAuthHandler_RefreshToken/Refresh_Token_Success
    assertions.go:120: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:635
        	Error:      	Expected value not to be nil.
        	Test:       	TestAuthHandler_RefreshToken/Refresh_Token_Success
        	Messages:   	Expected success response - response should not be nil
    auth_handler_test.go:638: FAIL:	FindByID(string,uint)
        		at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:619]
    auth_handler_test.go:638: FAIL: 0 out of 1 expectation(s) were met.
        	The code you are testing needs to make 1 more call(s).
        	at: [/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:638]
=== RUN   TestAuthHandler_RefreshToken/Refresh_Token_Invalid_Token
    auth_handler_test.go:655: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:655
        	Error:      	Not equal: 
        	            	expected: 401
        	            	actual  : 200
        	Test:       	TestAuthHandler_RefreshToken/Refresh_Token_Invalid_Token
    auth_handler_test.go:658: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:658
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestAuthHandler_RefreshToken/Refresh_Token_Invalid_Token
    auth_handler_test.go:659: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:659
        	Error:      	map[string]interface {}(nil) does not contain "error"
        	Test:       	TestAuthHandler_RefreshToken/Refresh_Token_Invalid_Token
    auth_handler_test.go:660: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/auth_handler_test.go:660
        	Error:      	Not equal: 
        	            	expected: float64(401)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestAuthHandler_RefreshToken/Refresh_Token_Invalid_Token
--- FAIL: TestAuthHandler_RefreshToken (0.00s)
    --- FAIL: TestAuthHandler_RefreshToken/Refresh_Token_Success (0.00s)
    --- FAIL: TestAuthHandler_RefreshToken/Refresh_Token_Invalid_Token (0.00s)
=== RUN   TestAuthHandler_Logout
=== RUN   TestAuthHandler_Logout/Logout_Success
--- PASS: TestAuthHandler_Logout (0.00s)
    --- PASS: TestAuthHandler_Logout/Logout_Success (0.00s)
=== RUN   TestClientHandler_GetClient
=== RUN   TestClientHandler_GetClient/Get_Client_Success
=== RUN   TestClientHandler_GetClient/Get_Client_Not_Found
--- PASS: TestClientHandler_GetClient (0.00s)
    --- PASS: TestClientHandler_GetClient/Get_Client_Success (0.00s)
    --- PASS: TestClientHandler_GetClient/Get_Client_Not_Found (0.00s)
=== RUN   TestDocumentHandler_UploadDocumentSuccess
--- PASS: TestDocumentHandler_UploadDocumentSuccess (0.01s)
=== RUN   TestDocumentHandler_GetDocumentSuccess
--- PASS: TestDocumentHandler_GetDocumentSuccess (0.00s)
=== RUN   TestDocumentHandler_GetDocumentInvalidID
    document_handler_test.go:390: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/document_handler_test.go:390
        	Error:      	Not equal: 
        	            	expected: 400
        	            	actual  : 200
        	Test:       	TestDocumentHandler_GetDocumentInvalidID
--- FAIL: TestDocumentHandler_GetDocumentInvalidID (0.00s)
=== RUN   TestDocumentHandler_GetDocumentNotFound
    document_handler_test.go:415: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/document_handler_test.go:415
        	Error:      	Not equal: 
        	            	expected: 404
        	            	actual  : 200
        	Test:       	TestDocumentHandler_GetDocumentNotFound
--- FAIL: TestDocumentHandler_GetDocumentNotFound (0.00s)
=== RUN   TestDocumentHandler_UpdateDocumentSuccess
--- PASS: TestDocumentHandler_UpdateDocumentSuccess (0.00s)
=== RUN   TestDocumentHandler_DeleteDocumentSuccess
--- PASS: TestDocumentHandler_DeleteDocumentSuccess (0.00s)
=== RUN   TestDocumentHandler_ListDocumentsSuccess
--- PASS: TestDocumentHandler_ListDocumentsSuccess (0.00s)
=== RUN   TestDocumentHandler_GetDocumentStatsSuccess
--- PASS: TestDocumentHandler_GetDocumentStatsSuccess (0.00s)
=== RUN   TestDocumentHandler_DownloadDocumentSuccess
--- PASS: TestDocumentHandler_DownloadDocumentSuccess (0.00s)
=== RUN   TestDocumentHandler_UploadDocumentMissingFile
    document_handler_test.go:612: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/document_handler_test.go:612
        	Error:      	Not equal: 
        	            	expected: 400
        	            	actual  : 200
        	Test:       	TestDocumentHandler_UploadDocumentMissingFile
--- FAIL: TestDocumentHandler_UploadDocumentMissingFile (0.00s)
=== RUN   TestDocumentHandler_UploadDocumentInvalidEntityID
    document_handler_test.go:647: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/document_handler_test.go:647
        	Error:      	Not equal: 
        	            	expected: 400
        	            	actual  : 200
        	Test:       	TestDocumentHandler_UploadDocumentInvalidEntityID
--- FAIL: TestDocumentHandler_UploadDocumentInvalidEntityID (0.00s)
=== RUN   TestDocumentHandler_UpdateDocumentInvalidJSON
    document_handler_test.go:669: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/document_handler_test.go:669
        	Error:      	Not equal: 
        	            	expected: 400
        	            	actual  : 200
        	Test:       	TestDocumentHandler_UpdateDocumentInvalidJSON
--- FAIL: TestDocumentHandler_UpdateDocumentInvalidJSON (0.00s)
=== RUN   TestDocumentHandler_ListDocumentsInvalidQuery
    document_handler_test.go:689: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/document_handler_test.go:689
        	Error:      	Not equal: 
        	            	expected: 400
        	            	actual  : 200
        	Test:       	TestDocumentHandler_ListDocumentsInvalidQuery
--- FAIL: TestDocumentHandler_ListDocumentsInvalidQuery (0.00s)
=== RUN   TestDocumentHandler_DownloadDocumentInvalidID
    document_handler_test.go:709: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/document_handler_test.go:709
        	Error:      	Not equal: 
        	            	expected: 400
        	            	actual  : 200
        	Test:       	TestDocumentHandler_DownloadDocumentInvalidID
--- FAIL: TestDocumentHandler_DownloadDocumentInvalidID (0.00s)
=== RUN   TestDocumentHandler_ServiceError
    document_handler_test.go:728: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/handlers/document_handler_test.go:728
        	Error:      	Not equal: 
        	            	expected: 500
        	            	actual  : 200
        	Test:       	TestDocumentHandler_ServiceError
--- FAIL: TestDocumentHandler_ServiceError (0.00s)
=== RUN   TestDocumentHandler_BoundaryCases
=== RUN   TestDocumentHandler_BoundaryCases/LargeEntityID
=== RUN   TestDocumentHandler_BoundaryCases/EmptyFormFields
=== RUN   TestDocumentHandler_BoundaryCases/SpecialCharactersInFormFields
--- PASS: TestDocumentHandler_BoundaryCases (0.00s)
    --- PASS: TestDocumentHandler_BoundaryCases/LargeEntityID (0.00s)
    --- PASS: TestDocumentHandler_BoundaryCases/EmptyFormFields (0.00s)
    --- PASS: TestDocumentHandler_BoundaryCases/SpecialCharactersInFormFields (0.00s)
=== RUN   TestDocumentHandler_Performance
--- PASS: TestDocumentHandler_Performance (0.00s)
=== RUN   TestUserHandler_GetUser
=== RUN   TestUserHandler_GetUser/Get_User_Success
=== RUN   TestUserHandler_GetUser/Get_User_Not_Found
--- PASS: TestUserHandler_GetUser (0.00s)
    --- PASS: TestUserHandler_GetUser/Get_User_Success (0.00s)
    --- PASS: TestUserHandler_GetUser/Get_User_Not_Found (0.00s)
FAIL
FAIL	law-oa-go/internal/handlers	2.028s
=== RUN   TestHealthCheckConfigurator_Basic
--- PASS: TestHealthCheckConfigurator_Basic (0.00s)
=== RUN   TestHealthCheckConfigurator_ConfigureDatabase
--- PASS: TestHealthCheckConfigurator_ConfigureDatabase (0.00s)
=== RUN   TestHealthCheckConfigurator_ConfigureCache
--- PASS: TestHealthCheckConfigurator_ConfigureCache (0.00s)
=== RUN   TestHealthCheckConfigurator_ConfigureConcurrency
--- PASS: TestHealthCheckConfigurator_ConfigureConcurrency (0.00s)
=== RUN   TestHealthCheckConfigurator_ConfigureExternalAPI
--- PASS: TestHealthCheckConfigurator_ConfigureExternalAPI (0.00s)
=== RUN   TestHealthCheckConfigurator_ConfigureStorage
--- PASS: TestHealthCheckConfigurator_ConfigureStorage (0.00s)
=== RUN   TestHealthCheckBuilder_Basic
--- PASS: TestHealthCheckBuilder_Basic (0.00s)
=== RUN   TestHealthCheckBuilder_WithConfig
--- PASS: TestHealthCheckBuilder_WithConfig (0.00s)
=== RUN   TestHealthCheckBuilder_WithDatabase
2025/09/16 09:49:11 INFO 注册健康检查 name=database
--- FAIL: TestHealthCheckBuilder_WithDatabase (0.00s)
panic: runtime error: invalid memory address or nil pointer dereference [recovered]
	panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x18 pc=0x3acbded]

goroutine 14 [running]:
testing.tRunner.func1.2({0x409da20, 0x4455700})
	/usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1632 +0x3fc
testing.tRunner.func1()
	/usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1635 +0x6b6
panic({0x409da20?, 0x4455700?})
	/usr/local/Cellar/go/1.23.6/libexec/src/runtime/panic.go:785 +0x132
database/sql.(*DB).conn(0xc00023ab60, {0x4128498, 0xc00046a7e0}, 0x1)
	/usr/local/Cellar/go/1.23.6/libexec/src/database/sql/sql.go:1423 +0xeed
database/sql.(*DB).PingContext.func1(0x1)
	/usr/local/Cellar/go/1.23.6/libexec/src/database/sql/sql.go:892 +0x66
database/sql.(*DB).retry(0x60?, 0xc0001fda20)
	/usr/local/Cellar/go/1.23.6/libexec/src/database/sql/sql.go:1568 +0x4b
database/sql.(*DB).PingContext(0xc00023ab60, {0x4128498, 0xc00046a7e0})
	/usr/local/Cellar/go/1.23.6/libexec/src/database/sql/sql.go:891 +0xf4
law-oa-go/internal/health.(*DatabaseHealthCheck).Check(0xc000010300, {0x4128498, 0xc00046a7e0})
	/Users/mac/Desktop/FT/law-oa-go/internal/health/health.go:313 +0x1e8
law-oa-go/internal/health.(*HealthChecker).RunChecks(0xc000023e50)
	/Users/mac/Desktop/FT/law-oa-go/internal/health/health.go:192 +0x24c
law-oa-go/internal/health.TestHealthCheckBuilder_WithDatabase(0xc00022e000)
	/Users/mac/Desktop/FT/law-oa-go/internal/health/config_test.go:157 +0x1f2
testing.tRunner(0xc00022e000, 0x4120038)
	/usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1690 +0x227
created by testing.(*T).Run in goroutine 1
	/usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1743 +0x826
FAIL	law-oa-go/internal/health	0.778s
=== RUN   TestMetricsBasicFunctionality
--- PASS: TestMetricsBasicFunctionality (0.01s)
=== RUN   TestPerformanceMonitorBasic
--- PASS: TestPerformanceMonitorBasic (0.01s)
=== RUN   TestBusinessMonitorBasic
--- PASS: TestBusinessMonitorBasic (0.02s)
=== RUN   TestMonitorServiceBasic
--- FAIL: TestMonitorServiceBasic (0.00s)
panic: duplicate metrics collector registration attempted [recovered]
	panic: duplicate metrics collector registration attempted

goroutine 21 [running]:
testing.tRunner.func1.2({0x5e0e4a0, 0xc000078d60})
	/usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1632 +0x3fc
testing.tRunner.func1()
	/usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1635 +0x6b6
panic({0x5e0e4a0?, 0xc000078d60?})
	/usr/local/Cellar/go/1.23.6/libexec/src/runtime/panic.go:785 +0x132
github.com/prometheus/client_golang/prometheus.(*Registry).MustRegister(0x60d7500, {0xc0000294d0, 0x1, 0x0?})
	/Users/mac/go/pkg/mod/github.com/prometheus/client_golang@v1.16.0/prometheus/registry.go:405 +0xa5
github.com/prometheus/client_golang/prometheus/promauto.Factory.NewGauge({{0x5e5f998?, 0x60d7500?}}, {{0x0, 0x0}, {0x0, 0x0}, {0x5d030bf, 0x12}, {0x5d07ff8, 0x1d}, ...})
	/Users/mac/go/pkg/mod/github.com/prometheus/client_golang@v1.16.0/prometheus/promauto/auto.go:297 +0x19d
github.com/prometheus/client_golang/prometheus/promauto.NewGauge(...)
	/Users/mac/go/pkg/mod/github.com/prometheus/client_golang@v1.16.0/prometheus/promauto/auto.go:191
law-oa-go/internal/metrics.NewApplicationMetrics()
	/Users/mac/Desktop/FT/law-oa-go/internal/metrics/application_metrics.go:87 +0x109
law-oa-go/internal/metrics.NewMonitorService.GetDefaultMetrics.func1()
	/Users/mac/Desktop/FT/law-oa-go/internal/metrics/application_metrics.go:76 +0x1d
sync.(*Once).doSlow(0x60f8b70, 0x5e59a48)
	/usr/local/Cellar/go/1.23.6/libexec/src/sync/once.go:76 +0xe2
sync.(*Once).Do(0x60f8b70, 0x5e59a48)
	/usr/local/Cellar/go/1.23.6/libexec/src/sync/once.go:67 +0x45
law-oa-go/internal/metrics.GetDefaultMetrics(...)
	/Users/mac/Desktop/FT/law-oa-go/internal/metrics/application_metrics.go:75
law-oa-go/internal/metrics.NewMonitorService({0x1, 0x1, 0x1, 0x6fc23ac00, 0x6fc23ac00, 0xdf8475800, 0x1})
	/Users/mac/Desktop/FT/law-oa-go/internal/metrics/monitor_service.go:83 +0x2c5
law-oa-go/internal/metrics.TestMonitorServiceBasic(0xc000299040)
	/Users/mac/Desktop/FT/law-oa-go/internal/metrics/metrics_basic_test.go:86 +0xe5
testing.tRunner(0xc000299040, 0x5e59930)
	/usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1690 +0x227
created by testing.(*T).Run in goroutine 1
	/usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1743 +0x826
FAIL	law-oa-go/internal/metrics	0.625s
=== RUN   TestErrorHandlingMiddleware_PanicError
    error_handler_test.go:56: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/middleware/error_handler_test.go:56
        	Error:      	Not equal: 
        	            	expected: "test-request-id"
        	            	actual  : ""
        	            	
        	            	Diff:
        	            	--- Expected
        	            	+++ Actual
        	            	@@ -1 +1 @@
        	            	-test-request-id
        	            	+
        	Test:       	TestErrorHandlingMiddleware_PanicError
--- FAIL: TestErrorHandlingMiddleware_PanicError (0.02s)
=== RUN   TestErrorHandlingMiddleware_BusinessError
    error_handler_test.go:97: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/middleware/error_handler_test.go:97
        	Error:      	Not equal: 
        	            	expected: "test-request-id"
        	            	actual  : ""
        	            	
        	            	Diff:
        	            	--- Expected
        	            	+++ Actual
        	            	@@ -1 +1 @@
        	            	-test-request-id
        	            	+
        	Test:       	TestErrorHandlingMiddleware_BusinessError
--- FAIL: TestErrorHandlingMiddleware_BusinessError (0.00s)
=== RUN   TestErrorHandlingMiddleware_ValidationError
    error_handler_test.go:130: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/middleware/error_handler_test.go:130
        	Error:      	Not equal: 
        	            	expected: "VALIDATION_ERROR"
        	            	actual  : "invalid_email_format"
        	            	
        	            	Diff:
        	            	--- Expected
        	            	+++ Actual
        	            	@@ -1 +1 @@
        	            	-VALIDATION_ERROR
        	            	+invalid_email_format
        	Test:       	TestErrorHandlingMiddleware_ValidationError
--- FAIL: TestErrorHandlingMiddleware_ValidationError (0.00s)
=== RUN   TestErrorHandlingMiddleware_AuthorizationError
    error_handler_test.go:164: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/middleware/error_handler_test.go:164
        	Error:      	Not equal: 
        	            	expected: "AUTHORIZATION_ERROR"
        	            	actual  : "authorization_error"
        	            	
        	            	Diff:
        	            	--- Expected
        	            	+++ Actual
        	            	@@ -1 +1 @@
        	            	-AUTHORIZATION_ERROR
        	            	+authorization_error
        	Test:       	TestErrorHandlingMiddleware_AuthorizationError
--- FAIL: TestErrorHandlingMiddleware_AuthorizationError (0.00s)
=== RUN   TestErrorHandlingMiddleware_MultipleErrors
--- PASS: TestErrorHandlingMiddleware_MultipleErrors (0.00s)
=== RUN   TestErrorHandlingMiddleware_NoError
--- PASS: TestErrorHandlingMiddleware_NoError (0.00s)
=== RUN   TestErrorHandlingMiddleware_Context
    error_handler_test.go:261: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/middleware/error_handler_test.go:261
        	Error:      	Not equal: 
        	            	expected: "test-request-id"
        	            	actual  : ""
        	            	
        	            	Diff:
        	            	--- Expected
        	            	+++ Actual
        	            	@@ -1 +1 @@
        	            	-test-request-id
        	            	+
        	Test:       	TestErrorHandlingMiddleware_Context
--- FAIL: TestErrorHandlingMiddleware_Context (0.00s)
=== RUN   TestErrorHandlingMiddleware_ProductionMode
    error_handler_test.go:301: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/middleware/error_handler_test.go:301
        	Error:      	Expected nil, but got: ""
        	Test:       	TestErrorHandlingMiddleware_ProductionMode
--- FAIL: TestErrorHandlingMiddleware_ProductionMode (0.00s)
=== RUN   TestErrorHandler_GetSuggestions
=== RUN   TestErrorHandler_GetSuggestions/validation_error
=== RUN   TestErrorHandler_GetSuggestions/database_error
=== RUN   TestErrorHandler_GetSuggestions/authorization_error
=== RUN   TestErrorHandler_GetSuggestions/unknown_error
--- PASS: TestErrorHandler_GetSuggestions (0.00s)
    --- PASS: TestErrorHandler_GetSuggestions/validation_error (0.00s)
    --- PASS: TestErrorHandler_GetSuggestions/database_error (0.00s)
    --- PASS: TestErrorHandler_GetSuggestions/authorization_error (0.00s)
    --- PASS: TestErrorHandler_GetSuggestions/unknown_error (0.00s)
=== RUN   TestErrorHandler_SeverityToString
=== RUN   TestErrorHandler_SeverityToString/LOW
=== RUN   TestErrorHandler_SeverityToString/MEDIUM
=== RUN   TestErrorHandler_SeverityToString/HIGH
=== RUN   TestErrorHandler_SeverityToString/CRITICAL
--- PASS: TestErrorHandler_SeverityToString (0.00s)
    --- PASS: TestErrorHandler_SeverityToString/LOW (0.00s)
    --- PASS: TestErrorHandler_SeverityToString/MEDIUM (0.00s)
    --- PASS: TestErrorHandler_SeverityToString/HIGH (0.00s)
    --- PASS: TestErrorHandler_SeverityToString/CRITICAL (0.00s)
=== RUN   TestConvenienceFunctions
--- PASS: TestConvenienceFunctions (0.00s)
=== RUN   TestSecurityMiddleware_Init
--- PASS: TestSecurityMiddleware_Init (0.05s)
=== RUN   TestSecurityMiddleware_RateLimiting
--- PASS: TestSecurityMiddleware_RateLimiting (0.01s)
=== RUN   TestSecurityMiddleware_JWTAuthentication
--- PASS: TestSecurityMiddleware_JWTAuthentication (0.01s)
=== RUN   TestSecurityMiddleware_CombinedMiddleware
--- PASS: TestSecurityMiddleware_CombinedMiddleware (0.01s)
=== RUN   TestSecurityMiddleware_TokenRefresh
--- PASS: TestSecurityMiddleware_TokenRefresh (0.11s)
=== RUN   TestSecurityMiddleware_IPManagement
--- PASS: TestSecurityMiddleware_IPManagement (0.00s)
=== RUN   TestSecurityMiddleware_LegacyCompatibility
--- PASS: TestSecurityMiddleware_LegacyCompatibility (0.01s)
=== RUN   TestSecurityMiddleware_LegacyRateLimiting
    security_test.go:438: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/middleware/security_test.go:438
        	Error:      	Not equal: 
        	            	expected: 429
        	            	actual  : 200
        	Test:       	TestSecurityMiddleware_LegacyRateLimiting
--- FAIL: TestSecurityMiddleware_LegacyRateLimiting (0.00s)
=== RUN   TestSecurityMiddleware_ConcurrentAccess
--- PASS: TestSecurityMiddleware_ConcurrentAccess (0.01s)
=== RUN   TestAuthMiddleware
--- PASS: TestAuthMiddleware (0.00s)
FAIL
FAIL	law-oa-go/internal/middleware	1.491s
?   	law-oa-go/internal/monitoring	[no test files]
?   	law-oa-go/internal/rbac	[no test files]
?   	law-oa-go/internal/router	[no test files]
?   	law-oa-go/internal/server	[no test files]
?   	law-oa-go/internal/validation	[no test files]
?   	law-oa-go/internal/validators	[no test files]
=== RUN   TestBaseRepository_Create
--- PASS: TestBaseRepository_Create (0.01s)
=== RUN   TestBaseRepository_GetByID

2025/09/16 09:49:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/base_repository.go:32 [35;1mrecord not found
[0m[33m[0.056ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE `users`.`id` = 999999 AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
--- PASS: TestBaseRepository_GetByID (0.01s)
=== RUN   TestBaseRepository_Update
--- PASS: TestBaseRepository_Update (0.00s)
=== RUN   TestBaseRepository_Delete

2025/09/16 09:49:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/base_repository_test.go:106 [35;1mrecord not found
[0m[33m[0.058ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE `users`.`id` = 1 AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
--- PASS: TestBaseRepository_Delete (0.00s)
=== RUN   TestBaseRepository_List
--- PASS: TestBaseRepository_List (0.00s)
=== RUN   TestBaseRepository_Count
--- PASS: TestBaseRepository_Count (0.00s)
=== RUN   TestBaseRepository_BatchCreate
--- PASS: TestBaseRepository_BatchCreate (0.00s)
=== RUN   TestBaseRepository_FindWithPreload
--- PASS: TestBaseRepository_FindWithPreload (0.00s)
=== RUN   TestBaseRepository_Transaction
--- PASS: TestBaseRepository_Transaction (0.00s)
=== RUN   TestCaseRepository_Create
--- PASS: TestCaseRepository_Create (0.00s)
=== RUN   TestCaseRepository_FindByID

2025/09/16 09:49:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/case_repository.go:30 [35;1mrecord not found
[0m[33m[0.043ms] [34;1m[rows:0][0m SELECT * FROM `cases` WHERE `cases`.`id` = 999999 AND `cases`.`deleted_at` IS NULL ORDER BY `cases`.`id` LIMIT 1
--- PASS: TestCaseRepository_FindByID (0.00s)
=== RUN   TestCaseRepository_Update
--- PASS: TestCaseRepository_Update (0.00s)
=== RUN   TestCaseRepository_Delete

2025/09/16 09:49:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/case_repository_test.go:110 [35;1mrecord not found
[0m[33m[0.055ms] [34;1m[rows:0][0m SELECT * FROM `cases` WHERE `cases`.`id` = 1 AND `cases`.`deleted_at` IS NULL ORDER BY `cases`.`id` LIMIT 1
--- PASS: TestCaseRepository_Delete (0.00s)
=== RUN   TestCaseRepository_List
--- PASS: TestCaseRepository_List (0.01s)
=== RUN   TestCaseRepository_List_ByClient
--- PASS: TestCaseRepository_List_ByClient (0.01s)
=== RUN   TestCaseRepository_List_ByLawyer
--- PASS: TestCaseRepository_List_ByLawyer (0.00s)
=== RUN   TestCaseRepository_List_Search
--- PASS: TestCaseRepository_List_Search (0.00s)
=== RUN   TestCaseRepository_GetStats
--- PASS: TestCaseRepository_GetStats (0.01s)
=== RUN   TestCaseRepository_AssignLawyer
--- PASS: TestCaseRepository_AssignLawyer (0.00s)
=== RUN   TestCaseRepository_UpdateStatus
--- PASS: TestCaseRepository_UpdateStatus (0.00s)
=== RUN   TestClientRepository_Create
--- PASS: TestClientRepository_Create (0.00s)
=== RUN   TestClientRepository_FindByID

2025/09/16 09:49:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:31 [35;1mrecord not found
[0m[33m[0.044ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE `clients`.`id` = 999999 AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
--- PASS: TestClientRepository_FindByID (0.00s)
=== RUN   TestClientRepository_FindByEmail

2025/09/16 09:49:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.117ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "nonexistent@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
--- PASS: TestClientRepository_FindByEmail (0.00s)
=== RUN   TestClientRepository_Update
--- PASS: TestClientRepository_Update (0.00s)
=== RUN   TestClientRepository_Delete

2025/09/16 09:49:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository_test.go:126 [35;1mrecord not found
[0m[33m[0.051ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE `clients`.`id` = 1 AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
--- PASS: TestClientRepository_Delete (0.00s)
=== RUN   TestClientRepository_List
--- PASS: TestClientRepository_List (0.00s)
=== RUN   TestClientRepository_List_Search
--- PASS: TestClientRepository_List_Search (0.00s)
=== RUN   TestClientRepository_GetStats
--- PASS: TestClientRepository_GetStats (0.00s)
=== RUN   TestClientRepository_Create_DuplicateEmail

2025/09/16 09:49:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:25 [35;1mUNIQUE constraint failed: clients.email
[0m[33m[0.317ms] [34;1m[rows:0][0m INSERT INTO `clients` (`created_at`,`updated_at`,`deleted_at`,`name`,`email`,`phone`,`address`,`company`,`notes`,`status`) VALUES ("2025-09-16 09:49:14.292","2025-09-16 09:49:14.292",NULL,"客户2","same@example.com","13900139001","地址2","","","active") RETURNING `id`
--- PASS: TestClientRepository_Create_DuplicateEmail (0.00s)
=== RUN   TestClientRepository_List_EmptyDatabase
--- PASS: TestClientRepository_List_EmptyDatabase (0.00s)
=== RUN   TestClientRepository_List_InvalidPagination
--- PASS: TestClientRepository_List_InvalidPagination (0.00s)
=== RUN   TestClientRepository_List_SearchCaseInsensitive
--- PASS: TestClientRepository_List_SearchCaseInsensitive (0.00s)
=== RUN   TestQueryBuilder_BasicOperations
=== RUN   TestQueryBuilder_BasicOperations/Where
=== RUN   TestQueryBuilder_BasicOperations/WhereIn
=== RUN   TestQueryBuilder_BasicOperations/WhereNot
=== RUN   TestQueryBuilder_BasicOperations/WhereLike
=== RUN   TestQueryBuilder_BasicOperations/Order
=== RUN   TestQueryBuilder_BasicOperations/OrderDesc
=== RUN   TestQueryBuilder_BasicOperations/OrderAsc
=== RUN   TestQueryBuilder_BasicOperations/Limit
=== RUN   TestQueryBuilder_BasicOperations/Offset
=== RUN   TestQueryBuilder_BasicOperations/Count
=== RUN   TestQueryBuilder_BasicOperations/Exists
=== RUN   TestQueryBuilder_BasicOperations/First

2025/09/16 09:49:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/base_repository.go:269 [35;1mrecord not found
[0m[33m[0.049ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE name = "不存在的用户" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
--- PASS: TestQueryBuilder_BasicOperations (0.01s)
    --- PASS: TestQueryBuilder_BasicOperations/Where (0.00s)
    --- PASS: TestQueryBuilder_BasicOperations/WhereIn (0.00s)
    --- PASS: TestQueryBuilder_BasicOperations/WhereNot (0.00s)
    --- PASS: TestQueryBuilder_BasicOperations/WhereLike (0.00s)
    --- PASS: TestQueryBuilder_BasicOperations/Order (0.00s)
    --- PASS: TestQueryBuilder_BasicOperations/OrderDesc (0.00s)
    --- PASS: TestQueryBuilder_BasicOperations/OrderAsc (0.00s)
    --- PASS: TestQueryBuilder_BasicOperations/Limit (0.00s)
    --- PASS: TestQueryBuilder_BasicOperations/Offset (0.00s)
    --- PASS: TestQueryBuilder_BasicOperations/Count (0.00s)
    --- PASS: TestQueryBuilder_BasicOperations/Exists (0.00s)
    --- PASS: TestQueryBuilder_BasicOperations/First (0.00s)
=== RUN   TestQueryBuilder_ComplexQueries
=== RUN   TestQueryBuilder_ComplexQueries/MultipleConditions
=== RUN   TestQueryBuilder_ComplexQueries/ORConditions
=== RUN   TestQueryBuilder_ComplexQueries/INConditions
=== RUN   TestQueryBuilder_ComplexQueries/LIKEConditions
=== RUN   TestQueryBuilder_ComplexQueries/Pagination
--- PASS: TestQueryBuilder_ComplexQueries (0.01s)
    --- PASS: TestQueryBuilder_ComplexQueries/MultipleConditions (0.00s)
    --- PASS: TestQueryBuilder_ComplexQueries/ORConditions (0.00s)
    --- PASS: TestQueryBuilder_ComplexQueries/INConditions (0.00s)
    --- PASS: TestQueryBuilder_ComplexQueries/LIKEConditions (0.00s)
    --- PASS: TestQueryBuilder_ComplexQueries/Pagination (0.00s)
=== RUN   TestQueryBuilder_JoinOperations
=== RUN   TestQueryBuilder_JoinOperations/Join
--- PASS: TestQueryBuilder_JoinOperations (0.00s)
    --- PASS: TestQueryBuilder_JoinOperations/Join (0.00s)
=== RUN   TestQueryBuilder_GroupAndHaving
=== RUN   TestQueryBuilder_GroupAndHaving/GroupBy
=== RUN   TestQueryBuilder_GroupAndHaving/Having
--- PASS: TestQueryBuilder_GroupAndHaving (0.00s)
    --- PASS: TestQueryBuilder_GroupAndHaving/GroupBy (0.00s)
    --- PASS: TestQueryBuilder_GroupAndHaving/Having (0.00s)
=== RUN   TestQueryBuilder_ErrorHandling
=== RUN   TestQueryBuilder_ErrorHandling/InvalidSQL

2025/09/16 09:49:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/base_repository.go:260 [35;1mno such column: invalid_column
[0m[33m[0.293ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE invalid_column = "value" AND `users`.`deleted_at` IS NULL
=== RUN   TestQueryBuilder_ErrorHandling/EmptyResult
--- PASS: TestQueryBuilder_ErrorHandling (0.00s)
    --- PASS: TestQueryBuilder_ErrorHandling/InvalidSQL (0.00s)
    --- PASS: TestQueryBuilder_ErrorHandling/EmptyResult (0.00s)
=== RUN   TestQueryBuilder_Preload
=== RUN   TestQueryBuilder_Preload/Preload
--- PASS: TestQueryBuilder_Preload (0.00s)
    --- PASS: TestQueryBuilder_Preload/Preload (0.00s)
=== RUN   TestQueryBuilder_LeftJoin
=== RUN   TestQueryBuilder_LeftJoin/LeftJoin
--- PASS: TestQueryBuilder_LeftJoin (0.00s)
    --- PASS: TestQueryBuilder_LeftJoin/LeftJoin (0.00s)
=== RUN   TestQueryBuilder_Distinct
=== RUN   TestQueryBuilder_Distinct/Distinct
--- PASS: TestQueryBuilder_Distinct (0.00s)
    --- PASS: TestQueryBuilder_Distinct/Distinct (0.00s)
=== RUN   TestQueryBuilder_Raw
=== RUN   TestQueryBuilder_Raw/Raw
--- PASS: TestQueryBuilder_Raw (0.00s)
    --- PASS: TestQueryBuilder_Raw/Raw (0.00s)
=== RUN   TestQueryBuilder_Exec
=== RUN   TestQueryBuilder_Exec/Exec
--- PASS: TestQueryBuilder_Exec (0.00s)
    --- PASS: TestQueryBuilder_Exec/Exec (0.00s)
=== RUN   TestQueryBuilder_ComplexChaining
=== RUN   TestQueryBuilder_ComplexChaining/ComplexChaining
--- PASS: TestQueryBuilder_ComplexChaining (0.00s)
    --- PASS: TestQueryBuilder_ComplexChaining/ComplexChaining (0.00s)
=== RUN   TestUserRepository_Create

2025/09/16 09:49:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.067ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "test@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
--- PASS: TestUserRepository_Create (0.01s)
=== RUN   TestUserRepository_Create_DuplicateEmail

2025/09/16 09:49:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.055ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "same@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
--- PASS: TestUserRepository_Create_DuplicateEmail (0.00s)
=== RUN   TestUserRepository_FindByID

2025/09/16 09:49:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/base_repository.go:32 [35;1mrecord not found
[0m[33m[0.042ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE `users`.`id` = 999999 AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
--- PASS: TestUserRepository_FindByID (0.00s)
=== RUN   TestUserRepository_FindByEmail

2025/09/16 09:49:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.045ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "nonexistent@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
--- PASS: TestUserRepository_FindByEmail (0.00s)
=== RUN   TestUserRepository_Update
--- PASS: TestUserRepository_Update (0.00s)
=== RUN   TestUserRepository_Delete

2025/09/16 09:49:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository_test.go:174 [35;1mrecord not found
[0m[33m[0.066ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE `users`.`id` = 1 AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
--- PASS: TestUserRepository_Delete (0.00s)
=== RUN   TestUserRepository_List
--- PASS: TestUserRepository_List (0.00s)
=== RUN   TestUserRepository_List_EmptyDatabase
--- PASS: TestUserRepository_List_EmptyDatabase (0.00s)
=== RUN   TestUserRepository_List_InvalidPagination
--- PASS: TestUserRepository_List_InvalidPagination (0.00s)
=== RUN   TestUserRepository_List_SearchCaseInsensitive
--- PASS: TestUserRepository_List_SearchCaseInsensitive (0.00s)
PASS
ok  	law-oa-go/internal/repositories	1.874s
=== RUN   TestWithRetry_Success
--- PASS: TestWithRetry_Success (0.00s)
=== RUN   TestWithRetry_EventualSuccess
--- PASS: TestWithRetry_EventualSuccess (0.24s)
=== RUN   TestWithRetry_MaxAttemptsReached
    retry_test.go:66: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/retry/retry_test.go:66
        	Error:      	Not equal: 
        	            	expected: 2
        	            	actual  : 1
        	Test:       	TestWithRetry_MaxAttemptsReached
    retry_test.go:67: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/retry/retry_test.go:67
        	Error:      	"temporary error" does not contain "max retry attempts"
        	Test:       	TestWithRetry_MaxAttemptsReached
--- FAIL: TestWithRetry_MaxAttemptsReached (0.00s)
=== RUN   TestWithRetry_ContextCanceled
--- PASS: TestWithRetry_ContextCanceled (0.10s)
=== RUN   TestWithRetry_NonRetryableError
    retry_test.go:103: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/retry/retry_test.go:103
        	Error:      	Not equal: 
        	            	expected: 1
        	            	actual  : 3
        	Test:       	TestWithRetry_NonRetryableError
    retry_test.go:104: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/retry/retry_test.go:104
        	Error:      	"max retry attempts (3) reached, last error: non-retryable error" should not contain "max retry attempts"
        	Test:       	TestWithRetry_NonRetryableError
--- FAIL: TestWithRetry_NonRetryableError (0.52s)
=== RUN   TestWithRetryContext
--- PASS: TestWithRetryContext (0.18s)
=== RUN   TestDatabaseRetryableOperation
--- PASS: TestDatabaseRetryableOperation (0.69s)
=== RUN   TestAPIRetryableOperation
--- PASS: TestAPIRetryableOperation (0.53s)
=== RUN   TestCalculateDelay
=== RUN   TestCalculateDelay/first_attempt
    retry_test.go:198: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/retry/retry_test.go:198
        	Error:      	Not equal: 
        	            	expected: 100ms
        	            	actual  : 0s
        	Test:       	TestCalculateDelay/first_attempt
=== RUN   TestCalculateDelay/second_attempt
    retry_test.go:198: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/retry/retry_test.go:198
        	Error:      	Not equal: 
        	            	expected: 200ms
        	            	actual  : 0s
        	Test:       	TestCalculateDelay/second_attempt
=== RUN   TestCalculateDelay/max_delay
--- FAIL: TestCalculateDelay (0.00s)
    --- FAIL: TestCalculateDelay/first_attempt (0.00s)
    --- FAIL: TestCalculateDelay/second_attempt (0.00s)
    --- PASS: TestCalculateDelay/max_delay (0.00s)
=== RUN   TestWithRetryResult
=== RUN   TestWithRetryResult/success
=== RUN   TestWithRetryResult/failure
    retry_test.go:226: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/retry/retry_test.go:226
        	Error:      	Not equal: 
        	            	expected: 1
        	            	actual  : 3
        	Test:       	TestWithRetryResult/failure
=== RUN   TestWithRetryResult/retry_success
--- FAIL: TestWithRetryResult (0.66s)
    --- PASS: TestWithRetryResult/success (0.00s)
    --- FAIL: TestWithRetryResult/failure (0.43s)
    --- PASS: TestWithRetryResult/retry_success (0.22s)
=== RUN   TestRetryMetrics
=== RUN   TestRetryMetrics/record_success
=== RUN   TestRetryMetrics/record_failure
=== RUN   TestRetryMetrics/record_retry_success
=== RUN   TestRetryMetrics/get_statistics
--- PASS: TestRetryMetrics (0.00s)
    --- PASS: TestRetryMetrics/record_success (0.00s)
    --- PASS: TestRetryMetrics/record_failure (0.00s)
    --- PASS: TestRetryMetrics/record_retry_success (0.00s)
    --- PASS: TestRetryMetrics/get_statistics (0.00s)
=== RUN   TestIsRetryableError
=== RUN   TestIsRetryableError/custom_retryable_error
=== RUN   TestIsRetryableError/non-retryable_error
=== RUN   TestIsRetryableError/database_error
=== RUN   TestIsRetryableError/network_error
=== RUN   TestIsRetryableError/validation_error
--- PASS: TestIsRetryableError (0.00s)
    --- PASS: TestIsRetryableError/custom_retryable_error (0.00s)
    --- PASS: TestIsRetryableError/non-retryable_error (0.00s)
    --- PASS: TestIsRetryableError/database_error (0.00s)
    --- PASS: TestIsRetryableError/network_error (0.00s)
    --- PASS: TestIsRetryableError/validation_error (0.00s)
=== RUN   TestRetryWithJitter
--- PASS: TestRetryWithJitter (0.00s)
=== RUN   TestRetryWithZeroInitialDelay
--- PASS: TestRetryWithZeroInitialDelay (0.00s)
=== RUN   TestRetryWithMaxDelay
--- PASS: TestRetryWithMaxDelay (0.00s)
=== RUN   TestCustomRetryableOperation
--- PASS: TestCustomRetryableOperation (0.00s)
FAIL
FAIL	law-oa-go/internal/retry	3.443s
=== RUN   TestAPISecurityService_NewAPISecurityService
=== RUN   TestAPISecurityService_NewAPISecurityService/创建API安全服务
--- PASS: TestAPISecurityService_NewAPISecurityService (0.00s)
    --- PASS: TestAPISecurityService_NewAPISecurityService/创建API安全服务 (0.00s)
=== RUN   TestAPISecurityService_SecurityMiddleware
=== RUN   TestAPISecurityService_SecurityMiddleware/安全中间件基本功能
=== RUN   TestAPISecurityService_SecurityMiddleware/安全中间件与真实路由集成
=== RUN   TestAPISecurityService_SecurityMiddleware/安全中间件与真实路由集成/有效GET请求
=== RUN   TestAPISecurityService_SecurityMiddleware/安全中间件与真实路由集成/OPTIONS预检请求
--- PASS: TestAPISecurityService_SecurityMiddleware (0.00s)
    --- PASS: TestAPISecurityService_SecurityMiddleware/安全中间件基本功能 (0.00s)
    --- PASS: TestAPISecurityService_SecurityMiddleware/安全中间件与真实路由集成 (0.00s)
        --- PASS: TestAPISecurityService_SecurityMiddleware/安全中间件与真实路由集成/有效GET请求 (0.00s)
        --- PASS: TestAPISecurityService_SecurityMiddleware/安全中间件与真实路由集成/OPTIONS预检请求 (0.00s)
=== RUN   TestAPISecurityService_ValidateRequest
=== RUN   TestAPISecurityService_ValidateRequest/基本请求验证
=== RUN   TestAPISecurityService_ValidateRequest/基本请求验证/有效请求
=== RUN   TestAPISecurityService_ValidateRequest/基本请求验证/包含授权头的请求
--- PASS: TestAPISecurityService_ValidateRequest (0.00s)
    --- PASS: TestAPISecurityService_ValidateRequest/基本请求验证 (0.00s)
        --- PASS: TestAPISecurityService_ValidateRequest/基本请求验证/有效请求 (0.00s)
        --- PASS: TestAPISecurityService_ValidateRequest/基本请求验证/包含授权头的请求 (0.00s)
=== RUN   TestAPISecurityService_IPWhitelistBlacklist
=== RUN   TestAPISecurityService_IPWhitelistBlacklist/IP白名单功能
=== RUN   TestAPISecurityService_IPWhitelistBlacklist/IP白名单功能/白名单IP
=== RUN   TestAPISecurityService_IPWhitelistBlacklist/IP白名单功能/白名单本地IP
=== RUN   TestAPISecurityService_IPWhitelistBlacklist/IP白名单功能/非白名单IP
=== RUN   TestAPISecurityService_IPWhitelistBlacklist/IP白名单功能/空IP
=== RUN   TestAPISecurityService_IPWhitelistBlacklist/IP黑名单功能
=== RUN   TestAPISecurityService_IPWhitelistBlacklist/IP黑名单功能/黑名单IP
=== RUN   TestAPISecurityService_IPWhitelistBlacklist/IP黑名单功能/黑名单本地IP
=== RUN   TestAPISecurityService_IPWhitelistBlacklist/IP黑名单功能/非黑名单IP
=== RUN   TestAPISecurityService_IPWhitelistBlacklist/IP黑名单功能/空IP
=== RUN   TestAPISecurityService_IPWhitelistBlacklist/IP白名单黑名单未启用
--- PASS: TestAPISecurityService_IPWhitelistBlacklist (0.00s)
    --- PASS: TestAPISecurityService_IPWhitelistBlacklist/IP白名单功能 (0.00s)
        --- PASS: TestAPISecurityService_IPWhitelistBlacklist/IP白名单功能/白名单IP (0.00s)
        --- PASS: TestAPISecurityService_IPWhitelistBlacklist/IP白名单功能/白名单本地IP (0.00s)
        --- PASS: TestAPISecurityService_IPWhitelistBlacklist/IP白名单功能/非白名单IP (0.00s)
        --- PASS: TestAPISecurityService_IPWhitelistBlacklist/IP白名单功能/空IP (0.00s)
    --- PASS: TestAPISecurityService_IPWhitelistBlacklist/IP黑名单功能 (0.00s)
        --- PASS: TestAPISecurityService_IPWhitelistBlacklist/IP黑名单功能/黑名单IP (0.00s)
        --- PASS: TestAPISecurityService_IPWhitelistBlacklist/IP黑名单功能/黑名单本地IP (0.00s)
        --- PASS: TestAPISecurityService_IPWhitelistBlacklist/IP黑名单功能/非黑名单IP (0.00s)
        --- PASS: TestAPISecurityService_IPWhitelistBlacklist/IP黑名单功能/空IP (0.00s)
    --- PASS: TestAPISecurityService_IPWhitelistBlacklist/IP白名单黑名单未启用 (0.00s)
=== RUN   TestAPISecurityService_CheckRateLimit
=== RUN   TestAPISecurityService_CheckRateLimit/限流检查
=== RUN   TestAPISecurityService_CheckRateLimit/限流未启用
--- PASS: TestAPISecurityService_CheckRateLimit (0.00s)
    --- PASS: TestAPISecurityService_CheckRateLimit/限流检查 (0.00s)
    --- PASS: TestAPISecurityService_CheckRateLimit/限流未启用 (0.00s)
=== RUN   TestAPISecurityService_DetectWAFAttack
=== RUN   TestAPISecurityService_DetectWAFAttack/WAF攻击检测
=== RUN   TestAPISecurityService_DetectWAFAttack/WAF攻击检测/正常请求
=== RUN   TestAPISecurityService_DetectWAFAttack/WAF攻击检测/可疑URL请求
=== RUN   TestAPISecurityService_DetectWAFAttack/WAF攻击检测/可疑User-Agent
--- PASS: TestAPISecurityService_DetectWAFAttack (0.00s)
    --- PASS: TestAPISecurityService_DetectWAFAttack/WAF攻击检测 (0.00s)
        --- PASS: TestAPISecurityService_DetectWAFAttack/WAF攻击检测/正常请求 (0.00s)
        --- PASS: TestAPISecurityService_DetectWAFAttack/WAF攻击检测/可疑URL请求 (0.00s)
        --- PASS: TestAPISecurityService_DetectWAFAttack/WAF攻击检测/可疑User-Agent (0.00s)
=== RUN   TestAPISecurityService_ValidateCSRFToken
=== RUN   TestAPISecurityService_ValidateCSRFToken/CSRF令牌验证
=== RUN   TestAPISecurityService_ValidateCSRFToken/CSRF令牌验证/无CSRF令牌
=== RUN   TestAPISecurityService_ValidateCSRFToken/CSRF令牌验证/头部CSRF令牌
=== RUN   TestAPISecurityService_ValidateCSRFToken/CSRF令牌验证/表单CSRF令牌
--- PASS: TestAPISecurityService_ValidateCSRFToken (0.00s)
    --- PASS: TestAPISecurityService_ValidateCSRFToken/CSRF令牌验证 (0.00s)
        --- PASS: TestAPISecurityService_ValidateCSRFToken/CSRF令牌验证/无CSRF令牌 (0.00s)
        --- PASS: TestAPISecurityService_ValidateCSRFToken/CSRF令牌验证/头部CSRF令牌 (0.00s)
        --- PASS: TestAPISecurityService_ValidateCSRFToken/CSRF令牌验证/表单CSRF令牌 (0.00s)
=== RUN   TestAPISecurityService_CORS
=== RUN   TestAPISecurityService_CORS/CORS中间件功能
=== RUN   TestAPISecurityService_CORS/CORS中间件功能/简单请求
=== RUN   TestAPISecurityService_CORS/CORS中间件功能/预检请求
=== RUN   TestAPISecurityService_CORS/CORS未启用
--- PASS: TestAPISecurityService_CORS (0.00s)
    --- PASS: TestAPISecurityService_CORS/CORS中间件功能 (0.00s)
        --- PASS: TestAPISecurityService_CORS/CORS中间件功能/简单请求 (0.00s)
        --- PASS: TestAPISecurityService_CORS/CORS中间件功能/预检请求 (0.00s)
    --- PASS: TestAPISecurityService_CORS/CORS未启用 (0.00s)
=== RUN   TestAPISecurityService_Integration
=== RUN   TestAPISecurityService_Integration/完整的安全中间件链
=== RUN   TestAPISecurityService_Integration/完整的安全中间件链/有效的安全请求
=== RUN   TestAPISecurityService_Integration/完整的安全中间件链/预检请求
--- PASS: TestAPISecurityService_Integration (0.00s)
    --- PASS: TestAPISecurityService_Integration/完整的安全中间件链 (0.00s)
        --- PASS: TestAPISecurityService_Integration/完整的安全中间件链/有效的安全请求 (0.00s)
        --- PASS: TestAPISecurityService_Integration/完整的安全中间件链/预检请求 (0.00s)
=== RUN   TestAuditService_NewAuditService
=== RUN   TestAuditService_NewAuditService/创建审计服务_-_启用审计
=== RUN   TestAuditService_NewAuditService/创建审计服务_-_禁用审计
--- PASS: TestAuditService_NewAuditService (0.01s)
    --- PASS: TestAuditService_NewAuditService/创建审计服务_-_启用审计 (0.00s)
    --- PASS: TestAuditService_NewAuditService/创建审计服务_-_禁用审计 (0.01s)
=== RUN   TestAuditService_LogEvent
=== RUN   TestAuditService_LogEvent/记录基本审计事件
2025/09/16 09:49:14 Audit worker 1 started
2025/09/16 09:49:14 Audit worker 1 started
2025/09/16 09:49:14 Audit worker 0 started
2025/09/16 09:49:14 Audit worker 4 started
2025/09/16 09:49:14 Audit worker 0 started
2025/09/16 09:49:14 Audit worker 3 started
2025/09/16 09:49:14 Audit worker 2 started
2025/09/16 09:49:14 Audit worker 2 started
2025/09/16 09:49:14 Audit worker 3 started
2025/09/16 09:49:14 Audit worker 4 started
=== RUN   TestAuditService_LogEvent/记录事件_-_审计禁用
--- PASS: TestAuditService_LogEvent (0.20s)
    --- PASS: TestAuditService_LogEvent/记录基本审计事件 (0.20s)
    --- PASS: TestAuditService_LogEvent/记录事件_-_审计禁用 (0.00s)
=== RUN   TestAuditService_SpecificEventMethods
=== RUN   TestAuditService_SpecificEventMethods/测试各种事件记录方法
2025/09/16 09:49:14 Audit worker 2 started
2025/09/16 09:49:14 Audit worker 1 started
2025/09/16 09:49:14 Audit worker 4 started
2025/09/16 09:49:14 Audit worker 0 started
2025/09/16 09:49:14 Audit worker 3 started
=== RUN   TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录登录事件
=== RUN   TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录登出事件
=== RUN   TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录密码重置事件
=== RUN   TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录权限变更事件
=== RUN   TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录数据访问事件
=== RUN   TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录数据修改事件
=== RUN   TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录数据删除事件
=== RUN   TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录安全事件
=== RUN   TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录API访问事件
=== RUN   TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录文件操作事件
--- PASS: TestAuditService_SpecificEventMethods (0.51s)
    --- PASS: TestAuditService_SpecificEventMethods/测试各种事件记录方法 (0.51s)
        --- PASS: TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录登录事件 (0.05s)
        --- PASS: TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录登出事件 (0.05s)
        --- PASS: TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录密码重置事件 (0.05s)
        --- PASS: TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录权限变更事件 (0.05s)
        --- PASS: TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录数据访问事件 (0.05s)
        --- PASS: TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录数据修改事件 (0.05s)
        --- PASS: TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录数据删除事件 (0.05s)
        --- PASS: TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录安全事件 (0.05s)
        --- PASS: TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录API访问事件 (0.05s)
        --- PASS: TestAuditService_SpecificEventMethods/测试各种事件记录方法/记录文件操作事件 (0.05s)
=== RUN   TestAuditService_QueryAuditLogs
=== RUN   TestAuditService_QueryAuditLogs/基本查询功能
=== RUN   TestAuditService_QueryAuditLogs/查询过滤条件
=== RUN   TestAuditService_QueryAuditLogs/查询过滤条件/按用户ID过滤
=== RUN   TestAuditService_QueryAuditLogs/查询过滤条件/按用户名模糊匹配
=== RUN   TestAuditService_QueryAuditLogs/查询过滤条件/按事件类型过滤
=== RUN   TestAuditService_QueryAuditLogs/查询过滤条件/按严重程度过滤
=== RUN   TestAuditService_QueryAuditLogs/查询过滤条件/按时间范围过滤
=== RUN   TestAuditService_QueryAuditLogs/查询过滤条件/按IP地址过滤
=== RUN   TestAuditService_QueryAuditLogs/查询过滤条件/分页查询
=== RUN   TestAuditService_QueryAuditLogs/查询过滤条件/排序查询
--- PASS: TestAuditService_QueryAuditLogs (0.00s)
    --- PASS: TestAuditService_QueryAuditLogs/基本查询功能 (0.00s)
    --- PASS: TestAuditService_QueryAuditLogs/查询过滤条件 (0.00s)
        --- PASS: TestAuditService_QueryAuditLogs/查询过滤条件/按用户ID过滤 (0.00s)
        --- PASS: TestAuditService_QueryAuditLogs/查询过滤条件/按用户名模糊匹配 (0.00s)
        --- PASS: TestAuditService_QueryAuditLogs/查询过滤条件/按事件类型过滤 (0.00s)
        --- PASS: TestAuditService_QueryAuditLogs/查询过滤条件/按严重程度过滤 (0.00s)
        --- PASS: TestAuditService_QueryAuditLogs/查询过滤条件/按时间范围过滤 (0.00s)
        --- PASS: TestAuditService_QueryAuditLogs/查询过滤条件/按IP地址过滤 (0.00s)
        --- PASS: TestAuditService_QueryAuditLogs/查询过滤条件/分页查询 (0.00s)
        --- PASS: TestAuditService_QueryAuditLogs/查询过滤条件/排序查询 (0.00s)
=== RUN   TestAuditService_AuditMiddleware
=== RUN   TestAuditService_AuditMiddleware/审计中间件基本功能
2025/09/16 09:49:15 Audit worker 0 started
2025/09/16 09:49:15 Audit worker 2 started
2025/09/16 09:49:15 Audit worker 1 started
2025/09/16 09:49:15 Audit worker 4 started
2025/09/16 09:49:15 Audit worker 3 started
=== RUN   TestAuditService_AuditMiddleware/审计中间件基本功能/带用户信息的请求
=== RUN   TestAuditService_AuditMiddleware/审计中间件基本功能/不带用户信息的请求
--- PASS: TestAuditService_AuditMiddleware (0.11s)
    --- PASS: TestAuditService_AuditMiddleware/审计中间件基本功能 (0.11s)
        --- PASS: TestAuditService_AuditMiddleware/审计中间件基本功能/带用户信息的请求 (0.05s)
        --- PASS: TestAuditService_AuditMiddleware/审计中间件基本功能/不带用户信息的请求 (0.05s)
=== RUN   TestAuditService_SecurityAlerts
=== RUN   TestAuditService_SecurityAlerts/安全事件告警
2025/09/16 09:49:15 Audit worker 0 started
2025/09/16 09:49:15 Audit worker 2 started
2025/09/16 09:49:15 Audit worker 3 started
2025/09/16 09:49:15 Audit worker 4 started
2025/09/16 09:49:15 Audit worker 1 started
2025/09/16 09:49:15 SECURITY ALERT: testuser - security_breach - Security breach detected
2025/09/16 09:49:15 SENSITIVE OPERATION ALERT: testuser performed security_breach on 
=== RUN   TestAuditService_SecurityAlerts/登录失败告警
2025/09/16 09:49:15 Audit worker 0 started
2025/09/16 09:49:15 Audit worker 1 started
2025/09/16 09:49:15 Audit worker 2 started
2025/09/16 09:49:15 Audit worker 3 started
2025/09/16 09:49:15 Audit worker 4 started
2025/09/16 09:49:15 LOGIN FAILURE ALERT: User testuser failed login from 192.168.1.100 (no cache available)
2025/09/16 09:49:15 LOGIN FAILURE ALERT: User testuser failed login from 192.168.1.100 (no cache available)
2025/09/16 09:49:15 LOGIN FAILURE ALERT: User testuser failed login from 192.168.1.100 (no cache available)
=== RUN   TestAuditService_SecurityAlerts/敏感操作告警
2025/09/16 09:49:15 Audit worker 1 started
2025/09/16 09:49:15 Audit worker 0 started
2025/09/16 09:49:15 Audit worker 3 started
2025/09/16 09:49:15 Audit worker 2 started
2025/09/16 09:49:15 Audit worker 4 started
2025/09/16 09:49:15 SENSITIVE OPERATION ALERT: admin performed permission_change on user_permissions
--- PASS: TestAuditService_SecurityAlerts (0.60s)
    --- PASS: TestAuditService_SecurityAlerts/安全事件告警 (0.20s)
    --- PASS: TestAuditService_SecurityAlerts/登录失败告警 (0.20s)
    --- PASS: TestAuditService_SecurityAlerts/敏感操作告警 (0.20s)
=== RUN   TestAuditService_DataMasking
=== RUN   TestAuditService_DataMasking/IP地址脱敏
=== RUN   TestAuditService_DataMasking/IP地址脱敏/192.168.1.100
=== RUN   TestAuditService_DataMasking/IP地址脱敏/10.0.0.1
=== RUN   TestAuditService_DataMasking/IP地址脱敏/127.0.0.1
=== RUN   TestAuditService_DataMasking/IP地址脱敏/255.255.255.255
=== RUN   TestAuditService_DataMasking/IP地址脱敏/invalid-ip
=== RUN   TestAuditService_DataMasking/IP地址脱敏/#00
=== RUN   TestAuditService_DataMasking/用户代理脱敏
=== RUN   TestAuditService_DataMasking/用户代理脱敏/Mozilla/5.0_(Windows_NT_10.0;_Win64;_x64)
=== RUN   TestAuditService_DataMasking/用户代理脱敏/short
=== RUN   TestAuditService_DataMasking/用户代理脱敏/#00
=== RUN   TestAuditService_DataMasking/用户代理脱敏/Mozilla/5.0
=== RUN   TestAuditService_DataMasking/用户代理脱敏/This_is_a_very_long_user_agent_string_that_should_be_truncated
--- PASS: TestAuditService_DataMasking (0.00s)
    --- PASS: TestAuditService_DataMasking/IP地址脱敏 (0.00s)
        --- PASS: TestAuditService_DataMasking/IP地址脱敏/192.168.1.100 (0.00s)
        --- PASS: TestAuditService_DataMasking/IP地址脱敏/10.0.0.1 (0.00s)
        --- PASS: TestAuditService_DataMasking/IP地址脱敏/127.0.0.1 (0.00s)
        --- PASS: TestAuditService_DataMasking/IP地址脱敏/255.255.255.255 (0.00s)
        --- PASS: TestAuditService_DataMasking/IP地址脱敏/invalid-ip (0.00s)
        --- PASS: TestAuditService_DataMasking/IP地址脱敏/#00 (0.00s)
    --- PASS: TestAuditService_DataMasking/用户代理脱敏 (0.00s)
        --- PASS: TestAuditService_DataMasking/用户代理脱敏/Mozilla/5.0_(Windows_NT_10.0;_Win64;_x64) (0.00s)
        --- PASS: TestAuditService_DataMasking/用户代理脱敏/short (0.00s)
        --- PASS: TestAuditService_DataMasking/用户代理脱敏/#00 (0.00s)
        --- PASS: TestAuditService_DataMasking/用户代理脱敏/Mozilla/5.0 (0.00s)
        --- PASS: TestAuditService_DataMasking/用户代理脱敏/This_is_a_very_long_user_agent_string_that_should_be_truncated (0.00s)
=== RUN   TestAuditService_Concurrency
=== RUN   TestAuditService_Concurrency/并发事件记录
2025/09/16 09:49:15 Audit worker 1 started
2025/09/16 09:49:15 Audit worker 4 started
2025/09/16 09:49:15 Audit worker 0 started
2025/09/16 09:49:15 Audit worker 2 started
2025/09/16 09:49:15 Audit worker 3 started
--- PASS: TestAuditService_Concurrency (0.30s)
    --- PASS: TestAuditService_Concurrency/并发事件记录 (0.30s)
=== RUN   TestAuditService_Stop
=== RUN   TestAuditService_Stop/正常停止服务
2025/09/16 09:49:16 Audit worker 0 started
2025/09/16 09:49:16 Audit worker 1 started
2025/09/16 09:49:16 Audit worker 2 started
2025/09/16 09:49:16 Audit worker 3 started
2025/09/16 09:49:16 Audit worker 4 started
2025/09/16 09:49:16 Audit worker 2 stopped
2025/09/16 09:49:16 Audit worker 4 stopped
2025/09/16 09:49:16 Audit worker 3 stopped
2025/09/16 09:49:16 Audit worker 0 stopped
2025/09/16 09:49:16 Audit worker 1 stopped
--- PASS: TestAuditService_Stop (0.20s)
    --- PASS: TestAuditService_Stop/正常停止服务 (0.20s)
=== RUN   TestAuditService_GetAuditStats
=== RUN   TestAuditService_GetAuditStats/获取审计统计信息
2025/09/16 09:49:16 Audit worker 0 started
2025/09/16 09:49:16 Audit worker 2 started
2025/09/16 09:49:16 Audit worker 4 started
2025/09/16 09:49:16 Audit worker 3 started
2025/09/16 09:49:16 Audit worker 1 started

2025/09/16 09:49:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/security/audit.go:740 [35;1mall expectations were already fulfilled, call to Query 'SELECT count(*) FROM `audit_events` WHERE timestamp >= ? AND timestamp < ?' with args [{Name: Ordinal:1 Value:2025-09-16 08:00:00 +0800 CST} {Name: Ordinal:2 Value:2025-09-17 08:00:00 +0800 CST}] was not expected
[0m[33m[17.933ms] [34;1m[rows:0][0m SELECT count(*) FROM `audit_events` WHERE timestamp >= '2025-09-16 08:00:00' AND timestamp < '2025-09-17 08:00:00'

2025/09/16 09:49:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/security/audit.go:746 [35;1mall expectations were already fulfilled, call to Query 'SELECT count(*) FROM `audit_events` WHERE timestamp >= ?' with args [{Name: Ordinal:1 Value:2025-09-14 08:00:00 +0800 CST}] was not expected
[0m[33m[0.213ms] [34;1m[rows:0][0m SELECT count(*) FROM `audit_events` WHERE timestamp >= '2025-09-14 08:00:00'

2025/09/16 09:49:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/security/audit.go:751 [35;1mall expectations were already fulfilled, call to Query 'SELECT count(*) FROM `audit_events` WHERE event_type = ?' with args [{Name: Ordinal:1 Value:security_event}] was not expected
[0m[33m[0.044ms] [34;1m[rows:0][0m SELECT count(*) FROM `audit_events` WHERE event_type = 'security_event'
--- PASS: TestAuditService_GetAuditStats (0.02s)
    --- PASS: TestAuditService_GetAuditStats/获取审计统计信息 (0.02s)
=== RUN   TestEncryptionService_NewEncryptionService
=== RUN   TestEncryptionService_NewEncryptionService/创建加密服务_-_无配置密钥
=== RUN   TestEncryptionService_NewEncryptionService/创建加密服务_-_有配置密钥
--- PASS: TestEncryptionService_NewEncryptionService (0.56s)
    --- PASS: TestEncryptionService_NewEncryptionService/创建加密服务_-_无配置密钥 (0.27s)
    --- PASS: TestEncryptionService_NewEncryptionService/创建加密服务_-_有配置密钥 (0.29s)
=== RUN   TestEncryptionService_AES_EncryptDecrypt
=== RUN   TestEncryptionService_AES_EncryptDecrypt/AES加密解密循环测试
=== RUN   TestEncryptionService_AES_EncryptDecrypt/AES加密解密循环测试/简单文本
=== RUN   TestEncryptionService_AES_EncryptDecrypt/AES加密解密循环测试/中文文本
=== RUN   TestEncryptionService_AES_EncryptDecrypt/AES加密解密循环测试/特殊字符
=== RUN   TestEncryptionService_AES_EncryptDecrypt/AES加密解密循环测试/长文本
=== RUN   TestEncryptionService_AES_EncryptDecrypt/AES加密解密循环测试/空字符串
=== RUN   TestEncryptionService_AES_EncryptDecrypt/AES加密错误处理
--- PASS: TestEncryptionService_AES_EncryptDecrypt (0.40s)
    --- PASS: TestEncryptionService_AES_EncryptDecrypt/AES加密解密循环测试 (0.15s)
        --- PASS: TestEncryptionService_AES_EncryptDecrypt/AES加密解密循环测试/简单文本 (0.00s)
        --- PASS: TestEncryptionService_AES_EncryptDecrypt/AES加密解密循环测试/中文文本 (0.00s)
        --- PASS: TestEncryptionService_AES_EncryptDecrypt/AES加密解密循环测试/特殊字符 (0.00s)
        --- PASS: TestEncryptionService_AES_EncryptDecrypt/AES加密解密循环测试/长文本 (0.00s)
        --- PASS: TestEncryptionService_AES_EncryptDecrypt/AES加密解密循环测试/空字符串 (0.00s)
    --- PASS: TestEncryptionService_AES_EncryptDecrypt/AES加密错误处理 (0.26s)
=== RUN   TestEncryptionService_RSA_EncryptDecrypt
=== RUN   TestEncryptionService_RSA_EncryptDecrypt/RSA加密解密循环测试
=== RUN   TestEncryptionService_RSA_EncryptDecrypt/RSA加密解密循环测试/短文本
=== RUN   TestEncryptionService_RSA_EncryptDecrypt/RSA加密解密循环测试/中等长度文本
=== RUN   TestEncryptionService_RSA_EncryptDecrypt/RSA加密解密循环测试/数字
--- PASS: TestEncryptionService_RSA_EncryptDecrypt (0.33s)
    --- PASS: TestEncryptionService_RSA_EncryptDecrypt/RSA加密解密循环测试 (0.33s)
        --- PASS: TestEncryptionService_RSA_EncryptDecrypt/RSA加密解密循环测试/短文本 (0.02s)
        --- PASS: TestEncryptionService_RSA_EncryptDecrypt/RSA加密解密循环测试/中等长度文本 (0.01s)
        --- PASS: TestEncryptionService_RSA_EncryptDecrypt/RSA加密解密循环测试/数字 (0.01s)
=== RUN   TestEncryptionService_PasswordHashing
=== RUN   TestEncryptionService_PasswordHashing/密码哈希验证
=== RUN   TestEncryptionService_PasswordHashing/密码哈希验证/简单密码
=== RUN   TestEncryptionService_PasswordHashing/密码哈希验证/复杂密码
=== RUN   TestEncryptionService_PasswordHashing/密码哈希验证/中文密码
=== RUN   TestEncryptionService_PasswordHashing/密码哈希验证/长密码
=== RUN   TestEncryptionService_PasswordHashing/密码哈希错误处理
--- PASS: TestEncryptionService_PasswordHashing (44.70s)
    --- PASS: TestEncryptionService_PasswordHashing/密码哈希验证 (44.29s)
        --- PASS: TestEncryptionService_PasswordHashing/密码哈希验证/简单密码 (11.08s)
        --- PASS: TestEncryptionService_PasswordHashing/密码哈希验证/复杂密码 (11.01s)
        --- PASS: TestEncryptionService_PasswordHashing/密码哈希验证/中文密码 (11.08s)
        --- PASS: TestEncryptionService_PasswordHashing/密码哈希验证/长密码 (11.02s)
    --- PASS: TestEncryptionService_PasswordHashing/密码哈希错误处理 (0.40s)
=== RUN   TestEncryptionService_FieldEncryption
=== RUN   TestEncryptionService_FieldEncryption/敏感字段加密
=== RUN   TestEncryptionService_FieldEncryption/敏感字段加密/加密邮箱字段
=== RUN   TestEncryptionService_FieldEncryption/敏感字段加密/加密手机字段
=== RUN   TestEncryptionService_FieldEncryption/敏感字段加密/加密身份证字段
=== RUN   TestEncryptionService_FieldEncryption/敏感字段加密/不加密姓名字段
=== RUN   TestEncryptionService_FieldEncryption/敏感字段加密/不加密地址字段
=== RUN   TestEncryptionService_FieldEncryption/字段加密未启用
--- PASS: TestEncryptionService_FieldEncryption (0.65s)
    --- PASS: TestEncryptionService_FieldEncryption/敏感字段加密 (0.43s)
        --- PASS: TestEncryptionService_FieldEncryption/敏感字段加密/加密邮箱字段 (0.00s)
        --- PASS: TestEncryptionService_FieldEncryption/敏感字段加密/加密手机字段 (0.00s)
        --- PASS: TestEncryptionService_FieldEncryption/敏感字段加密/加密身份证字段 (0.00s)
        --- PASS: TestEncryptionService_FieldEncryption/敏感字段加密/不加密姓名字段 (0.00s)
        --- PASS: TestEncryptionService_FieldEncryption/敏感字段加密/不加密地址字段 (0.00s)
    --- PASS: TestEncryptionService_FieldEncryption/字段加密未启用 (0.22s)
=== RUN   TestEncryptionService_UserDataEncryption
=== RUN   TestEncryptionService_UserDataEncryption/用户数据加密解密
--- PASS: TestEncryptionService_UserDataEncryption (0.22s)
    --- PASS: TestEncryptionService_UserDataEncryption/用户数据加密解密 (0.22s)
=== RUN   TestEncryptionService_UtilityFunctions
=== RUN   TestEncryptionService_UtilityFunctions/生成数据密钥
=== RUN   TestEncryptionService_UtilityFunctions/计算哈希值
=== RUN   TestEncryptionService_UtilityFunctions/输入清理
=== RUN   TestEncryptionService_UtilityFunctions/输入清理/正常输入
=== RUN   TestEncryptionService_UtilityFunctions/输入清理/包含控制字符
=== RUN   TestEncryptionService_UtilityFunctions/输入清理/SQL注入尝试
=== RUN   TestEncryptionService_UtilityFunctions/输入清理/包含SQL注入模式
=== RUN   TestEncryptionService_UtilityFunctions/输入清理/包含注释
=== RUN   TestEncryptionService_UtilityFunctions/输入清理/前后空格
=== RUN   TestEncryptionService_UtilityFunctions/生成安全令牌
--- PASS: TestEncryptionService_UtilityFunctions (1.66s)
    --- PASS: TestEncryptionService_UtilityFunctions/生成数据密钥 (0.35s)
    --- PASS: TestEncryptionService_UtilityFunctions/计算哈希值 (0.21s)
    --- PASS: TestEncryptionService_UtilityFunctions/输入清理 (0.10s)
        --- PASS: TestEncryptionService_UtilityFunctions/输入清理/正常输入 (0.00s)
        --- PASS: TestEncryptionService_UtilityFunctions/输入清理/包含控制字符 (0.00s)
        --- PASS: TestEncryptionService_UtilityFunctions/输入清理/SQL注入尝试 (0.00s)
        --- PASS: TestEncryptionService_UtilityFunctions/输入清理/包含SQL注入模式 (0.00s)
        --- PASS: TestEncryptionService_UtilityFunctions/输入清理/包含注释 (0.00s)
        --- PASS: TestEncryptionService_UtilityFunctions/输入清理/前后空格 (0.00s)
    --- PASS: TestEncryptionService_UtilityFunctions/生成安全令牌 (0.99s)
=== RUN   TestEncryptionService_KeyManagement
=== RUN   TestEncryptionService_KeyManagement/密钥旋转
=== RUN   TestEncryptionService_KeyManagement/获取加密状态
--- PASS: TestEncryptionService_KeyManagement (0.64s)
    --- PASS: TestEncryptionService_KeyManagement/密钥旋转 (0.27s)
    --- PASS: TestEncryptionService_KeyManagement/获取加密状态 (0.37s)
=== RUN   TestEncryptionService_Metrics
=== RUN   TestEncryptionService_Metrics/加密操作指标
--- PASS: TestEncryptionService_Metrics (3.91s)
    --- PASS: TestEncryptionService_Metrics/加密操作指标 (3.91s)
=== RUN   TestNewJWTKeyManager
--- PASS: TestNewJWTKeyManager (0.00s)
=== RUN   TestJWTKeyManager_CreateTokens
--- PASS: TestJWTKeyManager_CreateTokens (0.04s)
=== RUN   TestJWTKeyManager_VerifyToken
--- PASS: TestJWTKeyManager_VerifyToken (0.02s)
=== RUN   TestJWTKeyManager_RotateKeys
JWT key rotated from v1 to v2
--- PASS: TestJWTKeyManager_RotateKeys (0.00s)
=== RUN   TestNewRateLimiter
--- PASS: TestNewRateLimiter (0.00s)
=== RUN   TestRateLimiter_Middleware
    security_test.go:253: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/security/security_test.go:253
        	Error:      	Not equal: 
        	            	expected: 200
        	            	actual  : 429
        	Test:       	TestRateLimiter_Middleware
--- FAIL: TestRateLimiter_Middleware (0.02s)
=== RUN   TestRateLimiter_IPWhitelist
--- PASS: TestRateLimiter_IPWhitelist (0.00s)
=== RUN   TestRateLimiter_IPBlacklist
--- PASS: TestRateLimiter_IPBlacklist (0.00s)
=== RUN   TestRateLimiter_GetRateLimitInfo
--- PASS: TestRateLimiter_GetRateLimitInfo (0.00s)
=== RUN   TestRateLimiter_GetStats
--- PASS: TestRateLimiter_GetStats (0.00s)
=== RUN   TestRateLimiter_ClientIP
--- PASS: TestRateLimiter_ClientIP (0.00s)
FAIL
FAIL	law-oa-go/internal/security	56.021s
=== RUN   TestCaseService_CreateCase
=== RUN   TestCaseService_CreateCase/Basic_Test
--- PASS: TestCaseService_CreateCase (0.00s)
    --- PASS: TestCaseService_CreateCase/Basic_Test (0.00s)
=== RUN   TestCaseService_GetCaseByID
=== RUN   TestCaseService_GetCaseByID/Basic_Test
--- PASS: TestCaseService_GetCaseByID (0.00s)
    --- PASS: TestCaseService_GetCaseByID/Basic_Test (0.00s)
=== RUN   TestCaseService_UpdateCase
=== RUN   TestCaseService_UpdateCase/Basic_Test
--- PASS: TestCaseService_UpdateCase (0.00s)
    --- PASS: TestCaseService_UpdateCase/Basic_Test (0.00s)
=== RUN   TestCaseService_DeleteCase
=== RUN   TestCaseService_DeleteCase/Basic_Test
--- PASS: TestCaseService_DeleteCase (0.00s)
    --- PASS: TestCaseService_DeleteCase/Basic_Test (0.00s)
=== RUN   TestCaseService_ListCases
=== RUN   TestCaseService_ListCases/Basic_Test
--- PASS: TestCaseService_ListCases (0.00s)
    --- PASS: TestCaseService_ListCases/Basic_Test (0.00s)
=== RUN   TestCaseService_GetCaseStats
=== RUN   TestCaseService_GetCaseStats/Basic_Test
--- PASS: TestCaseService_GetCaseStats (0.00s)
    --- PASS: TestCaseService_GetCaseStats/Basic_Test (0.00s)
=== RUN   TestNewSearchService
--- PASS: TestNewSearchService (0.00s)
=== RUN   TestNewSearchServiceWithNilClient
--- PASS: TestNewSearchServiceWithNilClient (0.00s)
=== RUN   TestSearchService_Search_Success
--- PASS: TestSearchService_Search_Success (0.00s)
=== RUN   TestSearchService_Search_EmptyQuery
--- PASS: TestSearchService_Search_EmptyQuery (0.00s)
=== RUN   TestSearchService_Search_DefaultValues
--- PASS: TestSearchService_Search_DefaultValues (0.00s)
=== RUN   TestSearchService_Search_MaxPageSize
--- PASS: TestSearchService_Search_MaxPageSize (0.00s)
=== RUN   TestSearchService_Search_FallbackSearch
--- PASS: TestSearchService_Search_FallbackSearch (0.00s)
=== RUN   TestSearchService_Search_WithFilters
--- PASS: TestSearchService_Search_WithFilters (0.00s)
=== RUN   TestSearchService_GetSearchSuggestions_Success
--- PASS: TestSearchService_GetSearchSuggestions_Success (0.00s)
=== RUN   TestSearchService_GetSearchSuggestions_EmptyQuery
--- PASS: TestSearchService_GetSearchSuggestions_EmptyQuery (0.00s)
=== RUN   TestSearchService_GetSearchSuggestions_Limit
--- PASS: TestSearchService_GetSearchSuggestions_Limit (0.00s)
=== RUN   TestSearchService_GetSearchSuggestions_Fallback
--- PASS: TestSearchService_GetSearchSuggestions_Fallback (0.00s)
=== RUN   TestSearchService_IndexEntity_Success
Would index entity case/123: map[content:Test content status:active title:Test Case]
--- PASS: TestSearchService_IndexEntity_Success (0.00s)
=== RUN   TestSearchService_IndexEntity_Fallback
--- PASS: TestSearchService_IndexEntity_Fallback (0.00s)
=== RUN   TestSearchService_DeleteEntityFromIndex_Success
Would delete entity case/123 from index
--- PASS: TestSearchService_DeleteEntityFromIndex_Success (0.00s)
=== RUN   TestSearchService_DeleteEntityFromIndex_Fallback
--- PASS: TestSearchService_DeleteEntityFromIndex_Fallback (0.00s)
=== RUN   TestSearchService_ReindexAll_Success
Would reindex all entities
--- PASS: TestSearchService_ReindexAll_Success (0.00s)
=== RUN   TestSearchService_ReindexAll_Fallback
--- PASS: TestSearchService_ReindexAll_Fallback (0.00s)
=== RUN   TestSearchService_buildSearchQuery
--- PASS: TestSearchService_buildSearchQuery (0.00s)
=== RUN   TestSearchService_fallbackSearch
--- PASS: TestSearchService_fallbackSearch (0.00s)
=== RUN   TestSearchService_generateSuggestions
--- PASS: TestSearchService_generateSuggestions (0.00s)
=== RUN   TestSearchService_Performance
--- PASS: TestSearchService_Performance (0.00s)
=== RUN   TestSearchService_ConcurrentAccess
--- PASS: TestSearchService_ConcurrentAccess (0.00s)
=== RUN   TestSearchService_ContextCancellation
--- PASS: TestSearchService_ContextCancellation (0.00s)
=== RUN   TestSearchService_EdgeCases
=== RUN   TestSearchService_EdgeCases/VeryLongQuery
=== RUN   TestSearchService_EdgeCases/SpecialCharactersInQuery
=== RUN   TestSearchService_EdgeCases/EmptyTypesFilter
--- PASS: TestSearchService_EdgeCases (0.00s)
    --- PASS: TestSearchService_EdgeCases/VeryLongQuery (0.00s)
    --- PASS: TestSearchService_EdgeCases/SpecialCharactersInQuery (0.00s)
    --- PASS: TestSearchService_EdgeCases/EmptyTypesFilter (0.00s)
=== RUN   TestBasicFunctionality
--- PASS: TestBasicFunctionality (0.00s)
=== RUN   TestStringOperations
--- PASS: TestStringOperations (0.00s)
=== RUN   TestIntegerOperations
--- PASS: TestIntegerOperations (0.00s)
=== RUN   TestMockService
--- PASS: TestMockService (0.00s)
=== RUN   TestMockService_ConcurrentAccess
--- PASS: TestMockService_ConcurrentAccess (0.00s)
=== RUN   TestMockService_InvalidOperations
--- PASS: TestMockService_InvalidOperations (0.00s)
PASS
ok  	law-oa-go/internal/services	1.566s
=== RUN   TestUserService_CreateUser
=== RUN   TestUserService_CreateUser/Create_User_Success
=== RUN   TestUserService_CreateUser/Create_User_Email_Already_Exists
=== RUN   TestUserService_CreateUser/Create_User_Invalid_Role
=== RUN   TestUserService_CreateUser/Create_User_Weak_Password
=== RUN   TestUserService_CreateUser/Create_User_Database_Error
    user_service_test.go:141: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/services/test/user_service_test.go:141
        	Error:      	An error is expected but got nil.
        	Test:       	TestUserService_CreateUser/Create_User_Database_Error
    user_service_test.go:142: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/internal/services/test/user_service_test.go:142
        	Error:      	Expected nil, but got: &services.UserProfile{ID:0x0, Name:"Test User", Email:"test@example.com", Role:"user", Phone:"", Avatar:"", Status:"active", CreatedAt:time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), UpdatedAt:time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)}
        	Test:       	TestUserService_CreateUser/Create_User_Database_Error
--- FAIL: TestUserService_CreateUser (1.37s)
    --- PASS: TestUserService_CreateUser/Create_User_Success (0.69s)
    --- PASS: TestUserService_CreateUser/Create_User_Email_Already_Exists (0.00s)
    --- PASS: TestUserService_CreateUser/Create_User_Invalid_Role (0.00s)
    --- PASS: TestUserService_CreateUser/Create_User_Weak_Password (0.00s)
    --- FAIL: TestUserService_CreateUser/Create_User_Database_Error (0.68s)
panic: runtime error: invalid memory address or nil pointer dereference [recovered]
	panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x18 pc=0x5db65d2]

goroutine 9 [running]:
testing.tRunner.func1.2({0x5eae0c0, 0x614ae00})
	/usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1632 +0x3fc
testing.tRunner.func1()
	/usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1635 +0x6b6
panic({0x5eae0c0?, 0x614ae00?})
	/usr/local/Cellar/go/1.23.6/libexec/src/runtime/panic.go:785 +0x132
law-oa-go/internal/services/test_test.TestUserService_CreateUser.func5(0xc0000b9040)
	/Users/mac/Desktop/FT/law-oa-go/internal/services/test/user_service_test.go:143 +0x532
testing.tRunner(0xc0000b9040, 0xc000010360)
	/usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1690 +0x227
created by testing.(*T).Run in goroutine 5
	/usr/local/Cellar/go/1.23.6/libexec/src/testing/testing.go:1743 +0x826
FAIL	law-oa-go/internal/services/test	2.244s
FAIL
[0;31m[ERROR][0m 2025-09-16 09:50:10 单元测试失败
[0;34m[INFO][0m 2025-09-16 09:50:10 运行集成测试...
?   	law-oa-go/tests	[no test files]
# law-oa-go/tests/e2e [law-oa-go/tests/e2e.test]
tests/e2e/workflow_test.go:260:39: too many arguments in call to test.AssertSuccessResponse
	have (*testing.T, *httptest.ResponseRecorder, number)
	want (*testing.T, *httptest.ResponseRecorder)
tests/e2e/workflow_test.go:275:39: too many arguments in call to test.AssertSuccessResponse
	have (*testing.T, *httptest.ResponseRecorder, number)
	want (*testing.T, *httptest.ResponseRecorder)
tests/e2e/workflow_test.go:285:39: too many arguments in call to test.AssertSuccessResponse
	have (*testing.T, *httptest.ResponseRecorder, number)
	want (*testing.T, *httptest.ResponseRecorder)
tests/e2e/workflow_test.go:295:39: too many arguments in call to test.AssertSuccessResponse
	have (*testing.T, *httptest.ResponseRecorder, number)
	want (*testing.T, *httptest.ResponseRecorder)
tests/e2e/workflow_test.go:305:39: too many arguments in call to test.AssertSuccessResponse
	have (*testing.T, *httptest.ResponseRecorder, number)
	want (*testing.T, *httptest.ResponseRecorder)
tests/e2e/workflow_test.go:310:39: too many arguments in call to test.AssertSuccessResponse
	have (*testing.T, *httptest.ResponseRecorder, number)
	want (*testing.T, *httptest.ResponseRecorder)
tests/e2e/workflow_test.go:334:39: too many arguments in call to test.AssertSuccessResponse
	have (*testing.T, *httptest.ResponseRecorder, number)
	want (*testing.T, *httptest.ResponseRecorder)
tests/e2e/workflow_test.go:347:39: too many arguments in call to test.AssertSuccessResponse
	have (*testing.T, *httptest.ResponseRecorder, number)
	want (*testing.T, *httptest.ResponseRecorder)
tests/e2e/workflow_test.go:359:39: too many arguments in call to test.AssertSuccessResponse
	have (*testing.T, *httptest.ResponseRecorder, number)
	want (*testing.T, *httptest.ResponseRecorder)
tests/e2e/workflow_test.go:372:39: too many arguments in call to test.AssertSuccessResponse
	have (*testing.T, *httptest.ResponseRecorder, number)
	want (*testing.T, *httptest.ResponseRecorder)
tests/e2e/workflow_test.go:372:39: too many errors
FAIL	law-oa-go/tests/e2e [build failed]
# law-oa-go/tests/integration.test
ld: warning: '/private/var/folders/4p/bng36r_s65d26yqk0lfpw2rh0000gn/T/go-link-957206267/000023.o' has malformed LC_DYSYMTAB, expected 98 undefined symbols to start at index 884, found 95 undefined symbols starting at index 884
# law-oa-go/tests/performance.test
ld: warning: '/private/var/folders/4p/bng36r_s65d26yqk0lfpw2rh0000gn/T/go-link-597698096/000023.o' has malformed LC_DYSYMTAB, expected 98 undefined symbols to start at index 884, found 95 undefined symbols starting at index 884
=== RUN   TestAuthIntegration
=== RUN   TestAuthIntegration/Complete_Auth_Flow

2025/09/16 09:50:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.956ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "testuser@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.237ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "testuser@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    testutils.go:193: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:193
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/integration/auth_integration_test.go:131
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestAuthIntegration/Complete_Auth_Flow
=== RUN   TestAuthIntegration/Login_with_Different_Roles

2025/09/16 09:50:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.182ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.229ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.177ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:18 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.330ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:19 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.185ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "regular@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:19 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.183ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "regular@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
--- FAIL: TestAuthIntegration (4.87s)
    --- FAIL: TestAuthIntegration/Complete_Auth_Flow (0.70s)
    --- PASS: TestAuthIntegration/Login_with_Different_Roles (4.16s)
=== RUN   TestUserManagementIntegration

2025/09/16 09:50:20 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.073ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:21 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.196ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
=== RUN   TestUserManagementIntegration/Create_and_Manage_Users

2025/09/16 09:50:21 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.133ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "newuser@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:22 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.201ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "newuser@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    auth_integration_test.go:266: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/integration/auth_integration_test.go:266
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestUserManagementIntegration/Create_and_Manage_Users
--- FAIL: TestUserManagementIntegration (1.40s)
    --- FAIL: TestUserManagementIntegration/Create_and_Manage_Users (0.72s)
=== RUN   TestClientManagementIntegration

2025/09/16 09:50:22 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.069ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:22 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.207ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
=== RUN   TestClientManagementIntegration/Create_and_Manage_Clients

2025/09/16 09:50:22 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.120ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "client@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
    auth_integration_test.go:377: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/integration/auth_integration_test.go:377
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestClientManagementIntegration/Create_and_Manage_Clients
=== RUN   TestClientManagementIntegration/Client_Search_and_Filter

2025/09/16 09:50:22 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.077ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "abc@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
    auth_integration_test.go:446: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/integration/auth_integration_test.go:446
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestClientManagementIntegration/Client_Search_and_Filter

2025/09/16 09:50:22 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.052ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "xyz@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
    auth_integration_test.go:446: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/integration/auth_integration_test.go:446
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestClientManagementIntegration/Client_Search_and_Filter

2025/09/16 09:50:22 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.053ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "test@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
    auth_integration_test.go:446: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/integration/auth_integration_test.go:446
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestClientManagementIntegration/Client_Search_and_Filter
--- FAIL: TestClientManagementIntegration (0.70s)
    --- FAIL: TestClientManagementIntegration/Create_and_Manage_Clients (0.00s)
    --- FAIL: TestClientManagementIntegration/Client_Search_and_Filter (0.00s)
=== RUN   TestCaseManagementIntegration

2025/09/16 09:50:22 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.068ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:23 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.188ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:23 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.069ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "client@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
    auth_integration_test.go:517: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/integration/auth_integration_test.go:517
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestCaseManagementIntegration
=== RUN   TestCaseManagementIntegration/Create_and_Manage_Cases
    auth_integration_test.go:538: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/integration/auth_integration_test.go:538
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestCaseManagementIntegration/Create_and_Manage_Cases
    testutils.go:193: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:193
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/integration/auth_integration_test.go:539
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestCaseManagementIntegration/Create_and_Manage_Cases
--- FAIL: TestCaseManagementIntegration (0.70s)
    --- FAIL: TestCaseManagementIntegration/Create_and_Manage_Cases (0.00s)
=== RUN   TestAuthorizationIntegration

2025/09/16 09:50:23 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.079ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:24 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.180ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:24 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.064ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:24 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.192ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:24 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.069ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:25 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.184ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
=== RUN   TestAuthorizationIntegration/Role-Based_Access_Control

2025/09/16 09:50:25 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.100ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "newuser@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:26 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.184ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "newuser@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    auth_integration_test.go:655: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/integration/auth_integration_test.go:655
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestAuthorizationIntegration/Role-Based_Access_Control
    auth_integration_test.go:665: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/integration/auth_integration_test.go:665
        	Error:      	Not equal: 
        	            	expected: 403
        	            	actual  : 200
        	Test:       	TestAuthorizationIntegration/Role-Based_Access_Control
    auth_integration_test.go:675: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/integration/auth_integration_test.go:675
        	Error:      	Not equal: 
        	            	expected: 403
        	            	actual  : 200
        	Test:       	TestAuthorizationIntegration/Role-Based_Access_Control
=== RUN   TestAuthorizationIntegration/Resource_Access_Control
    auth_integration_test.go:706: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/integration/auth_integration_test.go:706
        	Error:      	Not equal: 
        	            	expected: 403
        	            	actual  : 200
        	Test:       	TestAuthorizationIntegration/Resource_Access_Control
--- FAIL: TestAuthorizationIntegration (2.71s)
    --- FAIL: TestAuthorizationIntegration/Role-Based_Access_Control (0.67s)
    --- FAIL: TestAuthorizationIntegration/Resource_Access_Control (0.00s)
=== RUN   TestDatabaseIntegration
=== RUN   TestDatabaseIntegration/User_CRUD_Operations

2025/09/16 09:50:26 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.590ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "test@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:26 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.202ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "test@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:26 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/base_repository.go:32 [35;1mrecord not found
[0m[33m[0.045ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE `users`.`id` = 1 AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
=== RUN   TestDatabaseIntegration/Client_CRUD_Operations

2025/09/16 09:50:26 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.096ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "client@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 09:50:26 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:31 [35;1mrecord not found
[0m[33m[0.041ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE `clients`.`id` = 1 AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
=== RUN   TestDatabaseIntegration/Case_CRUD_Operations

2025/09/16 09:50:26 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.117ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:27 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.231ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:27 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.094ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "client@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 09:50:27 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:25 [35;1mUNIQUE constraint failed: clients.email
[0m[33m[0.351ms] [34;1m[rows:0][0m INSERT INTO `clients` (`created_at`,`updated_at`,`deleted_at`,`name`,`email`,`phone`,`address`,`company`,`notes`,`status`) VALUES ("2025-09-16 09:50:27.422","2025-09-16 09:50:27.422",NULL,"Test Client","client@example.com","1234567890","123 Test St","","","active") RETURNING `id`
    database_integration_test.go:170: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/integration/database_integration_test.go:170
        	Error:      	Received unexpected error:
        	            	Failed to create client: UNIQUE constraint failed: clients.email
        	Test:       	TestDatabaseIntegration/Case_CRUD_Operations
=== RUN   TestDatabaseIntegration/Search_and_Filter_Operations

2025/09/16 09:50:27 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.087ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "alice@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:28 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.216ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "alice@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:28 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.086ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "bob@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:28 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.191ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "bob@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:28 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.061ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "alice2@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:29 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.203ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "alice2@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:29 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.061ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "charlie@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:30 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.202ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "charlie@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    database_integration_test.go:267: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/integration/database_integration_test.go:267
        	Error:      	Not equal: 
        	            	expected: 1
        	            	actual  : 2
        	Test:       	TestDatabaseIntegration/Search_and_Filter_Operations

2025/09/16 09:50:30 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.059ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "abc@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 09:50:30 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.070ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "xyz@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 09:50:30 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.089ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "abc2@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
    database_integration_test.go:303: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/integration/database_integration_test.go:303
        	Error:      	Not equal: 
        	            	expected: 1
        	            	actual  : 3
        	Test:       	TestDatabaseIntegration/Search_and_Filter_Operations
    database_integration_test.go:304: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/integration/database_integration_test.go:304
        	Error:      	Not equal: 
        	            	expected: "ABC Corporation"
        	            	actual  : "ABC Limited"
        	            	
        	            	Diff:
        	            	--- Expected
        	            	+++ Actual
        	            	@@ -1 +1 @@
        	            	-ABC Corporation
        	            	+ABC Limited
        	Test:       	TestDatabaseIntegration/Search_and_Filter_Operations
=== RUN   TestDatabaseIntegration/Concurrent_Operations

2025/09/16 09:50:30 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.147ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:30 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.093ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:30 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.113ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:30 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.310ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:30 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.094ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:30 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.114ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:30 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[2.567ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:30 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[3.704ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:30 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[3.636ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:30 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[5.610ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:30 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.165ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    database_integration_test.go:343: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/integration/database_integration_test.go:343
        	Error:      	Received unexpected error:
        	            	Batch create users failed: circuit breaker default is open
        	Test:       	TestDatabaseIntegration/Concurrent_Operations
=== RUN   TestDatabaseIntegration/Transaction_Rollback

2025/09/16 09:50:30 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.528ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "transaction@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:31 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.176ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "transaction@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:31 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.064ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "transactionclient@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 09:50:31 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/services/case_service.go:335 [35;1mrecord not found
[0m[33m[0.063ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE (id = 0 AND role = "lawyer") AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    database_integration_test.go:378: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/integration/database_integration_test.go:378
        	Error:      	Received unexpected error:
        	            	lawyer not found
        	Test:       	TestDatabaseIntegration/Transaction_Rollback
--- FAIL: TestDatabaseIntegration (5.35s)
    --- PASS: TestDatabaseIntegration/User_CRUD_Operations (0.63s)
    --- PASS: TestDatabaseIntegration/Client_CRUD_Operations (0.00s)
    --- FAIL: TestDatabaseIntegration/Case_CRUD_Operations (0.64s)
    --- FAIL: TestDatabaseIntegration/Search_and_Filter_Operations (2.61s)
    --- FAIL: TestDatabaseIntegration/Concurrent_Operations (0.81s)
    --- FAIL: TestDatabaseIntegration/Transaction_Rollback (0.64s)
FAIL
FAIL	law-oa-go/tests/integration	16.757s
=== RUN   TestLoginLoad

2025/09/16 09:50:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[2.186ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[2.626ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[2.979ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[7.536ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[2.871ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.079ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.077ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.066ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.065ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.065ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.072ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.067ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[5.402ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.727ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[5.734ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.089ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[5.700ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 09:50:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[4.995ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1



2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[1.515ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[1.518ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.063ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.319ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.184ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[3.536ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[1.975ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.241ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.102ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.674ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.055ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.197ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.481ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[2.640ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[2.576ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

=== Login Load Test Results ===
Duration: 10s
Workers: 10
Total Requests: 94037
Successful: 94037
Failed: 0
Error Rate: 0.00%
Average Response Time: 1.062521ms
Min Response Time: 114.042µs
Max Response Time: 70.087042ms
Requests per Second: 9403.70
--- PASS: TestLoginLoad (10.03s)
=== RUN   TestUserProfileLoad

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.074ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "loadtest@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:90
        	Error:      	Received unexpected error:
        	            	User not found: loadtest@example.com
        	Test:       	TestUserProfileLoad
--- FAIL: TestUserProfileLoad (0.03s)
=== RUN   TestClientOperationsLoad

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.215ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "clientops@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:121
        	Error:      	Received unexpected error:
        	            	User not found: clientops@example.com
        	Test:       	TestClientOperationsLoad
--- FAIL: TestClientOperationsLoad (0.02s)
=== RUN   TestConcurrentUsers

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.121ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent49@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.182ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent2@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.214ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent20@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.085ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent14@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.104ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent16@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.137ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent12@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.130ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent4@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.359ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent26@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent12@example.com
        	Test:       	TestConcurrentUsers

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.289ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent3@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.086ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent37@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.545ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent19@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent37@example.com
        	Test:       	TestConcurrentUsers

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.377ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent22@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.093ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent23@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.184ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent15@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent26@example.com
        	Test:       	TestConcurrentUsers

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.187ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent25@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.529ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent43@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent22@example.com
        	Test:       	TestConcurrentUsers

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.089ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent24@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.164ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent38@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.070ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent17@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.110ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent11@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.220ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent21@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.079ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent44@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.110ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent27@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.349ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent47@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.414ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent13@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent4@example.com
        	Test:       	TestConcurrentUsers

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[1.661ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent0@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[1.179ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent48@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.290ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent29@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.912ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent39@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.128ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent8@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.479ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent46@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.844ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent18@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.214ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent30@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent19@example.com
        	Test:       	TestConcurrentUsers

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.388ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent7@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.994ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent45@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent18@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent45@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent7@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent23@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent25@example.com
        	Test:       	TestConcurrentUsers

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.098ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent33@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.120ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent10@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.090ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent41@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.068ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent42@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.062ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent34@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.060ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent6@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.100ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent5@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.069ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent31@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.064ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent32@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.252ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent1@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.068ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent36@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.402ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent35@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.088ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent28@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent46@example.com
        	Test:       	TestConcurrentUsers

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.074ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent9@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent30@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent28@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent35@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent34@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent24@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent10@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent8@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent9@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent11@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent42@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent6@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent31@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent32@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent5@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent1@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent2@example.com
        	Test:       	TestConcurrentUsers

2025/09/16 10:16:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[2.229ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "concurrent40@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent17@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent41@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent21@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent20@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent13@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent49@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent44@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent15@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent47@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent27@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent38@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent33@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent36@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent14@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent29@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent48@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent3@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent39@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent43@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent0@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent16@example.com
        	Test:       	TestConcurrentUsers
    testutils.go:232: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/test/testutils.go:232
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/performance/load_test.go:191
        	            				/usr/local/Cellar/go/1.23.6/libexec/src/runtime/asm_amd64.s:1700
        	Error:      	Received unexpected error:
        	            	User not found: concurrent40@example.com
        	Test:       	TestConcurrentUsers

=== Concurrent Users Test Results ===
Total Users: 50
Requests per User: 100
Total Test Duration: 30s
All users completed successfully
--- FAIL: TestConcurrentUsers (0.04s)
FAIL
FAIL	law-oa-go/tests/performance	11.370s
FAIL
[0;31m[ERROR][0m 2025-09-16 10:16:42 集成测试失败
[0;34m[INFO][0m 2025-09-16 10:16:42 运行端到端测试...
scripts/run_tests.sh: line 295: docker-compose: command not found
[0;34m[INFO][0m 2025-09-16 10:16:42 等待测试环境就绪...
=== RUN   TestCompleteUserWorkflow

2025/09/16 10:17:10 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.674ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:10 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.145ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:10 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.138ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:10 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.092ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:10 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.117ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:11 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.120ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
=== RUN   TestCompleteUserWorkflow/New_User_Onboarding

2025/09/16 10:17:11 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.090ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "john.doe@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:11 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.119ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "john.doe@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:260
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestCompleteUserWorkflow/New_User_Onboarding
    test_helpers.go:252: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:252
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:275
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestCompleteUserWorkflow/New_User_Onboarding
--- FAIL: TestCompleteUserWorkflow (0.52s)
    --- FAIL: TestCompleteUserWorkflow/New_User_Onboarding (0.07s)
=== RUN   TestLawyerClientManagementWorkflow

2025/09/16 10:17:11 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.062ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:11 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.107ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:11 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.116ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:11 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.128ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:11 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.086ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:11 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.102ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
=== RUN   TestLawyerClientManagementWorkflow/Client_Management_Lifecycle

2025/09/16 10:17:11 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.104ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "contact@acme.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
    workflow_test.go:333: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:333
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestLawyerClientManagementWorkflow/Client_Management_Lifecycle
    test_helpers.go:248: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:248
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:334
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestLawyerClientManagementWorkflow/Client_Management_Lifecycle
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:334
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestLawyerClientManagementWorkflow/Client_Management_Lifecycle
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:347
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestLawyerClientManagementWorkflow/Client_Management_Lifecycle
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:359
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestLawyerClientManagementWorkflow/Client_Management_Lifecycle
    workflow_test.go:371: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:371
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestLawyerClientManagementWorkflow/Client_Management_Lifecycle
    test_helpers.go:248: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:248
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:372
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestLawyerClientManagementWorkflow/Client_Management_Lifecycle
    test_helpers.go:252: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:252
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:372
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestLawyerClientManagementWorkflow/Client_Management_Lifecycle
--- FAIL: TestLawyerClientManagementWorkflow (0.44s)
    --- FAIL: TestLawyerClientManagementWorkflow/Client_Management_Lifecycle (0.00s)
=== RUN   TestAdminUserManagementWorkflow

2025/09/16 10:17:11 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.025ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:11 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.124ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:11 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.101ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:11 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.102ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:11 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.076ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:11 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.125ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
=== RUN   TestAdminUserManagementWorkflow/User_Management_Operations
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:426
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestAdminUserManagementWorkflow/User_Management_Operations
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:431
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestAdminUserManagementWorkflow/User_Management_Operations
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:436
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestAdminUserManagementWorkflow/User_Management_Operations
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:441
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestAdminUserManagementWorkflow/User_Management_Operations
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:450
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestAdminUserManagementWorkflow/User_Management_Operations
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:455
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestAdminUserManagementWorkflow/User_Management_Operations
--- FAIL: TestAdminUserManagementWorkflow (0.44s)
    --- FAIL: TestAdminUserManagementWorkflow/User_Management_Operations (0.01s)
=== RUN   TestCaseManagementWorkflow

2025/09/16 10:17:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.044ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.126ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.119ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.133ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.101ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.101ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
=== RUN   TestCaseManagementWorkflow/Case_Management_Lifecycle

2025/09/16 10:17:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.086ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "legal@techstart.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
    workflow_test.go:486: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:486
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestCaseManagementWorkflow/Case_Management_Lifecycle

2025/09/16 10:17:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.027ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "legal@globalcorp.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
    workflow_test.go:486: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:486
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestCaseManagementWorkflow/Case_Management_Lifecycle
    workflow_test.go:524: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:524
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestCaseManagementWorkflow/Case_Management_Lifecycle
    workflow_test.go:528: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:528
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestCaseManagementWorkflow/Case_Management_Lifecycle
--- FAIL: TestCaseManagementWorkflow (0.44s)
    --- FAIL: TestCaseManagementWorkflow/Case_Management_Lifecycle (0.00s)
=== RUN   TestAuthorizationAndSecurity

2025/09/16 10:17:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.023ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.116ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.103ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.131ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.117ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.148ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
=== RUN   TestAuthorizationAndSecurity/Role-Based_Access_Control
    workflow_test.go:611: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:611
        	Error:      	Not equal: 
        	            	expected: 403
        	            	actual  : 200
        	Test:       	TestAuthorizationAndSecurity/Role-Based_Access_Control

2025/09/16 10:17:12 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.036ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "test@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:13 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.127ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "test@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    workflow_test.go:622: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:622
        	Error:      	Not equal: 
        	            	expected: 403
        	            	actual  : 200
        	Test:       	TestAuthorizationAndSecurity/Role-Based_Access_Control
    workflow_test.go:633: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:633
        	Error:      	Not equal: 
        	            	expected: 403
        	            	actual  : 200
        	Test:       	TestAuthorizationAndSecurity/Role-Based_Access_Control
    workflow_test.go:644: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:644
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestAuthorizationAndSecurity/Role-Based_Access_Control
=== RUN   TestAuthorizationAndSecurity/Unauthenticated_Access
    workflow_test.go:660: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:660
        	Error:      	Not equal: 
        	            	expected: 401
        	            	actual  : 200
        	Test:       	TestAuthorizationAndSecurity/Unauthenticated_Access
    workflow_test.go:660: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:660
        	Error:      	Not equal: 
        	            	expected: 401
        	            	actual  : 200
        	Test:       	TestAuthorizationAndSecurity/Unauthenticated_Access
    workflow_test.go:660: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:660
        	Error:      	Not equal: 
        	            	expected: 401
        	            	actual  : 200
        	Test:       	TestAuthorizationAndSecurity/Unauthenticated_Access
    workflow_test.go:660: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:660
        	Error:      	Not equal: 
        	            	expected: 401
        	            	actual  : 200
        	Test:       	TestAuthorizationAndSecurity/Unauthenticated_Access
=== RUN   TestAuthorizationAndSecurity/Invalid_Authentication
    workflow_test.go:671: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:671
        	Error:      	Not equal: 
        	            	expected: 401
        	            	actual  : 200
        	Test:       	TestAuthorizationAndSecurity/Invalid_Authentication
--- FAIL: TestAuthorizationAndSecurity (0.55s)
    --- FAIL: TestAuthorizationAndSecurity/Role-Based_Access_Control (0.07s)
    --- FAIL: TestAuthorizationAndSecurity/Unauthenticated_Access (0.00s)
    --- FAIL: TestAuthorizationAndSecurity/Invalid_Authentication (0.00s)
=== RUN   TestAPIErrorHandling

2025/09/16 10:17:13 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.028ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:13 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.149ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:13 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.098ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:13 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.136ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:13 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.121ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:13 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.091ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
=== RUN   TestAPIErrorHandling/Invalid_Input_Handling
    workflow_test.go:695: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:695
        	Error:      	Not equal: 
        	            	expected: 400
        	            	actual  : 200
        	Test:       	TestAPIErrorHandling/Invalid_Input_Handling
    workflow_test.go:704: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:704
        	Error:      	Not equal: 
        	            	expected: 400
        	            	actual  : 200
        	Test:       	TestAPIErrorHandling/Invalid_Input_Handling
    workflow_test.go:713: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:713
        	Error:      	Not equal: 
        	            	expected: 400
        	            	actual  : 200
        	Test:       	TestAPIErrorHandling/Invalid_Input_Handling
=== RUN   TestAPIErrorHandling/Resource_Not_Found

2025/09/16 10:17:13 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/base_repository.go:32 [35;1mrecord not found
[0m[33m[0.075ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE `users`.`id` = 99999 AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    workflow_test.go:721: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:721
        	Error:      	Not equal: 
        	            	expected: 404
        	            	actual  : 200
        	Test:       	TestAPIErrorHandling/Resource_Not_Found

2025/09/16 10:17:13 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:31 [35;1mrecord not found
[0m[33m[0.028ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE `clients`.`id` = 99999 AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
    workflow_test.go:725: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:725
        	Error:      	Not equal: 
        	            	expected: 404
        	            	actual  : 200
        	Test:       	TestAPIErrorHandling/Resource_Not_Found

2025/09/16 10:17:13 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/services/case_service.go:106 [35;1mrecord not found
[0m[33m[0.023ms] [34;1m[rows:0][0m SELECT * FROM `cases` WHERE `cases`.`id` = 99999 AND `cases`.`deleted_at` IS NULL ORDER BY `cases`.`id` LIMIT 1
    workflow_test.go:729: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:729
        	Error:      	Not equal: 
        	            	expected: 404
        	            	actual  : 200
        	Test:       	TestAPIErrorHandling/Resource_Not_Found
=== RUN   TestAPIErrorHandling/Duplicate_Resource_Creation

2025/09/16 10:17:13 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.030ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "duplicate@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
    workflow_test.go:742: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:742
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestAPIErrorHandling/Duplicate_Resource_Creation
    workflow_test.go:746: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:746
        	Error:      	Not equal: 
        	            	expected: 409
        	            	actual  : 200
        	Test:       	TestAPIErrorHandling/Duplicate_Resource_Creation
--- FAIL: TestAPIErrorHandling (0.44s)
    --- FAIL: TestAPIErrorHandling/Invalid_Input_Handling (0.00s)
    --- FAIL: TestAPIErrorHandling/Resource_Not_Found (0.00s)
    --- FAIL: TestAPIErrorHandling/Duplicate_Resource_Creation (0.00s)
=== RUN   TestAPIPerformance

2025/09/16 10:17:13 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.026ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:13 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.129ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:13 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.093ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:13 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.500ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:13 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.118ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:13 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.145ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
=== RUN   TestAPIPerformance/Response_Time_Testing
    workflow_test.go:765: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:765
        	Error:      	Not equal: 
        	            	expected: 200
        	            	actual  : 404
        	Test:       	TestAPIPerformance/Response_Time_Testing
=== RUN   TestAPIPerformance/Concurrent_Request_Handling

2025/09/16 10:17:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.095ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent6@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 10:17:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.535ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent1@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 10:17:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.165ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent3@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 10:17:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.268ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent9@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 10:17:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.915ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent2@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 10:17:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mdatabase table is locked: clients
[0m[33m[3.567ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent8@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 10:17:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mdatabase table is locked: clients
[0m[33m[3.750ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent5@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 10:17:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mdatabase table is locked: clients
[0m[33m[4.096ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent7@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 10:17:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mdatabase table is locked: clients
[0m[33m[3.776ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent4@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 10:17:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mdatabase table is locked: clients
[0m[33m[4.161ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent10@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 10:17:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:25 [35;1mdatabase table is locked
[0m[33m[3.745ms] [34;1m[rows:0][0m INSERT INTO `clients` (`created_at`,`updated_at`,`deleted_at`,`name`,`email`,`phone`,`address`,`company`,`notes`,`status`) VALUES ("2025-09-16 10:17:14.002","2025-09-16 10:17:14.002",NULL,"Concurrent Client 9","concurrent9@example.com","5550000009","","Company 9","","active") RETURNING `id`

2025/09/16 10:17:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:25 [35;1mdatabase table is locked
[0m[33m[3.266ms] [34;1m[rows:0][0m INSERT INTO `clients` (`created_at`,`updated_at`,`deleted_at`,`name`,`email`,`phone`,`address`,`company`,`notes`,`status`) VALUES ("2025-09-16 10:17:14.005","2025-09-16 10:17:14.005",NULL,"Concurrent Client 2","concurrent2@example.com","5550000002","","Company 2","","active") RETURNING `id`
    workflow_test.go:809: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:809
        	Error:      	Not equal: 
        	            	expected: 10
        	            	actual  : 0
        	Test:       	TestAPIPerformance/Concurrent_Request_Handling
        	Messages:   	All concurrent requests should succeed
--- FAIL: TestAPIPerformance (0.54s)
    --- FAIL: TestAPIPerformance/Response_Time_Testing (0.00s)
    --- FAIL: TestAPIPerformance/Concurrent_Request_Handling (0.01s)
FAIL
FAIL	law-oa-go/tests/e2e	4.081s
FAIL
[0;31m[ERROR][0m 2025-09-16 10:17:14 端到端测试失败
scripts/run_tests.sh: line 308: docker-compose: command not found
[0;34m[INFO][0m 2025-09-16 10:17:14 运行性能测试...
=== RUN   TestCompleteUserWorkflow

2025/09/16 10:17:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.685ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.214ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.128ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:14 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.156ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.133ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.149ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
=== RUN   TestCompleteUserWorkflow/New_User_Onboarding

2025/09/16 10:17:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.101ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "john.doe@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.135ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "john.doe@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:260
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestCompleteUserWorkflow/New_User_Onboarding
    test_helpers.go:252: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:252
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:275
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestCompleteUserWorkflow/New_User_Onboarding
--- FAIL: TestCompleteUserWorkflow (0.52s)
    --- FAIL: TestCompleteUserWorkflow/New_User_Onboarding (0.07s)
=== RUN   TestLawyerClientManagementWorkflow

2025/09/16 10:17:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.032ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.169ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.135ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.091ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.130ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.140ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
=== RUN   TestLawyerClientManagementWorkflow/Client_Management_Lifecycle

2025/09/16 10:17:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.135ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "contact@acme.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
    workflow_test.go:333: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:333
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestLawyerClientManagementWorkflow/Client_Management_Lifecycle
    test_helpers.go:248: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:248
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:334
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestLawyerClientManagementWorkflow/Client_Management_Lifecycle
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:334
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestLawyerClientManagementWorkflow/Client_Management_Lifecycle
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:347
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestLawyerClientManagementWorkflow/Client_Management_Lifecycle
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:359
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestLawyerClientManagementWorkflow/Client_Management_Lifecycle
    workflow_test.go:371: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:371
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestLawyerClientManagementWorkflow/Client_Management_Lifecycle
    test_helpers.go:248: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:248
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:372
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestLawyerClientManagementWorkflow/Client_Management_Lifecycle
    test_helpers.go:252: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:252
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:372
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestLawyerClientManagementWorkflow/Client_Management_Lifecycle
--- FAIL: TestLawyerClientManagementWorkflow (0.44s)
    --- FAIL: TestLawyerClientManagementWorkflow/Client_Management_Lifecycle (0.00s)
=== RUN   TestAdminUserManagementWorkflow

2025/09/16 10:17:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.024ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.097ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.142ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.120ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:15 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.105ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.081ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
=== RUN   TestAdminUserManagementWorkflow/User_Management_Operations
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:426
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestAdminUserManagementWorkflow/User_Management_Operations
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:431
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestAdminUserManagementWorkflow/User_Management_Operations
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:436
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestAdminUserManagementWorkflow/User_Management_Operations
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:441
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestAdminUserManagementWorkflow/User_Management_Operations
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:450
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestAdminUserManagementWorkflow/User_Management_Operations
    test_helpers.go:254: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/test_helpers.go:254
        	            				/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:455
        	Error:      	Not equal: 
        	            	expected: float64(0)
        	            	actual  : <nil>(<nil>)
        	Test:       	TestAdminUserManagementWorkflow/User_Management_Operations
--- FAIL: TestAdminUserManagementWorkflow (0.44s)
    --- FAIL: TestAdminUserManagementWorkflow/User_Management_Operations (0.00s)
=== RUN   TestCaseManagementWorkflow

2025/09/16 10:17:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.019ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.132ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.127ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.163ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.120ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.145ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
=== RUN   TestCaseManagementWorkflow/Case_Management_Lifecycle

2025/09/16 10:17:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.114ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "legal@techstart.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
    workflow_test.go:486: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:486
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestCaseManagementWorkflow/Case_Management_Lifecycle

2025/09/16 10:17:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.090ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "legal@globalcorp.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
    workflow_test.go:486: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:486
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestCaseManagementWorkflow/Case_Management_Lifecycle
    workflow_test.go:524: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:524
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestCaseManagementWorkflow/Case_Management_Lifecycle
    workflow_test.go:528: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:528
        	Error:      	Received unexpected error:
        	            	unexpected end of JSON input
        	Test:       	TestCaseManagementWorkflow/Case_Management_Lifecycle
--- FAIL: TestCaseManagementWorkflow (0.44s)
    --- FAIL: TestCaseManagementWorkflow/Case_Management_Lifecycle (0.00s)
=== RUN   TestAuthorizationAndSecurity

2025/09/16 10:17:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.032ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.144ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.125ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.130ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.110ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:16 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.144ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
=== RUN   TestAuthorizationAndSecurity/Role-Based_Access_Control
    workflow_test.go:611: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:611
        	Error:      	Not equal: 
        	            	expected: 403
        	            	actual  : 200
        	Test:       	TestAuthorizationAndSecurity/Role-Based_Access_Control

2025/09/16 10:17:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.035ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "test@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.120ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "test@example.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    workflow_test.go:622: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:622
        	Error:      	Not equal: 
        	            	expected: 403
        	            	actual  : 200
        	Test:       	TestAuthorizationAndSecurity/Role-Based_Access_Control
    workflow_test.go:633: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:633
        	Error:      	Not equal: 
        	            	expected: 403
        	            	actual  : 200
        	Test:       	TestAuthorizationAndSecurity/Role-Based_Access_Control
    workflow_test.go:644: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:644
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestAuthorizationAndSecurity/Role-Based_Access_Control
=== RUN   TestAuthorizationAndSecurity/Unauthenticated_Access
    workflow_test.go:660: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:660
        	Error:      	Not equal: 
        	            	expected: 401
        	            	actual  : 200
        	Test:       	TestAuthorizationAndSecurity/Unauthenticated_Access
    workflow_test.go:660: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:660
        	Error:      	Not equal: 
        	            	expected: 401
        	            	actual  : 200
        	Test:       	TestAuthorizationAndSecurity/Unauthenticated_Access
    workflow_test.go:660: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:660
        	Error:      	Not equal: 
        	            	expected: 401
        	            	actual  : 200
        	Test:       	TestAuthorizationAndSecurity/Unauthenticated_Access
    workflow_test.go:660: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:660
        	Error:      	Not equal: 
        	            	expected: 401
        	            	actual  : 200
        	Test:       	TestAuthorizationAndSecurity/Unauthenticated_Access
=== RUN   TestAuthorizationAndSecurity/Invalid_Authentication
    workflow_test.go:671: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:671
        	Error:      	Not equal: 
        	            	expected: 401
        	            	actual  : 200
        	Test:       	TestAuthorizationAndSecurity/Invalid_Authentication
--- FAIL: TestAuthorizationAndSecurity (0.54s)
    --- FAIL: TestAuthorizationAndSecurity/Role-Based_Access_Control (0.08s)
    --- FAIL: TestAuthorizationAndSecurity/Unauthenticated_Access (0.00s)
    --- FAIL: TestAuthorizationAndSecurity/Invalid_Authentication (0.00s)
=== RUN   TestAPIErrorHandling

2025/09/16 10:17:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.033ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.125ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.122ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.120ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.112ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.114ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
=== RUN   TestAPIErrorHandling/Invalid_Input_Handling
    workflow_test.go:695: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:695
        	Error:      	Not equal: 
        	            	expected: 400
        	            	actual  : 200
        	Test:       	TestAPIErrorHandling/Invalid_Input_Handling
    workflow_test.go:704: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:704
        	Error:      	Not equal: 
        	            	expected: 400
        	            	actual  : 200
        	Test:       	TestAPIErrorHandling/Invalid_Input_Handling
    workflow_test.go:713: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:713
        	Error:      	Not equal: 
        	            	expected: 400
        	            	actual  : 200
        	Test:       	TestAPIErrorHandling/Invalid_Input_Handling
=== RUN   TestAPIErrorHandling/Resource_Not_Found

2025/09/16 10:17:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/base_repository.go:32 [35;1mrecord not found
[0m[33m[0.064ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE `users`.`id` = 99999 AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
    workflow_test.go:721: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:721
        	Error:      	Not equal: 
        	            	expected: 404
        	            	actual  : 200
        	Test:       	TestAPIErrorHandling/Resource_Not_Found

2025/09/16 10:17:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:31 [35;1mrecord not found
[0m[33m[0.103ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE `clients`.`id` = 99999 AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
    workflow_test.go:725: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:725
        	Error:      	Not equal: 
        	            	expected: 404
        	            	actual  : 200
        	Test:       	TestAPIErrorHandling/Resource_Not_Found

2025/09/16 10:17:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/services/case_service.go:106 [35;1mrecord not found
[0m[33m[0.034ms] [34;1m[rows:0][0m SELECT * FROM `cases` WHERE `cases`.`id` = 99999 AND `cases`.`deleted_at` IS NULL ORDER BY `cases`.`id` LIMIT 1
    workflow_test.go:729: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:729
        	Error:      	Not equal: 
        	            	expected: 404
        	            	actual  : 200
        	Test:       	TestAPIErrorHandling/Resource_Not_Found
=== RUN   TestAPIErrorHandling/Duplicate_Resource_Creation

2025/09/16 10:17:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.035ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "duplicate@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
    workflow_test.go:742: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:742
        	Error:      	Not equal: 
        	            	expected: 201
        	            	actual  : 200
        	Test:       	TestAPIErrorHandling/Duplicate_Resource_Creation
    workflow_test.go:746: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:746
        	Error:      	Not equal: 
        	            	expected: 409
        	            	actual  : 200
        	Test:       	TestAPIErrorHandling/Duplicate_Resource_Creation
--- FAIL: TestAPIErrorHandling (0.46s)
    --- FAIL: TestAPIErrorHandling/Invalid_Input_Handling (0.00s)
    --- FAIL: TestAPIErrorHandling/Resource_Not_Found (0.00s)
    --- FAIL: TestAPIErrorHandling/Duplicate_Resource_Creation (0.00s)
=== RUN   TestAPIPerformance

2025/09/16 10:17:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.025ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.119ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "admin@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.106ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.128ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "lawyer@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:56 [35;1mrecord not found
[0m[33m[0.122ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1

2025/09/16 10:17:17 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/user_repository.go:30 [35;1mrecord not found
[0m[33m[0.133ms] [34;1m[rows:0][0m SELECT * FROM `users` WHERE email = "user@test.com" AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
=== RUN   TestAPIPerformance/Response_Time_Testing
    workflow_test.go:765: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:765
        	Error:      	Not equal: 
        	            	expected: 200
        	            	actual  : 404
        	Test:       	TestAPIPerformance/Response_Time_Testing
=== RUN   TestAPIPerformance/Concurrent_Request_Handling

2025/09/16 10:17:18 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.052ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent2@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 10:17:18 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.088ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent9@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 10:17:18 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mdatabase table is locked: clients
[0m[33m[1.803ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent4@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 10:17:18 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mdatabase table is locked: clients
[0m[33m[2.174ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent1@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 10:17:18 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.038ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent7@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 10:17:18 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mrecord not found
[0m[33m[0.111ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent5@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 10:17:18 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:25 [35;1mdatabase table is locked
[0m[33m[0.749ms] [34;1m[rows:0][0m INSERT INTO `clients` (`created_at`,`updated_at`,`deleted_at`,`name`,`email`,`phone`,`address`,`company`,`notes`,`status`) VALUES ("2025-09-16 10:17:18.027","2025-09-16 10:17:18.027",NULL,"Concurrent Client 7","concurrent7@example.com","5550000007","","Company 7","","active") RETURNING `id`

2025/09/16 10:17:18 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mdatabase table is locked: clients
[0m[33m[7.673ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent10@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 10:17:18 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mdatabase table is locked: clients
[0m[33m[7.816ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent6@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 10:17:18 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mdatabase table is locked: clients
[0m[33m[4.785ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent8@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1

2025/09/16 10:17:18 [31;1m/Users/mac/Desktop/FT/law-oa-go/internal/repositories/client_repository.go:44 [35;1mdatabase table is locked: clients
[0m[33m[2.127ms] [34;1m[rows:0][0m SELECT * FROM `clients` WHERE email = "concurrent3@example.com" AND `clients`.`deleted_at` IS NULL ORDER BY `clients`.`id` LIMIT 1
    workflow_test.go:809: 
        	Error Trace:	/Users/mac/Desktop/FT/law-oa-go/tests/e2e/workflow_test.go:809
        	Error:      	Not equal: 
        	            	expected: 10
        	            	actual  : 0
        	Test:       	TestAPIPerformance/Concurrent_Request_Handling
        	Messages:   	All concurrent requests should succeed
--- FAIL: TestAPIPerformance (0.46s)
    --- FAIL: TestAPIPerformance/Response_Time_Testing (0.00s)
    --- FAIL: TestAPIPerformance/Concurrent_Request_Handling (0.01s)
FAIL
FAIL	law-oa-go/tests/e2e	4.057s
FAIL
[0;31m[ERROR][0m 2025-09-16 10:17:18 端到端测试失败
scripts/run_tests.sh: line 308: docker-compose: command not found
[0;34m[INFO][0m 2025-09-16 10:17:18 运行性能测试...
[0;31m[ERROR][0m 2025-09-16 10:17:26 性能测试失败
[0;34m[INFO][0m 2025-09-16 10:17:26 运行安全测试...
[0;32m[SUCCESS][0m 2025-09-16 10:17:26 安全测试完成
[0;34m[INFO][0m 2025-09-16 10:17:26 生成测试报告: test_report.md
[0;31m[ERROR][0m 2025-09-16 10:17:30 性能测试失败
[0;34m[INFO][0m 2025-09-16 10:17:30 运行安全测试...
[0;32m[SUCCESS][0m 2025-09-16 10:17:30 安全测试完成
[0;34m[INFO][0m 2025-09-16 10:17:30 生成测试报告: test_report.md

---
*报告生成时间: 2025年 9月16日 星期二 10时17分31秒 CST*

