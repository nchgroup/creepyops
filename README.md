# creepyops

Creepy Ops - Red team tool to delivery payloads using different methods.

# build

```bash
go build -ldflags "-s -w -buildid=" -trimpath .
```

# help

```
Usage of ./creepyops:

Subcommands:
  deliver   - Deliver a file to specified paths
  smuggle  - Smuggle a file inside HTML
  server   - Start a file server

Global Options:
  -200
        Log only 200 responses.
  -banner string
        Custom Server header to include in responses for banner grabbing. (default "Microsoft-IIS/10.0")
  -bind string
        Bind address. (default "0.0.0.0:8000")
  -cert string
        Path to SSL certificate file. If provided, enables HTTPS.
  -h    Show help for all commands.
  -help
        Show help for all commands.
  -key string
        Path to SSL key file. If provided, enables HTTPS.

Subcommand 'deliver' options:
  -file string
        File to deliver. (required)
  -path string
        Comma-separated paths to deliver payload. (example: '/static/main.js,/delivery/payload') (default "/")

Subcommand 'smuggle' options:
  -file string
        File to smuggle. (required)
  -key int
        Random byte key length of XOR for the smuggled file. (required) (default 32)
  -name string
        Name of the file for download. (example: download.exe) (required)
  -path string
        Comma-separated paths to deliver payload. (example: '/download/older.html,/download/newer.html') (default "/")

Subcommand 'server' options:
  -dir string
        Directory to serve files. (default: current directory) (default ".")

Examples:
  ./creepyops deliver -file=/home/kali/payloads/revtcp.ps1 -path=/hello.jpg
  ./creepyops smuggle -file=file.exe -name=totallynotavirus.exe -path=/download
  ./creepyops server -dir=/home/kali/payloads -bind=0.0.0.0:8080
```