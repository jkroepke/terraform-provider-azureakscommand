package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func GetOidcTokenFromGithubActions(oidcRequestUrl string, oidcRequestToken string) (string, error) {
	requestURL := fmt.Sprintf("%s&audience=%s", oidcRequestUrl, "api://AzureADTokenExchange")
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "bearer "+oidcRequestToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var payload struct {
		Value string `json:"value"`
	}

	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&payload); err != nil {
		return "", err
	}
	return payload.Value, nil
}
