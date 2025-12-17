package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"text/template"
)

type EmailWithTemplateRequestBody struct {
	ToAddr   string            `json:"to_addr"`
	Subject  string            `json:"subject"`
	Template string            `json:"template"`
	Vars     map[string]string `json:"vars"`
}

func sendHtmlEmail(to []string, subject string, htmlBody string) error {
	auth := smtp.PlainAuth(
		"",
		os.Getenv("FROM_EMAIL"),
		os.Getenv("FROM_EMAIL_PASSWORD"),
		os.Getenv("FROM_EMAIL_SMTP"),
	)

	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";"
	message := "Subject: " + subject + "\n" + headers + "\n\n" + htmlBody

	addr := os.Getenv("SMTP_ADDR")
	if addr == "" {
		return fmt.Errorf("SMTP_ADDR environment variable is not set")
	}

	return smtp.SendMail(
		addr,
		auth,
		os.Getenv("FROM_EMAIL"),
		to,
		[]byte(message),
	)
}

func SendEmailLogic(req EmailWithTemplateRequestBody) error {
	to := strings.Split(req.ToAddr, ",")

	tmplPath := "templates/" + req.Template + ".html"
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", tmplPath, err)
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, req.Vars); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	if err := sendHtmlEmail(to, req.Subject, rendered.String()); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func HTMLTemplateEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var reqBody EmailWithTemplateRequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Bad request with json format", http.StatusBadRequest)
		return
	}

	if err := SendEmailLogic(reqBody); err != nil {
		log.Printf("Error sending email: %v", err)
		http.Error(w, "Failed to send email: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Email sent successfully"))
}
