// Copyright 2022 Canonical Ltd.

package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/canonical/charmed-temporal-image/tctl-plugins/cmd"
	"github.com/google/uuid"
	"github.com/skratchdot/open-golang/open"
)

const (
	authEndpoint = "https://accounts.google.com/o/oauth2/auth"
	scope        = "openid profile email"
)

var (
	server       *http.Server
	shutdown     = make(chan struct{}) // Custom channel for graceful shutdown signal
	state        = uuid.New().String()
	clientID     string
	clientSecret string
	codeVerifier string
	redirectURI  string
)

func main() {
	env := os.Getenv("TCTL_ENVIRONMENT")
	if env == "dev" {
		fmt.Fprintf(os.Stderr, "login command not defined for development environment.\n")
		os.Exit(1)
	}

	var err error
	clientID, err = cmd.ClientID()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading snap argument '%v-google-client-id': %v", env, err)
		os.Exit(1)
	}

	if clientID == "" {
		fmt.Fprintf(os.Stderr, "no google-client-id found for %v environment. use 'sudo snap set tctl %v-google-client-id=\"<client_id>\"'.\n", env, env)
		os.Exit(1)
	}

	clientSecret, err = cmd.ClientSecret()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading snap argument '%v-google-client-secret': %v", env, err)
		os.Exit(1)
	}

	if clientSecret == "" {
		fmt.Fprintf(os.Stderr, "no google-client-secret found for %v environment. use 'sudo snap set tctl %v-google-client-secret=\"<client_secret>\"'.\n", env, env)
		os.Exit(1)
	}

	// Error is ignored, as a failure to fetch the token will result in the initiation of the login flow.
	token, _ := cmd.FetchValidToken(clientID, clientSecret)
	if token != "" {
		fmt.Fprintf(os.Stdout, "valid access token fetched\n")
		os.Exit(0)
	}

	if err := getToken(); err != nil {
		fmt.Fprintf(os.Stderr, "unable to login: %s\n", err)
		os.Exit(1)
	}
}

// getFreePort returns an available TCP port number on the localhost.
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// getToken returns a valid Google OAuth 2.0 access token.
func getToken() error {
	ctx := context.Background()
	port, err := getFreePort()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting free port: %s\n", err)
		return err
	}

	redirectURI = fmt.Sprintf("http://localhost:%v/oauth2callback", port)
	authURL := generateAuthURL()

	// Automatically open the browser for the user to log in
	if err := open.Run(authURL); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open the browser: %s\n", err)
		return err
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start an HTTP server to handle the callback
	go func() {
		http.HandleFunc("/oauth2callback", handleCallback)
		server = &http.Server{Addr: fmt.Sprintf(":%v", port)}
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "HTTP server error: %s\n", err)
			close(shutdown)
		}
	}()

	// Wait for shutdown signal
	select {
	case <-shutdownChan:
		fmt.Println("Received shutdown signal. Shutting down...")

		// Gracefully shut down the server
		if err := server.Shutdown(ctx); err != nil {
			fmt.Fprintf(os.Stdout, "Error during server shutdown: %s\n", err)
		}
	case <-shutdown:
		fmt.Println("Received access token. Shutting down server...")
		if err := server.Shutdown(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error during server shutdown: %s\n", err)
		}
	}

	return nil
}

// generateAuthURL generates an authorization URL for the user to login with.
func generateAuthURL() string {
	codeVerifier = generateRandomCodeVerifier()
	codeChallenge := generateCodeChallenge(codeVerifier)

	queryParams := url.Values{}
	queryParams.Add("client_id", clientID)
	queryParams.Add("redirect_uri", redirectURI)
	queryParams.Add("scope", scope)
	queryParams.Add("response_type", "code")
	queryParams.Add("state", state)
	queryParams.Add("code_challenge", codeChallenge)
	queryParams.Add("code_challenge_method", "S256")

	authURL := authEndpoint + "?" + queryParams.Encode()
	return authURL
}

// generateRandomCodeVerifier generates a random code verifier string,
// suitable for use in oauth authorization requests.
func generateRandomCodeVerifier() string {
	const codeVerifierLength = 64
	b := make([]byte, codeVerifierLength)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(b)
}

// generateCodeChallenge generates a code challenge string from a code verifier.
// It takes a code verifier as input and calculates a code challenge string
// by applying the SHA-256 hash algorithm to the provided code verifier.
// The resulting hash is then base64 URL-encoded to create a code challenge
// suitable for use in OAuth 2.0 PKCE (Proof Key for Code Exchange) challenges.
func generateCodeChallenge(codeVerifier string) string {
	sha256sum := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(sha256sum[:])
	return codeChallenge
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	defer close(shutdown)
	r.ParseForm()
	if r.FormValue("state") != state {
		http.Error(w, "State mismatch", http.StatusBadRequest)
		fmt.Fprintf(os.Stderr, "State mismatch: %d\n", http.StatusBadRequest)
		return
	}

	authorizationCode := r.FormValue("code")
	if authorizationCode == "" {
		http.Error(w, "Authorization code missing", http.StatusBadRequest)
		fmt.Fprintf(os.Stderr, "Authorization code missing: %d\n", http.StatusBadRequest)
		return
	}

	// Exchange the authorization code for an access token
	tokenResp, err := exchangeCodeForToken(authorizationCode)
	if err != nil {
		http.Error(w, fmt.Sprintf("Token exchange failed: %s", err), http.StatusInternalServerError)
		return
	}

	// Store access token in snap application directory
	path := os.Getenv("SNAP_USER_DATA")
	err = cmd.WriteTokenToFile(path, tokenResp.AccessToken, "access")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing access token: %s\n", err)
	}

	err = cmd.WriteTokenToFile(path, tokenResp.RefreshToken, "refresh")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing refresh token: %s\n", err)
	}

	fmt.Fprintf(w, "Authentication successful. You can close this window.")
}

type tokenResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

// exchangeCodeForToken exchanges an authorization code for an access token
// using the OAuth 2.0 authorization code flow.
func exchangeCodeForToken(authorizationCode string) (*tokenResponse, error) {
	tokenEndpoint := "https://oauth2.googleapis.com/token"

	formData := url.Values{}
	formData.Set("client_id", clientID)
	formData.Set("client_secret", clientSecret)
	formData.Set("code", authorizationCode)
	formData.Set("redirect_uri", redirectURI)
	formData.Set("grant_type", "authorization_code")

	// Include the code verifier in the token request
	formData.Set("code_verifier", codeVerifier)

	response, err := http.PostForm(tokenEndpoint, formData)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status code: %d", response.StatusCode)
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(response.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}
