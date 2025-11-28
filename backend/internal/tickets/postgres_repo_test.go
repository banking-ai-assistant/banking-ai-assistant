package tickets

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestSaveAndGetLetter(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepo(db)

	letter := &Letter{
		ID:          "l1",
		OriginalID:  "o1",
		Subject:     "Subj",
		Content:     "Content",
		SenderEmail: "a@b.com",
		SenderName:  "John",
		ReceivedAt:  time.Now().UTC(),
		Status:      "new",
		Category:    "normal",
		AIAnalysis: &AIAnalysis{
			Type:     "info",
			Urgency:  "medium",
			Tone:     "neutral",
			Summary:  "summary",
			Keywords: []string{"kw"},
		},
	}

	aiBytes, _ := json.Marshal(letter.AIAnalysis)
	mock.ExpectExec("INSERT INTO letters").
		WithArgs(letter.ID, letter.OriginalID, letter.Subject, letter.Content,
			letter.SenderEmail, letter.SenderName, letter.ReceivedAt, letter.Status,
			string(aiBytes), letter.Category).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.SaveLetter(context.Background(), letter)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestListLetters(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepo(db)

	rows := sqlmock.NewRows([]string{"id", "original_id", "subject", "content", "sender_email", "sender_name", "received_at", "replied_at", "status", "ai_analysis", "category"}).
		AddRow("l1", "o1", "Subj", "Content", "a@b.com", "John", time.Now(), nil, "new", `{"Type":"info"}`, "normal")

	mock.ExpectQuery("SELECT id, original_id, subject, content, sender_email, sender_name, received_at, replied_at, status, ai_analysis, category FROM letters").
		WillReturnRows(rows)

	letters, err := repo.ListLetters(context.Background(), "", "")
	require.NoError(t, err)
	require.Len(t, letters, 1)
}

func TestGetLetterByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepo(db)

	aiStr := `{"Type":"info"}`
	rows := sqlmock.NewRows([]string{"id", "original_id", "subject", "content", "sender_email", "sender_name", "received_at", "replied_at", "status", "ai_analysis", "category"}).
		AddRow("l1", "o1", "Subj", "Content", "a@b.com", "John", time.Now(), nil, "new", aiStr, "normal")

	mock.ExpectQuery("SELECT id, original_id, subject, content, sender_email, sender_name, received_at, replied_at, status, ai_analysis, category FROM letters WHERE id=").
		WithArgs("l1").WillReturnRows(rows)

	l, err := repo.GetLetterByID(context.Background(), "l1")
	require.NoError(t, err)
	require.NotNil(t, l)
	require.Equal(t, "l1", l.ID)
	require.NotNil(t, l.AIAnalysis)
}

func TestSaveAndListGeneratedReply(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepo(db)

	reply := &GeneratedReply{
		ID:        "r1",
		LetterID:  "l1",
		Content:   "Reply",
		Style:     "formal",
		Status:    "draft",
		CreatedAt: time.Now().UTC(),
	}

	mock.ExpectExec("INSERT INTO generated_replies").
		WithArgs(reply.ID, reply.LetterID, reply.Content, reply.Style, reply.Status, reply.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.SaveGeneratedReply(context.Background(), reply)
	require.NoError(t, err)

	rows := sqlmock.NewRows([]string{"id", "letter_id", "content", "style", "status", "created_at"}).
		AddRow(reply.ID, reply.LetterID, reply.Content, reply.Style, reply.Status, reply.CreatedAt)

	mock.ExpectQuery("SELECT id, letter_id, content, style, status, created_at FROM generated_replies WHERE letter_id=").WithArgs("l1").
		WillReturnRows(rows)

	list, err := repo.ListRepliesByLetter(context.Background(), "l1")
	require.NoError(t, err)
	require.Len(t, list, 1)
	require.Equal(t, "r1", list[0].ID)
}
