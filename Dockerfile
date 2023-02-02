FROM golang:1.20

WORKDIR /go/src/github.com/airforce270/airbot

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o build/airbot

ENTRYPOINT ["build/airbot"]
