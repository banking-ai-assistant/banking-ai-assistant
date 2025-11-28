package tickets

import (
	"testing"
)

func TestCreateLetterRequestValidate(t *testing.T) {
	cases := []struct {
		name    string
		req     CreateLetterRequest
		wantErr bool
	}{
		{"valid", CreateLetterRequest{Content: "Hi", SenderEmail: "a@b.com"}, false},
		{"empty content", CreateLetterRequest{Content: "", SenderEmail: "a@b.com"}, true},
		{"empty email", CreateLetterRequest{Content: "Hi", SenderEmail: ""}, true},
		{"both empty", CreateLetterRequest{Content: "", SenderEmail: ""}, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.req.Validate()
			if (err != nil) != c.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, c.wantErr)
			}
		})
	}
}

func TestGenerateRequestValidate(t *testing.T) {
	cases := []struct {
		name    string
		req     GenerateRequest
		wantErr bool
	}{
		{"valid", GenerateRequest{Action: "create"}, false},
		{"empty action", GenerateRequest{Action: ""}, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.req.Validate()
			if (err != nil) != c.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, c.wantErr)
			}
		})
	}
}

func TestSendReplyRequestValidate(t *testing.T) {
	cases := []struct {
		name    string
		req     SendReplyRequest
		wantErr bool
	}{
		{"valid reply_id", SendReplyRequest{ReplyID: "r1"}, false},
		{"valid custom content", SendReplyRequest{CustomContent: "Hello"}, false},
		{"both empty", SendReplyRequest{}, true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.req.Validate()
			if (err != nil) != c.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, c.wantErr)
			}
		})
	}
}
