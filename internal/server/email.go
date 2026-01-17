package server

import (
	"fmt"
	"net/smtp"
	"os"
)

func (s *Server) SendVerificationEmail(to, username, url string) {
	subject := "Verify your Camagru account"
	body := fmt.Sprintf(`
Hello %s,

Thank you for registering with Camagru!

Please verify your account by clicking the following link:
%s

If you did not create this account, please ignore this email.

Best regards,
Camagru Team
`, username, url)

	s.SendEmail(to, subject, body)
}

func (s *Server) SendPasswordResetEmail(to, username, url string) {
	subject := "Reset your Camagru password"
	body := fmt.Sprintf(`
Hello %s,

You requested to reset your password. Click the following link to reset it:
%s

This link will expire in 30 minutes.

If you did not request this, please ignore this email.

Best regards,
Camagru Team
`, username, url)

	s.SendEmail(to, subject, body)
}

func (s *Server) SendCommentNotification(to, author, comment string) {
	subject := "New comment on your image"
	body := fmt.Sprintf(`
Hello,

%s commented on your image:
"%s"

Visit your gallery to see all comments.

Best regards,
Camagru Team
`, author, comment)

	s.SendEmail(to, subject, body)
}

func (s *Server) SendLikeNotification(to, author string) {
	subject := "Someone liked your image"
	body := fmt.Sprintf(`
Hello,

%s liked your image!

Visit your gallery to see all your likes.

Best regards,
Camagru Team
`, author)

	s.SendEmail(to, subject, body)
}

func (s *Server) SendEmail(to, subject, body string) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	fromEmail := os.Getenv("FROM_EMAIL")
	if smtpHost == "" {
		smtpHost = "mailhog"
	}
	if smtpPort == "" {
		smtpPort = "1025"
	}
	if fromEmail == "" {
		fromEmail = "noreply@camagru.local"
	}
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", fromEmail, to, subject, body)
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	if smtpUser != "" && smtpPass != "" {
		auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
		smtp.SendMail(addr, auth, fromEmail, []string{to}, []byte(msg))
	} else {
		smtp.SendMail(addr, nil, fromEmail, []string{to}, []byte(msg))
	}
}
