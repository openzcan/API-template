package tasks

import (
	"fmt"
	"myproject/api/database"
	"myproject/api/models"
	"os"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	GormLogger "gorm.io/gorm/logger"
)

func ClearOldRunOnceTasks(cfg database.ClusterConfig) {

	// clear run once tasks - does not affect periodic tasks
	db, err := gorm.Open(postgres.Open(cfg.Primary.GetDSN()), &gorm.Config{
		Logger:                                   GormLogger.Default.LogMode(GormLogger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	if database.GetParam("LOG_TASK_SQL") == "true" {
		fmt.Println("Logging Database queries turn off in bin/start.sh")
		db.Config.Logger.LogMode(GormLogger.Info)
	}

	//  gorm does connection pooling so no need to close the connection - defer db.Close()
	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println("Error getting *sql.DB object:", err)
		return
	}
	defer sqlDB.Close()

	db.Exec("delete from tasks where frequency = 'once' and run_at < now() - interval '1 minute'")
}

func DoEveryMinute(rdb *redis.Client, cfg database.ClusterConfig) {

	//fmt.Println("running every minute", time.Now())

	DeleteTempFiles(rdb, cfg)

}

func DeleteTempFiles(rdb *redis.Client, cfg database.ClusterConfig) {
	// Create a new lock client.

	locker := redislock.New(rdb)

	ctx := context.Background()

	// Try to obtain lock.
	lock, err := locker.Obtain(ctx, "DeleteTempFiles", 5000*time.Millisecond, nil)
	if err == redislock.ErrNotObtained {
		//fmt.Println("DeleteTempFiles Could not obtain lock!")
		return
	} else if err != nil {
		return // log.Fatalln(err)
	}

	//  defer Release.
	defer lock.Release(ctx)
	//fmt.Println("DeleteTempFiles I have a lock!")

	db, err := gorm.Open(postgres.Open(cfg.Primary.GetDSN()), &gorm.Config{
		Logger:                                   GormLogger.Default.LogMode(GormLogger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	if database.GetParam("LOG_TASK_SQL") == "true" {
		fmt.Println("Logging Database queries turn off in bin/start.sh")
		db.Config.Logger.LogMode(GormLogger.Info)
	}

	sqlDB, err := db.DB()
	defer sqlDB.Close()

	var tasks []models.Task

	db.Find(&tasks,
		"run_at < now() and enabled != false and (name = 'Delete temporary invoice file')   ")

	for _, task := range tasks {
		// delete the file
		os.Remove(task.Data)
		// delete the task
		db.Delete(&task)
	}
}
