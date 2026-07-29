package models

import (
	"strings"

	"law-oa-go/internal/security"
)

// ClientStats 客户统计信息
type ClientStats struct {
	TotalClients        int64 `json:"total_clients"`
	ActiveClients       int64 `json:"active_clients"`
	InactiveClients     int64 `json:"inactive_clients"`
	NewClientsThisMonth int64 `json:"new_clients_this_month"`
}

// TableName 指定表名
func (Client) TableName() string {
	return "clients"
}

// MaskIDCard 对身份证号进行脱敏处理，保留前3位和后4位
func MaskIDCard(idCard string) string {
	if len(idCard) <= 7 {
		return "***" // 太短，全部脱敏
	}
	return idCard[:3] + "***" + idCard[len(idCard)-4:]
}

// MaskPhone 对手机号进行脱敏处理，保留前3位和后4位
func MaskPhone(phone string) string {
	if len(phone) <= 7 {
		return "***" // 太短，全部脱敏
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// ClientSafeResponse 用于API响应的安全客户端信息，敏感字段已脱敏
type ClientSafeResponse struct {
	ID            uint   `json:"id"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Email         string `json:"email"`
	Phone         string `json:"phone"` // 脱敏后的手机号
	Address       string `json:"address"`
	Company       string `json:"company"`
	IDCard        string `json:"id_card"` // 脱敏后的身份证号
	Industry      string `json:"industry"`
	ContactPerson string `json:"contact_person"`
	ContactPhone  string `json:"contact_phone"` // 脱敏后的联系电话
	Source        string `json:"source"`
	Notes         string `json:"notes"`
	Status        string `json:"status"`
}

// ToSafeResponse 将 Client 转换为安全的 API 响应格式
func (c *Client) ToSafeResponse() ClientSafeResponse {
	return ClientSafeResponse{
		ID:            c.ID,
		CreatedAt:     c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Name:          c.Name,
		Type:          c.Type,
		Email:         c.Email,
		Phone:         MaskPhone(c.Phone),
		Address:       c.Address,
		Company:       c.Company,
		IDCard:        maskStoredIDCard(c),
		Industry:      c.Industry,
		ContactPerson: c.ContactPerson,
		ContactPhone:  MaskPhone(c.ContactPhone),
		Source:        c.Source,
		Notes:         c.Notes,
		Status:        c.Status,
	}
}

func (c *Client) HasIDCard() bool {
	if c == nil {
		return false
	}
	return security.IdentityPresent(c.IDCard, c.IDCardCiphertext, c.IDCardDigest)
}

func (c *Client) DecryptedIDCard() (string, error) {
	if c == nil {
		return "", nil
	}
	if strings.TrimSpace(c.IDCard) != "" {
		return strings.TrimSpace(c.IDCard), nil
	}
	return security.DecryptIdentityNumber(c.IDCardCiphertext)
}

func maskStoredIDCard(c *Client) string {
	value, _ := c.DecryptedIDCard()
	if value == "" && c.HasIDCard() {
		return "已登记（受保护）"
	}
	return MaskIDCard(value)
}

// ClientsToSafeResponse 批量转换 Client 列片为安全的 API 响应格式
func ClientsToSafeResponse(clients []Client) []ClientSafeResponse {
	responses := make([]ClientSafeResponse, len(clients))
	for i, client := range clients {
		responses[i] = client.ToSafeResponse()
	}
	return responses
}
