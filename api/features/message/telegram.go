package message

import (
	"context"
	"errors"
	"fmt"
	"log"
	"myproject/api/database"
	"myproject/api/models"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gorm.io/gorm"
)

var (
	// Menu texts
	firstMenu  = "<b>Menu 1</b>\n\nA beautiful menu with a shiny inline button."
	secondMenu = "<b>Menu 2</b>\n\nA better menu with even more shiny inline buttons."

	welcomeMessage       = "Welcome to the bot. Please use the following command to start getting notifications:\n\n/start (your phone number) - Start the bot and identify yourself with your phone number including ISO country code.\n\nE.g. /start +12223334444"
	notRecognisedMessage = "I don't recognise who you are. Please try again."

	// Button texts
	nextButton     = "Next"
	backButton     = "Back"
	tutorialButton = "Tutorial"

	// Store bot screaming status
	screaming = false
	bot       *tgbotapi.BotAPI

	// Keyboard layout for the first menu. One button, one row
	firstMenuMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(nextButton, nextButton),
		),
	)

	// Keyboard layout for the second menu. Two buttons, one per row
	secondMenuMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(backButton, backButton),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(tutorialButton, "https://core.telegram.org/bots/api"),
		),
	)
)

func RunTelegramBot(db *gorm.DB) {
	var err error
	botToken := database.GetParam("TELEGRAM_BOT_APIKEY")

	if botToken == "" {
		log.Panic("RunTelegramBot:No bot token provided")
	}

	// Create a new bot
	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Println("RunTelegramBot:Error creating bot: ", botToken)
		// Abort if something is wrong
		log.Panic(err)
	}

	//fmt.Println("running telegram bot is Parent", !fiber.IsChild())

	// Set this to true to log all interactions with telegram servers
	bot.Debug = false

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Create a new background context.
	ctx := context.Background()

	// `updates` is a golang channel which receives telegram updates
	updates := bot.GetUpdatesChan(u)

	// Pass background context to goroutine
	receiveUpdates(ctx, updates, db)

	// Tell the user the bot is online
	log.Println("RunTelegramBot:Start listening for updates.")

}

func receiveUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel, db *gorm.DB) {
	// `for {` means the loop is infinite until we manually stop it
	for {
		select {
		// stop looping if ctx is cancelled
		case <-ctx.Done():
			return
		// receive update from channel and then handle it
		case update := <-updates:
			handleUpdate(update, db)
		}
	}
}

func handleUpdate(update tgbotapi.Update, db *gorm.DB) {
	switch {
	// Handle messages
	case update.Message != nil:
		handleMessage(update.Message, db)
		break

	// Handle button clicks
	case update.CallbackQuery != nil:
		handleButton(update.CallbackQuery)
		break
	}
}

func handleMessage(message *tgbotapi.Message, db *gorm.DB) {
	user := message.From
	text := message.Text

	if user == nil {
		return
	}

	// Print to console
	log.Printf("%s wrote %s", user.FirstName, text)

	var err error
	if strings.HasPrefix(text, "/") {
		err = handleCommand(message.Chat.ID, text, user, db)
	} else if screaming && len(text) > 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, strings.ToUpper(text))
		// To preserve markdown, we attach entities (bold, italic..)
		msg.Entities = message.Entities
		_, err = bot.Send(msg)
	} else {
		// This is equivalent to forwarding, without the sender's name
		err = sendWelcomeMessage(message.Chat.ID)
	}

	if err != nil {
		log.Printf("An error occured: %s", err.Error())
	}
}

func findUserByNameAndPhone(db *gorm.DB, name, familyname string, phone string) (*models.User, error) {

	fullname := fmt.Sprintf("%s %s", name, familyname)
	// find the user
	var user models.User
	result := db.Where("phone = ? and  lower(name) = ?", phone, strings.ToLower(fullname)).First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		fmt.Println("User not found", fullname, phone)
		result = db.Where("phone = ? ", phone).First(&user)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
	}

	return &user, nil
}

// When we get a command, we react accordingly
func handleCommand(chatId int64, text string, user *tgbotapi.User, db *gorm.DB) error {
	var err error

	var res *models.User

	if strings.HasPrefix(text, "/start") {
		res, err = findUserByNameAndPhone(db, user.FirstName, user.LastName, strings.TrimPrefix(text, "/start "))

		if res == nil {
			err = sendNotRecognisedMessage(chatId)
		} else {
			res.Telegram = fmt.Sprintf("%d", chatId)
			err = db.Save(res).Error
			if err != nil {
				log.Printf("An error occured: %s", err.Error())
				err = sendNotRecognisedMessage(chatId)
			} else {
				err = sendDoneMessage(chatId)
			}
		}
	} else {
		err = sendWelcomeMessage(chatId)
	}

	return err
}

func handleButton(query *tgbotapi.CallbackQuery) {
	var text string

	markup := tgbotapi.NewInlineKeyboardMarkup()
	message := query.Message

	if query.Data == nextButton {
		text = secondMenu
		markup = secondMenuMarkup
	} else if query.Data == backButton {
		text = firstMenu
		markup = firstMenuMarkup
	}

	callbackCfg := tgbotapi.NewCallback(query.ID, "")
	bot.Send(callbackCfg)

	// Replace menu text and keyboard
	msg := tgbotapi.NewEditMessageTextAndMarkup(message.Chat.ID, message.MessageID, text, markup)
	msg.ParseMode = tgbotapi.ModeHTML
	bot.Send(msg)
}

func sendMenu(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, firstMenu)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = firstMenuMarkup
	_, err := bot.Send(msg)
	return err
}

func sendWelcomeMessage(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, welcomeMessage)
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := bot.Send(msg)
	return err
}

func sendNotRecognisedMessage(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, notRecognisedMessage)
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := bot.Send(msg)
	return err
}

func sendDoneMessage(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, "Welcome.  You will now be able to receive messages from the myproject system via telegram.")
	msg.ParseMode = tgbotapi.ModeHTML
	_, err := bot.Send(msg)
	return err
}
