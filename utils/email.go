package utils

import (
	"fmt"
	"log"
	"massage-scheduler/models"
	"net/smtp"
	"os"
)

func SendAdminNotification(newUser models.User) {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "adm@teste.com" // Fallback baseado no seedAdmin
	}

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

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
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, []string{adminEmail}, msg)
	if err != nil {
		log.Printf("Erro ao enviar e-mail de notificação: %v", err)
	} else {
		log.Printf("E-mail de notificação enviado para o admin (%s) sobre o usuário %s", adminEmail, newUser.Email)
	}
}
