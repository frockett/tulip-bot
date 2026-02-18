package tools

import (
	"github.com/codecrafters-io/claude-code-starter-go/app/pkg/registry"
	"github.com/openai/openai-go/v3"
)

func RegisterBuiltinTools(reg *registry.Registry) {
	RegisterReadTools(reg)
	RegisterListFilesTools(reg)
	RegisterWriteTools(reg)
}

func GetBuiltinToolDefinitions() []openai.ChatCompletionToolUnionParam {
	return []openai.ChatCompletionToolUnionParam{
		openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        "Read",
			Description: openai.String("Read and return the contents of a file"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"filePath": map[string]any{
						"type":        "string",
						"description": "The path of the file to read",
					},
				},
			},
		}),
		openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        "Write",
			Description: openai.String("Write something to a file"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"filePath": map[string]any{
						"type":        "string",
						"description": "The path of the file to write",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "The content to write to the file",
					},
				},
			},
		}),
		openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        "ListFiles",
			Description: openai.String("List files in a directory"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"directoryPath": map[string]any{
						"type":        "string",
						"description": "The path of the directory in which to list the files",
					},
				},
			},
		}),
	}
}
