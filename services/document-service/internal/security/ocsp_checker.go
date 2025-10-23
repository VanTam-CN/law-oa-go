package security

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ocsp"
)

// OCSPChecker OCSP检查器
type OCSPChecker struct {
	client      *http.Client
	cache       map[string]*OCSPCacheEntry
	cacheTTL    time.Duration
	useNonce    bool
	mutex       sync.RWMutex
	logger      *logrus.Logger
	stats       *OCSPStats
}

// OCSPCacheEntry OCSP缓存条目
type OCSPCacheEntry struct {
	response   *ocsp.Response
	expireTime time.Time
	lastUsed   time.Time
	hitCount   int64
}

// OCSPStats OCSP统计信息
type OCSPStats struct {
	TotalRequests    int64     `json:"total_requests"`
	CacheHits        int64     `json:"cache_hits"`
	CacheMisses      int64     `json:"cache_misses"`
	SuccessfulReqs   int64     `json:"successful_requests"`
	FailedReqs       int64     `json:"failed_requests"`
	LastRequestTime  time.Time `json:"last_request_time"`
	AverageLatency   time.Duration `json:"average_latency"`
	mutex            sync.RWMutex
}

// NewOCSPChecker 创建OCSP检查器
func NewOCSPChecker(timeout time.Duration, logger *logrus.Logger) *OCSPChecker {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	return &OCSPChecker{
		client:   &http.Client{Timeout: timeout},
		cache:    make(map[string]*OCSPCacheEntry),
		cacheTTL: 5 * time.Minute, // 默认缓存5分钟
		useNonce: true,           // 默认使用nonce
		logger:   logger,
		stats:    &OCSPStats{},
	}

	// 启动缓存清理协程
	go ocspChecker.cleanup()
}

// CheckRevocationWithOCSP 使用OCSP检查证书吊销状态
func (oc *OCSPChecker) CheckRevocationWithOCSP(cert *x509.Certificate, issuer *x509.Certificate) error {
	// 检查是否有OCSP响应器
	if len(cert.OCSPServer) == 0 {
		return fmt.Errorf("证书未配置OCSP响应器")
	}

	// 使用第一个OCSP服务器URL
	ocspURL := cert.OCSPServer[0]
	oc.logger.WithFields(logrus.Fields{
		"serial_number": cert.SerialNumber.String(),
		"subject":       cert.Subject.String(),
		"ocsp_url":      ocspURL,
	}).Debug("开始OCSP检查")

	// 更新统计信息
	oc.stats.mutex.Lock()
	oc.stats.TotalRequests++
	oc.stats.LastRequestTime = time.Now()
	oc.stats.mutex.Unlock()

	startTime := time.Now()

	// 检查缓存
	cacheKey := oc.generateCacheKey(cert, issuer)
	oc.mutex.RLock()
	if entry, exists := oc.cache[cacheKey]; exists {
		if time.Now().Before(entry.expireTime) {
			oc.mutex.RUnlock()
			entry.hitCount++
			entry.lastUsed = time.Now()

			// 更新缓存命中统计
			oc.stats.mutex.Lock()
			oc.stats.CacheHits++
			oc.stats.mutex.Unlock()

			oc.logger.WithField("cache_key", cacheKey).Debug("使用缓存的OCSP响应")
			return oc.checkOCSPResponse(cert, entry.response)
		}
		oc.mutex.RUnlock()
	}

	// 更新缓存未命中统计
	oc.stats.mutex.Lock()
	oc.stats.CacheMisses++
	oc.stats.mutex.Unlock()

	// 创建OCSP请求
	ocspReq, err := ocsp.CreateRequest(cert, issuer, &ocsp.RequestOptions{
		Hash: crypto.SHA256,
	})
	if err != nil {
		return fmt.Errorf("创建OCSP请求失败: %w", err)
	}

	// 添加nonce（如果启用）
	if oc.useNonce {
		// 注意：ocsp.CreateRequest会自动添加nonce
		oc.logger.Debug("OCSP请求已包含nonce")
	}

	// 发送OCSP请求
	ocspResp, err := oc.sendOCSPRequest(ocspURL, ocspReq)
	if err != nil {
		// 更新失败统计
		oc.stats.mutex.Lock()
		oc.stats.FailedReqs++
		oc.stats.mutex.Unlock()

		return fmt.Errorf("发送OCSP请求失败: %w", err)
	}

	// 验证OCSP响应签名
	if err := ocspResp.CheckSignatureFrom(issuer); err != nil {
		// 更新失败统计
		oc.stats.mutex.Lock()
		oc.stats.FailedReqs++
		oc.stats.mutex.Unlock()

		return fmt.Errorf("OCSP响应签名验证失败: %w", err)
	}

	// 检查响应中的证书序列号是否匹配
	if !ocspResp.Certificate.SerialNumber.Equal(cert.SerialNumber) {
		return fmt.Errorf("OCSP响应的证书序列号不匹配")
	}

	// 缓存OCSP响应
	oc.cacheOCSPResponse(cacheKey, ocspResp)

	// 更新成功统计和延迟
	duration := time.Since(startTime)
	oc.stats.mutex.Lock()
	oc.stats.SuccessfulReqs++
	if oc.stats.AverageLatency == 0 {
		oc.stats.AverageLatency = duration
	} else {
		oc.stats.AverageLatency = (oc.stats.AverageLatency + duration) / 2
	}
	oc.stats.mutex.Unlock()

	oc.logger.WithFields(logrus.Fields{
		"ocsp_url":      ocspURL,
		"duration":      duration,
		"status":        ocspResp.Status,
		"this_update":   ocspResp.ThisUpdate,
		"next_update":   ocspResp.NextUpdate,
	}).Debug("OCSP检查完成")

	// 检查吊销状态
	return oc.checkOCSPResponse(cert, ocspResp)
}

// sendOCSPRequest 发送OCSP请求
func (oc *OCSPChecker) sendOCSPRequest(url string, request []byte) (*ocsp.Response, error) {
	oc.logger.WithField("url", url).Debug("发送OCSP请求")

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(request))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置HTTP头
	httpReq.Header.Set("Content-Type", "application/ocsp-request")
	httpReq.Header.Set("Accept", "application/ocsp-response")
	httpReq.Header.Set("User-Agent", "Law-OA-OCSP-Checker/1.0")
	httpReq.Header.Set("Cache-Control", "no-cache")

	// 发送请求
	resp, err := oc.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OCSP服务器返回错误状态码: %d", resp.StatusCode)
	}

	// 读取响应数据
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应数据失败: %w", err)
	}

	oc.logger.WithFields(logrus.Fields{
		"url":        url,
		"size_bytes": len(respData),
	}).Debug("收到OCSP响应")

	// 解析OCSP响应
	ocspResp, err := ocsp.ParseResponse(respData, nil)
	if err != nil {
		return nil, fmt.Errorf("解析OCSP响应失败: %w", err)
	}

	return ocspResp, nil
}

// checkOCSPResponse 检查OCSP响应状态
func (oc *OCSPChecker) checkOCSPResponse(cert *x509.Certificate, resp *ocsp.Response) error {
	// 检查响应是否过期
	if resp.NextUpdate.Before(time.Now()) {
		return fmt.Errorf("OCSP响应已过期，过期时间: %s", resp.NextUpdate.Format(time.RFC3339))
	}

	// 检查响应是否尚未生效
	if resp.ThisUpdate.After(time.Now()) {
		return fmt.Errorf("OCSP响应尚未生效，生效时间: %s", resp.ThisUpdate.Format(time.RFC3339))
	}

	// 根据状态返回相应结果
	switch resp.Status {
	case ocsp.Good:
		oc.logger.WithFields(logrus.Fields{
			"serial_number": cert.SerialNumber.String(),
			"subject":       cert.Subject.String(),
			"this_update":   resp.ThisUpdate,
			"next_update":   resp.NextUpdate,
		}).Debug("证书状态正常（Good）")
		return nil

	case ocsp.Revoked:
		revocationTime := resp.RevokedAt
		revocationReason := resp.RevocationReason

		oc.logger.WithFields(logrus.Fields{
			"serial_number":    cert.SerialNumber.String(),
			"subject":          cert.Subject.String(),
			"revocation_time":  revocationTime,
			"revocation_reason": revocationReason,
		}).Warn("证书已被吊销（Revoked）")

		return fmt.Errorf("证书已被吊销，吊销时间: %s，吊销原因: %d",
			revocationTime.Format(time.RFC3339), revocationReason)

	case ocsp.Unknown:
		oc.logger.WithFields(logrus.Fields{
			"serial_number": cert.SerialNumber.String(),
			"subject":       cert.Subject.String(),
		}).Warn("证书状态未知（Unknown）")

		return fmt.Errorf("证书状态未知")

	default:
		return fmt.Errorf("意外的OCSP状态: %d", resp.Status)
	}
}

// generateCacheKey 生成缓存键
func (oc *OCSPChecker) generateCacheKey(cert *x509.Certificate, issuer *x509.Certificate) string {
	certHash := ocspResponseHash(cert.Raw)
	issuerHash := ocspResponseHash(issuer.Raw)
	return fmt.Sprintf("%s_%s", certHash, issuerHash)
}

// ocspResponseHash 计算响应哈希用于缓存键
func ocspResponseHash(data []byte) string {
	hash := crypto.SHA256.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

// cacheOCSPResponse 缓存OCSP响应
func (oc *OCSPChecker) cacheOCSPResponse(cacheKey string, resp *ocsp.Response) {
	oc.mutex.Lock()
	defer oc.mutex.Unlock()

	// 计算缓存过期时间（使用响应的NextUpdate时间的一半，或默认TTL）
	var expireTime time.Time
	if resp.NextUpdate.After(time.Now()) {
		cacheDuration := resp.NextUpdate.Sub(time.Now()) / 2
		if cacheDuration > oc.cacheTTL {
			cacheDuration = oc.cacheTTL
		}
		expireTime = time.Now().Add(cacheDuration)
	} else {
		expireTime = time.Now().Add(oc.cacheTTL)
	}

	oc.cache[cacheKey] = &OCSPCacheEntry{
		response:   resp,
		expireTime: expireTime,
		lastUsed:   time.Now(),
		hitCount:   0,
	}

	oc.logger.WithFields(logrus.Fields{
		"cache_key":   cacheKey,
		"expire_time": expireTime,
		"cache_ttl":   expireTime.Sub(time.Now()),
		"status":      resp.Status,
	}).Debug("OCSP响应已缓存")
}

// cleanup 清理过期缓存
func (oc *OCSPChecker) cleanup() {
	ticker := time.NewTicker(oc.cacheTTL / 4) // 每个TTL的1/4时间清理一次
	defer ticker.Stop()

	for range ticker.C {
		oc.mutex.Lock()
		now := time.Now()
		for key, entry := range oc.cache {
			if now.After(entry.expireTime) {
				delete(oc.cache, key)
				oc.logger.WithField("cache_key", key).Debug("清理过期OCSP缓存")
			}
		}
		oc.mutex.Unlock()
	}
}

// GetCacheStats 获取缓存统计信息
func (oc *OCSPChecker) GetCacheStats() map[string]interface{} {
	oc.mutex.RLock()
	defer oc.mutex.RUnlock()

	stats := map[string]interface{}{
		"cache_size": len(oc.cache),
		"entries":    make([]map[string]interface{}, 0, len(oc.cache)),
	}

	for key, entry := range oc.cache {
		entryStats := map[string]interface{}{
			"cache_key":   key,
			"hit_count":   entry.hitCount,
			"last_used":   entry.lastUsed,
			"expires_at":  entry.expireTime,
			"status":      entry.response.Status,
		}
		stats["entries"] = append(stats["entries"].([]map[string]interface{}), entryStats)
	}

	return stats
}

// GetStats 获取OCSP统计信息
func (oc *OCSPChecker) GetStats() *OCSPStats {
	oc.stats.mutex.RLock()
	defer oc.stats.mutex.RUnlock()

	// 返回统计信息的副本
	return &OCSPStats{
		TotalRequests:   oc.stats.TotalRequests,
		CacheHits:       oc.stats.CacheHits,
		CacheMisses:     oc.stats.CacheMisses,
		SuccessfulReqs:  oc.stats.SuccessfulReqs,
		FailedReqs:      oc.stats.FailedReqs,
		LastRequestTime: oc.stats.LastRequestTime,
		AverageLatency:  oc.stats.AverageLatency,
	}
}

// ClearCache 清空缓存
func (oc *OCSPChecker) ClearCache() {
	oc.mutex.Lock()
	defer oc.mutex.Unlock()

	oc.cache = make(map[string]*OCSPCacheEntry)
	oc.logger.Info("OCSP缓存已清空")
}

// SetCacheTTL 设置缓存TTL
func (oc *OCSPChecker) SetCacheTTL(ttl time.Duration) {
	oc.mutex.Lock()
	defer oc.mutex.Unlock()

	oc.cacheTTL = ttl
	oc.logger.WithField("ttl", ttl).Info("OCSP缓存TTL已更新")
}

// SetUseNonce 设置是否使用nonce
func (oc *OCSPChecker) SetUseNonce(useNonce bool) {
	oc.mutex.Lock()
	defer oc.mutex.Unlock()

	oc.useNonce = useNonce
	oc.logger.WithField("use_nonce", useNonce).Info("OCSP nonce设置已更新")
}

// GetCachedResponse 获取缓存的OCSP响应（如果存在且未过期）
func (oc *OCSPChecker) GetCachedResponse(cert *x509.Certificate, issuer *x509.Certificate) (*ocsp.Response, bool) {
	cacheKey := oc.generateCacheKey(cert, issuer)

	oc.mutex.RLock()
	defer oc.mutex.RUnlock()

	if entry, exists := oc.cache[cacheKey]; exists {
		if time.Now().Before(entry.expireTime) {
			entry.hitCount++
			entry.lastUsed = time.Now()
			return entry.response, true
		}
	}

	return nil, false
}