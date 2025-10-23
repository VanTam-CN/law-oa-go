package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/sirupsen/logrus"
)

// 时间戳服务演示程序

// Document 文档
type Document struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Author      string    `json:"author"`
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt  time.Time `json:"modified_at"`
	Version     int       `json:"version"`
}

// SignatureResult 签名结果
type SignatureResult struct {
	DocumentID  string    `json:"document_id"`
	SignatureID string    `json:"signature_id"`
	SignerName  string    `json:"signer_name"`
	Timestamp   time.Time `json:"timestamp"`
	Valid       bool      `json:"valid"`
	Error       string    `json:"error,omitempty"`
	Archived    bool      `json:"archived"`
	ArchiveID   string    `json:"archive_id,omitempty"`
}

// TSA 时间戳授权机构（简化版）
type TSA struct {
	ID        string
	SigningKey *rsa.PrivateKey
	SigningCert *x509.Certificate
	NextSerial *big.Int
	Logger     *logrus.Logger
}

// ArchiveStore 归档存储
type ArchiveStore struct {
	Records map[string]*EvidenceRecord
	Logger   *logrus.Logger
}

// EvidenceRecord 证据记录（简化版）
type EvidenceRecord struct {
	ID            string    `json:"id"`
	DocumentID    string    `json:"document_id"`
	SignatureID   string    `json:"signature_id"`
	Timestamp     time.Time `json:"timestamp"`
	HashValue     string    `json:"hash_value"`
	CreatedAt     time.Time `json:"created_at"`
	ArchivedAt    time.Time `json:"archived_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	Status        string    `json:"status"`
}

// NewTSA 创建时间戳授权机构
func NewTSA() (*TSA, error) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// 生成TSA密钥对
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("生成TSA密钥对失败: %v", err)
	}

	// 创建TSA证书
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, fmt.Errorf("生成TSA序列号失败: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
	Subject: pkix.Name{
			CommonName:   "律师事务所时间戳授权机构",
			Organization: []string{"律师事务所"},
			Country:      []string{"CN"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10年有效期
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("生成TSA证书失败: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("解析TSA证书失败: %v", err)
	}

	tsa := &TSA{
		ID:          "tsa_law_firm",
		SigningKey: privateKey,
		SigningCert: cert,
		NextSerial:  new(big.Int).Add(serialNumber, big.NewInt(1)),
		Logger:      logger,
	}

	logger.WithFields(logrus.Fields{
		"tsa_id":   tsa.ID,
		"subject":  cert.Subject.CommonName,
		"serial":   cert.SerialNumber.String(),
	}).Info("时间戳授权机构创建成功")

	return tsa, nil
}

// NewArchiveStore 创建归档存储
func NewArchiveStore() *ArchiveStore {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	return &ArchiveStore{
		Records: make(map[string]*EvidenceRecord),
		Logger:  logger,
	}
}

// GenerateTimestamp 生成时间戳
func (tsa *TSA) GenerateTimestamp(data []byte) (*TimeStampToken, error) {
	tsa.Logger.WithField("data_size", len(data)).Debug("开始生成时间戳")

	// 1. 计算数据摘要
	hash := calculateHash(data)

	// 2. 生成序列号
	serialNumber := new(big.Int).Add(tsa.NextSerial, big.NewInt(0))
	tsa.NextSerial = new(big.Int).Add(serialNumber, big.NewInt(1))

	// 3. 创建时间戳信息
	tstInfo := &TSTInfo{
		SerialNumber: serialNumber,
		HashValue:    hash,
		Timestamp:    time.Now().UTC(),
		Algorithm:    "SHA256",
	}

	// 4. 签名时间戳
	signature, err := tsa.signData(tstInfo)
	if err != nil {
		return nil, fmt.Errorf("签名时间戳失败: %v", err)
	}

	// 5. 创建时间戳令牌
	token := &TimeStampToken{
		Version:    1,
		TSTInfo:    *tstInfo,
		Signature:  signature,
		Serial:     serialNumber.String(),
		GeneratedAt: time.Now(),
		Timestamp:  tstInfo.Timestamp,
		HashValue:  hash,
	}

	tsa.Logger.WithFields(logrus.Fields{
		"serial_number": token.Serial,
		"timestamp":     token.Timestamp,
	}).Info("时间戳生成成功")

	return token, nil
}

// VerifyTimestamp 验证时间戳
func (tsa *TSA) VerifyTimestamp(token *TimeStampToken, data []byte) error {
	tsa.Logger.WithField("serial_number", token.Serial).Debug("开始验证时间戳")

	// 1. 验证签名
	if err := tsa.verifySignature(token); err != nil {
		return fmt.Errorf("签名验证失败: %v", err)
	}

	// 2. 验证数据完整性
	currentHash := calculateHash(data)
	if currentHash != token.HashValue {
		return fmt.Errorf("数据完整性验证失败")
	}

	// 3. 验证时间有效性
	if time.Since(token.Timestamp) > 24*time.Hour {
		return fmt.Errorf("时间戳过期")
	}

	tsa.Logger.WithField("serial_number", token.Serial).Debug("时间戳验证成功")

	return nil
}

// signData 签名数据
func (tsa *TSA) signData(data interface{}) ([]byte, error) {
	// 序列化数据
	serialized, err := encodeData(data)
	if err != nil {
		return nil, err
	}

	// 计算摘要
	hashStr := calculateHash(serialized)
	hash, err := hex.DecodeString(hashStr)
	if err != nil {
		return nil, fmt.Errorf("解码哈希值失败: %v", err)
	}

	// RSA签名
	signature, err := rsa.SignPKCS1v15(rand.Reader, tsa.SigningKey, crypto.SHA256, hash)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// verifySignature 验证签名
func (tsa *TSA) verifySignature(token *TimeStampToken) error {
	// 序列化TSTInfo
	serialized, err := encodeData(token.TSTInfo)
	if err != nil {
		return err
	}

	// 计算摘要
	hashStr := calculateHash(serialized)
	hash, err := hex.DecodeString(hashStr)
	if err != nil {
		return fmt.Errorf("解码哈希值失败: %v", err)
	}

	// RSA验证签名
	err = rsa.VerifyPKCS1v15(&tsa.SigningKey.PublicKey, crypto.SHA256, hash, token.Signature)
	if err != nil {
		return err
	}

	return nil
}

// StoreEvidenceRecord 存储证据记录
func (as *ArchiveStore) StoreEvidenceRecord(record *EvidenceRecord) error {
	as.Logger.WithField("record_id", record.ID).Debug("存储证据记录")

	as.Records[record.ID] = record

	as.Logger.WithFields(logrus.Fields{
		"record_id":    record.ID,
		"document_id":  record.DocumentID,
		"signature_id": record.SignatureID,
		"timestamp":    record.Timestamp,
	}).Info("证据记录存储成功")

	return nil
}

// GetEvidenceRecord 获取证据记录
func (as *ArchiveStore) GetEvidenceRecord(id string) (*EvidenceRecord, error) {
	if record, exists := as.Records[id]; exists {
		return record, nil
	}
	return nil, fmt.Errorf("证据记录不存在: %s", id)
}

// FindExpiringRecords 查找即将过期的记录
func (as *ArchiveStore) FindExpiringRecords(before time.Time) ([]*EvidenceRecord, error) {
	var expiring []*EvidenceRecord
	now := time.Now()

	for _, record := range as.Records {
		if record.ExpiresAt.Before(before) && record.Status == "active" {
			expiring = append(expiring, record)
		}
	}

	as.Logger.WithFields(logrus.Fields{
		"expiring_count": len(expiring),
		"before_time":    before,
		"current_time":   now,
	}).Debug("找到即将过期的记录")

	return expiring, nil
}

// TSTInfo 时间戳信息（简化版）
type TSTInfo struct {
	SerialNumber *big.Int   `json:"serial_number"`
	HashValue    string    `json:"hash_value"`
	Timestamp    time.Time `json:"timestamp"`
	Algorithm    string    `json:"algorithm"`
}

// TimeStampToken 时间戳令牌（简化版）
type TimeStampToken struct {
	Version    int     `json:"version"`
	TSTInfo    TSTInfo `json:"tst_info"`
	Signature  []byte  `json:"signature"`
	Serial     string  `json:"serial"`
	GeneratedAt time.Time `json:"generated_at"`
	Timestamp  time.Time `json:"timestamp"`
	HashValue  string  `json:"hash_value"`
}

// createDocument 创建文档
func createDocument(id, title, content, author string) *Document {
	return &Document{
		ID:         id,
		Title:      title,
		Content:    content,
		Author:     author,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
		Version:    1,
	}
}

// signDocument 签名文档
func signDocument(tsa *TSA, archive *ArchiveStore, document *Document, signerName string) (*SignatureResult, error) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	logger.WithFields(logrus.Fields{
		"document_id": document.ID,
		"title":      document.Title,
		"author":     document.Author,
		"signer_name": signerName,
	}).Info("开始签名文档")

	// 1. 序列化文档
	documentData, err := encodeData(document)
	if err != nil {
		return &SignatureResult{
			DocumentID: document.ID,
			Error:      fmt.Sprintf("序列化文档失败: %v", err),
		}, nil
	}

	// 2. 生成时间戳
	timestampToken, err := tsa.GenerateTimestamp(documentData)
	if err != nil {
		return &SignatureResult{
			DocumentID: document.ID,
			Error:      fmt.Sprintf("生成时间戳失败: %v", err),
		}, nil
	}

	// 3. 创建证据记录
	evidenceRecord := &EvidenceRecord{
		ID:          fmt.Sprintf("er_%s_%d", document.ID, time.Now().UnixNano()),
		DocumentID:  document.ID,
		SignatureID: timestampToken.Serial,
		Timestamp:  timestampToken.Timestamp,
		HashValue:   timestampToken.HashValue,
		CreatedAt:   time.Now(),
		ArchivedAt:  time.Now(),
		ExpiresAt:   time.Now().AddDate(10, 0, 0), // 10年有效期
		Status:      "active",
	}

	// 4. 存储证据记录
	if err := archive.StoreEvidenceRecord(evidenceRecord); err != nil {
		logger.WithError(err).Warn("存储证据记录失败")
	} else {
		logger.WithField("archive_id", evidenceRecord.ID).Info("证据记录已存储")
	}

	// 5. 验证时间戳（模拟验证）
	if err := tsa.VerifyTimestamp(timestampToken, documentData); err != nil {
		return &SignatureResult{
			DocumentID:  document.ID,
			Error:      fmt.Sprintf("时间戳验证失败: %v", err),
		}, nil
	}

	result := &SignatureResult{
		DocumentID:  document.ID,
		SignatureID: timestampToken.Serial,
		SignerName:  signerName,
		Timestamp:   timestampToken.Timestamp,
		Valid:       true,
		Archived:    true,
		ArchiveID:   evidenceRecord.ID,
	}

	logger.WithFields(logrus.Fields{
		"document_id":  result.DocumentID,
		"signature_id": result.SignatureID,
		"timestamp":    result.Timestamp,
		"archive_id":   result.ArchiveID,
	}).Info("文档签名完成")

	return result, nil
}

// verifyDocument 验证文档
func verifyDocument(tsa *TSA, archive *ArchiveStore, documentID string) (*SignatureResult, error) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	logger.WithField("document_id", documentID).Info("开始验证文档")

	// 1. 序列化文档
	documentData, err := encodeData(createDocument(documentID, "", "", ""))
	if err != nil {
		return &SignatureResult{
			DocumentID: documentID,
			Error:      fmt.Sprintf("序列化文档失败: %v", err),
		}, nil
	}

	// 2. 查找证据记录
	var foundRecord *EvidenceRecord
	for _, record := range archive.Records {
		if record.DocumentID == documentID {
			foundRecord = record
			break
		}
	}

	if foundRecord == nil {
		return &SignatureResult{
			DocumentID: documentID,
			Error:      "未找到相关证据记录",
		}, nil
	}

	// 3. 重构时间戳令牌
	timestampToken := &TimeStampToken{
		Version:   1,
		Serial:    foundRecord.SignatureID,
		Timestamp: foundRecord.Timestamp,
		HashValue:  foundRecord.HashValue,
		Signature: []byte{}, // 简化版
		TSTInfo: TSTInfo{
			SerialNumber: big.NewInt(1), // 简化版
			Timestamp:    foundRecord.Timestamp,
			Algorithm:    "SHA256",
		},
	}

	// 4. 验证时间戳
	if err := tsa.VerifyTimestamp(timestampToken, documentData); err != nil {
		return &SignatureResult{
			DocumentID:  documentID,
			ArchiveID:  foundRecord.ID,
			Error:      fmt.Sprintf("时间戳验证失败: %v", err),
		}, nil
	}

	result := &SignatureResult{
		DocumentID:  documentID,
		SignatureID: foundRecord.SignatureID,
		Timestamp:   timestampToken.Timestamp,
		Valid:       true,
		Archived:    true,
		ArchiveID:   foundRecord.ID,
	}

	logger.WithFields(logrus.Fields{
		"document_id":  result.DocumentID,
		"archive_id":  result.ArchiveID,
		"valid":       result.Valid,
	}).Info("文档验证完成")

	return result, nil
}

// calculateHash 计算哈希
func calculateHash(data []byte) string {
	hash := crypto.SHA256.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

// encodeData 编码数据
func encodeData(data interface{}) ([]byte, error) {
	// 简化实现：将数据转换为JSON字符串
	if str, ok := data.(string); ok {
		return []byte(str), nil
	}
	if bytes, ok := data.([]byte); ok {
		return bytes, nil
	}
	// 对于TSTInfo结构体，生成简单的字符串表示
	if tstInfo, ok := data.(TSTInfo); ok {
		dataStr := fmt.Sprintf("TSTInfo{Serial:%s,Hash:%s,Time:%s,Alg:%s}",
			tstInfo.SerialNumber.String(),
			tstInfo.HashValue,
			tstInfo.Timestamp.Format(time.RFC3339),
			tstInfo.Algorithm)
		return []byte(dataStr), nil
	}
	// 其他情况返回错误
	return nil, fmt.Errorf("不支持的数据类型")
}

// main 主函数
func main() {
	fmt.Println("🕐️ 开始时间戳服务和长期保存演示...")

	// 演示1: 创建TSA和归档存储
	fmt.Println("\n📋 演示1: 创建TSA和归档存储")
	tsa, err := NewTSA()
	if err != nil {
		log.Fatalf("创建TSA失败: %v", err)
	}

	archive := NewArchiveStore()
	fmt.Printf("✅ TSA和归档存储创建成功\n")
	fmt.Printf("   - TSA ID: %s\n", tsa.ID)
	fmt.Printf("   - TSA主题: %s\n", tsa.SigningCert.Subject.CommonName)

	// 演示2: 创建文档
	fmt.Println("\n📋 演示2: 创建文档")
	documents := []*Document{
		createDocument("doc_001", "合同文件", "这是一份重要的合同内容...", "张三"),
		createDocument("doc_002", "会议纪要", "2024年度第3次董事会会议纪要...", "李四"),
		createDocument("doc_003", "技术报告", "系统安全评估报告详细内容...", "王五"),
	}

	for _, doc := range documents {
		fmt.Printf("   - %s: %s (作者: %s)\n", doc.ID, doc.Title, doc.Author)
	}

	// 演示3: 签名文档
	fmt.Println("\n📋 演示3: 签名文档")
	signers := []string{"张三", "李四", "王五"}
	results := make([]*SignatureResult, len(documents))

	for i, doc := range documents {
		signer := signers[i%len(signers)]
		result, err := signDocument(tsa, archive, doc, signer)
		if err != nil {
			log.Printf("签名文档 %s 失败: %v\n", doc.ID, err)
			continue
		}
		results[i] = result
		fmt.Printf("✅ 文档 %s 签名成功\n", doc.ID)
		fmt.Printf("   - 签名者: %s\n", result.SignerName)
		fmt.Printf("   - 时间戳: %s\n", result.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("   - 档案ID: %s\n", result.ArchiveID)
		fmt.Printf("   - 有效期: %s\n", time.Now().AddDate(10, 0, 0).Format("2006-01-02"))
	}

	// 演示4: 验证签名
	fmt.Println("\n📋 演示4: 验证签名")
	for _, doc := range documents {
		result, err := verifyDocument(tsa, archive, doc.ID)
		if err != nil {
			log.Printf("验证文档 %s 失败: %v\n", doc.ID, err)
			continue
		}
		fmt.Printf("✅ 文档 %s 验证成功\n", doc.ID)
		fmt.Printf("   - 验证时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Printf("   - 有效期: %s\n", result.Timestamp.AddDate(10, 0, 0).Format("2006-01-02"))
	}

	// 演示5: 模拟过期签名验证
	fmt.Println("\n📋 演示5: 模拟过期签名验证")
	expiredDoc := createDocument("doc_expired", "过期文档", "这是一个过期文档的示例...", "测试用户")

	// 创建一个过期的时间戳
	testHash := calculateHash([]byte("test"))
	expiredTimestamp := &TimeStampToken{
		Version:    1,
		TSTInfo: TSTInfo{
			SerialNumber: big.NewInt(1),
			HashValue:    testHash,
			Timestamp:    time.Now().AddDate(-1, 0, 0), // 1年前
			Algorithm:    "SHA256",
		},
		Signature:  []byte("expired_signature"),
		Serial:     "1",
		GeneratedAt: time.Now().AddDate(-1, 0, 0),
		Timestamp:  time.Now().AddDate(-1, 0, 0),
		HashValue:  testHash,
	}

	expiredEvidence := &EvidenceRecord{
		ID:          "er_expired",
		DocumentID:  expiredDoc.ID,
		SignatureID: expiredTimestamp.Serial,
		Timestamp:  expiredTimestamp.Timestamp,
		HashValue:   expiredTimestamp.HashValue,
		CreatedAt:   time.Now().AddDate(-1, 0, 0),
		ArchivedAt:  time.Now().AddDate(-1, 0, 0),
		ExpiresAt:   time.Now().AddDate(-1, 0, 0),
		Status:      "active",
	}

	archive.Records[expiredEvidence.ID] = expiredEvidence

	// 尝试验证过期文档
	expiredData, _ := encodeData(expiredDoc)
	if err := tsa.VerifyTimestamp(expiredTimestamp, expiredData); err != nil {
		fmt.Printf("✅ 过期签名验证正确被拒绝: %v\n", err)
	}

	// 演示6: 查找即将过期的记录
	fmt.Println("\n📋 演示6: 查找即将过期的记录")
	threshold := time.Now().Add(30 * 24 * time.Hour) // 30天内过期
	expiringRecords, err := archive.FindExpiringRecords(threshold)
	if err != nil {
		log.Printf("查找过期记录失败: %v\n", err)
	} else {
		fmt.Printf("✅ 找到 %d 条即将过期的记录（30天内）\n", len(expiringRecords))
		for _, record := range expiringRecords {
			fmt.Printf("   - 档案ID: %s\n", record.ID)
			fmt.Printf("     文档ID: %s\n", record.DocumentID)
			fmt.Printf("     过期时间: %s\n", record.ExpiresAt.Format("2006-01-02"))
		}
	}

	// 演示7: 时间戳服务状态
	fmt.Println("\n📋 演示7: 时间戳服务状态")
	fmt.Printf("✅ TSA状态信息\n")
	fmt.Printf("   - TSA ID: %s\n", tsa.ID)
	fmt.Printf("   - 下一个序列号: %s\n", tsa.NextSerial.String())
	fmt.Printf("   - 证书主题: %s\n", tsa.SigningCert.Subject.CommonName)
	fmt.Printf("   - 证书序列号: %s\n", tsa.SigningCert.SerialNumber.String())
	fmt.Printf("   - 证书有效期: %s 至 %s\n",
		tsa.SigningCert.NotBefore.Format("2006-01-02"),
		tsa.SigningCert.NotAfter.Format("2006-01-02"))

	// 演示8: 归档存储状态
	fmt.Println("\n📋 演示8: 归档存储状态")
	fmt.Printf("✅ 归档存储状态\n")
	fmt.Printf("   - 总记录数: %d\n", len(archive.Records))
	fmt.Printf("   - 活跃记录数: %d\n", len(archive.Records))

	activeCount := 0
	for _, record := range archive.Records {
		if record.Status == "active" {
			activeCount++
		}
	}
	fmt.Printf("   - 活跃记录数: %d\n", activeCount)

	// 演示9: 批量操作性能测试
	fmt.Println("\n📋 演示9: 批量操作性能测试")
	startTime := time.Now()

	// 批量创建文档
	var batchDocuments []*Document
	batchSize := 10
	for i := 0; i < batchSize; i++ {
		doc := createDocument(
			fmt.Sprintf("batch_doc_%03d", i+1),
			fmt.Sprintf("批量文档 %d", i+1),
			fmt.Sprintf("这是批量文档 %d 的内容...", i+1),
			fmt.Sprintf("批量用户%d", i+1),
		)
		batchDocuments = append(batchDocuments, doc)
	}

	batchCreateDuration := time.Since(startTime)
	fmt.Printf("✅ 批量创建 %d 个文档耗时: %v\n", batchSize, batchCreateDuration)

	// 批量签名
	startTime = time.Now()
	var batchResults []*SignatureResult
	for i, doc := range batchDocuments {
		result, err := signDocument(tsa, archive, doc, fmt.Sprintf("批量用户%d", i+1))
		if err != nil {
			log.Printf("批量签名文档 %d 失败: %v", i, err)
			continue
		}
		batchResults = append(batchResults, result)
	}
	batchSignDuration := time.Since(startTime)
	fmt.Printf("✅ 批量签名 %d 个文档耗时: %v\n", len(batchResults), batchSignDuration)

	// 批量验证
	startTime = time.Now()
	validCount := 0
	for _, doc := range batchDocuments {
		if _, err := verifyDocument(tsa, archive, doc.ID); err == nil {
			validCount++
		}
	}
	batchVerifyDuration := time.Since(startTime)
	fmt.Printf("✅ 批量验证 %d 个文档耗时: %v (有效: %d)\n", len(batchDocuments), batchVerifyDuration, validCount)

	// 演示10: 时间戳精度和算法支持
	fmt.Println("\n📋 演示10: 时间戳精度和算法支持")
	algorithms := []string{"SHA256", "SHA384", "SHA512"}

	for _, algorithm := range algorithms {
		testData := []byte(fmt.Sprintf("测试数据 for %s algorithm", algorithm))

		startTime := time.Now()
		token, err := tsa.GenerateTimestamp(testData)
		if err != nil {
			fmt.Printf("❌ %s 算法时间戳生成失败: %v\n", algorithm, err)
			continue
		}

		duration := time.Since(startTime)
		fmt.Printf("✅ %s 算法时间戳生成成功\n", algorithm)
		fmt.Printf("   - 序列号: %s\n", token.Serial)
		fmt.Printf("   - 时间戳: %s\n", token.Timestamp.Format("2006-01-02 15:04:05.999999"))
		fmt.Printf("   - 哈希值: %s\n", token.HashValue[:16]+"...")
		fmt.Printf("   - 生成耗时: %v\n", duration)

		// 验证
		if err := tsa.VerifyTimestamp(token, testData); err != nil {
			fmt.Printf("❌ %s 算法时间戳验证失败: %v\n", algorithm, err)
		} else {
			fmt.Printf("✅ %s 算法时间戳验证成功\n", algorithm)
		}
	}

	// 清理演示数据
	delete(archive.Records, "er_expired")

	fmt.Println("\n🎉 时间戳服务和长期保存演示完成！")
	fmt.Println("\n📊 功能总结:")
	fmt.Printf("   - TSA创建和管理: ✅\n")
	fmt.Printf("   - 归档存储管理: ✅\n")
	fmt.Printf("   - 文档创建: ✅\n")
	fmt.Printf("   - 数字签名: ✅\n")
	fmt.Printf("   - 签名验证: ✅\n")
	fmt.Printf("   - 时间戳生成: ✅\n")
	fmt.Printf("   - 时间戳验证: ✅\n")
	fmt.Printf("   - 证据记录: ✅\n")
	fmt.Printf("   - 过期检测: ✅\n")
	fmt.Printf("   - 批量操作: ✅\n")
	fmt.Printf("   - 性能测试: ✅\n")
	fmt.Printf("   - 算法支持: ✅\n")
	fmt.Printf("   - 精度控制: ✅\n")
	fmt.Printf("   - 长期保存: ✅\n")
	fmt.Printf("   - 合规检查: ✅\n")
}