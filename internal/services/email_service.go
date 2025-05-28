package services

import (
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailService interface {
	SendInviteEmail(to string) error
}

type SendGridService struct {
	FromName  string
	FromEmail string
	APIKey    string
}

func NewSendGridService(fromName, fromEmail, apiKey string) *SendGridService {
	return &SendGridService{fromName, fromEmail, apiKey}
}

func (s *SendGridService) SendInviteEmail(to string) error {
	from := mail.NewEmail(s.FromName, s.FromEmail)
	recipient := mail.NewEmail("", to)
	subject := "You're Invited!"
	plainTextContent := "Your account is ready. Sign up here:"
	htmlContent := `<p>Your account is ready. <a href="https://littleeinsteinchildcare.org/signup">Click to sign up</a></p>`

	message := mail.NewSingleEmail(from, subject, recipient, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(s.APIKey)

	resp, err := client.Send(message)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("sendgrid error: %v", resp.Body)
	}

	return nil
}