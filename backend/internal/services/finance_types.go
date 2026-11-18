package services

// FinanceOverview 财务概览
type FinanceOverview struct {
	ContractStats   *ContractStats   `json:"contract_stats"`
	InvoiceStats    *InvoiceStats    `json:"invoice_stats"`
	PaymentStats    *PaymentStats    `json:"payment_stats"`
	CommissionStats *CommissionStats `json:"commission_stats"`
}
