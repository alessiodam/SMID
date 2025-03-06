package utils

import (
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func SendRequest(domain, phpSessId, method, url, body string, extraHeaders *http.Header) (*http.Response, string, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("https://%s%s", domain, url), strings.NewReader(body))
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("Cookie", fmt.Sprintf("PHPSESSID=%s", phpSessId))
	req.Header.Set("Host", domain)
	req.Header.Set("Origin", fmt.Sprintf("https://%s", domain))

	if extraHeaders != nil {
		for k, v := range *extraHeaders {
			req.Header[k] = v
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, "", err
	}
	return resp, string(respBody), nil
}

func SendXmlRequest(domain, phpSessId, method, url, body string, extraHeaders *http.Header) (*http.Response, string, error) {
	if extraHeaders == nil {
		extraHeaders = &http.Header{}
	}
	extraHeaders.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	extraHeaders.Set("X-Requested-With", "XMLHttpRequest")
	return SendRequest(domain, phpSessId, method, url, body, extraHeaders)
}

func PerformUpstreamRequest(domain, phpSessId string) (string, error) {
	resp, body, err := SendXmlRequest(domain, phpSessId, "GET", "/", "", nil)
	if err != nil {
		return "", fmt.Errorf("error sending XML request: %v", err)
	}
	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusUnauthorized {
		return "", errors.New("not authenticated, invalid PhpSessId (PHPSESSID)")
	} else if resp.StatusCode != http.StatusOK {
		return "", errors.New("could not check if authenticated")
	}
	idx := strings.Index(body, "/planner/main/user/")
	if idx == -1 {
		return "", errors.New("could not extract user ID from response")
	}
	start := idx + len("/planner/main/user/")
	remaining := body[start:]
	end := strings.Index(remaining, "/")
	if end == -1 {
		return "", errors.New("could not extract user ID from response")
	}
	extracted := remaining[:end]
	parts := strings.Split(extracted, "_")
	if len(parts) < 2 {
		return "", errors.New("unexpected user ID format")
	}
	uniqueUserID := parts[0] + "_" + parts[1]
	return uniqueUserID, nil
}

func HashPhpSessId(phpSessId string) string {
	hash := sha512.Sum512([]byte(phpSessId))
	return hex.EncodeToString(hash[:])
}
