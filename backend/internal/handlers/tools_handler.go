package handlers

import (
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"law-oa-go/internal/common"
)

type ToolsHandler struct{}

func NewToolsHandler() *ToolsHandler {
	return &ToolsHandler{}
}

type LitigationFeeRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type LitigationFeeBracket struct {
	From   float64 `json:"from"`
	To     float64 `json:"to,omitempty"`
	Rate   float64 `json:"rate"`
	Amount float64 `json:"amount"`
	Fee    float64 `json:"fee"`
}

type LitigationFeeResponse struct {
	Amount          float64                `json:"amount"`
	Fee             float64                `json:"fee"`
	CalculationTime int64                  `json:"calculationTime"`
	Brackets        []LitigationFeeBracket `json:"brackets"`
}

type InterestCalculatorRequest struct {
	Principal float64 `json:"principal" binding:"required,gt=0"`
	Rate      float64 `json:"rate" binding:"required,gte=0"`
	Days      int     `json:"days" binding:"required,gt=0"`
	Type      string  `json:"type" binding:"required,oneof=simple compound penalty"`
}

type InterestCalculatorResponse struct {
	Principal float64 `json:"principal"`
	Rate      float64 `json:"rate"`
	Days      int     `json:"days"`
	Type      string  `json:"type"`
	Interest  float64 `json:"interest"`
	Total     float64 `json:"total"`
}

type DeadlineCalculatorRequest struct {
	StartDate       string `json:"startDate" binding:"required"`
	Days            int    `json:"days" binding:"required,gt=0"`
	ExcludeWeekends bool   `json:"excludeWeekends"`
	ExcludeHolidays bool   `json:"excludeHolidays"`
}

type DeadlineCalculatorResponse struct {
	StartDate       string `json:"startDate"`
	Days            int    `json:"days"`
	ExcludeWeekends bool   `json:"excludeWeekends"`
	ExcludeHolidays bool   `json:"excludeHolidays"`
	EndDate         string `json:"endDate"`
	WorkDays        int    `json:"workDays"`
	TotalDays       int    `json:"totalDays"`
}

func (h *ToolsHandler) CalculateLitigationFee(c *gin.Context) {
	var req LitigationFeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	fee, brackets := calculatePropertyLitigationFee(req.Amount)
	common.APISuccess(c, LitigationFeeResponse{
		Amount:          roundMoney(req.Amount),
		Fee:             roundMoney(fee),
		CalculationTime: time.Now().UnixMilli(),
		Brackets:        brackets,
	})
}

func (h *ToolsHandler) CalculateInterest(c *gin.Context) {
	var req InterestCalculatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	annualRate := req.Rate / 100
	days := float64(req.Days)
	var interest float64
	switch req.Type {
	case "compound":
		interest = req.Principal*math.Pow(1+annualRate/365, days) - req.Principal
	case "penalty":
		interest = req.Principal * annualRate * days / 365
	default:
		interest = req.Principal * annualRate * days / 365
	}

	common.APISuccess(c, InterestCalculatorResponse{
		Principal: roundMoney(req.Principal),
		Rate:      req.Rate,
		Days:      req.Days,
		Type:      req.Type,
		Interest:  roundMoney(interest),
		Total:     roundMoney(req.Principal + interest),
	})
}

func (h *ToolsHandler) CalculateDeadline(c *gin.Context) {
	var req DeadlineCalculatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.APIBadRequest(c, "请求参数错误", err.Error())
		return
	}

	startDate, err := parseToolDate(req.StartDate)
	if err != nil {
		common.APIBadRequest(c, "开始日期格式错误", "请使用 YYYY-MM-DD 日期格式")
		return
	}

	endDate := startDate
	workDays := 0
	totalDays := 0
	for workDays < req.Days {
		endDate = endDate.AddDate(0, 0, 1)
		totalDays++
		if shouldCountToolDate(endDate, req.ExcludeWeekends, req.ExcludeHolidays) {
			workDays++
		}
	}

	common.APISuccess(c, DeadlineCalculatorResponse{
		StartDate:       startDate.Format("2006-01-02"),
		Days:            req.Days,
		ExcludeWeekends: req.ExcludeWeekends,
		ExcludeHolidays: req.ExcludeHolidays,
		EndDate:         endDate.Format("2006-01-02"),
		WorkDays:        workDays,
		TotalDays:       totalDays,
	})
}

func calculatePropertyLitigationFee(amount float64) (float64, []LitigationFeeBracket) {
	if amount <= 10000 {
		return 50, []LitigationFeeBracket{{From: 0, To: 10000, Rate: 0, Amount: roundMoney(amount), Fee: 50}}
	}

	fee := 50.0
	brackets := []LitigationFeeBracket{{From: 0, To: 10000, Rate: 0, Amount: 10000, Fee: 50}}
	tiers := []struct {
		from float64
		to   float64
		rate float64
	}{
		{10000, 100000, 0.025},
		{100000, 200000, 0.020},
		{200000, 500000, 0.015},
		{500000, 1000000, 0.010},
		{1000000, 2000000, 0.009},
		{2000000, 5000000, 0.008},
		{5000000, 10000000, 0.007},
		{10000000, 20000000, 0.006},
		{20000000, math.Inf(1), 0.005},
	}

	for _, tier := range tiers {
		if amount <= tier.from {
			break
		}
		upper := math.Min(amount, tier.to)
		taxable := upper - tier.from
		itemFee := taxable * tier.rate
		fee += itemFee
		brackets = append(brackets, LitigationFeeBracket{
			From:   tier.from,
			To:     finiteTo(tier.to),
			Rate:   tier.rate,
			Amount: roundMoney(taxable),
			Fee:    roundMoney(itemFee),
		})
		if amount <= tier.to {
			break
		}
	}

	return fee, brackets
}

func finiteTo(value float64) float64 {
	if math.IsInf(value, 1) {
		return 0
	}
	return value
}

func parseToolDate(value string) (time.Time, error) {
	if parsed, err := time.ParseInLocation("2006-01-02", value, time.Local); err == nil {
		return parsed, nil
	}
	return time.Parse(time.RFC3339, value)
}

func shouldCountToolDate(date time.Time, excludeWeekends, excludeHolidays bool) bool {
	if excludeWeekends && (date.Weekday() == time.Saturday || date.Weekday() == time.Sunday) {
		return false
	}
	if excludeHolidays && isFixedMainlandHoliday(date) {
		return false
	}
	return true
}

func isFixedMainlandHoliday(date time.Time) bool {
	month := date.Month()
	day := date.Day()
	if month == time.January && day == 1 {
		return true
	}
	if month == time.May && day == 1 {
		return true
	}
	if month == time.October && day >= 1 && day <= 7 {
		return true
	}
	return false
}

func roundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}
