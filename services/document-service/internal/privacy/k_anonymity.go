package privacy

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// KAnonymityProcessor k-匿名化处理器
type KAnonymityProcessor struct {
	K                 int
	QuasiIdentifiers  []string
	SensitiveAttr     string
	SuppressionRate   float64
	Generalization    *GeneralizationHierarchy
	Clustering        *DataClusterer
	logger            *slog.Logger
	mutex             sync.RWMutex
}

// GeneralizationHierarchy 泛化层次结构
type GeneralizationHierarchy struct {
	Hierarchies map[string]*Hierarchy
	mutex       sync.RWMutex
}

// Hierarchy 层次结构
type Hierarchy struct {
	Levels    []map[string]string
	Domain    []string
	MaxLevel  int
}

// DataClusterer 数据聚类器
type DataClusterer struct {
	Algorithm string // "kmeans", "hierarchical", "density"
	Params    map[string]interface{}
	mutex     sync.RWMutex
}

// DataRecord 数据记录
type DataRecord struct {
	ID        string
	Values    map[string]interface{}
	Sensitive string
	ClusterID int
}

// Cluster 聚类
type Cluster struct {
	ID       int
	Records  []*DataRecord
	Center   map[string]interface{}
	Size     int
}

// EquivalenceClass 等价类
type EquivalenceClass struct {
	QuasiValues map[string]string
	Records      []*DataRecord
	Size         int
	SensitiveValues map[string]int
}

// AnonymityResult 匿名化结果
type AnonymityResult struct {
	AnonymizedData    []map[string]interface{}
	Statistics        *AnonymityStatistics
	ProcessingTime    time.Duration
	Quality           float64
	SuppressedCount   int
	AnonymityLevel    string
}

// AnonymityStatistics 匿名化统计信息
type AnonymityStatistics struct {
	OriginalSize      int
	AnonymizedSize    int
	KValue            int
	LValue            int
	TCloseness        float64
	Precision         float64
	Discernibility    float64
	InformationLoss   float64
	ReidentificationRisk float64
}

// NewKAnonymityProcessor 创建k-匿名化处理器
func NewKAnonymityProcessor(k int, quasiIdentifiers []string, sensitiveAttr string, logger *slog.Logger) *KAnonymityProcessor {
	return &KAnonymityProcessor{
		K:                k,
		QuasiIdentifiers: quasiIdentifiers,
		SensitiveAttr:    sensitiveAttr,
		SuppressionRate:  0.05, // 默认5%抑制率
		Generalization:   NewGeneralizationHierarchy(),
		Clustering:       NewDataClusterer("kmeans", map[string]interface{}{"k": k}),
		logger:           logger,
	}
}

// ProcessDataset 处理数据集
func (kap *KAnonymityProcessor) ProcessDataset(ctx context.Context, data []map[string]interface{}) (*AnonymityResult, error) {
	startTime := time.Now()

	kap.mutex.Lock()
	defer kap.mutex.Unlock()

	kap.logger.Info("开始k-匿名化处理",
		"dataset_size", len(data),
		"k_value", kap.K,
		"quasi_identifiers", kap.QuasiIdentifiers,
	)

	// 1. 数据预处理
	records := kap.preprocessData(data)

	// 2. 聚类分析
	clusters, err := kap.performClustering(records)
	if err != nil {
		return nil, fmt.Errorf("聚类分析失败: %w", err)
	}

	// 3. 构建等价类
	equivalenceClasses := kap.buildEquivalenceClasses(clusters)

	// 4. 检查k-匿名性
	validClasses := kap.validateKAnonymity(equivalenceClasses)

	// 5. 泛化和抑制
	anonymizedData := kap.generalizeAndSuppress(validClasses)

	// 6. 计算统计信息
	statistics := kap.calculateStatistics(records, anonymizedData)

	processingTime := time.Since(startTime)

	result := &AnonymityResult{
		AnonymizedData:  anonymizedData,
		Statistics:      statistics,
		ProcessingTime:  processingTime,
		Quality:         kap.calculateQuality(statistics),
		SuppressedCount: len(records) - len(anonymizedData),
		AnonymityLevel:  fmt.Sprintf("k=%d", kap.K),
	}

	kap.logger.Info("k-匿名化处理完成",
		"original_size", len(data),
		"anonymized_size", len(anonymizedData),
		"suppressed_count", result.SuppressedCount,
		"processing_time", processingTime,
		"quality", result.Quality,
	)

	return result, nil
}

// preprocessData 数据预处理
func (kap *KAnonymityProcessor) preprocessData(data []map[string]interface{}) []*DataRecord {
	records := make([]*DataRecord, 0, len(data))

	for i, record := range data {
		dataRecord := &DataRecord{
			ID:        fmt.Sprintf("record_%d", i),
			Values:    make(map[string]interface{}),
			Sensitive: fmt.Sprintf("%v", record[kap.SensitiveAttr]),
		}

		// 提取准标识符值
		for _, quasi := range kap.QuasiIdentifiers {
			if value, exists := record[quasi]; exists {
				dataRecord.Values[quasi] = value
			}
		}

		records = append(records, dataRecord)
	}

	return records
}

// performClustering 执行聚类
func (kap *KAnonymityProcessor) performClustering(records []*DataRecord) ([]*Cluster, error) {
	switch kap.Clustering.Algorithm {
	case "kmeans":
		return kap.kmeansClustering(records)
	case "hierarchical":
		return kap.hierarchicalClustering(records)
	case "density":
		return kap.densityBasedClustering(records)
	default:
		return kap.kmeansClustering(records)
	}
}

// kmeansClustering K-均值聚类
func (kap *KAnonymityProcessor) kmeansClustering(records []*DataRecord) ([]*Cluster, error) {
	if len(records) < kap.K {
		return nil, fmt.Errorf("记录数量小于k值")
	}

	// 初始化聚类中心
	clusters := kap.initializeClusters(records)

	// 迭代优化
	maxIterations := 100
	for iteration := 0; iteration < maxIterations; iteration++ {
		// 分配记录到最近的聚类中心
		changed := kap.assignRecordsToClusters(records, clusters)
		if !changed {
			break
		}

		// 更新聚类中心
		kap.updateClusterCenters(clusters)
	}

	// 确保每个聚类至少包含k个记录
	clusters = kap.mergeSmallClusters(clusters, records)

	return clusters, nil
}

// initializeClusters 初始化聚类
func (kap *KAnonymityProcessor) initializeClusters(records []*DataRecord) []*Cluster {
	clusters := make([]*Cluster, kap.K)

	// 随机选择k个记录作为初始中心
	selected := make(map[int]bool)
	for i := 0; i < kap.K && i < len(records); i++ {
		centerIdx := i
		for selected[centerIdx] {
			centerIdx = (centerIdx + 1) % len(records)
		}
		selected[centerIdx] = true

		clusters[i] = &Cluster{
			ID:      i,
			Records: make([]*DataRecord, 0),
			Center:  make(map[string]interface{}),
			Size:    0,
		}

		// 复制准标识符值作为聚类中心
		for _, quasi := range kap.QuasiIdentifiers {
			if value, exists := records[centerIdx].Values[quasi]; exists {
				clusters[i].Center[quasi] = value
			}
		}
	}

	return clusters
}

// assignRecordsToClusters 分配记录到聚类
func (kap *KAnonymityProcessor) assignRecordsToClusters(records []*DataRecord, clusters []*Cluster) bool {
	changed := false

	// 清空现有聚类
	for _, cluster := range clusters {
		cluster.Records = make([]*DataRecord, 0)
		cluster.Size = 0
	}

	// 分配每个记录到最近的聚类
	for _, record := range records {
		closestCluster := kap.findClosestCluster(record, clusters)
		closestCluster.Records = append(closestCluster.Records, record)
		closestCluster.Size++

		if record.ClusterID != closestCluster.ID {
			record.ClusterID = closestCluster.ID
			changed = true
		}
	}

	return changed
}

// findClosestCluster 找到最近的聚类
func (kap *KAnonymityProcessor) findClosestCluster(record *DataRecord, clusters []*Cluster) *Cluster {
	minDistance := math.MaxFloat64
	closestCluster := clusters[0]

	for _, cluster := range clusters {
		distance := kap.calculateDistance(record.Values, cluster.Center)
		if distance < minDistance {
			minDistance = distance
			closestCluster = cluster
		}
	}

	return closestCluster
}

// calculateDistance 计算距离
func (kap *KAnonymityProcessor) calculateDistance(values1, values2 map[string]interface{}) float64 {
	var distance float64

	for _, quasi := range kap.QuasiIdentifiers {
		val1, exists1 := values1[quasi]
		val2, exists2 := values2[quasi]

		if !exists1 || !exists2 {
			continue
		}

		// 简单的距离计算（可以扩展为更复杂的度量）
		if val1 != val2 {
			distance += 1.0
		}
	}

	return distance
}

// updateClusterCenters 更新聚类中心
func (kap *KAnonymityProcessor) updateClusterCenters(clusters []*Cluster) {
	for _, cluster := range clusters {
		if cluster.Size == 0 {
			continue
		}

		// 计算每个准标识符的众数作为新的聚类中心
		for _, quasi := range kap.QuasiIdentifiers {
			valueCounts := make(map[string]int)

			for _, record := range cluster.Records {
				if value, exists := record.Values[quasi]; exists {
					valueStr := fmt.Sprintf("%v", value)
					valueCounts[valueStr]++
				}
			}

			// 找到众数
			maxCount := 0
			modeValue := ""
			for value, count := range valueCounts {
				if count > maxCount {
					maxCount = count
					modeValue = value
				}
			}

			cluster.Center[quasi] = modeValue
		}
	}
}

// mergeSmallClusters 合并小聚类
func (kap *KAnonymityProcessor) mergeSmallClusters(clusters []*Cluster, records []*DataRecord) []*Cluster {
	// 找出小于k的聚类
	smallClusters := make([]*Cluster, 0)
	largeClusters := make([]*Cluster, 0)

	for _, cluster := range clusters {
		if cluster.Size < kap.K {
			smallClusters = append(smallClusters, cluster)
		} else {
			largeClusters = append(largeClusters, cluster)
		}
	}

	// 将小聚类的记录分配到最近的大聚类
	for _, smallCluster := range smallClusters {
		if len(largeClusters) == 0 {
			// 如果没有大聚类，将小聚类合并到最近的其他聚类
			closestCluster := kap.findClosestClusterForCluster(smallCluster, clusters)
			closestCluster.Records = append(closestCluster.Records, smallCluster.Records...)
			closestCluster.Size += smallCluster.Size
		} else {
			for _, record := range smallCluster.Records {
				closestLarge := kap.findClosestCluster(record, largeClusters)
				closestLarge.Records = append(closestLarge.Records, record)
				closestLarge.Size++
				record.ClusterID = closestLarge.ID
			}
		}
	}

	// 返回有效的大聚类
	validClusters := make([]*Cluster, 0)
	for _, cluster := range clusters {
		if cluster.Size >= kap.K {
			validClusters = append(validClusters, cluster)
		}
	}

	return validClusters
}

// findClosestClusterForCluster 为聚类找到最近的聚类
func (kap *KAnonymityProcessor) findClosestClusterForCluster(cluster *Cluster, clusters []*Cluster) *Cluster {
	if len(clusters) <= 1 {
		return cluster
	}

	minDistance := math.MaxFloat64
	closestCluster := clusters[0]

	if clusters[0] == cluster {
		closestCluster = clusters[1]
	}

	for _, otherCluster := range clusters {
		if otherCluster == cluster {
			continue
		}

		distance := kap.calculateClusterDistance(cluster, otherCluster)
		if distance < minDistance {
			minDistance = distance
			closestCluster = otherCluster
		}
	}

	return closestCluster
}

// calculateClusterDistance 计算聚类间距离
func (kap *KAnonymityProcessor) calculateClusterDistance(cluster1, cluster2 *Cluster) float64 {
	return kap.calculateDistance(cluster1.Center, cluster2.Center)
}

// hierarchicalClustering 层次聚类
func (kap *KAnonymityProcessor) hierarchicalClustering(records []*DataRecord) ([]*Cluster, error) {
	// 简化的层次聚类实现
	clusters := make([]*Cluster, 0, len(records))

	// 初始时每个记录是一个聚类
	for i, record := range records {
		cluster := &Cluster{
			ID:      i,
			Records: []*DataRecord{record},
			Center:  record.Values,
			Size:    1,
		}
		clusters = append(clusters, cluster)
		record.ClusterID = i
	}

	// 迭代合并最近的聚类，直到聚类数适合k值
	for len(clusters) > len(records)/kap.K {
		closestPair := kap.findClosestClusterPair(clusters)
		if closestPair == nil {
			break
		}

		// 合并聚类
		kap.mergeClusters(closestPair.cluster1, closestPair.cluster2)
	}

	return clusters, nil
}

// findClosestClusterPair 找到最近的聚类对
func (kap *KAnonymityProcessor) findClosestClusterPair(clusters []*Cluster) *struct {
	cluster1 *Cluster
	cluster2 *Cluster
} {
	var closest struct {
		cluster1 *Cluster
		cluster2 *Cluster
	}
	minDistance := math.MaxFloat64

	for i, cluster1 := range clusters {
		for j, cluster2 := range clusters {
			if i >= j {
				continue
			}

			distance := kap.calculateClusterDistance(cluster1, cluster2)
			if distance < minDistance {
				minDistance = distance
				closest.cluster1 = cluster1
				closest.cluster2 = cluster2
			}
		}
	}

	if minDistance == math.MaxFloat64 {
		return nil
	}

	return &closest
}

// mergeClusters 合并聚类
func (kap *KAnonymityProcessor) mergeClusters(cluster1, cluster2 *Cluster) {
	cluster1.Records = append(cluster1.Records, cluster2.Records...)
	cluster1.Size += cluster2.Size
	kap.updateClusterCenters([]*Cluster{cluster1})
}

// densityBasedClustering 基于密度的聚类
func (kap *KAnonymityProcessor) densityBasedClustering(records []*DataRecord) ([]*Cluster, error) {
	// 简化的DBSCAN实现
	eps := 1.0 // 邻域半径
	minPts := kap.K // 最小点数

	clusters := make([]*Cluster, 0)
	visited := make(map[string]bool)

	for _, record := range records {
		if visited[record.ID] {
			continue
		}

		neighbors := kap.getNeighbors(record, records, eps)
		if len(neighbors) < minPts {
			// 噪声点
			visited[record.ID] = true
			continue
		}

		// 创建新聚类
		cluster := &Cluster{
			ID:      len(clusters),
			Records: make([]*DataRecord, 0),
			Center:  make(map[string]interface{}),
			Size:    0,
		}

		// 扩展聚类
		kap.expandCluster(record, neighbors, cluster, records, eps, minPts, visited)
		clusters = append(clusters, cluster)
	}

	return clusters, nil
}

// getNeighbors 获取邻居点
func (kap *KAnonymityProcessor) getNeighbors(record *DataRecord, records []*DataRecord, eps float64) []*DataRecord {
	neighbors := make([]*DataRecord, 0)

	for _, otherRecord := range records {
		if otherRecord.ID == record.ID {
			continue
		}

		distance := kap.calculateDistance(record.Values, otherRecord.Values)
		if distance <= eps {
			neighbors = append(neighbors, otherRecord)
		}
	}

	return neighbors
}

// expandCluster 扩展聚类
func (kap *KAnonymityProcessor) expandCluster(record *DataRecord, neighbors []*DataRecord, cluster *Cluster, allRecords []*DataRecord, eps float64, minPts int, visited map[string]bool) {
	cluster.Records = append(cluster.Records, record)
	cluster.Size++
	visited[record.ID] = true

	i := 0
	for i < len(neighbors) {
		neighbor := neighbors[i]

		if !visited[neighbor.ID] {
			visited[neighbor.ID] = true
			neighborNeighbors := kap.getNeighbors(neighbor, allRecords, eps)

			if len(neighborNeighbors) >= minPts {
				neighbors = append(neighbors, neighborNeighbors...)
			}
		}

		// 将邻居添加到聚类
		alreadyInCluster := false
		for _, clusterRecord := range cluster.Records {
			if clusterRecord.ID == neighbor.ID {
				alreadyInCluster = true
				break
			}
		}

		if !alreadyInCluster {
			cluster.Records = append(cluster.Records, neighbor)
			cluster.Size++
		}

		i++
	}
}

// buildEquivalenceClasses 构建等价类
func (kap *KAnonymityProcessor) buildEquivalenceClasses(clusters []*Cluster) []*EquivalenceClass {
	equivalenceClasses := make([]*EquivalenceClass, 0)

	for _, cluster := range clusters {
		if cluster.Size < kap.K {
			continue
		}

		eqClass := &EquivalenceClass{
			QuasiValues:    make(map[string]string),
			Records:         cluster.Records,
			Size:            cluster.Size,
			SensitiveValues: make(map[string]int),
		}

		// 构建准标识符的泛化值
		for _, quasi := range kap.QuasiIdentifiers {
			generalizedValue := kap.generalizeValue(cluster, quasi)
			eqClass.QuasiValues[quasi] = generalizedValue
		}

		// 统计敏感属性值分布
		for _, record := range cluster.Records {
			eqClass.SensitiveValues[record.Sensitive]++
		}

		equivalenceClasses = append(equivalenceClasses, eqClass)
	}

	return equivalenceClasses
}

// generalizeValue 泛化值
func (kap *KAnonymityProcessor) generalizeValue(cluster *Cluster, attribute string) string {
	values := make(map[string]int)

	for _, record := range cluster.Records {
		if value, exists := record.Values[attribute]; exists {
			valueStr := fmt.Sprintf("%v", value)
			values[valueStr]++
		}
	}

	// 找到最一般的泛化值
	if len(values) == 1 {
		for value := range values {
			return value
		}
	}

	// 尝试找到共同的泛化层次
	return kap.findCommonGeneralization(values, attribute)
}

// findCommonGeneralization 找到共同泛化
func (kap *KAnonymityProcessor) findCommonGeneralization(values map[string]int, attribute string) string {
	// 简化的泛化逻辑
	// 实际实现中应该有更复杂的泛化层次结构

	switch attribute {
	case "age":
		// 年龄泛化到年龄段
		return "[年龄区间]"
	case "zipcode":
		// 邮编泛化到前缀
		return "[邮编前缀]"
	case "salary":
		// 薪资泛化到范围
		return "[薪资范围]"
	case "education":
		// 教育程度泛化
		return "[教育水平]"
	default:
		return "[泛化值]"
	}
}

// validateKAnonymity 验证k-匿名性
func (kap *KAnonymityProcessor) validateKAnonymity(equivalenceClasses []*EquivalenceClass) []*EquivalenceClass {
	validClasses := make([]*EquivalenceClass, 0)

	for _, eqClass := range equivalenceClasses {
		if eqClass.Size >= kap.K {
			validClasses = append(validClasses, eqClass)
		}
	}

	return validClasses
}

// generalizeAndSuppress 泛化和抑制
func (kap *KAnonymityProcessor) generalizeAndSuppress(equivalenceClasses []*EquivalenceClass) []map[string]interface{} {
	anonymizedData := make([]map[string]interface{}, 0)

	for _, eqClass := range equivalenceClasses {
		// 检查是否需要抑制
		if kap.shouldSuppress(eqClass) {
			continue // 抑制这个等价类
		}

		// 为每个记录生成匿名化数据
		for _, record := range eqClass.Records {
			anonymizedRecord := make(map[string]interface{})

			// 复制非准标识符字段
			for key, value := range record.Values {
				if !kap.isQuasiIdentifier(key) {
					anonymizedRecord[key] = value
				}
			}

			// 添加泛化的准标识符
			for quasi, generalizedValue := range eqClass.QuasiValues {
				anonymizedRecord[quasi] = generalizedValue
			}

			// 添加敏感属性
			anonymizedRecord[kap.SensitiveAttr] = record.Sensitive

			anonymizedData = append(anonymizedData, anonymizedRecord)
		}
	}

	return anonymizedData
}

// shouldSuppress 判断是否应该抑制
func (kap *KAnonymityProcessor) shouldSuppress(eqClass *EquivalenceClass) bool {
	// 检查l-多样性
	if !kap.checkLDiversity(eqClass) {
		return true
	}

	// 检查t-接近度
	if !kap.checkTCloseness(eqClass) {
		return true
	}

	// 检查抑制率
	suppressionThreshold := int(float64(len(eqClass.Records)) * kap.SuppressionRate)
	if len(eqClass.Records) < suppressionThreshold {
		return true
	}

	return false
}

// checkLDiversity 检查l-多样性
func (kap *KAnonymityProcessor) checkLDiversity(eqClass *EquivalenceClass) bool {
	l := kap.K // 简化：使用k值作为l值
	return len(eqClass.SensitiveValues) >= l
}

// checkTCloseness 检查t-接近度
func (kap *KAnonymityProcessor) checkTCloseness(eqClass *EquivalenceClass) bool {
	// 简化的t-接近度检查
	t := 0.2 // 默认t值

	// 计算敏感属性分布与全局分布的距离
	// 这里简化为检查敏感属性值的分布是否过于集中
	total := float64(eqClass.Size)
	maxFreq := 0.0

	for _, count := range eqClass.SensitiveValues {
		freq := float64(count) / total
		if freq > maxFreq {
			maxFreq = freq
		}
	}

	return maxFreq <= (1.0 - t)
}

// isQuasiIdentifier 检查是否为准标识符
func (kap *KAnonymityProcessor) isQuasiIdentifier(attribute string) bool {
	for _, quasi := range kap.QuasiIdentifiers {
		if quasi == attribute {
			return true
		}
	}
	return false
}

// calculateStatistics 计算统计信息
func (kap *KAnonymityProcessor) calculateStatistics(originalRecords []*DataRecord, anonymizedData []map[string]interface{}) *AnonymityStatistics {
	statistics := &AnonymityStatistics{
		OriginalSize:   len(originalRecords),
		AnonymizedSize: len(anonymizedData),
		KValue:         kap.K,
		LValue:         kap.K, // 简化设置
		TCloseness:     kap.calculateTCloseness(anonymizedData),
		Precision:      kap.calculatePrecision(originalRecords, anonymizedData),
		Discernibility: kap.calculateDiscernibility(anonymizedData),
	}

	statistics.InformationLoss = 1.0 - statistics.Precision
	statistics.ReidentificationRisk = kap.calculateReidentificationRisk(anonymizedData)

	return statistics
}

// calculateTCloseness 计算t-接近度
func (kap *KAnonymityProcessor) calculateTCloseness(data []map[string]interface{}) float64 {
	// 简化的t-接近度计算
	return 0.8 // 示例值
}

// calculatePrecision 计算精度
func (kap *KAnonymityProcessor) calculatePrecision(original []*DataRecord, anonymized []map[string]interface{}) float64 {
	if len(original) == 0 {
		return 0.0
	}

	// 简化的精度计算
	return float64(len(anonymized)) / float64(len(original))
}

// calculateDiscernibility 计算可区分性
func (kap *KAnonymityProcessor) calculateDiscernibility(data []map[string]interface{}) float64 {
	// 简化的可区分性计算
	return 0.7 // 示例值
}

// calculateReidentificationRisk 计算重识别风险
func (kap *KAnonymityProcessor) calculateReidentificationRisk(data []map[string]interface{}) float64 {
	// 简化的重识别风险计算
	return 1.0 / float64(kap.K)
}

// calculateQuality 计算匿名化质量
func (kap *KAnonymityProcessor) calculateQuality(statistics *AnonymityStatistics) float64 {
	// 综合质量指标
	weights := map[string]float64{
		"precision":    0.3,
		"discernibility": 0.2,
		"risk":         0.3,
		"information_loss": 0.2,
	}

	quality := weights["precision"]*statistics.Precision +
		weights["discernibility"]*statistics.Discernibility +
		weights["risk"]*(1-statistics.ReidentificationRisk) +
		weights["information_loss"]*(1-statistics.InformationLoss)

	return math.Max(0.0, math.Min(1.0, quality))
}

// NewGeneralizationHierarchy 创建泛化层次结构
func NewGeneralizationHierarchy() *GeneralizationHierarchy {
	return &GeneralizationHierarchy{
		Hierarchies: make(map[string]*Hierarchy),
	}
}

// NewDataClusterer 创建数据聚类器
func NewDataClusterer(algorithm string, params map[string]interface{}) *DataClusterer {
	return &DataClusterer{
		Algorithm: algorithm,
		Params:    params,
	}
}

// LDiversityProcessor l-多样性处理器
type LDiversityProcessor struct {
	L           int
	SensitiveAttr string
	logger      *slog.Logger
}

// NewLDiversityProcessor 创建l-多样性处理器
func NewLDiversityProcessor(l int, sensitiveAttr string, logger *slog.Logger) *LDiversityProcessor {
	return &LDiversityProcessor{
		L:            l,
		SensitiveAttr: sensitiveAttr,
		logger:       logger,
	}
}

// EnforceLDiversity 强制l-多样性
func (ldp *LDiversityProcessor) EnforceLDiversity(equivalenceClasses []*EquivalenceClass) []*EquivalenceClass {
	validClasses := make([]*EquivalenceClass, 0)

	for _, eqClass := range equivalenceClasses {
		if len(eqClass.SensitiveValues) >= ldp.L {
			validClasses = append(validClasses, eqClass)
		} else {
			// l-多样性不足，需要进行抑制或泛化
			ldp.logger.Warn("等价类不满足l-多样性",
				"class_size", eqClass.Size,
				"diversity", len(eqClass.SensitiveValues),
				"required_l", ldp.L,
			)
		}
	}

	return validClasses
}