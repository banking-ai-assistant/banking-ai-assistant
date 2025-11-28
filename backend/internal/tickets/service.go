package tickets

import "context"

type Service interface {
	ListLetters(ctx context.Context, category, status string) ([]Letter, error)
	CreateLetter(ctx context.Context, req *CreateLetterRequest) (*Letter, error)
	GetLetter(ctx context.Context, id string) (*LetterFull, error)
	GenerateReplies(ctx context.Context, id string, action string) ([]GeneratedReply, error)
	SendReply(ctx context.Context, id string, req *SendReplyRequest) (string, error)

	StartWorkers(ctx context.Context)
	StopWorkers()
}
