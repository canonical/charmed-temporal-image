package authorizer

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const serviceAccountSuffix = ".iam.gserviceaccount.com"

// TokenInfo holds information parsed from a Google OAuth token.
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

// Verifier provides configuration parameters for verifying Google OAuth tokens.
type Verifier struct {
	GoogleClientID string
	TokenURL       string
	RequiredScope  string
}

// NewVerifier returns a new Verifier implementation.
func NewVerifier(clientID string, tokenUrl string, requiredScope string) *Verifier {
	return &Verifier{
		GoogleClientID: clientID,
		TokenURL:       tokenUrl,
		RequiredScope:  requiredScope,
	}
}

// GetTokenInfo fetches a given access token's information.
func (v Verifier) GetTokenInfo(accessToken string) (*TokenInfo, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", v.TokenURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request error: %s", resp.Status)
	}

	var tokenInfo TokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, err
	}

	return &tokenInfo, nil
}

// VerifyToken verifies that a given TokenInfo is valid by
// checking that the required scope, email and expiry time
// restrictions exist.
func (v Verifier) VerifyToken(token *TokenInfo) error {
	intExp, err := strconv.ParseInt(token.Exp, 0, 64)
	if err != nil {
		return fmt.Errorf("error validating token: %v", err)
	}

	expirationTime := time.Unix(intExp, 0)
	currentTime := time.Now()

	if !strings.HasSuffix(token.Email, serviceAccountSuffix) && v.GoogleClientID != "" && token.Azp != v.GoogleClientID {
		return errors.New("incorrect token client id")
	}

	if v.RequiredScope != "" && !strings.Contains(token.Scope, v.RequiredScope) {
		return errors.New("token scope must include email")
	}

	if token.EmailVerified != "true" {
		return errors.New("token email not verified")
	}

	if currentTime.After(expirationTime) {
		return errors.New("token expired")
	}

	return nil
}
