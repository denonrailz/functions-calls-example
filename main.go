package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"os"
	"os/exec"
)

var (
	token = os.Getenv("OPENAI_API_KEY")

	systemPrompt = `You are a linux terminal commands explainer. User will provide you question, terminal command that he run and terminal output. You need to answer his question as short as possible. Is FORBIDDEN_COMMAND, means that user try to run command that can change something in system. You need to explain that, as assistant you cant execute commands that can modify something in system. Very shortly explain what can modify in system this command.
`

	systemFunctionPrompt = `I want you to act as a linux terminal commands assistant, who knows file system operations only. If users asks about adding, modifying or delete something, answer with FORBIDDEN_COMMAND with arguments of command and arguments.
For example:
Question: Delete all files in current directory
FunctionCall: {"command":"FORBIDDEN_COMMAND","args":"rm -rf *"}

Question: Move files from current directory to root
FunctionCall: {"command":"FORBIDDEN_COMMAND","args":"mv * /"}

Question: List all .txt files in current directory
FunctionCall: {"command":"ls","args":"*.txt"}
`

	question01     = "In which directory I am?"
	question02     = "Create new directory call test-folder"
	question03     = "in current folder list all files with it sizes. calculate total size of all files"
	questionsArray = []string{question01, question02, question03}

	promptWithTerminalOutput = `Answer user question, using provided terminal output
	Question: %v
	Command: %v
	Output: %v`
)

const (
	model  = openai.GPT4TurboPreview
	myFunc = `[{
    "name": "get_terminal_command_with_args",
	"description": "Provide linux/mac terminal command with arguments to run. If command is not view only and can modify something, return command: 'FORBIDDEN_COMMAND' with command and arguments.",
    "parameters": {
      "type": "object",
      "properties": {
        "command": {
          "type": "string",
           "description": "existing linux/mac terminal command"
        },
        "args": {
          "description": "array of arguments for running with commands",
          "type": "string"
        }
      },
      "required": [
        "command"
      ]
    }
  }]
`
)

// getFunction parse json string and return openai.Tool
func getFunction(funcToParse string) ([]openai.Tool, error) {
	var funcDef []openai.FunctionDefinition
	err := json.Unmarshal([]byte(funcToParse), &funcDef)
	if err != nil {
		return nil, fmt.Errorf("error unmarshal function json: %v", err)
	}

	var toolFunc []openai.Tool
	for _, f := range funcDef {
		toolFunc = append(toolFunc, openai.Tool{
			Type:     openai.ToolTypeFunction,
			Function: f,
		})
	}
	return toolFunc, nil
}

func main() {
	//Initialisation. Reading env variables OPENAI_API_KEY and OpenAI Client
	if token == "" {
		fmt.Println("OPENAI_API_KEY environment variable is not set")
		os.Exit(1)
	}
	client := openai.NewClient(token)

	//Parsing GPT function
	gptFunction, err := getFunction(myFunc)
	if err != nil {
		fmt.Printf("error parsing function: %v\n", err)
		os.Exit(1)
	}

	//Setting up function call
	chatFunctionRequest := openai.ChatCompletionRequest{
		Model:       model,
		Tools:       gptFunction,
		ToolChoice:  gptFunction[0],
		Temperature: 0,
	}

	//Create function chat completion request
	chatRequest := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
		},
	}

	fmt.Printf("\nSYSTEM CHAT PROMPT: %v\n", systemPrompt)
	fmt.Printf("SYSTEM FUNCTION PROMPT: %v\n", systemFunctionPrompt)
	fmt.Printf("*****\n\n")

	//let's iterate over our question
	for i, userQuestion := range questionsArray {

		fmt.Printf("Question %v. %v\n", i+1, userQuestion)
		//Switching user question
		//chatFunctionRequest.Messages[1].Content = userQuestion
		chatFunctionRequest.Messages = []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemFunctionPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userQuestion,
			},
		}

		//Calling GPT function
		fmt.Printf("function call in progress...\n")
		funcResp, err := client.CreateChatCompletion(context.Background(), chatFunctionRequest)
		if err != nil {
			fmt.Printf("Question %v. ChatCompletion error: %v\n", i, err)
			os.Exit(1)
		}
		fmt.Printf("assistant: '%v'\n", funcResp.Choices[0].Message.Content)
		fmt.Printf("function call: %+v\n", funcResp.Choices[0].Message.ToolCalls[0].Function)

		//handling function response
		//binding function response to map
		var cmdMap map[string]string
		err = json.Unmarshal([]byte(funcResp.Choices[0].Message.ToolCalls[0].Function.Arguments), &cmdMap)
		if err != nil {
			fmt.Printf("error unmarshal function json: %v\n", err)
			os.Exit(1)
		}

		//running our functions called from GPT
		getExecOutput, err := execute(cmdMap["command"], cmdMap["args"])
		if err != nil {
			fmt.Printf("error execute command: %v\n", err)
			os.Exit(1)
		}

		//putting execution output to user question, and asking for providing answer
		queryWithOutput := fmt.Sprintf(
			promptWithTerminalOutput,
			userQuestion,
			fmt.Sprintf("%v %v", cmdMap["command"], cmdMap["args"]),
			getExecOutput)
		fmt.Printf("\nGPTChat call starts...\n")
		fmt.Printf("user: %v\n", queryWithOutput)

		chatRequest.Messages = append(chatRequest.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: queryWithOutput,
		})
		chatResp, err := client.CreateChatCompletion(context.Background(), chatRequest)
		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("response from assistant: %v\n------\n", chatResp.Choices[0].Message.Content)
		//adding response to chatRequest messages
		chatRequest.Messages = append(chatRequest.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: chatResp.Choices[0].Message.Content,
		})
	}
}

func execute(command, params string) (string, error) {
	var (
		cmd *exec.Cmd
	)
	if command == "FORBIDDEN_COMMAND" {
		return "FORBIDDEN_COMMAND", nil
	}
	if params != "" {
		cmd = exec.Command(command, params)
	} else {
		cmd = exec.Command(command)
	}
	output, err := cmd.Output()
	return string(output), err
}
