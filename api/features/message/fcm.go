package message

import (
	//"errors"
	"fmt"
	"myproject/api/database"

	//"myproject/api/features/service"
	"myproject/api/models"
	"os"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

var packageName = "org.myproject.app"

func sendToTopic(packageName string,
	topic string, data map[string]string,
	notification *messaging.Notification,
	force bool) error {

	ctx := context.Background()

	opt := option.WithCredentialsFile("./firebase_myproject_auth.json")

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		fmt.Println("error creating firebase app", err)
		return err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		fmt.Println("error creating messaging client", err)
		return err
	}

	duration, _ := time.ParseDuration("10m")

	title := "Data message"
	body := "Empty message"

	message := &messaging.Message{
		Data:  data,
		Topic: topic,
		Android: &messaging.AndroidConfig{
			Priority:              "high",
			TTL:                   &duration,
			RestrictedPackageName: packageName,
		},
		/* 	APNS: &messaging.APNSConfig{

		}, */
		/* 	FCMOptions: &messaging.FCMOptions{

		}, */
	}

	if notification != nil {
		title = notification.Title
		body = notification.Body

		message = &messaging.Message{
			Data:         data,
			Notification: notification,
			Topic:        topic,
			Android: &messaging.AndroidConfig{
				Priority:              "high",
				TTL:                   &duration,
				RestrictedPackageName: packageName,
				Notification: &messaging.AndroidNotification{
					Sound:       "default",
					ClickAction: "FLUTTER_NOTIFICATION_CLICK",
					//ClickAction: "CLICK_ACTIVITY",
					Priority: messaging.PriorityMax,
				},
			},
			/* 	APNS: &messaging.APNSConfig{

			}, */
			/* 	FCMOptions: &messaging.FCMOptions{

			}, */
		}
	}

	if os.Getenv("USE_DOCKER") == "true" && !force {
		fmt.Println("sendToTopic docker mode - not sending message to topic:", message.Topic, title, body)
		return nil
	}

	if (os.Getenv("TEST_MODE") == "true" || database.GetParam("DEV_MODE") == "true") && !force {
		fmt.Println("sendToTopic test/dev mode - not sending message to topic:", message.Topic, title, body)
		return nil
	}

	response, err := client.Send(ctx, message)
	if err != nil {
		fmt.Println("error sending message to topic:", err, message.Topic, title, body)
		return err
	}
	// Response is a message ID string.
	if os.Getenv("TEST_MODE") == "true" || database.GetParam("DEV_MODE") == "true" {
		fmt.Println("sendToTopic Successfully sent message:", response, message.Topic, title, body)
	}
	return nil
}

func SendDataToTopic(packageName string,
	topic string, data map[string]string,
	force bool) error {
	return sendToTopic(packageName, topic, data, nil, force)
}

func SendPNtoTopic(packageName string,
	topic string, title string,
	content string, data map[string]string,
	force bool) error {

	notification := &messaging.Notification{
		Title: title,
		Body:  content,
	}
	return sendToTopic(packageName, topic, data, notification, force)
}

func SendPushNotification(packageName string,
	token string, title string,
	content string, data map[string]string, force bool) error {

	if token == "" {
		fmt.Println("SendPushNotification - not sending message to empty token:", title, content)
		return nil
	}
	ctx := context.Background()

	opt := option.WithCredentialsFile("./firebase_myproject_auth.json")

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		fmt.Println("error creating firebase app", err)
		return err
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		fmt.Println("error creating messaging client", err)
		return err
	}

	duration, _ := time.ParseDuration("10m")

	message := &messaging.Message{
		Data: data,
		Notification: &messaging.Notification{
			Title: title,
			Body:  content,
		},
		Token: token,
		Android: &messaging.AndroidConfig{
			Priority:              "high",
			TTL:                   &duration,
			RestrictedPackageName: packageName,
			Notification: &messaging.AndroidNotification{
				Sound:       "default",
				ClickAction: "FLUTTER_NOTIFICATION_CLICK",
				//ClickAction: "CLICK_ACTIVITY",
				Priority:              messaging.PriorityMax,
				DefaultVibrateTimings: true,
				//DefaultSound:          false,
			},
		},
		/* 	APNS: &messaging.APNSConfig{

		}, */
		/* 	FCMOptions: &messaging.FCMOptions{

		}, */
	}

	if os.Getenv("USE_DOCKER") == "true" && !force {
		fmt.Println("SendPushNotification docker mode - not sending message to token:", message.Token, message.Notification.Title, message.Notification.Body)
		return nil
	}

	if (os.Getenv("TEST_MODE") == "true" || database.GetParam("DEV_MODE") == "true") && !force {
		fmt.Println("SendPushNotification test mode - would send message to token:", message.Token, message.Notification.Title, message.Notification.Body)
		return nil
	}

	response, err := client.Send(ctx, message)
	if err != nil {
		fmt.Println("error sending message to token:", err, message.Token, message.Notification.Title, message.Notification.Body)
		return err
	}
	// Response is a message ID string.
	if os.Getenv("TEST_MODE") == "true" || database.GetParam("DEV_MODE") == "true" {
		fmt.Println("SendPushNotification Successfully sent message:", response, message.Topic, title)
	}
	return nil
}

func EmitMapPushNotification(data map[string]string, token string, topic string, title string, content string, force bool) error {

	if content == "" {
		content = "A customer requires service for an item of equipment.  Please open myproject and respond promptly to the request"
	}

	if token != "" {
		// emit to a token
		if err := SendPushNotification(packageName, token, title,
			content, data,
			force,
		); err != nil {
			fmt.Println(err)

			return err
		}
	} else {
		// send to a topic

		topic = strings.ReplaceAll(topic, " ", "")

		if err := SendPNtoTopic(packageName, topic, title, content, data, force); err != nil {
			fmt.Println(err)

			return err
		}

	}
	return nil
}

func SendMessageToUserMobileApp(db *gorm.DB, user models.User, title, content string, data map[string]string, force bool) error {

	var envPackage = database.GetParam("PACKAGE_NAME")

	packageName := "org.myproject.app"
	if envPackage != "" {
		packageName = envPackage
	}

	var uapp models.UserApp
	result := db.First(&uapp, "user_id = ? and package = ?", user.ID, packageName)

	if result.Error != nil {
		return result.Error
	}

	if uapp.Token != "" {
		fmt.Println("Send PN to token", uapp.Token, "for user", user.Name)
	}

	return SendPushNotification(packageName, uapp.Token, title, content, data, force)
}
