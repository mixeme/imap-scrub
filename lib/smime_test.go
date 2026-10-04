package lib

import "testing"

func TestIsSMIMEProtected(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		params      map[string]string
		want        bool
	}{
		{"opaque pkcs7-mime (smime.p7m)", "application/pkcs7-mime", map[string]string{"smime-type": "signed-data", "name": "smime.p7m"}, true},
		{"opaque pkcs7-mime compressed (smime.p7z)", "application/pkcs7-mime", map[string]string{"smime-type": "compressed-data", "name": "smime.p7z"}, true},
		{"x-pkcs7-mime alias", "application/x-pkcs7-mime", nil, true},
		{"standalone pkcs7-signature", "application/pkcs7-signature", map[string]string{"name": "smime.p7s"}, true},
		{"x-pkcs7-signature alias", "application/x-pkcs7-signature", nil, true},
		{"multipart/signed with pkcs7-signature protocol", "multipart/signed", map[string]string{"protocol": "application/pkcs7-signature"}, true},
		{"multipart/signed protocol case-insensitive", "multipart/signed", map[string]string{"protocol": "APPLICATION/PKCS7-SIGNATURE"}, true},
		{"multipart/signed with x-pkcs7-signature protocol", "multipart/signed", map[string]string{"protocol": "application/x-pkcs7-signature"}, true},
		{"multipart/signed with pgp protocol", "multipart/signed", map[string]string{"protocol": "application/pgp-signature"}, true},
		{"multipart/signed pgp protocol case-insensitive", "multipart/signed", map[string]string{"protocol": "APPLICATION/PGP-SIGNATURE"}, true},
		{"standalone pgp-signature", "application/pgp-signature", nil, true},
		{"multipart/signed without protocol", "multipart/signed", nil, false},
		{"multipart/mixed", "multipart/mixed", nil, false},
		{"plain text", "text/plain", nil, false},
		{"octet-stream attachment", "application/octet-stream", map[string]string{"name": "file.bin"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSMIMEProtected(tt.contentType, tt.params); got != tt.want {
				t.Errorf("IsSMIMEProtected(%q, %v) = %v, want %v", tt.contentType, tt.params, got, tt.want)
			}
		})
	}
}

func TestIsOpenPGPPublicKeyPart(t *testing.T) {
	publicKeyBody := "-----BEGIN PGP PUBLIC KEY BLOCK-----\r\n\r\n" +
		"mQENBFakeKeyIDABC\r\n" +
		"=abcd\r\n" +
		"-----END PGP PUBLIC KEY BLOCK-----\r\n"
	signatureBody := "-----BEGIN PGP SIGNATURE-----\r\n\r\n" +
		"iEYEARECAAYFAkl1EZkACgkQ\r\n" +
		"=cacd\r\n" +
		"-----END PGP SIGNATURE-----\r\n"
	privateKeyBody := "-----BEGIN PGP PRIVATE KEY BLOCK-----\r\n\r\n" +
		"fake\r\n" +
		"-----END PGP PRIVATE KEY BLOCK-----\r\n"

	tests := []struct {
		name        string
		contentType string
		filename    string
		body        string
		want        bool
	}{
		{"application/pgp-keys", "application/pgp-keys", "key.asc", "binary", true},
		{"application/pgp-key", "application/pgp-key", "", "", true},
		{"armored octet-stream", "application/octet-stream", "pubkey.asc", publicKeyBody, true},
		{"armored text plain", "text/plain", "key.asc", publicKeyBody, true},
		{"pgp signature armor", "application/octet-stream", "signature.asc", signatureBody, false},
		{"private key armor", "application/octet-stream", "secret.asc", privateKeyBody, false},
		{"plain binary", "application/octet-stream", "file.bin", "hello", false},
		{"public key markers without .asc name", "application/octet-stream", "file.bin", publicKeyBody, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsOpenPGPPublicKeyPart(tt.contentType, tt.filename, []byte(tt.body)); got != tt.want {
				t.Errorf("IsOpenPGPPublicKeyPart() = %v, want %v", got, tt.want)
			}
		})
	}
}
