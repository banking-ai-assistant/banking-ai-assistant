package ai

type AnalyzeRequest struct {
	Content string `json:"content"`
}

type AnalyzeResponse struct {
	Type     string   `json:"type"`
	Urgency  string   `json:"urgency"`
	Tone     string   `json:"tone"`
	Summary  string   `json:"summary"`
	Keywords []string `json:"keywords"`
}

type GenerateRequest struct {
	Content string `json:"content"`
	Action  string `json:"action"`
}

type GenerateOption struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Style   string `json:"style"`
	Summary string `json:"summary"`
}

type GenerateResponse struct {
	Options []GenerateOption `json:"options"`
}
