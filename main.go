package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

type Handler func(w http.ResponseWriter, r *http.Request)

type OhttpsRequest struct {
	Timestamp int `json:"timestamp"`
	Payload   struct {
		CertificateName           string   `json:"certificateName"`
		CertificateDomains        []string `json:"certificateDomains"`
		CertificateCertKey        string   `json:"certificateCertKey"`
		CertificateFullchainCerts string   `json:"certificateFullchainCerts"`
		CertificateExpireAt       int      `json:"certificateExpireAt"`
	} `json:"payload"`
	Sign string `json:"sign"`
}

func Auth(h Handler, username string, password string) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || !strings.EqualFold(username, user) || !strings.EqualFold(password, pass) {
			w.Header().Set("WWW-Authenticate", `Basic realm="please input real username and passowrd"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}
		h(w, r)
	}
}
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	tpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		w.Write([]byte("templates/index.html 不存在"))
	}
	certs := make(map[string]OhttpsRequest)
	dirs, err := os.ReadDir(*outdir)

	if err == nil {
		for _, v := range dirs {
			certData, _ := os.ReadFile(*outdir + "/" + v.Name() + "/full.json")
			var data OhttpsRequest
			json.Unmarshal(certData, &data)
			certs[v.Name()] = data
		}
	}
	reader := bytes.NewBuffer([]byte{})
	tpl.Execute(reader, map[string]any{
		"certs": certs,
	})
	w.Write(reader.Bytes())
}
func CertHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Method", r.Method)
	if r.Method != "POST" {

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"success\":false}"))
	}
	r.ParseForm()

	body, err := io.ReadAll(r.Body)
	fmt.Println("body", err)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"success\":false}"))
	}
	var form OhttpsRequest
	if err := json.Unmarshal(body, &form); err != nil {
		fmt.Println("json", err)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"success\":false}"))
		return
	}
	byt16 := md5.Sum([]byte(fmt.Sprintf("%d:%s", form.Timestamp, *ohttps_token)))
	if hex.EncodeToString(byt16[:]) != form.Sign {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"success\":false}"))
		return
	}

	os.Mkdir(*outdir+"/"+form.Payload.CertificateName, 0755)
	os.WriteFile(*outdir+"/"+form.Payload.CertificateName+"/"+"full.json", body, 0755)
	os.WriteFile(*outdir+"/"+form.Payload.CertificateName+"/"+"domains", []byte(strings.Join(form.Payload.CertificateDomains, "\n")), 0755)
	os.WriteFile(*outdir+"/"+form.Payload.CertificateName+"/"+"cert.key", []byte(form.Payload.CertificateCertKey), 0755)
	os.WriteFile(*outdir+"/"+form.Payload.CertificateName+"/"+"fullchain.pem", []byte(form.Payload.CertificateFullchainCerts), 0755)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"success\":true}"))
}

var port *int = flag.Int("port", 8080, "")
var username *string = flag.String("user", "user", "")
var password *string = flag.String("pass", "pass", "")
var ohttps_token *string = flag.String("ohttps_token", "", "")
var outdir *string = flag.String("outdir", "certs", "")

func main() {
	flag.Parse()
	ln, err := net.Listen("tcp", ":"+fmt.Sprint(*port))
	if err != nil {
		log.Fatalln(err)
	}
	outDirInfo, err := os.Stat(*outdir)
	if err != nil {
		os.MkdirAll(*outdir, 0755)
	} else {
		if !outDirInfo.IsDir() {
			log.Fatalln(*outdir + " is not dir.")
		}
	}

	mutx := http.NewServeMux()
	fmt.Printf("listen 0.0.0.0:%d\n", *port)
	mutx.HandleFunc("/", Auth(IndexHandler, *username, *password))
	mutx.HandleFunc("/cert", CertHandler)
	http.Serve(ln, mutx)
}
