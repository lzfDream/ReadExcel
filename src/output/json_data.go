package output

import (
	"encoding/json"
	"fmt"
	"os"
)

func OutputJson(path, fileName string, data map[string]interface{}) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	jsonData = append(jsonData, '\n')
	err = os.WriteFile(path+"/"+fileName+".json", jsonData, 0644)
	if err != nil {
		return fmt.Errorf("写入文件%s失败: %w", fileName, err)
	}

	return nil
}
