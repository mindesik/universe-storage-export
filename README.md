Export tool for cs

## Building

`$ env GOOS="windows" go build -o ./dist/export.exe src/src.go`

## Configuration

Create file `config.json`:

`"connection": "SYSDBA:masterkey@127.0.0.1"` Firebird connection settigns

`"dbPath": "C:\\Program Files (x86)\\...` .DB database path

`"remoteURL": "https://..."` Remote url for HTTP POST requests

`"login": "..."` Login for HTTP Basic auth

`"password": "..."` Password for HTTP Basic auth

## Usage

`export.exe` - Run export via HTTP POST

`export.exe file` - Run export to file (export.json)