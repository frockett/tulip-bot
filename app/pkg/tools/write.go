package tools

import (
	"encoding/json"
	"os"

	"github.com/codecrafters-io/claude-code-starter-go/app/pkg/registry"
)

type WriteParams struct {
	FilePath string `json:"filePath"`
	Content  string `json:"content"`
}

func Write(filePath string, content string) (string, error) {
	err := os.WriteFile(filePath, []byte(content), 0644)

	if err != nil {
		return "", err
	}

	return "success", nil
}

func RegisterWriteTools(reg *registry.Registry) {
	reg.Register("Write", func(args string) (string, error) {
		var params WriteParams
		if err := json.Unmarshal([]byte(args), &params); err != nil {
			return "", err
		}
		return Write(params.FilePath, params.Content)
	})
}
