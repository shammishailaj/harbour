package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	oidc "github.com/coreos/go-oidc"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

var (
	clientID          = os.Getenv("OIDC_CLIENT_ID")
	clientSecret      = os.Getenv("OIDC_CLIENT_SECRET")
	oidcDiscovery     = os.Getenv("OIDC_URL")
	oidcLoginCallback = os.Getenv("OIDC_LOGIN_CALLBACK")
)

func main() {
	fmt.Println("Hello from harbour.rocks/iam")

	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, oidcDiscovery)
	if err != nil {
		log.Fatal(err)
	}
	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}
	verifier := provider.Verifier(oidcConfig)

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  oidcLoginCallback,
		Scopes:       []string{oidc.ScopeOpenID, "email"},
	}

	state := "foobar" // Don't do this in production.

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, config.AuthCodeURL(state), http.StatusFound)
	})

	http.HandleFunc("/auth/oidc/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "state did not match", http.StatusBadRequest)
			return
		}

		oauth2Token, err := config.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}
		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
			return
		}
		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		oauth2Token.AccessToken = "*REDACTED*"

		resp := struct {
			OAuth2Token   *oauth2.Token
			IDTokenClaims *json.RawMessage // ID Token payload is just JSON.
		}{oauth2Token, new(json.RawMessage)}

		if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		data, err := json.MarshalIndent(resp, "", "    ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(data)
	})

	log.Printf("listening on http://%s/", "127.0.0.1:5100")
	log.Fatal(http.ListenAndServe("127.0.0.1:5100", nil))
}