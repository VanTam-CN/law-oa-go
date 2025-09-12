package models

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
