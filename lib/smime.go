package lib

import (
	"bytes"
	"strings"
)

const (
	pgpPublicKeyBegin = "-----BEGIN PGP PUBLIC KEY BLOCK-----"
	pgpPublicKeyEnd   = "-----END PGP PUBLIC KEY BLOCK-----"
)

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

// IsOpenPGPPublicKeyPart reports whether an attachment part is a standalone
// OpenPGP public key (RFC 3156 armor or RFC 9788 application/pgp-keys).
func IsOpenPGPPublicKeyPart(contentType, filename string, body []byte) bool {
	mediaType := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	switch mediaType {
	case "application/pgp-keys", "application/pgp-key":
		return true
	}

	return hasPGPPublicKeyArmor(body)
}

func hasPGPPublicKeyArmor(body []byte) bool {
	if len(body) == 0 {
		return false
	}
	if bytes.Contains(body, []byte("-----BEGIN PGP PRIVATE KEY BLOCK-----")) ||
		bytes.Contains(body, []byte("-----BEGIN PGP MESSAGE-----")) ||
		bytes.Contains(body, []byte("-----BEGIN PGP SIGNATURE-----")) {
		return false
	}
	return bytes.Contains(body, []byte(pgpPublicKeyBegin)) &&
		bytes.Contains(body, []byte(pgpPublicKeyEnd))
}
