package services

import (
	"bre_new_backend/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	APIURL = "https://ark.cn-beijing.volces.com/api/v3/chat/completions" // Standard OpenAI compatible endpoint usually works for Ark
	Model  = "glm-4-7-251222"
)

// Note: The user provided /api/v3/responses in the curl, but usually /chat/completions is the standard for chat.
// If /chat/completions fails, we will fallback to /responses or whatever the curl used.
// Let's try to match the curl's payload structure but use Go structs.
// The curl used "input" instead of "messages", which is specific to the /responses endpoint (Bot API?).
// However, Volcengine Ark usually supports OpenAI compatible /chat/completions.
// To be safe and follow the user's explicit curl example, I will use the endpoint and format from the curl file.

const RealAPIURL = "https://ark.cn-beijing.volces.com/api/v3/responses"

type AIRequest struct {
	Model string    `json:"model"`
	Input []AIInput `json:"input"`
	Tools []AITool  `json:"tools"`
}

type AITool struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  struct {
		Type       string `json:"type"`
		Properties struct {
			Location struct {
				Type        string `json:"type"`
				Description string `json:"description"`
			} `json:"location"`
		} `json:"properties"`
		Required []string `json:"required"`
	} `json:"parameters"`
}

type AIInput struct {
	Role    string      `json:"role"`
	Content []AIContent `json:"content"`
}

type AIContent struct {
	Type string `json:"type"` // "input_text"
	Text string `json:"text"`
}

type AIResponse struct {
	Choices []struct {
		Message struct {
			// Content might be a string (standard) or a list (bot api)
			Content interface{} `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	// The /responses endpoint might return differently.
	// Let's assume a generic map first to debug or stick to the likely response.
	// Actually, let's try to make a generic request helper.
}

// Response from /responses might be different.
// Let's treat the response as a raw JSON and parse what we need.

func CallAI(prompt string) (string, error) {

	reqBody := AIRequest{
		Model: Model,
		Input: []AIInput{
			{
				Role: "user",
				Content: []AIContent{
					{
						Type: "input_text",
						Text: prompt,
					},
				},
			},
		},
		Tools: []AITool{
			{
				Type:        "function",
				Name:        "get_news",
				Description: "Get news for a specific location",
				Parameters: struct {
					Type       string `json:"type"`
					Properties struct {
						Location struct {
							Type        string `json:"type"`
							Description string `json:"description"`
						} `json:"location"`
					} `json:"properties"`
					Required []string `json:"required"`
				}{
					Type: "object",
					Properties: struct {
						Location struct {
							Type        string `json:"type"`
							Description string `json:"description"`
						} `json:"location"`
					}{},
					Required: []string{"location"},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", RealAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+config.AppConfig.AI.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response struct {
		Choices []struct {
			Message struct {
				Content interface{} `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Output []struct {
			Type    string      `json:"type"`
			Role    string      `json:"role"`
			Content interface{} `json:"content"`
		} `json:"output"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	// Helper to extract text from content interface
	extractText := func(content interface{}) string {
		switch v := content.(type) {
		case string:
			return v
		case []interface{}:
			for _, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					if text, ok := m["text"].(string); ok {
						return text
					}
				}
			}
		}
		return fmt.Sprintf("%v", content)
	}

	if len(response.Choices) > 0 {
		return extractText(response.Choices[0].Message.Content), nil
	}

	if len(response.Output) > 0 {
		for _, item := range response.Output {
			if item.Role == "assistant" || item.Type == "message" {
				return extractText(item.Content), nil
			}
		}
	}

	return string(body), nil // Fallback
}

type NewsData struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

func GetDailyNews() ([]NewsData, error) {
	today := time.Now().Format("2006-01-02")
	prompt := `联网、联网，全网总结(至少 10 个平台) ` + today + ` 当天的国内外热点新闻并标记每个新闻的发布时间，要从多个新闻网站获取数据，
	对相同的内容的新闻进行去重处理，并总结成 20 条，请严格按照 JSON 对象数组格式输出，
	每个对象包含 title 和 url 字段。其中 url 必须是该新闻真实存在的原始报道链接（如新华网、人民网、Reuters 等），
	绝不要臆造无法访问的链接。如果无法获取真实链接则不采纳此新闻，url 必须要有保证可靠。
	不要包含 Markdown 标记或其他多余文字。
	例如：[{"title": "[年月日时分秒]新闻1", "url": "https://real-news-link..."}, {"title": "[年月日时分秒]新闻2", "url": ""}]`
	log.Printf("AI Prompt: %s", prompt)

	systemPrompt := "你是AI新闻助手。请搜索今日热点新闻。必须直接返回JSON数组格式结果，不要输出任何思考过程或Markdown标记。"
	response, err := CallAIWebSearch(systemPrompt, prompt)
	if err != nil {
		return nil, err
	}

	// Clean up response if it contains markdown code blocks
	cleanResponse := strings.TrimSpace(response)
	start := strings.Index(cleanResponse, "[")
	end := strings.LastIndex(cleanResponse, "]")
	if start != -1 && end != -1 && end > start {
		cleanResponse = cleanResponse[start : end+1]
	}

	var newsList []NewsData
	if err := json.Unmarshal([]byte(cleanResponse), &newsList); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %v. Response: %s", err, response)
	}

	return newsList, nil
}

func AnalyzeNews(newsContent string, days int) (string, error) {
	prompt := fmt.Sprintf("以下是过去 %d 天的新闻内容，请进行简要的财经分析，并推荐相关的3个板块及匹配度(只能在我给定的内容中总结分析，不要分散)：\n%s", days, newsContent)
	log.Println("AI Prompt: %s", prompt)
	return CallAI(prompt)
}

const WebSearchModel = "doubao-seed-1-8-251228"

type AIWebSearchRequest struct {
	Model  string          `json:"model"`
	Input  []AIInput       `json:"input"`
	Tools  []WebSearchTool `json:"tools"`
	Stream bool            `json:"stream"`
}

type WebSearchTool struct {
	Type         string             `json:"type"`
	Limit        int                `json:"limit,omitempty"`
	Sources      []string           `json:"sources,omitempty"`
	UserLocation *WebSearchLocation `json:"user_location,omitempty"`
}

type WebSearchLocation struct {
	Type    string `json:"type"`
	Country string `json:"country,omitempty"`
	Region  string `json:"region,omitempty"`
	City    string `json:"city,omitempty"`
}

func CallAIWebSearch(systemPrompt, prompt string) (string, error) {
	reqBody := AIWebSearchRequest{
		Model: Model,
		Input: []AIInput{
			{
				Role: "system",
				Content: []AIContent{
					{
						Type: "input_text",
						Text: systemPrompt,
					},
				},
			},
			{
				Role: "user",
				Content: []AIContent{
					{
						Type: "input_text",
						Text: prompt,
					},
				},
			},
		},
		Tools: []WebSearchTool{
			{
				Type:  "web_search",
				Limit: 50,
				//Sources: []string{"toutiao", "douyin", "moji"},
			},
		},
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", RealAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+config.AppConfig.AI.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content interface{} `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Output []struct {
			Type    string      `json:"type"`
			Role    string      `json:"role"`
			Content interface{} `json:"content"`
		} `json:"output"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	// Helper to extract text from content interface
	extractText := func(content interface{}) string {
		switch v := content.(type) {
		case string:
			return v
		case []interface{}:
			var sb strings.Builder
			for _, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					if text, ok := m["text"].(string); ok {
						sb.WriteString(text)
					}
				}
			}
			return sb.String()
		}
		return fmt.Sprintf("%v", content)
	}

	if len(response.Choices) > 0 {
		return extractText(response.Choices[0].Message.Content), nil
	}

	if len(response.Output) > 0 {
		for _, item := range response.Output {
			if item.Role == "assistant" || item.Type == "message" {
				return extractText(item.Content), nil
			}
		}
	}

	return string(body), nil
}
