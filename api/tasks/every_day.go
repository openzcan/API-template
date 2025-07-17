package tasks

import (
	"fmt"
	"myproject/api/database"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	GormLogger "gorm.io/gorm/logger"
)

func DoEveryDay(rdb *redis.Client, cfg database.ClusterConfig) {

	fmt.Println("running every day", time.Now())

	//clearAppEvents(rdb, connectionString)

	clearTempFiles(rdb, cfg)

}

func clearAppEvents(rdb *redis.Client, cfg database.ClusterConfig) {
	// Create a new lock client.

	locker := redislock.New(rdb)

	ctx := context.Background()

	// Try to obtain lock.
	lock, err := locker.Obtain(ctx, "clearAppEvents", 5000*time.Millisecond, nil)
	if err == redislock.ErrNotObtained {
		//fmt.Println("IgnoredOrders Could not obtain lock!")
		return
	} else if err != nil {
		return // log.Fatalln(err)
	}

	//  defer Release.
	defer lock.Release(ctx)
	//fmt.Println("IgnoredOrders I have a lock!")

	db, err := gorm.Open(postgres.Open(cfg.Primary.GetDSN()), &gorm.Config{
		Logger:                                   GormLogger.Default.LogMode(GormLogger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	if database.GetParam("LOG_TASK_SQL") == "true" {
		//fmt.Println("Logging Database queries turn off in bin/start.sh")
		db.Config.Logger.LogMode(GormLogger.Info)
	}

	sqlDB, err := db.DB()
	defer sqlDB.Close()

	db.Exec("delete from app_events where created_at < now() - interval '3 days'")

}

func clearTempFiles(rdb *redis.Client, cfg database.ClusterConfig) {
	// Create a new lock client.

	locker := redislock.New(rdb)

	ctx := context.Background()

	// Try to obtain lock.
	lock, err := locker.Obtain(ctx, "clearTempFiles", 5000*time.Millisecond, nil)
	if err == redislock.ErrNotObtained {
		//fmt.Println("clearTempFiles Could not obtain lock!")
		return
	} else if err != nil {
		return // log.Fatalln(err)
	}

	//  defer Release.
	defer lock.Release(ctx)
	fmt.Println("clearTempFiles I have a lock!")

	// delete all files in the public/temp directory older than 1 day
	// get all the files in the directory
	// for each file, check the modified time
	// if the modified time is older than 1 day, delete the file

}
