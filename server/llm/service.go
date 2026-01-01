package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type Service interface {
	ParseIntent(input, timezone string) (*TodoIntent, error)
}

type OpenAIService struct {
	apiKey string
	model  string
}

type TodoIntent struct {
	Summary   string `json:"summary"`
	DueAt     string `json:"due_at,omitempty"`     // ISO 8601
	Priority  string `json:"priority,omitempty"`   // "low", "medium", "high"
	Assignee  string `json:"assignee,omitempty"`   // username without @
}

func NewOpenAIService(apiKey, model string) *OpenAIService {
	return &OpenAIService{
		apiKey: apiKey,
		model:  model,
	}
}

func (s *OpenAIService) ParseIntent(input, timezone string) (*TodoIntent, error) {
	now := time.Now().UTC()
	
	systemPrompt := fmt.Sprintf(`You are a Todo Assistant.
Current Time (UTC): %s
User Timezone provided: %s

Your goal is to extract structued data from user input.
Return ONLY a valid JSON object with the following fields:
- summary: The task content (remove time/assignee references if extracted).
- due_at: ISO 8601 datetime (e.g. 2023-10-27T10:00:00Z). Calculate absolute time based on user input relative to current time. If time is not specified, omit.
- priority: "low", "medium", or "high". Default to "medium" if not specified.
- assignee: The username mentioned (remove @). If "me" or "myself", omit or return empty strings.

Example Input: "Call Hao at 5pm tomorrow urgent"
Example Output: {"summary": "Call Hao", "due_at": "2023-10-28T17:00:00Z", "priority": "high", "assignee": "haolb"}
`, now.Format(time.RFC3339), timezone)

	requestBody, _ := json.Marshal(map[string]interface{}{
		"model": s.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": input},
		},
		"temperature": 0.0,
	})

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to call OpenAI")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned status: %d", resp.StatusCode)
	}

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, errors.Wrap(err, "failed to decode response")
	}

	if len(openAIResp.Choices) == 0 {
		return nil, errors.New("no choices returned from OpenAI")
	}

	content := openAIResp.Choices[0].Message.Content
	
	// Remove potential markdown code blocks
	if len(content) > 7 && content[:3] == "```" {
		content = content[3:]
		if len(content) > 3 && content[len(content)-3:] == "```" {
			content = content[:len(content)-3]
		}
		// Strip "json" if present
		if len(content) > 4 && content[:4] == "json" {
			content = content[4:]
		}
	}

	var intent TodoIntent
	if err := json.Unmarshal([]byte(content), &intent); err != nil {
		return nil, errors.Wrapf(err, "failed to parse JSON from LLM: %s", content)
	}

	return &intent, nil
}
