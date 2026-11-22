package client

import (
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// requestConfigErrorPrefixes are the prefixes of the plain errors kuma.New
// rejects the API key, the headers, the User-Agent and the base URL's userinfo
// with, see kuma.WithHeaders and kuma.WithAPIKey. The client has no sentinel
// for them, so the prefixes of its messages are all there is to match.
//
// The client's "credentials: " errors, which reject an API key next to a
// username and password, a session token, a TOTP secret or autosetup, are not
// matched: the provider's own validation (validateAPIKey) rejects every such
// combination before the client is called, and the provider never asks for
// autosetup, so they are unreachable from here.
//
//nolint:gochecknoglobals // Read-only table of the client's message prefixes.
var requestConfigErrorPrefixes = []string{"headers: ", "api key: ", userinfoErrorPrefix}

// userinfoErrorPrefix is the prefix of the error kuma.New rejects a base URL
// with userinfo with, next to an API key or an Authorization header.
const userinfoErrorPrefix = "base url: userinfo"

// IsRequestConfigError reports whether err is kuma.New rejecting the
// configured API key, headers or User-Agent, or the userinfo of the base URL
// next to them. Such an error is the configuration's and no further attempt
// overcomes it. The match is on the client's message text, so it is a best
// effort: an error it misses is merely retried and reported as a connection
// failure.
func IsRequestConfigError(err error) bool {
	// The API key sentinels are the server's answer rather than a rejected
	// configuration, and two of them share the "api key: " prefix.
	if err == nil ||
		errors.Is(err, kuma.ErrAPIKeyUnsupported) ||
		errors.Is(err, kuma.ErrAPIKeyRejected) ||
		errors.Is(err, kuma.ErrAPIKeySetupRequired) {
		return false
	}

	msg := err.Error()

	for _, prefix := range requestConfigErrorPrefixes {
		if strings.HasPrefix(msg, prefix) || strings.Contains(msg, ": "+prefix) {
			return true
		}
	}

	return false
}

// IsUserinfoError reports whether err is kuma.New rejecting a base URL with
// userinfo (user:password@) next to an API key or an Authorization header. It
// is a request configuration error, see IsRequestConfigError, whose fix lies
// in the endpoint rather than in the headers.
func IsUserinfoError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()

	return strings.HasPrefix(msg, userinfoErrorPrefix) || strings.Contains(msg, ": "+userinfoErrorPrefix)
}

// IsRefusedRedirect reports whether err is the client refusing a redirect that
// leaves the base URL's origin, an upgrade from http to https on the same host
// included. The client follows no such redirect, so the base URL itself has to
// change, which the error names. Like IsRequestConfigError, it matches the
// client's message text.
func IsRefusedRedirect(err error) bool {
	return IsHTTPSUpgradeRedirect(err) || (err != nil && strings.Contains(err.Error(), "leaves the server's origin"))
}

// IsHTTPSUpgradeRedirect reports whether err is the client refusing a
// redirect from http to https on the same host, see IsRefusedRedirect.
func IsHTTPSUpgradeRedirect(err error) bool {
	return err != nil && strings.Contains(err.Error(), "upgrades to https;")
}

// IsAuthSchemeProbeError reports whether err is the client's auth scheme
// detection getting an answer that is neither Uptime Kuma 2.x nor 3.x: a
// redirect within the server's origin, or a status and Content-Type no Uptime
// Kuma gives, such as an access proxy refusing the request. It matches the
// client's message text, like IsRequestConfigError. Such an error is retried,
// because a 5xx from a proxy may well pass, unless IsPermanentProbeError.
func IsAuthSchemeProbeError(err error) bool {
	if err == nil {
		return false
	}

	msg := err.Error()

	return strings.Contains(msg, "detect auth scheme: GET ") &&
		(strings.Contains(msg, " was redirected to ") || strings.Contains(msg, " answered status "))
}

// RedactURL returns rawURL with the password of its userinfo replaced, see
// url.URL.Redacted, for a message that reaches the user. A value that does not
// parse as a URL is not repeated, because it may hold a password all the same.
func RedactURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "(endpoint that is not a valid URL)"
	}

	return parsed.Redacted()
}

// probeStatusPattern finds the status in the error of an auth scheme detection
// that got an answer no Uptime Kuma gives, see IsAuthSchemeProbeError. The text
// is the client's: "detect auth scheme: GET <url> answered status <code> with
// Content-Type <type>".
var probeStatusPattern = regexp.MustCompile(`detect auth scheme: GET \S+ answered status (\d{3}) with Content-Type `)

// IsPermanentProbeError reports whether err is an auth scheme detection that
// got a 4xx answer no Uptime Kuma gives, see IsAuthSchemeProbeError. Uptime
// Kuma answers the detection with JSON, or with a 404 or 429 the client
// reads as 2.x, so a 4xx is something in front of it refusing the request:
// an access proxy that wants credentials, which another attempt with the same
// configuration does not bring. A 5xx and a redirect within the origin may
// pass, and are retried.
func IsPermanentProbeError(err error) bool {
	if err == nil {
		return false
	}

	match := probeStatusPattern.FindStringSubmatch(err.Error())
	if match == nil {
		return false
	}

	status, convErr := strconv.Atoi(match[1])

	return convErr == nil && status >= 400 && status < 500
}
