//go:build ignore

// 综合PGO性能剖析工作负载
// 包含数据库操作、并发处理、HTTP请求等多种工作负载

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	// 导入项目包
	"law-oa-go/internal/concurrency"
	"law-oa-go/internal/config"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
	"law-oa-go/internal/services"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 工作负载类型
type WorkloadType int

const (
	WorkloadTypeHTTP WorkloadType = iota
	WorkloadTypeDatabase
	WorkloadTypeConcurrent
	WorkloadTypeMixed
)

// 综合工作负载配置
type ComprehensiveWorkloadConfig struct {
	// 基本配置
	Duration          time.Duration `json:"duration"`
	ConcurrentWorkers int           `json:"concurrent_workers"`

	// 工作负载权重
	HTTPWeight       float64 `json:"http_weight"`
	DatabaseWeight   float64 `json:"database_weight"`
	ConcurrentWeight float64 `json:"concurrent_weight"`

	// 数据库配置
	DBOperations []string `json:"db_operations"`
	RecordCount  int      `json:"record_count"`

	// 并发配置
	ConcurrentTasks int           `json:"concurrent_tasks"`
	TaskDuration    time.Duration `json:"task_duration"`
}

// 综合性能统计
type ComprehensiveStats struct {
	// 基本统计
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	TotalDuration time.Duration `json:"total_duration"`

	// HTTP统计
	HTTPRequests int64         `json:"http_requests"`
	HTTPSuccess  int64         `json:"http_success"`
	HTTPFailed   int64         `json:"http_failed"`
	HTTPAvgTime  time.Duration `json:"http_avg_time"`

	// 数据库统计
	DBOperations int64         `json:"db_operations"`
	DBSuccess    int64         `json:"db_success"`
	DBFailed     int64         `json:"db_failed"`
	DBAvgTime    time.Duration `json:"db_avg_time"`

	// 并发统计
	ConcurrentTasks   int64         `json:"concurrent_tasks"`
	ConcurrentSuccess int64         `json:"concurrent_success"`
	ConcurrentFailed  int64         `json:"concurrent_failed"`
	ConcurrentAvgTime time.Duration `json:"concurrent_avg_time"`

	// 系统统计
	MemoryUsage    int64 `json:"memory_usage_peak"`
	GoroutineCount int   `json:"goroutine_count_peak"`
}

// 工作负载执行器
type ComprehensiveWorkload struct {
	config       ComprehensiveWorkloadConfig
	stats        *ComprehensiveStats
	mu           sync.RWMutex
	db           *gorm.DB
	userRepo     repositories.UserRepository
	caseRepo     repositories.CaseRepository
	clientRepo   repositories.ClientRepository
	userService  *services.UserService
	caseService  *services.CaseService
	batchService *services.BatchService
	workerPool   *concurrency.WorkerPool
}

func mainComprehensive() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	// 设置工作负载配置
	workloadConfig := ComprehensiveWorkloadConfig{
		Duration:          10 * time.Minute,
		ConcurrentWorkers: 20,
		HTTPWeight:        0.4, // 40% HTTP请求
		DatabaseWeight:    0.4, // 40% 数据库操作
		ConcurrentWeight:  0.2, // 20% 并发处理
		DBOperations:      []string{"create", "read", "update", "delete", "search"},
		RecordCount:       1000,
		ConcurrentTasks:   100,
		TaskDuration:      100 * time.Millisecond,
	}

	log.Printf("开始综合PGO性能剖析工作负载...")
	log.Printf("配置: %+v", workloadConfig)

	// 创建工作负载执行器
	workload, err := NewComprehensiveWorkload(cfg, workloadConfig)
	if err != nil {
		log.Fatalf("创建工作负载失败: %v", err)
	}
	defer workload.Cleanup()

	// 执行工作负载
	err = workload.Run()
	if err != nil {
		log.Fatalf("工作负载执行失败: %v", err)
	}

	// 输出统计结果
	workload.PrintStats()
}

// 创建综合工作负载执行器
func NewComprehensiveWorkload(cfg *config.Config, workloadConfig ComprehensiveWorkloadConfig) (*ComprehensiveWorkload, error) {
	// 初始化数据库
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(&models.User{}, &models.Case{}, &models.Client{})
	if err != nil {
		return nil, fmt.Errorf("数据库迁移失败: %v", err)
	}

	// 创建仓库
	userRepo := repositories.NewUserRepository(db)
	caseRepo := repositories.NewCaseRepository(db)
	clientRepo := repositories.NewClientRepository(db)

	// 创建服务
	userService := services.NewUserService(userRepo)
	caseService := services.NewCaseService(db)
	batchService := services.NewBatchService(db)

	// 创建工作池
	workerPool := concurrency.NewWorkerPool(10, 100, 5*time.Minute)
	workerPool.Start()

	return &ComprehensiveWorkload{
		config:       workloadConfig,
		stats:        &ComprehensiveStats{},
		db:           db,
		userRepo:     userRepo,
		caseRepo:     caseRepo,
		clientRepo:   clientRepo,
		userService:  userService,
		caseService:  caseService,
		batchService: batchService,
		workerPool:   workerPool,
	}, nil
}

// 执行工作负载
func (cw *ComprehensiveWorkload) Run() error {
	log.Println("启动综合工作负载执行...")
	cw.stats.StartTime = time.Now()

	// 创建上下文用于控制执行时间
	ctx, cancel := context.WithTimeout(context.Background(), cw.config.Duration)
	defer cancel()

	// 创建测试数据
	err := cw.createTestData()
	if err != nil {
		return fmt.Errorf("创建测试数据失败: %v", err)
	}

	// 启动统计收集器
	go cw.collectStats(ctx)

	// 启动工作负载
	var wg sync.WaitGroup
	wg.Add(cw.config.ConcurrentWorkers)

	for i := 0; i < cw.config.ConcurrentWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()
			cw.runWorkerWorkload(ctx, workerID)
		}(i)
	}

	// 等待所有工作完成或超时
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("所有工作负载完成")
	case <-ctx.Done():
		log.Println("工作负载超时完成")
	}

	cw.stats.EndTime = time.Now()
	cw.stats.TotalDuration = cw.stats.EndTime.Sub(cw.stats.StartTime)

	return nil
}

// 执行单个工作负载
func (cw *ComprehensiveWorkload) runWorkerWorkload(ctx context.Context, workerID int) {
	log.Printf("工作器 %d 开始工作负载", workerID)

	operationCount := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 选择工作负载类型
			workloadType := cw.selectWorkloadType()

			// 执行对应的工作负载
			switch workloadType {
			case WorkloadTypeHTTP:
				cw.executeHTTPWorkload(ctx, workerID, operationCount)
			case WorkloadTypeDatabase:
				cw.executeDatabaseWorkload(ctx, workerID, operationCount)
			case WorkloadTypeConcurrent:
				cw.executeConcurrentWorkload(ctx, workerID, operationCount)
			case WorkloadTypeMixed:
				cw.executeMixedWorkload(ctx, workerID, operationCount)
			}

			operationCount++

			// 小延迟避免过度消耗CPU
			time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
		}
	}
}

// 执行HTTP工作负载
func (cw *ComprehensiveWorkload) executeHTTPWorkload(ctx context.Context, workerID, operationID int) {
	startTime := time.Now()

	// 模拟HTTP请求处理
	// 这里主要是为了PGO收集HTTP处理相关的性能数据
	request := map[string]interface{}{
		"worker_id":    workerID,
		"operation_id": operationID,
		"timestamp":    time.Now(),
		"data":         generateRandomData(100),
	}

	// 模拟JSON序列化/反序列化（HTTP常见操作）
	_, err := json.Marshal(request)
	if err != nil {
		cw.mu.Lock()
		cw.stats.HTTPFailed++
		cw.mu.Unlock()
		return
	}

	// 模拟响应处理
	time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

	responseTime := time.Since(startTime)
	cw.mu.Lock()
	cw.stats.HTTPRequests++
	cw.stats.HTTPSuccess++
	cw.stats.HTTPAvgTime = (cw.stats.HTTPAvgTime*time.Duration(cw.stats.HTTPSuccess-1) + responseTime) / time.Duration(cw.stats.HTTPSuccess)
	cw.mu.Unlock()
}

// 执行数据库工作负载
func (cw *ComprehensiveWorkload) executeDatabaseWorkload(ctx context.Context, workerID, operationID int) {
	startTime := time.Now()

	// 随机选择数据库操作
	operation := cw.config.DBOperations[rand.Intn(len(cw.config.DBOperations))]

	ctx = context.Background()
	var err error

	switch operation {
	case "create":
		// 创建用户
		user := &models.User{
			Name:  generateRandomName(),
			Email: generateRandomEmail(),
			Role:  "user",
		}
		err = cw.userRepo.Create(ctx, user)

	case "read":
		// 查询用户
		_, err = cw.userRepo.FindByID(ctx, uint(rand.Intn(cw.config.RecordCount)+1))

	case "update":
		// 更新用户
		userID := uint(rand.Intn(cw.config.RecordCount) + 1)
		user, err := cw.userRepo.FindByID(ctx, userID)
		if err == nil {
			user.Name = generateRandomName()
			err = cw.userRepo.Update(ctx, user)
		}

	case "delete":
		// 删除用户（使用随机ID，可能不存在）
		userID := uint(rand.Intn(cw.config.RecordCount*2) + 1) // 增加范围以测试错误处理
		err = cw.userRepo.Delete(ctx, userID)

	case "search":
		// 搜索用户 - 使用List方法
		params := &repositories.UserListParams{
			Page:     1,
			PageSize: 10,
			Search:   generateRandomName(),
		}
		_, _, err = cw.userRepo.List(ctx, params)
	}

	responseTime := time.Since(startTime)
	cw.mu.Lock()
	cw.stats.DBOperations++
	if err == nil {
		cw.stats.DBSuccess++
		cw.stats.DBAvgTime = (cw.stats.DBAvgTime*time.Duration(cw.stats.DBSuccess-1) + responseTime) / time.Duration(cw.stats.DBSuccess)
	} else {
		cw.stats.DBFailed++
	}
	cw.mu.Unlock()
}

// 执行并发工作负载
func (cw *ComprehensiveWorkload) executeConcurrentWorkload(ctx context.Context, workerID, operationID int) {
	startTime := time.Now()

	// 创建并发任务
	taskID := fmt.Sprintf("worker-%d-op-%d", workerID, operationID)

	task := &concurrency.DatabaseTask{
		TaskID:       taskID,
		TaskType:     "database",
		TaskPriority: 1, // 正常优先级
		Operation: func(ctx context.Context) error {
			// 模拟数据库操作
			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
			return nil
		},
		Context: ctx,
	}

	// 提交任务到工作池
	result, err := cw.workerPool.SubmitWithResult(ctx, task)
	if err != nil {
		cw.mu.Lock()
		cw.stats.ConcurrentFailed++
		cw.mu.Unlock()
		return
	}

	// 等待结果
	// 注意：这里简化了处理，实际使用中需要根据具体的并发库实现
	responseTime := time.Since(startTime)
	cw.mu.Lock()
	cw.stats.ConcurrentTasks++
	if result.Error == nil {
		cw.stats.ConcurrentSuccess++
		cw.stats.ConcurrentAvgTime = (cw.stats.ConcurrentAvgTime*time.Duration(cw.stats.ConcurrentSuccess-1) + responseTime) / time.Duration(cw.stats.ConcurrentSuccess)
	} else {
		cw.stats.ConcurrentFailed++
	}
	cw.mu.Unlock()
}

// 执行混合工作负载
func (cw *ComprehensiveWorkload) executeMixedWorkload(ctx context.Context, workerID, operationID int) {
	// 随机选择混合操作
	switch rand.Intn(3) {
	case 0:
		// 批量操作
		cw.executeBatchWorkload(ctx, workerID, operationID)
	case 1:
		// 事务操作
		cw.executeTransactionWorkload(ctx, workerID, operationID)
	case 2:
		// 复杂查询
		cw.executeComplexQueryWorkload(ctx, workerID, operationID)
	}
}

// 执行批量工作负载
func (cw *ComprehensiveWorkload) executeBatchWorkload(ctx context.Context, workerID, operationID int) {
	// 创建批量用户请求
	requests := make([]*services.CreateUserRequest, 10)
	for i := 0; i < 10; i++ {
		requests[i] = &services.CreateUserRequest{
			Name:     generateRandomName(),
			Email:    generateRandomEmail(),
			Password: "Password123!",
			Role:     "user",
		}
	}

	// 执行批量创建
	_, err := cw.userService.BatchCreateUsers(ctx, requests)
	if err != nil {
		log.Printf("批量创建用户失败: %v", err)
	}

	// 更新统计（简化处理）
	cw.mu.Lock()
	cw.stats.DBOperations++
	cw.stats.DBSuccess++
	cw.mu.Unlock()
}

// 执行事务工作负载
func (cw *ComprehensiveWorkload) executeTransactionWorkload(ctx context.Context, workerID, operationID int) {
	ctx = context.Background()

	// 开始事务
	tx := cw.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 执行多个数据库操作
	user := &models.User{
		Name:  generateRandomName(),
		Email: generateRandomEmail(),
		Role:  "user",
	}

	err := tx.Create(user).Error
	if err != nil {
		tx.Rollback()
		return
	}

	// 创建客户
	client := &models.Client{
		Name:    generateRandomName(),
		Email:   generateRandomEmail(),
		Address: generateRandomString(30),
	}

	err = tx.Create(client).Error
	if err != nil {
		tx.Rollback()
		return
	}

	// 创建相关案件
	caseRecord := &models.Case{
		Title:       "测试案件 " + generateRandomString(10),
		Description: generateRandomString(50),
		ClientID:    client.ID,
		LawyerID:    user.ID,
		Status:      "pending",
	}

	err = tx.Create(caseRecord).Error
	if err != nil {
		tx.Rollback()
		return
	}

	// 提交事务
	err = tx.Commit().Error
	if err != nil {
		return
	}
}

// 执行复杂查询工作负载
func (cw *ComprehensiveWorkload) executeComplexQueryWorkload(ctx context.Context, workerID, operationID int) {
	ctx = context.Background()

	// 执行复杂查询（联表查询、聚合、分组等）
	var result []map[string]interface{}

	err := cw.db.Table("users").
		Select("users.*, cases.title as case_title, cases.status as case_status").
		Joins("LEFT JOIN cases ON cases.user_id = users.id").
		Where("users.role = ?", "user").
		Group("users.id").
		Limit(20).
		Offset(rand.Intn(100)).
		Find(&result).Error

	if err != nil {
		log.Printf("复杂查询失败: %v", err)
	}
}

// 选择工作负载类型
func (cw *ComprehensiveWorkload) selectWorkloadType() WorkloadType {
	totalWeight := cw.config.HTTPWeight + cw.config.DatabaseWeight + cw.config.ConcurrentWeight
	random := rand.Float64() * totalWeight

	if random < cw.config.HTTPWeight {
		return WorkloadTypeHTTP
	} else if random < cw.config.HTTPWeight+cw.config.DatabaseWeight {
		return WorkloadTypeDatabase
	} else {
		return WorkloadTypeConcurrent
	}
}

// 创建测试数据
func (cw *ComprehensiveWorkload) createTestData() error {
	log.Println("创建测试数据...")

	// 创建测试用户
	for i := 0; i < cw.config.RecordCount; i++ {
		user := &models.User{
			Name:  generateRandomName(),
			Email: generateRandomEmail(),
			Role:  "user",
		}

		err := cw.userRepo.Create(context.Background(), user)
		if err != nil {
			return fmt.Errorf("创建测试用户失败: %v", err)
		}

		// 每10个用户创建一个案件
		if i%10 == 0 {
			caseRecord := &models.Case{
				Title:       "测试案件 " + generateRandomString(10),
				Description: generateRandomString(50),
				ClientID:    uint(i) + 1, // 使用简单的ID，实际应该创建客户
				LawyerID:    user.ID,
				Status:      []string{"pending", "active", "closed"}[rand.Intn(3)],
			}

			err = cw.caseRepo.Create(context.Background(), caseRecord)
			if err != nil {
				return fmt.Errorf("创建测试案件失败: %v", err)
			}
		}
	}

	log.Printf("创建了 %d 个测试用户", cw.config.RecordCount)
	return nil
}

// 收集统计信息
func (cw *ComprehensiveWorkload) collectStats(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cw.printIntermediateStats()
		}
	}
}

// 打印中间统计
func (cw *ComprehensiveWorkload) printIntermediateStats() {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	log.Printf("中间统计 - HTTP: %d/%d, DB: %d/%d, 并发: %d/%d",
		cw.stats.HTTPSuccess, cw.stats.HTTPRequests,
		cw.stats.DBSuccess, cw.stats.DBOperations,
		cw.stats.ConcurrentSuccess, cw.stats.ConcurrentTasks)
}

// 打印最终统计
func (cw *ComprehensiveWorkload) PrintStats() {
	cw.mu.RLock()
	defer cw.mu.RUnlock()

	fmt.Println("\n=== 综合PGO工作负载性能统计 ===")
	fmt.Printf("总执行时间: %v\n", cw.stats.TotalDuration)
	fmt.Printf("HTTP请求: 成功 %d/%d (成功率: %.2f%%), 平均时间: %v\n",
		cw.stats.HTTPSuccess, cw.stats.HTTPRequests,
		float64(cw.stats.HTTPSuccess)/float64(cw.stats.HTTPRequests)*100,
		cw.stats.HTTPAvgTime)
	fmt.Printf("数据库操作: 成功 %d/%d (成功率: %.2f%%), 平均时间: %v\n",
		cw.stats.DBSuccess, cw.stats.DBOperations,
		float64(cw.stats.DBSuccess)/float64(cw.stats.DBOperations)*100,
		cw.stats.ConcurrentAvgTime)
	fmt.Printf("并发任务: 成功 %d/%d (成功率: %.2f%%), 平均时间: %v\n",
		cw.stats.ConcurrentSuccess, cw.stats.ConcurrentTasks,
		float64(cw.stats.ConcurrentSuccess)/float64(cw.stats.ConcurrentTasks)*100,
		cw.stats.ConcurrentAvgTime)
	fmt.Printf("峰值内存使用: %d MB\n", cw.stats.MemoryUsage/1024/1024)
	fmt.Printf("峰值Goroutine数量: %d\n", cw.stats.GoroutineCount)
	fmt.Println("=================================")

	// 输出JSON格式的统计
	statsJSON, _ := json.MarshalIndent(cw.stats, "", "  ")
	fmt.Printf("\nJSON统计:\n%s\n", statsJSON)
}

// 清理资源
func (cw *ComprehensiveWorkload) Cleanup() {
	if cw.workerPool != nil {
		cw.workerPool.Stop()
	}
}

func mainComprehensiveDetailed() {
	mainComprehensive()
}

// 辅助函数：生成随机名称
func generateRandomName() string {
	return fmt.Sprintf("User_%s", generateRandomString(8))
}

// 辅助函数：生成随机邮箱
func generateRandomEmail() string {
	return fmt.Sprintf("%s@example.com", generateRandomString(10))
}

// 辅助函数：生成随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 辅助函数：生成随机数据
func generateRandomData(size int) map[string]interface{} {
	data := make(map[string]interface{})
	for i := 0; i < size/10; i++ {
		key := generateRandomString(5)
		value := generateRandomString(rand.Intn(20) + 5)
		data[key] = value
	}
	return data
}
