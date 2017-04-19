JumpCloud README
================

## Build

go build --ldflags "-s -w" -o jumpcloud jumpcloud.go

## Run

./jumpcloud

Server will listen on port 8080. Ctrl-C or using the `kill' command will gracefully
terminate the server.
