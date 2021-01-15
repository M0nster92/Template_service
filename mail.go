package main

import (
	//	"bufio"

	"bytes"
	"html/template"
	"net/smtp"

	log "github.com/sirupsen/logrus"

	"github.com/mohamedattahri/mail"
)

func NewTemplateMail(server string, from mail.Address, subject string, recipients []mail.Address) *TemplateMail {
	obj := TemplateMail{
		From:       from,
		Subject:    subject,
		SMTPHost:   server,
		Recipients: recipients,
	}
	obj.msg = mail.NewMessage()
	obj.msg.SetFrom(&from)
	obj.msg.SetSubject(obj.Subject)
	for _, addr := range recipients {
		obj.msg.To().Add(&addr)
		obj.rcpt = append(obj.rcpt, addr.Address)
	}
	obj.parts = mail.NewMultipart("multipart/alternative", obj.msg)
	return &obj
}

func (obj *TemplateMail) AddHtml(tpl string, data interface{}) error {
	//fmt.Println("Object in AddHtml ", data)
	t, err := template.New("htmlemail").Parse(tpl)
	var buf bytes.Buffer
	//	var buf2 bytes.Buffer
	if err != nil {
		log.Errorf("Problem parsing HTML template %v", err)
		return err
	}

	err = t.Execute(&buf, data)
	if err != nil {
		log.Errorf("Problem executing HTML template %v", err)
		return err
	}

	err = obj.parts.AddText("text/html", &buf)
	if err != nil {
		log.Error(err)
	} else {
		obj.hashtml = true
	}

	return err
}

func (obj *TemplateMail) AddText(tpl string, data interface{}) error {
	t, err := template.New("textemail").Parse(tpl)
	var buf bytes.Buffer
	if err != nil {
		log.Errorf("Problem parsing Text template %v", err)
		return err
	}
	err = t.Execute(&buf, data)
	if err != nil {
		log.Errorf("Problem executing Text template %v", err)
		return err
	}
	err = obj.parts.AddText("text/plain", &buf)
	if err != nil {
		log.Error(err)
	} else {
		obj.hastext = true
	}
	return err
}

// Send - Send the e-mail
func (obj *TemplateMail) Send() error {

	var err error

	obj.parts.Close()

	body := obj.msg.Bytes()
	log.Debugf("Body being sent is %v", string(body))

	err = smtp.SendMail(obj.SMTPHost, obj.SMTPAuth, obj.From.Address, obj.rcpt, body)
	if err != nil {
		log.Errorf("Error sending e-mail %v", err)
	}
	return err
}
