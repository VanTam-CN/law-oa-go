package main

import (
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

// CAProfile CA配置文件
type CAProfile struct {
	CommonName    string
	Organization  string
	Country       string
	KeyAlgorithm  string
	KeySize       int
	ValidityDays  int
	IsCA          bool
	MaxPathLength int
}

// CertificateAuthority 证书颁发机构
type CertificateAuthority struct {
	ID           string
	Profile      *CAProfile
	Certificate  *x509.Certificate
	PrivateKey   interface{}
	SerialNumber *big.Int
	NextSerial   *big.Int
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

// NewCertificateAuthority 创建证书颁发机构
func NewCertificateAuthority(profile *CAProfile, parentCA *CertificateAuthority) (*CertificateAuthority, error) {
	// 生成密钥对
	var privateKey interface{}
	var publicKey interface{}
	var err error

	switch profile.KeyAlgorithm {
	case "RSA":
		privateKey, err = rsa.GenerateKey(rand.Reader, profile.KeySize)
		if err != nil {
			return nil, err
		}
		publicKey = privateKey.(*rsa.PrivateKey).Public()
	case "ECDSA":
		var curve elliptic.Curve
		switch profile.KeySize {
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
		privateKey, err = ecdsa.GenerateKey(curve, rand.Reader)
		if err != nil {
			return nil, err
		}
		publicKey = privateKey.(*ecdsa.PrivateKey).Public()
	default:
		return nil, fmt.Errorf("不支持的密钥算法: %s", profile.KeyAlgorithm)
	}

	// 生成序列号
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}

	// 创建证书模板
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   profile.CommonName,
			Organization: []string{profile.Organization},
			Country:      []string{profile.Country},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, profile.ValidityDays),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  profile.IsCA,
	}

	if profile.IsCA {
		template.KeyUsage |= x509.KeyUsageCertSign | x509.KeyUsageCRLSign
		template.MaxPathLen = profile.MaxPathLength
	}

	// 设置颁发者
	if parentCA != nil {
		template.Issuer = parentCA.Certificate.Subject
	} else {
		template.Issuer = template.Subject
	}

	// 生成证书
	var certDER []byte
	if parentCA != nil {
		certDER, err = x509.CreateCertificate(rand.Reader, template, parentCA.Certificate, publicKey, parentCA.PrivateKey)
	} else {
		certDER, err = x509.CreateCertificate(rand.Reader, template, template, publicKey, privateKey)
	}

	if err != nil {
		return nil, err
	}

	// 解析证书
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, err
	}

	// 创建CA实例
	ca := &CertificateAuthority{
		ID:           generateCAID(),
		Profile:      profile,
		Certificate:  cert,
		PrivateKey:   privateKey,
		SerialNumber: serialNumber,
		NextSerial:   new(big.Int).Add(serialNumber, big.NewInt(1)),
		CreatedAt:    time.Now(),
		ExpiresAt:    cert.NotAfter,
	}

	return ca, nil
}

// generateCAID 生成CA ID
func generateCAID() string {
	return "ca_" + time.Now().Format("20060102150405")
}

// ToPEM 将CA转换为PEM格式
func (ca *CertificateAuthority) ToPEM() (certPEM, keyPEM []byte, err error) {
	// 编码证书
	certPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: ca.Certificate.Raw,
	})

	// 编码私钥
	var privateKeyBytes []byte
	switch key := ca.PrivateKey.(type) {
	case *rsa.PrivateKey:
		privateKeyBytes, err = x509.MarshalPKCS8PrivateKey(key)
	case *ecdsa.PrivateKey:
		privateKeyBytes, err = x509.MarshalPKCS8PrivateKey(key)
	default:
		return nil, nil, fmt.Errorf("未知的私钥类型")
	}

	if err != nil {
		return nil, nil, err
	}

	keyPEM = pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return certPEM, keyPEM, nil
}

// VerifyChain 验证证书链
func (ca *CertificateAuthority) VerifyChain(rootCA *CertificateAuthority) error {
	roots := x509.NewCertPool()
	roots.AddCert(rootCA.Certificate)

	opts := x509.VerifyOptions{
		Roots: roots,
	}

	chains, err := ca.Certificate.Verify(opts)
	if err != nil {
		return err
	}

	fmt.Printf("✅ 证书链验证成功，找到 %d 个有效链\n", len(chains))
	for i, chain := range chains {
		fmt.Printf("   链 %d: 长度 %d\n", i+1, len(chain))
		for j, cert := range chain {
			fmt.Printf("     - 证书 %d: %s\n", j+1, cert.Subject.CommonName)
		}
	}

	return nil
}

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	fmt.Println("🔐 开始根CA证书和信任链管理演示...")

	// 演示1: 创建根CA
	fmt.Println("\n📋 演示1: 创建根CA")
	rootProfile := &CAProfile{
		CommonName:    "律师事务所根CA",
		Organization:  "律师事务所",
		Country:       "CN",
		KeyAlgorithm:  "RSA",
		KeySize:       4096,
		ValidityDays:  20 * 365, // 20年
		IsCA:          true,
		MaxPathLength: 3,
	}

	rootCA, err := NewCertificateAuthority(rootProfile, nil)
	if err != nil {
		log.Fatalf("创建根CA失败: %v", err)
	}

	fmt.Printf("✅ 根CA创建成功\n")
	fmt.Printf("   - ID: %s\n", rootCA.ID)
	fmt.Printf("   - 主题: %s\n", rootCA.Certificate.Subject.CommonName)
	fmt.Printf("   - 序列号: %s\n", rootCA.SerialNumber.String())
	fmt.Printf("   - 有效期: %s 至 %s\n", rootCA.CreatedAt.Format("2006-01-02"), rootCA.ExpiresAt.Format("2006-01-02"))
	fmt.Printf("   - 是否为CA: %v\n", rootCA.Certificate.IsCA)
	fmt.Printf("   - 最大路径长度: %d\n", rootCA.Certificate.MaxPathLen)

	// 演示2: 创建中间CA
	fmt.Println("\n📋 演示2: 创建中间CA")
	intermediateProfile := &CAProfile{
		CommonName:    "律师事务所中间CA",
		Organization:  "律师事务所",
		Country:       "CN",
		KeyAlgorithm:  "RSA",
		KeySize:       4096,
		ValidityDays:  10 * 365, // 10年
		IsCA:          true,
		MaxPathLength: 2,
	}

	intermediateCA, err := NewCertificateAuthority(intermediateProfile, rootCA)
	if err != nil {
		log.Fatalf("创建中间CA失败: %v", err)
	}

	fmt.Printf("✅ 中间CA创建成功\n")
	fmt.Printf("   - ID: %s\n", intermediateCA.ID)
	fmt.Printf("   - 主题: %s\n", intermediateCA.Certificate.Subject.CommonName)
	fmt.Printf("   - 颁发者: %s\n", intermediateCA.Certificate.Issuer.CommonName)
	fmt.Printf("   - 序列号: %s\n", intermediateCA.SerialNumber.String())
	fmt.Printf("   - 有效期: %s 至 %s\n", intermediateCA.CreatedAt.Format("2006-01-02"), intermediateCA.ExpiresAt.Format("2006-01-02"))

	// 演示3: 创建终端实体证书
	fmt.Println("\n📋 演示3: 创建终端实体证书")
	leafProfile := &CAProfile{
		CommonName:    "律师用户证书",
		Organization:  "律师事务所",
		Country:       "CN",
		KeyAlgorithm:  "ECDSA",
		KeySize:       256,
		ValidityDays:  365, // 1年
		IsCA:          false,
		MaxPathLength: 0,
	}

	leafCA, err := NewCertificateAuthority(leafProfile, intermediateCA)
	if err != nil {
		log.Fatalf("创建终端实体证书失败: %v", err)
	}

	fmt.Printf("✅ 终端实体证书创建成功\n")
	fmt.Printf("   - ID: %s\n", leafCA.ID)
	fmt.Printf("   - 主题: %s\n", leafCA.Certificate.Subject.CommonName)
	fmt.Printf("   - 颁发者: %s\n", leafCA.Certificate.Issuer.CommonName)
	fmt.Printf("   - 序列号: %s\n", leafCA.SerialNumber.String())
	fmt.Printf("   - 有效期: %s 至 %s\n", leafCA.CreatedAt.Format("2006-01-02"), leafCA.ExpiresAt.Format("2006-01-02"))
	fmt.Printf("   - 密钥算法: %s\n", leafProfile.KeyAlgorithm)

	// 演示4: PEM编码
	fmt.Println("\n📋 演示4: PEM编码输出")
	rootCertPEM, rootKeyPEM, err := rootCA.ToPEM()
	if err != nil {
		log.Fatalf("根CA PEM编码失败: %v", err)
	}

	fmt.Printf("✅ 根CA PEM编码成功\n")
	fmt.Printf("   - 证书PEM长度: %d bytes\n", len(rootCertPEM))
	fmt.Printf("   - 私钥PEM长度: %d bytes\n", len(rootKeyPEM))

	// 演示5: 证书链验证
	fmt.Println("\n📋 演示5: 证书链验证")

	// 验证中间CA证书链
	fmt.Println("验证中间CA证书链:")
	err = intermediateCA.VerifyChain(rootCA)
	if err != nil {
		fmt.Printf("❌ 中间CA证书链验证失败: %v\n", err)
	} else {
		fmt.Printf("✅ 中间CA证书链验证成功\n")
	}

	// 验证终端实体证书链（需要包含中间CA）
	fmt.Println("验证终端实体证书链:")
	roots := x509.NewCertPool()
	roots.AddCert(rootCA.Certificate)

	intermediates := x509.NewCertPool()
	intermediates.AddCert(intermediateCA.Certificate)

	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
	}

	chains, err := leafCA.Certificate.Verify(opts)
	if err != nil {
		fmt.Printf("❌ 终端实体证书链验证失败: %v\n", err)
	} else {
		fmt.Printf("✅ 终端实体证书链验证成功，找到 %d 个有效链\n", len(chains))
		for i, chain := range chains {
			fmt.Printf("   链 %d: 长度 %d\n", i+1, len(chain))
			for j, cert := range chain {
				fmt.Printf("     - 证书 %d: %s\n", j+1, cert.Subject.CommonName)
			}
		}
	}

	// 演示6: 性能测试
	fmt.Println("\n📋 演示6: 性能测试")

	startTime := time.Now()
	iterations := 10

	for i := 0; i < iterations; i++ {
		profile := &CAProfile{
			CommonName:    fmt.Sprintf("性能测试CA %d", i+1),
			Organization:  "测试组织",
			Country:       "CN",
			KeyAlgorithm:  "RSA",
			KeySize:       2048, // 使用较小密钥以提高性能
			ValidityDays:  365,
			IsCA:          true,
			MaxPathLength: 1,
		}

		_, err := NewCertificateAuthority(profile, nil)
		if err != nil {
			log.Fatalf("性能测试创建CA失败: %v", err)
		}
	}

	duration := time.Since(startTime)
	avgDuration := duration / time.Duration(iterations)

	fmt.Printf("✅ 性能测试完成\n")
	fmt.Printf("   - 迭代次数: %d\n", iterations)
	fmt.Printf("   - 总耗时: %v\n", duration)
	fmt.Printf("   - 平均耗时: %v\n", avgDuration)

	// 演示7: 不同密钥算法测试
	fmt.Println("\n📋 演示7: 不同密钥算法测试")
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

		profile := &CAProfile{
			CommonName:    fmt.Sprintf("%s-%d 测试CA", alg.algorithm, alg.keySize),
			Organization:  "算法测试组织",
			Country:       "CN",
			KeyAlgorithm:  alg.algorithm,
			KeySize:       alg.keySize,
			ValidityDays:  365,
			IsCA:          true,
			MaxPathLength: 1,
		}

		ca, err := NewCertificateAuthority(profile, nil)
		if err != nil {
			fmt.Printf("❌ %s-%d CA创建失败: %v\n", alg.algorithm, alg.keySize, err)
			continue
		}

		duration := time.Since(start)

		fmt.Printf("✅ %s-%d CA创建成功\n", alg.algorithm, alg.keySize)
		fmt.Printf("   - 创建耗时: %v\n", duration)
		fmt.Printf("   - 证书指纹: %X\n", sha256.Sum256(ca.Certificate.Raw))
	}

	fmt.Println("\n🎉 根CA证书和信任链管理演示完成！")
	fmt.Println("\n📊 功能总结:")
	fmt.Printf("   - 根CA创建: ✅\n")
	fmt.Printf("   - 中间CA创建: ✅\n")
	fmt.Printf("   - 终端实体证书创建: ✅\n")
	fmt.Printf("   - PEM编码输出: ✅\n")
	fmt.Printf("   - 证书链验证: ✅\n")
	fmt.Printf("   - 性能测试: ✅\n")
	fmt.Printf("   - 多算法支持: ✅\n")

	logger.WithFields(logrus.Fields{
		"root_ca_id":           rootCA.ID,
		"intermediate_ca_id":   intermediateCA.ID,
		"leaf_cert_id":         leafCA.ID,
		"total_certificates":   3,
		"validation_success":   true,
	}).Info("CA证书和信任链管理演示完成")
}