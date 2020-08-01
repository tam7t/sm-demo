package main

import (
	"context"
	"crypto/subtle"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	sm "cloud.google.com/go/secretmanager/apiv1"
	smpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

func main() {
	ctx := context.Background()

	client, err := sm.NewClient(ctx)
	if err != nil {
		log.Fatal("Could not initialize secret client: ", err)
	}

	d := dog{
		SM: client,
	}

	// occasionally reload the secret
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				d.loadSecret()
			}
		}
	}()

	// start the http server
	http.HandleFunc("/", d.bark)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := http.Server{
		Addr: ":" + port,
	}

	log.Printf("Listening on port %s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Printf("server error: %s", err)
	}
}

// dog holds configuration necessary for the translating slack handlers.
type dog struct {
	SM                *sm.Client
	VerificationToken string
}

func (d *dog) loadSecret() {
	secret, err := d.SM.AccessSecretVersion(context.Background(), &smpb.AccessSecretVersionRequest{
		Name: os.Getenv("SECRET_RESOURCE_NAME"),
	})
	if err != nil {
		log.Printf("Error loading secret: %s", err)
		return
	}

	d.VerificationToken = string(secret.GetPayload().GetData())
}

// bark handles the "/bark" slack command webhook and translates the provided
// text into dog.
func (d *dog) bark(w http.ResponseWriter, r *http.Request) {
	if d.VerificationToken == "" {
		e(w, "no secret loaded", http.StatusInternalServerError)
		return
	}

	if subtle.ConstantTimeCompare([]byte(r.FormValue("token")), []byte(d.VerificationToken)) == 0 {
		e(w, "slack verification failed", http.StatusUnauthorized)
		return
	}

	if r.FormValue("command") != "/bark" {
		e(w, "wrong command", http.StatusBadRequest)
		return
	}

	w.Write([]byte("the dog says:"))
	for range strings.Fields(r.FormValue("text")) {
		w.Write([]byte(" bark"))
	}
}

// e helps write HTTP error messages.
func e(w http.ResponseWriter, msg string, code int) {
	w.WriteHeader(code)
	w.Write([]byte(msg))
}
