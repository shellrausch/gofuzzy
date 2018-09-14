package utils

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// NormalizeURL normalizes example.com:80 to http://example.com:80
func NormalizeURL(u string) (*url.URL, error) {
	if !isHTTPPrepended(u) {
		u = prependHTTP(u)
	}

	// We do it like browsers, just remove the trailing slash.
	// This will save us from a lot of problems later.
	u = strings.TrimSuffix(u, "/")

	parsedURL, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse URL/hostname: %s. %s", u, err)
	}

	return parsedURL, nil
}

// CountWordlistLines counts all lines in a file.
func CountWordlistLines(file string) uint {
	fh, _ := os.Open(file)
	defer fh.Close()

	s := bufio.NewScanner(fh)
	var lc uint
	for s.Scan() {
		lc++
	}

	return lc
}

// CountWords counts all words for a given string. A word consists just of unicode letters.
func CountWords(bytes *[]byte) int {
	numWords := 0
	isWord := false

	for _, c := range *bytes {
		r := rune(c)
		if unicode.IsLetter(r) {
			isWord = true
		} else if isWord && !unicode.IsLetter(r) {
			numWords++
			isWord = false
		}
	}

	return numWords
}

// HeaderSize calculates the whole header size.
func HeaderSize(h http.Header) int {
	l := 0
	for field, value := range h {
		l += len(field)
		for _, v := range value {
			l += len(v)
		}
	}

	return l
}

// MapToStrArray inserts map keys into an array.
func MapToStrArray(m map[string]bool) []string {
	s := []string{}
	for k := range m {
		s = append(s, k)
	}

	return s
}

// IsExtFormatValid checks if an extension has a valid format: \.[a-z0-9]+
func IsExtFormatValid(ext string) bool {
	if string(ext[0]) != "." {
		return false
	}

	for _, letter := range ext[1:] {
		if !unicode.IsLetter(rune(letter)) && !unicode.IsDigit(rune(letter)) {
			return false
		}
	}

	return true
}

// SplitHeaderFields splits header fields by a ":".
func SplitHeaderFields(h, sep string) map[string]string {
	header := make(map[string]string)

	if len(h) == 0 {
		return header
	}

	headerLine := strings.Split(h, sep)
	for _, h := range headerLine {
		sepIndex := strings.Index(h, ":")

		if sepIndex == -1 {
			log.Fatalln("Malformed header name/value. Missing separator colon ':', like name:value")
			continue
		}

		name := strings.TrimSpace(h[:sepIndex])
		value := strings.TrimSpace(h[sepIndex+1:])
		header[name] = value
	}

	return header
}

// MapSplit splits a string by separator, converts the tokens to an int and adds the converted token as a key to an map.
func MapSplit(argval, sep string) map[int]bool {
	return convertToIntMap(strings.Split(argval, sep))
}

func isHTTPPrepended(hostname string) bool {
	match, _ := regexp.MatchString("^http(s)?://", hostname)
	return match
}

func prependHTTP(hostname string) string {
	return "http://" + hostname
}

func convertToIntMap(arr []string) map[int]bool {
	m := map[int]bool{}
	for _, v := range arr {
		if i, err := strconv.Atoi(v); err == nil {
			m[i] = true
		}
	}

	return m
}
