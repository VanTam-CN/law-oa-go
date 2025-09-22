//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"law-oa-go/internal/config"
)

// 测试数据
var testDocuments = []map[string]interface{}{
	{
		"title":     "合同纠纷案例",
		"content":   "这是一个关于合同纠纷的法律案例，涉及两方当事人的合同履行问题。",
		"type":      "case",
		"client":    "张三",
		"status":    "active",
		"priority":  "high",
		"created_at": "2024-01-15T10:30:00Z",
		"tags":      []string{"合同", "纠纷", "商业"},
	},
	{
		"title":     "劳动争议调解",
		"content":   "员工与公司之间的劳动争议，涉及工资支付和劳动合同问题。",
		"type":      "case",
		"client":    "李四",
		"status":    "pending",
		"priority":  "medium",
		"created_at": "2024-02-20T14:15:00Z",
		"tags":      []string{"劳动", "争议", "调解"},
	},
	{
		"title":     "知识产权侵权",
		"content":   "商标侵权案件，原告指控被告未经授权使用其注册商标。",
		"type":      "case",
		"client":    "王五",
		"status":    "closed",
		"priority":  "high",
		"created_at": "2024-03-10T09:45:00Z",
		"tags":      []string{"知识产权", "商标", "侵权"},
	},
	{
		"title":     "客户资料：ABC科技公司",
		"content":   "ABC科技公司是一家专注于人工智能开发的高科技企业，需要法律服务支持。",
		"type":      "client",
		"company":   "ABC科技有限公司",
		"status":    "active",
		"industry":  "科技",
		"created_at": "2024-01-05T11:20:00Z",
		"tags":      []string{"客户", "科技公司", "AI"},
	},
	{
		"title":     "客户资料：XYZ制造企业",
		"content":   "XYZ制造企业主要生产精密机械零件，面临供应链合同问题。",
		"type":      "client",
		"company":   "XYZ制造有限公司",
		"status":    "active",
		"industry":  "制造业",
		"created_at": "2024-02-12T16:30:00Z",
		"tags":      []string{"客户", "制造业", "供应链"},
	},
	{
		"title":     "保密协议模板",
		"content":   "标准保密协议模板，适用于商业合作中的信息保护需求。",
		"type":      "document",
		"category":  "合同模板",
		"file_type": "docx",
		"size":      1024000,
		"created_at": "2024-01-01T00:00:00Z",
		"tags":      []string{"模板", "保密协议", "合同"},
	},
	{
		"title":     "股权转让协议",
		"content":   "公司股权转让协议，包含转让条件、价格条款和交割安排。",
		"type":      "document",
		"category":  "合同模板",
		"file_type": "pdf",
		"size":      2048000,
		"created_at": "2024-02-15T12:00:00Z",
		"tags":      []string{"模板", "股权", "转让"},
	},
}

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 初始化Elasticsearch客户端
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{cfg.GetElasticsearchURL()},
		Username:  cfg.Elasticsearch.Username,
		Password:  cfg.Elasticsearch.Password,
	})
	if err != nil {
		log.Fatal("连接Elasticsearch失败:", err)
	}

	// 测试连接
	res, err := es.Info()
	if err != nil {
		log.Fatal("获取Elasticsearch信息失败:", err)
	}
	res.Body.Close()

	fmt.Println("Elasticsearch连接成功")

	// 创建索引
	indexName := "lawoa_test"
	if err := createIndex(es, indexName); err != nil {
		log.Printf("创建索引失败: %v", err)
	}

	// 索引测试数据
	for i, doc := range testDocuments {
		if err := indexDocument(es, indexName, fmt.Sprintf("doc_%d", i+1), doc); err != nil {
			log.Printf("索引文档 %d 失败: %v", i+1, err)
		} else {
			fmt.Printf("成功索引文档: %s\n", doc["title"])
		}
	}

	fmt.Println("测试数据索引完成")

	// 等待索引刷新
	time.Sleep(2 * time.Second)

	// 测试搜索
	testSearch(es, indexName)
}

func createIndex(es *elasticsearch.Client, indexName string) error {
	// 删除现有索引（如果存在）
	es.Indices.Delete([]string{indexName})

	// 创建索引映射
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
					"boost":    3.0,
				},
				"content": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
					"boost":    1.0,
				},
				"type": map[string]interface{}{
					"type": "keyword",
				},
				"client": map[string]interface{}{
					"type": "keyword",
				},
				"status": map[string]interface{}{
					"type": "keyword",
				},
				"priority": map[string]interface{}{
					"type": "keyword",
				},
				"company": map[string]interface{}{
					"type": "text",
				},
				"industry": map[string]interface{}{
					"type": "keyword",
				},
				"category": map[string]interface{}{
					"type": "keyword",
				},
				"file_type": map[string]interface{}{
					"type": "keyword",
				},
				"size": map[string]interface{}{
					"type": "long",
				},
				"created_at": map[string]interface{}{
					"type": "date",
				},
				"tags": map[string]interface{}{
					"type": "keyword",
				},
				"suggest": map[string]interface{}{
					"type": "completion",
				},
			},
		},
	}

	res, err := es.Indices.Create(indexName, es.Indices.Create.WithBody(strings.NewReader(toJSON(mapping))))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("创建索引失败: %s", res.Status())
	}

	fmt.Printf("索引 %s 创建成功\n", indexName)
	return nil
}

func indexDocument(es *elasticsearch.Client, indexName, docID string, doc map[string]interface{}) error {
	// 添加suggest字段用于自动完成
	if title, ok := doc["title"].(string); ok {
			doc["suggest"] = map[string]interface{}{
				"input":  []string{title},
				"weight": 10,
			}
	}

	res, err := es.Index(indexName, strings.NewReader(toJSON(doc)), es.Index.WithDocumentID(docID))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("索引文档失败: %s", res.Status())
	}

	return nil
}

func testSearch(es *elasticsearch.Client, indexName string) {
	fmt.Println("\n=== 测试搜索功能 ===")

	// 测试关键词搜索
	testQueries := []string{
		"合同",
		"纠纷", 
		"客户",
		"模板",
		"知识产权",
		"科技公司",
	}

	for _, query := range testQueries {
		fmt.Printf("\n搜索关键词: '%s'\n", query)
		
		searchQuery := map[string]interface{}{
			"query": map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query": query,
					"fields": []string{"title^3", "content"},
				},
			},
			"highlight": map[string]interface{}{
				"fields": map[string]interface{}{
					"title":   map[string]interface{}{},
					"content": map[string]interface{}{},
				},
				"pre_tags":  []string{"<em>"},
				"post_tags": []string{"</em>"},
			},
		}

		res, err := es.Search(
			es.Search.WithIndex(indexName),
			es.Search.WithBody(strings.NewReader(toJSON(searchQuery))),
		)
		if err != nil {
			log.Printf("搜索失败: %v", err)
			continue
		}

		var result map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			log.Printf("解析搜索结果失败: %v", err)
			res.Body.Close()
			continue
		}
		res.Body.Close()

		hits := result["hits"].(map[string]interface{})
		total := hits["total"].(map[string]interface{})
		totalValue := total["value"].(float64)

		fmt.Printf("找到 %d 个结果\n", int(totalValue))

		if hitsList, ok := hits["hits"].([]interface{}); ok {
			for i, hit := range hitsList {
				if i >= 3 { // 只显示前3个结果
					break
				}
				if hitMap, ok := hit.(map[string]interface{}); ok {
					source := hitMap["_source"].(map[string]interface{})
					title := source["title"].(string)
					docType := source["type"].(string)
					fmt.Printf("  %d. [%s] %s\n", i+1, docType, title)
				}
			}
		}
	}
}

// toJSON 辅助函数
func toJSON(data map[string]interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(jsonData)
}
