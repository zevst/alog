////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package email

import (
	"github.com/golang/mock/gomock"
	"github.com/mylockerteam/alog/mocks"
	"html/template"
	"runtime/debug"
	"testing"

	"github.com/mylockerteam/mailSender"
	"gopkg.in/gomail.v2"
)

func TestEmailStrategy_Write(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	type args struct {
		p []byte
	}

	from := "no-reply@example.com"
	to := []string{"<test@example.com>"}

	msg := gomail.NewMessage()
	msg.SetHeader("From", "Example <no-reply@example.com>")
	msg.SetHeader("Bcc", to...)
	msg.SetHeader("Subject", "Debug message")

	mockSendCloser := mocks.NewMockSendCloser(mockCtrl)
	mockSendCloser.EXPECT().Send(from, []string{"test@example.com"}, msg).Return(nil).AnyTimes()

	sender := mailSender.Create(&mailSender.Sender{
		Channel: make(chan mailSender.Message, 1),
		Closer:  mockSendCloser,
	})

	tpl, _ := template.New("test").Parse("<pre><code>{{ .Data }}</code></pre>")

	strategy := Get(sender, msg, tpl)
	stack := debug.Stack()
	tests := []struct {
		name    string
		fields  *Strategy
		args    args
		wantN   int
		wantErr bool
	}{
		{
			fields: strategy.(*Strategy),
			args: args{
				p: stack,
			},
			wantN:   len(stack),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Strategy{
				Writer:   tt.fields.Writer,
				sender:   tt.fields.sender,
				Message:  tt.fields.Message,
				Template: tt.fields.Template,
			}
			gotN, err := s.Write(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("EmailStrategy.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotN != tt.wantN {
				t.Errorf("EmailWrite() = %v, want %v", gotN, tt.wantN)
			}
		})
	}
}
