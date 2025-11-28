package tickets

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"banking-ai-assistant/pkg/utils"
)

type postgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) Repository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) SaveLetter(ctx context.Context, l *Letter) error {
	if l.ID == "" {
		l.ID = utils.NewID()
	}
	if l.ReceivedAt.IsZero() {
		l.ReceivedAt = time.Now().UTC()
	}
	aiBytes := []byte("null")
	if l.AIAnalysis != nil {
		b, _ := json.Marshal(l.AIAnalysis)
		aiBytes = b
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO letters (id, original_id, subject, content, sender_email, sender_name, received_at, status, ai_analysis, category)
         VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
         ON CONFLICT (id) DO UPDATE SET subject = EXCLUDED.subject, content = EXCLUDED.content, sender_email = EXCLUDED.sender_email, sender_name = EXCLUDED.sender_name, status = EXCLUDED.status, ai_analysis = EXCLUDED.ai_analysis, category = EXCLUDED.category`,
		l.ID, l.OriginalID, l.Subject, l.Content, l.SenderEmail, l.SenderName, l.ReceivedAt, l.Status, string(aiBytes), l.Category,
	)
	return err
}

func (r *postgresRepo) ListLetters(ctx context.Context, category, status string) ([]Letter, error) {
	q := `SELECT id, original_id, subject, content, sender_email, sender_name, received_at, replied_at, status, ai_analysis, category FROM letters WHERE 1=1`
	args := []any{}
	if category != "" {
		q += ` AND category = $1`
		args = append(args, category)
	}
	if status != "" {
		if len(args) == 0 {
			q += ` AND status = $1`
		} else {
			q += ` AND status = $2`
		}
		args = append(args, status)
	}
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Letter
	for rows.Next() {
		var l Letter
		var aiRaw sql.NullString
		var repliedAt sql.NullTime
		if err := rows.Scan(&l.ID, &l.OriginalID, &l.Subject, &l.Content, &l.SenderEmail, &l.SenderName, &l.ReceivedAt, &repliedAt, &l.Status, &aiRaw, &l.Category); err != nil {
			return nil, err
		}
		if repliedAt.Valid {
			t := repliedAt.Time
			l.RepliedAt = &t
		}
		if aiRaw.Valid && aiRaw.String != "" {
			var ai AIAnalysis
			if err := json.Unmarshal([]byte(aiRaw.String), &ai); err == nil {
				l.AIAnalysis = &ai
			}
		}
		out = append(out, l)
	}
	return out, nil
}

func (r *postgresRepo) GetLetterByID(ctx context.Context, id string) (*Letter, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, original_id, subject, content, sender_email, sender_name, received_at, replied_at, status, ai_analysis, category FROM letters WHERE id=$1`, id)
	var l Letter
	var aiRaw sql.NullString
	var repliedAt sql.NullTime
	if err := row.Scan(&l.ID, &l.OriginalID, &l.Subject, &l.Content, &l.SenderEmail, &l.SenderName, &l.ReceivedAt, &repliedAt, &l.Status, &aiRaw, &l.Category); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if repliedAt.Valid {
		t := repliedAt.Time
		l.RepliedAt = &t
	}
	if aiRaw.Valid && aiRaw.String != "" {
		var ai AIAnalysis
		if err := json.Unmarshal([]byte(aiRaw.String), &ai); err == nil {
			l.AIAnalysis = &ai
		}
	}
	return &l, nil
}

func (r *postgresRepo) SaveGeneratedReply(ctx context.Context, rr *GeneratedReply) error {
	if rr.ID == "" {
		rr.ID = utils.NewID()
	}
	if rr.CreatedAt.IsZero() {
		rr.CreatedAt = time.Now().UTC()
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO generated_replies (id, letter_id, content, style, status, created_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		rr.ID, rr.LetterID, rr.Content, rr.Style, rr.Status, rr.CreatedAt)
	return err
}

func (r *postgresRepo) ListRepliesByLetter(ctx context.Context, letterID string) ([]GeneratedReply, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, letter_id, content, style, status, created_at FROM generated_replies WHERE letter_id=$1 ORDER BY created_at DESC`, letterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []GeneratedReply
	for rows.Next() {
		var g GeneratedReply
		if err := rows.Scan(&g.ID, &g.LetterID, &g.Content, &g.Style, &g.Status, &g.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, nil
}

func (r *postgresRepo) UpdateLetterStatus(ctx context.Context, id, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE letters SET status=$2, updated_at=now() WHERE id=$1`, id, status)
	return err
}

func (r *postgresRepo) MarkLetterReplied(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE letters SET replied_at=now(), status='replied', updated_at=now() WHERE id=$1`, id)
	return err
}
