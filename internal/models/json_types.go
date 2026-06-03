package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSON 自定义JSON类型，用于处理GORM的JSON字段
type JSON map[string]interface{}

// Value 实现driver.Valuer接口
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现sql.Scanner接口
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSON)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return scanJSONMap(v, j)
	case string:
		return scanJSONMap([]byte(v), j)
	default:
		return fmt.Errorf("cannot scan %T into JSON", value)
	}
}

func scanJSONMap(data []byte, target *JSON) error {
	var object map[string]interface{}
	if err := json.Unmarshal(data, &object); err == nil {
		*target = JSON(object)
		return nil
	}

	var array []interface{}
	if err := json.Unmarshal(data, &array); err != nil {
		return err
	}
	*target = JSON{"uses": array}
	return nil
}

// String 返回JSON的字符串表示
func (j JSON) String() string {
	if j == nil {
		return "null"
	}
	data, _ := json.Marshal(j)
	return string(data)
}

// Get 获取JSON中的值
func (j JSON) Get(key string) interface{} {
	if j == nil {
		return nil
	}
	return j[key]
}

// Set 设置JSON中的值
func (j *JSON) Set(key string, value interface{}) {
	if *j == nil {
		*j = make(JSON)
	}
	(*j)[key] = value
}

// ToMap 转换为map[string]interface{}
func (j JSON) ToMap() map[string]interface{} {
	if j == nil {
		return make(map[string]interface{})
	}
	return map[string]interface{}(j)
}

// FromMap 从map[string]interface{}创建JSON
func FromMap(m map[string]interface{}) JSON {
	if m == nil {
		return make(JSON)
	}
	return JSON(m)
}

// JSONStringArray 自定义字符串数组JSON类型
type JSONStringArray []string

// Value 实现driver.Valuer接口
func (j JSONStringArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现sql.Scanner接口
func (j *JSONStringArray) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONStringArray, 0)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return fmt.Errorf("cannot scan %T into JSONStringArray", value)
	}
}

// String 返回JSONStringArray的字符串表示
func (j JSONStringArray) String() string {
	if j == nil {
		return "null"
	}
	data, _ := json.Marshal(j)
	return string(data)
}

// ToSlice 转换为[]string
func (j JSONStringArray) ToSlice() []string {
	if j == nil {
		return make([]string, 0)
	}
	return []string(j)
}

// FromSlice 从[]string创建JSONStringArray
func FromSlice(s []string) JSONStringArray {
	if s == nil {
		return make(JSONStringArray, 0)
	}
	return JSONStringArray(s)
}

// JSONArray JSON 数组类型，用于存储整数数组等
type JSONArray []interface{}

// Value 实现 driver.Valuer 接口
func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONArray, 0)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return fmt.Errorf("cannot scan %T into JSONArray", value)
	}
}

// ToIntSlice 将 JSONArray 转换为 []int
func (j JSONArray) ToIntSlice() []int {
	result := make([]int, 0, len(j))
	for _, item := range j {
		if num, ok := item.(float64); ok {
			result = append(result, int(num))
		} else if num, ok := item.(int); ok {
			result = append(result, num)
		}
	}
	return result
}

// FromIntSlice 从 []int 创建 JSONArray
func FromIntSlice(s []int) JSONArray {
	if s == nil {
		return make(JSONArray, 0)
	}
	result := make(JSONArray, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}
