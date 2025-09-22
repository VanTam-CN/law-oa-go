package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

func main() {
	// 创建Elasticsearch客户端
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
	})
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s", err)
	}

	// 测试连接
	res, err := es.Info()
	if err != nil {
		log.Fatalf("Error connecting to Elasticsearch: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Fatalf("Elasticsearch returned error: %s", res.Status())
	}

	fmt.Println("Elasticsearch连接成功")

	// 创建包含completion字段的索引模板
	createIndexTemplate(es)

	// 等待索引创建完成
	time.Sleep(2 * time.Second)

	// 准备测试数据
	documents := []map[string]interface{}{
		{
			"id":          1,
			"title":       "合同模板-标准版",
			"description": "标准合同模板，适用于一般业务场景",
			"type":        "case",
			"entity_id":   "1",
			"status":      "active",
			"created_at":  time.Now().Format(time.RFC3339),
			"suggest":     []string{"合同模板", "标准版", "合同", "模板", "标准合同"},
		},
		{
			"id":          2,
			"title":       "客户身份证复印件",
			"description": "客户身份证扫描件",
			"type":        "client",
			"entity_id":   "1",
			"status":      "active",
			"created_at":  time.Now().Format(time.RFC3339),
			"suggest":     []string{"客户身份证", "身份证", "客户", "复印件", "身份证扫描"},
		},
		{
			"id":          3,
			"title":       "法律意见书",
			"description": "关于XX案件的法律意见",
			"type":        "case",
			"entity_id":   "2",
			"status":      "active",
			"created_at":  time.Now().Format(time.RFC3339),
			"suggest":     []string{"法律意见书", "法律意见", "意见书", "法律", "案件"},
		},
		{
			"id":          4,
			"title":       "证据清单",
			"description": "案件证据材料清单",
			"type":        "case",
			"entity_id":   "2",
			"status":      "active",
			"created_at":  time.Now().Format(time.RFC3339),
			"suggest":     []string{"证据清单", "证据", "清单", "案件证据", "材料清单"},
		},
		{
			"id":          5,
			"title":       "委托协议",
			"description": "与客户签订的委托协议",
			"type":        "case",
			"entity_id":   "3",
			"status":      "active",
			"created_at":  time.Now().Format(time.RFC3339),
			"suggest":     []string{"委托协议", "委托", "协议", "客户委托", "签订协议"},
		},
		{
			"id":          6,
			"title":       "案件相关照片",
			"description": "案件现场照片",
			"type":        "case",
			"entity_id":   "3",
			"status":      "active",
			"created_at":  time.Now().Format(time.RFC3339),
			"suggest":     []string{"案件照片", "现场照片", "案件", "照片", "现场"},
		},
		{
			"id":          7,
			"title":       "法院传票",
			"description": "法院送达的传票文件",
			"type":        "case",
			"entity_id":   "4",
			"status":      "active",
			"created_at":  time.Now().Format(time.RFC3339),
			"suggest":     []string{"法院传票", "传票", "法院", "送达", "传票文件"},
		},
	}

	// 批量索引文档
	for _, doc := range documents {
		indexName := "law_oa_" + doc["type"].(string)
		docID := fmt.Sprintf("%v", doc["id"])
		err := indexDocument(es, indexName, docID, doc)
		if err != nil {
			log.Printf("Error indexing document %s: %v", docID, err)
		} else {
			log.Printf("Successfully indexed document %s", docID)
		}
	}

	// 等待索引完成
	time.Sleep(2 * time.Second)

	// 测试搜索建议功能
	testSuggestions(es)

	fmt.Println("搜索建议测试数据创建完成")
}

// 创建索引模板
func createIndexTemplate(es *elasticsearch.Client) {
	template := map[string]interface{}{
		"index_patterns": []string{"law_oa_*"},
		"template": map[string]interface{}{
			"settings": map[string]interface{}{
				"number_of_shards":   1,
				"number_of_replicas": 1,
			},
			"mappings": map[string]interface{}{
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type": "long",
					},
					"title": map[string]interface{}{
						"type": "text",
						"analyzer": "standard",
					},
					"description": map[string]interface{}{
						"type": "text",
						"analyzer": "standard",
					},
					"type": map[string]interface{}{
						"type": "keyword",
					},
					"entity_id": map[string]interface{}{
						"type": "keyword",
					},
					"status": map[string]interface{}{
						"type": "keyword",
					},
					"created_at": map[string]interface{}{
						"type": "date",
					},
					"suggest": map[string]interface{}{
						"type": "completion",
						"analyzer": "standard",
						"search_analyzer": "standard",
					},
				},
			},
		},
	}

	templateJSON, err := json.Marshal(template)
	if err != nil {
		log.Fatalf("Error marshaling template: %v", err)
	}

	req := es.Indices.PutIndexTemplate
	res, err := req(
		"law_oa_template",
		strings.NewReader(string(templateJSON)),
	)
	if err != nil {
		log.Fatalf("Error creating index template: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("Error response from Elasticsearch: %s", res.Status())
		body, _ := io.ReadAll(res.Body)
		log.Printf("Error body: %s", string(body))
		return
	}

	fmt.Println("索引模板创建成功")
}

// 索引文档
func indexDocument(es *elasticsearch.Client, indexName, docID string, doc map[string]interface{}) error {
	docJSON, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	req := es.Index
	res, err := req(
		indexName,
		strings.NewReader(string(docJSON)),
		req.WithDocumentID(docID),
		req.WithRefresh("true"),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("indexing error: %s", res.Status())
	}

	return nil
}

// 测试搜索建议功能
func testSuggestions(es *elasticsearch.Client) {
	fmt.Println("\n=== 测试搜索建议功能 ===")

	// 测试1: 搜索"法律"
	fmt.Println("\n测试1: 搜索 '法律'")
	suggestions := getSuggestions(es, "法律")
	fmt.Printf("建议结果: %v\n", suggestions)

	// 测试2: 搜索"合同"
	fmt.Println("\n测试2: 搜索 '合同'")
	suggestions = getSuggestions(es, "合同")
	fmt.Printf("建议结果: %v\n", suggestions)

	// 测试3: 搜索"客户"
	fmt.Println("\n测试3: 搜索 '客户'")
	suggestions = getSuggestions(es, "客户")
	fmt.Printf("建议结果: %v\n", suggestions)

	// 测试4: 搜索"协议"
	fmt.Println("\n测试4: 搜索 '协议'")
	suggestions = getSuggestions(es, "协议")
	fmt.Printf("建议结果: %v\n", suggestions)
}

// 获取搜索建议
func getSuggestions(es *elasticsearch.Client, query string) []string {
	suggestQuery := map[string]interface{}{
		"suggest": map[string]interface{}{
			"text": query,
			"completion": map[string]interface{}{
				"field": "suggest",
				"size":  10,
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(suggestQuery); err != nil {
		log.Printf("Error encoding suggest query: %v", err)
		return []string{}
	}

	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("law_oa_*"),
		es.Search.WithBody(&buf),
	)
	if err != nil {
		log.Printf("Error executing suggest query: %v", err)
		return []string{}
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("Suggest query error: %s", res.Status())
		return []string{}
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.Printf("Error decoding suggest response: %v", err)
		return []string{}
	}

	return parseSuggestions(result, 10)
}

// 解析建议结果
func parseSuggestions(result map[string]interface{}, limit int) []string {
	suggest, ok := result["suggest"].(map[string]interface{})
	if !ok {
		return []string{}
	}

	var suggestions []string
	for _, suggestField := range suggest {
		if suggestList, ok := suggestField.([]interface{}); ok && len(suggestList) > 0 {
			if firstSuggest, ok := suggestList[0].(map[string]interface{}); ok {
				if options, ok := firstSuggest["options"].([]interface{}); ok {
					for _, option := range options {
						if optMap, ok := option.(map[string]interface{}); ok {
							if text, ok := optMap["text"].(string); ok {
								suggestions = append(suggestions, text)
								if len(suggestions) >= limit {
									return suggestions
								}
							}
						}
					}
				}
			}
		}
	}

	return suggestions
}