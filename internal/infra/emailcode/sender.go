package emailcode

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/miebyte/authkit"

	"github.com/superwhys/rent-manager/config"
)

var ErrDeliveryFailed = errors.New("验证码邮件发送失败，请稍后重试")

type Sender struct {
	config      config.SMTPConfig
	tlsConfig   *tls.Config
	dialContext func(context.Context, string, string) (net.Conn, error)
}

var _ authkit.CodeSender = (*Sender)(nil)

func NewSender(cfg config.SMTPConfig) (*Sender, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if !cfg.Configured() {
		return nil, errors.New("尚未配置邮件发送服务")
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &Sender{
		config: cfg,
		tlsConfig: &tls.Config{
			ServerName: cfg.Host,
			MinVersion: tls.VersionTLS12,
		},
		dialContext: (&net.Dialer{}).DialContext,
	}, nil
}

func (s *Sender) SendCode(ctx context.Context, email, code string) error {
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || strings.ContainsAny(email, "\r\n") {
		return ErrDeliveryFailed
	}
	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()
	if err := s.send(ctx, email, code); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		// 连接 deadline 可能先于 Context 的定时回调触发。
		if deadline, ok := ctx.Deadline(); ok && !time.Now().Before(deadline) {
			return context.DeadlineExceeded
		}
		// SMTP 响应可能回显验证码或认证信息，不向调用方传递原始错误。
		return ErrDeliveryFailed
	}
	return nil
}

func (s *Sender) send(ctx context.Context, email, code string) error {
	conn, err := s.dialContext(
		ctx,
		"tcp",
		net.JoinHostPort(s.config.Host, strconv.Itoa(s.config.Port)),
	)
	if err != nil {
		return err
	}
	defer conn.Close()
	rawConn := conn
	stop := context.AfterFunc(ctx, func() { _ = rawConn.Close() })
	defer stop()
	deadline, _ := ctx.Deadline()
	if err := conn.SetDeadline(deadline); err != nil {
		return err
	}
	if s.config.Port == 465 {
		tlsConn := tls.Client(conn, s.tlsConfig)
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			return err
		}
		conn = tlsConn
	}
	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return err
	}
	defer client.Close()
	if s.config.Port != 465 {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return ErrDeliveryFailed
		}
		if err := client.StartTLS(s.tlsConfig); err != nil {
			return err
		}
	}
	if s.config.Username != "" {
		if err := client.Auth(
			smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host),
		); err != nil {
			return err
		}
	}
	if err := client.Mail(s.config.From); err != nil {
		return err
	}
	if err := client.Rcpt(email); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	from := (&mail.Address{Name: s.config.FromName, Address: s.config.From}).String()
	message := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n您的租住管理验证码是：%s\r\n\r\n验证码在 %d 分钟内有效，请勿向他人透露。若非本人操作，请忽略此邮件。\r\n",
		from,
		(&mail.Address{Address: email}).String(),
		mime.QEncoding.Encode("UTF-8", "租住管理 · 邮箱验证码"),
		code,
		int(authkit.CodeTTL/time.Minute),
	)
	if _, err := writer.Write([]byte(message)); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	// DATA 已被服务器接受；退出会话失败不能把已发送邮件误报为发送失败。
	_ = client.Quit()
	return nil
}
