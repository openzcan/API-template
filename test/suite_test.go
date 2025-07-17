package test

import (
	"log"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type MyTestSuite struct {
	suite.Suite
	db  *gorm.DB
	app *fiber.App
}

var tt *testing.T

func TestMyTestSuite(t *testing.T) {
	tt = t
	suite.Run(t, new(MyTestSuite))
}

func (s *MyTestSuite) SetupSuite() {
	log.Println("SetupSuite()")
	// connect the database, save to 's.db'
	app, db, err := SetupTestApp()
	if err != nil {
		tt.Fatalf("Failed to setup test app: %v", err)
	}
	s.db = db
	s.app = app
}

func (s *MyTestSuite) TearDownSuite() {
	log.Println("TearDownSuite()")
	// delete the created database
}

func (s *MyTestSuite) SetupTest() {
	log.Println("SetupTest()")
}

func (s *MyTestSuite) TearDownTest() {
	log.Println("TearDownTest()")
}

func (s *MyTestSuite) BeforeTest(suiteName, testName string) {
	log.Println("BeforeTest()", suiteName, testName)
}

func (s *MyTestSuite) AfterTest(suiteName, testName string) {
	log.Println("AfterTest()", suiteName, testName)
}

func (s *MyTestSuite) TestExample1() {
	s.Equal(true, true)
}

func (s *MyTestSuite) TestExample2() {
	s.Equal(true, true)
}
