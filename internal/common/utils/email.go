package common_utils

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"massage-scheduler/internal/common/models"
	"net"
	"net/smtp"
	"os"
	"sync"
	"time"
)

const smtpTimeout = 15 * time.Second

var (
	notifyOnce sync.Once
	notifyCh   chan common_models.User
	notifyWG   sync.WaitGroup
	notifyStop sync.Once
)

func ensureNotificationWorker() {
	notifyOnce.Do(func() {
		notifyCh = make(chan common_models.User, 64)
		notifyWG.Add(1)
		go func() {
			defer notifyWG.Done()
			for u := range notifyCh {
				SendAdminNotification(u)
			}
		}()
	})
}

func EnqueueAdminNotification(newUser common_models.User) {
	ensureNotificationWorker()
	select {
	case notifyCh <- newUser:
	default:
		log.Printf("Fila de notificações cheia; pulando e-mail para novo usuário: %s", newUser.Email)
	}
}

func ShutdownNotifications(ctx context.Context) {
	if notifyCh == nil {
		return
	}
	notifyStop.Do(func() {
		close(notifyCh)
	})
	done := make(chan struct{})
	go func() {
		notifyWG.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
		log.Println("Timeout aguardando envio de notificações pendentes.")
	}
}

func sendMailWithTimeout(addr, host string, auth smtp.Auth, from string, to []string, msg []byte) error {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return err
	}
	_ = conn.SetDeadline(time.Now().Add(smtpTimeout))

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		_ = conn.Close()
		return err
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: host}); err != nil {
			return err
		}
	}
	if auth != nil {
		if err := c.Auth(auth); err != nil {
			return err
		}
	}
	if err := c.Mail(from); err != nil {
		return err
	}
	for _, rcpt := range to {
		if err := c.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

func SendAdminNotification(newUser common_models.User) {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "edynobre@gmail.com" // Fallback baseado no seedAdmin
	}

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

	// Se nada estiver configurado, assume que a notificação por e-mail está desativada e não avisa.
	if smtpHost == "" && smtpPort == "" && smtpUser == "" && smtpPass == "" {
		return
	}

	// Se houver configuração parcial, avisa que está incompleta.
	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" {
		log.Printf("Configuração SMTP incompleta. Pulando envio de e-mail para novo usuário: %s", newUser.Email)
		return
	}

	subject := "Novo Cadastro no Sistema de Agenda"
	body := fmt.Sprintf("Olá Administrador,\n\nUm novo usuário se cadastrou no sistema:\n\nNome: %s\nE-mail: %s\nEmpresa: %s\nData: %s\n\nAtenciosamente,\nSistema de Agenda",
		newUser.Name, newUser.Email, newUser.Business, newUser.CreatedAt.Format("02/01/2006 15:04:05"))

	msg := []byte("To: " + adminEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	err := sendMailWithTimeout(net.JoinHostPort(smtpHost, smtpPort), smtpHost, auth, smtpUser, []string{adminEmail}, msg)
	if err != nil {
		log.Printf("Erro ao enviar e-mail de notificação: %v", err)
	} else {
		log.Printf("E-mail de notificação enviado para o admin (%s) sobre o usuário %s", adminEmail, newUser.Email)
	}
}
