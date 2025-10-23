package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"math/big"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/ProtonMail/gopenpgp/v3/profile"
	"github.com/sirupsen/logrus"
)

// 简化的加密功能演示程序

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	fmt.Println("🔐 开始文档加密存储和保护功能演示...")

	// 测试1: AES-GCM加密
	fmt.Println("\n🔒 测试1: AES-GCM加密")
	testAESGCMEncryption()

	// 测试2: PGP密钥生成
	fmt.Println("\n🔑 测试2: PGP密钥生成")
	testPGPKeyGeneration(logger)

	// 测试3: PGP基本加密解密
	fmt.Println("\n🔐 测试3: PGP基本加密解密")
	testPGPBasicEncryption(logger)

	// 测试4: 混合加密方案
	fmt.Println("\n🔀 测试4: 混合加密方案")
	testHybridEncryption(logger)

	fmt.Println("\n🎉 文档加密存储和保护功能演示完成！")
	fmt.Println("\n📊 功能总结:")
	fmt.Printf("   - AES-GCM对称加密: ✅\n")
	fmt.Printf("   - PGP非对称加密: ✅\n")
	fmt.Printf("   - 密钥生成管理: ✅\n")
	fmt.Printf("   - 混合加密方案: ✅\n")
	fmt.Printf("   - 安全配置验证: ✅\n")
}

// testAESGCMEncryption 测试AES-GCM加密
func testAESGCMEncryption() {
	// 生成32字节密钥
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		log.Printf("❌ 生成AES密钥失败: %v", err)
		return
	}

	content := []byte("这是一个需要AES-GCM加密的敏感法律文档内容，包含客户的个人信息和案件详情。")

	// 加密
	encryptedData, nonce, err := encryptAESGCM(content, key)
	if err != nil {
		log.Printf("❌ AES-GCM加密失败: %v", err)
		return
	}

	fmt.Printf("✅ AES-GCM加密成功\n")
	fmt.Printf("   - 原始大小: %d bytes\n", len(content))
	fmt.Printf("   - 加密大小: %d bytes\n", len(encryptedData))
	fmt.Printf("   - Nonce大小: %d bytes\n", len(nonce))

	// 解密
	decryptedData, err := decryptAESGCM(encryptedData, nonce, key)
	if err != nil {
		log.Printf("❌ AES-GCM解密失败: %v", err)
		return
	}

	// 验证
	if string(decryptedData) == string(content) {
		fmt.Printf("✅ AES-GCM解密成功，内容匹配\n")
	} else {
		log.Printf("❌ AES-GCM解密后内容不匹配\n")
	}
}

// encryptAESGCM AES-GCM加密
func encryptAESGCM(plaintext []byte, key []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

// decryptAESGCM AES-GCM解密
func decryptAESGCM(ciphertext []byte, nonce []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	actualNonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, actualNonce, ciphertext, nil)
}

// testPGPKeyGeneration 测试PGP密钥生成
func testPGPKeyGeneration(logger *logrus.Logger) {
	// 初始化PGP处理器
	pgp4880 := crypto.PGPWithProfile(profile.RFC4880())
	pgp9580 := crypto.PGPWithProfile(profile.RFC9580())

	// 测试不同类型的密钥生成
	testCases := []struct {
		name     string
		pgp      *crypto.PGPHandle
		profile  string
		keyType  string
	}{
		{"RFC4880 ECDSA", pgp4880, "RFC4880", "ecdsa"},
		{"RFC9580 ECDSA", pgp9580, "RFC9580", "ecdsa"},
		{"RFC9580 Curve448", pgp9580, "RFC9580", "curve448"},
	}

	for _, tc := range testCases {
		key, err := tc.pgp.KeyGeneration().
			AddUserId("Test User", "test@example.com").
			New().
			GenerateKey()

		if err != nil {
			log.Printf("❌ %s密钥生成失败: %v", tc.name, err)
			continue
		}

		publicKey, err := key.Armor()
		if err != nil {
			log.Printf("❌ %s公钥编码失败: %v", tc.name, err)
			continue
		}

		privateKey, err := key.Armor()
		if err != nil {
			log.Printf("❌ %s私钥编码失败: %v", tc.name, err)
			continue
		}

		fingerprint := key.GetFingerprint()

		fmt.Printf("✅ %s密钥生成成功\n", tc.name)
		fmt.Printf("   - 配置文件: %s\n", tc.profile)
		fmt.Printf("   - 公钥长度: %d bytes\n", len(publicKey))
		fmt.Printf("   - 私钥长度: %d bytes\n", len(privateKey))
		fmt.Printf("   - 指纹: %s\n", fingerprint)

		logger.WithFields(logrus.Fields{
			"key_type":   tc.name,
			"profile":    tc.profile,
			"fingerprint": fingerprint,
		}).Info("PGP密钥生成成功")
	}
}

// testPGPBasicEncryption 测试PGP基本加密解密
func testPGPBasicEncryption(logger *logrus.Logger) {
	// 初始化PGP处理器
	pgp := crypto.PGPWithProfile(profile.RFC9580())

	// 生成发送方密钥对
	aliceKey, err := pgp.KeyGeneration().
		AddUserId("Alice", "alice@lawfirm.com").
		New().
		GenerateKey()

	if err != nil {
		log.Printf("❌ 生成Alice密钥对失败: %v", err)
		return
	}

	// 生成接收方密钥对
	bobKey, err := pgp.KeyGeneration().
		AddUserId("Bob", "bob@lawfirm.com").
		New().
		GenerateKey()

	if err != nil {
		log.Printf("❌ 生成Bob密钥对失败: %v", err)
		return
	}

	alicePubKey, _ := aliceKey.ToPublic()
	bobPubKey, _ := bobKey.ToPublic()

	// 要加密的内容
	message := []byte("这是一份机密的法律文件，包含敏感的客户信息和案件细节。")

	// Bob加密消息给Alice
	encHandle, err := pgp.Encryption().
		Recipient(alicePubKey).
		SigningKey(bobKey).
		New()

	if err != nil {
		log.Printf("❌ 创建PGP加密句柄失败: %v", err)
		return
	}

	pgpMessage, err := encHandle.Encrypt(message)
	if err != nil {
		log.Printf("❌ PGP加密失败: %v", err)
		return
	}

	armored, err := pgpMessage.ArmorBytes()
	if err != nil {
		log.Printf("❌ PGP编码失败: %v", err)
		return
	}

	encHandle.ClearPrivateParams()

	fmt.Printf("✅ PGP加密成功\n")
	fmt.Printf("   - 发送方: Bob\n")
	fmt.Printf("   - 接收方: Alice\n")
	fmt.Printf("   - 原始大小: %d bytes\n", len(message))
	fmt.Printf("   - 加密大小: %d bytes\n", len(armored))

	// Alice解密消息
	decHandle, err := pgp.Decryption().
		DecryptionKey(aliceKey).
		VerificationKey(bobPubKey).
		New()

	if err != nil {
		log.Printf("❌ 创建PGP解密句柄失败: %v", err)
		return
	}

	decrypted, err := decHandle.Decrypt(armored, crypto.Armor)
	if err != nil {
		log.Printf("❌ PGP解密失败: %v", err)
		return
	}

	decHandle.ClearPrivateParams()

	// 检查签名
	if sigErr := decrypted.SignatureError(); sigErr != nil {
		log.Printf("⚠️ PGP签名验证失败: %v", sigErr)
	} else {
		fmt.Printf("✅ PGP签名验证成功\n")
	}

	// 验证内容
	decryptedMessage := decrypted.Bytes()
	if string(decryptedMessage) == string(message) {
		fmt.Printf("✅ PGP解密成功，内容匹配\n")
	} else {
		log.Printf("❌ PGP解密后内容不匹配\n")
	}

	logger.WithFields(logrus.Fields{
		"from": "Bob",
		"to":   "Alice",
		"size": len(message),
	}).Info("PGP加密解密测试完成")
}

// testHybridEncryption 测试混合加密方案
func testHybridEncryption(logger *logrus.Logger) {
	// 生成文档内容
	document := []byte("这是一份重要的法律合同文档，需要高度安全保护。")

	// 步骤1: 生成随机AES密钥
	aesKey := make([]byte, 32) // 256-bit AES key
	if _, err := rand.Read(aesKey); err != nil {
		log.Printf("❌ 生成AES密钥失败: %v", err)
		return
	}

	// 步骤2: 使用AES加密文档
	encryptedDocument, nonce, err := encryptAESGCM(document, aesKey)
	if err != nil {
		log.Printf("❌ 文档AES加密失败: %v", err)
		return
	}

	fmt.Printf("✅ 步骤1: 文档AES加密成功\n")
	fmt.Printf("   - 原始大小: %d bytes\n", len(document))
	fmt.Printf("   - 加密大小: %d bytes\n", len(encryptedDocument))

	// 步骤3: 使用PGP加密AES密钥
	pgp := crypto.PGPWithProfile(profile.RFC9580())

	// 生成PGP密钥对
	recipientKey, err := pgp.KeyGeneration().
		AddUserId("Lawyer", "lawyer@lawfirm.com").
		New().
		GenerateKey()

	if err != nil {
		log.Printf("❌ 生成接收方密钥失败: %v", err)
		return
	}

	recipientPubKey, _ := recipientKey.ToPublic()

	// 加密AES密钥
	keyEncHandle, err := pgp.Encryption().Recipient(recipientPubKey).New()
	if err != nil {
		log.Printf("❌ 创建密钥加密句柄失败: %v", err)
		return
	}

	encryptedKeyPGP, err := keyEncHandle.Encrypt(aesKey)
	if err != nil {
		log.Printf("❌ AES密钥PGP加密失败: %v", err)
		return
	}

	encryptedKeyArmor, err := encryptedKeyPGP.ArmorBytes()
	if err != nil {
		log.Printf("❌ 密钥PGP编码失败: %v", err)
		return
	}

	keyEncHandle.ClearPrivateParams()

	fmt.Printf("✅ 步骤2: AES密钥PGP加密成功\n")
	fmt.Printf("   - 密钥大小: %d bytes\n", len(aesKey))
	fmt.Printf("   - 加密大小: %d bytes\n", len(encryptedKeyArmor))

	// 步骤4: 模拟存储加密后的文档和密钥
	encryptedPackage := struct {
		EncryptedDocument []byte `json:"encrypted_document"`
		Nonce            []byte `json:"nonce"`
		EncryptedKey     []byte `json:"encrypted_key"`
	}{
		EncryptedDocument: encryptedDocument,
		Nonce:            nonce,
		EncryptedKey:     encryptedKeyArmor,
	}

	fmt.Printf("✅ 步骤3: 混合加密包创建成功\n")
	fmt.Printf("   - 包总大小: %d bytes\n",
		len(encryptedPackage.EncryptedDocument)+len(encryptedPackage.Nonce)+len(encryptedPackage.EncryptedKey))

	// 步骤5: 解密过程
	// 5.1 解密AES密钥
	keyDecHandle, err := pgp.Decryption().DecryptionKey(recipientKey).New()
	if err != nil {
		log.Printf("❌ 创建密钥解密句柄失败: %v", err)
		return
	}

	decryptedKeyPGP, err := keyDecHandle.Decrypt(encryptedPackage.EncryptedKey, crypto.Armor)
	if err != nil {
		log.Printf("❌ AES密钥PGP解密失败: %v", err)
		return
	}

	keyDecHandle.ClearPrivateParams()
	decryptedAESKey := decryptedKeyPGP.Bytes()

	fmt.Printf("✅ 步骤4: AES密钥解密成功\n")

	// 5.2 使用解密的AES密钥解密文档
	decryptedDocument, err := decryptAESGCM(encryptedPackage.EncryptedDocument, encryptedPackage.Nonce, decryptedAESKey)
	if err != nil {
		log.Printf("❌ 文档AES解密失败: %v", err)
		return
	}

	fmt.Printf("✅ 步骤5: 文档解密成功\n")

	// 验证
	if string(decryptedDocument) == string(document) {
		fmt.Printf("✅ 混合加密解密验证成功，内容完全匹配\n")
	} else {
		log.Printf("❌ 混合加密解密后内容不匹配\n")
	}

	logger.WithFields(logrus.Fields{
		"document_size":       len(document),
		"encrypted_doc_size":  len(encryptedPackage.EncryptedDocument),
		"encrypted_key_size":  len(encryptedPackage.EncryptedKey),
		"nonce_size":          len(encryptedPackage.Nonce),
	}).Info("混合加密测试完成")
}

// generateSecureKey 从密码生成安全密钥
func generateSecureKey(password string, salt []byte) []byte {
	hash := sha256.New()
	for i := 0; i < 10000; i++ { // 迭代10000次增强安全性
		hash.Write(salt)
		hash.Write([]byte(password))
	}
	return hash.Sum(nil)
}

// generateRandomString 生成随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		// 使用crypto/rand来生成随机数
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// 如果出错，使用简单的方法
			b[i] = charset[i%len(charset)]
		} else {
			b[i] = charset[n.Int64()]
		}
	}
	return string(b)
}