package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/models"
	"law-oa-go/internal/repositories"
)

// TrustBalanceMiddleware 代管款余额校验中间件
type TrustBalanceMiddleware struct {
	accountRepo repositories.TrustAccountRepository
}

// NewTrustBalanceMiddleware 创建余额校验中间件实例
func NewTrustBalanceMiddleware(
	accountRepo repositories.TrustAccountRepository,
) *TrustBalanceMiddleware {
	return &TrustBalanceMiddleware{
		accountRepo: accountRepo,
	}
}

// CheckBalance 检查账户余额
// 从路径参数中获取账户ID，从查询参数或请求体中获取金额
func (m *TrustBalanceMiddleware) CheckBalance() gin.HandlerFunc {
	return func(c *gin.Context) {
		accountID := c.Param("account_id")
		if accountID == "" {
			// 尝试从查询参数获取
			accountID = c.Query("account_id")
		}

		if accountID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "缺少账户ID"})
			c.Abort()
			return
		}

		id, err := strconv.ParseUint(accountID, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的账户ID"})
			c.Abort()
			return
		}

		// 获取账户信息
		account, err := m.accountRepo.FindByID(c.Request.Context(), uint(id))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询账户失败"})
			c.Abort()
			return
		}
		if account == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "账户不存在"})
			c.Abort()
			return
		}

		// 检查账户状态
		if account.Status != "active" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": fmt.Sprintf("账户状态不正确，当前状态: %s", account.Status),
				"status": account.Status,
			})
			c.Abort()
			return
		}

		// 将账户信息存入上下文
		c.Set("trust_account", account)
		c.Next()
	}
}

// RequireSufficientBalance 要求余额充足
// 检查请求中的金额参数是否超过可用余额
func (m *TrustBalanceMiddleware) RequireSufficientBalance() gin.HandlerFunc {
	return func(c *gin.Context) {
		account, exists := c.Get("trust_account")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "账户信息未找到"})
			c.Abort()
			return
		}

		trustAccount, ok := account.(*models.ClientTrustAccount)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "账户类型错误"})
			c.Abort()
			return
		}

		// 获取交易金额
		amount, err := m.extractAmount(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// 计算可用余额
		availableBalance := trustAccount.Balance - trustAccount.FrozenAmount

		if amount > availableBalance {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "可用余额不足",
				"available_balance": availableBalance,
				"required_amount":    amount,
				"shortage":           amount - availableBalance,
			})
			c.Abort()
			return
		}

		// 将可用余额存入上下文
		c.Set("available_balance", availableBalance)
		c.Next()
	}
}

// extractAmount 从请求中提取金额
func (m *TrustBalanceMiddleware) extractAmount(c *gin.Context) (float64, error) {
	// 1. 尝试从查询参数获取
	if amountStr := c.Query("amount"); amountStr != "" {
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return 0, errors.New("无效的金额格式")
		}
		return amount, nil
	}

	// 2. 尝试从表单获取
	if amountStr := c.PostForm("amount"); amountStr != "" {
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return 0, errors.New("无效的金额格式")
		}
		return amount, nil
	}

	// 3. 尝试从 JSON 请求体获取
	var jsonBody map[string]interface{}
	if err := c.BindJSON(&jsonBody); err == nil {
		if amountVal, ok := jsonBody["amount"]; ok {
			switch v := amountVal.(type) {
			case float64:
				return v, nil
			case float32:
				return float64(v), nil
			case int:
				return float64(v), nil
			case int64:
				return float64(v), nil
			case string:
				amount, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return 0, errors.New("无效的金额格式")
				}
				return amount, nil
			default:
				return 0, errors.New("无效的金额类型")
			}
		}
	}

	return 0, errors.New("未找到金额参数")
}

// RequireSufficientBalanceForAmount 检查指定金额是否充足
// 用于需要在中间件中指定特定金额的场景
func (m *TrustBalanceMiddleware) RequireSufficientBalanceForAmount(amountGetter func(*gin.Context) (float64, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		account, exists := c.Get("trust_account")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "账户信息未找到"})
			c.Abort()
			return
		}

		trustAccount, ok := account.(*models.ClientTrustAccount)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "账户类型错误"})
			c.Abort()
			return
		}

		// 获取金额
		amount, err := amountGetter(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// 计算可用余额
		availableBalance := trustAccount.Balance - trustAccount.FrozenAmount

		if amount > availableBalance {
			c.JSON(http.StatusForbidden, gin.H{
				"error":             "可用余额不足",
				"available_balance": availableBalance,
				"required_amount":    amount,
				"shortage":           amount - availableBalance,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetAccountBalance 获取账户余额信息
func (m *TrustBalanceMiddleware) GetAccountBalance(c *gin.Context) (balance, available, frozen float64, err error) {
	accountID := c.Param("account_id")
	if accountID == "" {
		accountID = c.Query("account_id")
	}

	if accountID == "" {
		return 0, 0, 0, errors.New("缺少账户ID")
	}

	id, err := strconv.ParseUint(accountID, 10, 32)
	if err != nil {
		return 0, 0, 0, errors.New("无效的账户ID")
	}

	account, err := m.accountRepo.FindByID(c.Request.Context(), uint(id))
	if err != nil {
		return 0, 0, 0, err
	}
	if account == nil {
		return 0, 0, 0, errors.New("账户不存在")
	}

	return account.Balance, account.Balance - account.FrozenAmount, account.FrozenAmount, nil
}

// BalanceResponse 余额信息响应
type BalanceResponse struct {
	AccountID        uint    `json:"account_id"`
	AccountCode      string  `json:"account_code"`
	Balance          float64 `json:"balance"`
	FrozenAmount     float64 `json:"frozen_amount"`
	AvailableBalance float64 `json:"available_balance"`
	Currency         string  `json:"currency"`
	Status           string  `json:"status"`
}
