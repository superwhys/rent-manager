package emailcode

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"errors"
	"io"
	"math/big"
	"net"
	"net/mail"
	"net/textproto"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/superwhys/rent-manager/config"
)

func TestSendCodeUsesEncryptedSMTP(t *testing.T) {
	for _, directTLS := range []bool{false, true} {
		t.Run(strconv.FormatBool(directTLS), func(t *testing.T) {
			sender, result := newTestSender(t, smtpScenario{directTLS: directTLS, offerTLS: true})
			if err := sender.SendCode(
				context.Background(),
				"tenant@example.com",
				"123456",
			); err != nil {
				t.Fatal(err)
			}
			received := <-result
			if received.err != nil {
				t.Fatal(received.err)
			}
			if !received.authEncrypted ||
				received.auth != "\x00notice@example.com\x00smtp-password" {
				t.Fatal("SMTP 身份验证必须在 TLS 内完成")
			}
			if received.from != "FROM:<notice@example.com> BODY=8BITMIME" &&
				received.from != "FROM:<notice@example.com>" {
				t.Fatalf("unexpected sender: %s", received.from)
			}
			if received.to != "TO:<tenant@example.com>" {
				t.Fatalf("unexpected recipient: %s", received.to)
			}
			message, err := mail.ReadMessage(strings.NewReader(received.message))
			if err != nil {
				t.Fatal(err)
			}
			body, err := io.ReadAll(message.Body)
			if err != nil || !strings.Contains(string(body), "123456") ||
				!strings.Contains(string(body), "10 分钟") {
				t.Fatal("邮件应包含验证码和有效期")
			}
		})
	}
}

func TestSendCodeRejectsUnencryptedSMTP(t *testing.T) {
	sender, result := newTestSender(t, smtpScenario{})
	if err := sender.SendCode(
		context.Background(),
		"tenant@example.com",
		"123456",
	); !errors.Is(
		err,
		ErrDeliveryFailed,
	) {
		t.Fatalf("unexpected error: %v", err)
	}
	received := <-result
	if received.auth != "" || received.message != "" {
		t.Fatal("不能向未启用 TLS 的服务器发送凭据或验证码")
	}
}

func TestSendCodeRejectsUntrustedCertificate(t *testing.T) {
	sender, result := newTestSender(t, smtpScenario{offerTLS: true})
	sender.tlsConfig.RootCAs = x509.NewCertPool()
	if err := sender.SendCode(
		context.Background(),
		"tenant@example.com",
		"123456",
	); !errors.Is(
		err,
		ErrDeliveryFailed,
	) {
		t.Fatalf("unexpected error: %v", err)
	}
	received := <-result
	if received.auth != "" || received.message != "" {
		t.Fatal("不能向证书不受信任的服务器发送凭据或验证码")
	}
}

func TestSendCodeHidesSMTPFailureDetails(t *testing.T) {
	sender, result := newTestSender(t, smtpScenario{offerTLS: true, rejectData: true})
	err := sender.SendCode(context.Background(), "tenant@example.com", "123456")
	if !errors.Is(err, ErrDeliveryFailed) || strings.Contains(err.Error(), "123456") ||
		strings.Contains(err.Error(), "smtp-password") {
		t.Fatalf("SMTP 错误不应透出邮件或凭据: %v", err)
	}
	<-result
}

func TestSendCodeAcceptedMessageSurvivesQuitFailure(t *testing.T) {
	sender, result := newTestSender(t, smtpScenario{offerTLS: true, rejectQuit: true})
	if err := sender.SendCode(context.Background(), "tenant@example.com", "123456"); err != nil {
		t.Fatalf("邮件已接受后不应误报发送失败: %v", err)
	}
	<-result
}

func TestSendCodeCancellationClosesConnection(t *testing.T) {
	for _, explicitCancel := range []bool{false, true} {
		t.Run(strconv.FormatBool(explicitCancel), func(t *testing.T) {
			sender, err := NewSender(config.SMTPConfig{
				Host:    "smtp.example.com",
				Port:    587,
				From:    "notice@example.com",
				Timeout: 30 * time.Millisecond,
			})
			if err != nil {
				t.Fatal(err)
			}
			client, server := net.Pipe()
			defer server.Close()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			sender.dialContext = func(context.Context, string, string) (net.Conn, error) {
				if explicitCancel {
					cancel()
				}
				return client, nil
			}
			start := time.Now()
			err = sender.SendCode(ctx, "tenant@example.com", "123456")
			if explicitCancel && !errors.Is(err, context.Canceled) {
				t.Fatalf("取消后应返回 context.Canceled: %v", err)
			}
			if !explicitCancel && !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("超时应返回 context.DeadlineExceeded: %v", err)
			}
			if time.Since(start) > time.Second {
				t.Fatal("取消或超时不能继续等待 SMTP")
			}
			if _, err := server.Read(make([]byte, 1)); !errors.Is(err, io.EOF) {
				t.Fatalf("取消或超时应关闭连接: %v", err)
			}
		})
	}
}

func TestSendCodeRejectsRecipientHeaderInjectionBeforeDial(t *testing.T) {
	sender, err := NewSender(
		config.SMTPConfig{Host: "smtp.example.com", Port: 587, From: "notice@example.com"},
	)
	if err != nil {
		t.Fatal(err)
	}
	sender.dialContext = func(context.Context, string, string) (net.Conn, error) {
		t.Fatal("无效收件地址不应发起连接")
		return nil, nil
	}
	for _, email := range []string{"invalid", "tenant@example.com\r\nBcc: other@example.com", "Tenant <tenant@example.com>"} {
		if err := sender.SendCode(
			context.Background(),
			email,
			"123456",
		); !errors.Is(
			err,
			ErrDeliveryFailed,
		) {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}

type smtpScenario struct {
	directTLS  bool
	offerTLS   bool
	rejectData bool
	rejectQuit bool
}

type smtpResult struct {
	authEncrypted bool
	auth          string
	from          string
	to            string
	message       string
	err           error
}

func newTestSender(t *testing.T, scenario smtpScenario) (*Sender, <-chan smtpResult) {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	certificate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "smtp.example.com"},
		DNSNames: []string{
			"smtp.example.com",
		},
		NotBefore:   time.Now().Add(-time.Hour),
		NotAfter:    time.Now().Add(time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, certificate, certificate, publicKey, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	serverTLS := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: privateKey}},
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	result := make(chan smtpResult, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		result <- serveSMTP(listener, serverTLS, scenario)
	}()
	t.Cleanup(func() {
		_ = listener.Close()
		<-done
	})
	port := 587
	if scenario.directTLS {
		port = 465
	}
	sender, err := NewSender(config.SMTPConfig{
		Host: "smtp.example.com", Port: port, From: "notice@example.com", FromName: "租住管理",
		Username: "notice@example.com", Password: "smtp-password", Timeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	sender.tlsConfig.RootCAs = x509.NewCertPool()
	sender.tlsConfig.RootCAs.AddCert(leaf)
	sender.dialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, listener.Addr().String())
	}
	return sender, result
}

func serveSMTP(
	listener net.Listener,
	tlsConfig *tls.Config,
	scenario smtpScenario,
) (result smtpResult) {
	conn, err := listener.Accept()
	if err != nil {
		result.err = err
		return
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
	encrypted := scenario.directTLS
	if encrypted {
		conn = tls.Server(conn, tlsConfig)
	}
	stream := textproto.NewConn(conn)
	_ = stream.PrintfLine("220 smtp.example.com ESMTP")
	for {
		line, err := stream.ReadLine()
		if err != nil {
			return
		}
		command, arg, _ := strings.Cut(line, " ")
		switch command {
		case "EHLO":
			_ = stream.PrintfLine("250-smtp.example.com")
			if !encrypted && scenario.offerTLS {
				_ = stream.PrintfLine("250-STARTTLS")
			}
			_ = stream.PrintfLine("250 AUTH PLAIN")
		case "STARTTLS":
			_ = stream.PrintfLine("220 Ready")
			conn = tls.Server(conn, tlsConfig)
			stream = textproto.NewConn(conn)
			encrypted = true
		case "AUTH":
			_, encoded, _ := strings.Cut(arg, " ")
			decoded, _ := base64.StdEncoding.DecodeString(encoded)
			result.auth, result.authEncrypted = string(decoded), encrypted
			_ = stream.PrintfLine("235 Authenticated")
		case "MAIL":
			result.from = arg
			_ = stream.PrintfLine("250 Accepted")
		case "RCPT":
			result.to = arg
			_ = stream.PrintfLine("250 Accepted")
		case "DATA":
			_ = stream.PrintfLine("354 Start mail")
			body, err := stream.ReadDotBytes()
			if err != nil {
				result.err = err
				return
			}
			result.message = string(body)
			if scenario.rejectData {
				_ = stream.PrintfLine("554 Rejected 123456 smtp-password")
			} else {
				_ = stream.PrintfLine("250 Queued")
			}
		case "QUIT":
			if scenario.rejectQuit {
				_ = stream.PrintfLine("550 Unavailable")
			} else {
				_ = stream.PrintfLine("221 Bye")
			}
			return
		default:
			_ = stream.PrintfLine("500 Unsupported")
		}
	}
}
