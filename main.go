package main

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"os"
)

var (
	token      = os.Getenv("OPENAI_API_KEY")
	system     = "I want you to act as a linux terminal commands assistant. You knows only file system operations only. You can't provide answer with commands that can modify or delete files. Only view. Do not write explanations. Just write commands."
	question01 = "In which directory I am?"
	question02 = "Create new directory call test-folder"
	question03 = "Provide information about system?"
	question04 = "List all files and provide size of all files in current folder"
)

const (
	model = openai.GPT4TurboPreview
)

func main() {
	//Reading env variables OPENAI_API_KEY and OPENAI_ORG_ID
	if token == "" {
		fmt.Println("OPENAI_API_KEY environment variable is not set")
		os.Exit(1)
	}

	//Init OpenAI client
	client := openai.NewClient(token)

	//Create chat completion request
	chatRequest := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: system,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: question01,
			},
		},
	}

	fmt.Printf("\nsystem: %v\n\n", chatRequest.Messages[0].Content)
	fmt.Printf("user: %v\n\n", chatRequest.Messages[1].Content)

	//Send request to OpenAI API
	resp, err := client.CreateChatCompletion(context.Background(), chatRequest)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}
	//append GPT answer to messages
	fmt.Printf("assistant: %v\n-----\n", resp.Choices[0].Message.Content)
	chatRequest.Messages = append(chatRequest.Messages, resp.Choices[0].Message)

	//Adding second question to ChatGPT messages list
	fmt.Printf("user: %v\n\n", question02)
	chatRequest.Messages = append(chatRequest.Messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question02,
	})
	resp, err = client.CreateChatCompletion(context.Background(), chatRequest)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("assistant: %v\n-----\n", resp.Choices[0].Message.Content)
	chatRequest.Messages = append(chatRequest.Messages, resp.Choices[0].Message)

	//Third question to ChatGPT
	fmt.Printf("user: %v\n\n", question03)
	chatRequest.Messages = append(chatRequest.Messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question03,
	})
	resp, err = client.CreateChatCompletion(context.Background(), chatRequest)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("assistant: %v\n-----\n", resp.Choices[0].Message.Content)
	chatRequest.Messages = append(chatRequest.Messages, resp.Choices[0].Message)

	//Fourth question to ChatGPT
	fmt.Printf("user: %v\n\n", question04)
	chatRequest.Messages = append(chatRequest.Messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: question04,
	})

	resp, err = client.CreateChatCompletion(context.Background(), chatRequest)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("assistant: %v\n-----\n", resp.Choices[0].Message.Content)
	chatRequest.Messages = append(chatRequest.Messages, resp.Choices[0].Message)
}
