package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/errgo.v1"
)

const emailScope = "https://www.googleapis.com/auth/userinfo.email"

type tokenResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

var ErrNoEmailScope = errgo.New("token scope must include email")
var ErrEmailNotVerified = errgo.New("token email not verified")

// ClientID returns the '<env>-google-client-id' snapctl configuration.
func ClientID() (string, error) {
	env := os.Getenv("TCTL_ENVIRONMENT")
	clientID, err := GetSnapctlArg("google-client-id")
	if err != nil {
		return "", err
	}

	if clientID == "" {
		return "", fmt.Errorf("no google-client-id found for %v environment. use 'sudo snap set tctl %v-google-client-id=\"<client_id>\"'", env, env)
	}

	return clientID, nil
}

// ClientSecret returns the '<env>-google-client-secret' snapctl configuration.
func ClientSecret() (string, error) {
	env := os.Getenv("TCTL_ENVIRONMENT")
	clientSecret, err := GetSnapctlArg("google-client-secret")
	if err != nil {
		return "", err
	}

	if clientSecret == "" {
		return "", fmt.Errorf("no google-client-secret found for %v environment. use 'sudo snap set tctl %v-google-client-secret=\"<client_secret>\"'", env, env)
	}

	return clientSecret, nil
}

// FetchValidToken checks for the existence of a valid Google OAuth access token.
// If found, it returns the token if it is not expired.
// If expired, it refreshes the access token and returns it.
func FetchValidToken(clientID string, clientSecret string) (string, error) {
	path := os.Getenv("SNAP_USER_DATA")
	accessToken, err := readTokenFromFile(path, "access")
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "error reading access token from file at path: %v, %v", path, err)
		}

		return "", err
	}

	err = verifyToken(accessToken)
	if err != nil {
		// Check if original token is missing scope or email verification
		if errgo.Cause(err) == ErrNoEmailScope || errgo.Cause(err) == ErrEmailNotVerified {
			return "", err
		}

		// Refresh access token is a refresh token is available
		refreshToken, err := readTokenFromFile(path, "refresh")
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				fmt.Fprintf(os.Stderr, "error reading refresh token from file at path: %v, %v", path, err)
			}

			return "", err
		}

		respToken, err := refreshAccessToken(clientID, clientSecret, refreshToken)
		if err != nil {
			return "", err
		}

		fmt.Fprintf(os.Stdout, "access token refreshed\n")
		return respToken.AccessToken, nil
	}

	return accessToken, nil
}

// verifyToken verifies that a given TokenInfo is valid by
// checking that the required scope, email and expiry time
// restrictions exist.
func verifyToken(accessToken string) error {
	token, err := getTokenInfo(accessToken)
	if err != nil {
		return fmt.Errorf("error fetching token info: %v. use tctl login to refresh token", err)
	}

	intExp, err := strconv.ParseInt(token.Exp, 0, 64)
	if err != nil {
		return fmt.Errorf("error validating token: %v", err)
	}

	expirationTime := time.Unix(intExp, 0)
	currentTime := time.Now()

	if !strings.Contains(token.Scope, emailScope) {
		return errgo.WithCausef(nil, ErrNoEmailScope, "")
	}

	if token.EmailVerified != "true" {
		return errgo.WithCausef(nil, ErrEmailNotVerified, "")
	}

	if currentTime.After(expirationTime) {
		return errors.New("token expired")
	}

	return nil
}

// WriteTokenToFile writes an access token to a file
// in the specified directory.
func WriteTokenToFile(directory string, token string, token_type string) error {
	env := os.Getenv("TCTL_ENVIRONMENT")
	filePath := filepath.Join(directory, fmt.Sprintf("%v_%v_token.txt", env, token_type))
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(token)
	if err != nil {
		return err
	}

	return nil
}

// readTokenFromFile reads an access token from a file
// in the specified directory.
func readTokenFromFile(directory string, token_type string) (string, error) {
	env := os.Getenv("TCTL_ENVIRONMENT")
	filePath := filepath.Join(directory, fmt.Sprintf("%v_%v_token.txt", env, token_type))
	data, err := ioutil.ReadFile(filePath)

	if err != nil {
		if pathErr, ok := err.(*os.PathError); ok {
			// Check if file does not exist
			if pathErr.Err == os.ErrNotExist {
				return "", os.ErrNotExist
			}
		}

		return "", err
	}

	token := string(data)
	return token, nil
}

type TokenInfo struct {
	Azp           string `json:"azp"`
	Aud           string `json:"aud"`
	Sub           string `json:"sub"`
	Scope         string `json:"scope"`
	Exp           string `json:"exp"`
	ExpiresIn     string `json:"expires_in"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	AccessType    string `json:"access_type"`
}

// getTokenInfo fetches a given access token's information.
func getTokenInfo(accessToken string) (*TokenInfo, error) {
	url := "https://www.googleapis.com/oauth2/v3/tokeninfo"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %s", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request error: %s", response.Status)
	}

	var tokenInfo TokenInfo
	if err := json.NewDecoder(response.Body).Decode(&tokenInfo); err != nil {
		return nil, fmt.Errorf("error decoding token response: %s", err)
	}

	return &tokenInfo, nil
}

// refreshAccessToken refreshes an access token using a refresh token and
// the Google OAuth2 token endpoint.
func refreshAccessToken(clientID string, clientSecret string, refreshToken string) (*tokenResponse, error) {
	tokenEndpoint := "https://oauth2.googleapis.com/token"

	formData := url.Values{}
	formData.Set("client_id", clientID)
	formData.Set("client_secret", clientSecret)
	formData.Set("refresh_token", refreshToken)
	formData.Set("grant_type", "refresh_token")

	response, err := http.PostForm(tokenEndpoint, formData)
	if err != nil {
		return nil, fmt.Errorf("request error: %s", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status code: %d", response.StatusCode)
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(response.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	path := os.Getenv("SNAP_USER_DATA")
	err = WriteTokenToFile(path, tokenResp.AccessToken, "access")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing access token: %v\n", err)
	}

	fmt.Fprintf(os.Stdout, "access token refreshed at %v\n", path)

	return &tokenResp, nil
}

// GetSnapctlArg retrieves a configuration value using the "snapctl get" command
// with the specified argument and the current TCTL_ENVIRONMENT.
func GetSnapctlArg(arg string) (string, error) {
	env := os.Getenv("TCTL_ENVIRONMENT")
	execCmd := exec.Command("snapctl", "get", fmt.Sprintf("%v-%v", env, arg))

	output, err := execCmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}
