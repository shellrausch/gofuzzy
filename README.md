# GoFuzzy

GoFuzzy is a web directory/file finder and a HTTP request fuzzer. GoFuzzy is inspired by `wfuzz` which is one of my favorite tools.
An occurence of the `FUZZ` keyword (anywhere in the request) will be replaced by an entry from the wordlist. Recommended wordlists: see [SecList](https://github.com/danielmiessler/SecLists/tree/master/Discovery/Web-Content).


## Sample run

```bash
$ gofuzzy -u localhost -w wordlists/wl.txt

   _ __________ + _________     *
_ _ __ /  ________/ ____________________ ___   *
    __/  /  / _  /  __/  / /__ /___  /  /  / -+
   +  \____/\___/__/  \___/_____/____\_   /
           *          - -+          /____/        *

---------------------------------------------------------------------------------
Chars(-hh)    Words(-hw)   Lines(-hl)   Header(-hr)  Code(-hc)    Result
---------------------------------------------------------------------------------
185           22           7            140          301          Admin
185           22           7            140          301          Login
185           22           7            140          301          login
0             0            0            198          200          passwords
185           22           7            119          301          test
```

## Build

Make sure Go is [installed](https://golang.org/doc/install) and the `$GOPATH` is set correctly.

```bash
go get github.com/shellrausch/gofuzzy
cd $GOPATH/src/github.com/shellrausch/gofuzzy
go build
```

## Install

Make sure you have followed the _Build_ step above.

```bash
cd $GOPATH/src/github.com/shellrausch/gofuzzy
go install
```

After the installation the `gofuzzy` binary will be in `$PATH`. That means you can call `gofuzzy` from everywhere.

## Usage

Find hidden files or directories

```bash
gofuzzy -u example.com -w wl.txt
```

```bash
gofuzzy -u example.com/subdir/FUZZ/config.bak -w wl.txt
```

Brute force a header field

```bash
gofuzzy -u example.com -w wl.txt -H "User-Agent: FUZZ"
```

Brute force a file extension

```bash
gofuzzy -u example.com/file.FUZZ -w ext.txt
```

Brute force a password send via a form with POST

```bash
gofuzzy -u example.com/login.php -w wl.txt -m POST \
    -d "user=admin&passwd=FUZZ&submit=s"
    -H "Content-Type: application/x-www-form-urlencoded"
```

Brute force HTTP methods

```bash
gofuzzy -u example.com -w wl.txt -m FUZZ
```

## Arguments and Help

```bash
-404
	Show 404 status code responses.
-H string
	Custom header fields, separated by comma. Example: -H 'User-Agent:Chrome,Cookie:Session=abcd'
-a string
	User-Agent.
-c string
	Cookie.
-d string
	Post data.
-f	Follow 30x redirects.
-hc string
	Hide results with specific HTTP codes, separated by comma. Example: -hc 404,500
-hh string
	Hide results with specific number of chars, separated by comma. Example: -hh 48,1024
-hl string
	Hide results with specific number of lines, separated by comma. Example: -hl 48,1024
-hr string
	Hide results with specific header length, separated by comma. Example: -hr 48,1024
-hw string
	Hide results with specific number of words, separated by comma. Example: -hw 48,1024
-m string
	HTTP method. GET, POST, <CUSTOM>, ... (default "GET")
-o string
	Output file for the results.
-of string
	Format of output file. Currently supported: csv, txt, json. Example: -of txt
-p	Progress output. (default true)
-s int
	Sleep time in milliseconds between requests per Go routine.
-t int
	Concurrency level. (default 8)
-to int
	HTTP timeout in milliseconds. (default 10000)
-u string
	URL/Hostname.
-w string
	Wordlist file.
-x string
	Appended file extension to the path, separated by comma. Example: -x .php,.html,.jpg
```

## Docker

- Build the image.

```bash
cd $GOPATH/src/github.com/shellrausch/gofuzzy
docker build -t gofuzzy .
```

- Run GoFuzzy in a container.

```bash
docker run -v $(pwd)/wordlists:/wordlists gofuzzy -u localhost -w /wordlists/wl.txt
```
