package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"myproject/api/database"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

//////// DO spaces images

func FixupPhone(phone string) string {
	// fixup the phone number
	// missing, + prefix, missing +1 prefix,  +216..
	// +12166456767

	// first remove all spaces
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")

	if len(phone) == 12 && strings.HasPrefix(phone, "+1") {
		// all good US number
		return phone
	}
	if len(phone) == 11 && strings.HasPrefix(phone, "1") {
		// missing + prefix
		return fmt.Sprintf("+%s", phone)
	}

	if len(phone) == 10 {
		phone = strings.TrimPrefix(phone, "+")
		phone = fmt.Sprintf("+1%s", phone)
	}

	return phone
}

func GeneratePin() uint {
	return uint(rand.Intn(79000) + 10000)
}

func ExtractLatLng(latlng string) (float64, float64) {
	var lat, lng float64

	if strings.HasPrefix(latlng, "[") {
		fmt.Sscanf(latlng, "[%f,%f]", &lat, &lng)
	} else {
		fmt.Sscanf(latlng, "%f,%f", &lat, &lng)
	}

	return lat, lng
}

func GeocodeAddress(address string) (string, error) {
	// use the google maps api to get the lat/lng for the address
	// https://maps.googleapis.com/maps/api/geocode/json?address=1600+Amphitheatre+Parkway,+Mountain+View,+CA&key=YOUR_API_KEY

	// uriencode the address
	address = strings.ReplaceAll(address, " ", "+")
	key := database.GetParam("GOOGLE_MAPS_API_KEY")

	if key == "" {
		return "", fmt.Errorf("no google maps api key found in database")
	}

	uri := fmt.Sprintf("https://maps.googleapis.com/maps/api/geocode/json?address=%s&key=%s", address, key)

	resp, err := http.Get(uri)
	if err != nil {
		log.Print(err)
		return "", err
	}

	defer resp.Body.Close()

	barr, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
		return "", err
	}

	body := string(barr[:])

	// fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)
	// fmt.Println("response Body:", body)

	var result map[string]interface{}
	json.Unmarshal([]byte(body), &result)

	// extract the lat/lng
	results, ok := result["results"].([]interface{})
	if !ok || len(results) == 0 {
		return "", fmt.Errorf("no results found")
	}

	geometry, ok := results[0].(map[string]interface{})["geometry"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("no geometry found")
	}

	// extract the lat/lng
	latlng, ok := geometry["location"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("no location found")
	}

	lat := latlng["lat"].(float64)
	lng := latlng["lng"].(float64)

	// fmt.Println("lat/lng", lat, lng)

	return fmt.Sprintf("%f,%f", lat, lng), nil

}

// Helper function to convert numbers to words
func NumberToWords(num uint) string {
	if num == 0 {
		return "Zero"
	}

	// Arrays to store words for numbers
	ones := []string{"", "One", "Two", "Three", "Four", "Five", "Six", "Seven", "Eight", "Nine", "Ten",
		"Eleven", "Twelve", "Thirteen", "Fourteen", "Fifteen", "Sixteen", "Seventeen", "Eighteen", "Nineteen"}
	tens := []string{"", "", "Twenty", "Thirty", "Forty", "Fifty", "Sixty", "Seventy", "Eighty", "Ninety"}

	var result string

	// Handle billions
	if num >= 1000000000 {
		result += NumberToWords(num/1000000000) + " Billion "
		num %= 1000000000
	}

	// Handle millions
	if num >= 1000000 {
		result += NumberToWords(num/1000000) + " Million "
		num %= 1000000
	}

	// Handle thousands
	if num >= 1000 {
		result += NumberToWords(num/1000) + " Thousand "
		num %= 1000
	}

	// Handle hundreds
	if num >= 100 {
		result += NumberToWords(num/100) + " Hundred "
		num %= 100
	}

	// Handle tens and ones
	if num > 0 {
		if num < 20 {
			result += ones[num]
		} else {
			result += tens[num/10]
			if num%10 > 0 {
				result += "-" + ones[num%10]
			}
		}
	}

	return strings.TrimSpace(result)
}

func StringFromMapValue(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key]; ok {
		if reflect.TypeOf(v) == reflect.TypeOf(string("")) {
			return v.(string)
		}
	}
	return ""
}

func FloatFromMapValue(m map[string]interface{}, key string) float64 {
	if m == nil {
		return 0
	}
	if v, ok := m[key]; ok {

		if reflect.TypeOf(v) == reflect.TypeOf(float64(0)) {
			return v.(float64)
		} else if reflect.TypeOf(v) == reflect.TypeOf(int(0)) {
			return float64(v.(int))
		} else if reflect.TypeOf(v) == reflect.TypeOf(string("")) {
			// parse the float from v
			if f, err := strconv.ParseFloat(v.(string), 64); err == nil {
				return f
			}
			return 0
		}

	}
	return 0
}

func UintFromMapValue(m map[string]interface{}, key string) uint {
	if m == nil {
		//fmt.Println("map value is nil for key", key)
		return 0
	}
	if v, ok := m[key]; ok {
		if reflect.TypeOf(v) == reflect.TypeOf(float64(0)) {

			//fmt.Println("map value is float for key", key)
			return uint(v.(float64))
		} else if reflect.TypeOf(v) == reflect.TypeOf(uint(0)) {

			//fmt.Println("map value is uint for key", key)
			return v.(uint)
		} else if reflect.TypeOf(v) == reflect.TypeOf(int(0)) {

			//fmt.Println("map value is int for key", key)
			return uint(v.(int))
		} else if reflect.TypeOf(v) == reflect.TypeOf(string("")) {

			//fmt.Println("map value is string for key", key)
			// parse the uint from v
			if f, err := strconv.ParseUint(v.(string), 10, 64); err == nil {
				return uint(f)
			}
			return 0
		}
	}
	return 0
}
