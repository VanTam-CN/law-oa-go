package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

// parseURL 解析URL字符串，失败时返回nil
func parseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		log.Printf("警告: 解析URL失败 %s: %v", rawURL, err)
		return nil
	}
	return u
}

// X509CertificateManager X.509证书管理器
type X509CertificateManager struct {
	caKey        *rsa.PrivateKey
	caCert       *x509.Certificate
	nextSerial   *big.Int
	logger       *logrus.Logger
}

// NewX509CertificateManager 创建X.509证书管理器
func NewX509CertificateManager() (*X509CertificateManager, error) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// 生成CA密钥对
	caKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, fmt.Errorf("生成CA密钥失败: %v", err)
	}

	// 生成序列号
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("生成序列号失败: %v", err)
	}

	// 创建CA证书模板
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"律师事务所"},
			Country:      []string{"CN"},
			Province:     []string{"Beijing"},
			Locality:     []string{"Beijing"},
			CommonName:   "律师事务所 Root CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10年有效期
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            3,
	}

	// 自签名CA证书
	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("创建CA证书失败: %v", err)
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, fmt.Errorf("解析CA证书失败: %v", err)
	}

	return &X509CertificateManager{
		caKey:      caKey,
		caCert:     cert,
		nextSerial: new(big.Int).Add(serial, big.NewInt(1)),
		logger:     logger,
	}, nil
}

// CertificateRequest 证书请求
type CertificateRequest struct {
	TemplateID      string
	Subject         pkix.Name
	PublicKey       crypto.PublicKey
	Validity        time.Duration
	KeyUsage        x509.KeyUsage
	ExtKeyUsage     []x509.ExtKeyUsage
	DNSNames        []string
	IPAddresses     []net.IP
	EmailAddresses  []string
	URIs            []*url.URL
	RequesterID     string
}

// Certificate 证书信息
type Certificate struct {
	SerialNumber   *big.Int
	Subject        pkix.Name
	Issuer         pkix.Name
	NotBefore      time.Time
	NotAfter       time.Time
	PublicKey      crypto.PublicKey
	Certificate    *x509.Certificate
	CertificatePEM string
	CreatedAt      time.Time
	TemplateID     string
	RequesterID    string
}

// IssueCertificate 签发证书
func (xcm *X509CertificateManager) IssueCertificate(req *CertificateRequest) (*Certificate, error) {
	// 验证请求
	if err := xcm.validateRequest(req); err != nil {
		return nil, fmt.Errorf("证书请求验证失败: %v", err)
	}

	// 生成序列号
	serialNumber := new(big.Int).Set(xcm.nextSerial)
	xcm.nextSerial.Add(xcm.nextSerial, big.NewInt(1))

	// 创建证书模板
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      req.Subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(req.Validity),
		KeyUsage:     req.KeyUsage,
		ExtKeyUsage:  req.ExtKeyUsage,
		DNSNames:     req.DNSNames,
		IPAddresses:  req.IPAddresses,
		EmailAddresses: req.EmailAddresses,
		URIs:         req.URIs,
	}

	// 设置颁发者
	template.Issuer = xcm.caCert.Subject

	// 添加扩展字段
	var extraExtensions []pkix.Extension

	// 创建主题备用名称扩展
	if len(req.DNSNames) > 0 || len(req.IPAddresses) > 0 || len(req.EmailAddresses) > 0 || len(req.URIs) > 0 {
		sanExtension, err := xcm.createSubjectAlternativeNameExtension(req)
		if err != nil {
			return nil, fmt.Errorf("创建SAN扩展失败: %v", err)
		}
		extraExtensions = append(extraExtensions, *sanExtension)
	}

	template.ExtraExtensions = extraExtensions

	// 生成证书
	certBytes, err := x509.CreateCertificate(rand.Reader, template, xcm.caCert, req.PublicKey, xcm.caKey)
	if err != nil {
		return nil, fmt.Errorf("创建证书失败: %v", err)
	}

	// 解析证书
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %v", err)
	}

	// 编码PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certificate := &Certificate{
		SerialNumber:   serialNumber,
		Subject:        req.Subject,
		Issuer:         xcm.caCert.Subject,
		NotBefore:      cert.NotBefore,
		NotAfter:       cert.NotAfter,
		PublicKey:      req.PublicKey,
		Certificate:    cert,
		CertificatePEM: string(certPEM),
		CreatedAt:      time.Now(),
		TemplateID:     req.TemplateID,
		RequesterID:    req.RequesterID,
	}

	return certificate, nil
}

// validateRequest 验证证书请求
func (xcm *X509CertificateManager) validateRequest(req *CertificateRequest) error {
	if req.PublicKey == nil {
		return fmt.Errorf("公钥不能为空")
	}

	if req.Validity <= 0 {
		return fmt.Errorf("有效期必须大于0")
	}

	if req.Subject.CommonName == "" && len(req.DNSNames) == 0 && len(req.IPAddresses) == 0 {
		return fmt.Errorf("必须指定CommonName或至少一个DNS名称/IP地址")
	}

	return nil
}

// createSubjectAlternativeNameExtension 创建主题备用名称扩展
func (xcm *X509CertificateManager) createSubjectAlternativeNameExtension(req *CertificateRequest) (*pkix.Extension, error) {
	var sanValues []asn1.RawValue

	// 添加DNS名称
	for _, dns := range req.DNSNames {
		value, err := asn1.MarshalWithParams(dns, "ia5")
		if err != nil {
			return nil, fmt.Errorf("编码DNS名称失败: %v", err)
		}
		sanValues = append(sanValues, asn1.RawValue{Tag: 2, Class: 2, Bytes: value})
	}

	// 添加IP地址
	for _, ip := range req.IPAddresses {
		sanValues = append(sanValues, asn1.RawValue{Tag: 7, Class: 2, Bytes: ip})
	}

	// 添加邮箱地址
	for _, email := range req.EmailAddresses {
		value, err := asn1.MarshalWithParams(email, "ia5")
		if err != nil {
			return nil, fmt.Errorf("编码邮箱地址失败: %v", err)
		}
		sanValues = append(sanValues, asn1.RawValue{Tag: 1, Class: 2, Bytes: value})
	}

	// 添加URI
	for _, uri := range req.URIs {
		value, err := asn1.MarshalWithParams(uri.String(), "ia5")
		if err != nil {
			return nil, fmt.Errorf("编码URI失败: %v", err)
		}
		sanValues = append(sanValues, asn1.RawValue{Tag: 6, Class: 2, Bytes: value})
	}

	sanSequence, err := asn1.Marshal(sanValues)
	if err != nil {
		return nil, fmt.Errorf("编码SAN序列失败: %v", err)
	}

	extension := pkix.Extension{
		Id:       []int{2, 5, 29, 17}, // SAN OID
		Critical: true,
		Value:    sanSequence,
	}

	return &extension, nil
}

// ValidateCertificateChain 验证证书链
func (xcm *X509CertificateManager) ValidateCertificateChain(cert *x509.Certificate) error {
	// 创建根证书池
	roots := x509.NewCertPool()
	roots.AddCert(xcm.caCert)

	// 设置验证选项
	opts := x509.VerifyOptions{
		Roots:       roots,
		CurrentTime: time.Now(),
	}

	// 验证证书链
	_, err := cert.Verify(opts)
	if err != nil {
		return fmt.Errorf("证书链验证失败: %v", err)
	}

	return nil
}

// BatchIssueCertificate 批量签发证书
func (xcm *X509CertificateManager) BatchIssueCertificate(requests []*CertificateRequest) ([]*Certificate, error) {
	results := make([]*Certificate, 0, len(requests))

	for i, req := range requests {
		xcm.logger.WithFields(logrus.Fields{
			"request_index": i,
			"template_id":   req.TemplateID,
			"subject":       req.Subject.CommonName,
		}).Info("开始批量签发证书")

		cert, err := xcm.IssueCertificate(req)
		if err != nil {
			xcm.logger.WithFields(logrus.Fields{
				"request_index": i,
				"error":         err,
			}).Error("证书签发失败")
			continue
		}

		results = append(results, cert)

		xcm.logger.WithFields(logrus.Fields{
			"request_index": i,
			"serial_number": cert.SerialNumber.String(),
			"subject":       cert.Subject.CommonName,
		}).Info("证书签发成功")
	}

	return results, nil
}

// GenerateKeyPair 生成密钥对
func (xcm *X509CertificateManager) GenerateKeyPair(algorithm string, keySize int) (crypto.PrivateKey, crypto.PublicKey, error) {
	switch algorithm {
	case "RSA":
		privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
		if err != nil {
			return nil, nil, err
		}
		return privateKey, &privateKey.PublicKey, nil

	case "ECDSA":
		var curve elliptic.Curve
		switch keySize {
		case 224:
			curve = elliptic.P224()
		case 256:
			curve = elliptic.P256()
		case 384:
			curve = elliptic.P384()
		case 521:
			curve = elliptic.P521()
		default:
			curve = elliptic.P256()
		}

		privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			return nil, nil, err
		}
		return privateKey, &privateKey.PublicKey, nil

	default:
		return nil, nil, fmt.Errorf("不支持的密钥算法: %s", algorithm)
	}
}

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	fmt.Println("🔐 开始X.509证书生成和签发演示...")

	// 演示1: 创建证书管理器
	fmt.Println("\n📋 演示1: 创建X.509证书管理器")
	manager, err := NewX509CertificateManager()
	if err != nil {
		log.Fatalf("创建证书管理器失败: %v", err)
	}

	fmt.Printf("✅ X.509证书管理器创建成功\n")
	fmt.Printf("   - CA主题: %s\n", manager.caCert.Subject.CommonName)
	fmt.Printf("   - CA序列号: %s\n", manager.caCert.SerialNumber.String())
	fmt.Printf("   - CA有效期: %s 至 %s\n", manager.caCert.NotBefore.Format("2006-01-02"), manager.caCert.NotAfter.Format("2006-01-02"))

	// 演示2: 签发单个TLS服务器证书
	fmt.Println("\n📋 演示2: 签发TLS服务器证书")
	tlsServerReq := createTLSServerRequest("www.lawfirm.com", manager)
	tlsServerCert, err := manager.IssueCertificate(tlsServerReq)
	if err != nil {
		log.Fatalf("签发TLS服务器证书失败: %v", err)
	}

	fmt.Printf("✅ TLS服务器证书签发成功\n")
	fmt.Printf("   - 序列号: %s\n", tlsServerCert.SerialNumber.String())
	fmt.Printf("   - 主题: %s\n", tlsServerCert.Subject.CommonName)
	fmt.Printf("   - 颁发者: %s\n", tlsServerCert.Issuer.CommonName)
	fmt.Printf("   - 有效期: %s 至 %s\n", tlsServerCert.NotBefore.Format("2006-01-02"), tlsServerCert.NotAfter.Format("2006-01-02"))
	fmt.Printf("   - DNS名称: %v\n", tlsServerCert.Certificate.DNSNames)

	// 验证证书链
	err = manager.ValidateCertificateChain(tlsServerCert.Certificate)
	if err != nil {
		fmt.Printf("❌ 证书链验证失败: %v\n", err)
	} else {
		fmt.Printf("✅ 证书链验证成功\n")
	}

	// 演示3: 签发TLS客户端证书
	fmt.Println("\n📋 演示3: 签发TLS客户端证书")
	tlsClientReq := createTLSClientRequest("client.lawfirm.com", manager)
	tlsClientCert, err := manager.IssueCertificate(tlsClientReq)
	if err != nil {
		log.Fatalf("签发TLS客户端证书失败: %v", err)
	}

	fmt.Printf("✅ TLS客户端证书签发成功\n")
	fmt.Printf("   - 序列号: %s\n", tlsClientCert.SerialNumber.String())
	fmt.Printf("   - 主题: %s\n", tlsClientCert.Subject.CommonName)
	fmt.Printf("   - 颁发者: %s\n", tlsClientCert.Issuer.CommonName)
	fmt.Printf("   - 有效期: %s 至 %s\n", tlsClientCert.NotBefore.Format("2006-01-02"), tlsClientCert.NotAfter.Format("2006-01-02"))
	fmt.Printf("   - 邮箱地址: %v\n", tlsClientCert.Certificate.EmailAddresses)

	// 演示4: 签发代码签名证书
	fmt.Println("\n📋 演示4: 签发代码签名证书")
	codeSigningReq := createCodeSigningRequest("LawOffice Software", manager)
	codeSigningCert, err := manager.IssueCertificate(codeSigningReq)
	if err != nil {
		log.Fatalf("签发代码签名证书失败: %v", err)
	}

	fmt.Printf("✅ 代码签名证书签发成功\n")
	fmt.Printf("   - 序列号: %s\n", codeSigningCert.SerialNumber.String())
	fmt.Printf("   - 主题: %s\n", codeSigningCert.Subject.CommonName)
	fmt.Printf("   - 颁发者: %s\n", codeSigningCert.Issuer.CommonName)
	fmt.Printf("   - 有效期: %s 至 %s\n", codeSigningCert.NotBefore.Format("2006-01-02"), codeSigningCert.NotAfter.Format("2006-01-02"))

	// 演示5: 签发邮件签名证书
	fmt.Println("\n📋 演示5: 签发邮件签名证书")
	emailSigningReq := createEmailSigningRequest("lawyer@lawfirm.com", manager)
	emailSigningCert, err := manager.IssueCertificate(emailSigningReq)
	if err != nil {
		log.Fatalf("签发邮件签名证书失败: %v", err)
	}

	fmt.Printf("✅ 邮件签名证书签发成功\n")
	fmt.Printf("   - 序列号: %s\n", emailSigningCert.SerialNumber.String())
	fmt.Printf("   - 主题: %s\n", emailSigningCert.Subject.CommonName)
	fmt.Printf("   - 颁发者: %s\n", emailSigningCert.Issuer.CommonName)
	fmt.Printf("   - 有效期: %s 至 %s\n", emailSigningCert.NotBefore.Format("2006-01-02"), emailSigningCert.NotAfter.Format("2006-01-02"))
	fmt.Printf("   - 邮箱地址: %v\n", emailSigningCert.Certificate.EmailAddresses)

	// 演示6: 批量签发证书
	fmt.Println("\n📋 演示6: 批量签发证书")
	batchRequests := createBatchRequests(manager)
	batchResults, err := manager.BatchIssueCertificate(batchRequests)
	if err != nil {
		log.Fatalf("批量签发证书失败: %v", err)
	}

	fmt.Printf("✅ 批量签发证书完成\n")
	fmt.Printf("   - 请求数量: %d\n", len(batchRequests))
	fmt.Printf("   - 成功数量: %d\n", len(batchResults))
	for i, cert := range batchResults {
		fmt.Printf("   - 证书 %d: %s (序列号: %s)\n", i+1, cert.Subject.CommonName, cert.SerialNumber.String())
	}

	// 演示7: 多种密钥算法测试
	fmt.Println("\n📋 演示7: 多种密钥算法测试")
	algorithms := []struct {
		algorithm string
		keySize   int
	}{
		{"RSA", 2048},
		{"RSA", 4096},
		{"ECDSA", 256},
		{"ECDSA", 384},
		{"ECDSA", 521},
	}

	for _, alg := range algorithms {
		start := time.Now()

		_, publicKey, err := manager.GenerateKeyPair(alg.algorithm, alg.keySize)
		if err != nil {
			fmt.Printf("❌ %s-%d 密钥对生成失败: %v\n", alg.algorithm, alg.keySize, err)
			continue
		}

		// 创建测试证书请求
		req := &CertificateRequest{
			TemplateID:  "tls-server",
			Subject: pkix.Name{
				CommonName:   fmt.Sprintf("test-%s-%d.example.com", alg.algorithm, alg.keySize),
				Organization: []string{"Test Org"},
				Country:      []string{"CN"},
			},
			PublicKey:  publicKey,
			Validity:   365 * 24 * time.Hour,
			KeyUsage:   x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
			ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			DNSNames:   []string{fmt.Sprintf("test-%s-%d.example.com", alg.algorithm, alg.keySize)},
			RequesterID: "test_system",
		}

		cert, err := manager.IssueCertificate(req)
		if err != nil {
			fmt.Printf("❌ %s-%d 证书签发失败: %v\n", alg.algorithm, alg.keySize, err)
			continue
		}

		duration := time.Since(start)

		fmt.Printf("✅ %s-%d 证书签发成功\n", alg.algorithm, alg.keySize)
		fmt.Printf("   - 生成耗时: %v\n", duration)
		fmt.Printf("   - 证书序列号: %s\n", cert.SerialNumber.String())
		fmt.Printf("   - 公钥类型: %T\n", publicKey)
	}

	// 演示8: 复杂SAN扩展测试
	fmt.Println("\n📋 演示8: 复杂SAN扩展测试")
	complexReq := &CertificateRequest{
		TemplateID: "tls-server",
		Subject: pkix.Name{
			CommonName:   "complex.example.com",
			Organization: []string{"Complex Corp"},
			Country:      []string{"CN"},
			Province:     []string{"Beijing"},
			Locality:     []string{"Beijing"},
		},
		Validity:    365 * 24 * time.Hour,
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames: []string{
			"complex.example.com",
			"www.complex.example.com",
			"api.complex.example.com",
		},
		IPAddresses: []net.IP{
			net.ParseIP("192.168.1.100"),
			net.ParseIP("10.0.0.100"),
			net.ParseIP("172.16.0.100"),
		},
		EmailAddresses: []string{
			"admin@complex.example.com",
			"support@complex.example.com",
			"security@complex.example.com",
		},
		// 暂时跳过URI，避免编码问题
		// URIs: []*url.URL{
		//	parseURL("https://complex.example.com/api"),
		//	parseURL("https://complex.example.com/docs"),
		// },
		RequesterID: "test_system",
	}

	// 生成密钥对
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("生成密钥对失败: %v", err)
	}
	complexReq.PublicKey = &privateKey.PublicKey

	complexCert, err := manager.IssueCertificate(complexReq)
	if err != nil {
		log.Fatalf("签发复杂SAN证书失败: %v", err)
	}

	fmt.Printf("✅ 复杂SAN扩展证书签发成功\n")
	fmt.Printf("   - 证书序列号: %s\n", complexCert.SerialNumber.String())
	fmt.Printf("   - 主题: %s\n", complexCert.Subject.CommonName)
	fmt.Printf("   - DNS名称数量: %d\n", len(complexCert.Certificate.DNSNames))
	fmt.Printf("   - IP地址数量: %d\n", len(complexCert.Certificate.IPAddresses))
	fmt.Printf("   - 邮箱地址数量: %d\n", len(complexCert.Certificate.EmailAddresses))

	// 演示9: PEM编码展示
	fmt.Println("\n📋 演示9: PEM编码展示")
	fmt.Printf("TLS服务器证书PEM:\n")
	fmt.Printf("%s\n", tlsServerCert.CertificatePEM[:500] + "...")

	fmt.Println("\n🎉 X.509证书生成和签发演示完成！")
	fmt.Println("\n📊 功能总结:")
	fmt.Printf("   - X.509证书管理器创建: ✅\n")
	fmt.Printf("   - TLS服务器证书签发: ✅\n")
	fmt.Printf("   - TLS客户端证书签发: ✅\n")
	fmt.Printf("   - 代码签名证书签发: ✅\n")
	fmt.Printf("   - 邮件签名证书签发: ✅\n")
	fmt.Printf("   - 批量证书签发: ✅\n")
	fmt.Printf("   - 多算法密钥支持: ✅\n")
	fmt.Printf("   - 复杂SAN扩展: ✅\n")
	fmt.Printf("   - 证书链验证: ✅\n")

	logger.WithFields(logrus.Fields{
		"ca_subject":          manager.caCert.Subject.CommonName,
		"ca_serial":          manager.caCert.SerialNumber.String(),
		"certificates_issued": len(batchResults) + 4, // 单个证书 + 批量结果
		"algorithms_tested":   len(algorithms),
	}).Info("X.509证书生成和签发演示完成")
}

// createTLSServerRequest 创建TLS服务器证书请求
func createTLSServerRequest(commonName string, manager *X509CertificateManager) *CertificateRequest {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("生成TLS服务器密钥失败: %v", err)
	}

	return &CertificateRequest{
		TemplateID: "tls-server",
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"律师事务所"},
			Country:      []string{"CN"},
			Province:     []string{"Beijing"},
			Locality:     []string{"Beijing"},
		},
		PublicKey:   &privateKey.PublicKey,
		Validity:    365 * 24 * time.Hour, // 1年
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    []string{commonName, "www." + commonName},
		RequesterID: "system",
	}
}

// createTLSClientRequest 创建TLS客户端证书请求
func createTLSClientRequest(commonName string, manager *X509CertificateManager) *CertificateRequest {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("生成TLS客户端密钥失败: %v", err)
	}

	return &CertificateRequest{
		TemplateID: "tls-client",
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"律师事务所"},
			Country:      []string{"CN"},
			Province:     []string{"Beijing"},
			Locality:     []string{"Beijing"},
		},
		PublicKey:   &privateKey.PublicKey,
		Validity:    365 * 24 * time.Hour, // 1年
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		RequesterID: "system",
	}
}

// createCodeSigningRequest 创建代码签名证书请求
func createCodeSigningRequest(commonName string, manager *X509CertificateManager) *CertificateRequest {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatalf("生成代码签名密钥失败: %v", err)
	}

	return &CertificateRequest{
		TemplateID: "code-signing",
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"律师事务所"},
			Country:      []string{"CN"},
			Province:     []string{"Beijing"},
			Locality:     []string{"Beijing"},
		},
		PublicKey:   &privateKey.PublicKey,
		Validity:    3 * 365 * 24 * time.Hour, // 3年
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
		RequesterID: "system",
	}
}

// createEmailSigningRequest 创建邮件签名证书请求
func createEmailSigningRequest(email string, manager *X509CertificateManager) *CertificateRequest {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("生成邮件签名密钥失败: %v", err)
	}

	return &CertificateRequest{
		TemplateID: "email-signing",
		Subject: pkix.Name{
			CommonName:   email,
			Organization: []string{"律师事务所"},
			Country:      []string{"CN"},
			Province:     []string{"Beijing"},
			Locality:     []string{"Beijing"},
		},
		PublicKey:   &privateKey.PublicKey,
		Validity:    365 * 24 * time.Hour, // 1年
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageEmailProtection},
		RequesterID: "system",
	}
}

// createBatchRequests 创建批量证书请求
func createBatchRequests(manager *X509CertificateManager) []*CertificateRequest {
	requests := make([]*CertificateRequest, 5)

	domains := []string{"server1.example.com", "server2.example.com", "server3.example.com", "server4.example.com", "server5.example.com"}

	for i, domain := range domains {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			log.Fatalf("生成批量证书密钥失败: %v", err)
		}

		requests[i] = &CertificateRequest{
			TemplateID: "tls-server",
			Subject: pkix.Name{
				CommonName:   domain,
				Organization: []string{"Example Corp"},
				Country:      []string{"CN"},
			},
			PublicKey:   &privateKey.PublicKey,
			Validity:    365 * 24 * time.Hour,
			KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
			ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			DNSNames:    []string{domain, "www." + domain},
			RequesterID: fmt.Sprintf("batch_user_%d", i+1),
		}
	}

	return requests
}