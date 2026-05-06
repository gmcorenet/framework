package mailer

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net/smtp"
)

type MailerInterface interface {
	Send(email *Email) error
}

type Email struct {
	From        Mailbox
	To          []Mailbox
	Cc          []Mailbox
	Bcc         []Mailbox
	Subject     string
	Body        string
	Html        bool
	Attachments []Attachment
	Headers     map[string]string
}

type Mailbox struct {
	Address string
	Name    string
}

type Attachment struct {
	Name        string
	Content     []byte
	ContentType string
}

func NewEmail() *Email {
	return &Email{
		Headers: make(map[string]string),
	}
}

func (e *Email) SetFrom(address, name string) *Email {
	e.From = Mailbox{Address: address, Name: name}
	return e
}

func (e *Email) AddTo(address, name string) *Email {
	e.To = append(e.To, Mailbox{Address: address, Name: name})
	return e
}

func (e *Email) AddCc(address, name string) *Email {
	e.Cc = append(e.Cc, Mailbox{Address: address, Name: name})
	return e
}

func (e *Email) AddBcc(address, name string) *Email {
	e.Bcc = append(e.Bcc, Mailbox{Address: address, Name: name})
	return e
}

func (e *Email) SetSubject(subject string) *Email {
	e.Subject = subject
	return e
}

func (e *Email) SetBody(body string) *Email {
	e.Body = body
	return e
}

func (e *Email) SetHtml(html bool) *Email {
	e.Html = html
	return e
}

func (e *Email) AddAttachment(name string, content []byte, contentType string) *Email {
	e.Attachments = append(e.Attachments, Attachment{
		Name:        name,
		Content:     content,
		ContentType: contentType,
	})
	return e
}

func (e *Email) AddHeader(key, value string) *Email {
	e.Headers[key] = value
	return e
}

type SMTPTransport struct {
	host     string
	port     int
	username string
	password string
}

func NewSMTPTransport(host string, port int, username, password string) *SMTPTransport {
	return &SMTPTransport{
		host:     host,
		port:     port,
		username: username,
		password: password,
	}
}

func (t *SMTPTransport) Send(email *Email) error {
	addr := fmt.Sprintf("%s:%d", t.host, t.port)

	headers := make(map[string]string)
	headers["From"] = formatMailbox(email.From)
	headers["To"] = formatMailboxes(email.To)
	if len(email.Cc) > 0 {
		headers["Cc"] = formatMailboxes(email.Cc)
	}
	headers["Subject"] = email.Subject
	headers["MIME-Version"] = "1.0"

	if email.Html {
		headers["Content-Type"] = "text/html; charset=\"utf-8\""
	} else {
		headers["Content-Type"] = "text/plain; charset=\"utf-8\""
	}

	for key, value := range email.Headers {
		headers[key] = value
	}

	var msg bytes.Buffer
	for key, value := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	msg.WriteString("\r\n")
	msg.WriteString(email.Body)

	recipients := make([]string, 0)
	for _, to := range email.To {
		recipients = append(recipients, to.Address)
	}
	for _, cc := range email.Cc {
		recipients = append(recipients, cc.Address)
	}

	var auth smtp.Auth
	if t.username != "" {
		auth = smtp.PlainAuth("", t.username, t.password, t.host)
	}

	return smtp.SendMail(addr, auth, email.From.Address, recipients, msg.Bytes())
}

func formatMailbox(m Mailbox) string {
	if m.Name == "" {
		return m.Address
	}
	return fmt.Sprintf("%s <%s>", m.Name, m.Address)
}

func formatMailboxes(mailboxes []Mailbox) string {
	result := ""
	for i, m := range mailboxes {
		if i > 0 {
			result += ", "
		}
		result += formatMailbox(m)
	}
	return result
}

type SendmailTransport struct {
	command string
}

func NewSendmailTransport(command string) *SendmailTransport {
	return &SendmailTransport{command: command}
}

func (t *SendmailTransport) Send(email *Email) error {
	return errors.New("sendmail transport is not supported in this environment")
}

type FileTransport struct {
	dir string
}

func NewFileTransport(dir string) *FileTransport {
	return &FileTransport{dir: dir}
}

func (t *FileTransport) Send(email *Email) error {
	return errors.New("file transport is not supported in this environment")
}

type MemoryTransport struct {
	emails []*Email
}

func NewMemoryTransport() *MemoryTransport {
	return &MemoryTransport{
		emails: make([]*Email, 0),
	}
}

func (t *MemoryTransport) Send(email *Email) error {
	t.emails = append(t.emails, email)
	return nil
}

func (t *MemoryTransport) GetEmails() []*Email {
	return t.emails
}

func (t *MemoryTransport) Clear() {
	t.emails = make([]*Email, 0)
}

type TransportManager struct {
	transports map[string]MailerInterface
}

func NewTransportManager() *TransportManager {
	return &TransportManager{
		transports: make(map[string]MailerInterface),
	}
}

func (m *TransportManager) Add(name string, transport MailerInterface) {
	m.transports[name] = transport
}

func (m *TransportManager) Get(name string) MailerInterface {
	return m.transports[name]
}

func encodeRFC2047(word string) string {
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(word)) + "?="
}
