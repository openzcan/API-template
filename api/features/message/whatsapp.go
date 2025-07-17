package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"myproject/api/database"
	"net/http"
	"os"
	"strings"
)

func SendToWhatsapp(phone string, payload io.Reader) (string, error) {

	token := database.GetParam("WHATSAPP_ACCESS_TOKEN")

	if token == "" {
		//panic("environment has no whatsapp access token")
		return "", nil
	}

	phoneId := database.GetParam("WHATSAPP_PHONE_ID")
	if phoneId == "" {
		panic("environment has no whatsapp phone number")
	}

	endpoint := fmt.Sprintf("https://graph.facebook.com/v15.0/%s/messages", phoneId)

	if os.Getenv("TEST_MODE") == "true" || database.GetParam("DEV_MODE") == "true" {
		//fmt.Println("Test mode - NOT sending whatsapp message", params, "to", phone)
		if phone != "573125333768" {
			fmt.Println("TEST_MODE/DEV_MODE - message.go:SendToWhatsapp() - NOT Sending message to whatsapp", phone, "using token", token)
			return fmt.Sprintf("[OK:%s]", phone), nil
		} else {
			fmt.Println("TEST_MODE/DEV_MODE (to 573125333768) -  message.go:SendToWhatsapp() - Sending message to whatsapp", phone, "using token", token)
		}
	}

	//create new POST request to the url and encoded form Data
	r, err := http.NewRequest("POST", endpoint, payload)
	if err != nil {
		log.Fatal(err)
	}

	//set headers to the request
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	//send request and get the response
	client := &http.Client{}
	resp, err := client.Do(r)

	if err != nil {
		//handle error
		log.Fatal(err)
		return "", err
	}

	defer resp.Body.Close()

	barr, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	body := string(barr[:])

	// fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)
	// fmt.Println("response Body:", body)

	return body, nil
}

func SendWhatsappTemplate(recipient string, template string, params []map[string]string) {

	// send a message via whatsapp to the phone
	data := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                recipient,
		"type":              "template",
		"template": map[string]interface{}{
			"name": template,
			"language": map[string]string{
				"code": "en",
			},
			"components": []map[string]interface{}{
				{
					"type":       "body",
					"parameters": params,
				},
			},
		},
	}

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(data)

	msg := payloadBuf.String()

	fmt.Println("whatsapp.go:sendWhatsappMessage() send message", msg)

	_, err := SendToWhatsapp(recipient, strings.NewReader(msg))

	if err != nil {
		fmt.Println(err)
	}
}

func WhatsappTextMessage(phone string, content string) {
	// send a message via whatsapp to the phone
	data := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                phone,
		"type":              "text",
		"text": map[string]interface{}{ // the text object
			"preview_url": false,
			"body":        content,
		},
	}

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(data)

	msg := payloadBuf.String()

	fmt.Println("send text message", msg)

	res, err := SendToWhatsapp(phone, strings.NewReader(msg))
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Success sending to whatsapp", phone, res)
	}
}

func WhatsappOTPMessage(phone string, code string) {
	data := map[string]interface{}{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                phone,
		"type":              "template",
		"template": map[string]interface{}{
			"name": "otp_message",
			"language": map[string]string{
				"code": "es",
			},
			"components": []map[string]interface{}{
				{
					"type": "body",
					"parameters": []map[string]string{
						{
							"type": "text",
							"text": code,
						},
					},
				},
				{
					"type":     "button",
					"sub_type": "url",
					"index":    "0",
					"parameters": []map[string]string{
						{
							"type": "text",
							"text": code,
						},
					},
				},
			},
		},
	}

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(data)

	if os.Getenv("TEST_MODE") == "true" || database.GetParam("DEV_MODE") == "true" {
		//fmt.Println("Test mode - NOT sending whatsapp message", params, "to", phone)
		fmt.Println("TEST_MODE - not sending OTP message to", phone, payloadBuf.String())
		return
	}

	msg := payloadBuf.String()

	fmt.Println("send OTP message", msg)

	res, err := SendToWhatsapp(phone, strings.NewReader(msg))
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Success sending OTP message", phone, res)
	}
}
