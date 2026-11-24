package services

import (
	"time"
)

// RuleMatch 规则匹配结果
type RuleMatch struct {
	RuleID       string                 `json:"ruleId"`
	RuleName     string                 `json:"ruleName"`
	RuleType     string                 `json:"ruleType"`
	Matched      bool                   `json:"matched"`
	MatchDetails *RuleMatchDetails      `json:"matchDetails,omitempty"`
	RiskScore    float64                `json:"riskScore"`
	EvaluatedAt  time.Time              `json:"evaluatedAt"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// RuleMatchDetails 规则匹配详情
type RuleMatchDetails struct {
	MatchType      string   `json:"matchType"`
	MatchedFields  []string `json:"matchedFields"`
	MatchedValues  []string `json:"matchedValues"`
	Confidence     float64  `json:"confidence"`
	Description    string   `json:"description"`
	Recommendation string   `json:"recommendation"`
}
