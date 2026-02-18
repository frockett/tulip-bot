package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/codecrafters-io/claude-code-starter-go/app/pkg/registry"
	"github.com/codecrafters-io/claude-code-starter-go/app/pkg/tools"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type UserInstructions struct {
	FileName     string
	Instructions string
}

func loadUserInstructions(instructions *UserInstructions) (err error) {

	files := []string{"claude.md", "gemini.md", "instructions.md"}

	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			instFile, err := os.ReadFile(file)
			if err != nil {
				log.Fatal(err)
			}

			instructions.Instructions = string(instFile)
			instructions.FileName = file
			return nil
		}
	}
	return fmt.Errorf("no instructions file found")
}

func main() {
	// var prompt string
	// flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	// flag.Parse()

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

	// Initiate the messages array and get the tool definitions
	var messages []openai.ChatCompletionMessageParamUnion
	var toolsParams []openai.ChatCompletionToolUnionParam
	toolsParams = tools.GetBuiltinToolDefinitions()

	userInstructions := UserInstructions{}
	err := loadUserInstructions(&userInstructions)
	if err != nil {
		log.Fatal(err)
	}
	if userInstructions.FileName != "" {
		fmt.Println("Loading user instructions from", userInstructions.FileName)
		messages = append(messages, openai.ChatCompletionMessageParamUnion{
			OfSystem: &openai.ChatCompletionSystemMessageParam{
				Content: openai.ChatCompletionSystemMessageParamContentUnion{
					OfString: openai.String(userInstructions.Instructions),
				},
			},
		})
	}

	reader := bufio.NewReader(os.Stdin)

	for {

		fmt.Print("> ")

		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if strings.HasPrefix(line, "/") {
			if line == "/exit\n" {
				fmt.Println("Exiting...")
				os.Exit(0)
			}
			if line == "/help\n" {
				fmt.Println("Available commands:")
				fmt.Println("/exit - Exit the program")
				fmt.Println("/help - Show this help message")
				continue
			}
			if line == "/instructions\n" {
				fmt.Println("Current instructions:")
				fmt.Println(userInstructions.Instructions)
				continue
			}
			fmt.Println("Unknown command:", line)
			continue
		}

		messages = append(messages, openai.ChatCompletionMessageParamUnion{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.String(line),
				},
			}})

		// fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")
		//

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

			for stream.Next() {
				chunk := stream.Current()
				acc.AddChunk(chunk)

				// Print streaming content as it arrives
				if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
					fmt.Print(chunk.Choices[0].Delta.Content)
				}

				if content, ok := acc.JustFinishedContent(); ok {
					assistantContent = content
				}

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

			// New line at end of agent stream
			fmt.Print("\n")

			// No more tool calls, we're done
			break
		}
	}

}
