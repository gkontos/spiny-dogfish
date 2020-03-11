env GOOS=windows GOARCH=amd64 go build -o spiny-dogfish-windows-amd64.exe .
env GOOS=darwin GOARCH=amd64 go build -o spiny-dogfish-darwin-amd64 .
env GOOS=linux GOARCH=amd64 go build -o spiny-dogfish-linux-amd64 .