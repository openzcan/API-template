package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"myproject/api/models"
	"myproject/test"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func TestAInjectUsers(t *testing.T) {
	_, db, err := test.SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test app: %v", err)
	}

	type Locations struct {
		Province string
		Cities   []string
	}

	var locations []Locations = []Locations{
		{
			Province: "Ohio",
			Cities:   []string{"Cleveland", "Lakewood", "Brooklyn", "Westlake", "Parma", "Berea"},
		},
		{
			Province: "Michigan",
			Cities: []string{
				"Detroit", "Ann Arbor", "Lansing", "Saginaw", "Toledo", "Holland",
			},
		},
		{
			Province: "Indiana",
			Cities: []string{
				"Indiana", "Muncie", "Lafayette", "Bloomington", "Greenwood", "Terre Haute",
			},
		},
		{
			Province: "Wisconsin",
			Cities: []string{
				"Chicago", "Milwaukie", "Rockford", "Madison", "Janesville", "Appleton",
			},
		},
		{
			Province: "Iowa",
			Cities: []string{
				"Des Moines", "Cedar Rapids", "Waterloo", "Fort Dodge", "Dubuque", "Ottumwa",
			},
		},
		{
			Province: "Kentucky",
			Cities: []string{
				"Louisville", "Lexington", "Richmond", "Evansville", "Frankfort", "Elizabethtown",
			},
		},
	}

	rand.Seed(time.Now().Unix())

	nums := makeRange(1, 999)

	// create 1000 users with a business and multiple equipment with QR codes
	for id := range nums {

		pn := rand.Intn(5)
		cn := rand.Intn(5)

		var loc = locations[pn]

		// create the api key
		apikey, _ := uuid.NewRandom()

		data := models.User{

			ApiKey:      apikey,
			UnlockToken: "",
			Name:        apikey.String()[0:20],
			Email:       fmt.Sprintf("user_%d@gmail.com", id),
			Phone:       fmt.Sprintf("+1216%d123456", id),
			City:        loc.Cities[cn],
			Province:    loc.Province,
			Country:     "USA",
		}

		db.Create(&data)
	}

}

func GetJsonTestRequestResponse(app *fiber.App, method string, url string, reqBody any) (code int, respBody map[string]any, err error) {
	bodyJson := []byte("")
	if reqBody != nil {
		bodyJson, _ = json.Marshal(reqBody)
	}
	req := httptest.NewRequest(method, url, bytes.NewReader(bodyJson))
	resp, err := app.Test(req, 10)
	code = resp.StatusCode
	// If error we're done
	if err != nil {
		return
	}
	// If no body content, we're done
	if resp.ContentLength == 0 {
		return
	}
	bodyData := make([]byte, resp.ContentLength)
	_, _ = resp.Body.Read(bodyData)
	err = json.Unmarshal(bodyData, &respBody)
	return
}

func TestBInjectBusinesses(t *testing.T) {
	_, db, err := test.SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test app: %v", err)
	}

	var users []models.User

	db.Find(&users, "email like 'user_%'")

	for _, u := range users {
		fmt.Println("user", u.ID, u.Email, u.Country)

		uuid, _ := uuid.NewRandom()

		var business = models.Business{
			Name:     uuid.String()[0:15],
			UserId:   u.ID,
			City:     u.City,
			Province: u.Province,
			Country:  u.Country,
			Email:    u.Email,
			Phone:    u.Phone,
			Enabled:  true,
			Photo:    "/images/food_display.png",
			Uuid:     uuid.String(),
		}

		db.Create(&business)

		var loc = models.Location{
			Name:       uuid.String()[0:15],
			Address:    uuid.String()[10:25],
			UserId:     u.ID,
			BusinessId: business.ID,
			Phone:      u.Phone,
			City:       u.City,
			Province:   u.Province,
			Country:    u.Country,
			Photo:      "/images/food_display.png",
			Uuid:       uuid.String()[0:8],
		}

		db.Create(&loc)

	}

}

func TestDInjectTeamMembers(t *testing.T) {
	_, db, err := test.SetupTestApp()
	if err != nil {
		t.Fatalf("Failed to setup test app: %v", err)
	}

	var role = models.BusinessRole{

		RoleId:      3,
		BusinessId:  8,
		LocationId:  8,
		Type:        "partner",
		Permissions: "read",
		Services:    []string{},
	}

	db.Create(&role)

	var users []models.User
	db.Limit(10).Find(&users, "email like 'user_%'")

	for _, u := range users {
		fmt.Println("user", u.ID, u.Email, u.Country)

		var member = models.BusinessRole{

			RoleId:      u.ID,
			BusinessId:  8,
			LocationId:  8,
			Type:        "partner",
			Permissions: "read",
			Services:    []string{},
		}

		db.Create(&member)

	}

}
