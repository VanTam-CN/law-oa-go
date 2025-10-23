package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SimpleCAProfile 简化的CA配置文件
type SimpleCAProfile struct {
	CommonName    string
	Organization  string
	Country       string
	KeyAlgorithm  string
	KeySize       int
	ValidityDays  int
	IsCA          bool
	MaxPathLength int
}

// SimpleCA 简化的证书颁发机构
type SimpleCA struct {
	ID           string
	Profile      *SimpleCAProfile
	Certificate  *x509.Certificate
	PrivateKey   interface{}
	SerialNumber *big.Int
	NextSerial   *big.Int
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

// NewSimpleCA 创建简化的CA
func NewSimpleCA(profile *SimpleCAProfile, parentCA *SimpleCA) (*SimpleCA, error) {
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
		return nil, err
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
	ca := &SimpleCA{
		ID:           generateSimpleCAID(),
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

// generateSimpleCAID 生成CA ID
func generateSimpleCAID() string {
	return "simple_ca_" + time.Now().Format("20060102150405")
}

// TestSimpleCACreation 测试简化CA创建
func TestSimpleCACreation(t *testing.T) {
	profile := &SimpleCAProfile{
		CommonName:    "测试根CA",
		Organization:  "测试律师事务所",
		Country:       "CN",
		KeyAlgorithm:  "RSA",
		KeySize:       2048,
		ValidityDays:  365,
		IsCA:          true,
		MaxPathLength: 3,
	}

	// 创建根CA
	rootCA, err := NewSimpleCA(profile, nil)
	require.NoError(t, err)
	require.NotNil(t, rootCA)

	// 验证根CA属性
	assert.Equal(t, "测试根CA", rootCA.Profile.CommonName)
	assert.Equal(t, "测试律师事务所", rootCA.Profile.Organization)
	assert.True(t, rootCA.Profile.IsCA)
	assert.Equal(t, 3, rootCA.Profile.MaxPathLength)
	assert.NotNil(t, rootCA.Certificate)
	assert.NotNil(t, rootCA.PrivateKey)
	assert.NotNil(t, rootCA.SerialNumber)
	assert.True(t, rootCA.CreatedAt.Before(time.Now().Add(time.Minute)))
	assert.True(t, rootCA.ExpiresAt.After(time.Now().AddDate(0, 0, 364)))

	// 验证证书内容
	assert.Equal(t, "测试根CA", rootCA.Certificate.Subject.CommonName)
	assert.Equal(t, []string{"测试律师事务所"}, rootCA.Certificate.Subject.Organization)
	assert.Equal(t, []string{"CN"}, rootCA.Certificate.Subject.Country)
	assert.True(t, rootCA.Certificate.IsCA)
	assert.Equal(t, 3, rootCA.Certificate.MaxPathLen)

	t.Logf("✅ 简化CA创建测试通过")
	t.Logf("   - CA ID: %s", rootCA.ID)
	t.Logf("   - 主题: %s", rootCA.Certificate.Subject.CommonName)
	t.Logf("   - 序列号: %s", rootCA.SerialNumber.String())
	t.Logf("   - 有效期: %s 至 %s", rootCA.CreatedAt.Format("2006-01-02"), rootCA.ExpiresAt.Format("2006-01-02"))
}

// TestSimpleCAChain 测试简化CA链
func TestSimpleCAChain(t *testing.T) {
	// 创建根CA
	rootProfile := &SimpleCAProfile{
		CommonName:    "根CA",
		Organization:  "律师事务所",
		Country:       "CN",
		KeyAlgorithm:  "RSA",
		KeySize:       2048,
		ValidityDays:  365 * 10, // 10年
		IsCA:          true,
		MaxPathLength: 2,
	}

	rootCA, err := NewSimpleCA(rootProfile, nil)
	require.NoError(t, err)

	// 创建中间CA
	intermediateProfile := &SimpleCAProfile{
		CommonName:    "中间CA",
		Organization:  "律师事务所",
		Country:       "CN",
		KeyAlgorithm:  "RSA",
		KeySize:       2048,
		ValidityDays:  365 * 5, // 5年
		IsCA:          true,
		MaxPathLength: 1,
	}

	intermediateCA, err := NewSimpleCA(intermediateProfile, rootCA)
	require.NoError(t, err)

	// 验证CA链关系
	assert.Equal(t, rootCA.Certificate.Subject, intermediateCA.Certificate.Issuer)
	assert.Equal(t, "根CA", intermediateCA.Certificate.Issuer.CommonName)
	assert.Equal(t, "中间CA", intermediateCA.Certificate.Subject.CommonName)

	// 创建终端实体证书
	leafProfile := &SimpleCAProfile{
		CommonName:    "用户证书",
		Organization:  "律师事务所",
		Country:       "CN",
		KeyAlgorithm:  "RSA",
		KeySize:       2048,
		ValidityDays:  365, // 1年
		IsCA:          false,
		MaxPathLength: 0,
	}

	leafCA, err := NewSimpleCA(leafProfile, intermediateCA)
	require.NoError(t, err)

	// 验证终端实体证书
	assert.Equal(t, intermediateCA.Certificate.Subject, leafCA.Certificate.Issuer)
	assert.Equal(t, "中间CA", leafCA.Certificate.Issuer.CommonName)
	assert.Equal(t, "用户证书", leafCA.Certificate.Subject.CommonName)
	assert.False(t, leafCA.Certificate.IsCA)

	t.Logf("✅ 简化CA链测试通过")
	t.Logf("   - 根CA: %s", rootCA.Certificate.Subject.CommonName)
	t.Logf("   - 中间CA: %s", intermediateCA.Certificate.Subject.CommonName)
	t.Logf("   - 终端实体: %s", leafCA.Certificate.Subject.CommonName)
}

// TestSimpleCAAESandECDSA 测试不同密钥算法
func TestSimpleCAAESandECDSA(t *testing.T) {
	testCases := []struct {
		name         string
		keyAlgorithm string
		keySize      int
	}{
		{"RSA-2048", "RSA", 2048},
		{"RSA-4096", "RSA", 4096},
		{"ECDSA-P256", "ECDSA", 256},
		{"ECDSA-P384", "ECDSA", 384},
		{"ECDSA-P521", "ECDSA", 521},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			profile := &SimpleCAProfile{
				CommonName:    tc.name + " 测试CA",
				Organization:  "测试组织",
				Country:       "CN",
				KeyAlgorithm:  tc.keyAlgorithm,
				KeySize:       tc.keySize,
				ValidityDays:  365,
				IsCA:          true,
				MaxPathLength: 1,
			}

			ca, err := NewSimpleCA(profile, nil)
			require.NoError(t, err)
			require.NotNil(t, ca)

			assert.Equal(t, tc.keyAlgorithm, ca.Profile.KeyAlgorithm)
			assert.Equal(t, tc.keySize, ca.Profile.KeySize)

			// 验证私钥类型
			switch tc.keyAlgorithm {
			case "RSA":
				_, ok := ca.PrivateKey.(*rsa.PrivateKey)
				assert.True(t, ok, "RSA密钥类型验证失败")
			case "ECDSA":
				_, ok := ca.PrivateKey.(*ecdsa.PrivateKey)
				assert.True(t, ok, "ECDSA密钥类型验证失败")
			}

			t.Logf("✅ %s CA创建成功", tc.name)
			t.Logf("   - 证书序列号: %s", ca.SerialNumber.String())
		})
	}
}

// TestSimpleCAValidation 测试简化CA验证
func TestSimpleCAValidation(t *testing.T) {
	// 创建根CA
	rootProfile := &SimpleCAProfile{
		CommonName:    "验证测试根CA",
		Organization:  "律师事务所",
		Country:       "CN",
		KeyAlgorithm:  "RSA",
		KeySize:       2048,
		ValidityDays:  365,
		IsCA:          true,
		MaxPathLength: 2,
	}

	rootCA, err := NewSimpleCA(rootProfile, nil)
	require.NoError(t, err)

	// 创建中间CA
	intermediateProfile := &SimpleCAProfile{
		CommonName:    "验证测试中间CA",
		Organization:  "律师事务所",
		Country:       "CN",
		KeyAlgorithm:  "RSA",
		KeySize:       2048,
		ValidityDays:  365,
		IsCA:          true,
		MaxPathLength: 1,
	}

	intermediateCA, err := NewSimpleCA(intermediateProfile, rootCA)
	require.NoError(t, err)

	// 验证证书链
	roots := x509.NewCertPool()
	roots.AddCert(rootCA.Certificate)

	intermediates := x509.NewCertPool()
	intermediates.AddCert(intermediateCA.Certificate)

	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
	}

	chains, err := intermediateCA.Certificate.Verify(opts)
	require.NoError(t, err)
	assert.Len(t, chains, 1)
	assert.Len(t, chains[0], 2) // 中间CA + 根CA

	t.Logf("✅ 简化CA验证测试通过")
	t.Logf("   - 证书链长度: %d", len(chains[0]))
	for i, cert := range chains[0] {
		t.Logf("   - 证书%d: %s", i+1, cert.Subject.CommonName)
	}
}

// TestSimpleCAPemEncoding 测试PEM编码
func TestSimpleCAPemEncoding(t *testing.T) {
	profile := &SimpleCAProfile{
		CommonName:    "PEM测试CA",
		Organization:  "测试组织",
		Country:       "CN",
		KeyAlgorithm:  "RSA",
		KeySize:       2048,
		ValidityDays:  365,
		IsCA:          true,
		MaxPathLength: 1,
	}

	ca, err := NewSimpleCA(profile, nil)
	require.NoError(t, err)

	// 编码证书为PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: ca.Certificate.Raw,
	})
	assert.NotEmpty(t, certPEM)
	assert.Contains(t, string(certPEM), "BEGIN CERTIFICATE")
	assert.Contains(t, string(certPEM), "END CERTIFICATE")

	// 编码私钥为PEM
	var privateKeyBytes []byte
	switch key := ca.PrivateKey.(type) {
	case *rsa.PrivateKey:
		privateKeyBytes, err = x509.MarshalPKCS8PrivateKey(key)
	case *ecdsa.PrivateKey:
		privateKeyBytes, err = x509.MarshalPKCS8PrivateKey(key)
	default:
		t.Fatalf("未知的私钥类型")
	}
	require.NoError(t, err)

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	assert.NotEmpty(t, privateKeyPEM)
	assert.Contains(t, string(privateKeyPEM), "BEGIN PRIVATE KEY")
	assert.Contains(t, string(privateKeyPEM), "END PRIVATE KEY")

	t.Logf("✅ PEM编码测试通过")
	t.Logf("   - 证书PEM长度: %d bytes", len(certPEM))
	t.Logf("   - 私钥PEM长度: %d bytes", len(privateKeyPEM))
}

// BenchmarkSimpleCACreation 简化CA创建性能测试
func BenchmarkSimpleCACreation(b *testing.B) {
	profile := &SimpleCAProfile{
		CommonName:    "性能测试CA",
		Organization:  "测试组织",
		Country:       "CN",
		KeyAlgorithm:  "RSA",
		KeySize:       2048,
		ValidityDays:  365,
		IsCA:          true,
		MaxPathLength: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewSimpleCA(profile, nil)
		if err != nil {
			b.Fatalf("创建CA失败: %v", err)
		}
	}
}

// BenchmarkSimpleCAChainCreation 简化CA链创建性能测试
func BenchmarkSimpleCAChainCreation(b *testing.B) {
	rootProfile := &SimpleCAProfile{
		CommonName:    "性能测试根CA",
		Organization:  "测试组织",
		Country:       "CN",
		KeyAlgorithm:  "RSA",
		KeySize:       2048,
		ValidityDays:  365,
		IsCA:          true,
		MaxPathLength: 2,
	}

	rootCA, err := NewSimpleCA(rootProfile, nil)
	if err != nil {
		b.Fatalf("创建根CA失败: %v", err)
	}

	intermediateProfile := &SimpleCAProfile{
		CommonName:    "性能测试中间CA",
		Organization:  "测试组织",
		Country:       "CN",
		KeyAlgorithm:  "RSA",
		KeySize:       2048,
		ValidityDays:  365,
		IsCA:          true,
		MaxPathLength: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewSimpleCA(intermediateProfile, rootCA)
		if err != nil {
			b.Fatalf("创建中间CA失败: %v", err)
		}
	}
}