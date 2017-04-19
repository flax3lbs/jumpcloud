JumpCloud README
================

There are two source files, essentially very similar solutions.  jumpcloud2.go uses
a couple of global structs, whereas jumpcloud.go initializes the same structs in 
func main and passes the pointers to the handlers.  jumpcloud.go appeared to be
working, but I started getting crashes.  jumpcloud2.go may be more stable.  I 
haven't had any crashes with it so far.


## Build

go build --ldflags "-s -w" -o jumpcloud2 jumpcloud2.go

## Run

./jumpcloud2

Server will listen on port 8080. Ctrl-C or using the `kill' command will gracefully
terminate the server.
