package client

import (
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidTOTPSecret reports a TOTPSecret that is not a base32 shared
// secret, and from which no one-time code can be derived.
//
// It is a sentinel so that the provider can name the attribute to fix. The
// upstream client rejects such a secret too, but with an error that carries no
// sentinel and arrives from the same call as an unreachable server, so a typo
// in the configuration would otherwise be reported as a connection failure.
var ErrInvalidTOTPSecret = errors.New("invalid totp secret")

// ValidateTOTPSecret reports whether secret decodes into the key a one-time
// code is derived from.
//
// It accepts what the upstream client accepts, and for the same reason: a
// secret copied out of Uptime Kuma's two-factor dialog carries spaces,
// hyphens, lower case or no padding, and none of that changes the key. See
// WithTOTPSecret in github.com/breml/go-uptime-kuma-client.
//
// The check is made here as well as upstream so that a malformed secret fails
// once, before the first connection attempt, instead of being retried as
// though the server were at fault.
func ValidateTOTPSecret(secret string) error {
	normalized := strings.ToUpper(strings.NewReplacer(" ", "", "-", "", "=", "").Replace(secret))
	if normalized == "" {
		return fmt.Errorf("%w: empty", ErrInvalidTOTPSecret)
	}

	if padding := len(normalized) % 8; padding != 0 {
		normalized += strings.Repeat("=", 8-padding)
	}

	_, err := base32.StdEncoding.DecodeString(normalized)
	if err != nil {
		return fmt.Errorf("%w: not base32: %w", ErrInvalidTOTPSecret, err)
	}

	return nil
}
