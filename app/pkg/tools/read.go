package tools

import (
	"encoding/json"
	"os"

	"github.com/codecrafters-io/claude-code-starter-go/app/pkg/registry"
)

type ReadParams struct {
	FilePath string `json:"filePath"`
}

func Read(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func RegisterReadTools(reg *registry.Registry) {
	reg.Register("Read", func(args string) (string, error) {
		var params ReadParams
		if err := json.Unmarshal([]byte(args), &params); err != nil {
			return "", err
		}
		return Read(params.FilePath)
	})
}
