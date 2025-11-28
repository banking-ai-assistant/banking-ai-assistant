package tickets

import (
	"strings"
)

func DetectCategory(analysis AIAnalysis) string {
	u := strings.ToLower(strings.TrimSpace(analysis.Urgency))
	t := strings.ToLower(strings.TrimSpace(analysis.Type))

	if u == "critical" || u == "high" || t == "regulatory" {
		return "critical"
	}
	if u == "medium" || t == "complaint" {
		return "important"
	}
	return "normal"
}

func GenerateSummary(content string, analysis AIAnalysis) string {
	t := strings.ToLower(strings.TrimSpace(analysis.Type))

	switch t {
	case "document_request":
		return "Запрос документов: " + extractKeyPhrase(content)
	case "complaint":
		return "Жалоба: " + extractComplaintReason(content)
	case "regulatory":
		return "Регуляторный запрос: срочно, требуется реакция"
	case "partnership":
		return "Коммерческое предложение"
	default:
		if strings.TrimSpace(analysis.Summary) != "" {
			return analysis.Summary
		}
		return "Общий запрос: " + extractKeyPhrase(content)
	}
}

func extractKeyPhrase(content string) string {
	const max = 80
	s := strings.TrimSpace(content)
	if len(s) == 0 {
		return ""
	}
	if idx := strings.IndexAny(s, ".\n"); idx >= 0 && idx <= max {
		return strings.TrimSpace(s[:idx+1])
	}
	if len(s) > max {
		return strings.TrimSpace(s[:max]) + "..."
	}
	return s
}

func extractComplaintReason(content string) string {
	l := strings.ToLower(content)
	found := []string{"жалоб", "претенз", "недовольн", "нарушен", "неудовлетвор"}
	for _, kw := range found {
		if i := strings.Index(l, kw); i >= 0 {
			start := i - 20
			if start < 0 {
				start = 0
			}
			end := i + 60
			if end > len(l) {
				end = len(l)
			}
			frag := strings.TrimSpace(content[start:end])
			return "Причина: " + frag
		}
	}

	return extractKeyPhrase(content)
}
