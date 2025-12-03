package email

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os"
)

// SendVerificationEmail delivers the account verification link via SMTP.
func SendVerificationEmail(to, token string) error {
	host := getenv("SMTP_HOST", "localhost")
	port := getenv("SMTP_PORT", "1025")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	from := getenv("SMTP_FROM", "no-reply@camagru.local")

	subject := "Verify your Camagru account"
	verificationURL := fmt.Sprintf("%s://%s/verify?token=%s", getenv("APP_SCHEME", "http"), getenv("APP_HOST", host+":"+port), token)
	msg := buildMessage(from, to, subject, buildVerificationEmailHTML(verificationURL))

	addr := net.JoinHostPort(host, port)
	auth := smtp.PlainAuth("", user, pass, host)

	if os.Getenv("SMTP_TLS") == "true" {
		return sendWithTLS(addr, auth, from, []string{to}, msg)
	}
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

func buildMessage(from, to, subject, body string) []byte {
	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=\"utf-8\"\r\n\r\n", from, to, subject)
	return []byte(headers + body)
}

func buildVerificationEmailHTML(url string) string {
	return fmt.Sprintf(`<html><body><p>Welcome to Camagru!</p><p>Please verify your account by clicking <a href="%s">this link</a>.</p></body></html>`, url)
}

func sendWithTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	host, _, _ := net.SplitHostPort(addr)
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: host, InsecureSkipVerify: true})
	if err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()

	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return err
		}
	}
	if err = client.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = writer.Write(msg); err != nil {
		return err
	}
	return writer.Close()
}

func getenv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}



































































}	return body.String(), nil	}		return "", fmt.Errorf("executing template: %w", err)	if err := t.Execute(&body, link); err != nil {	var body bytes.Buffer	}		return "", fmt.Errorf("parsing template: %w", err)	if err != nil {	`)	</html>	</body>		<p><a href="{{.}}">{{.}}</a></p>		<p>Please verify your email address by clicking the link below:</p>	<body>	</head>		<title>Verification</title>		<meta charset="utf-8">	<head>	<html>	<!DOCTYPE html>	t, err := template.New("verification").Parse(`func buildVerificationEmailHTML(link string) (string, error) {}	return smtp.SendMail(host+":"+port, auth, from, []string{to}, []byte(msg))	}		auth = smtp.PlainAuth("", from, pass, host)	if pass != "" {	var auth smtp.Auth		body		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +		"Subject: " + subject + "\n" +		"To: " + to + "\n" +	msg := "From: " + from + "\n" +	}		return err	if err != nil {	body, err := buildVerificationEmailHTML(subject)	}		port = "1025" // mailhog default	if port == "" {	port := os.Getenv("SMTP_PORT")	}		host = "127.0.0.1"	if host == "" {	host := os.Getenv("SMTP_HOST")	pass := os.Getenv("SMTP_PASS")	}		from = "noreply@localhost"	if from == "" {	from := os.Getenv("SMTP_FROM")func SendVerificationEmail(to, subject string) error {)	"os"	"net/smtp"	"html/template"	"fmt"	"bytes"import (