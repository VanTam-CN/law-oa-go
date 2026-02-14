//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"law-oa-go/internal/cache"
	"law-oa-go/internal/config"
	"law-oa-go/internal/database"
)

func main() {
	fmt.Println("=== 缓存服务测试 ===")

	// 1. 测试配置加载
	fmt.Println("1. 加载配置...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("配置加载失败:", err)
	}
	fmt.Println("配置加载成功")

	// 2. 测试数据库模块初始化
	fmt.Println("2. 初始化数据库模块...")
	if err := database.InitOptimizedComponents(cfg); err != nil {
		log.Fatal("数据库模块初始化失败:", err)
	}
	fmt.Println("数据库模块初始化成功")

	// 3. 测试获取缓存服务并模拟main.go逻辑
	fmt.Println("3. 获取缓存服务并模拟main.go逻辑...")
	cacheService := database.GetCacheService()
	fmt.Printf("数据库模块缓存服务: %p\n", cacheService)
	fmt.Printf("全局缓存服务 (前): %p\n", cache.DefaultCacheService)
	
	// 模拟main.go的逻辑
	if cacheService != nil {
		cache.DefaultCacheService = cacheService
		fmt.Println("使用数据库模块的缓存服务")
	} else {
		fmt.Println("数据库模块缓存服务不可用，尝试初始化独立缓存服务...")
		
		// 初始化独立缓存服务
		if err := cache.InitCache(); err != nil {
			log.Fatal("缓存服务初始化失败:", err)
		}
		cacheService = cache.DefaultCacheService
	}
	
	fmt.Printf("全局缓存服务 (后): %p\n", cache.DefaultCacheService)
	
	if cacheService == nil {
		log.Fatal("缓存服务仍然为 nil")
	}
	fmt.Println("缓存服务获取成功")

	// 4. 测试缓存连接
	fmt.Println("4. 测试缓存连接...")
	if cacheService.Client() == nil {
		log.Fatal("缓存客户端为 nil")
	}
	
	if !cacheService.Ping() {
		log.Fatal("缓存连接测试失败")
	}
	fmt.Println("缓存连接测试成功")

	// 5. 测试缓存操作
	fmt.Println("5. 测试缓存操作...")
	ctx := context.Background()
	testKey := "test:key"
	testData := map[string]interface{}{
		"id":    1,
		"name":  "测试数据",
		"value": "这是一个缓存测试",
	}

	// 测试设置
	if err := cacheService.Set(ctx, testKey, testData, time.Minute); err != nil {
		log.Fatal("缓存设置失败:", err)
	}
	fmt.Println("缓存设置成功")

	// 测试获取
	var result map[string]interface{}
	if err := cacheService.Get(ctx, testKey, &result); err != nil {
		log.Fatal("缓存获取失败:", err)
	}
	fmt.Printf("缓存获取成功: %+v\n", result)

	// 测试删除
	if err := cacheService.Delete(ctx, testKey); err != nil {
		log.Fatal("缓存删除失败:", err)
	}
	fmt.Println("缓存删除成功")

	// 6. 测试全局缓存服务
	fmt.Println("6. 测试全局缓存服务...")
	if cache.DefaultCacheService == nil {
		log.Fatal("全局缓存服务为 nil")
	}
	
	// 测试全局缓存服务操作
	globalTestKey := "global:test"
	globalTestData := "全局测试数据"
	
	if err := cache.DefaultCacheService.Set(ctx, globalTestKey, globalTestData, time.Minute); err != nil {
		log.Fatal("全局缓存设置失败:", err)
	}
	
	var globalResult string
	if err := cache.DefaultCacheService.Get(ctx, globalTestKey, &globalResult); err != nil {
		log.Fatal("全局缓存获取失败:", err)
	}
	fmt.Printf("全局缓存测试成功: %s\n", globalResult)

	fmt.Println("=== 缓存服务测试完成 ===")
}