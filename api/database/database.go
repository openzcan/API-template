package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	GormLogger "gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
)

var (
	DBConn       *gorm.DB
	AddAppEvents bool = true
	LogPath           = ""
	SystemParams      = make(map[string]string)
	SystemConfig AppConfig
)

type ServiceConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Dbname   string `yaml:"dbname"`
}

type ClusterConfig struct {
	Primary  ServiceConfig            `yaml:"primary"`
	Replicas map[string]ServiceConfig `yaml:"replicas"`
}

type AppConfig struct {
	Port     uint              `yaml:"port"`
	Redis    ClusterConfig     `yaml:"redis"`
	Database ClusterConfig     `yaml:"database"`
	Params   map[string]string `yaml:"params"`
}

func (c *ServiceConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", c.Host, c.Port, c.User, c.Dbname, c.Password)
}

func (c *ServiceConfig) GetShardDSN(id uint) string {
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s_%d password=%s sslmode=disable", c.Host, c.Port, c.User, c.Dbname, id, c.Password)
}

// common database tables

type Tag struct {
	ID  uint   `gorm:"primary_key" json:"id"`
	Tag string `json:"tag"`
}

func GetParam(key string) string {
	value, ok := SystemParams[strings.ToLower(key)]

	if !ok {
		return ""
	}
	return value
}

func InitDatabase(rdb *redis.Client, cfg ClusterConfig) *gorm.DB {

	if cfg.Primary.Host == "" {
		panic("config has no database host")
	}

	var logLevel = GormLogger.Warn

	if os.Getenv("DEV_MODE") == "true" {
		logLevel = GormLogger.Info
	} else if os.Getenv("LOG_SQL") == "true" {
		fmt.Println("Logging Database queries turn off LOG_SQL in environment")
		logLevel = GormLogger.Info
	} else if GetParam("LOG_SQL") == "true" {
		fmt.Println("Logging Database queries turn off LOG_SQL in config.params")
		logLevel = GormLogger.Info
	} else if rdb != nil {
		// query redis client for the key "LOG_SQL"
		if rdb.Get(context.Background(), "LOG_SQL").Val() == "true" {
			fmt.Println("Logging Database queries turn off LOG_SQL in redis")
			logLevel = GormLogger.Info
		}
	}

	//var err error
	DBConn, err := gorm.Open(postgres.Open(cfg.Primary.GetDSN()),
		&gorm.Config{
			Logger:                                   GormLogger.Default.LogMode(logLevel),
			PrepareStmt:                              false,
			DisableForeignKeyConstraintWhenMigrating: true,
		})

	if err != nil {
		fmt.Println(err, cfg.Primary.Host)
		panic("failed to connect database")
	}

	if len(cfg.Replicas) > 0 {
		//fmt.Println("connecting to ", len(cfg.Replicas), "replicas")

		var replicas []gorm.Dialector
		// for each replica in cfg.Replicas, add to the array
		for _, repCfg := range cfg.Replicas {
			replicas = append(replicas, postgres.Open(repCfg.GetDSN()))
		}

		DBConn.Use(dbresolver.Register(dbresolver.Config{
			// use `db2` as sources, `db3`, `db4` as replicas
			Sources:  []gorm.Dialector{postgres.Open(cfg.Primary.GetDSN())},
			Replicas: replicas,
			// sources/replicas load balancing policy
			Policy: dbresolver.RandomPolicy{},
			// print sources/replicas mode in logger
			TraceResolverMode: true,
		}))
	}

	// Add query timing metrics
	DBConn.Callback().Query().Before("gorm:query").Register("metrics:before_query", func(db *gorm.DB) {
		db.Statement.Context = context.WithValue(db.Statement.Context, "query_start_time", time.Now())
	})

	DBConn.Callback().Query().After("gorm:query").Register("metrics:after_query", func(db *gorm.DB) {
		if startTime, ok := db.Statement.Context.Value("query_start_time").(time.Time); ok {
			duration := time.Since(startTime).Milliseconds()
			if duration > 100 { // Log slow queries
				log.Printf("SLOW QUERY (%dms): %s", duration, db.Statement.SQL.String())
			}
		}
	})

	return DBConn
}

func ConnectRedis(cfg ClusterConfig) (*redis.Client, error) {

	if cfg.Primary.Host == "" {
		panic("config has no redis address")
	}

	username := "default"

	if cfg.Primary.User != "" {
		username = cfg.Primary.User
	}

	var rdb *redis.Client

	if cfg.Primary.Password != "" {
		if os.Getenv("TEST_MODE") == "true" {
			//fmt.Println("connecting to redis instance at", fmt.Sprintf("%s:%s with user %s and password", cfg.Primary.Host, cfg.Primary.Port, username), cfg.Primary.Password)
		}
		rdb = redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:%s", cfg.Primary.Host, cfg.Primary.Port), Username: username, Password: cfg.Primary.Password})
	} else {
		//fmt.Println("connecting to redis instance at", fmt.Sprintf("%s:%s", cfg.Primary.Host, cfg.Primary.Port))
		rdb = redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:%s", cfg.Primary.Host, cfg.Primary.Port)})
	}

	return rdb, nil
}

func LoadConfig() (AppConfig, error) {

	var cfg AppConfig

	if os.Getenv("USE_DOCKER") == "true" {
		yamlFile, err := os.ReadFile("/app/WOIRTUMNSDFOEWR983745/docker.yml")
		if err != nil {
			panic(err)
		}
		err = yaml.Unmarshal(yamlFile, &cfg)
		if err != nil {
			panic(err)
		}

		return cfg, nil
	}

	if os.Getenv("DEV_MODE") == "true" {
		yamlFile, err := os.ReadFile("./WOIRTUMNSDFOEWR983745/development.yml")
		if err != nil {
			panic(err)
		}
		err = yaml.Unmarshal(yamlFile, &cfg)
		if err != nil {
			panic(err)
		}

		SystemConfig = cfg

		return cfg, nil
	}

	yamlFile, err := os.ReadFile("./WOIRTUMNSDFOEWR983745/production.yml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		panic(err)
	}

	SystemConfig = cfg
	SystemParams = cfg.Params

	return cfg, nil
}
