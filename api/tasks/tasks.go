package tasks

import (
	"fmt"
	"log"
	"myproject/api/database"
	"myproject/api/models"
	"os"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	GormLogger "gorm.io/gorm/logger"
)

func SetupTasks(rdb *redis.Client, cfg database.ClusterConfig) {

	if os.Getenv("RUN_TASK_QUEUE") == "true" {

		// remove any tasks that have a run_at older than 1 minute
		ClearOldRunOnceTasks(cfg)

		s := gocron.NewScheduler(time.UTC)

		s.Every(60).Seconds().Do(func() { DoEveryMinute(rdb, cfg) })

		s.Every(1).Hour().Do(func() {
			DoEveryHour(rdb, cfg)
		})

		s.Every(1).Day().Do(func() {
			DoEveryDay(rdb, cfg)
		})

		// Schedule each first day of the month
		s.Every(1).Month(1).Do(func() { DoEveryMonth(rdb, cfg) })

		// Add task metrics collection
		s.Every(1).Minute().Do(func() {
			conn := postgres.Open(cfg.Primary.GetDSN())
			db, err := gorm.Open(conn, &gorm.Config{
				Logger:                                   GormLogger.Default.LogMode(GormLogger.Silent),
				DisableForeignKeyConstraintWhenMigrating: true,
			})

			if err != nil {
				fmt.Println(err)
				return
			}
			var taskCounts struct {
				Total    int64
				Pending  int64
				Running  int64
				Complete int64
				Failed   int64
			}

			sqlDB, err := db.DB()
			defer sqlDB.Close()

			db.Model(&models.Task{}).Count(&taskCounts.Total)
			db.Model(&models.Task{}).Where("run_at > ?", time.Now()).Count(&taskCounts.Pending)

			// Log metrics or expose via Prometheus
			log.Printf("Task metrics - Total: %d, Pending: %d",
				taskCounts.Total, taskCounts.Pending)
		})

		s.StartAsync()
	}
}
