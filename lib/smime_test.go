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
		{"multipart/signed with pgp protocol", "multipart/signed", map[string]string{"protocol": "application/pgp-signature"}, false},
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
