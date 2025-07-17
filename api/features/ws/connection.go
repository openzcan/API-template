package ws

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/websocket/v2"
	"github.com/redis/go-redis/v9"
)

type Connection struct {
	name            string
	channelsHandler *redis.PubSub
	channels        []string

	stopListenerChan chan struct{}
	listening        bool

	MessageChan chan redis.Message
	conn        *websocket.Conn
}

// Connect connect user to user channels on redis
func Connect(rdb *redis.Client, name string, conn *websocket.Conn) (*Connection, error) {

	u := &Connection{
		name:             name,
		stopListenerChan: make(chan struct{}),
		MessageChan:      make(chan redis.Message),
		conn:             conn,
	}

	if err := u.connect(rdb); err != nil {
		return nil, err
	}

	return u, nil
}

func (u *Connection) Subscribe(rdb *redis.Client, channel string) error {

	u.channels = append(u.channels, channel)

	return u.connect(rdb)
}

func (u *Connection) Unsubscribe(rdb *redis.Client, channel string) error {

	// remove channel
	for i, v := range u.channels {
		if v == channel {
			u.channels = append(u.channels[:i], u.channels[i+1:]...)
			break
		}
	}

	return u.connect(rdb)
}

func (u *Connection) connect(rdb *redis.Client) error {
	ctx := context.Background()

	if u.channelsHandler != nil {
		if err := u.channelsHandler.Unsubscribe(ctx); err != nil {
			return err
		}
		if err := u.channelsHandler.Close(); err != nil {
			return err
		}
	}

	if u.listening {
		u.stopListenerChan <- struct{}{}
	}

	return u.doConnect(rdb, u.channels...)
}

func (u *Connection) doConnect(rdb *redis.Client, channels ...string) error {
	ctx := context.Background()

	// subscribe all channels in one request
	pubSub := rdb.Subscribe(ctx, channels...)
	// keep channel handler to be used in unsubscribe
	u.channelsHandler = pubSub

	// The Listener
	go func() {
		u.listening = true
		if len(channels) > 0 {
			fmt.Println("starting the listener for client:", u.name, "on channels:", channels)
		}
		for {
			select {
			case msg, ok := <-pubSub.Channel():
				if !ok {
					return
				}
				//u.MessageChan <- *msg

				fmt.Println("write message to client:", u.name, channels)

				if err := u.conn.WriteJSON(msg); err != nil {
					fmt.Println(err)
				}

			case <-u.stopListenerChan:
				//fmt.Println("stopping the listener for client:", u.name)
				return
			}
		}
	}()
	return nil
}

func (u *Connection) Disconnect() error {
	ctx := context.Background()

	if u.channelsHandler != nil {
		if err := u.channelsHandler.Unsubscribe(ctx); err != nil {
			return err
		}
		if err := u.channelsHandler.Close(); err != nil {
			return err
		}
	}
	if u.listening {
		u.stopListenerChan <- struct{}{}
	}

	//fmt.Println("disconnect", u.name)

	close(u.MessageChan)

	return nil
}

func Chat(rdb *redis.Client, channel string, content string) error {
	//fmt.Println("publish on channel", channel)
	ctx := context.Background()

	return rdb.Publish(ctx, channel, content).Err()
}

func (u *Connection) Broadcast(rdb *redis.Client, prefix string, content string) error {
	//fmt.Println("broadcast for", u.name, u.channels)
	ctx := context.Background()

	for _, channel := range u.channels {
		if strings.HasPrefix(channel, "location") {
			//fmt.Println("broadcast on channel", channel)
			if err := rdb.Publish(ctx, channel, content); err != nil {
				return err.Err()
			}
		} else {
			fmt.Println("no match", channel)
		}

	}
	return nil
}
