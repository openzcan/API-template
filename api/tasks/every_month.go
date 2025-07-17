package tasks

import (
	"fmt"
	"myproject/api/database"
	"time"

	"github.com/redis/go-redis/v9"
)

func DoEveryMonth(rdb *redis.Client, cfg database.ClusterConfig) {

	fmt.Println("running every month", time.Now())

}
