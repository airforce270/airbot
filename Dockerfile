FROM golang:1.21

WORKDIR /go/src/github.com/airforce270

# To use local copies of dependencies (for development), run the copy here.
# For example:
# COPY helix/. ./helix

WORKDIR /go/src/github.com/airforce270/airbot

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY airbot/go.mod airbot/go.sum ./
RUN go mod download && go mod verify

COPY airbot/. .
RUN go build -v -o airbot/build/airbot

ENTRYPOINT ["airbot/build/airbot"]
