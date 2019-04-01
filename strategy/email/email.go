////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package email

import (
	"github.com/mylockerteam/mailSender"
	"gopkg.in/gomail.v2"
	"html/template"
	"io"
)

// Strategy logging strategy in the email
// You can use it for errors and other types of messages
type Strategy struct {
	sender   mailSender.AsyncSender
	Message  *gomail.Message
	Template *template.Template
	io.Writer
}

//Get waiting for a parameter ess in format host:port;username;password
func Get(sender mailSender.AsyncSender, msg *gomail.Message, tpl *template.Template) io.Writer {
	return &Strategy{
		sender:   sender,
		Message:  msg,
		Template: tpl,
	}
}

func (s *Strategy) Write(p []byte) (n int, err error) {
	s.sender.SendAsync(mailSender.Message{
		Message:  s.Message,
		Template: s.Template,
		Data:     mailSender.EmailData{"Data": string(p)},
	})
	return len(p), nil
}
