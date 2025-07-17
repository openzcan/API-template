package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"myproject/api/database"
	"myproject/api/models"
	"myproject/api/services"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/gofiber/websocket/v2"
)

const (
	onSubscribe   = "subscribe"
	onJoin        = "join"
	onUnsubscribe = "unsubscribe"
	onLeave       = "leave"
	onMessage     = "message"
	onPing        = "ping"
)

func handler(c *websocket.Conn) {
	// c.Locals is added to the *websocket.Conn
	// log.Println(c.Locals("allowed"))  // true
	// log.Println(c.Params("username")) // 123
	// log.Println(c.Query("v"))         // 1.0
	// log.Println(c.Cookies("session")) // ""

	// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
	var (
		mt       int
		pkt      []byte
		err      error
		u        *Connection
		msg      models.WebsocketMessage
		username = c.Params("username")
	)

	cfg, err := database.LoadConfig()

	// see also https://pkg.go.dev/github.com/go-redis/redis/v7@v7.4.1#ClusterClient
	rdb, _ := database.ConnectRedis(cfg.Redis)

	u, _ = Connect(rdb, username, c)

out:
	for {
		if mt, pkt, err = c.ReadMessage(); err != nil {
			log.Println("read:", username, err)
			u.Disconnect()
			break
		}
		//log.Printf("recv: %d %s", mt, pkt)
		// unmarshal into Message struct
		if err = json.Unmarshal(pkt, &msg); err != nil {
			continue out
		}
		//log.Println("Message", username, msg.Event, msg.Channel, msg.Subject, msg.Payload)

		switch msg.Event {
		case onSubscribe, onJoin:
			fmt.Println("subscribe", msg.Channel)
			if err := u.Subscribe(rdb, msg.Channel); err != nil {
				fmt.Println("Error", err)
				u.Disconnect()
				break out
			}
			// echo subscribe command
			if err = c.WriteMessage(mt, pkt); err != nil {
				log.Println("write:", err)
			}
		case onUnsubscribe, onLeave:
			//fmt.Println("unsubscribe", msg.Channel)
			if err := u.Unsubscribe(rdb, msg.Channel); err != nil {
				fmt.Println("Error", err)
				u.Disconnect()
				break out
			}

		case onMessage:
			//fmt.Println("message to", msg.Channel)
			if err := Publish(rdb, msg.Channel, string(pkt[:])); err != nil {
				fmt.Println("Error", err)
				u.Disconnect()
				break
			}
			// if this is a merchant to a client copy to the merchant channels
			// for other connected instances of the same location
			if strings.HasPrefix(msg.Channel, "guest") {
				if err := u.Broadcast(rdb, "location", string(pkt[:])); err != nil {
					fmt.Println("Error", err)
					u.Disconnect()
					break out
				}
			} else {
				fmt.Println("not guest", msg.Channel, strings.HasPrefix("guest", msg.Channel))

			}
		case onPing:
			// the ping from the client keeps the connection alive
			// if os.Getenv("TEST_MODE") == "true" || database.GetParam("DEV_MODE") == "true" {
			// 	fmt.Println("ping from", username)
			// }
			msg.Payload = "pong"
			if err := c.WriteJSON(msg); err != nil {
				fmt.Println(err)
				u.Disconnect()
				break out
			}
		}
	}

}

func SetupWebsocketRoutes(app fiber.Router) {

	//fmt.Println("setup websockets handlers")

	app.Get("/chat/user/:username", websocket.New(handler))
	app.Get("/chat/business/:username", websocket.New(handler))
}

// routes prefixed with /api/v1/ui
func SetupRestrictedApiRoutes(app fiber.Router) {

	// publish an order update to Redis
	app.Post("/order/publish/:publisher", func(c *fiber.Ctx) error {
		if err := services.VerifyMatchingUser(c.Locals("currentUser").(models.User).ID, c); err != nil {
			return err
		}

		//publisher := c.Params("publisher")

		message := new(models.MultiChannelMessage)
		if err := c.BodyParser(message); err != nil {
			fmt.Println(err)
			c.Status(503).SendString(err.Error())
			return err
		}

		if err := PublishToChannels(message); err != nil {
			return err
		}

		return c.JSON(fiber.Map{"result": "OK"})
	})

}
