package utils

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultTimeout = 10 * time.Second

func SendRequest(domain, phpSessId, method, urlPath, body string, extraHeaders *http.Header) (*http.Response, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	fullURL := fmt.Sprintf("https://%s%s", domain, urlPath)
	req, err := http.NewRequestWithContext(ctx, method, fullURL, strings.NewReader(body))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create new request: %w", err)
	}

	req.Header.Set("Cookie", fmt.Sprintf("PHPSESSID=%s", phpSessId))
	req.Header.Set("Host", domain)
	req.Header.Set("Origin", fmt.Sprintf("https://%s", domain))

	if extraHeaders != nil {
		for k, v := range *extraHeaders {
			req.Header[k] = v
		}
	}

	client := &http.Client{
		Timeout: defaultTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to execute request: %w", err)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, "", fmt.Errorf("failed to read response body: %w", err)
	}

	return resp, string(respBody), nil
}

func SendXmlRequest(domain, phpSessId, method, urlPath, body string, extraHeaders *http.Header) (*http.Response, string, error) {
	if extraHeaders == nil {
		extraHeaders = &http.Header{}
	}
	extraHeaders.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	extraHeaders.Set("X-Requested-With", "XMLHttpRequest")
	return SendRequest(domain, phpSessId, method, urlPath, body, extraHeaders)
}

func PerformUpstreamRequest(domain, phpSessId string) (string, error) {
	resp, body, err := SendXmlRequest(domain, phpSessId, http.MethodGet, "/", "", nil)
	if err != nil {
		return "", fmt.Errorf("error sending XML request: %w", err)
	}

	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusUnauthorized {
		return "", errors.New("not authenticated: invalid PHPSESSID")
	} else if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	const marker = "/planner/main/user/"
	idx := strings.Index(body, marker)
	if idx == -1 {
		return "", errors.New("failed to locate user marker in response")
	}

	start := idx + len(marker)
	remaining := body[start:]
	end := strings.Index(remaining, "/")
	if end == -1 {
		return "", errors.New("failed to parse user ID from response")
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
