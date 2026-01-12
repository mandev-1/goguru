package services

import (
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"github.com/yourusername/camagru/internal/config"
)

type EmailService struct {
	cfg *config.SMTPConfig
}

func NewEmailService(cfg *config.SMTPConfig) *EmailService {
	return &EmailService{cfg: cfg}
}

func (s *EmailService) SendVerificationEmail(to, verifyURL string) error {
	addr := net.JoinHostPort(s.cfg.Host, s.cfg.Port)
	subj := "Camagru: Verify your email"
	body := s.buildVerificationEmailHTML(verifyURL)

	msg := strings.Join([]string{
		"From: " + s.cfg.From,
		"To: " + to,
		"Subject: " + subj,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	if s.cfg.User != "" && s.cfg.Pass != "" {
		auth = smtp.PlainAuth("", s.cfg.User, s.cfg.Pass, s.cfg.Host)
	}

	return smtp.SendMail(addr, auth, s.cfg.From, []string{to}, []byte(msg))
}

func (s *EmailService) buildVerificationEmailHTML(verifyURL string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset='utf-8'>
<title>Email Verification</title>
</head>
<body style='font-family: sans-serif; background-color: #f5f5f5; padding: 20px;'>
<div style='max-width: 600px; margin: 0 auto; background: white; padding: 40px; border-radius: 8px;'>
<h2 style='color: #333; text-align: center;'>Welcome to Camagru!</h2>
<p style='color: #666; line-height: 1.6;'>Thank you for registering. Please verify your email address by clicking the button below:</p>
<div style='text-align: center; margin: 30px 0;'>
<a href='%s' style='display: inline-block; background: #00BABC; color: white; padding: 12px 30px; text-decoration: none; border-radius: 4px; font-weight: bold;'>Verify Email</a>
</div>
<p style='color: #999; font-size: 12px; line-height: 1.6;'>If the button doesn't work, copy and paste this URL into your browser:</p>
<p style='color: #999; font-size: 12px; word-break: break-all;'>%s</p>
<hr style='border: none; border-top: 1px solid #eee; margin: 30px 0;'>
<p style='color: #999; font-size: 11px; text-align: center;'>This email was sent by Camagru</p>
</div>
</body>
</html>`, verifyURL, verifyURL)
}
