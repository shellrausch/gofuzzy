package fuzz

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Opts contains all passed cmdline args as well as the parsed ones.
type Opts struct {
	URLRaw                  string
	HTTPHideBodyLinesRaw    string
	HTTPHideBodyLengthRaw   string
	HTTPHideNumWordsRaw     string
	HTTPHideHeaderLengthRaw string
	HTTPHideCodesRaw        string
	FileExtensionsRaw       string
	CustomHeader            string
	UserAgent               string
	Cookie                  string
	HTTPMethod              string
	Wordlist                string
	BodyData                string
	OutputFile              string
	OutputFormat            string
	SleepRaw                int
	Timeout                 int
	Concurrency             int
	Debug                   bool
	FollowRedirects         bool
	ProgressOutput          bool
	Show404                 bool
	FileExtensions          []string
	HTTPHideBodyLines       map[int]bool
	HTTPHideBodyLength      map[int]bool
	HTTPHideNumWords        map[int]bool
	HTTPHideHeaderLength    map[int]bool
	HTTPHideCodes           map[int]bool
	URL                     *url.URL
	Sleep                   time.Duration

	// Meta options that are set during the runtime.
	fuzzKeyword            string
	headerFieldSep         string
	cmdLineValueSep        string
	maxRequestRetries      uint8
	numApproxRequests      uint
	numDoneRequests        uint
	wordlistLineCount      uint
	progressSendInterval   int
	fuzzKeywordPresent     bool
	wordlistReadComplete   chan bool
	supportedOutputFormats map[string]bool
}

// Parse parses and validates the cmdline args.
func Parse(outputFormats map[string]bool) (*Opts, error) {
	o := &Opts{supportedOutputFormats: outputFormats}

	fs := flag.NewFlagSet("gofuzzy", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Print("USAGE: gofuzzy -u example.com -w wl.txt [options]")
		fmt.Println("\n   If the keyword '\x1b[31mFUZZ\x1b[0m' is provided somewhere, gofuzzy will replace it with an entry from the wordlist.")
		fmt.Println("\nEXAMPLES:")
		fmt.Println("   Find hidden files or directories:")
		fmt.Println("   # gofuzzy -u example.com -w wl.txt")
		fmt.Println("\n   Brute force a header field:")
		fmt.Println("   # gofuzzy -u example.com -w wl.txt -H 'User-Agent: \x1b[31mFUZZ\x1b[0m'")
		fmt.Println("\n   Brute force a file extension:")
		fmt.Println("   # gofuzzy -u example.com/file.\x1b[31mFUZZ\x1b[0m -w ext.txt")
		fmt.Println("\n   Brute force a password send via a form:")
		fmt.Println("   # gofuzzy -u example.com/login.php -w wl.txt -m POST -d 'user=admin&passwd=\x1b[31mFUZZ\x1b[0m&submit=s' -H 'Content-Type: application/x-www-form-urlencoded'")
		fmt.Println("\nOPTIONS:")
		fs.PrintDefaults()
	}

	fs.StringVar(&o.URLRaw, "u", "", "URL/Hostname.")
	fs.StringVar(&o.Wordlist, "w", "", "Wordlist file.")
	fs.StringVar(&o.HTTPMethod, "m", http.MethodGet, "HTTP method. GET, POST, <CUSTOM>, ...")
	fs.StringVar(&o.HTTPHideCodesRaw, "hc", "", "Hide results with specific HTTP codes, separated by comma. Example: -hc 404,500")
	fs.StringVar(&o.HTTPHideBodyLinesRaw, "hl", "", "Hide results with specific number of lines, separated by comma. Example: -hl 48,1024")
	fs.StringVar(&o.HTTPHideBodyLengthRaw, "hh", "", "Hide results with specific number of chars, separated by comma. Example: -hh 48,1024")
	fs.StringVar(&o.HTTPHideNumWordsRaw, "hw", "", "Hide results with specific number of words, separated by comma. Example: -hw 48,1024")
	fs.StringVar(&o.HTTPHideHeaderLengthRaw, "hr", "", "Hide results with specific header length, separated by comma. Example: -hr 48,1024")
	fs.StringVar(&o.FileExtensionsRaw, "x", "", "Appended file extension to the path, separated by comma. Example: -x .php,.html,.jpg")
	fs.StringVar(&o.CustomHeader, "H", "", "Custom header fields, separated by comma. Example: -H 'User-Agent:Chrome,Cookie:Session=abcd'")
	fs.StringVar(&o.BodyData, "d", "", "Post data.")
	fs.StringVar(&o.UserAgent, "a", "", "User-Agent.")
	fs.StringVar(&o.Cookie, "c", "", "Cookie.")
	fs.StringVar(&o.OutputFile, "o", "", "Output file for the results.")
	fs.StringVar(&o.OutputFormat, "of", "", "Format of output file. Currently supported: "+strings.Join(mapToStrArray(o.supportedOutputFormats), ", ")+". Example: -of txt")
	fs.IntVar(&o.Concurrency, "t", 8, "Concurrency level.")
	fs.IntVar(&o.Timeout, "to", 10000, "HTTP timeout in milliseconds.")
	fs.IntVar(&o.SleepRaw, "s", 0, "Sleep time in milliseconds between requests per Go routine.")
	fs.BoolVar(&o.FollowRedirects, "f", false, "Follow 30x redirects.")
	fs.BoolVar(&o.ProgressOutput, "p", true, "Progress output.")
	fs.BoolVar(&o.Show404, "404", false, "Show 404 status code responses.")
	// flag.BoolVar(&o.Debug, "debug", false, "Debug mode.")

	fs.Parse(os.Args[1:])

	if err := validate(o); err != nil {
		return nil, err
	}

	_init(o)

	return o, nil
}

func validate(o *Opts) error {
	log.SetLevel(log.ErrorLevel)

	if o.Debug {
		log.SetLevel(log.DebugLevel)
	}

	if o.URLRaw == "" {
		return fmt.Errorf("No URL/hostname provided. Use flag: -u example.com")
	}

	if _, _, err := normalizeURL(o.URLRaw); err != nil {
		return err
	}

	if o.Wordlist == "" {
		return fmt.Errorf("No wordlist provided. Use flag: -w wl.txt")
	}

	if _, err := os.Stat(o.Wordlist); os.IsNotExist(err) {
		return fmt.Errorf("Wordlist not found at '%s'", o.Wordlist)
	}

	if o.FileExtensionsRaw != "" {
		for _, ext := range strings.Split(o.FileExtensionsRaw, ",") {
			if !isExtFormatValid(ext) {
				return fmt.Errorf("Invalid extension %s. Extensions must contain a period followed by alphanummeric letters. Example: .php,.html", ext)
			}
		}
	}

	if o.Concurrency < 1 || o.Concurrency > 100 {
		return fmt.Errorf("The concurrency level is invalid. Must be >=1 and <=100")
	}

	if o.OutputFile != "" {
		if o.OutputFormat == "" {
			return fmt.Errorf("Provide an output format with -of")
		}

		if _, err := os.Create(o.OutputFile); err != nil {
			return err
		}
	}

	if o.OutputFormat != "" {
		if o.OutputFile == "" {
			return fmt.Errorf("Provide an output filename with -o")
		}

		o.OutputFormat = strings.ToLower(o.OutputFormat)

		if !o.supportedOutputFormats[o.OutputFormat] {
			return fmt.Errorf("Only the following output formats are supported: %s", strings.Join(mapToStrArray(o.supportedOutputFormats), ", "))
		}
	}

	return nil
}

func _init(o *Opts) {
	o.wordlistReadComplete = make(chan bool)
	go func() {
		o.wordlistLineCount = countLines(o.Wordlist)
		o.numApproxRequests = o.wordlistLineCount * uint(len(o.FileExtensions))
		o.wordlistReadComplete <- true
	}()

	o.fuzzKeyword = "FUZZ"
	o.cmdLineValueSep, o.headerFieldSep = ",", ","
	o.maxRequestRetries = 3
	o.progressSendInterval = 75
	o.URLRaw, o.URL, _ = normalizeURL(o.URLRaw)
	o.Sleep = time.Duration(o.SleepRaw) * time.Millisecond
	o.HTTPMethod = strings.ToUpper(o.HTTPMethod)
	o.HTTPHideCodes = convertSeparatedCmdArg(o.HTTPHideCodesRaw, o.cmdLineValueSep)
	o.HTTPHideBodyLength = convertSeparatedCmdArg(o.HTTPHideBodyLengthRaw, o.cmdLineValueSep)
	o.HTTPHideNumWords = convertSeparatedCmdArg(o.HTTPHideNumWordsRaw, o.cmdLineValueSep)
	o.HTTPHideBodyLines = convertSeparatedCmdArg(o.HTTPHideBodyLinesRaw, o.cmdLineValueSep)
	o.HTTPHideHeaderLength = convertSeparatedCmdArg(o.HTTPHideHeaderLengthRaw, o.cmdLineValueSep)

	if !o.Show404 {
		o.HTTPHideCodes[http.StatusNotFound] = true
	}

	if o.FileExtensionsRaw != "" {
		for _, ext := range strings.Split(o.FileExtensionsRaw, o.cmdLineValueSep) {
			o.FileExtensions = append(o.FileExtensions, ext)
		}
	} else {
		// With an empty extension we can enter the main loop once.
		o.FileExtensions = append(o.FileExtensions, "")
	}

	o.fuzzKeywordPresent = func(o *Opts) bool {
		return strings.Contains(o.URL.Path, o.fuzzKeyword) ||
			strings.Contains(o.URL.RawQuery, o.fuzzKeyword) ||
			strings.Contains(o.CustomHeader, o.fuzzKeyword) ||
			strings.Contains(o.BodyData, o.fuzzKeyword) ||
			strings.Contains(o.HTTPMethod, o.fuzzKeyword) ||
			strings.Contains(o.FileExtensionsRaw, o.fuzzKeyword) ||
			strings.Contains(o.UserAgent, o.fuzzKeyword) ||
			strings.Contains(o.Cookie, o.fuzzKeyword)
	}(o)
}
