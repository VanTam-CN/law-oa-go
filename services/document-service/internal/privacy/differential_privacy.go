package privacy

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math"
	"math/big"
	"sync"
	"time"
)

// DifferentialPrivacyMechanism 差分隐私机制接口
type DifferentialPrivacyMechanism interface {
	AddNoise(value float64) float64
	GetEpsilon() float64
	GetDelta() float64
	GetSensitivity() float64
}

// LaplaceMechanism 拉普拉斯机制
type LaplaceMechanism struct {
	Epsilon     float64
	Sensitivity float64
}

// NewLaplaceMechanism 创建拉普拉斯机制
func NewLaplaceMechanism(epsilon, sensitivity float64) *LaplaceMechanism {
	return &LaplaceMechanism{
		Epsilon:     epsilon,
		Sensitivity: sensitivity,
	}
}

// AddNoise 添加拉普拉斯噪声
func (lm *LaplaceMechanism) AddNoise(value float64) float64 {
	scale := lm.Sensitivity / lm.Epsilon
	noise := laplaceNoise(scale)
	return value + noise
}

// GetEpsilon 获取epsilon参数
func (lm *LaplaceMechanism) GetEpsilon() float64 {
	return lm.Epsilon
}

// GetDelta 获取delta参数
func (lm *LaplaceMechanism) GetDelta() float64 {
	return 0.0 // 拉普拉斯机制delta为0
}

// GetSensitivity 获取敏感度
func (lm *LaplaceMechanism) GetSensitivity() float64 {
	return lm.Sensitivity
}

// GaussianMechanism 高斯机制
type GaussianMechanism struct {
	Epsilon     float64
	Delta       float64
	Sensitivity float64
}

// NewGaussianMechanism 创建高斯机制
func NewGaussianMechanism(epsilon, delta, sensitivity float64) *GaussianMechanism {
	return &GaussianMechanism{
		Epsilon:     epsilon,
		Delta:       delta,
		Sensitivity: sensitivity,
	}
}

// AddNoise 添加高斯噪声
func (gm *GaussianMechanism) AddNoise(value float64) float64 {
	// 计算高斯分布的标准差
	sigma := gm.calculateSigma()
	noise := gaussianNoise(sigma)
	return value + noise
}

// calculateSigma 计算高斯机制的标准差
func (gm *GaussianMechanism) calculateSigma() float64 {
	// (ε, δ)-差分隐私的高斯机制标准差计算
	// σ = sensitivity * sqrt(2*ln(1.25/δ)) / ε
	return gm.Sensitivity * math.Sqrt(2*math.Log(1.25/gm.Delta)) / gm.Epsilon
}

// GetEpsilon 获取epsilon参数
func (gm *GaussianMechanism) GetEpsilon() float64 {
	return gm.Epsilon
}

// GetDelta 获取delta参数
func (gm *GaussianMechanism) GetDelta() float64 {
	return gm.Delta
}

// GetSensitivity 获取敏感度
func (gm *GaussianMechanism) GetSensitivity() float64 {
	return gm.Sensitivity
}

// DynamicPrivacyBudgetAllocation 动态隐私预算分配
type DynamicPrivacyBudgetAllocation struct {
	TotalEpsilon     float64
	TotalDelta       float64
	UsedEpsilon      float64
	UsedDelta        float64
	NodeHistory      map[string]*NodePrivacyInfo
	ConvergenceRate  float64
	Mutex            sync.RWMutex
}

// NodePrivacyInfo 节点隐私信息
type NodePrivacyInfo struct {
	NodeID            string
	ParticipationCount int64
	LastAccess        time.Time
	DataSensitivity   float64
	AdaptiveEpsilon   float64
	AdaptiveDelta     float64
}

// NewDynamicPrivacyBudgetAllocation 创建动态隐私预算分配器
func NewDynamicPrivacyBudgetAllocation(totalEpsilon, totalDelta float64) *DynamicPrivacyBudgetAllocation {
	return &DynamicPrivacyBudgetAllocation{
		TotalEpsilon:    totalEpsilon,
		TotalDelta:      totalDelta,
		UsedEpsilon:     0.0,
		UsedDelta:       0.0,
		NodeHistory:     make(map[string]*NodePrivacyInfo),
		ConvergenceRate: 1.0,
	}
}

// AllocateBudget 为节点分配隐私预算
func (dpba *DynamicPrivacyBudgetAllocation) AllocateBudget(nodeID string, dataSensitivity float64) (float64, float64, error) {
	dpba.Mutex.Lock()
	defer dpba.Mutex.Unlock()

	// 检查总预算是否足够
	remainingEpsilon := dpba.TotalEpsilon - dpba.UsedEpsilon
	remainingDelta := dpba.TotalDelta - dpba.UsedDelta

	if remainingEpsilon <= 0 || remainingDelta <= 0 {
		return 0, 0, fmt.Errorf("隐私预算耗尽")
	}

	// 获取或创建节点信息
	nodeInfo, exists := dpba.NodeHistory[nodeID]
	if !exists {
		nodeInfo = &NodePrivacyInfo{
			NodeID:           nodeID,
			DataSensitivity:  dataSensitivity,
			AdaptiveEpsilon:  dpba.TotalEpsilon / 100, // 初始分配
			AdaptiveDelta:    dpba.TotalDelta / 100,
		}
		dpba.NodeHistory[nodeID] = nodeInfo
	}

	// 计算自适应预算
	adaptiveEpsilon := dpba.calculateAdaptiveEpsilon(nodeInfo)
	adaptiveDelta := dpba.calculateAdaptiveDelta(nodeInfo)

	// 检查预算是否足够
	if adaptiveEpsilon > remainingEpsilon {
		adaptiveEpsilon = remainingEpsilon
	}
	if adaptiveDelta > remainingDelta {
		adaptiveDelta = remainingDelta
	}

	// 更新使用情况
	dpba.UsedEpsilon += adaptiveEpsilon
	dpba.UsedDelta += adaptiveDelta
	nodeInfo.LastAccess = time.Now()
	nodeInfo.ParticipationCount++

	return adaptiveEpsilon, adaptiveDelta, nil
}

// calculateAdaptiveEpsilon 计算自适应epsilon
func (dpba *DynamicPrivacyBudgetAllocation) calculateAdaptiveEpsilon(nodeInfo *NodePrivacyInfo) float64 {
	// 基于节点数据敏感性、参与度和收敛率计算自适应epsilon
	baseEpsilon := dpba.TotalEpsilon / 100

	// 参与度因子：参与次数越多，分配越少
	participationFactor := 1.0 / math.Sqrt(1.0+float64(nodeInfo.ParticipationCount))

	// 数据敏感性因子：敏感性越高，分配越少
	sensitivityFactor := 1.0 / (1.0 + nodeInfo.DataSensitivity)

	// 收敛率因子：收敛率越高，分配越多
	convergenceFactor := dpba.ConvergenceRate

	adaptiveEpsilon := baseEpsilon * participationFactor * sensitivityFactor * convergenceFactor

	// 确保在合理范围内
	if adaptiveEpsilon < 0.001 {
		adaptiveEpsilon = 0.001
	}
	if adaptiveEpsilon > 1.0 {
		adaptiveEpsilon = 1.0
	}

	nodeInfo.AdaptiveEpsilon = adaptiveEpsilon
	return adaptiveEpsilon
}

// calculateAdaptiveDelta 计算自适应delta
func (dpba *DynamicPrivacyBudgetAllocation) calculateAdaptiveDelta(nodeInfo *NodePrivacyInfo) float64 {
	// Delta通常设置为epsilon的函数
	epsilon := nodeInfo.AdaptiveEpsilon
	delta := math.Min(1e-5, epsilon/1000) // 保守的delta设置

	nodeInfo.AdaptiveDelta = delta
	return delta
}

// GetRemainingBudget 获取剩余预算
func (dpba *DynamicPrivacyBudgetAllocation) GetRemainingBudget() (float64, float64) {
	dpba.Mutex.RLock()
	defer dpba.Mutex.RUnlock()

	return dpba.TotalEpsilon - dpba.UsedEpsilon, dpba.TotalDelta - dpba.UsedDelta
}

// LocalDifferentialPrivacy 本地差分隐私
type LocalDifferentialPrivacy struct {
	Epsilon   float64
	Delta     float64
	Mechanism string // "Laplace", "Gaussian", "Binary"
}

// NewLocalDifferentialPrivacy 创建本地差分隐私处理器
func NewLocalDifferentialPrivacy(epsilon, delta float64, mechanism string) *LocalDifferentialPrivacy {
	return &LocalDifferentialPrivacy{
		Epsilon:   epsilon,
		Delta:     delta,
		Mechanism: mechanism,
	}
}

// ApplyPrivacy 应用本地差分隐私
func (ldp *LocalDifferentialPrivacy) ApplyPrivacy(data []float64) []float64 {
	privateData := make([]float64, len(data))

	for i, value := range data {
		var noise float64
		switch ldp.Mechanism {
		case "Laplace":
			scale := 1.0 / ldp.Epsilon
			noise = laplaceNoise(scale)
		case "Gaussian":
			sigma := math.Sqrt(2*math.Log(1.25/ldp.Delta)) / ldp.Epsilon
			noise = gaussianNoise(sigma)
		case "Binary":
			// 二值随机响应机制
			if math.Abs(value) > 0.5 {
				noise = 1.0
			} else {
				noise = 0.0
			}
		default:
			scale := 1.0 / ldp.Epsilon
			noise = laplaceNoise(scale)
		}
		privateData[i] = value + noise
	}

	return privateData
}

// DifferentialPrivacyService 差分隐私服务
type DifferentialPrivacyService struct {
	budgetAllocator    *DynamicPrivacyBudgetAllocation
	localDP           *LocalDifferentialPrivacy
	logger            *slog.Logger
	mutex             sync.RWMutex
}

// NewDifferentialPrivacyService 创建差分隐私服务
func NewDifferentialPrivacyService(totalEpsilon, totalDelta float64, logger *slog.Logger) *DifferentialPrivacyService {
	return &DifferentialPrivacyService{
		budgetAllocator: NewDynamicPrivacyBudgetAllocation(totalEpsilon, totalDelta),
		localDP:        NewLocalDifferentialPrivacy(1.0, 1e-5, "Laplace"),
		logger:         logger,
	}
}

// PrivateQuery 私有查询
func (dps *DifferentialPrivacyService) PrivateQuery(ctx context.Context, query string, nodeID string, dataSensitivity float64) (float64, error) {
	dps.mutex.Lock()
	defer dps.mutex.Unlock()

	// 分配隐私预算
	epsilon, delta, err := dps.budgetAllocator.AllocateBudget(nodeID, dataSensitivity)
	if err != nil {
		dps.logger.Error("隐私预算分配失败", "error", err, "nodeID", nodeID)
		return 0, err
	}

	// 执行查询（这里简化为返回模拟值）
	queryResult := dps.executeQuery(query)

	// 应用差分隐私噪声
	mechanism := NewLaplaceMechanism(epsilon, dataSensitivity)
	privateResult := mechanism.AddNoise(queryResult)

	dps.logger.Info("私有查询完成",
		"query", query,
		"nodeID", nodeID,
		"epsilon", epsilon,
		"delta", delta,
		"originalResult", queryResult,
		"privateResult", privateResult,
	)

	return privateResult, nil
}

// executeQuery 执行查询（简化实现）
func (dps *DifferentialPrivacyService) executeQuery(query string) float64 {
	// 这里应该是实际的查询逻辑
	// 为了演示，返回模拟值
	switch query {
	case "count_documents":
		return 1000.0
	case "average_case_value":
		return 500000.0
	case "client_satisfaction":
		return 4.5
	default:
		return 100.0
	}
}

// BatchPrivateQuery 批量私有查询
func (dps *DifferentialPrivacyService) BatchPrivateQuery(ctx context.Context, queries []QueryRequest) ([]QueryResponse, error) {
	responses := make([]QueryResponse, len(queries))

	for i, query := range queries {
		result, err := dps.PrivateQuery(ctx, query.Query, query.NodeID, query.DataSensitivity)
		if err != nil {
			responses[i] = QueryResponse{
				QueryID: query.QueryID,
				Error:   err.Error(),
			}
		} else {
			responses[i] = QueryResponse{
				QueryID: query.QueryID,
				Result:  result,
			}
		}
	}

	return responses, nil
}

// GetPrivacyReport 获取隐私报告
func (dps *DifferentialPrivacyService) GetPrivacyReport() map[string]interface{} {
	dps.mutex.RLock()
	defer dps.mutex.RUnlock()

	remainingEpsilon, remainingDelta := dps.budgetAllocator.GetRemainingBudget()

	return map[string]interface{}{
		"total_epsilon":        dps.budgetAllocator.TotalEpsilon,
		"total_delta":          dps.budgetAllocator.TotalDelta,
		"used_epsilon":         dps.budgetAllocator.UsedEpsilon,
		"used_delta":           dps.budgetAllocator.UsedDelta,
		"remaining_epsilon":    remainingEpsilon,
		"remaining_delta":      remainingDelta,
		"budget_usage_rate":    dps.budgetAllocator.UsedEpsilon / dps.budgetAllocator.TotalEpsilon,
		"active_nodes":         len(dps.budgetAllocator.NodeHistory),
		"convergence_rate":     dps.budgetAllocator.ConvergenceRate,
	}
}

// QueryRequest 查询请求
type QueryRequest struct {
	QueryID         string
	Query           string
	NodeID          string
	DataSensitivity float64
}

// QueryResponse 查询响应
type QueryResponse struct {
	QueryID string
	Result  float64
	Error   string
}

// 辅助函数

// laplaceNoise 生成拉普拉斯噪声
func laplaceNoise(scale float64) float64 {
	// 使用指数分布生成拉普拉斯噪声
	// Laplace(0, scale) = Exponential(1/scale) - Exponential(1/scale)
	u1, _ := rand.Float64()
	u2, _ := rand.Float64()

	if u1 < 0.5 {
		return scale * math.Log(2*u2)
	}
	return -scale * math.Log(2*(1-u2))
}

// gaussianNoise 生成高斯噪声（Box-Muller变换）
func gaussianNoise(sigma float64) float64 {
	// Box-Muller变换生成正态分布
	u1, _ := rand.Float64()
	u2, _ := rand.Float64()

	z0 := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
	return z0 * sigma
}

// PrivacyBudgetManager 隐私预算管理器
type PrivacyBudgetManager struct {
	TotalEpsilon    float64
	TotalDelta      float64
	Queries         []QueryRecord
	RollingWindow   time.Duration
	mutex           sync.RWMutex
}

// QueryRecord 查询记录
type QueryRecord struct {
	QueryID      string
	NodeID       string
	Epsilon      float64
	Delta        float64
	Timestamp    time.Time
	QueryType    string
}

// NewPrivacyBudgetManager 创建隐私预算管理器
func NewPrivacyBudgetManager(totalEpsilon, totalDelta float64, window time.Duration) *PrivacyBudgetManager {
	return &PrivacyBudgetManager{
		TotalEpsilon:  totalEpsilon,
		TotalDelta:    totalDelta,
		Queries:       make([]QueryRecord, 0),
		RollingWindow: window,
	}
}

// RecordQuery 记录查询
func (pbm *PrivacyBudgetManager) RecordQuery(record QueryRecord) error {
	pbm.mutex.Lock()
	defer pbm.mutex.Unlock()

	// 检查窗口内的预算使用情况
	windowUsage := pbm.calculateWindowUsage()

	if windowUsage.Epsilon+record.Epsilon > pbm.TotalEpsilon {
		return fmt.Errorf("窗口内隐私预算超限")
	}

	pbm.Queries = append(pbm.Queries, record)

	// 清理过期记录
	pbm.cleanupOldRecords()

	return nil
}

// calculateWindowUsage 计算窗口内的使用量
func (pbm *PrivacyBudgetManager) calculateWindowUsage() struct{Epsilon, Delta float64} {
	now := time.Now()
	var epsilon, delta float64

	for _, query := range pbm.Queries {
		if now.Sub(query.Timestamp) <= pbm.RollingWindow {
			epsilon += query.Epsilon
			delta += query.Delta
		}
	}

	return struct {
		Epsilon float64
		Delta   float64
	}{epsilon, delta}
}

// cleanupOldRecords 清理过期记录
func (pbm *PrivacyBudgetManager) cleanupOldRecords() {
	now := time.Now()
	cutoff := now.Add(-pbm.RollingWindow)

	validQueries := make([]QueryRecord, 0)
	for _, query := range pbm.Queries {
		if query.Timestamp.After(cutoff) {
			validQueries = append(validQueries, query)
		}
	}

	pbm.Queries = validQueries
}

// GetBudgetStatus 获取预算状态
func (pbm *PrivacyBudgetManager) GetBudgetStatus() map[string]interface{} {
	pbm.mutex.RLock()
	defer pbm.mutex.RUnlock()

	windowUsage := pbm.calculateWindowUsage()

	return map[string]interface{}{
		"total_epsilon":           pbm.TotalEpsilon,
		"total_delta":             pbm.TotalDelta,
		"window_epsilon_used":     windowUsage.Epsilon,
		"window_delta_used":       windowUsage.Delta,
		"window_epsilon_remaining": pbm.TotalEpsilon - windowUsage.Epsilon,
		"window_delta_remaining":   pbm.TotalDelta - windowUsage.Delta,
		"total_queries":           len(pbm.Queries),
		"rolling_window":          pbm.RollingWindow.String(),
	}
}