analysis:
	GODEBUG=allocfreetrace=1 go build -gcflags=-m escape.go

run:
	GOGC=200 GODEBUG=allocfreetrace=1 go run escape.go