package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type CustomerRequest struct {
	RequestText string `json:"requestText"`
}

type GatewayResponse struct {
	ResponseText string `json:"responseText,omitempty"`
	Error        string `json:"error,omitempty"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env file: %+v", err)
	}

	http.HandleFunc("/gateway", gatewayHandler)

	log.Println("Gateway service running on :8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func gatewayHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	var customerRequest CustomerRequest
	err := json.NewDecoder(r.Body).Decode(&customerRequest)
	if err != nil || customerRequest.RequestText == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	responseText, err := processWithOpenAI(customerRequest.RequestText)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(GatewayResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GatewayResponse{ResponseText: responseText})
}

func processWithOpenAI(query string) (string, error) {
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	)

	ctx := context.Background()
	tools := []openai.ChatCompletionToolParam{
		{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(openai.FunctionDefinitionParam{
				Name:        openai.String("get_balances"),
				Description: openai.String("Retrieve user balances"),
			}),
		},
		{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(openai.FunctionDefinitionParam{
				Name:        openai.String("get_transactions"),
				Description: openai.String("Retrieve user transactions"),
			}),
		},
		{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(openai.FunctionDefinitionParam{
				Name:        openai.String("get_exchange_rates"),
				Description: openai.String("Retrieve exchange rates for a given symbol"),
				Parameters: openai.F(openai.FunctionParameters{
					"type": "object",
					"properties": map[string]interface{}{
						"symbol": map[string]string{
							"type":        "string",
							"description": "The trading symbol, e.g., BTC/USDT, ETH/USD, SOL/GEL",
						},
					},
					"required": []string{"symbol"},
				}),
			}),
		},
	}

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(query),
	}

	log.Println("messages", messages)

	params := openai.ChatCompletionNewParams{
		Messages: openai.F(messages),
		Tools:    openai.F(tools),
		Model:    openai.F(openai.ChatModelGPT4oMini),
	}

	completion, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %v", err)
	}

	toolCalls := completion.Choices[0].Message.ToolCalls

	if len(toolCalls) == 0 {
		fmt.Printf("No function call\n")
		return completion.Choices[0].Message.Content, nil
	}

	params.Messages.Value = append(params.Messages.Value, completion.Choices[0].Message)
	for _, toolCall := range toolCalls {
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			panic(err)
		}

		// TODO: make this dynamic
		args["userId"] = "1"

		data, err := callToolboxAPI(toolCall.Function.Name, args)
		if err != nil {
			panic(err)
		}

		params.Messages.Value = append(params.Messages.Value, openai.ToolMessage(toolCall.ID, data))
	}

	completion, err = client.Chat.Completions.New(ctx, params)
	if err != nil {
		panic(err)
	}

	log.Println(completion.Choices[0].Message.Content)

	return completion.Choices[0].Message.Content, nil
}

func callToolboxAPI(toolName string, params map[string]interface{}) (string, error) {
	endpoints := map[string]string{
		"get_balances":       "/balances",
		"get_transactions":   "/transactions",
		"get_exchange_rates": "/exchange-rates",
	}

	endpoint, exists := endpoints[toolName]
	if !exists {
		return "", fmt.Errorf("unknown tool: %s", toolName)
	}

	baseURL := os.Getenv("TOOLBOX_API_BASE_URL") + endpoint

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		log.Fatal(err)
	}

	urlParams := url.Values{}
	for k, v := range params {
		urlParams.Add(k, v.(string))
	}

	parsedURL.RawQuery = urlParams.Encode()

	log.Println("parsedURL.String()", parsedURL.String())

	resp, err := http.Get(parsedURL.String())
	if err != nil {
		return "", fmt.Errorf("toolbox API error: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Println("tool calling response body", string(body))
	return string(body), nil
}
