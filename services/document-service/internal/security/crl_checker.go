package security

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// CRLChecker CRL检查器
type CRLChecker struct {
	client    *http.Client
	cache     map[string]*CRLCacheEntry
	cacheTTL  time.Duration
	mutex     sync.RWMutex
	logger    *logrus.Logger
}

// CRLCacheEntry CRL缓存条目
type CRLCacheEntry struct {
	crl        *x509.RevocationList
	expireTime time.Time
	lastUsed   time.Time
	hitCount   int64
}

// NewCRLChecker 创建CRL检查器
func NewCRLChecker(timeout time.Duration, logger *logrus.Logger) *CRLChecker {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	return &CRLChecker{
		client:   &http.Client{Timeout: timeout},
		cache:    make(map[string]*CRLCacheEntry),
		cacheTTL: 30 * time.Minute, // 默认缓存30分钟
		logger:   logger,
	}

	// 启动缓存清理协程
	go crlChecker.cleanup()
}

// CheckRevocationWithCRL 使用CRL检查证书吊销状态
func (cc *CRLChecker) CheckRevocationWithCRL(cert *x509.Certificate, issuer *x509.Certificate) error {
	// 检查是否有CRL分发点
	if len(cert.CRLDistributionPoints) == 0 {
		return fmt.Errorf("证书未配置CRL分发点")
	}

	// 获取主CRL分发点URL
	crlURL := cert.CRLDistributionPoints[0]
	cc.logger.WithFields(logrus.Fields{
		"serial_number": cert.SerialNumber.String(),
		"subject":       cert.Subject.String(),
		"crl_url":       crlURL,
	}).Debug("开始CRL检查")

	// 检查缓存
	cc.mutex.RLock()
	if entry, exists := cc.cache[crlURL]; exists {
		if time.Now().Before(entry.expireTime) {
			cc.mutex.RUnlock()
			entry.hitCount++
			entry.lastUsed = time.Now()
			cc.logger.WithField("crl_url", crlURL).Debug("使用缓存的CRL")
			return cc.checkCRLStatus(cert, entry.crl)
		}
		cc.mutex.RUnlock()
	}

	// 下载CRL
	crl, err := cc.downloadCRL(crlURL)
	if err != nil {
		return fmt.Errorf("下载CRL失败: %w", err)
	}

	// 验证CRL签名
	if err := issuer.CheckCRLSignature(crl); err != nil {
		return fmt.Errorf("CRL签名验证失败: %w", err)
	}

	// 检查CRL是否过期
	if time.Now().After(crl.NextUpdate) {
		return fmt.Errorf("CRL已过期，过期时间: %s", crl.NextUpdate.Format(time.RFC3339))
	}

	// 缓存CRL
	cc.cacheCRL(crlURL, crl)

	// 检查证书吊销状态
	return cc.checkCRLStatus(cert, crl)
}

// downloadCRL 下载CRL
func (cc *CRLChecker) downloadCRL(url string) (*x509.RevocationList, error) {
	cc.logger.WithField("url", url).Debug("开始下载CRL")

	startTime := time.Now()
	resp, err := cc.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP错误状态码: %d", resp.StatusCode)
	}

	crlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应数据失败: %w", err)
	}

	cc.logger.WithFields(logrus.Fields{
		"url":           url,
		"size_bytes":    len(crlData),
		"download_time": duration,
	}).Debug("CRL下载完成")

	// 尝试解析PEM编码的CRL
	block, _ := pem.Decode(crlData)
	if block != nil && block.Type == "X509 CRL" {
		crlData = block.Bytes
		cc.logger.Debug("检测到PEM编码的CRL")
	}

	// 解析CRL
	crl, err := x509.ParseRevocationList(crlData)
	if err != nil {
		return nil, fmt.Errorf("解析CRL失败: %w", err)
	}

	cc.logger.WithFields(logrus.Fields{
		"url":              url,
		"issuer":           crl.Issuer.String(),
		"this_update":      crl.ThisUpdate,
		"next_update":      crl.NextUpdate,
		"revoked_certs":    len(crl.RevokedCertificates),
	}).Info("CRL解析成功")

	return crl, nil
}

// checkCRLStatus 检查证书在CRL中的吊销状态
func (cc *CRLChecker) checkCRLStatus(cert *x509.Certificate, crl *x509.RevocationList) error {
	// 检查CRL是否过期
	if time.Now().Before(crl.ThisUpdate) {
		return fmt.Errorf("CRL尚未生效，生效时间: %s", crl.ThisUpdate.Format(time.RFC3339))
	}

	// 查找证书是否在吊销列表中
	for _, revokedCert := range crl.RevokedCertificates {
		if cert.SerialNumber.Cmp(revokedCert.SerialNumber) == 0 {
			cc.logger.WithFields(logrus.Fields{
				"serial_number":     cert.SerialNumber.String(),
				"subject":           cert.Subject.String(),
				"revocation_time":   revokedCert.RevocationTime,
				"revocation_reason": revokedCert.RevocationReason,
			}).Warn("证书已被吊销")

			return fmt.Errorf("证书已被吊销，吊销时间: %s，吊销原因: %d",
				revokedCert.RevocationTime.Format(time.RFC3339),
				revokedCert.RevocationReason)
		}
	}

	cc.logger.WithFields(logrus.Fields{
		"serial_number": cert.SerialNumber.String(),
		"subject":       cert.Subject.String(),
		"crl_issuer":    crl.Issuer.String(),
	}).Debug("证书未在CRL中，状态正常")

	return nil
}

// cacheCRL 缓存CRL
func (cc *CRLChecker) cacheCRL(url string, crl *x509.RevocationList) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	// 计算缓存过期时间（使用CRL更新时间的一半，或默认TTL）
	var expireTime time.Time
	if crl.NextUpdate.After(time.Now()) {
		cacheDuration := crl.NextUpdate.Sub(time.Now()) / 2
		if cacheDuration > cc.cacheTTL {
			cacheDuration = cc.cacheTTL
		}
		expireTime = time.Now().Add(cacheDuration)
	} else {
		expireTime = time.Now().Add(cc.cacheTTL)
	}

	cc.cache[url] = &CRLCacheEntry{
		crl:        crl,
		expireTime: expireTime,
		lastUsed:   time.Now(),
		hitCount:   0,
	}

	cc.logger.WithFields(logrus.Fields{
		"url":         url,
		"expire_time": expireTime,
		"cache_ttl":   expireTime.Sub(time.Now()),
	}).Debug("CRL已缓存")
}

// cleanup 清理过期缓存
func (cc *CRLChecker) cleanup() {
	ticker := time.NewTicker(cc.cacheTTL / 4) // 每个TTL的1/4时间清理一次
	defer ticker.Stop()

	for range ticker.C {
		cc.mutex.Lock()
		now := time.Now()
		for url, entry := range cc.cache {
			if now.After(entry.expireTime) {
				delete(cc.cache, url)
				cc.logger.WithField("url", url).Debug("清理过期CRL缓存")
			}
		}
		cc.mutex.Unlock()
	}
}

// GetCacheStats 获取缓存统计信息
func (cc *CRLChecker) GetCacheStats() map[string]interface{} {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()

	stats := map[string]interface{}{
		"cache_size": len(cc.cache),
		"entries":    make([]map[string]interface{}, 0, len(cc.cache)),
	}

	for url, entry := range cc.cache {
		entryStats := map[string]interface{}{
			"url":       url,
			"hit_count": entry.hitCount,
			"last_used": entry.lastUsed,
			"expires_at": entry.expireTime,
		}
		stats["entries"] = append(stats["entries"].([]map[string]interface{}), entryStats)
	}

	return stats
}

// ClearCache 清空缓存
func (cc *CRLChecker) ClearCache() {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	cc.cache = make(map[string]*CRLCacheEntry)
	cc.logger.Info("CRL缓存已清空")
}

// SetCacheTTL 设置缓存TTL
func (cc *CRLChecker) SetCacheTTL(ttl time.Duration) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	cc.cacheTTL = ttl
	cc.logger.WithField("ttl", ttl).Info("CRL缓存TTL已更新")
}

// GetCachedCRL 获取缓存的CRL（如果存在且未过期）
func (cc *CRLChecker) GetCachedCRL(url string) (*x509.RevocationList, bool) {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()

	if entry, exists := cc.cache[url]; exists {
		if time.Now().Before(entry.expireTime) {
			entry.hitCount++
			entry.lastUsed = time.Now()
			return entry.crl, true
		}
	}

	return nil, false
}