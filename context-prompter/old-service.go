package context_prompter

// import (
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"strings"
// 	"time"

// 	"github.com/hupe1980/go-tiktoken"
// )

// func getTotalTokens(msg Message) int {
// 	// Load encoding for cl100k_base (used by GPT-3.5 / GPT-4)
// 	encoding, err := tiktoken.NewEncodingForModel("gpt-3.5-turbo")
// 	if err != nil {
// 		fmt.Printf("failed to load encoding: %v", err)
// 	}

// 	ids, _, err := encoding.Encode(msg.Content[0].Text.Value, nil, nil)
// 	if err != nil {
// 		fmt.Printf("failed to encode: %v", err)
// 	}

// 	return len(ids)
// }

// func buildContext(tokenCap int, historyCap int) ([]Message, error) {
// 	roastHistory, err := database.GetRoastHistory(historyCap)
// 	if err != nil {
// 		return nil, fmt.Errorf("database error: %w", err)
// 	}

// 	var messages []Message
// 	totalTokens := 0
// 	now := time.Now()
// 	lenght := len(roastHistory) + 5

// 	// Process in order (oldest first, newest last)
// 	for i, history := range roastHistory {
// 		// Calculate timestamps with 15-second intervals
// 		userTime := now.Add(-time.Duration(lenght-i) * 30 * time.Second)
// 		assistantTime := userTime.Add(15 * time.Second)

// 		userMsg := Message{
// 			Role: "user",
// 			Content: []MessageContent{
// 				{
// 					Type:        "text",
// 					Text:        MessageText{Value: history.Tweet},
// 					Annotations: []interface{}{},
// 				},
// 			},
// 			CreatedAt: userTime,
// 		}

// 		assistantMsg := Message{
// 			Role: "assistant",
// 			Content: []MessageContent{
// 				{
// 					Type:        "text",
// 					Text:        MessageText{Value: fmt.Sprintf(`{"topic":"%s","brand":"%s","score":%d}`, history.Topic, history.Brand, history.Score)},
// 					Annotations: []interface{}{},
// 				},
// 			},
// 			CreatedAt: assistantTime,
// 		}

// 		// Token calculation
// 		pairTokens := getTotalTokens(userMsg) + getTotalTokens(assistantMsg) + 5
// 		if totalTokens+pairTokens > int(float64(tokenCap)*0.8) {
// 			break
// 		}

// 		messages = append(messages, userMsg, assistantMsg)
// 		totalTokens += pairTokens
// 	}

// 	return messages, nil
// }

// // clean strips code fences and markdown emphasis.
// func clean(raw string) string {
// 	s := strings.TrimSpace(raw)
// 	// remove triple-backtick blocks
// 	if strings.HasPrefix(s, "```") {
// 		parts := strings.SplitN(s, "\n", 2)
// 		if len(parts) == 2 {
// 			s = parts[1]
// 		}
// 		if idx := strings.LastIndex(s, "```"); idx != -1 {
// 			s = s[:idx]
// 		}
// 	}
// 	// remove any leftover backticks or bold markers
// 	s = strings.ReplaceAll(s, "`", "")
// 	s = strings.ReplaceAll(s, "**", "")
// 	return strings.TrimSpace(s)
// }

// // normalize result
// func clampResult(result RoastResult) RoastResult {
// 	if result.Score < 0 {
// 		result.Score = 0
// 	} else if result.Score > 100 {
// 		result.Score = 100
// 	}
// 	if result.Topic == "" {
// 		result.Topic = "unknown"
// 	}
// 	if result.BrandUsername == "" {
// 		result.BrandUsername = "general"
// 	}

// 	// External filter: if AI returns "toobad_bot" as brand, map it to "bad_chain"
// 	// This handles cases where the AI ignores the prompt instruction
// 	if result.BrandUsername == "toobad_bot" {
// 		result.BrandUsername = "bad_chain"
// 	}

// 	return result
// }

// // Get Roast Results with context builder
// func GetRoastResults(text string, imageURL string, userMentions []string) (RoastResult, error) {
// 	ctx := context.Background()

// 	if text == "" && imageURL == "" {
// 		return RoastResult{}, fmt.Errorf("at least tweet text or image URL must be provided")
// 	}

// 	// If it's an image, use the old method
// 	if imageURL != "" {
// 		result, err := openai.GetRoastResult(text, imageURL, userMentions)
// 		return RoastResult(result), err
// 	}

// 	// 1. Build your context messages
// 	msgs, err := buildContext(token_cap, n_cap)
// 	if err != nil {
// 		return RoastResult{}, err
// 	}
// 	assistant := NewContextualAssistant()

// 	var msgReq []MessageRequest
// 	for _, msg := range msgs {
// 		msgReq = append(msgReq, MessageRequest{
// 			Role:    msg.Role,
// 			Content: msg.Content[0].Text.Value,
// 			Metadata: map[string]interface{}{
// 				"created_at": msg.CreatedAt.Format(time.RFC3339),
// 			},
// 		})
// 	}

// 	// append the new roast
// 	msgReq = append(msgReq, MessageRequest{
// 		Role:    "user",
// 		Content: text,
// 		Metadata: map[string]interface{}{
// 			"created_at": time.Now().Format(time.RFC3339),
// 		},
// 	})
// 	for _, msg := range msgReq {
// 		fmt.Printf("---------------- \n")
// 		fmt.Printf("msg.Role: %v\n", msg.Role)
// 		fmt.Printf("msg.Content: %v\n", msg.Content)
// 		fmt.Printf("msg.CreatedAt: %v\n", msg.Metadata["created_at"])
// 		fmt.Printf("---------------- \n")
// 	}
// 	// Run with full context
// 	run, err := assistant.RunWithContext(ctx, RunOptions{
// 		Messages: msgReq,
// 		Timeout:  60 * time.Second,
// 	})

// 	if err != nil {
// 		log.ErrorF("Error at RunWithContext: %v", err)
// 		return RoastResult{}, err
// 	}

// 	// 8. Find the most recent assistant message
// 	var lastAssistantMsg string
// 	var timeLate time.Time
// 	for _, msg := range run.Messages {
// 		fmt.Printf("msg.Role: %v\n", msg.Role)
// 		fmt.Printf("timeLate: %v\n", timeLate)
// 		fmt.Printf("msg.CreatedAt: %v\n", msg.CreatedAt)
// 		if msg.Role == "assistant" && (timeLate.IsZero() || timeLate.Before(msg.CreatedAt)) {
// 			lastAssistantMsg = msg.Content[0].Text.Value
// 			break
// 		}
// 	}

// 	if lastAssistantMsg == "" {
// 		return RoastResult{}, errors.New("no assistant response found")
// 	}

// 	// 9. Parse and return the result
// 	cleaned := clean(lastAssistantMsg)
// 	var result RoastResult
// 	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
// 		return RoastResult{}, fmt.Errorf("parsing JSON (%s): %w", cleaned, err)
// 	}

// 	return clampResult(result), nil
// }
