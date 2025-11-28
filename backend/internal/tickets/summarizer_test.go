package tickets

import (
	"testing"
)

func TestDetectCategory(t *testing.T) {
	cases := []struct {
		name     string
		analysis AIAnalysis
		want     string
	}{
		{"critical urgency", AIAnalysis{Type: "notification", Urgency: "high"}, "critical"},
		{"regulatory type", AIAnalysis{Type: "regulatory", Urgency: "low"}, "critical"},
		{"complaint type", AIAnalysis{Type: "complaint", Urgency: "low"}, "important"},
		{"medium urgency", AIAnalysis{Type: "information_request", Urgency: "medium"}, "important"},
		{"normal case", AIAnalysis{Type: "information_request", Urgency: "low"}, "normal"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := DetectCategory(c.analysis)
			if got != c.want {
				t.Errorf("DetectCategory() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestGenerateSummary(t *testing.T) {
	cases := []struct {
		name     string
		content  string
		analysis AIAnalysis
		want     string
	}{
		{"document request", "Присылайте документы по счету", AIAnalysis{Type: "document_request"}, "Запрос документов: Присылайте документы по счету"},
		{"complaint with reason", "У меня жалоба на сервис", AIAnalysis{Type: "complaint"}, "Жалоба: Причина: У меня жалоба на сервис"},
		{"regulatory", "Регуляторное письмо", AIAnalysis{Type: "regulatory"}, "Регуляторный запрос: срочно, требуется реакция"},
		{"partnership", "Коммерческое предложение", AIAnalysis{Type: "partnership"}, "Коммерческое предложение"},
		{"fallback summary", "Запрос информации", AIAnalysis{Type: "other", Summary: "Пользователь запрашивает информацию"}, "Пользователь запрашивает информацию"},
		{"fallback keyphrase", "Общий запрос данных по счету", AIAnalysis{Type: "other"}, "Общий запрос: Общий запрос данных по счету"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := GenerateSummary(c.content, c.analysis)
			if got != c.want {
				t.Errorf("GenerateSummary() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestExtractKeyPhrase(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    string
	}{
		{"short text", "Привет.", "Привет."},
		{"long text", "Это очень длинное письмо с большим количеством слов, которое должно обрезаться по максимуму.", "Это очень длинное письмо с большим количест..."},
		{"empty", "", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := extractKeyPhrase(c.content)
			if got != c.want {
				t.Errorf("extractKeyPhrase() = %v, want %v", got, c.want)
			}
		})
	}
}

func TestExtractComplaintReason(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    string
	}{
		{"with keyword", "Я подаю жалобу на сервис", "Причина: Я подаю жалобу на сервис"},
		{"without keyword", "Просто обычное письмо", "Просто обычное письмо"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := extractComplaintReason(c.content)
			if got != c.want {
				t.Errorf("extractComplaintReason() = %v, want %v", got, c.want)
			}
		})
	}
}
