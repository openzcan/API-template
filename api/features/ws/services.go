package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"myproject/api/database"
	"myproject/api/models"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

func Publish(rdb *redis.Client, channel string, content string) error {
	ctx := context.Background()

	fmt.Println("publish on channel", channel, content)
	return rdb.Publish(ctx, channel, content).Err()
}

func hasChannel(rdb *redis.Client, channel string) (bool, error) {
	ctx := context.Background()

	result := rdb.PubSubNumSub(ctx, channel)

	if result.Err() != nil {
		return false, result.Err()
	}

	mapResult := result.Val()

	return mapResult[channel] > 0, nil

}

func PublishToChannel(channel string, message *models.WebsocketMessage) error {

	cfg, err := database.LoadConfig()

	rdb, _ := database.ConnectRedis(cfg.Redis)

	// free up the redis connection when we are done
	defer rdb.Close()

	// check the channel is live
	if exists, err := hasChannel(rdb, channel); err != nil {
		fmt.Println("PublishToChannel Error", err)
		return err
	} else if !exists {
		fmt.Println("PublishToChannel channel", channel, "does not exist")
		return errors.New("channel does not exist")
	}

	//fmt.Println("publish to", message)
	pkt, err := json.Marshal(message)

	if err != nil {
		fmt.Println("PublishToChannel Error", err)
		return err
	}

	if err := Publish(rdb, channel, string(pkt[:])); err != nil {
		fmt.Println("PublishToChannel Error", err)
		return err
	}

	return nil
}

func PublishToChannels(message *models.MultiChannelMessage) error {

	// publish a message to redis

	cfg, _ := database.LoadConfig()

	rdb, _ := database.ConnectRedis(cfg.Redis)

	fmt.Println("publish", message.Subject, "to", message.Channels)

	for _, channel := range message.Channels {

		msg := models.WebsocketMessage{
			Payload: message.Payload,
			Channel: channel,
			Event:   message.Event,
			Subject: message.Subject,
			Uuid:    message.Uuid,
		}
		pkt, err := json.Marshal(msg)

		if err != nil {
			fmt.Println("Error", err)
			// free up the redis connection
			rdb.Close()
			return err
		}

		if err := Publish(rdb, channel, string(pkt[:])); err != nil {
			fmt.Println("Error", err)
			// free up the redis connection
			rdb.Close()
			return err
		}
	}

	// free up the redis connection
	rdb.Close()

	return nil
}
