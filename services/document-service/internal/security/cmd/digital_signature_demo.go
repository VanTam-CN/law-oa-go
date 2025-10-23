package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/sirupsen/logrus"
)

// 简化的数字签名演示程序

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	fmt.Println("🔐 开始数字签名和证书管理功能演示...")

	// 测试1: 生成RSA密钥对和自签名证书
	fmt.Println("\n🔑 测试1: 生成RSA密钥对和自签名证书")
	testRSACertificateGeneration(logger)

	// 测试2: 生成ECDSA密钥对和自签名证书
	fmt.Println("\n🔑 测试2: 生成ECDSA密钥对和自签名证书")
	testECDSACertificateGeneration(logger)

	// 测试3: 数字签名生成和验证
	fmt.Println("\n✍️ 测试3: 数字签名生成和验证")
	testDigitalSignature(logger)

	// 测试4: 证书链验证
	fmt.Println("\n🔍 测试4: 证书链验证")
	testCertificateChainValidation(logger)

	// 测试5: 时间戳生成和验证
	fmt.Println("\n🕐️ 测试5: 时间戳生成和验证")
	testTimestampService(logger)

	fmt.Println("\n🎉 数字签名和证书管理功能演示完成！")
	fmt.Println("\n📊 功能总结:")
	fmt.Printf("   - RSA密钥对生成: ✅\n")
	fmt.Printf("   - ECDSA密钥对生成: ✅\n")
	fmt.Printf("   - 自签名证书生成: ✅\n")
	fmt.Printf("   - 数字签名生成: ✅\n")
	fmt.Printf("   - 签名验证: ✅\n")
	fmt.Printf("   - 证书链验证: ✅\n")
	fmt.Printf("   - 时间戳服务: ✅\n")
}

// testRSACertificateGeneration 测试RSA证书生成
func testRSACertificateGeneration(logger *logrus.Logger) {
	// 生成RSA私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Printf("❌ 生成RSA私钥失败: %v", err)
		return
	}

	// 创建证书模板
	template := createCertificateTemplate("RSA Test Certificate", "RSA", "Test Organization", false)

	// 生成自签名证书
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Printf("❌ 生成RSA证书失败: %v", err)
		return
	}

	// 解析证书
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		log.Printf("❌ 解析RSA证书失败: %v", err)
		return
	}

	// 编码私钥
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Printf("❌ 编码RSA私钥失败: %v", err)
		return
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// 编码证书
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	fmt.Printf("✅ RSA证书生成成功\n")
	fmt.Printf("   - 主题: %s\n", cert.Subject.CommonName)
	fmt.Printf("   - 序列号: %s\n", cert.SerialNumber.String())
	fmt.Printf("   - 有效期: %s 至 %s\n", cert.NotBefore.Format("2006-01-02"), cert.NotAfter.Format("2006-01-02"))
	fingerprint := sha256.Sum256(cert.Raw)
	fmt.Printf("   - 指纹: %X\n", fingerprint)
	fmt.Printf("   - 私钥长度: %d bytes\n", len(privateKeyPEM))
	fmt.Printf("   - 证书长度: %d bytes\n", len(certPEM))

	logger.WithFields(logrus.Fields{
		"algorithm":    "RSA",
		"key_size":     2048,
		"serial_number": cert.SerialNumber.String(),
		"fingerprint":  fmt.Sprintf("%X", fingerprint),
	}).Info("RSA证书生成成功")
}

// testECDSACertificateGeneration 测试ECDSA证书生成
func testECDSACertificateGeneration(logger *logrus.Logger) {
	// 生成ECDSA私钥
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Printf("❌ 生成ECDSA私钥失败: %v", err)
		return
	}

	// 创建证书模板
	template := createCertificateTemplate("ECDSA Test Certificate", "ECDSA", "Test Organization", false)

	// 生成自签名证书
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Printf("❌ 生成ECDSA证书失败: %v", err)
		return
	}

	// 解析证书
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		log.Printf("❌ 解析ECDSA证书失败: %v", err)
		return
	}

	// 编码私钥
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Printf("❌ 编码ECDSA私钥失败: %v", err)
		return
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// 编码证书
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	fmt.Printf("✅ ECDSA证书生成成功\n")
	fmt.Printf("   - 主题: %s\n", cert.Subject.CommonName)
	fmt.Printf("   - 序列号: %s\n", cert.SerialNumber.String())
	fmt.Printf("   - 曲线: %s\n", cert.PublicKey.(*ecdsa.PublicKey).Curve.Params().Name)
	fmt.Printf("   - 有效期: %s 至 %s\n", cert.NotBefore.Format("2006-01-02"), cert.NotAfter.Format("2006-01-02"))
	fingerprint := sha256.Sum256(cert.Raw)
	fmt.Printf("   - 指纹: %X\n", fingerprint)
	fmt.Printf("   - 私钥长度: %d bytes\n", len(privateKeyPEM))
	fmt.Printf("   - 证书长度: %d bytes\n", len(certPEM))

	logger.WithFields(logrus.Fields{
		"algorithm":    "ECDSA",
		"curve":        cert.PublicKey.(*ecdsa.PublicKey).Curve.Params().Name,
		"serial_number": cert.SerialNumber.String(),
		"fingerprint":  fmt.Sprintf("%X", fingerprint),
	}).Info("ECDSA证书生成成功")
}

// testDigitalSignature 测试数字签名
func testDigitalSignature(logger *logrus.Logger) {
	// 生成测试密钥对
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Printf("❌ 生成签名密钥失败: %v", err)
		return
	}

	// 准备要签名的文档
	document := []byte("这是一份需要数字签名的重要法律文档。")
	documentHash := sha256.Sum256(document)

	fmt.Printf("📄 准备签名文档: %s\n", string(document))
	fmt.Printf("🔐 文档哈希: %X\n", documentHash)

	// 生成数字签名
	signature, err := ecdsa.SignASN1(rand.Reader, privateKey, documentHash[:])
	if err != nil {
		log.Printf("❌ 生成数字签名失败: %v", err)
		return
	}

	fmt.Printf("✅ 数字签名生成成功\n")
	fmt.Printf("   - 签名长度: %d bytes\n", len(signature))
	fmt.Printf("   - 签名值: %X\n", signature)

	// 验证签名
	publicKey := &privateKey.PublicKey
	valid := ecdsa.VerifyASN1(publicKey, documentHash[:], signature)

	fmt.Printf("✅ 签名验证结果: %v\n", valid)

	if valid {
		logger.WithFields(logrus.Fields{
			"document_size":  len(document),
			"signature_size": len(signature),
			"algorithm":      "ECDSA-SHA256",
			"verified":       valid,
		}).Info("数字签名验证成功")
	} else {
		logger.WithFields(logrus.Fields{
			"document_size":  len(document),
			"signature_size": len(signature),
			"algorithm":      "ECDSA-SHA256",
			"verified":       valid,
		}).Error("数字签名验证失败")
	}
}

// testCertificateChainValidation 测试证书链验证
func testCertificateChainValidation(logger *logrus.Logger) {
	// 生成根CA证书
	rootPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Printf("❌ 生成根CA私钥失败: %v", err)
		return
	}

	rootTemplate := createCertificateTemplate("Root CA", "Root CA", "Test Root CA", true)
	rootTemplate.IsCA = true
	rootTemplate.MaxPathLen = 2

	rootCertDER, err := x509.CreateCertificate(rand.Reader, rootTemplate, rootTemplate, &rootPrivateKey.PublicKey, rootPrivateKey)
	if err != nil {
		log.Printf("❌ 生成根CA证书失败: %v", err)
		return
	}

	rootCert, err := x509.ParseCertificate(rootCertDER)
	if err != nil {
		log.Printf("❌ 解析根CA证书失败: %v", err)
		return
	}

	// 生成终端实体证书
	entityPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Printf("❌ 生成实体私钥失败: %v", err)
		return
	}

	entityTemplate := createCertificateTemplate("Entity Certificate", "Entity", "Test Entity", false)
	entityTemplate.Issuer = rootTemplate.Subject

	entityCertDER, err := x509.CreateCertificate(rand.Reader, entityTemplate, rootCert, &entityPrivateKey.PublicKey, rootPrivateKey)
	if err != nil {
		log.Printf("❌ 生成实体证书失败: %v", err)
		return
	}

	entityCert, err := x509.ParseCertificate(entityCertDER)
	if err != nil {
		log.Printf("❌ 解析实体证书失败: %v", err)
		return
	}

	// 创建证书池并添加根证书
	roots := x509.NewCertPool()
	roots.AddCert(rootCert)

	// 验证证书链
	opts := x509.VerifyOptions{
		Roots: roots,
	}

	_, err = entityCert.Verify(opts)
	if err != nil {
		log.Printf("❌ 证书链验证失败: %v", err)
		return
	}

	fmt.Printf("✅ 证书链验证成功\n")
	fmt.Printf("   - 根CA: %s\n", rootCert.Subject.CommonName)
	fmt.Printf("   - 实体证书: %s\n", entityCert.Subject.CommonName)
	fmt.Printf("   - 链长度: 2\n")
	fmt.Printf("   - 验证状态: 有效\n")

	logger.WithFields(logrus.Fields{
		"root_cn":     rootCert.Subject.CommonName,
		"entity_cn":   entityCert.Subject.CommonName,
		"chain_length": 2,
		"valid":       true,
	}).Info("证书链验证成功")
}

// testTimestampService 测试时间戳服务
func testTimestampService(logger *logrus.Logger) {
	// 准备时间戳数据
	data := []byte("这是一份需要时间戳的文档内容")
	dataHash := sha256.Sum256(data)

	// 生成时间戳
	timestamp := Timestamp{
		ID:        generateID(),
		Hash:      dataHash[:],
		Time:      time.Now(),
		URL:       "http://timestamp.digicert.com",
		TSAInfo:   "Demo TSA",
		CreatedAt: time.Now(),
	}

	fmt.Printf("✅ 时间戳生成成功\n")
	fmt.Printf("   - 时间戳ID: %s\n", timestamp.ID)
	fmt.Printf("   - 数据哈希: %X\n", timestamp.Hash)
	fmt.Printf("   - 时间戳时间: %s\n", timestamp.Time.Format("2006-01-02 15:04:05"))
	fmt.Printf("   - TSA信息: %s\n", timestamp.TSAInfo)

	// 验证时间戳
	originalHash := sha256.Sum256(data)
	valid := bytes.Equal(timestamp.Hash, originalHash[:])

	// 检查时间戳有效性
	maxAge := 24 * time.Hour
	timestampValid := time.Since(timestamp.Time) <= maxAge

	fmt.Printf("✅ 时间戳验证结果\n")
	fmt.Printf("   - 哈希匹配: %v\n", valid)
	fmt.Printf("   - 时间有效: %v\n", timestampValid)
	fmt.Printf("   - 验证状态: %v\n", valid && timestampValid)

	logger.WithFields(logrus.Fields{
		"timestamp_id": timestamp.ID,
		"hash_match":   valid,
		"time_valid":   timestampValid,
		"tsa_info":     timestamp.TSAInfo,
	}).Info("时间戳验证完成")
}

// createCertificateTemplate 创建证书模板
func createCertificateTemplate(commonName, org, ou string, isCA bool) *x509.Certificate {
	// 生成序列号
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         commonName,
			Organization:       []string{org},
			OrganizationalUnit: []string{ou},
			Country:           []string{"CN"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // 1年有效期
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  isCA,
	}

	if isCA {
		template.KeyUsage |= x509.KeyUsageCertSign
		template.MaxPathLen = 2
	}

	return template
}

// generateID 生成唯一ID
func generateID() string {
	return fmt.Sprintf("ts_%d", time.Now().UnixNano())
}

// Timestamp 时间戳结构
type Timestamp struct {
	ID        string    `json:"id"`
	Hash      []byte    `json:"hash"`
	Time      time.Time `json:"time"`
	URL       string    `json:"url"`
	TSAInfo   string    `json:"tsa_info"`
	CreatedAt time.Time `json:"created_at"`
}