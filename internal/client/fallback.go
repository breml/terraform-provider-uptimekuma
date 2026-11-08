package client

import (
	"errors"

	kuma "github.com/breml/go-uptime-kuma-client"
)

// PasswordFallbackCause returns the error the password login failed with,
// when err is, or wraps, the *kuma.SessionTokenFallbackError kuma.New reports
// for a session token the server refused followed by a password login that
// failed as well, and nil for every other error.
//
// That error wraps both failures, so errors.Is matches
// kuma.ErrInvalidSessionToken on it whatever the password login failed with;
// this is how a caller judges it by its password half instead. It is
// recognised by its type rather than by the configuration, because a
// configured password does not mean the fallback ran: the client skips it for
// kuma.ErrUserInactive and for a token failure that is not a rejection.
func PasswordFallbackCause(err error) error {
	var fallbackErr *kuma.SessionTokenFallbackError
	if !errors.As(err, &fallbackErr) {
		return nil
	}

	return fallbackErr.PasswordErr
}
