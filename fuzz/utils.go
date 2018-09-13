package fuzz

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	log "github.com/Sirupsen/logrus"
)

// normalizeURL normalizes example.com:80 to http://example.com:80/
func normalizeURL(fuzzURL string) (string, *url.URL, error) {
	if !isHTTPPrepended(fuzzURL) {
		fuzzURL = prependHTTP(fuzzURL)
	}

	// We do it like browsers, just remove the trailing slash.
	// This will save us from a lot of problems later.
	fuzzURL = strings.TrimSuffix(fuzzURL, "/")

	p, err := url.Parse(fuzzURL)
	if err != nil {
		return "", nil, fmt.Errorf("Unable to parse url/hostname %s. %s", fuzzURL, err)
	}

	scheme := p.Scheme + "://"

	port := ""
	if p.Port() != "" {
		port = ":" + p.Port()
	}

	query := ""
	if p.RawQuery != "" {
		query = "?" + p.RawQuery
	}

	completeURL := scheme + p.Hostname() + port + p.Path + query
	// Parse the URL again to get a clean result and to validate our built construction.
	parsedURL, _ := url.Parse(completeURL)

	return completeURL, parsedURL, nil
}

func countLines(file string) uint {
	fh, _ := os.Open(file)
	defer fh.Close()

	s := bufio.NewScanner(fh)
	var lc uint
	for s.Scan() {
		lc++
	}

	return lc
}

func countWords(bytes *[]byte) int {
	isWord := false
	numWords := 0
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

func headerSize(h http.Header) int {
	l := 0
	for name, values := range h {
		l += len(name)
		for _, value := range values {
			l += len(value)
		}
	}

	return l
}

func strArrayToMapStrBool(arr []string) map[int]bool {
	m := map[int]bool{}
	for _, v := range arr {
		if i, err := strconv.Atoi(v); err == nil {
			m[i] = true
		}
	}

	return m
}

func mapToStrArray(m map[string]bool) []string {
	s := []string{}
	for k := range m {
		s = append(s, k)
	}

	return s
}

func isExtFormatValid(ext string) bool {
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

func isHTTPPrepended(hostname string) bool {
	match, _ := regexp.MatchString("^http(s)?://", hostname)
	return match
}

func prependHTTP(hostname string) string {
	return fmt.Sprintf("http://%s", hostname)
}

func splitHeaderFields(h, sep string) map[string]string {
	header := make(map[string]string)

	if len(h) == 0 {
		return header
	}

	headerFields := strings.Split(h, sep)

	for _, h := range headerFields {
		split := strings.Split(h, ":")
		if len(split) != 2 {
			log.Errorf("Header not complete or the header value contains a separator char '%s' (without quotes). Broken header is: %v", sep, split)
			continue
		}

		name := strings.Trim(split[0], " ")
		value := strings.Trim(split[1], " ")
		header[name] = value
	}

	return header
}

func convertSeparatedCmdArg(argval, sep string) map[int]bool {
	return strArrayToMapStrBool(strings.Split(argval, sep))
}

func dumpRequest(req *http.Request) string {
	b, _ := httputil.DumpRequest(req, true)
	return string(b)
}

func dumpResponse(resp *http.Response) string {
	b, _ := httputil.DumpResponse(resp, true)
	return string(b)
}
