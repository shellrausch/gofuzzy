# Stage: Building
FROM golang:1.10-alpine AS builder

WORKDIR /go/src/github.com/shellrausch/gofuzzy/
COPY . .

RUN CGO_ENABLED=0 go build -a -o gofuzzy

# Stage: Running
FROM alpine:latest

COPY --from=builder /go/src/github.com/shellrausch/gofuzzy/gofuzzy /usr/local/bin/

ENTRYPOINT [ "/usr/local/bin/gofuzzy" ]
CMD [ "" ]
