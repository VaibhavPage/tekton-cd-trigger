trigger:
	go build -v -o dist/trigger main.go

trigger-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make trigger

