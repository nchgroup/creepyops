# creepyops

Creepy Ops - Red team tool to delivery payloads using different methods.

# build

```bash
go build -ldflags "-s -w -buildid=" -trimpath .
```

# help

```
 $ ./creepyops -h
Error: Unknown subcommand. Use 'deliver', 'smuggle', or 'server'.
Usage of ./creepyops:

Subcommands:
  deliver  - Deliver a file to specified paths
  smuggle  - Smuggle a file inside HTML
  server   - Start a file server

Global Options:
  -bind       Bind address.
  -cert       Path to SSL certificate file.
  -key        Path to SSL key file.
  -200        Log only 200 responses.
  -banner     Custom Server header for banner grabbing.
  -help, -h   Show help for all commands.

Subcommand 'deliver' options:
  -file string
        File to deliver. (required)
  -path string
        Comma-separated paths to deliver payload. (example: '/static/main.js,/delivery/payload') (default "/")

Subcommand 'smuggle' options:
  -file string
        File to smuggle. (required)
  -len int
        Random byte key length of XOR for the smuggled file. (default 32)
  -name string
        Name of the file for download. (example: download.exe) (required)
  -path string
        Comma-separated paths to deliver payload. (example: '/download/older.html,/download/newer.html') (default "/")

Subcommand 'server' options:
  -dir string
        Directory to serve files. (default: current directory) (default ".")

Examples:
  ./creepyops deliver -file=/home/kali/payloads/revtcp.ps1 -path=/hello.jpg
  ./creepyops -bind=0.0.0.0:80 smuggle -file=file.exe -name=totallynotavirus.exe -path=/download
  ./creepyops server -dir=/home/kali/payloads
```
