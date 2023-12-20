package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkoukk/tiktoken-go"
	"os"
	"strconv"
)

type MsgBody struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
	Model      string `json:"model"`
}

func HandleRequest(ctx context.Context, event events.APIGatewayProxyResponse) (events.APIGatewayProxyResponse, error) {
	var msgBody MsgBody
	err := json.Unmarshal([]byte(event.Body), &msgBody)
	if err != nil {
		fmt.Printf("Error unmarshaling message: %v\n", err)
	}

	tke, err := tiktoken.EncodingForModel(msgBody.Model)
	if err != nil {
		fmt.Printf("getEncoding: %v\n", err)
	}

	promptTokens := len(tke.Encode(msgBody.Prompt, nil, nil))
	completionTokens := len(tke.Encode(msgBody.Completion, nil, nil))

	promptCost := 0.0
	completionCost := 0.0
	totalCost := 0.0

	envBaseRatio := os.Getenv("BASE_RATIO")
	if envBaseRatio == "" {
		envBaseRatio = "1.0"
	}
	baseRatio, err := strconv.ParseFloat(envBaseRatio, 64)
	if err != nil {
		baseRatio = 1.0
	}

	switch msgBody.Model {
	case "gpt-4-vision-preview":
		promptCost = float64(promptTokens) * 0.01 * baseRatio / 1000
		completionCost = float64(completionTokens) * 0.03 * baseRatio / 1000
	case "gpt-4-1106-preview":
		promptCost = float64(promptTokens) * 0.01 * baseRatio / 1000
		completionCost = float64(completionTokens) * 0.03 * baseRatio / 1000
	case "gpt-4-0314":
		promptCost = float64(promptTokens) * 0.03 * baseRatio / 1000
		completionCost = float64(completionTokens) * 0.06 * baseRatio / 1000
	case "gpt-4":
		promptCost = float64(promptTokens) * 0.03 * baseRatio / 1000
		completionCost = float64(completionTokens) * 0.06 * baseRatio / 1000
	case "gpt-3.5-turbo-0301":
		promptCost = float64(promptTokens) * 0.0016 * baseRatio / 1000
		completionCost = float64(completionTokens) * 0.002 * baseRatio / 1000
	case "gpt-3.5-turbo":
		promptCost = float64(promptTokens) * 0.0016 * baseRatio / 1000
		completionCost = float64(completionTokens) * 0.002 * baseRatio / 1000
	case "gpt-3.5-turbo-16k":
		promptCost = float64(promptTokens) * 0.003 * baseRatio / 1000
		completionCost = float64(completionTokens) * 0.004 * baseRatio / 1000
	case "gpt-3.5-turbo-1106":
		promptCost = float64(promptTokens) * 0.001 * baseRatio / 1000
		completionCost = float64(completionTokens) * 0.002 * baseRatio / 1000
	default:
		fmt.Println("Model not supported")
	}

	totalCost = promptCost + completionCost

	return events.APIGatewayProxyResponse{Body: fmt.Sprintf(`{
		"usage": {
			"prompt_tokens": %d,
			"completion_tokens": %d,
			"total_tokens": %d
		},
		"cost": {
			"prompt_cost": %f,
			"completion_cost": %f,
			"total_cost": %f
		}
	}`, promptTokens, completionTokens, promptTokens+completionTokens, promptCost, completionCost, totalCost), StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}