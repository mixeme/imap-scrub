package lib

import "strings"

// IsSMIMEProtected reports whether a message's top-level Content-Type marks it
// as cryptographically signed mail that remove_attachments must not rewrite when
// keep_signatures is true: S/MIME per RFC 8551 (opaque pkcs7-mime, pkcs7-signature,
// multipart/signed with a pkcs7 protocol) or OpenPGP per RFC 3156 (standalone
// application/pgp-signature or multipart/signed with that protocol).
func IsSMIMEProtected(contentType string, params map[string]string) bool {
	switch strings.ToLower(contentType) {
	case "application/pkcs7-mime", "application/x-pkcs7-mime",
		"application/pkcs7-signature", "application/x-pkcs7-signature",
		"application/pgp-signature":
		return true
	case "multipart/signed":
		protocol := strings.ToLower(strings.TrimSpace(params["protocol"]))
		switch protocol {
		case "application/pkcs7-signature", "application/x-pkcs7-signature",
			"application/pgp-signature":
			return true
		}
	}
	return false
}
