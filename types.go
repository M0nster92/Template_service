package main

import (
	"net/smtp"
	"time"

	"github.com/mohamedattahri/mail"
)

type TemplateMail struct {
	From       mail.Address
	Recipients []mail.Address
	Subject    string
	SMTPHost   string
	SMTPAuth   smtp.Auth
	HTML       string
	Text       string
	msg        *mail.Message
	parts      *mail.Multipart
	rcpt       []string
	hastext    bool
	hashtml    bool
}

type Template struct {
	ID              string    `bson:"template_id"`
	Description     string    `bson:"Description"`
	Application     []string  `bson:"Application"`
	Subject         string    `bson:"Subject"`
	Text            string    `bson:"Text"`
	CreatedDate     time.Time `bson:"CreatedDate"`
	LastUpdate      time.Time `bson:"LastUpdate"`
	CreatedUsername string    `bson:"CreatedUsername"`
	UpdatedUsername string    `bson:"UpdatedUsername"`
	JsonDesign      string    `bson:"JsonDesign"`
}

type Response struct {
	Status string      `json:"status"`
	Error  string      `json:"error,omitempty"`
	Data   interface{} `json:"data,omitempty"`
}
