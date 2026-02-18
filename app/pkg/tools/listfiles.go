package tools

import (
	"encoding/json"
	"os"

	"github.com/codecrafters-io/claude-code-starter-go/app/pkg/registry"
)

type ListFilesParams struct {
	DirectoryPath string `json:"directoryPath"`
}

type ListFilesResponse struct {
	Name  string `json:"name"`
	IsDir bool   `json:"is_dir"`
}

func ListFiles(directorypath string) (string, error) {
	files, err := os.ReadDir(directorypath)

	if err != nil {
		return "", err
	}

	dirContents := make([]ListFilesResponse, len(files))

	for _, entry := range files {
		dirContents = append(dirContents, ListFilesResponse{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
		})
	}

	jsonData, err := json.Marshal(dirContents)

	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func RegisterListFilesTools(reg *registry.Registry) {
	reg.Register("ListFiles", func(args string) (string, error) {
		var params ListFilesParams
		if err := json.Unmarshal([]byte(args), &params); err != nil {
			return "", err
		}
		return ListFiles(params.DirectoryPath)
	})
}
