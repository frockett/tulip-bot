package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/codecrafters-io/claude-code-starter-go/app/pkg/registry"
	"github.com/codecrafters-io/claude-code-starter-go/app/pkg/tools"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func main() {
	var prompt string
	flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	flag.Parse()

	if prompt == "" {
		panic("Prompt must not be empty")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseUrl := os.Getenv("OPENROUTER_BASE_URL")
	if baseUrl == "" {
		baseUrl = "https://openrouter.ai/api/v1"
	}

	if apiKey == "" {
		panic("Env variable OPENROUTER_API_KEY not found")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseUrl))

	reg := registry.New()
	tools.RegisterBuiltinTools(reg)

	var messages []openai.ChatCompletionMessageParamUnion
	var toolsParams []openai.ChatCompletionToolUnionParam

	messages = append(messages, openai.ChatCompletionMessageParamUnion{
		OfUser: &openai.ChatCompletionUserMessageParam{
			Content: openai.ChatCompletionUserMessageParamContentUnion{
				OfString: openai.String(prompt),
			},
		}})

	toolsParams = tools.GetBuiltinToolDefinitions()

	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	// Main conversation loop - handle streaming and tool calls
	for {
		stream := client.Chat.Completions.NewStreaming(context.Background(),
			openai.ChatCompletionNewParams{
				Model:    openai.ChatModel("anthropic/claude-haiku-4.5"),
				Messages: messages,
				Tools:    toolsParams,
			},
		)

		acc := openai.ChatCompletionAccumulator{}
		var assistantContent string
		var toolCallsToExecute []map[string]any

		// Process the stream
		for stream.Next() {
			chunk := stream.Current()
			acc.AddChunk(chunk)

			// Print streaming content as it arrives
			// if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			// 	fmt.Print(chunk.Choices[0].Delta.Content)
			// }
			//
			if len(chunk.Choices) > 0 {

				// fmt.Fprintf(os.Stderr, "DEBUG: chunk content: %q\n", chunk.Choices[0].Delta.Content)

				// Manually accumulate content

				if chunk.Choices[0].Delta.Content != "" {

					assistantContent += chunk.Choices[0].Delta.Content

				}

			}

			// Detect when content finishes
			if content, ok := acc.JustFinishedContent(); ok {
				assistantContent = content
			}

			// Detect when a tool call finishes
			if tool, ok := acc.JustFinishedToolCall(); ok {
				// Track tool calls for later execution
				toolCallsToExecute = append(toolCallsToExecute, map[string]any{
					"id":   tool.ID,
					"name": tool.Name,
					"args": tool.Arguments,
				})
			}
		}

		if stream.Err() != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", stream.Err())
			os.Exit(1)
		}

		// // Debug: Check state before adding messages
		// fmt.Fprintf(os.Stderr, "\n[DEBUG] Messages before assistant: %d\n", len(messages))
		// fmt.Fprintf(os.Stderr, "[DEBUG] Tool calls to execute: %d\n", len(toolCallsToExecute))
		// for i, tc := range toolCallsToExecute {
		// 	fmt.Fprintf(os.Stderr, "[DEBUG] Tool call %d: ID=%s, Name=%s\n", i, tc["id"], tc["name"])
		// }

		// Build assistant message with tool calls
		assistantMsg := &openai.ChatCompletionAssistantMessageParam{
			Content: openai.ChatCompletionAssistantMessageParamContentUnion{
				OfString: openai.String(assistantContent),
			},
		}

		// Add tool calls to assistant message using extra fields
		if len(toolCallsToExecute) > 0 {
			// Format tool calls for the API structure
			formattedToolCalls := []map[string]any{}
			for _, tc := range toolCallsToExecute {
				formattedToolCalls = append(formattedToolCalls, map[string]any{
					"id":   tc["id"],
					"type": "function",
					"function": map[string]any{
						"name":      tc["name"],
						"arguments": tc["args"],
					},
				})
			}
			assistantMsg.SetExtraFields(map[string]any{
				"tool_calls": formattedToolCalls,
			})
		}

		// Add assistant message
		messages = append(messages, openai.ChatCompletionMessageParamUnion{
			OfAssistant: assistantMsg,
		})

		// Execute tools and add results
		for _, toolCall := range toolCallsToExecute {
			// fmt.Printf("Executing tool call: %s\n", toolCall["name"].(string))
			toolName := toolCall["name"].(string)
			toolID := toolCall["id"].(string)
			toolArgs := toolCall["args"].(string)

			// Execute the tool
			result, err := reg.Execute(toolName, toolArgs)

			if err != nil {
				// Handle error - add error result to messages
				messages = append(messages, openai.ChatCompletionMessageParamUnion{
					OfTool: &openai.ChatCompletionToolMessageParam{
						ToolCallID: toolID,
						Content: openai.ChatCompletionToolMessageParamContentUnion{
							OfString: openai.String(fmt.Sprintf("Error: %v", err))},
					}})
				continue
			}

			// Add tool result to messages
			// fmt.Fprintf(os.Stderr, "[DEBUG] Adding tool result for ID: %s\n", toolID)
			messages = append(messages, openai.ChatCompletionMessageParamUnion{
				OfTool: &openai.ChatCompletionToolMessageParam{
					ToolCallID: toolID,
					Content: openai.ChatCompletionToolMessageParamContentUnion{
						OfString: openai.String(result)},
				}})
		}

		if len(toolCallsToExecute) == 0 {
			fmt.Print(assistantContent)
		}

		// If there were tool calls, continue the conversation
		if len(toolCallsToExecute) > 0 {
			continue
		}

		// No more tool calls, we're done
		break
	}
}
