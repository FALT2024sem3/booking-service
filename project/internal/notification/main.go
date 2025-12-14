package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"text/template"

	"github.com/joho/godotenv"
)

func sendHtmlEmail(to []string, subject string, htmlBody string) error {
	auth := smtp.PlainAuth(
		"",
		os.Getenv("FROM_EMAIL"),
		os.Getenv("FROM_EMAIL_PASSWORD"),
		os.Getenv("FROM_EMAIL_SMTP"),
	)

	headers := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";"

	message := "Subject: " + subject + "\n" + headers + "\n\n" + htmlBody
	return smtp.SendMail(
		os.Getenv("SMTP_ADDR"),
		auth,
		os.Getenv("FROM_EMAIL"),
		to,
		[]byte(message),
	)
}

type EmailWithTemplateRequestBody struct {
	ToAddr   string            `json:"to_addr"`
	Subject  string            `json:"subject"`
	Template string            `json:"template"`
	Vars     map[string]string `json:"vars"`
}

func HTMLTemplateEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var reqBody EmailWithTemplateRequestBody
	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Bad request with json format", http.StatusBadRequest)
		return
	}

	to := strings.Split(reqBody.ToAddr, ",")

	tmpl, err := template.ParseFiles("./templates/" + reqBody.Template + ".html")
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, reqBody.Vars); err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	log.Println(rendered.String())

	err = sendHtmlEmail(to, reqBody.Subject, rendered.String())
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "Bad request with smtp", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Email sent successfully"))
}

func main() {
	godotenv.Load()
	addr := ":8080"

	mux := http.NewServeMux()
	mux.HandleFunc("/html_email", HTMLTemplateEmailHandler)

	log.Printf("server is listening at %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
