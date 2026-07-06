/* ============================================================================
   Section 16.2 : TLS post-quantique (Go 1.26) et crypto/hpke
   Description : Les deux nouveautés de Go 1.26. (§2.4) Un serveur TLS 1.3 et
                 son client négocient AUTOMATIQUEMENT une courbe hybride
                 post-quantique (X25519MLKEM768) — aucun changement de code.
                 (§2.5) crypto/hpke chiffre un message vers la clé publique
                 d'un destinataire HORS de TLS, avec un KEM hybride
                 post-quantique : émission (Seal) et déchiffrement (Open).
   Fichier source : 16.2 (02-cryptographie-tls.md)
   ============================================================================ */

package main

import (
	"crypto/hpke"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
)

func demoTLSPostQuantique() {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	}))
	ts.TLS = &tls.Config{MinVersion: tls.VersionTLS13}
	ts.StartTLS()
	defer ts.Close()

	resp, err := ts.Client().Get(ts.URL)
	if err != nil {
		fmt.Println("tls:", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("TLS         : version=0x%x · courbe négociée=%v (hybride post-quantique)\n",
		resp.TLS.Version, tls.CurveID(resp.TLS.CurveID))
}

func demoHPKE() {
	kem, kdf, aead := hpke.MLKEM768X25519(), hpke.HKDFSHA256(), hpke.AES256GCM()

	// Destinataire : paire de clés (KEM hybride post-quantique).
	priv, err := kem.GenerateKey()
	if err != nil {
		fmt.Println("hpke keygen:", err)
		return
	}
	pub, err := kem.NewPublicKey(priv.PublicKey().Bytes())
	if err != nil {
		fmt.Println("hpke pub:", err)
		return
	}

	// Émetteur → chiffre vers la clé publique ; destinataire → déchiffre.
	ct, err := hpke.Seal(pub, kdf, aead, []byte("contexte app"), []byte("message HPKE"))
	if err != nil {
		fmt.Println("hpke seal:", err)
		return
	}
	pt, err := hpke.Open(priv, kdf, aead, []byte("contexte app"), ct)
	fmt.Printf("hpke        : roundtrip=%q (err=%v)\n", pt, err)
}
