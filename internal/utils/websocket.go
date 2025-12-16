package utils

import (
	"encoding/json"
)

func BindData(data interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, target)
}

func GetReceiverId(data interface{}) string {
	temp := make(map[string]interface{})
	jsonData, _ := json.Marshal(data)
	json.Unmarshal(jsonData, &temp)

	if val, ok := temp["receiver_id"].(string); ok {
		return val
	}
	return ""
}