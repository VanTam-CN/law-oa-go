package common

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// StreamResponse 流式响应
type StreamResponse struct {
	Data     chan interface{} `json:"-"`
	Done     chan bool        `json:"-"`
	Metadata interface{}      `json:"metadata,omitempty"`
}

// NewStreamResponse 创建新的流式响应
func NewStreamResponse() *StreamResponse {
	return &StreamResponse{
		Data: make(chan interface{}, 100),
		Done: make(chan bool),
	}
}

// Send 发送数据
func (sr *StreamResponse) Send(data interface{}) {
	select {
	case sr.Data <- data:
	case <-sr.Done:
	}
}

// Close 关闭流式响应
func (sr *StreamResponse) Close() {
	close(sr.Done)
	close(sr.Data)
}

// WriteContentType 实现gin.Render接口
func (sr *StreamResponse) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Accel-Buffering", "no")
}

// Render 实现gin.Render接口
func (sr *StreamResponse) Render(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Accel-Buffering", "no") // 禁用Nginx缓冲
	
	// 写入响应头
	w.Write([]byte(`{"data": [`))
	
	first := true
	encoder := json.NewEncoder(w)
	
	for data := range sr.Data {
		if !first {
			w.Write([]byte(","))
		}
		first = false
		
		if err := encoder.Encode(data); err != nil {
			return err
		}
		
		// 立即刷新缓冲区
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
	
	w.Write([]byte(`]}`))
	
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	
	return nil
}

// StreamSuccess 返回流式成功响应
func StreamSuccess(c *gin.Context, dataChan <-chan interface{}) {
	response := NewStreamResponse()
	defer response.Close()
	
	// 启动goroutine转发数据
	go func() {
		for data := range dataChan {
			response.Send(data)
		}
	}()
	
	c.Render(http.StatusOK, response)
}

// StreamPaginatedResponse 流式分页响应
type StreamPaginatedResponse struct {
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	Data     chan interface{} `json:"-"`
	Done     chan bool      `json:"-"`
}

// NewStreamPaginatedResponse 创建新的流式分页响应
func NewStreamPaginatedResponse(total int64, page, pageSize int) *StreamPaginatedResponse {
	return &StreamPaginatedResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     make(chan interface{}, 100),
		Done:     make(chan bool),
	}
}

// Send 发送数据
func (sr *StreamPaginatedResponse) Send(data interface{}) {
	select {
	case sr.Data <- data:
	case <-sr.Done:
	}
}

// Close 关闭流式响应
func (sr *StreamPaginatedResponse) Close() {
	close(sr.Done)
	close(sr.Data)
}

// WriteContentType 实现gin.Render接口
func (sr *StreamPaginatedResponse) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Accel-Buffering", "no")
}

// Render 实现gin.Render接口
func (sr *StreamPaginatedResponse) Render(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("X-Accel-Buffering", "no")
	
	// 写入响应头
	header := map[string]interface{}{
		"total":     sr.Total,
		"page":      sr.Page,
		"page_size": sr.PageSize,
		"data":      []interface{}{},
	}
	
	headerData, err := json.Marshal(header)
	if err != nil {
		return err
	}
	
	w.Write(headerData[:len(headerData)-2]) // 去掉最后的"]}
	w.Write([]byte(`,"data":[`))
	
	first := true
	encoder := json.NewEncoder(w)
	
	for data := range sr.Data {
		if !first {
			w.Write([]byte(","))
		}
		first = false
		
		if err := encoder.Encode(data); err != nil {
			return err
		}
		
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
	
	w.Write([]byte(`]}}`))
	
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
	
	return nil
}

// StreamPaginatedSuccess 返回流式分页成功响应
func StreamPaginatedSuccess(c *gin.Context, total int64, page, pageSize int, dataChan <-chan interface{}) {
	response := NewStreamPaginatedResponse(total, page, pageSize)
	defer response.Close()
	
	go func() {
		for data := range dataChan {
			response.Send(data)
		}
	}()
	
	c.Render(http.StatusOK, response)
}