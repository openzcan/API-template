package utils

import (
	"math/rand"
	"net/url"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func RemoveAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	output, _, e := transform.String(t, s)
	if e != nil {
		panic(e)
	}
	return output
}

func UnescapeAccents(s string) (string, error) {
	decodedValue, err := url.QueryUnescape(s)
	if err != nil {
		return "", err
	}

	return RemoveAccents(decodedValue), nil
}

func NormalizeAddress(s string) string {
	return strings.Title(strings.ToLower(RemoveAccents(strings.Trim(s, " "))))
}

func GenerateUUID() string {
	rand, _ := uuid.NewRandom()
	//fmt.Println(rand)
	return rand.String()
}

func CheckChars(r rune) rune {
	switch {
	case r >= 'A' && r <= 'Z':
		return r
	case r >= 'a' && r <= 'z':
		return r
	case r >= '0' && r <= '9':
		return r
	case r == '/' || r == '.':
		return r
	}
	return '_'
}

func GeneratePassword(length int) string {
	//rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyz" + "0123456789" + "!@#$%^&*()")
	password := make([]rune, length)
	for i := 0; i < length; i++ {
		password[i] = chars[rand.Intn(len(chars))]
	}
	return string(password)
}
