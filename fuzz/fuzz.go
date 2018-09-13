package fuzz

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

// ResultChannels is just a wrapper of all public result channels.
type ResultChannels struct {
	Result   chan *Result
	Progress chan *Progress
	Finished chan bool
}

// Result contains response results and results which are calculated
// for a response at runtime.
type Result struct {
	ContentLength int
	NumWords      int
	StatusCode    int
	NumLines      int
	HeaderSize    int
	Result        string
}

// Progress contains the actual progress information.
type Progress struct {
	NumDoneRequests   uint
	NumApproxRequests uint
}

// request contains all information needed to make a plain HTTP request.
// This struct is just a stub.
type request struct {
	url     string
	data    string
	method  string
	entry   string
	ext     string
	retries uint8
	header  map[string]string
}

var httpClient http.Client
var resultChans ResultChannels

// NewFuzz initializes all public channels, so that the caller
// can receive results on them.
func NewFuzz(o *Opts) ResultChannels {
	resultChans = ResultChannels{
		Result:   make(chan *Result, o.Concurrency),
		Progress: make(chan *Progress, o.Concurrency), // Just a buffer which is large enough
		Finished: make(chan bool),
	}

	return resultChans
}

// Start starts the main fuzzing process for a given option set.
func Start(o *Opts) {
	httpClient = initHTTPClient(o)

	// We minimize the chance to be 'blocked' by the filesystem as we will
	// fetch more data at once (buffered channel), so the channel remains constantly filled.
	queuedReqsCh := make(chan *request, o.Concurrency*o.Concurrency)

	// When a producer is done (reading the wordlist and creating request stubs)
	// the channel sends a value. We use it as a barrier.
	producerDoneCh := make(chan bool)

	// Synchronizes the number of Go routines which are provided with -t arg.
	concurrencyWg := new(sync.WaitGroup)

	go produceRequests(o, queuedReqsCh, producerDoneCh)

	for i := 0; i < o.Concurrency; i++ {
		concurrencyWg.Add(1)

		go func() {
			for {
				fuzzReq, open := <-queuedReqsCh
				if !open {
					concurrencyWg.Done()
					return
				}
				consumeRequest(o, fuzzReq)

				time.Sleep(o.Sleep)
			}
		}()
	}

	go produceProgress(o)

	// Order matters for a proper termination of all Go routines.
	<-producerDoneCh
	close(queuedReqsCh)
	concurrencyWg.Wait()

	resultChans.Finished <- true
	close(resultChans.Result)
	close(resultChans.Finished)
}

// produceRequests reads an entry from the wordlist and produces a request-stub
// with all relevant information to invoke a request.
func produceRequests(o *Opts, queuedReqsCh chan *request, producerDoneCh chan bool) {
	fh, _ := os.Open(o.Wordlist)

	url := strings.TrimSuffix(o.URLRaw, "/")
	header := splitHeaderFields(o.CustomHeader, o.headerFieldSep)

	s := bufio.NewScanner(fh)
	for s.Scan() {
		for _, ext := range o.FileExtensions {
			queuedReqsCh <- &request{
				method: o.HTTPMethod,
				url:    url,
				header: header,
				data:   o.BodyData,
				ext:    ext,
				entry:  s.Text(),
			}
		}
	}

	fh.Close()
	producerDoneCh <- true
}

// produceProgress produces progress information in a defined interval and
// sends them via a channel.
func produceProgress(o *Opts) {
	if o.ProgressOutput {
		// Especially on large wordlists this barrier is important.
		// Because reading the whole wordlist can take some time.
		// Otherwise we could get a division by zero in further progress calculation.
		<-o.wordlistReadComplete

		tick := time.Tick(time.Millisecond * time.Duration(o.progressSendInterval))
		progress := &Progress{}
		for {
			select {
			case <-tick:
				progress.NumDoneRequests = o.numDoneRequests
				progress.NumApproxRequests = o.numApproxRequests
				resultChans.Progress <- progress
			}
		}
	}
}

// consumeRequest takes a given request stub and invokes the HTTP request.
// If an error occurs the request is repeated a number of times
// before the request is getting canceled.
func consumeRequest(o *Opts, r *request) {
	res, err := invokeRequest(o, r)

	o.numDoneRequests++ // We don't care here for race conditions. It's just a nice to have progress value.

	if err == nil && isInHideFilter(o, res) {
		resultChans.Result <- res
	}

	if err != nil {
		if r.retries < o.maxRequestRetries {
			r.retries++

			o.numApproxRequests++ // We don't care for race conditions here. It's just a nice to have progress value.

			consumeRequest(o, r)
		} else {
			log.Errorf("Giving up a request, too many errors: %s", err)
		}
	}
}

// invokeRequest does the raw HTTP request. Before a HTTP request is finally done
// the first occurence of the FUZZ keyword will be replaced by a wordlist entry.
func invokeRequest(o *Opts, r *request) (*Result, error) {
	var req *http.Request
	var err error

	url := r.url
	if !o.fuzzKeywordPresent {
		r.entry = strings.TrimPrefix(r.entry, "/")
		url = r.url + "/" + r.entry + r.ext
	}

	req, err = http.NewRequest(r.method, url, strings.NewReader(r.data))

	if err != nil {
		return nil, err
	}

	if o.UserAgent != "" {
		req.Header.Set("User-Agent", o.UserAgent)
	}

	if o.Cookie != "" {
		req.Header.Set("Cookie", o.Cookie)
	}

	for h, v := range r.header {
		req.Header.Set(h, v)
	}

	if o.fuzzKeywordPresent {
		req, err = replaceFuzzKeywordOccurence(o, req, r)

		if err != nil {
			return nil, err
		}
	}

	// if o.Debug {
	// 	log.Debug(dumpRequest(req))
	// }

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// if o.Debug {
	// 	log.Debug(dumpResponse(resp))
	// }

	result := populateResult(resp, r.entry)
	defer resp.Body.Close()

	_, err = io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		return nil, err
	}

	return result, err
}

// The FUZZ keyword can be everywhere in the HTTP request.
// We replace the first occurency of the keyword FUZZ with an entry from the wordlist.
func replaceFuzzKeywordOccurence(o *Opts, req *http.Request, r *request) (*http.Request, error) {
	reqBytes, _ := httputil.DumpRequest(req, true)

	fuzzKeywordBytes := []byte(o.fuzzKeyword)
	entryBytes := []byte(r.entry)

	// Replaces most of the FUZZ places in the request.
	replaced := bytes.Replace(reqBytes, fuzzKeywordBytes, entryBytes, -1)
	// Go renames header fields automatically to the following format:
	// "FUZZ: text/html" to "Fuzz: text/html".
	// Therefore we make also an additional 'Fuzz' replacement.
	replaced = bytes.Replace(replaced, bytes.Title(bytes.ToLower(fuzzKeywordBytes)), entryBytes, -1)

	// Creates and validates a request from a textual (raw) request.
	reqCopy, err := http.ReadRequest(bufio.NewReader(bytes.NewBuffer(replaced)))

	// Replace extension.
	ext := strings.Replace(r.ext, o.fuzzKeyword, r.entry, -1)
	// Replace URL.
	url := strings.Replace(req.URL.String()+ext, o.fuzzKeyword, r.entry, -1)

	if err != nil {
		return nil, err
	}

	// Replace request body
	body := strings.Replace(r.data, o.fuzzKeyword, r.entry, -1)

	req, err = http.NewRequest(reqCopy.Method, url, strings.NewReader(body))
	req.Header = reqCopy.Header

	if err != nil {
		return nil, err
	}

	return req, nil
}

// populateResult creates the Result.
// In the next step the Result is enriched with additional results which are
// calculated at runtime, e.g. number of words.
func populateResult(resp *http.Response, entry string) *Result {
	b, _ := ioutil.ReadAll(resp.Body)

	// -1 indicates the length is unknown. Hence we count the body size manually.
	// This condition occures often with HTTP status codes 30x and 40x.
	if resp.ContentLength == -1 {
		resp.ContentLength = int64(len(b))
	}

	return &Result{
		ContentLength: int(resp.ContentLength),
		NumLines:      bytes.Count(b, []byte{'\n'}),
		NumWords:      countWords(&b),
		HeaderSize:    headerSize(resp.Header),
		StatusCode:    resp.StatusCode,
		Result:        entry,
	}
}

// isInHideFilter determines if some values, sizes, lengths, ...
// of the result should be filtered.
func isInHideFilter(o *Opts, res *Result) bool {
	return !o.HTTPHideCodes[res.StatusCode] &&
		!o.HTTPHideBodyLines[res.NumLines] &&
		!o.HTTPHideBodyLength[res.ContentLength] &&
		!o.HTTPHideNumWords[res.NumWords] &&
		!o.HTTPHideHeaderLength[res.HeaderSize]
}

// initHTTPClient initialises the default HTTP client with fundamental
// connection options for every request.
func initHTTPClient(o *Opts) http.Client {
	return http.Client{
		Timeout: time.Duration(o.Timeout) * time.Millisecond,
		// Do not follow redirects (HTTP status codes 30x).
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !o.FollowRedirects {
				return http.ErrUseLastResponse
			}
			return nil
		},
		// Ignore invalid certs by default, since we are interested in the content.
		// And -hopefully- we know what we are doing :}
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}
