package elasticsearch

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

// Client Elasticsearch客户端封装
type Client struct {
	client *elasticsearch.Client
}

// NewClient 创建Elasticsearch客户端。
//
// 该包目前仅保留给历史兼容代码使用，因此不再依赖应用级 Config。
func NewClient(host, port, username, password string) (*Client, error) {
	host = strings.TrimSpace(host)
	port = strings.TrimSpace(port)
	if host == "" || port == "" {
		return nil, fmt.Errorf("elasticsearch host and port must be configured")
	}

	esConfig := elasticsearch.Config{
		Addresses: []string{fmt.Sprintf("http://%s:%s", host, port)},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// 如果有用户名密码，添加认证
	if username != "" && password != "" {
		esConfig.Username = username
		esConfig.Password = password
	}

	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, err
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.Info(
		client.Info.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}

// IndexDocument 索引文档
func (c *Client) IndexDocument(ctx context.Context, index string, id string, body map[string]interface{}) error {
	res, err := c.client.Index(
		index,
		strings.NewReader(toJSON(body)),
		c.client.Index.WithContext(ctx),
		c.client.Index.WithDocumentID(id),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index error: %s", res.Status())
	}

	return nil
}

// Search 执行搜索
func (c *Client) Search(ctx context.Context, index string, query map[string]interface{}) (map[string]interface{}, error) {
	res, err := c.client.Search(
		c.client.Search.WithContext(ctx),
		c.client.Search.WithIndex(index),
		c.client.Search.WithBody(strings.NewReader(toJSON(query))),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.Status())
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteDocument 删除文档
func (c *Client) DeleteDocument(ctx context.Context, index string, id string) error {
	res, err := c.client.Delete(
		index,
		id,
		c.client.Delete.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("delete error: %s", res.Status())
	}

	return nil
}

// CreateIndex 创建索引
func (c *Client) CreateIndex(ctx context.Context, index string, mapping map[string]interface{}) error {
	body := map[string]interface{}{
		"mappings": mapping,
	}

	res, err := c.client.Indices.Create(
		index,
		c.client.Indices.Create.WithContext(ctx),
		c.client.Indices.Create.WithBody(strings.NewReader(toJSON(body))),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("create index error: %s", res.Status())
	}

	return nil
}

// DeleteIndex 删除索引
func (c *Client) DeleteIndex(ctx context.Context, index string) error {
	res, err := c.client.Indices.Delete(
		[]string{index},
		c.client.Indices.Delete.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("delete index error: %s", res.Status())
	}

	return nil
}

// IndexExists 检查索引是否存在
func (c *Client) IndexExists(ctx context.Context, index string) (bool, error) {
	res, err := c.client.Indices.Exists(
		[]string{index},
		c.client.Indices.Exists.WithContext(ctx),
	)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		return true, nil
	}
	if res.StatusCode == 404 {
		return false, nil
	}

	return false, fmt.Errorf("index exists check error: %s", res.Status())
}

// GetIndexMapping 获取索引映射
func (c *Client) GetIndexMapping(ctx context.Context, index string) (map[string]interface{}, error) {
	res, err := c.client.Indices.GetMapping(
		c.client.Indices.GetMapping.WithContext(ctx),
		c.client.Indices.GetMapping.WithIndex(index),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("get mapping error: %s", res.Status())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetClient 获取原始ES客户端
func (c *Client) GetClient() *elasticsearch.Client {
	return c.client
}

// GetRawClient 获取原始ES客户端（与database包兼容）
func (c *Client) GetRawClient() *elasticsearch.Client {
	return c.client
}

// Helper function to convert map to JSON string
func toJSON(data map[string]interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return `{"query": {"match_all": {}}}`
	}
	return string(jsonData)
}
