# Stage: Building
FROM golang:alpine AS builder

WORKDIR /go/src/github.com/shellrausch/gofuzzy/

COPY . .

RUN apk update && apk add git
RUN go get -d
RUN CGO_ENABLED=0 GOOS=linux go build -a -o gofuzzy

# Stage: Running
FROM alpine:latest

COPY --from=builder /go/src/github.com/shellrausch/gofuzzy/gofuzzy /usr/local/bin/
WORKDIR /usr/local/bin

ENTRYPOINT [ "gofuzzy" ]
CMD [ "" ]
