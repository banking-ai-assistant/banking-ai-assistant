package tickets

import "time"

type AIAnalysis struct {
	Type     string   `json:"type"`
	Urgency  string   `json:"urgency"`
	Tone     string   `json:"tone"`
	Summary  string   `json:"summary"`
	Keywords []string `json:"keywords"`
}

type Letter struct {
	ID          string      `json:"id"`
	OriginalID  string      `json:"original_id"`
	Subject     string      `json:"subject"`
	Content     string      `json:"content"`
	SenderEmail string      `json:"sender_email"`
	SenderName  string      `json:"sender_name"`
	ReceivedAt  time.Time   `json:"received_at"`
	RepliedAt   *time.Time  `json:"replied_at,omitempty"`
	Status      string      `json:"status"`
	AIAnalysis  *AIAnalysis `json:"ai_analysis,omitempty"`
	Category    string      `json:"category"`
}

type GeneratedReply struct {
	ID        string    `json:"id"`
	LetterID  string    `json:"letter_id"`
	Content   string    `json:"content"`
	Style     string    `json:"style"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type LetterFull struct {
	Letter           *Letter          `json:"letter"`
	GeneratedReplies []GeneratedReply `json:"generated_replies"`
}
