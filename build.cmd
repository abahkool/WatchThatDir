cd /D "%~dp0"
go mod tidy  
go build -ldflags "-s -w" -o .\bin\