package editing

import (
	"bufio"
	"context"
	"fmt"
	"html"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ============ 高级差异对比核心接口 ============

// DiffEngine 差异引擎接口
type DiffEngine interface {
	ComputeDiff(oldContent, newContent string) ([]*DiffBlock, error)
	ComputeUnifiedDiff(oldContent, newContent string, options *UnifiedDiffOptions) (string, error)
	ComputeSideBySideDiff(oldContent, newContent string, options *SideBySideOptions) (string, error)
	ComputeSemanticDiff(oldContent, newContent string, options *SemanticDiffOptions) ([]*SemanticDiff, error)
}

// DiffFormatter 差异格式化器接口
type DiffFormatter interface {
	FormatUnified(diffs []*DiffBlock, options *UnifiedDiffOptions) (string, error)
	FormatHTML(diffs []*DiffBlock, options *HTMLDiffOptions) (string, error)
	FormatSideBySide(diffs []*DiffBlock, options *SideBySideOptions) (string, error)
	FormatJSON(diffs []*DiffBlock, options *JSONDiffOptions) (string, error)
}

// ConflictSolver 冲突解决器接口
type ConflictSolver interface {
	DetectConflicts(base, left, right string) ([]*Conflict, error)
	AutoResolveConflicts(conflicts []*Conflict) ([]*ConflictResolution, error)
	SuggestResolution(conflict *Conflict) (string, error)
}

// ============ 数据结构定义 ============

// DiffBlock 差异块
type DiffBlock struct {
	Type        DiffBlockType   `json:"type"`
	OldLines    []string        `json:"old_lines"`
	NewLines    []string        `json:"new_lines"`
	OldStart    int             `json:"old_start"`
	NewStart    int             `json:"new_start"`
	OldLength   int             `json:"old_length"`
	NewLength   int             `json:"new_length"`
	Context     []string        `json:"context,omitempty"`
	Similarity  float64         `json:"similarity,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// DiffBlockType 差异块类型
type DiffBlockType string

const (
	DiffBlockEqual   DiffBlockType = "equal"
	DiffBlockInsert  DiffBlockType = "insert"
	DiffBlockDelete  DiffBlockType = "delete"
	DiffBlockReplace DiffBlockType = "replace"
)

// SemanticDiff 语义差异
type SemanticDiff struct {
	Type        SemanticDiffType `json:"type"`
	Description string           `json:"description"`
	OldValue    interface{}      `json:"old_value"`
	NewValue    interface{}      `json:"new_value"`
	Position    Position         `json:"position"`
	Confidence  float64          `json:"confidence"`
}

// SemanticDiffType 语义差异类型
type SemanticDiffType string

const (
	SemanticDiffStructural    SemanticDiffType = "structural"
	SemanticDiffContent       SemanticDiffType = "content"
	SemanticDiffFormatting    SemanticDiffType = "formatting"
	SemanticDiffComments      SemanticDiffType = "comments"
	SemanticDiffWhitespace    SemanticDiffType = "whitespace"
	SemanticDiffIdentifiers   SemanticDiffType = "identifiers"
)

// Position 位置信息
type Position struct {
	Line      int `json:"line"`
	Column    int `json:"column"`
	Length    int `json:"length"`
	Offset    int `json:"offset"`
}

// Conflict 冲突信息
type Conflict struct {
	ID          string     `json:"id"`
	Type        ConflictType `json:"type"`
	Position    Position   `json:"position"`
	BaseContent string     `json:"base_content"`
	LeftContent string     `json:"left_content"`
	RightContent string    `json:"right_content"`
	Description string     `json:"description"`
	Resolution  string     `json:"resolution,omitempty"`
	AutoResolved bool     `json:"auto_resolved"`
	CreatedAt   time.Time  `json:"created_at"`
}

// ConflictType 冲突类型
type ConflictType string

const (
	ConflictTypeContent      ConflictType = "content"
	ConflictTypeFormatting   ConflictType = "formatting"
	ConflictTypeStructure    ConflictType = "structure"
	ConflictTypeWhitespace   ConflictType = "whitespace"
	ConflictTypeLineEnding   ConflictType = "line_ending"
)

// ConflictResolution 冲突解决方案
type ConflictResolution struct {
	Action      ConflictAction `json:"action"`
	Content     string         `json:"content"`
	Reason      string         `json:"reason"`
	AutoResolved bool          `json:"auto_resolved"`
	Confidence  float64        `json:"confidence"`
}

// ConflictAction 冲突解决动作
type ConflictAction string

const (
	ConflictActionAcceptBase   ConflictAction = "accept_base"
	ConflictActionAcceptLeft   ConflictAction = "accept_left"
	ConflictActionAcceptRight  ConflictAction = "accept_right"
	ConflictActionMerge        ConflictAction = "merge"
	ConflictActionManual       ConflictAction = "manual"
)

// UnifiedDiffOptions 统一差异选项
type UnifiedDiffOptions struct {
	ContextLines    int               `json:"context_lines"`
	ShowLineNumbers bool              `json:"show_line_numbers"`
	IgnoreWhitespace bool             `json:"ignore_whitespace"`
	IgnoreComments  bool              `json:"ignore_comments"`
	MaxDiffBlocks   int               `json:"max_diff_blocks"`
	OutputFormat    DiffOutputFormat   `json:"output_format"`
}

// SideBySideOptions 并排对比选项
type SideBySideOptions struct {
	ContextLines     int               `json:"context_lines"`
	ShowLineNumbers  bool              `json:"show_line_numbers"`
	HighlightChanges bool              `json:"highlight_changes"`
	MatchSimilarLines bool             `json:"match_similar_lines"`
	OutputFormat     DiffOutputFormat   `json:"output_format"`
}

// SemanticDiffOptions 语义差异选项
type SemanticDiffOptions struct {
	Language         string            `json:"language"`
	IgnoreFormatting bool              `json:"ignore_formatting"`
	IgnoreComments   bool              `json:"ignore_comments"`
	NormalizeIdentifiers bool           `json:"normalize_identifiers"`
	ConfidenceThreshold float64         `json:"confidence_threshold"`
}

// HTMLDiffOptions HTML差异选项
type HTMLDiffOptions struct {
	ShowLineNumbers  bool              `json:"show_line_numbers"`
	SyntaxHighlight   bool              `json:"syntax_highlight"`
	Language         string            `json:"language"`
	CSSClasses       map[string]string `json:"css_classes"`
	TableLayout      bool              `json:"table_layout"`
}

// JSONDiffOptions JSON差异选项
type JSONDiffOptions struct {
	PrettyPrint      bool              `json:"pretty_print"`
	IncludeMetadata  bool              `json:"include_metadata"`
	CompressionLevel int               `json:"compression_level"`
}

// DiffOutputFormat 差异输出格式
type DiffOutputFormat string

const (
	DiffFormatUnified     DiffOutputFormat = "unified"
	DiffFormatSideBySide  DiffOutputFormat = "side_by_side"
	DiffFormatHTML        DiffOutputFormat = "html"
	DiffFormatJSON        DiffOutputFormat = "json"
	DiffFormatMarkdown    DiffOutputFormat = "markdown"
)

// ============ Myers差异算法实现 ============

// MyersDiffer 基于Myers算法的差异计算器
type MyersDiffer struct {
	logger *logrus.Logger
}

// NewMyersDiffer 创建Myers差异计算器
func NewMyersDiffer(logger *logrus.Logger) DiffEngine {
	return &MyersDiffer{
		logger: logger,
	}
}

// ComputeDiff 计算差异
func (m *MyersDiffer) ComputeDiff(oldContent, newContent string) ([]*DiffBlock, error) {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	diffs := m.computeMyersDiff(oldLines, newLines)
	blocks := m.groupDiffsIntoBlocks(diffs, oldLines, newLines)

	return blocks, nil
}

// computeMyersDiff 执行Myers差异算法
func (m *MyersDiffer) computeMyersDiff(oldLines, newLines []string) []Edit {
	n, p := len(oldLines), len(newLines)
	max := n + p

	// 动态规划表
	v := make([]int, 2*max+1)
	var trace [][]int

	for d := 0; d <= max; d++ {
		trace = append(trace, make([]int, len(v)))
		copy(trace[d], v)

		for k := -d; k <= d; k += 2 {
			var x int
			if k == -d || (k != d && v[k-1] < v[k+1]) {
				x = v[k+1]
			} else {
				x = v[k-1] + 1
			}

			y := x - k
			for x < n && y < p && oldLines[x] == newLines[y] {
				x++
				y++
			}

			v[k] = x
			if x >= n && y >= p {
				return m.backtrack(trace, n, p, d, oldLines, newLines)
			}
		}
	}

	return []Edit{}
}

// Edit 编辑操作
type Edit struct {
	Type     EditType
	OldStart int
	OldEnd   int
	NewStart int
	NewEnd   int
}

// EditType 编辑类型
type EditType int

const (
	EditEqual EditType = iota
	EditInsert
	EditDelete
)

// backtrack 回溯差异路径
func (m *MyersDiffer) backtrack(trace [][]int, n, p, d int, oldLines, newLines []string) []Edit {
	var edits []Edit

	x, y := n, p
	for k := d; k > 0; k-- {
		v := trace[k]
		if x == 0 && y == 0 {
			break
		}

		var prevK int
		if k == -d || (k != d && v[k-1] < v[k+1]) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}

		prevX := v[prevK]
		prevY := prevX - prevK

		// 处理相等的行
		for x > prevX && y > prevY {
			edits = append(edits, Edit{
				Type:     EditEqual,
				OldStart: x - 1,
				OldEnd:   x,
				NewStart: y - 1,
				NewEnd:   y,
			})
			x--
			y--
		}

		// 处理编辑操作
		if x == prevX {
			// 插入操作
			edits = append(edits, Edit{
				Type:     EditInsert,
				OldStart: x,
				OldEnd:   x,
				NewStart: prevY,
				NewEnd:   y,
			})
		} else if y == prevY {
			// 删除操作
			edits = append(edits, Edit{
				Type:     EditDelete,
				OldStart: prevX,
				OldEnd:   x,
				NewStart: y,
				NewEnd:   y,
			})
		}

		x, y = prevX, prevY
	}

	// 处理剩余的相等行
	for x > 0 && y > 0 {
		edits = append(edits, Edit{
			Type:     EditEqual,
			OldStart: x - 1,
			OldEnd:   x,
			NewStart: y - 1,
			NewEnd:   y,
		})
		x--
		y--
	}

	// 反转编辑顺序
	for i, j := 0, len(edits)-1; i < j; i, j = i+1, j-1 {
		edits[i], edits[j] = edits[j], edits[i]
	}

	return edits
}

// groupDiffsIntoBlocks 将差异组合成块
func (m *MyersDiffer) groupDiffsIntoBlocks(edits []Edit, oldLines, newLines []string) []*DiffBlock {
	var blocks []*DiffBlock
	var currentBlock *DiffBlock

	for _, edit := range edits {
		block := &DiffBlock{
			OldStart: edit.OldStart,
			NewStart: edit.NewStart,
		}

		switch edit.Type {
		case EditEqual:
			if currentBlock != nil {
				blocks = append(blocks, currentBlock)
				currentBlock = nil
			}
			continue

		case EditInsert:
			block.Type = DiffBlockInsert
			block.NewLines = newLines[edit.NewStart:edit.NewEnd]
			block.NewLength = edit.NewEnd - edit.NewStart

		case EditDelete:
			block.Type = DiffBlockDelete
			block.OldLines = oldLines[edit.OldStart:edit.OldEnd]
			block.OldLength = edit.OldEnd - edit.OldStart
		}

		if currentBlock != nil && m.canMergeBlocks(currentBlock, block) {
			currentBlock = m.mergeBlocks(currentBlock, block)
		} else {
			if currentBlock != nil {
				blocks = append(blocks, currentBlock)
			}
			currentBlock = block
		}
	}

	if currentBlock != nil {
		blocks = append(blocks, currentBlock)
	}

	return blocks
}

// canMergeBlocks 检查是否可以合并块
func (m *MyersDiffer) canMergeBlocks(block1, block2 *DiffBlock) bool {
	// 相同类型且连续的块可以合并
	return block1.Type == block2.Type &&
		block1.OldStart+block1.OldLength == block2.OldStart &&
		block1.NewStart+block1.NewLength == block2.NewStart
}

// mergeBlocks 合并块
func (m *MyersDiffer) mergeBlocks(block1, block2 *DiffBlock) *DiffBlock {
	merged := &DiffBlock{
		Type:     block1.Type,
		OldStart: block1.OldStart,
		NewStart: block1.NewStart,
	}

	merged.OldLines = append(append([]string{}, block1.OldLines...), block2.OldLines...)
	merged.NewLines = append(append([]string{}, block1.NewLines...), block2.NewLines...)
	merged.OldLength = block1.OldLength + block2.OldLength
	merged.NewLength = block1.NewLength + block2.NewLength

	return merged
}

// ComputeUnifiedDiff 计算统一格式差异
func (m *MyersDiffer) ComputeUnifiedDiff(oldContent, newContent string, options *UnifiedDiffOptions) (string, error) {
	blocks, err := m.ComputeDiff(oldContent, newContent)
	if err != nil {
		return "", err
	}

	formatter := &UnifiedFormatter{}
	return formatter.FormatUnified(blocks, options)
}

// ComputeSideBySideDiff 计算并排差异
func (m *MyersDiffer) ComputeSideBySideDiff(oldContent, newContent string, options *SideBySideOptions) (string, error) {
	blocks, err := m.ComputeDiff(oldContent, newContent)
	if err != nil {
		return "", err
	}

	formatter := &SideBySideFormatter{}
	return formatter.FormatSideBySide(blocks, options)
}

// ComputeSemanticDiff 计算语义差异
func (m *MyersDiffer) ComputeSemanticDiff(oldContent, newContent string, options *SemanticDiffOptions) ([]*SemanticDiff, error) {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	// 基础语法分析
	oldTokens := m.tokenizeContent(oldContent, options.Language)
	newTokens := m.tokenizeContent(newContent, options.Language)

	// 计算语义差异
	var semanticDiffs []*SemanticDiff

	// 结构差异
	if options.Language != "" {
		structDiffs := m.analyzeStructuralDifferences(oldTokens, newTokens)
		semanticDiffs = append(semanticDiffs, structDiffs...)
	}

	// 内容差异
	contentDiffs := m.analyzeContentDifferences(oldLines, newLines, options)
	semanticDiffs = append(semanticDiffs, contentDiffs...)

	// 格式化差异
	if !options.IgnoreFormatting {
		formatDiffs := m.analyzeFormattingDifferences(oldContent, newContent)
		semanticDiffs = append(semanticDiffs, formatDiffs...)
	}

	// 过滤低置信度的差异
	var filteredDiffs []*SemanticDiff
	for _, diff := range semanticDiffs {
		if diff.Confidence >= options.ConfidenceThreshold {
			filteredDiffs = append(filteredDiffs, diff)
		}
	}

	return filteredDiffs, nil
}

// tokenizeContent 对内容进行分词
func (m *MyersDiffer) tokenizeContent(content, language string) []Token {
	// 简化的分词实现
	var tokens []Token
	lines := strings.Split(content, "\n")

	for lineNum, line := range lines {
		if language == "go" || language == "javascript" {
			codeTokens := m.tokenizeCode(line)
			for _, token := range codeTokens {
				token.Position.Line = lineNum
				tokens = append(tokens, token)
			}
		} else {
			// 普通文本分词
			words := strings.Fields(line)
			for _, word := range words {
				tokens = append(tokens, Token{
					Type:  TokenTypeWord,
					Value: word,
					Position: Position{
						Line: lineNum,
					},
				})
			}
		}
	}

	return tokens
}

// Token 词法单元
type Token struct {
	Type     TokenType
	Value    string
	Position Position
}

// TokenType 词法单元类型
type TokenType int

const (
	TokenTypeKeyword TokenType = iota
	TokenTypeIdentifier
	TokenTypeString
	TokenTypeComment
	TokenTypeWhitespace
	TokenTypeOperator
	TokenTypeNumber
	TokenTypeWord
	TokenTypePunctuation
)

// tokenizeCode 对代码进行分词
func (m *MyersDiffer) tokenizeCode(line string) []Token {
	// 简化的代码分词
	var tokens []Token

	// Go语言关键字
	keywords := map[string]bool{
		"func": true, "var": true, "const": true, "type": true,
		"struct": true, "interface": true, "if": true, "else": true,
		"for": true, "range": true, "return": true, "import": true,
		"package": true, "go": true, "defer": true, "select": true,
		"case": true, "default": true, "switch": true, "break": true,
		"continue": true, "fallthrough": true, "goto": true,
	}

	// 正则表达式匹配
	identifierRegex := regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`)
	stringRegex := regexp.MustCompile(`"([^"\\]|\\.)*"`)
	commentRegex := regexp.MustCompile(`//.*|/\*[\s\S]*?\*/`)
	numberRegex := regexp.MustCompile(`\d+(\.\d+)?`)

	// 提取注释
	comments := commentRegex.FindAllStringIndex(line, -1)
	for _, match := range comments {
		tokens = append(tokens, Token{
			Type:  TokenTypeComment,
			Value: line[match[0]:match[1]],
			Position: Position{
				Column: match[0],
				Length: match[1] - match[0],
			},
		})
	}

	// 提取字符串
	strings := stringRegex.FindAllStringIndex(line, -1)
	for _, match := range strings {
		tokens = append(tokens, Token{
			Type:  TokenTypeString,
			Value: line[match[0]:match[1]],
			Position: Position{
				Column: match[0],
				Length: match[1] - match[0],
			},
		})
	}

	// 提取数字
	numbers := numberRegex.FindAllStringIndex(line, -1)
	for _, match := range numbers {
		tokens = append(tokens, Token{
			Type:  TokenTypeNumber,
			Value: line[match[0]:match[1]],
			Position: Position{
				Column: match[0],
				Length: match[1] - match[0],
			},
		})
	}

	// 提取标识符和关键字
	identifiers := identifierRegex.FindAllStringIndex(line, -1)
	for _, match := range identifiers {
		value := line[match[0]:match[1]]
		tokenType := TokenTypeIdentifier
		if keywords[value] {
			tokenType = TokenTypeKeyword
		}

		tokens = append(tokens, Token{
			Type:  tokenType,
			Value: value,
			Position: Position{
				Column: match[0],
				Length: match[1] - match[0],
			},
		})
	}

	return tokens
}

// analyzeStructuralDifferences 分析结构差异
func (m *MyersDiffer) analyzeStructuralDifferences(oldTokens, newTokens []Token) []*SemanticDiff {
	var diffs []*SemanticDiff

	// 分析函数定义变化
	oldFuncs := m.extractFunctions(oldTokens)
	newFuncs := m.extractFunctions(newTokens)

	// 比较函数
	funcMap := make(map[string]Token)
	for _, token := range oldFuncs {
		funcMap[token.Value] = token
	}

	for _, newFunc := range newFuncs {
		if oldFunc, exists := funcMap[newFunc.Value]; exists {
			if m.hasSignificantChange(oldFunc, newFunc) {
				diffs = append(diffs, &SemanticDiff{
					Type:        SemanticDiffStructural,
					Description: fmt.Sprintf("函数 %s 的实现发生变化", newFunc.Value),
					OldValue:    oldFunc.Value,
					NewValue:    newFunc.Value,
					Position:    newFunc.Position,
					Confidence:  0.9,
				})
			}
			delete(funcMap, newFunc.Value)
		} else {
			diffs = append(diffs, &SemanticDiff{
				Type:        SemanticDiffStructural,
				Description: fmt.Sprintf("新增函数 %s", newFunc.Value),
				OldValue:    nil,
				NewValue:    newFunc.Value,
				Position:    newFunc.Position,
				Confidence:  1.0,
			})
		}
	}

	// 检查删除的函数
	for _, oldFunc := range funcMap {
		diffs = append(diffs, &SemanticDiff{
			Type:        SemanticDiffStructural,
			Description: fmt.Sprintf("删除函数 %s", oldFunc.Value),
			OldValue:    oldFunc.Value,
			NewValue:    nil,
			Position:    oldFunc.Position,
			Confidence:  1.0,
		})
	}

	return diffs
}

// extractFunctions 提取函数定义
func (m *MyersDiffer) extractFunctions(tokens []Token) []Token {
	var functions []Token
	for i := 0; i < len(tokens)-1; i++ {
		if tokens[i].Type == TokenTypeKeyword && tokens[i].Value == "func" {
			if i+1 < len(tokens) && tokens[i+1].Type == TokenTypeIdentifier {
				functions = append(functions, tokens[i+1])
			}
		}
	}
	return functions
}

// hasSignificantChange 检查是否有显著变化
func (m *MyersDiffer) hasSignificantChange(oldToken, newToken Token) bool {
	// 简化实现：如果位置变化很大，认为有显著变化
	return math.Abs(float64(oldToken.Position.Line-newToken.Position.Line)) > 5
}

// analyzeContentDifferences 分析内容差异
func (m *MyersDiffer) analyzeContentDifferences(oldLines, newLines []string, options *SemanticDiffOptions) []*SemanticDiff {
	var diffs []*SemanticDiff

	// 计算行级别的差异
	editDistance := m.computeLevenshteinDistance(strings.Join(oldLines, "\n"), strings.Join(newLines, "\n"))
	totalLength := max(len(oldLines), len(newLines))

	if totalLength > 0 {
		similarity := 1.0 - float64(editDistance)/float64(totalLength)
		if similarity < 0.8 {
			diffs = append(diffs, &SemanticDiff{
				Type:        SemanticDiffContent,
				Description: "文档内容有显著变化",
				OldValue:    fmt.Sprintf("%d lines", len(oldLines)),
				NewValue:    fmt.Sprintf("%d lines", len(newLines)),
				Confidence:  1.0 - similarity,
			})
		}
	}

	return diffs
}

// computeLevenshteinDistance 计算编辑距离
func (m *MyersDiffer) computeLevenshteinDistance(s1, s2 string) int {
	m, n := len(s1), len(s2)
	if m == 0 {
		return n
	}
	if n == 0 {
		return m
	}

	// 创建DP表
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	// 初始化
	for i := 0; i <= m; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= n; j++ {
		dp[0][j] = j
	}

	// 填充DP表
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if s1[i-1] == s2[j-1] {
				dp[i][j] = dp[i-1][j-1]
			} else {
				dp[i][j] = 1 + min(
					dp[i-1][j],   // 删除
					dp[i][j-1],   // 插入
					dp[i-1][j-1], // 替换
				)
			}
		}
	}

	return dp[m][n]
}

// analyzeFormattingDifferences 分析格式化差异
func (m *MyersDiffer) analyzeFormattingDifferences(oldContent, newContent string) []*SemanticDiff {
	var diffs []*SemanticDiff

	// 检查空白字符差异
	oldWhitespace := regexp.MustCompile(`\s+`).FindAllString(oldContent, -1)
	newWhitespace := regexp.MustCompile(`\s+`).FindAllString(newContent, -1)

	if len(oldWhitespace) != len(newWhitespace) {
		diffs = append(diffs, &SemanticDiff{
			Type:        SemanticDiffFormatting,
			Description: "空白字符格式发生变化",
			OldValue:    len(oldWhitespace),
			NewValue:    len(newWhitespace),
			Confidence:  0.7,
		})
	}

	// 检查行尾差异
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	oldHasTrailingWhitespace := false
	newHasTrailingWhitespace := false

	for _, line := range oldLines {
		if strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t") {
			oldHasTrailingWhitespace = true
			break
		}
	}

	for _, line := range newLines {
		if strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t") {
			newHasTrailingWhitespace = true
			break
		}
	}

	if oldHasTrailingWhitespace != newHasTrailingWhitespace {
		diffs = append(diffs, &SemanticDiff{
			Type:        SemanticDiffFormatting,
			Description: "行尾空白字符处理发生变化",
			OldValue:    oldHasTrailingWhitespace,
			NewValue:    newHasTrailingWhitespace,
			Confidence:  0.6,
		})
	}

	return diffs
}

// ============ 统一差异格式化器 ============

// UnifiedFormatter 统一差异格式化器
type UnifiedFormatter struct{}

// FormatUnified 格式化为统一差异格式
func (u *UnifiedFormatter) FormatUnified(blocks []*DiffBlock, options *UnifiedDiffOptions) (string, error) {
	var builder strings.Builder

	for _, block := range blocks {
		switch block.Type {
		case DiffBlockInsert:
			for _, line := range block.NewLines {
				if options.ShowLineNumbers {
					builder.WriteString(fmt.Sprintf("+%d: %s\n", block.NewStart, line))
				} else {
					builder.WriteString(fmt.Sprintf("+%s\n", line))
				}
			}

		case DiffBlockDelete:
			for _, line := range block.OldLines {
				if options.ShowLineNumbers {
					builder.WriteString(fmt.Sprintf("-%d: %s\n", block.OldStart, line))
				} else {
					builder.WriteString(fmt.Sprintf("-%s\n", line))
				}
			}

		case DiffBlockReplace:
			// 先显示删除的行
			for _, line := range block.OldLines {
				if options.ShowLineNumbers {
					builder.WriteString(fmt.Sprintf("-%d: %s\n", block.OldStart, line))
				} else {
					builder.WriteString(fmt.Sprintf("-%s\n", line))
				}
			}
			// 再显示新增的行
			for _, line := range block.NewLines {
				if options.ShowLineNumbers {
					builder.WriteString(fmt.Sprintf("+%d: %s\n", block.NewStart, line))
				} else {
					builder.WriteString(fmt.Sprintf("+%s\n", line))
				}
			}

		case DiffBlockEqual:
			if options.ContextLines > 0 {
				for i := 0; i < min(options.ContextLines, len(block.OldLines)); i++ {
					if options.ShowLineNumbers {
						builder.WriteString(fmt.Sprintf(" %d: %s\n", block.OldStart+i, block.OldLines[i]))
					} else {
						builder.WriteString(fmt.Sprintf(" %s\n", block.OldLines[i]))
					}
				}
			}
		}
	}

	return builder.String(), nil
}

// FormatHTML 格式化为HTML
func (u *UnifiedFormatter) FormatHTML(diffs []*DiffBlock, options *HTMLDiffOptions) (string, error) {
	var builder strings.Builder

	builder.WriteString("<div class=\"diff-container\">\n")

	for _, block := range diffs {
		switch block.Type {
		case DiffBlockInsert:
			for _, line := range block.NewLines {
				builder.WriteString(fmt.Sprintf(
					`<div class="diff-line added"><span class="line-number">%d</span><span class="line-content">%s</span></div>`,
					block.NewStart, html.EscapeString(line),
				))
			}

		case DiffBlockDelete:
			for _, line := range block.OldLines {
				builder.WriteString(fmt.Sprintf(
					`<div class="diff-line removed"><span class="line-number">%d</span><span class="line-content">%s</span></div>`,
					block.OldStart, html.EscapeString(line),
				))
			}

		case DiffBlockReplace:
			// 删除的行
			for _, line := range block.OldLines {
				builder.WriteString(fmt.Sprintf(
					`<div class="diff-line removed"><span class="line-number">%d</span><span class="line-content">%s</span></div>`,
					block.OldStart, html.EscapeString(line),
				))
			}
			// 新增的行
			for _, line := range block.NewLines {
				builder.WriteString(fmt.Sprintf(
					`<div class="diff-line added"><span class="line-number">%d</span><span class="line-content">%s</span></div>`,
					block.NewStart, html.EscapeString(line),
				))
			}

		case DiffBlockEqual:
			for _, line := range block.OldLines {
				builder.WriteString(fmt.Sprintf(
					`<div class="diff-line unchanged"><span class="line-number">%d</span><span class="line-content">%s</span></div>`,
					block.OldStart, html.EscapeString(line),
				))
			}
		}
	}

	builder.WriteString("</div>\n")
	return builder.String(), nil
}

// FormatSideBySide 格式化为并排对比
func (u *UnifiedFormatter) FormatSideBySide(diffs []*DiffBlock, options *SideBySideOptions) (string, error) {
	var builder strings.Builder

	builder.WriteString("<table class=\"diff-side-by-side\">\n")
	builder.WriteString("<thead><tr><th>旧版本</th><th>新版本</th></tr></thead>\n")
	builder.WriteString("<tbody>\n")

	for _, block := range diffs {
		switch block.Type {
		case DiffBlockInsert:
			for _, line := range block.NewLines {
				builder.WriteString(fmt.Sprintf(
					"<tr><td class=\"empty\"></td><td class=\"added\">%s</td></tr>",
					html.EscapeString(line),
				))
			}

		case DiffBlockDelete:
			for _, line := range block.OldLines {
				builder.WriteString(fmt.Sprintf(
					"<tr><td class=\"removed\">%s</td><td class=\"empty\"></td></tr>",
					html.EscapeString(line),
				))
			}

		case DiffBlockReplace:
			maxLines := max(len(block.OldLines), len(block.NewLines))
			for i := 0; i < maxLines; i++ {
				oldLine := ""
				newLine := ""
				oldClass := "empty"
				newClass := "empty"

				if i < len(block.OldLines) {
					oldLine = html.EscapeString(block.OldLines[i])
					oldClass = "removed"
				}
				if i < len(block.NewLines) {
					newLine = html.EscapeString(block.NewLines[i])
					newClass = "added"
				}

				builder.WriteString(fmt.Sprintf(
					"<tr><td class=\"%s\">%s</td><td class=\"%s\">%s</td></tr>",
					oldClass, oldLine, newClass, newLine,
				))
			}

		case DiffBlockEqual:
			for _, line := range block.OldLines {
				builder.WriteString(fmt.Sprintf(
					"<tr><td class=\"unchanged\">%s</td><td class=\"unchanged\">%s</td></tr>",
					html.EscapeString(line), html.EscapeString(line),
				))
			}
		}
	}

	builder.WriteString("</tbody>\n</table>\n")
	return builder.String(), nil
}

// FormatJSON 格式化为JSON
func (u *UnifiedFormatter) FormatJSON(diffs []*DiffBlock, options *JSONDiffOptions) (string, error) {
	var result struct {
		Blocks []map[string]interface{} `json:"blocks"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}

	for _, block := range diffs {
		blockData := map[string]interface{}{
			"type":       string(block.Type),
			"old_start":  block.OldStart,
			"new_start":  block.NewStart,
			"old_length": block.OldLength,
			"new_length": block.NewLength,
		}

		if len(block.OldLines) > 0 {
			blockData["old_lines"] = block.OldLines
		}
		if len(block.NewLines) > 0 {
			blockData["new_lines"] = block.NewLines
		}
		if len(block.Context) > 0 {
			blockData["context"] = block.Context
		}
		if block.Similarity > 0 {
			blockData["similarity"] = block.Similarity
		}

		result.Blocks = append(result.Blocks, blockData)
	}

	if options.IncludeMetadata {
		result.Metadata = map[string]interface{}{
			"total_blocks":   len(diffs),
			"generated_at":   time.Now().Format(time.RFC3339),
			"formatter":      "unified",
		}
	}

	// 简化的JSON序列化
	return fmt.Sprintf("%+v", result), nil
}

// ============ 并排差异格式化器 ============

// SideBySideFormatter 并排差异格式化器
type SideBySideFormatter struct{}

// FormatSideBySide 格式化为并排对比
func (s *SideBySideFormatter) FormatSideBySide(blocks []*DiffBlock, options *SideBySideOptions) (string, error) {
	formatter := &UnifiedFormatter{}
	return formatter.FormatSideBySide(blocks, &SideBySideOptions{
		ContextLines:     options.ContextLines,
		ShowLineNumbers:  options.ShowLineNumbers,
		HighlightChanges: options.HighlightChanges,
		MatchSimilarLines: options.MatchSimilarLines,
		OutputFormat:     options.OutputFormat,
	})
}

// ============ 智能冲突解决器 ============

// SmartConflictSolver 智能冲突解决器
type SmartConflictSolver struct {
	logger     *logrus.Logger
	strategies []ConflictResolutionStrategy
}

// ConflictResolutionStrategy 冲突解决策略接口
type ConflictResolutionStrategy interface {
	CanResolve(conflict *Conflict) bool
	Resolve(conflict *Conflict) (*ConflictResolution, error)
	Priority() int
}

// NewSmartConflictSolver 创建智能冲突解决器
func NewSmartConflictSolver(logger *logrus.Logger) ConflictSolver {
	solver := &SmartConflictSolver{
		logger: logger,
	}

	// 注册解决策略
	solver.strategies = []ConflictResolutionStrategy{
		&WhitespaceResolver{logger: logger},
		&ContentResolver{logger: logger},
		&LineEndingResolver{logger: logger},
		&FormattingResolver{logger: logger},
	}

	// 按优先级排序
	sort.Slice(solver.strategies, func(i, j int) bool {
		return solver.strategies[i].Priority() > solver.strategies[j].Priority()
	})

	return solver
}

// DetectConflicts 检测冲突
func (s *SmartConflictSolver) DetectConflicts(base, left, right string) ([]*Conflict, error) {
	baseLines := strings.Split(base, "\n")
	leftLines := strings.Split(left, "\n")
	rightLines := strings.Split(right, "\n")

	var conflicts []*Conflict

	// 使用三方合并算法检测冲突
	merger := &ThreeWayMerger{}
	mergeConflicts, err := merger.DetectConflicts(baseLines, leftLines, rightLines)
	if err != nil {
		return nil, fmt.Errorf("检测冲突失败: %w", err)
	}

	for _, conflict := range mergeConflicts {
		conflicts = append(conflicts, &Conflict{
			ID:          s.generateConflictID(),
			Type:        s.categorizeConflict(conflict),
			Position:    conflict.Position,
			BaseContent: conflict.BaseContent,
			LeftContent: conflict.LeftContent,
			RightContent: conflict.RightContent,
			Description: conflict.Description,
			CreatedAt:   time.Now(),
		})
	}

	return conflicts, nil
}

// AutoResolveConflicts 自动解决冲突
func (s *SmartConflictSolver) AutoResolveConflicts(conflicts []*Conflict) ([]*ConflictResolution, error) {
	var resolutions []*ConflictResolution

	for _, conflict := range conflicts {
		resolved := false

		// 尝试使用各种策略解决冲突
		for _, strategy := range s.strategies {
			if strategy.CanResolve(conflict) {
				resolution, err := strategy.Resolve(conflict)
				if err != nil {
					s.logger.WithError(err).Warn("解决冲突失败", "conflict_id", conflict.ID)
					continue
				}

				resolution.AutoResolved = true
				resolutions = append(resolutions, resolution)
				resolved = true
				break
			}
		}

		if !resolved {
			// 无法自动解决，返回手动解决建议
			suggestion, err := s.SuggestResolution(conflict)
			if err != nil {
				suggestion = "需要手动解决此冲突"
			}

			resolutions = append(resolutions, &ConflictResolution{
				Action:       ConflictActionManual,
				Content:      "",
				Reason:       suggestion,
				AutoResolved: false,
				Confidence:   0.0,
			})
		}
	}

	return resolutions, nil
}

// SuggestResolution 建议解决方案
func (s *SmartConflictSolver) SuggestResolution(conflict *Conflict) (string, error) {
	switch conflict.Type {
	case ConflictTypeWhitespace:
		return "建议保留更规范的空白字符格式", nil
	case ConflictTypeFormatting:
		return "建议保留更一致的代码格式", nil
	case ConflictTypeContent:
		return "建议仔细检查内容差异，选择合适的版本", nil
	case ConflictTypeLineEnding:
		return "建议统一使用LF或CRLF换行符", nil
	default:
		return "建议手动检查并选择合适的解决方案", nil
	}
}

// generateConflictID 生成冲突ID
func (s *SmartConflictSolver) generateConflictID() string {
	return fmt.Sprintf("conflict_%d", time.Now().UnixNano())
}

// categorizeConflict 分类冲突
func (s *SmartConflictSolver) categorizeConflict(conflict *MergeConflict) ConflictType {
	// 简化的冲突分类逻辑
	if s.isOnlyWhitespaceChange(conflict) {
		return ConflictTypeWhitespace
	}
	if s.isLineEndingChange(conflict) {
		return ConflictTypeLineEnding
	}
	if s.isFormattingChange(conflict) {
		return ConflictTypeFormatting
	}
	return ConflictTypeContent
}

// isOnlyWhitespaceChange 检查是否只是空白字符变化
func (s *SmartConflictSolver) isOnlyWhitespaceChange(conflict *MergeConflict) bool {
	baseTrimmed := strings.TrimSpace(conflict.BaseContent)
	leftTrimmed := strings.TrimSpace(conflict.LeftContent)
	rightTrimmed := strings.TrimSpace(conflict.RightContent)

	return baseTrimmed == leftTrimmed && baseTrimmed == rightTrimmed
}

// isLineEndingChange 检查是否只是换行符变化
func (s *SmartConflictSolver) isLineEndingChange(conflict *MergeConflict) bool {
	baseNormalized := strings.ReplaceAll(conflict.BaseContent, "\r\n", "\n")
	baseNormalized = strings.ReplaceAll(baseNormalized, "\r", "\n")

	leftNormalized := strings.ReplaceAll(conflict.LeftContent, "\r\n", "\n")
	leftNormalized = strings.ReplaceAll(leftNormalized, "\r", "\n")

	rightNormalized := strings.ReplaceAll(conflict.RightContent, "\r\n", "\n")
	rightNormalized = strings.ReplaceAll(rightNormalized, "\r", "\n")

	return baseNormalized == leftNormalized && baseNormalized == rightNormalized
}

// isFormattingChange 检查是否是格式化变化
func (s *SmartConflictSolver) isFormattingChange(conflict *MergeConflict) bool {
	// 简化实现：检查是否只有空格和标点的变化
	baseAlnum := s.countAlphanumeric(conflict.BaseContent)
	leftAlnum := s.countAlphanumeric(conflict.LeftContent)
	rightAlnum := s.countAlphanumeric(conflict.RightContent)

	return baseAlnum == leftAlnum && baseAlnum == rightAlnum
}

// countAlphanumeric 统计字母数字字符数量
func (s *SmartConflictSolver) countAlphanumeric(text string) int {
	count := 0
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			count++
		}
	}
	return count
}

// ============ 具体冲突解决策略 ============

// WhitespaceResolver 空白字符冲突解决器
type WhitespaceResolver struct {
	logger *logrus.Logger
}

func (w *WhitespaceResolver) CanResolve(conflict *Conflict) bool {
	return conflict.Type == ConflictTypeWhitespace
}

func (w *WhitespaceResolver) Resolve(conflict *Conflict) (*ConflictResolution, error) {
	// 选择更规范的空白字符格式
	leftNormalized := strings.Join(strings.Fields(conflict.LeftContent), " ")
	rightNormalized := strings.Join(strings.Fields(conflict.RightContent), " ")

	var chosenContent string
	var reason string

	if len(leftNormalized) > len(rightNormalized) {
		chosenContent = conflict.LeftContent
		reason = "选择更完整的空白字符格式"
	} else {
		chosenContent = conflict.RightContent
		reason = "选择更简洁的空白字符格式"
	}

	return &ConflictResolution{
		Action:       ConflictActionMerge,
		Content:      chosenContent,
		Reason:       reason,
		AutoResolved: true,
		Confidence:   0.9,
	}, nil
}

func (w *WhitespaceResolver) Priority() int {
	return 100
}

// ContentResolver 内容冲突解决器
type ContentResolver struct {
	logger *logrus.Logger
}

func (c *ContentResolver) CanResolve(conflict *Conflict) bool {
	return conflict.Type == ConflictTypeContent
}

func (c *ContentResolver) Resolve(conflict *Conflict) (*ConflictResolution, error) {
	// 简单的内容冲突解决策略
	if len(conflict.LeftContent) > len(conflict.RightContent) {
		return &ConflictResolution{
			Action:       ConflictActionAcceptLeft,
			Content:      conflict.LeftContent,
			Reason:       "选择更详细的内容",
			AutoResolved: true,
			Confidence:   0.6,
		}, nil
	}

	return &ConflictResolution{
		Action:       ConflictActionAcceptRight,
		Content:      conflict.RightContent,
		Reason:       "选择更新的内容",
		AutoResolved: true,
		Confidence:   0.6,
	}, nil
}

func (c *ContentResolver) Priority() int {
	return 50
}

// LineEndingResolver 换行符冲突解决器
type LineEndingResolver struct {
	logger *logrus.Logger
}

func (l *LineEndingResolver) CanResolve(conflict *Conflict) bool {
	return conflict.Type == ConflictTypeLineEnding
}

func (l *LineEndingResolver) Resolve(conflict *Conflict) (*ConflictResolution, error) {
	// 统一使用LF换行符
	content := strings.ReplaceAll(conflict.LeftContent, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	return &ConflictResolution{
		Action:       ConflictActionMerge,
		Content:      content,
		Reason:       "统一使用LF换行符",
		AutoResolved: true,
		Confidence:   0.95,
	}, nil
}

func (l *LineEndingResolver) Priority() int {
	return 90
}

// FormattingResolver 格式化冲突解决器
type FormattingResolver struct {
	logger *logrus.Logger
}

func (f *FormattingResolver) CanResolve(conflict *Conflict) bool {
	return conflict.Type == ConflictTypeFormatting
}

func (f *FormattingResolver) Resolve(conflict *Conflict) (*ConflictResolution, error) {
	// 选择更一致的格式
	leftConsistent := f.isConsistentFormatting(conflict.LeftContent)
	rightConsistent := f.isConsistentFormatting(conflict.RightContent)

	if leftConsistent && !rightConsistent {
		return &ConflictResolution{
			Action:       ConflictActionAcceptLeft,
			Content:      conflict.LeftContent,
			Reason:       "选择更一致的格式",
			AutoResolved: true,
			Confidence:   0.8,
		}, nil
	} else if !leftConsistent && rightConsistent {
		return &ConflictResolution{
			Action:       ConflictActionAcceptRight,
			Content:      conflict.RightContent,
			Reason:       "选择更一致的格式",
			AutoResolved: true,
			Confidence:   0.8,
		}, nil
	}

	// 如果都一致或都不一致，选择较新的版本
	return &ConflictResolution{
		Action:       ConflictActionAcceptRight,
		Content:      conflict.RightContent,
		Reason:       "选择较新的版本",
		AutoResolved: true,
		Confidence:   0.5,
	}, nil
}

func (f *FormattingResolver) isConsistentFormatting(content string) bool {
	// 简化的格式一致性检查
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return true
	}

	// 检查缩进是否一致
	var indentStyle string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		indent := f.getLineIndent(line)
		if indentStyle == "" {
			indentStyle = indent
		} else if indent != indentStyle && indent != "" {
			return false
		}
	}

	return true
}

func (f *FormattingResolver) getLineIndent(line string) string {
	var indent strings.Builder
	for _, r := range line {
		if r == ' ' || r == '\t' {
			indent.WriteRune(r)
		} else {
			break
		}
	}
	return indent.String()
}

func (f *FormattingResolver) Priority() int {
	return 70
}

// ============ 三方合并器 ============

// ThreeWayMerger 三方合并器
type ThreeWayMerger struct{}

// MergeConflict 合并冲突
type MergeConflict struct {
	Position    Position
	BaseContent string
	LeftContent string
	RightContent string
	Description string
}

// DetectConflicts 检测冲突
func (t *ThreeWayMerger) DetectConflicts(base, left, right []string) ([]*MergeConflict, error) {
	var conflicts []*MergeConflict

	// 简化的冲突检测
	maxLines := max(len(base), max(len(left), len(right)))

	for i := 0; i < maxLines; i++ {
		var baseLine, leftLine, rightLine string

		if i < len(base) {
			baseLine = base[i]
		}
		if i < len(left) {
			leftLine = left[i]
		}
		if i < len(right) {
			rightLine = right[i]
		}

		// 检测冲突
		if leftLine != rightLine && leftLine != baseLine && rightLine != baseLine {
			conflicts = append(conflicts, &MergeConflict{
				Position: Position{
					Line: i,
				},
				BaseContent: baseLine,
				LeftContent: leftLine,
				RightContent: rightLine,
				Description: fmt.Sprintf("第%d行存在冲突", i+1),
			})
		}
	}

	return conflicts, nil
}

// ============ 高级差异对比服务 ============

// AdvancedDiffService 高级差异对比服务
type AdvancedDiffService struct {
	differ     DiffEngine
	formatter  map[DiffOutputFormat]DiffFormatter
	resolver   ConflictSolver
	logger     *logrus.Logger
	cache      map[string]*DiffCache
	cacheMutex sync.RWMutex
}

// DiffCache 差异缓存
type DiffCache struct {
	Result    interface{}
	CreatedAt time.Time
	TTL       time.Duration
}

// NewAdvancedDiffService 创建高级差异对比服务
func NewAdvancedDiffService(logger *logrus.Logger) *AdvancedDiffService {
	service := &AdvancedDiffService{
		differ:    NewMyersDiffer(logger),
		formatter: make(map[DiffOutputFormat]DiffFormatter),
		resolver:  NewSmartConflictSolver(logger),
		logger:    logger,
		cache:     make(map[string]*DiffCache),
	}

	// 注册格式化器
	service.formatter[DiffFormatUnified] = &UnifiedFormatter{}
	service.formatter[DiffFormatHTML] = &UnifiedFormatter{}
	service.formatter[DiffFormatSideBySide] = &SideBySideFormatter{}
	service.formatter[DiffFormatJSON] = &UnifiedFormatter{}

	return service
}

// CompareVersions 比较版本
func (a *AdvancedDiffService) CompareVersions(ctx context.Context, oldContent, newContent string, format DiffOutputFormat) (string, error) {
	// 检查缓存
	cacheKey := a.generateCacheKey(oldContent, newContent, format)
	if cached := a.getFromCache(cacheKey); cached != nil {
		return cached.(string), nil
	}

	// 计算差异
	var result string
	var err error

	switch format {
	case DiffFormatUnified:
		options := &UnifiedDiffOptions{
			ContextLines:    3,
			ShowLineNumbers: true,
		}
		result, err = a.differ.ComputeUnifiedDiff(oldContent, newContent, options)

	case DiffFormatSideBySide:
		options := &SideBySideOptions{
			ContextLines:     3,
			ShowLineNumbers:  true,
			HighlightChanges: true,
		}
		result, err = a.differ.ComputeSideBySideDiff(oldContent, newContent, options)

	case DiffFormatHTML:
		blocks, diffErr := a.differ.ComputeDiff(oldContent, newContent)
		if diffErr != nil {
			return "", diffErr
		}
		options := &HTMLDiffOptions{
			ShowLineNumbers: true,
			SyntaxHighlight: true,
			TableLayout:     true,
		}
		formatter := a.formatter[DiffFormatHTML]
		result, err = formatter.FormatHTML(blocks, options)

	case DiffFormatJSON:
		blocks, diffErr := a.differ.ComputeDiff(oldContent, newContent)
		if diffErr != nil {
			return "", diffErr
		}
		options := &JSONDiffOptions{
			PrettyPrint:     true,
			IncludeMetadata: true,
		}
		formatter := a.formatter[DiffFormatJSON]
		result, err = formatter.FormatJSON(blocks, options)

	default:
		return "", fmt.Errorf("不支持的差异格式: %s", format)
	}

	if err != nil {
		return "", fmt.Errorf("计算差异失败: %w", err)
	}

	// 缓存结果
	a.putToCache(cacheKey, result, 10*time.Minute)

	return result, nil
}

// DetectAndResolveConflicts 检测并解决冲突
func (a *AdvancedDiffService) DetectAndResolveConflicts(ctx context.Context, base, left, right string) ([]*ConflictResolution, error) {
	// 检测冲突
	conflicts, err := a.resolver.DetectConflicts(base, left, right)
	if err != nil {
		return nil, fmt.Errorf("检测冲突失败: %w", err)
	}

	if len(conflicts) == 0 {
		return []*ConflictResolution{}, nil
	}

	// 自动解决冲突
	resolutions, err := a.resolver.AutoResolveConflicts(conflicts)
	if err != nil {
		return nil, fmt.Errorf("自动解决冲突失败: %w", err)
	}

	a.logger.WithFields(logrus.Fields{
		"total_conflicts":    len(conflicts),
		"auto_resolved":      len(resolutions),
		"manual_required":    a.countManualResolutions(resolutions),
	}).Info("冲突检测和解决完成")

	return resolutions, nil
}

// ComputeSemanticDiff 计算语义差异
func (a *AdvancedDiffService) ComputeSemanticDiff(ctx context.Context, oldContent, newContent string, language string) ([]*SemanticDiff, error) {
	options := &SemanticDiffOptions{
		Language:             language,
		IgnoreFormatting:     false,
		IgnoreComments:       false,
		NormalizeIdentifiers: false,
		ConfidenceThreshold:  0.7,
	}

	return a.differ.ComputeSemanticDiff(oldContent, newContent, options)
}

// GenerateDiffReport 生成差异报告
func (a *AdvancedDiffService) GenerateDiffReport(ctx context.Context, oldContent, newContent string) (*DiffReport, error) {
	blocks, err := a.differ.ComputeDiff(oldContent, newContent)
	if err != nil {
		return nil, fmt.Errorf("计算差异失败: %w", err)
	}

	report := &DiffReport{
		GeneratedAt: time.Now(),
		TotalBlocks: len(blocks),
		Summary:     a.generateSummary(blocks),
		Blocks:      blocks,
		Metadata:    a.generateMetadata(oldContent, newContent),
	}

	return report, nil
}

// DiffReport 差异报告
type DiffReport struct {
	GeneratedAt time.Time      `json:"generated_at"`
	TotalBlocks int             `json:"total_blocks"`
	Summary     DiffSummary     `json:"summary"`
	Blocks      []*DiffBlock    `json:"blocks"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// DiffSummary 差异摘要
type DiffSummary struct {
	Additions    int     `json:"additions"`
	Deletions    int     `json:"deletions"`
	Modifications int     `json:"modifications"`
	Similarity   float64 `json:"similarity"`
	Confidence   float64 `json:"confidence"`
}

// generateSummary 生成差异摘要
func (a *AdvancedDiffService) generateSummary(blocks []*DiffBlock) DiffSummary {
	var summary DiffSummary

	for _, block := range blocks {
		switch block.Type {
		case DiffBlockInsert:
			summary.Additions += block.NewLength
		case DiffBlockDelete:
			summary.Deletions += block.OldLength
		case DiffBlockReplace:
			summary.Modifications += max(block.OldLength, block.NewLength)
		}
	}

	totalChanges := summary.Additions + summary.Deletions + summary.Modifications
	if totalChanges > 0 {
		summary.Similarity = 1.0 - float64(summary.Deletions+summary.Modifications)/float64(totalChanges)
		summary.Confidence = math.Min(0.9, summary.Similarity)
	} else {
		summary.Similarity = 1.0
		summary.Confidence = 1.0
	}

	return summary
}

// generateMetadata 生成元数据
func (a *AdvancedDiffService) generateMetadata(oldContent, newContent string) map[string]interface{} {
	return map[string]interface{}{
		"old_size":    len(oldContent),
		"new_size":    len(newContent),
		"old_lines":   len(strings.Split(oldContent, "\n")),
		"new_lines":   len(strings.Split(newContent, "\n")),
		"algorithm":   "myers",
		"engine":      "advanced_diff",
	}
}

// 辅助方法

// generateCacheKey 生成缓存键
func (a *AdvancedDiffService) generateCacheKey(oldContent, newContent string, format DiffOutputFormat) string {
	return fmt.Sprintf("diff:%s:%x:%x", format, a.hashString(oldContent), a.hashString(newContent))
}

// hashString 计算字符串哈希
func (a *AdvancedDiffService) hashString(s string) string {
	// 简化的哈希实现
	return fmt.Sprintf("%x", len(s))
}

// getFromCache 从缓存获取
func (a *AdvancedDiffService) getFromCache(key string) interface{} {
	a.cacheMutex.RLock()
	defer a.cacheMutex.RUnlock()

	if cached, exists := a.cache[key]; exists {
		if time.Since(cached.CreatedAt) < cached.TTL {
			return cached.Result
		}
		delete(a.cache, key)
	}

	return nil
}

// putToCache 存入缓存
func (a *AdvancedDiffService) putToCache(key string, result interface{}, ttl time.Duration) {
	a.cacheMutex.Lock()
	defer a.cacheMutex.Unlock()

	a.cache[key] = &DiffCache{
		Result:    result,
		CreatedAt: time.Now(),
		TTL:       ttl,
	}
}

// countManualResolutions 统计需要手动解决的冲突数量
func (a *AdvancedDiffService) countManualResolutions(resolutions []*ConflictResolution) int {
	count := 0
	for _, resolution := range resolutions {
		if !resolution.AutoResolved {
			count++
		}
	}
	return count
}

// min 返回较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max 返回较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}