package message

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"myproject/api/database"
	"myproject/api/models"
	"net/smtp"
	"os"
	"strconv"
	"time"

	myk "github.com/domodwyer/mailyak"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"gorm.io/gorm"
)

func SendEmailWithMailyak(msg *models.EmailMessage, params interface{}) error {
	smtp_host := database.GetParam("SMTP_HOST")
	smtp_username := database.GetParam("SMTP_USERNAME")
	smtp_password := database.GetParam("SMTP_PASSWORD")
	smtp_port := database.GetParam("SMTP_PORT")

	if smtp_host == "" || smtp_username == "" || smtp_password == "" || smtp_port == "" {
		fmt.Println("environment has no SMTP credentials, not sending email")
		return nil
	}

	mail := myk.New(fmt.Sprintf("%s:%s", smtp_host, smtp_port), smtp.PlainAuth("", smtp_username, smtp_password, smtp_host))

	mail.To(msg.Email)
	mail.From(msg.From)
	mail.FromName(msg.FromName)

	mail.Subject(msg.Subject)

	mail.AddHeader("Message-Id", fmt.Sprintf("<%d@%s>", time.Now().UnixNano(), smtp_host))

	// Compile a template
	tmpl, err := template.ParseFiles(msg.Template)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Execute the template directly into the email body
	if err := tmpl.Execute(mail.HTML(), params); err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("send email to", msg.To, msg.Email, "from", msg.From, msg.FromName)

	if os.Getenv("TEST_MODE") == "true" || database.GetParam("DEV_MODE") == "true" {
		if msg.Email != "damianham@gmail.com" {
			fmt.Println("TEST_MODE - Not sending email to", msg.To, msg.Email, "from", msg.From, msg.FromName)
			return nil
		} else {
			fmt.Println("TEST_MODE - sending email to dev only", msg.To, msg.Email, "from", msg.From, msg.FromName)
		}
	}

	if err := mail.Send(); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func SendEmailViaSendgrid(msg *models.EmailMessage, params interface{}) error {

	fmt.Println("Subject", msg.Subject)
	fmt.Println("to", msg.To)
	fmt.Println("from", msg.From)
	fmt.Println("fromName", msg.FromName)
	fmt.Println("email", msg.Email)
	fmt.Println("text", msg.Text)

	if os.Getenv("TEST_MODE") == "true" || database.GetParam("DEV_MODE") == "true" {
		if msg.Email != "damianham@gmail.com" {
			fmt.Println("TEST_MODE - Not sending email to", msg.To, msg.Email, "from", msg.From, msg.FromName)
			return nil
		} else {
			fmt.Println("TEST_MODE - sending email to dev only", msg.To, msg.Email, "from", msg.From, msg.FromName)
		}
	}

	from := mail.NewEmail("MyAPI Team", "noreply@mydomain.com")
	subject := msg.Subject
	to := mail.NewEmail(msg.To, msg.Email)
	plainTextContent := msg.Text

	// Compile a template
	tmpl, err := template.ParseFiles(msg.Template)
	if err != nil {
		fmt.Println(err)
		return err
	}

	var buffer bytes.Buffer
	// Execute the template directly into the email body
	if err := tmpl.Execute(&buffer, params); err != nil {
		fmt.Println(err)
		return err
	}

	htmlContent := buffer.String()

	//t.Println("html", htmlContent)

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(database.GetParam("SENDGRID_API_KEY"))
	response, err := client.Send(message)
	if err != nil {
		log.Println(err)
		return err
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}

	return nil
}

func SendTelegramMessage(chatID int64, message string) error {
	botToken := database.GetParam("TELEGRAM_BOT_APIKEY")
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return fmt.Errorf("failed to create telegram bot: %w", err)
	}

	msg := tgbotapi.NewMessage(chatID, message)
	_, err = bot.Send(msg)
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	return nil
}

func SendMessageToUserIDViaTelegram(db *gorm.DB, id, message string) error {
	user := models.User{}
	db.First(&user, id)
	if user.ID == 0 {
		return fmt.Errorf("user not found")
	}

	return SendMessageToUserViaTelegram(db, &user, message)
}

func SendMessageToUserViaTelegram(db *gorm.DB, user *models.User, message string) error {

	chatID, err := strconv.ParseInt(user.Telegram, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse telegram chat ID: %w", err)
	}

	return SendTelegramMessage(chatID, message)
}
