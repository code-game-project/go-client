package cg

import (
	"crypto/tls"
	"fmt"
	"net"
	neturl "net/url"
	"strings"
	"time"
)

// trimURL removes the protocol component and trailing slashes.
func trimURL(url string) string {
	u, err := neturl.Parse(url)
	if err != nil {
		return url
	}
	u.Scheme = ""
	return strings.TrimSuffix(u.String(), "/")
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
func isTLS(trimmedURL string) (isTLS bool) {
	url, err := neturl.Parse("https://" + trimmedURL)
	if err != nil {
		return false
	}
	host := url.Host
	if url.Port() == "" {
		host = host + ":443"
	}

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, "tcp", host, &tls.Config{})
	if err != nil {
		return false
	}
	defer conn.Close()

	err = conn.VerifyHostname(url.Hostname())
	if err != nil {
		return false
	}

	expiry := conn.ConnectionState().PeerCertificates[0].NotAfter
	return !time.Now().After(expiry)
}
