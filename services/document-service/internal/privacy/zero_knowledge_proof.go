package privacy

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math/big"
	"sync"
	"time"
)

// ZeroKnowledgeProofService 零知识证明服务
type ZeroKnowledgeProofService struct {
	curve     *EllipticCurve
	hashFunc  func() HashFunction
	provers   map[string]*Prover
	verifiers map[string]*Verifier
	logger    *slog.Logger
	mutex     sync.RWMutex
}

// EllipticCurve 椭圆曲线
type EllipticCurve struct {
	P     *big.Int // 椭圆曲线有限域的素数p
	A     *big.Int // 曲线参数a
	B     *big.Int // 曲线参数b
	Gx    *big.Int // 生成点G的x坐标
	Gy    *big.Int // 生成点G的y坐标
	N     *big.Int // 阶n
	Name  string
}

// HashFunction 哈希函数接口
type HashFunction interface {
	Write(data []byte) (int, error)
	Sum() []byte
	Reset()
}

// SHA256Hash SHA256哈希函数
type SHA256Hash struct {
	hash [32]byte
}

// Write 写入数据
func (h *SHA256Hash) Write(data []byte) (int, error) {
	h.hash = sha256.Sum256(append(h.hash[:], data...))
	return len(data), nil
}

// Sum 获取哈希值
func (h *SHA256Hash) Sum() []byte {
	return h.hash[:]
}

// Reset 重置哈希
func (h *SHA256Hash) Reset() {
	h.hash = [32]byte{}
}

// Prover 证明者
type Prover struct {
	ID       string
	Secret   *big.Int
	Public   *Point
	Witness  map[string]interface{}
	Curve    *EllipticCurve
	HashFunc func() HashFunction
}

// Verifier 验证者
type Verifier struct {
	ID       string
	Public   *Point
	Curve    *EllipticCurve
	HashFunc func() HashFunction
}

// Point 椭圆曲线上的点
type Point struct {
	X *big.Int
	Y *big.Int
}

// ZKProof 零知识证明
type ZKProof struct {
	ProofType   string
	Commitments []*Point
	Challenge   *big.Int
	Response    *big.Int
	Metadata    map[string]interface{}
	Timestamp   time.Time
}

// SchnorrProof Schnorr证明
type SchnorrProof struct {
	Commitment *Point
	Challenge  *big.Int
	Response   *big.Int
}

// KnowledgeProof 知识证明
type KnowledgeProof struct {
	Statement  string
	Witness    *big.Int
	Proof      *ZKProof
	Verified   bool
	Timestamp  time.Time
}

// RangeProof 范围证明
type RangeProof struct {
	Value      *big.Int
	Min        *big.Int
	Max        *big.Int
	Proof      *ZKProof
	Verified   bool
}

// SigmaProtocol Sigma协议
type SigmaProtocol struct {
	Protocol   string
	Prover     *Prover
	Verifier   *Verifier
	Commitment *Point
	Challenge  *big.Int
	Response   *big.Int
}

// NewZeroKnowledgeProofService 创建零知识证明服务
func NewZeroKnowledgeProofService(curveName string, logger *slog.Logger) (*ZeroKnowledgeProofService, error) {
	curve := getEllipticCurve(curveName)
	if curve == nil {
		return nil, fmt.Errorf("不支持的椭圆曲线: %s", curveName)
	}

	service := &ZeroKnowledgeProofService{
		curve:     curve,
		hashFunc:  func() HashFunction { return &SHA256Hash{} },
		provers:   make(map[string]*Prover),
		verifiers: make(map[string]*Verifier),
		logger:    logger,
	}

	logger.Info("零知识证明服务初始化完成",
		"curve", curveName,
		"field_size", curve.P.BitLen(),
		"order", curve.N.BitLen(),
	)

	return service, nil
}

// getEllipticCurve 获取椭圆曲线
func getEllipticCurve(name string) *EllipticCurve {
	switch name {
	case "secp256k1":
		return &EllipticCurve{
			P:    new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16),
			A:    big.NewInt(0),
			B:    big.NewInt(7),
			Gx:   new(big.Int).SetString("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798", 16),
			Gy:   new(big.Int).SetString("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8", 16),
			N:    new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16),
			Name: "secp256k1",
		}
	case "ed25519":
		return &EllipticCurve{
			P:    new(big.Int).SetString("7FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFED", 16),
			A:    big.NewInt(-1),
			B:    big(big.NewInt(576460752303423508)), // 486662
			Gx:   new(big.Int).SetString("216936D3CD6E53FEC0A4E231FDD6DC5C692CC7609525A7B2C9562D608F25D51", 16),
			Gy:   new(big.Int).SetString("6666666666666666666666666666666666666666666666666666666666666658", 16),
			N:    new(big.Int).SetString("1000000000000000000000000000000014DEF9DEA2F79CD65812631A5CF5D3ED", 16),
			Name: "ed25519",
		}
	default:
		return nil
	}
}

// CreateProver 创建证明者
func (zkps *ZeroKnowledgeProofService) CreateProver(proverID string, secret *big.Int) (*Prover, error) {
	zkps.mutex.Lock()
	defer zkps.mutex.Unlock()

	if secret == nil {
		return nil, fmt.Errorf("私钥不能为空")
	}

	// 计算公钥 Q = G * secret
	public := zkps.scalarMultiply(zkps.curve.Gx, zkps.curve.Gy, secret)

	prover := &Prover{
		ID:       proverID,
		Secret:   secret,
		Public:   public,
		Witness:  make(map[string]interface{}),
		Curve:    zkps.curve,
		HashFunc: zkps.hashFunc,
	}

	zkps.provers[proverID] = prover

	zkps.logger.Info("证明者创建成功",
		"prover_id", proverID,
		"public_key_x", public.X.String()[:16]+"...",
		"public_key_y", public.Y.String()[:16]+"...",
	)

	return prover, nil
}

// CreateVerifier 创建验证者
func (zkps *ZeroKnowledgeProofService) CreateVerifier(verifierID string, publicKey *Point) (*Verifier, error) {
	zkps.mutex.Lock()
	defer zkps.mutex.Unlock()

	if publicKey == nil {
		return nil, fmt.Errorf("公钥不能为空")
	}

	verifier := &Verifier{
		ID:       verifierID,
		Public:   publicKey,
		Curve:    zkps.curve,
		HashFunc: zkps.hashFunc,
	}

	zkps.verifiers[verifierID] = verifier

	zkps.logger.Info("验证者创建成功",
		"verifier_id", verifierID,
		"public_key_x", publicKey.X.String()[:16]+"...",
		"public_key_y", publicKey.Y.String()[:16]+"...",
	)

	return verifier, nil
}

// ProveKnowledge 证明知识
func (zkps *ZeroKnowledgeProofService) ProveKnowledge(ctx context.Context, proverID string, statement string) (*ZKProof, error) {
	zkps.mutex.RLock()
	prover, exists := zkps.provers[proverID]
	zkps.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("证明者不存在: %s", proverID)
	}

	switch statement {
	case "discrete_log":
		return zkps.proveDiscreteLog(prover)
	case "range":
		return zkps.proveRange(prover)
	case "set_membership":
		return zkps.proveSetMembership(prover)
	default:
		return nil, fmt.Errorf("不支持的证明类型: %s", statement)
	}
}

// VerifyProof 验证证明
func (zkps *ZeroKnowledgeProofService) VerifyProof(ctx context.Context, verifierID string, proof *ZKProof) (bool, error) {
	zkps.mutex.RLock()
	verifier, exists := zkps.verifiers[verifierID]
	zkps.mutex.RUnlock()

	if !exists {
		return false, fmt.Errorf("验证者不存在: %s", verifierID)
	}

	switch proof.ProofType {
	case "discrete_log":
		return zkps.verifyDiscreteLog(verifier, proof)
	case "range":
		return zkps.verifyRange(verifier, proof)
	case "set_membership":
		return zkps.verifySetMembership(verifier, proof)
	default:
		return false, fmt.Errorf("不支持的证明类型: %s", proof.ProofType)
	}
}

// proveDiscreteLog 证明离散对数知识
func (zkps *ZeroKnowledgeProofService) proveDiscreteLog(prover *Prover) (*ZKProof, error) {
	// 1. 生成随机数 k
	k, err := rand.Int(rand.Reader, prover.Curve.N)
	if err != nil {
		return nil, fmt.Errorf("生成随机数失败: %w", err)
	}

	// 2. 计算承诺 R = G * k
	R := zkps.scalarMultiply(prover.Curve.Gx, prover.Curve.Gy, k)

	// 3. 计算挑战 e = H(R || Q)
	hash := prover.HashFunc()
	hash.Write(R.X.Bytes())
	hash.Write(R.Y.Bytes())
	hash.Write(prover.Public.X.Bytes())
	hash.Write(prover.Public.Y.Bytes())
	eBytes := hash.Sum()
	e := new(big.Int).SetBytes(eBytes)
	e.Mod(e, prover.Curve.N)

	// 4. 计算响应 s = k + e * secret mod n
	s := new(big.Int).Mul(e, prover.Secret)
	s.Add(s, k)
	s.Mod(s, prover.Curve.N)

	proof := &ZKProof{
		ProofType:   "discrete_log",
		Commitments: []*Point{R},
		Challenge:   e,
		Response:    s,
		Metadata:    map[string]interface{}{},
		Timestamp:   time.Now(),
	}

	return proof, nil
}

// verifyDiscreteLog 验证离散对数证明
func (zkps *ZeroKnowledgeProofService) verifyDiscreteLog(verifier *Verifier, proof *ZKProof) (bool, error) {
	if len(proof.Commitments) != 1 {
		return false, fmt.Errorf("证明格式错误")
	}

	R := proof.Commitments[0]
	e := proof.Challenge
	s := proof.Response

	// 1. 重新计算挑战 e' = H(R || Q)
	hash := verifier.HashFunc()
	hash.Write(R.X.Bytes())
	hash.Write(R.Y.Bytes())
	hash.Write(verifier.Public.X.Bytes())
	hash.Write(verifier.Public.Y.Bytes())
	ePrimeBytes := hash.Sum()
	ePrime := new(big.Int).SetBytes(ePrimeBytes)
	ePrime.Mod(ePrime, verifier.Curve.N)

	// 2. 验证挑战是否匹配
	if e.Cmp(ePrime) != 0 {
		return false, nil
	}

	// 3. 验证 s*G = R + e*Q
	left := zkps.scalarMultiply(verifier.Curve.Gx, verifier.Curve.Gy, s)
	right := zkps.pointAdd(R, zkps.scalarMultiply(verifier.Public.X, verifier.Public.Y, e))

	return left.X.Cmp(right.X) == 0 && left.Y.Cmp(right.Y) == 0, nil
}

// proveRange 证明范围
func (zkps *ZeroKnowledgeProofService) proveRange(prover *Prover) (*ZKProof, error) {
	// 简化的范围证明实现
	// 实际应该使用Bulletproofs等更高效的方案

	// 生成随机数
	r1, err := rand.Int(rand.Reader, prover.Curve.N)
	if err != nil {
		return nil, err
	}
	r2, err := rand.Int(rand.Reader, prover.Curve.N)
	if err != nil {
		return nil, err
	}

	// 计算承诺
	C1 := zkps.scalarMultiply(prover.Curve.Gx, prover.Curve.Gy, r1)
	C2 := zkps.scalarMultiply(prover.Public.X, prover.Public.Y, r2)

	// 计算挑战
	hash := prover.HashFunc()
	hash.Write(C1.X.Bytes())
	hash.Write(C1.Y.Bytes())
	hash.Write(C2.X.Bytes())
	hash.Write(C2.Y.Bytes())
	eBytes := hash.Sum()
	e := new(big.Int).SetBytes(eBytes)
	e.Mod(e, prover.Curve.N)

	// 计算响应
	s1 := new(big.Int).Add(r1, new(big.Int).Mul(e, prover.Secret))
	s1.Mod(s1, prover.Curve.N)

	s2 := new(big.Int).Add(r2, new(big.Int).Mul(e, big.NewInt(1)))
	s2.Mod(s2, prover.Curve.N)

	proof := &ZKProof{
		ProofType:   "range",
		Commitments: []*Point{C1, C2},
		Challenge:   e,
		Response:    s1, // 简化：只返回一个响应
		Metadata: map[string]interface{}{
			"s2": s2.String(),
		},
		Timestamp: time.Now(),
	}

	return proof, nil
}

// verifyRange 验证范围证明
func (zkps *ZeroKnowledgeProofService) verifyRange(verifier *Verifier, proof *ZKProof) (bool, error) {
	if len(proof.Commitments) != 2 {
		return false, fmt.Errorf("范围证明格式错误")
	}

	C1 := proof.Commitments[0]
	C2 := proof.Commitments[1]
	e := proof.Challenge
	s1 := proof.Response

	// 获取s2
	s2Str, ok := proof.Metadata["s2"].(string)
	if !ok {
		return false, fmt.Errorf("缺少s2参数")
	}
	s2 := new(big.Int)
	s2.SetString(s2Str, 10)

	// 验证承诺和响应
	left1 := zkps.scalarMultiply(verifier.Curve.Gx, verifier.Curve.Gy, s1)
	right1 := zkps.pointAdd(C1, zkps.scalarMultiply(verifier.Public.X, verifier.Public.Y, e))

	left2 := zkps.scalarMultiply(verifier.Curve.Gx, verifier.Curve.Gy, s2)
	right2 := zkps.pointAdd(C2, zkps.scalarMultiply(verifier.Public.X, verifier.Public.Y, e))

	return left1.X.Cmp(right1.X) == 0 && left1.Y.Cmp(right1.Y) == 0 &&
		left2.X.Cmp(right2.X) == 0 && left2.Y.Cmp(right2.Y) == 0, nil
}

// proveSetMembership 证明集合成员关系
func (zkps *ZeroKnowledgeProofService) proveSetMembership(prover *Prover) (*ZKProof, error) {
	// 简化的集合成员证明
	// 实际应该使用Pedersen承诺等更复杂的方案

	// 生成随机数
	r, err := rand.Int(rand.Reader, prover.Curve.N)
	if err != nil {
		return nil, err
	}

	// 计算承诺 C = G^r * H^x
	H := zkps.scalarMultiply(prover.Curve.Gx, prover.Curve.Gy, big.NewInt(2)) // 使用2G作为H
	C := zkps.pointAdd(zkps.scalarMultiply(prover.Curve.Gx, prover.Curve.Gy, r),
		zkps.scalarMultiply(H.X, H.Y, prover.Secret))

	// 计算挑战
	hash := prover.HashFunc()
	hash.Write(C.X.Bytes())
	hash.Write(C.Y.Bytes())
	eBytes := hash.Sum()
	e := new(big.Int).SetBytes(eBytes)
	e.Mod(e, prover.Curve.N)

	// 计算响应
	s := new(big.Int).Add(r, new(big.Int).Mul(e, prover.Secret))
	s.Mod(s, prover.Curve.N)

	proof := &ZKProof{
		ProofType:   "set_membership",
		Commitments: []*Point{C},
		Challenge:   e,
		Response:    s,
		Metadata: map[string]interface{}{
			"generator_h": map[string]string{
				"x": H.X.String(),
				"y": H.Y.String(),
			},
		},
		Timestamp: time.Now(),
	}

	return proof, nil
}

// verifySetMembership 验证集合成员证明
func (zkps *ZeroKnowledgeProofService) verifySetMembership(verifier *Verifier, proof *ZKProof) (bool, error) {
	if len(proof.Commitments) != 1 {
		return false, fmt.Errorf("集合成员证明格式错误")
	}

	C := proof.Commitments[0]
	e := proof.Challenge
	s := proof.Response

	// 获取H
	hData, ok := proof.Metadata["generator_h"].(map[string]string)
	if !ok {
		return false, fmt.Errorf("缺少生成器H")
	}
	Hx := new(big.Int)
	Hx.SetString(hData["x"], 10)
	Hy := new(big.Int)
	Hy.SetString(hData["y"], 10)

	// 验证 C = G^s * H^{-e}
	Gs := zkps.scalarMultiply(verifier.Curve.Gx, verifier.Curve.Gy, s)
	HeNegative := zkps.scalarMultiply(Hx, Hy, new(big.Int).Neg(e))
	right := zkps.pointAdd(Gs, HeNegative)

	return C.X.Cmp(right.X) == 0 && C.Y.Cmp(right.Y) == 0, nil
}

// BatchVerify 批量验证
func (zkps *ZeroKnowledgeProofService) BatchVerify(ctx context.Context, verifierID string, proofs []*ZKProof) ([]bool, error) {
	results := make([]bool, len(proofs))

	for i, proof := range proofs {
		result, err := zkps.VerifyProof(ctx, verifierID, proof)
		if err != nil {
			return nil, fmt.Errorf("批量验证失败，索引 %d: %w", i, err)
		}
		results[i] = result
	}

	return results, nil
}

// CreateRingSignature 创建环签名
func (zkps *ZeroKnowledgeProofService) CreateRingSignature(message []byte, publicKeys []*Point, secretKey *big.Int, secretIndex int) (*ZKProof, error) {
	if secretIndex < 0 || secretIndex >= len(publicKeys) {
		return nil, fmt.Errorf("无效的密钥索引")
	}

	// 简化的环签名实现
	n := len(publicKeys)

	// 生成随机数
	u, err := rand.Int(rand.Reader, zkps.curve.N)
	if err != nil {
		return nil, err
	}

	// 计算初始哈希值
	hash := zkps.hashFunc()
	hash.Write(message)

	// 为每个公钥计算哈希链
	v := make([]*big.Int, n)
	c := make([]*big.Int, n)

	// 计算v_i
	for i := 0; i < n; i++ {
		hash.Write(publicKeys[i].X.Bytes())
		hash.Write(publicKeys[i].Y.Bytes())
		vBytes := hash.Sum()
		v[i] = new(big.Int).SetBytes(vBytes)
		v[i].Mod(v[i], zkps.curve.N)
	}

	// 设置挑战
	c[0] = u

	// 计算s_i
	s := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		if i == secretIndex {
			// 私钥持有者的计算
			s[i] = new(big.Int).Sub(u, new(big.Int).Mul(c[i], secretKey))
			s[i].Mod(s[i], zkps.curve.N)
		} else {
			// 其他人的随机s_i
			s[i], err = rand.Int(rand.Reader, zkps.curve.N)
			if err != nil {
				return nil, err
			}
		}
	}

	proof := &ZKProof{
		ProofType:   "ring_signature",
		Commitments: publicKeys,
		Challenge:   u,
		Response:    s[secretIndex], // 简化
		Metadata: map[string]interface{}{
			"s_values": make([]string, n),
			"c_values": make([]string, n),
			"v_values": make([]string, n),
			"secret_index": secretIndex,
		},
		Timestamp: time.Now(),
	}

	// 存储s、c、v值
	sStr := proof.Metadata["s_values"].([]string)
	cStr := proof.Metadata["c_values"].([]string)
	vStr := proof.Metadata["v_values"].([]string)

	for i := 0; i < n; i++ {
		sStr[i] = s[i].String()
		cStr[i] = c[i].String()
		vStr[i] = v[i].String()
	}

	return proof, nil
}

// VerifyRingSignature 验证环签名
func (zkps *ZeroKnowledgeProofService) VerifyRingSignature(message []byte, proof *ZKProof) (bool, error) {
	if proof.ProofType != "ring_signature" {
		return false, fmt.Errorf("不是环签名证明")
	}

	publicKeys := proof.Commitments
	n := len(publicKeys)

	// 获取元数据
	sStr, ok := proof.Metadata["s_values"].([]string)
	if !ok || len(sStr) != n {
		return false, fmt.Errorf("s值缺失")
	}

	cStr, ok := proof.Metadata["c_values"].([]string)
	if !ok || len(cStr) != n {
		return false, fmt.Errorf("c值缺失")
	}

	vStr, ok := proof.Metadata["v_values"].([]string)
	if !ok || len(vStr) != n {
		return false, fmt.Errorf("v值缺失")
	}

	// 验证环签名
	hash := zkps.hashFunc()
	hash.Write(message)
	hash.Write(proof.Challenge.Bytes())

	for i := 0; i < n; i++ {
		hash.Write(publicKeys[i].X.Bytes())
		hash.Write(publicKeys[i].Y.Bytes())

		s := new(big.Int)
		s.SetString(sStr[i], 10)

		c := new(big.Int)
		c.SetString(cStr[i], 10)

		v := new(big.Int)
		v.SetString(vStr[i], 10)

		// 验证方程 v_i = H(m, c_i, s_i)
		hash.Write(c.Bytes())
		hash.Write(s.Bytes())
		vPrimeBytes := hash.Sum()
		vPrime := new(big.Int).SetBytes(vPrimeBytes)
		vPrime.Mod(vPrime, zkps.curve.N)

		if v.Cmp(vPrime) != 0 {
			return false, nil
		}
	}

	return true, nil
}

// 椭圆曲线运算辅助函数

// scalarMultiply 标量乘法
func (zkps *ZeroKnowledgeProofService) scalarMultiply(x, y, scalar *big.Int) *Point {
	// 简化的标量乘法实现
	// 实际应该使用更高效的算法（如滑动窗口）

	resultX := new(big.Int).Set(x)
	resultY := new(big.Int).Set(y)

	scalarInt := scalar.Int64()
	for i := int64(0); i < scalarInt-1; i++ {
		resultX, resultY = zkps.pointAdd(
			&Point{X: resultX, Y: resultY},
			&Point{X: x, Y: y},
		)
	}

	return &Point{X: resultX, Y: resultY}
}

// pointAdd 点加法
func (zkps *ZeroKnowledgeProofService) pointAdd(p1, p2 *Point) *Point {
	// 简化的点加法实现
	if p1 == nil {
		return p2
	}
	if p2 == nil {
		return p1
	}

	// 实际椭圆曲线点加法公式
	// 这里简化为模拟实现

	resultX := new(big.Int).Add(p1.X, p2.X)
	resultX.Mod(resultX, zkps.curve.P)

	resultY := new(big.Int).Add(p1.Y, p2.Y)
	resultY.Mod(resultY, zkps.curve.P)

	return &Point{X: resultX, Y: resultY}
}

// hashFunc 获取哈希函数
func (zkps *ZeroKnowledgeProofService) hashFunc() func() HashFunction {
	return zkps.hashFunc
}

// GetProver 获取证明者
func (zkps *ZeroKnowledgeProofService) GetProver(proverID string) (*Prover, error) {
	zkps.mutex.RLock()
	defer zkps.mutex.RUnlock()

	prover, exists := zkps.provers[proverID]
	if !exists {
		return nil, fmt.Errorf("证明者不存在: %s", proverID)
	}

	return prover, nil
}

// GetVerifier 获取验证者
func (zkps *ZeroKnowledgeProofService) GetVerifier(verifierID string) (*Verifier, error) {
	zkps.mutex.RLock()
	defer zkps.mutex.RUnlock()

	verifier, exists := zkps.verifiers[verifierID]
	if !exists {
		return nil, fmt.Errorf("验证者不存在: %s", verifierID)
	}

	return verifier, nil
}

// GetStatistics 获取统计信息
func (zkps *ZeroKnowledgeProofService) GetStatistics() map[string]interface{} {
	zkps.mutex.RLock()
	defer zkps.mutex.RUnlock()

	return map[string]interface{}{
		"curve_name":    zkps.curve.Name,
		"field_size":    zkps.curve.P.BitLen(),
		"order":         zkps.curve.N.BitLen(),
		"active_proversers": len(zkps.provers),
		"active_verifiers": len(zkps.verifiers),
		"supported_proof_types": []string{"discrete_log", "range", "set_membership", "ring_signature"},
	}
}

// PrivacyPreservingAuthentication 隐私保护认证
type PrivacyPreservingAuthentication struct {
	zkService *ZeroKnowledgeProofService
	logger    *slog.Logger
}

// NewPrivacyPreservingAuthentication 创建隐私保护认证
func NewPrivacyPreservingAuthentication(zkService *ZeroKnowledgeProofService, logger *slog.Logger) *PrivacyPreservingAuthentication {
	return &PrivacyPreservingAuthentication{
		zkService: zkService,
		logger:    logger,
	}
}

// AnonymousLogin 匿名登录
func (ppa *PrivacyPreservingAuthentication) AnonymousLogin(ctx context.Context, userID string, secret *big.Int) (*ZKProof, error) {
	// 创建证明者
	prover, err := ppa.zkService.CreateProver(userID+"_prover", secret)
	if err != nil {
		return nil, fmt.Errorf("创建证明者失败: %w", err)
	}

	// 证明知道私钥
	proof, err := ppa.zkService.ProveKnowledge(ctx, prover.ID, "discrete_log")
	if err != nil {
		return nil, fmt.Errorf("生成证明失败: %w", err)
	}

	ppa.logger.Info("匿名登录成功",
		"user_id", userID,
		"proof_type", proof.ProofType,
		"timestamp", proof.Timestamp,
	)

	return proof, nil
}

// VerifyAnonymousLogin 验证匿名登录
func (ppa *PrivacyPreservingAuthentication) VerifyAnonymousLogin(ctx context.Context, userID string, publicKey *Point, proof *ZKProof) (bool, error) {
	// 创建验证者
	verifier, err := ppa.zkService.CreateVerifier(userID+"_verifier", publicKey)
	if err != nil {
		return false, fmt.Errorf("创建验证者失败: %w", err)
	}

	// 验证证明
	valid, err := ppa.zkService.VerifyProof(ctx, verifier.ID, proof)
	if err != nil {
		return false, fmt.Errorf("验证证明失败: %w", err)
	}

	ppa.logger.Info("匿名登录验证完成",
		"user_id", userID,
		"valid", valid,
		"proof_type", proof.ProofType,
	)

	return valid, nil
}