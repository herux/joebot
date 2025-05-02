package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
)

func (server *Server) wrapTokenAuth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the auth token from cookie
		token, err := r.Cookie("authToken")
		if err != nil || token.Value == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		iPs, err := server.getUserIPWhitelisted(token.Value)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Unable to parse IP", http.StatusInternalServerError)
			return
		}

		remoteIP := net.ParseIP(host)
		if remoteIP == nil {
			http.Error(w, "Invalid IP address", http.StatusInternalServerError)
			return
		}

		isAuthorized := false
		for _, ip := range iPs {
			ip = strings.TrimSpace(ip)
			parsedIP := net.ParseIP(ip)
			if parsedIP == nil {
				continue
			}

			if remoteIP.Equal(parsedIP) {
				isAuthorized = true
				break
			}
		}

		fmt.Printf("isAuthorized: %v\n", isAuthorized)
		if !isAuthorized {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func (server *Server) getUserIPWhitelisted(token string) ([]string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://"+server.options.JoebotWebPortalHost+":"+server.options.JoebotWebPortalPort+"/api/users/ip-whitelisted", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user IP whitelisted: %v", err)
	}
	defer resp.Body.Close()

	var ips []string
	if err := json.NewDecoder(resp.Body).Decode(&ips); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %v", err)
	}

	return ips, nil
}
