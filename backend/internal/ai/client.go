package ai

import (
	"context"
	"time"

	"banking-ai-assistant/pkg/utils"
)

type Client interface {
	Analyze(ctx context.Context, content string) (*AnalyzeResponse, error)
	Generate(ctx context.Context, content, action string) (*GenerateResponse, error)
}

type stubClient struct{}

func NewStubClient() Client {
	return &stubClient{}
}

func (s *stubClient) Analyze(ctx context.Context, content string) (*AnalyzeResponse, error) {
	select {
	case <-time.After(300 * time.Millisecond):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	resp := &AnalyzeResponse{
		Type:    "document_request",
		Urgency: "low",
		Tone:    "formal",
		Summary: "Запрос документов: краткое содержание (stub).",
		Keywords: []string{
			"выписка", "договор",
		},
	}
	if len(content) > 300 {
		resp.Urgency = "medium"
		resp.Type = "complaint"
		resp.Summary = "Длинный текст, возможно претензия."
	}
	if containsWord(content, "регул") || containsWord(content, "надзор") {
		resp.Type = "regulatory"
		resp.Urgency = "high"
		resp.Summary = "Регуляторный запрос: требует срочной реакции."
	}

	return resp, nil
}

func (s *stubClient) Generate(ctx context.Context, content, action string) (*GenerateResponse, error) {
	select {
	case <-time.After(400 * time.Millisecond):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	options := []GenerateOption{
		{ID: utils.NewID(), Content: "Вариант отказа — примерный текст.", Style: "refusal", Summary: "Вежливый отказ"},
		{ID: utils.NewID(), Content: "Вариант подтверждения — ваш запрос будет выполнен.", Style: "approval", Summary: "Подтверждение"},
		{ID: utils.NewID(), Content: "Вариант совета — рекомендация.", Style: "advice", Summary: "Рекомендация"},
	}
	return &GenerateResponse{Options: options}, nil
}

func containsWord(s, w string) bool {
	return len(s) > 0 && (stringContainsIgnoreCase(s, w))
}

func stringContainsIgnoreCase(s, sub string) bool {
	return len(s) >= len(sub) && (indexOfIgnoreCase(s, sub) >= 0)
}

func indexOfIgnoreCase(s, sub string) int {
	S := []rune(s)
	T := []rune(sub)
	n := len(S)
	m := len(T)
	if m == 0 {
		return 0
	}
	for i := 0; i <= n-m; i++ {
		match := true
		for j := 0; j < m; j++ {
			a := S[i+j]
			b := T[j]
			if toLower(a) != toLower(b) {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func toLower(r rune) rune {
	if 'A' <= r && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}
