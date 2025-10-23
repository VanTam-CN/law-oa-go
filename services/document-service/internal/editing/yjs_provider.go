package editing

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"law-oa-go/internal/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// YjsProviderImpl Yjs提供者实现
type YjsProviderImpl struct {
	documents map[string]*YjsDocumentImpl
	mutex     sync.RWMutex
	logger    *logrus.Logger
}

// YjsDocumentImpl Yjs文档实现
type YjsDocumentImpl struct {
	id     string
	texts  map[string]*YjsTextImpl
	mutex  sync.RWMutex
	logger *logrus.Logger
}

// YjsTextImpl Yjs文本实现
type YjsTextImpl struct {
	id        string
	content   string
	delta     *models.DeltaData
	length    int
	listeners []func(interface{})
	mutex     sync.RWMutex
	logger    *logrus.Logger
}

// YjsChangeEvent Yjs变更事件
type YjsChangeEvent struct {
	Type      string      `json:"type"`
	Document  string      `json:"document"`
	Text      string      `json:"text"`
	Operation interface{} `json:"operation"`
	Timestamp time.Time   `json:"timestamp"`
	UserID    string      `json:"user_id"`
}

// NewYjsProvider 创建Yjs提供者
func NewYjsProvider() *YjsProviderImpl {
	return &YjsProviderImpl{
		documents: make(map[string]*YjsDocumentImpl),
		logger:    logrus.New(),
	}
}

// Initialize 初始化Yjs文档
func (p *YjsProviderImpl) Initialize(documentID string) (YjsDocument, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 检查文档是否已存在
	if doc, exists := p.documents[documentID]; exists {
		return doc, nil
	}

	// 创建新文档
	doc := &YjsDocumentImpl{
		id:     documentID,
		texts:  make(map[string]*YjsTextImpl),
		logger: p.logger,
	}
	doc.texts["content"] = &YjsTextImpl{
		id:        uuid.New().String(),
		content:   "",
		delta:     &models.DeltaData{Ops: []interface{}{}},
		length:    0,
		listeners: make([]func(interface{}), 0),
		logger:    p.logger,
	}

	p.documents[documentID] = doc

	p.logger.WithField("document_id", documentID).Info("初始化Yjs文档")

	return doc, nil
}

// GetText 获取Yjs文本
func (p *YjsProviderImpl) GetText(name string) YjsText {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	// 这里简化实现，总是返回第一个文档的content文本
	if len(p.documents) == 0 {
		return nil
	}

	for _, doc := range p.documents {
		if text, exists := doc.texts[name]; exists {
			return text
		}
	}

	return nil
}

// Subscribe 订阅文档变更
func (p *YjsProviderImpl) Subscribe(callback func(delta interface{})) error {
	// 这里简化实现，实际中需要更复杂的事件系统
	p.logger.Info("订阅Yjs文档变更事件")
	return nil
}

// Unsubscribe 取消订阅
func (p *YjsProviderImpl) Unsubscribe() error {
	p.logger.Info("取消Yjs文档订阅")
	return nil
}

// Destroy 销毁Yjs提供者
func (p *YjsProviderImpl) Destroy() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, doc := range p.documents {
		doc.Destroy()
	}

	p.documents = make(map[string]*YjsDocumentImpl)
	p.logger.Info("销毁Yjs提供者")

	return nil
}

// GetDocument 获取文档（内部方法）
func (p *YjsProviderImpl) GetDocument(documentID string) *YjsDocumentImpl {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	return p.documents[documentID]
}

// GetID 获取文档ID
func (d *YjsDocumentImpl) GetID() string {
	return d.id
}

// GetText 获取文本
func (d *YjsDocumentImpl) GetText(name string) YjsText {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	return d.texts[name]
}

// Destroy 销毁文档
func (d *YjsDocumentImpl) Destroy() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	for _, text := range d.texts {
		text.Destroy()
	}

	d.texts = make(map[string]*YjsTextImpl)
	d.logger.WithField("document_id", d.id).Info("销毁Yjs文档")

	return nil
}

// GetID 获取文本ID
func (t *YjsTextImpl) GetID() string {
	return t.id
}

// Insert 插入文本
func (t *YjsTextImpl) Insert(index int, text string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// 简化实现：在指定位置插入文本
	if index < 0 || index > t.length {
		index = t.length
	}

	t.content = t.content[:index] + text + t.content[index:]
	t.length += len(text)

	// 更新Delta表示
	t.delta = &models.DeltaData{
		Ops: []interface{}{
			map[string]interface{}{
				"insert": text,
			},
		},
	}

	// 通知监听者
	t.notifyListeners()
}

// Delete 删除文本
func (t *YjsTextImpl) Delete(index int, length int) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// 简化实现：删除指定范围的文本
	if index < 0 || index >= t.length || length <= 0 {
		return
	}

	if index+length > t.length {
		length = t.length - index
	}

	t.content = t.content[:index] + t.content[index+length:]
	t.length -= length

	// 更新Delta表示
	t.delta = &models.DeltaData{
		Ops: []interface{}{
			map[string]interface{}{
				"delete": float64(length),
			},
		},
	}

	// 通知监听者
	t.notifyListeners()
}

// Format 格式化文本
func (t *YjsTextImpl) Format(index int, length int, attributes map[string]interface{}) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// 简化实现：应用格式到指定范围
	if index < 0 || index >= t.length || length <= 0 {
		return
	}

	if index+length > t.length {
		length = t.length - index
	}

	// 这里简化处理，实际中需要更复杂的格式管理
	// 只是记录格式信息，实际的格式化在转换时处理

	// 通知监听者
	t.notifyListeners()
}

// GetDelta 获取Delta
func (t *YjsTextImpl) GetDelta() *models.DeltaData {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.delta
}

// GetLength 获取长度
func (t *YjsTextImpl) GetLength() int {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.length
}

// Observe 观察变更
func (t *YjsTextImpl) Observe(callback func(event interface{})) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.listeners = append(t.listeners, callback)
}

// Unobserve 取消观察
func (t *YjsTextImpl) Unobserve() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.listeners = make([]func(interface{}), 0)
}

// notifyListeners 通知监听者
func (t *YjsTextImpl) notifyListeners() {
	// 简化实现：通知所有监听器
	event := YjsChangeEvent{
		Type:      "text-change",
		Operation: t.delta,
		Timestamp: time.Now(),
	}

	for _, listener := range t.listeners {
		listener(event)
	}
}

// GetContent 获取内容（内部方法）
func (t *YjsTextImpl) GetContent() string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	return t.content
}

// UpdateContent 更新内容（内部方法）
func (t *YjsTextImpl) UpdateContent(content string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.content = content
	t.length = len(content)

	// 重新生成Delta
	t.delta = &models.DeltaData{
		Ops: []interface{}{
			map[string]interface{}{
				"insert": content,
			},
		},
	}

	t.notifyListeners()
}

// GetYjsProviderForRoom 获取房间Yjs提供者（用于协作）
func GetYjsProviderForRoom(roomID string) YjsProvider {
	provider := NewYjsProvider()
	// 这里可以初始化房间特定的配置
	return provider
}

// CreateYjsDocument 创建Yjs文档（辅助函数）
func CreateYjsDocument(documentID string, initialContent string) YjsDocument {
	doc := &YjsDocumentImpl{
		id:     documentID,
		texts:  make(map[string]*YjsTextImpl),
		logger: logrus.New(),
	}

	// 创建初始文本
	text := &YjsTextImpl{
		id:        uuid.New().String(),
		content:   initialContent,
		delta: &models.DeltaData{
			Ops: []interface{}{
				map[string]interface{}{
					"insert": initialContent,
				},
			},
		},
		length:    len(initialContent),
		listeners: make([]func(interface{}), 0),
		logger:    logrus.New(),
	}

	doc.texts["content"] = text

	return doc
}