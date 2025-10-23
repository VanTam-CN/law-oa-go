package ca

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/sirupsen/logrus"
)

// CAType 证书颁发机构类型
type CAType string

const (
	CATypeRoot     CAType = "root"     // 根CA
	CATypeIntermediate CAType = "intermediate" // 中间CA
	CATypeLeaf     CAType = "leaf"     // 终端实体证书
)

// CAStatus CA状态
type CAStatus string

const (
	CAStatusActive    CAStatus = "active"     // 激活状态
	CAStatusSuspended CAStatus = "suspended"  // 暂停状态
	CAStatusRevoked   CAStatus = "revoked"    // 吊销状态
	CAStatusExpired   CAStatus = "expired"    // 过期状态
)

// CAProfile CA配置文件
type CAProfile struct {
	Name           string    `json:"name"`
	Type           CAType    `json:"type"`
	CommonName     string    `json:"common_name"`
	Organization   string    `json:"organization"`
	OrganizationalUnit string `json:"organizational_unit"`
	Country        string    `json:"country"`
	Province       string    `json:"province"`
	Locality       string    `json:"locality"`
	Email          string    `json:"email"`
	KeyAlgorithm   string    `json:"key_algorithm"`
	KeySize        int       `json:"key_size"`
	HashAlgorithm  string    `json:"hash_algorithm"`
	ValidityPeriod int       `json:"validity_period"` // 有效期(天)
	MaxPathLength  int       `json:"max_path_length"` // 最大路径长度
	KeyUsage       x509.KeyUsage `json:"key_usage"`
	ExtKeyUsage    []x509.ExtKeyUsage `json:"ext_key_usage"`
}

// CertificateAuthority 证书颁发机构
type CertificateAuthority struct {
	ID              string                 `json:"id"`
	Profile         *CAProfile             `json:"profile"`
	Certificate     *x509.Certificate      `json:"-"`
	PrivateKey      crypto.PrivateKey      `json:"-"`
	CertificatePEM  string                 `json:"certificate_pem"`
	PrivateKeyPEM   string                 `json:"private_key_pem"`
	PublicKeyPEM    string                 `json:"public_key_pem"`
	ParentCA        *CertificateAuthority  `json:"-"` // 父CA
	Children        []*CertificateAuthority `json:"-"` // 子CA列表
	Status          CAStatus               `json:"status"`
	SerialNumber    *big.Int               `json:"serial_number"`
	NextSerial      *big.Int               `json:"next_serial"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	ExpiresAt       time.Time              `json:"expires_at"`
	LastUsedAt      time.Time              `json:"last_used_at"`
	UsageCount      int64                  `json:"usage_count"`
	CRLDistributionPoint string           `json:"crl_distribution_point"`
	OCSPServer      string                 `json:"ocsp_server"`
	PolicyOIDs      []string               `json:"policy_oids"`
}

// TrustStore 信任存储
type TrustStore struct {
	ID               string                   `json:"id"`
	Name             string                   `json:"name"`
	RootCertificates  []*x509.Certificate     `json:"-"`
	RootCertsPEM     []string                 `json:"root_certs_pem"`
	IntermediateCerts []*x509.Certificate     `json:"-"`
	IntermediateCertsPEM []string             `json:"intermediate_certs_pem"`
	TrustedAnchors   map[string]*x509.Certificate `json:"-"`
	LastUpdated      time.Time                `json:"last_updated"`
	AutoUpdate       bool                     `json:"auto_update"`
	UpdateInterval   time.Duration            `json:"update_interval"`
	NextUpdate       time.Time                `json:"next_update"`
}

// ChainValidationResult 证书链验证结果
type ChainValidationResult struct {
	IsValid         bool                   `json:"is_valid"`
	Chain           []*x509.Certificate    `json:"chain"`
	TrustAnchor     *x509.Certificate      `json:"trust_anchor"`
	Errors          []string               `json:"errors"`
	Warnings        []string               `json:"warnings"`
	ValidationTime  time.Time              `json:"validation_time"`
	PolicyOID       string                 `json:"policy_oid"`
	RevocationCheck bool                   `json:"revocation_check"`
	OCSPResponse    string                 `json:"ocsp_response"`
	CRLChecked      bool                   `json:"crl_checked"`
}

// CAConfiguration CA配置
type CAConfiguration struct {
	OrganizationName     string            `json:"organization_name"`
	RootCAValidity       int               `json:"root_ca_validity"`       // 根CA有效期(年)
	IntermediateCAValidity int            `json:"intermediate_ca_validity"` // 中间CA有效期(年)
	DefaultKeySize       int               `json:"default_key_size"`
	DefaultHashAlgorithm string            `json:"default_hash_algorithm"`
	EnableOCSP          bool              `json:"enable_ocsp"`
	EnableCRL           bool              `json:"enable_crl"`
	CRLUpdateInterval   time.Duration     `json:"crl_update_interval"`
	OCSPCacheInterval   time.Duration     `json:"ocsp_cache_interval"`
	HSMEnabled          bool              `json:"hsm_enabled"`
	HSMConfig           HSMConfiguration  `json:"hsm_config"`
	BackupEnabled       bool              `json:"backup_enabled"`
	BackupLocation      string            `json:"backup_location"`
	BackupInterval      time.Duration     `json:"backup_interval"`
}

// HSMConfiguration HSM配置
type HSMConfiguration struct {
	Provider       string `json:"provider"`
	ModulePath     string `json:"module_path"`
	TokenLabel     string `json:"token_label"`
	UserPIN        string `json:"user_pin"`
	SO_PIN         string `json:"so_pin"`
	SlotNumber     int    `json:"slot_number"`
	Enabled        bool   `json:"enabled"`
}

// DefaultCAConfiguration 默认CA配置
func DefaultCAConfiguration() *CAConfiguration {
	return &CAConfiguration{
		OrganizationName:       "律师事务所",
		RootCAValidity:         20,                    // 根CA有效期20年
		IntermediateCAValidity: 10,                    // 中间CA有效期10年
		DefaultKeySize:         4096,                  // RSA 4096位
		DefaultHashAlgorithm:   "SHA256",
		EnableOCSP:            true,
		EnableCRL:             true,
		CRLUpdateInterval:     24 * time.Hour,
		OCSPCacheInterval:     1 * time.Hour,
		HSMEnabled:            false,
		BackupEnabled:         true,
		BackupLocation:        "/etc/law-oa/ca/backup",
		BackupInterval:        24 * time.Hour,
	}
}

// DefaultRootCAProfile 默认根CA配置文件
func DefaultRootCAProfile() *CAProfile {
	return &CAProfile{
		Name:                "Root Certificate Authority",
		Type:                CATypeRoot,
		CommonName:          "Law Office Root CA",
		Organization:        "律师事务所",
		OrganizationalUnit:  "Certificate Authority",
		Country:             "CN",
		Province:            "Beijing",
		Locality:            "Beijing",
		KeyAlgorithm:        "RSA",
		KeySize:             4096,
		HashAlgorithm:       "SHA256",
		ValidityPeriod:      20 * 365, // 20年
		MaxPathLength:       3,
		KeyUsage:            x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:         []x509.ExtKeyUsage{},
	}
}

// DefaultIntermediateCAProfile 默认中间CA配置文件
func DefaultIntermediateCAProfile(name string) *CAProfile {
	return &CAProfile{
		Name:                fmt.Sprintf("Intermediate Certificate Authority - %s", name),
		Type:                CATypeIntermediate,
		CommonName:          fmt.Sprintf("Law Office Intermediate CA - %s", name),
		Organization:        "律师事务所",
		OrganizationalUnit:  fmt.Sprintf("Intermediate CA - %s", name),
		Country:             "CN",
		Province:            "Beijing",
		Locality:            "Beijing",
		KeyAlgorithm:        "RSA",
		KeySize:             4096,
		HashAlgorithm:       "SHA256",
		ValidityPeriod:      10 * 365, // 10年
		MaxPathLength:       2,
		KeyUsage:            x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:         []x509.ExtKeyUsage{},
	}
}

// CAManager CA管理器
type CAManager struct {
	config      *CAConfiguration
	trustStore  *TrustStore
	rootCA      *CertificateAuthority
	logger      *logrus.Logger
	initialized bool
}

// NewCAManager 创建CA管理器
func NewCAManager(config *CAConfiguration, logger *logrus.Logger) *CAManager {
	if config == nil {
		config = DefaultCAConfiguration()
	}

	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	return &CAManager{
		config:     config,
		trustStore: NewTrustStore("Law Office Trust Store"),
		logger:     logger,
	}
}

// Initialize 初始化CA管理器
func (cm *CAManager) Initialize() error {
	cm.logger.Info("初始化CA管理器...")

	// 初始化信任存储
	if err := cm.initializeTrustStore(); err != nil {
		return fmt.Errorf("初始化信任存储失败: %w", err)
	}

	// 初始化根CA
	if err := cm.initializeRootCA(); err != nil {
		return fmt.Errorf("初始化根CA失败: %w", err)
	}

	cm.initialized = true
	cm.logger.Info("CA管理器初始化完成")
	return nil
}

// initializeTrustStore 初始化信任存储
func (cm *CAManager) initializeTrustStore() error {
	if cm.trustStore == nil {
		cm.trustStore = NewTrustStore(fmt.Sprintf("%s Trust Store", cm.config.OrganizationName))
	}

	cm.trustStore.LastUpdated = time.Now()
	cm.trustStore.AutoUpdate = true
	cm.trustStore.UpdateInterval = 24 * time.Hour
	cm.trustStore.NextUpdate = time.Now().Add(cm.trustStore.UpdateInterval)

	cm.logger.WithFields(logrus.Fields{
		"trust_store": cm.trustStore.Name,
		"auto_update": cm.trustStore.AutoUpdate,
	}).Info("信任存储初始化完成")

	return nil
}

// initializeRootCA 初始化根CA
func (cm *CAManager) initializeRootCA() error {
	// 检查是否已存在根CA
	if cm.rootCA != nil && cm.rootCA.Status == CAStatusActive {
		return nil
	}

	// 生成新的根CA
	profile := DefaultRootCAProfile()
	profile.Organization = cm.config.OrganizationName

	ca, err := cm.createCertificateAuthority(profile, nil)
	if err != nil {
		return fmt.Errorf("创建根CA失败: %w", err)
	}

	ca.Status = CAStatusActive
	cm.rootCA = ca

	// 添加到信任存储
	cm.trustStore.AddRootCertificate(ca.Certificate)

	cm.logger.WithFields(logrus.Fields{
		"ca_id":       ca.ID,
		"common_name": ca.Profile.CommonName,
		"serial":      ca.SerialNumber.String(),
		"expires_at":  ca.ExpiresAt.Format("2006-01-02"),
	}).Info("根CA初始化完成")

	return nil
}

// createCertificateAuthority 创建证书颁发机构
func (cm *CAManager) createCertificateAuthority(profile *CAProfile, parentCA *CertificateAuthority) (*CertificateAuthority, error) {
	// 生成密钥对
	privateKey, publicKey, err := cm.generateKeyPair(profile.KeyAlgorithm, profile.KeySize)
	if err != nil {
		return nil, fmt.Errorf("生成密钥对失败: %w", err)
	}

	// 生成序列号
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("生成序列号失败: %w", err)
	}

	// 创建证书模板
	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         profile.CommonName,
			Organization:       []string{profile.Organization},
			OrganizationalUnit: []string{profile.OrganizationalUnit},
			Country:           []string{profile.Country},
			Province:          []string{profile.Province},
			Locality:          []string{profile.Locality},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, profile.ValidityPeriod),
		KeyUsage:              profile.KeyUsage,
		ExtKeyUsage:           profile.ExtKeyUsage,
		BasicConstraintsValid: true,
		IsCA:                  profile.Type != CATypeLeaf,
		MaxPathLen:            profile.MaxPathLength,
	}

	// 如果是中间CA，设置颁发者
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
		return nil, fmt.Errorf("创建证书失败: %w", err)
	}

	// 解析证书
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %w", err)
	}

	// 编码PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	privateKeyPEM, publicKeyPEM, err := cm.encodeKeyPair(privateKey, publicKey)
	if err != nil {
		return nil, fmt.Errorf("编码密钥对失败: %w", err)
	}

	// 创建CA实例
	ca := &CertificateAuthority{
		ID:                 cm.generateCAID(),
		Profile:            profile,
		Certificate:        cert,
		PrivateKey:         privateKey,
		CertificatePEM:     string(certPEM),
		PrivateKeyPEM:      string(privateKeyPEM),
		PublicKeyPEM:       string(publicKeyPEM),
		ParentCA:           parentCA,
		Children:           make([]*CertificateAuthority, 0),
		Status:             CAStatusActive,
		SerialNumber:       serialNumber,
		NextSerial:         new(big.Int).Add(serialNumber, big.NewInt(1)),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		ExpiresAt:          cert.NotAfter,
		UsageCount:         0,
	}

	return ca, nil
}

// generateKeyPair 生成密钥对
func (cm *CAManager) generateKeyPair(algorithm string, keySize int) (crypto.PrivateKey, crypto.PublicKey, error) {
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

	case "Ed25519":
		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return nil, nil, err
		}
		return privateKey, publicKey, nil

	default:
		return nil, nil, fmt.Errorf("不支持的密钥算法: %s", algorithm)
	}
}

// encodeKeyPair 编码密钥对
func (cm *CAManager) encodeKeyPair(privateKey crypto.PrivateKey, publicKey crypto.PublicKey) ([]byte, []byte, error) {
	// 编码私钥
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("编码私钥失败: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	// 编码公钥
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("编码公钥失败: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return privateKeyPEM, publicKeyPEM, nil
}

// generateCAID 生成CA ID
func (cm *CAManager) generateCAID() string {
	return fmt.Sprintf("ca_%d", time.Now().UnixNano())
}

// NewTrustStore 创建信任存储
func NewTrustStore(name string) *TrustStore {
	return &TrustStore{
		ID:               fmt.Sprintf("trust_%d", time.Now().UnixNano()),
		Name:             name,
		RootCertificates:  make([]*x509.Certificate, 0),
		RootCertsPEM:     make([]string, 0),
		IntermediateCerts: make([]*x509.Certificate, 0),
		IntermediateCertsPEM: make([]string, 0),
		TrustedAnchors:   make(map[string]*x509.Certificate),
		LastUpdated:      time.Now(),
		AutoUpdate:       true,
		UpdateInterval:   24 * time.Hour,
		NextUpdate:       time.Now().Add(24 * time.Hour),
	}
}

// AddRootCertificate 添加根证书到信任存储
func (ts *TrustStore) AddRootCertificate(cert *x509.Certificate) {
	ts.RootCertificates = append(ts.RootCertificates, cert)

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
	ts.RootCertsPEM = append(ts.RootCertsPEM, string(certPEM))

	ts.TrustedAnchors[cert.SerialNumber.String()] = cert
	ts.LastUpdated = time.Now()
}

// AddIntermediateCertificate 添加中间证书到信任存储
func (ts *TrustStore) AddIntermediateCertificate(cert *x509.Certificate) {
	ts.IntermediateCerts = append(ts.IntermediateCerts, cert)

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
	ts.IntermediateCertsPEM = append(ts.IntermediateCertsPEM, string(certPEM))

	ts.LastUpdated = time.Now()
}

// GetRootPool 获取根证书池
func (ts *TrustStore) GetRootPool() *x509.CertPool {
	pool := x509.NewCertPool()
	for _, cert := range ts.RootCertificates {
		pool.AddCert(cert)
	}
	return pool
}

// GetIntermediatePool 获取中间证书池
func (ts *TrustStore) GetIntermediatePool() *x509.CertPool {
	pool := x509.NewCertPool()
	for _, cert := range ts.IntermediateCerts {
		pool.AddCert(cert)
	}
	return pool
}

// ValidateCertificateChain 验证证书链
func (ts *TrustStore) ValidateCertificateChain(cert *x509.Certificate) (*ChainValidationResult, error) {
	result := &ChainValidationResult{
		IsValid:        false,
		Chain:          make([]*x509.Certificate, 0),
		Errors:         make([]string, 0),
		Warnings:       make([]string, 0),
		ValidationTime: time.Now(),
		RevocationCheck: false,
		CRLChecked:     false,
	}

	// 设置验证选项
	opts := x509.VerifyOptions{
		Roots:         ts.GetRootPool(),
		Intermediates: ts.GetIntermediatePool(),
		CurrentTime:   result.ValidationTime,
	}

	// 验证证书链
	chains, err := cert.Verify(opts)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("证书链验证失败: %v", err))
		return result, nil
	}

	if len(chains) == 0 {
		result.Errors = append(result.Errors, "没有找到有效的证书链")
		return result, nil
	}

	// 使用第一个有效链
	result.Chain = chains[0]
	result.IsValid = true
	result.TrustAnchor = ts.findTrustAnchor(result.Chain)

	return result, nil
}

// findTrustAnchor 查找信任锚
func (ts *TrustStore) findTrustAnchor(chain []*x509.Certificate) *x509.Certificate {
	for _, cert := range chain {
		if ts.TrustedAnchors[cert.SerialNumber.String()] != nil {
			return cert
		}
	}
	return nil
}

// GetRootCA 获取根CA
func (cm *CAManager) GetRootCA() *CertificateAuthority {
	return cm.rootCA
}

// GetTrustStore 获取信任存储
func (cm *CAManager) GetTrustStore() *TrustStore {
	return cm.trustStore
}

// IsInitialized 检查是否已初始化
func (cm *CAManager) IsInitialized() bool {
	return cm.initialized
}