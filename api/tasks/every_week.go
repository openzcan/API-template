package tasks

import (
	"fmt"
	"myproject/api/database"
	"time"

	"github.com/redis/go-redis/v9"
)

func DoEveryWeek(rdb *redis.Client, cfg database.ServiceConfig) {

	fmt.Println("running every week", time.Now())

}
