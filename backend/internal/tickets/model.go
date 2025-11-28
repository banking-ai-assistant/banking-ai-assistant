package tickets

type AIAnalysis struct {
	Type     string   `json:"type"`
	Urgency  string   `json:"urgency"`
	Tone     string   `json:"tone"`
	Summary  string   `json:"summary"`
	Keywords []string `json:"keywords"`
}
