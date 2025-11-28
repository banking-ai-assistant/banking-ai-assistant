package tickets

import "errors"

type CreateLetterRequest struct {
	OriginalID  string `json:"original_id"`
	Subject     string `json:"subject"`
	Content     string `json:"content"`
	SenderEmail string `json:"sender_email"`
	SenderName  string `json:"sender_name"`
}

func (r *CreateLetterRequest) Validate() error {
	if r.Content == "" {
		return errors.New("content required")
	}
	if r.SenderEmail == "" {
		return errors.New("sender_email required")
	}
	return nil
}

type GenerateRequest struct {
	Action string `json:"action"`
}

func (g *GenerateRequest) Validate() error {
	if g.Action == "" {
		return errors.New("action required")
	}
	return nil
}

type SendReplyRequest struct {
	ReplyID       string `json:"reply_id"`
	CustomContent string `json:"custom_content"`
}

func (s *SendReplyRequest) Validate() error {
	if s.ReplyID == "" && s.CustomContent == "" {
		return errors.New("reply_id or custom_content required")
	}
	return nil
}
