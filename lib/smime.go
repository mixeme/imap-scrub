package lib

import "strings"

// IsSMIMEProtected reports whether a message's top-level Content-Type marks it
// as S/MIME signed or enveloped data per RFC 8551: either an opaque
// application/pkcs7-mime part (smime.p7m / smime.p7z), a standalone
// application/pkcs7-signature part (smime.p7s), or a multipart/signed
// container whose protocol is one of those signature types.
func IsSMIMEProtected(contentType string, params map[string]string) bool {
	switch strings.ToLower(contentType) {
	case "application/pkcs7-mime", "application/x-pkcs7-mime",
		"application/pkcs7-signature", "application/x-pkcs7-signature":
		return true
	case "multipart/signed":
		protocol := strings.ToLower(strings.TrimSpace(params["protocol"]))
		return protocol == "application/pkcs7-signature" || protocol == "application/x-pkcs7-signature"
	}
	return false
}
