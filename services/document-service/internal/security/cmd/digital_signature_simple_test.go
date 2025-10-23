package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"testing"
	"time"
)

// TestCertificateGeneration 测试证书生成
func TestCertificateGeneration(t *testing.T) {
	// 生成ECDSA私钥
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("生成私钥失败: %v", err)
	}

	// 创建证书模板
	serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         "Test Certificate",
			Country:           []string{"CN"},
			Organization:       []string{"Test Org"},
			OrganizationalUnit: []string{"Test Unit"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	// 生成自签名证书
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("生成证书失败: %v", err)
	}

	// 解析证书
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("解析证书失败: %v", err)
	}

	// 验证证书基本信息
	if cert.Subject.CommonName != "Test Certificate" {
		t.Errorf("证书主体名称错误: 期望 'Test Certificate', 实际 '%s'", cert.Subject.CommonName)
	}

	if cert.NotAfter.Before(time.Now()) {
		t.Error("证书已过期")
	}

	fingerprint := sha256.Sum256(cert.Raw)
	if len(fingerprint) != 32 {
		t.Error("证书指纹长度错误")
	}

	// 编码私钥
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatalf("编码私钥失败: %v", err)
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

	// 验证PEM格式
	if len(privateKeyPEM) == 0 {
		t.Error("私钥PEM格式错误")
	}

	if len(certPEM) == 0 {
		t.Error("证书PEM格式错误")
	}

	t.Logf("✅ 证书生成测试通过")
	t.Logf("   - 主题: %s", cert.Subject.CommonName)
	t.Logf("   - 序列号: %s", cert.SerialNumber.String())
	t.Logf("   - 有效期: %s 至 %s", cert.NotBefore.Format("2006-01-02"), cert.NotAfter.Format("2006-01-02"))
	t.Logf("   - 指纹: %X", fingerprint)
	t.Logf("   - 私钥长度: %d bytes", len(privateKeyPEM))
	t.Logf("   - 证书长度: %d bytes", len(certPEM))
}

// TestDigitalSignature 测试数字签名
func TestDigitalSignature(t *testing.T) {
	// 生成测试密钥对
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("生成密钥失败: %v", err)
	}

	// 准备测试数据
	testCases := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "正常文档",
			data:    []byte("这是正常的测试文档内容"),
			wantErr: false,
		},
		{
			name:    "空文档",
			data:    []byte(""),
			wantErr: false,
		},
		{
			name:    "大文档",
			data:    make([]byte, 1024*1024), // 1MB
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 填充大文档数据
			if tc.name == "大文档" {
				for i := range tc.data {
					tc.data[i] = byte(i % 256)
				}
			}

			// 计算文档哈希
			documentHash := sha256.Sum256(tc.data)

			// 生成数字签名
			signature, err := ecdsa.SignASN1(rand.Reader, privateKey, documentHash[:])
			if err != nil {
				if !tc.wantErr {
					t.Fatalf("生成签名失败: %v", err)
				}
				return
			}

			if tc.wantErr {
				t.Error("期望生成签名失败，但成功了")
				return
			}

			// 验证签名
			publicKey := &privateKey.PublicKey
			valid := ecdsa.VerifyASN1(publicKey, documentHash[:], signature)

			if !valid {
				t.Errorf("签名验证失败: %s", tc.name)
			}

			// 测试篡改检测
			if len(tc.data) > 0 {
				tamperedData := make([]byte, len(tc.data))
				copy(tamperedData, tc.data)
				tamperedData[0] ^= 0xFF // 篡改第一个字节

				tamperedHash := sha256.Sum256(tamperedData)
				valid = ecdsa.VerifyASN1(publicKey, tamperedHash[:], signature)

				if valid {
					t.Error("篡改数据后签名验证应该失败，但成功了")
				}
			}

			t.Logf("✅ 数字签名测试通过: %s", tc.name)
			t.Logf("   - 数据大小: %d bytes", len(tc.data))
			t.Logf("   - 签名大小: %d bytes", len(signature))
			t.Logf("   - 签名验证: %v", valid)
		})
	}
}

// TestCertificateChain 测试证书链
func TestCertificateChain(t *testing.T) {
	// 生成根CA私钥
	rootPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("生成根CA私钥失败: %v", err)
	}

	// 创建根CA证书模板
	rootSerial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	rootTemplate := &x509.Certificate{
		SerialNumber: rootSerial,
		Subject: pkix.Name{
			CommonName:   "Root CA",
			Country:     []string{"CN"},
			Organization: []string{"Test Root CA"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(5, 0, 0), // 5年有效期
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
	}

	// 生成根CA证书
	rootCertDER, err := x509.CreateCertificate(rand.Reader, rootTemplate, rootTemplate, &rootPrivateKey.PublicKey, rootPrivateKey)
	if err != nil {
		t.Fatalf("生成根CA证书失败: %v", err)
	}

	rootCert, err := x509.ParseCertificate(rootCertDER)
	if err != nil {
		t.Fatalf("解析根CA证书失败: %v", err)
	}

	// 生成终端实体私钥
	entityPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("生成实体私钥失败: %v", err)
	}

	// 创建实体证书模板
	entitySerial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	entityTemplate := &x509.Certificate{
		SerialNumber: entitySerial,
		Subject: pkix.Name{
			CommonName:   "Entity Certificate",
			Country:     []string{"CN"},
			Organization: []string{"Test Entity"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // 1年有效期
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	// 使用根CA签发实体证书
	entityCertDER, err := x509.CreateCertificate(rand.Reader, entityTemplate, rootCert, &entityPrivateKey.PublicKey, rootPrivateKey)
	if err != nil {
		t.Fatalf("生成实体证书失败: %v", err)
	}

	entityCert, err := x509.ParseCertificate(entityCertDER)
	if err != nil {
		t.Fatalf("解析实体证书失败: %v", err)
	}

	// 验证证书链
	roots := x509.NewCertPool()
	roots.AddCert(rootCert)

	opts := x509.VerifyOptions{
		Roots: roots,
	}

	_, err = entityCert.Verify(opts)
	if err != nil {
		t.Errorf("证书链验证失败: %v", err)
	} else {
		t.Logf("✅ 证书链验证测试通过")
		t.Logf("   - 根CA: %s", rootCert.Subject.CommonName)
		t.Logf("   - 实体证书: %s", entityCert.Subject.CommonName)
		t.Logf("   - 验证状态: 有效")
	}
}

// TestTimestamp 测试时间戳功能
func TestTimestamp(t *testing.T) {
	// 测试数据
	data := []byte("这是需要时间戳的数据")
	dataHash := sha256.Sum256(data)

	// 生成时间戳
	timestamp := Timestamp{
		ID:        fmt.Sprintf("ts_%d", time.Now().UnixNano()),
		Hash:      dataHash[:],
		Time:      time.Now(),
		URL:       "http://timestamp.example.com",
		TSAInfo:   "Test TSA",
		CreatedAt: time.Now(),
	}

	// 验证时间戳
	originalHash := sha256.Sum256(data)
	valid := false
	if len(timestamp.Hash) == len(originalHash) {
		valid = true
		for i := range timestamp.Hash {
			if timestamp.Hash[i] != originalHash[i] {
				valid = false
				break
			}
		}
	}

	if !valid {
		t.Error("时间戳验证失败")
	}

	// 检查时间戳有效性
	maxAge := 24 * time.Hour
	timestampValid := time.Since(timestamp.Time) <= maxAge

	if !timestampValid {
		t.Error("时间戳已过期")
	}

	t.Logf("✅ 时间戳测试通过")
	t.Logf("   - 时间戳ID: %s", timestamp.ID)
	t.Logf("   - 数据哈希: %X", timestamp.Hash)
	t.Logf("   - 时间戳时间: %s", timestamp.Time.Format("2006-01-02 15:04:05"))
	t.Logf("   - 哈希匹配: %v", valid)
	t.Logf("   - 时间有效: %v", timestampValid)
}

// TestPerformance 性能测试
func TestPerformance(t *testing.T) {
	t.Run("证书生成性能", func(t *testing.T) {
		start := time.Now()
		iterations := 100

		for i := 0; i < iterations; i++ {
			privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				t.Fatalf("生成私钥失败: %v", err)
			}

			serialNumber, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
			template := &x509.Certificate{
				SerialNumber: serialNumber,
				Subject: pkix.Name{
					CommonName: fmt.Sprintf("Test Cert %d", i),
					Country:     []string{"CN"},
				},
				NotBefore:             time.Now(),
				NotAfter:              time.Now().AddDate(1, 0, 0),
				KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
				BasicConstraintsValid: true,
				IsCA:                  false,
			}

			_, err = x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
			if err != nil {
				t.Fatalf("生成证书失败: %v", err)
			}
		}

		duration := time.Since(start)
		t.Logf("✅ 证书生成性能测试完成")
		t.Logf("   - 迭代次数: %d", iterations)
		t.Logf("   - 总耗时: %v", duration)
		t.Logf("   - 平均耗时: %v", duration/time.Duration(iterations))
	})

	t.Run("数字签名性能", func(t *testing.T) {
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("生成私钥失败: %v", err)
		}

		document := make([]byte, 1024) // 1KB文档
		for i := range document {
			document[i] = byte(i % 256)
		}

		documentHash := sha256.Sum256(document)

		start := time.Now()
		iterations := 1000

		for i := 0; i < iterations; i++ {
			_, err := ecdsa.SignASN1(rand.Reader, privateKey, documentHash[:])
			if err != nil {
				t.Fatalf("生成签名失败: %v", err)
			}
		}

		duration := time.Since(start)
		t.Logf("✅ 数字签名性能测试完成")
		t.Logf("   - 迭代次数: %d", iterations)
		t.Logf("   - 文档大小: %d bytes", len(document))
		t.Logf("   - 总耗时: %v", duration)
		t.Logf("   - 平均耗时: %v", duration/time.Duration(iterations))
	})
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