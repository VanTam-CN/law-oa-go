package config

// ConflictDetectionConfig 冲突检测配置
type ConflictDetectionConfig struct {
	Enabled                    bool    `json:"enabled" yaml:"enabled"`
	AutoCheckOnCaseCreation    bool    `json:"autoCheckOnCaseCreation" yaml:"autoCheckOnCaseCreation"`
	DefaultSearchYears         int     `json:"defaultSearchYears" yaml:"defaultSearchYears"`
	DefaultSearchDepth         string  `json:"defaultSearchDepth" yaml:"defaultSearchDepth"`
	IncludeCorporateRelations  bool    `json:"includeCorporateRelations" yaml:"includeCorporateRelations"`
	HighRiskThreshold          float64 `json:"highRiskThreshold" yaml:"highRiskThreshold"`
	MediumRiskThreshold        float64 `json:"mediumRiskThreshold" yaml:"mediumRiskThreshold"`
	RequireApprovalForHighRisk bool    `json:"requireApprovalForHighRisk" yaml:"requireApprovalForHighRisk"`
	AllowSkipConflictCheck     bool    `json:"allowSkipConflictCheck" yaml:"allowSkipConflictCheck"`
}

// DefaultConflictDetectionConfig 默认冲突检测配置
func DefaultConflictDetectionConfig() *ConflictDetectionConfig {
	return &ConflictDetectionConfig{
		Enabled:                 true,
		AutoCheckOnCaseCreation: true,
		// Zero means full historical coverage. A bounded default can create a
		// false sense of safety for former clients and is not allowed by P0.
		DefaultSearchYears:         0,
		DefaultSearchDepth:         "deep",
		IncludeCorporateRelations:  true,
		HighRiskThreshold:          0.7,
		MediumRiskThreshold:        0.4,
		RequireApprovalForHighRisk: true,
		AllowSkipConflictCheck:     false,
	}
}

// Validate 验证配置
func (c *ConflictDetectionConfig) Validate() error {
	if c.DefaultSearchYears < 0 {
		return ErrInvalidSearchYears
	}

	if c.DefaultSearchDepth == "" {
		return ErrInvalidSearchDepth
	}

	if c.HighRiskThreshold <= 0 || c.HighRiskThreshold > 1 {
		return ErrInvalidRiskThreshold
	}

	if c.MediumRiskThreshold <= 0 || c.MediumRiskThreshold > 1 {
		return ErrInvalidRiskThreshold
	}

	if c.MediumRiskThreshold >= c.HighRiskThreshold {
		return ErrInvalidRiskThresholdRange
	}

	return nil
}

// 配置错误定义
var (
	ErrInvalidSearchYears        = NewConfigError("search years must be zero (full history) or greater")
	ErrInvalidSearchDepth        = NewConfigError("search depth cannot be empty")
	ErrInvalidRiskThreshold      = NewConfigError("risk threshold must be between 0 and 1")
	ErrInvalidRiskThresholdRange = NewConfigError("medium risk threshold must be less than high risk threshold")
)

// ConfigError 配置错误
type ConfigError struct {
	Message string
}

func NewConfigError(message string) *ConfigError {
	return &ConfigError{Message: message}
}

func (e *ConfigError) Error() string {
	return e.Message
}
