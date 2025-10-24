package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// 文档管理系统测试运行器
func main() {
	fmt.Println("🚀 文档管理系统测试套件启动")
	fmt.Println(strings.Repeat("=", 50))

	startTime := time.Now()

	// 运行单元测试
	fmt.Println("📋 运行单元测试...")
	if err := runUnitTests(); err != nil {
		log.Printf("❌ 单元测试失败: %v", err)
		os.Exit(1)
	}

	fmt.Println("✅ 单元测试完成")

	// 运行集成测试
	fmt.Println("\n🔗 运行集成测试...")
	if err := runIntegrationTests(); err != nil {
		log.Printf("❌ 集成测试失败: %v", err)
		os.Exit(1)
	}

	fmt.Println("✅ 集成测试完成")

	// 运行性能测试
	fmt.Println("\n⚡ 运行性能测试...")
	if err := runBenchmarkTests(); err != nil {
		log.Printf("❌ 性能测试失败: %v", err)
		os.Exit(1)
	}

	fmt.Println("✅ 性能测试完成")

	// 生成测试报告
	fmt.Println("\n📊 生成测试报告...")
	if err := generateTestReport(); err != nil {
		log.Printf("⚠️ 测试报告生成失败: %v", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("\n" + strings.Repeat("=", 50) + "\n")
	fmt.Printf("✅ 所有测试完成! 耗时: %v\n", duration)
	fmt.Println(strings.Repeat("=", 50))
}

func runUnitTests() error {
	fmt.Println("   - 文档上传服务测试...")
	if err := runTest("TestDocumentService"); err != nil {
		return err
	}

	fmt.Println("   - 文档预览服务测试...")
	if err := runTest("TestDocumentPreviewService"); err != nil {
		return err
	}

	fmt.Println("   - 文档版本管理测试...")
	if err := runTest("TestDocumentVersionService"); err != nil {
		return err
	}

	fmt.Println("   - 文档权限控制测试...")
	if err := runTest("TestDocumentPermissionService"); err != nil {
		return err
	}

	fmt.Println("   - 文档回收站测试...")
	if err := runTest("TestDocumentRecycleService"); err != nil {
		return err
	}

	fmt.Println("   - 文档搜索功能测试...")
	if err := runTest("TestDocumentSearchService"); err != nil {
		return err
	}

	fmt.Println("   - 文档统计功能测试...")
	if err := runTest("TestDocumentStatsService"); err != nil {
		return err
	}

	return nil
}

func runIntegrationTests() error {
	fmt.Println("   - 文档生命周期集成测试...")
	if err := runTest("TestIntegration"); err != nil {
		return err
	}
	return nil
}

func runBenchmarkTests() error {
	fmt.Println("   - 搜索性能测试...")
	if err := runBenchmark("BenchmarkSearchDocument"); err != nil {
		return err
	}

	fmt.Println("   - 统计性能测试...")
	if err := runBenchmark("BenchmarkDocumentStats"); err != nil {
		return err
	}
	return nil
}

func runTest(testName string) error {
	cmd := exec.Command("go", "test", "-v", "-run", "^"+testName+"$", "./internal/services")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func runBenchmark(benchmarkName string) error {
	cmd := exec.Command("go", "test", "-bench", "^"+benchmarkName+"$", "-benchmem", "./internal/services")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func generateTestReport() error {
	reportDir := "test-reports"
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return fmt.Errorf("failed to create test reports directory: %w", err)
	}

	// 生成覆盖率报告
	fmt.Println("   - 生成覆盖率报告...")
	cmd := exec.Command("go", "test", "-coverprofile=coverage.out", "-covermode=atomic", "./internal/services")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate coverage: %w", err)
	}

	// 生成HTML覆盖率报告
	cmd = exec.Command("go", "tool", "cover", "-html=coverage.out", "-o", filepath.Join(reportDir, "coverage.html"))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate HTML coverage: %w", err)
	}

	// 读取覆盖率数据
	cmd = exec.Command("go", "tool", "cover", "-func=coverage.out")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to read coverage data: %w", err)
	}

	// 解析覆盖率
	lines := strings.Split(string(output), "\n")
	var totalCoverage string
	for _, line := range lines {
		if strings.Contains(line, "total:") {
			totalCoverage = strings.TrimSpace(strings.Split(line, ":")[1])
			break
		}
	}

	// 创建测试报告文件
	reportContent := fmt.Sprintf(`# 文档管理系统测试报告

## 测试概览
- 测试时间: %s
- 总体覆盖率: %s

## 功能测试覆盖

### 核心服务测试
- [x] 文档上传服务 (DocumentService)
- [x] 文档预览服务 (DocumentPreviewService)
- [x] 文档版本管理 (DocumentVersionService)
- [x] 文档权限控制 (DocumentPermissionService)
- [x] 文档回收站 (DocumentRecycleService)
- [x] 文档搜索功能 (DocumentSearchService)
- [x] 文档统计功能 (DocumentStatsService)

### 集成测试
- [x] 文档生命周期集成测试

### 性能测试
- [x] 文档搜索性能基准
- [x] 统计功能性能基准

## 测试统计

### 功能覆盖
- ✅ 文档CRUD操作
- ✅ 文件上传和下载
- ✅ 多格式文档预览
- ✅ 版本管理和比较
- ✅ 权限控制和分享
- ✅ 回收站和恢复
- ✅ 高级搜索和过滤
- ✅ 统计报表和导出

### 测试环境
- Go版本: %s
- 测试框架: Go标准testing包
- 覆盖率工具: go test -cover

## 结论
文档管理系统的所有核心功能都通过了完整的单元测试、集成测试和性能测试，确保了代码质量和系统稳定性。
`,
		time.Now().Format("2006-01-02 15:04:05"),
		totalCoverage,
		getGoVersion(),
	)

	reportPath := filepath.Join(reportDir, "test-report.md")
	if err := os.WriteFile(reportPath, []byte(reportContent), 0644); err != nil {
		return fmt.Errorf("failed to write test report: %w", err)
	}

	fmt.Printf("   ✅ 测试报告已生成: %s\n", reportPath)
	fmt.Printf("   ✅ 覆盖率报告已生成: %s\n", filepath.Join(reportDir, "coverage.html"))
	fmt.Printf("   ✅ 总体覆盖率: %s\n", totalCoverage)

	return nil
}

func getGoVersion() string {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}