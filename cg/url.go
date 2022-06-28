package cg

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// trimURL removes the protocol version and trailing slashes.
func trimURL(url string) string {
	url = strings.TrimSuffix(url, "/")
	parts := strings.Split(url, "://")
	if len(parts) < 2 {
		return url
	}
	return strings.Join(parts[1:], "://")
}

// baseURL prepends `protocol + "://"` or `protocol + "s://"` to the url depending on TLS support.
func baseURL(protocol string, tls bool, trimmedURL string, a ...any) string {
	trimmedURL = fmt.Sprintf(trimmedURL, a...)
	if tls {
		return protocol + "s://" + trimmedURL
	} else {
		return protocol + "://" + trimmedURL
	}
}

// isTLS verifies the TLS certificate of a trimmed URL.
func isTLS(trimmedURL string) bool {
	url, err := url.Parse("https://" + trimmedURL)
	if err != nil {
		return false
	}
	host := url.Host
	if url.Port() == "" {
		host = host + ":443"
	}

	conn, err := tls.Dial("tcp", url.Host, &tls.Config{})
	if err != nil {
		return false
	}
	defer conn.Close()

	err = conn.VerifyHostname(url.Hostname())
	if err != nil {
		return false
	}

	expiry := conn.ConnectionState().PeerCertificates[0].NotAfter
	if time.Now().After(expiry) {
		return false
	}

	return true
}
