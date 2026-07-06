/* ============================================================================
   Section 16.2 : Les primitives crypto/* — les bonnes briques, bien posées
   Description : Le tour d'horizon des primitives de la section, chacune posée
                 correctement. Aléa → crypto/rand ; empreinte → SHA-256 ;
                 mot de passe → bcrypt (lent, salé, et qui REJETTE au-delà de
                 72 octets sur x/crypto récent) ; dérivation → HKDF ;
                 comparaison de secrets → subtle/hmac.Equal (temps constant) ;
                 chiffrement → AES-GCM (AEAD, altération détectée) ; signature
                 → Ed25519. « N'écrivez pas de crypto : assemblez la stdlib. »
   Fichier source : 16.2 (02-cryptographie-tls.md)
   ============================================================================ */

package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func demoPrimitives() {
	// §1.1 crypto/rand : toute valeur sensible en vient (jamais math/rand).
	tok := make([]byte, 16)
	_, _ = rand.Read(tok)
	fmt.Printf("crypto/rand : jeton = %s…\n", base64.RawURLEncoding.EncodeToString(tok)[:10])

	// §1.2 SHA-256 : empreinte d'intégrité (pas pour les mots de passe).
	sum := sha256.Sum256([]byte("données"))
	fmt.Printf("sha256      : %x…\n", sum[:6])

	// §1.3 bcrypt : lent, salé — et REJETTE au-delà de 72 octets.
	_, errLong := bcrypt.GenerateFromPassword([]byte(strings.Repeat("a", 73)), bcrypt.DefaultCost)
	h, _ := bcrypt.GenerateFromPassword([]byte("motdepasse"), bcrypt.DefaultCost)
	fmt.Printf("bcrypt      : 73 octets rejeté=%t · vérif=%t\n",
		errLong != nil, bcrypt.CompareHashAndPassword(h, []byte("motdepasse")) == nil)

	// §1.4 HKDF : dériver plusieurs clés d'un secret à forte entropie.
	key, _ := hkdf.Key(sha256.New, []byte("secret-maitre-aleatoire-32-octets!"), []byte("sel"), "session v1", 32)
	fmt.Printf("hkdf        : %d octets dérivés\n", len(key))

	// §1.5 subtle + hmac.Equal : comparaison à temps constant.
	mac := hmac.New(sha256.New, []byte("clé"))
	mac.Write([]byte("message"))
	digest := mac.Sum(nil)
	fmt.Printf("temps const : subtle=%t · hmac.Equal=%t\n",
		subtle.ConstantTimeCompare([]byte("abc"), []byte("abc")) == 1,
		hmac.Equal(digest, digest))

	// §1.6 AES-GCM (AEAD) : chiffre ET authentifie — l'altération est détectée.
	ck := make([]byte, 32)
	_, _ = rand.Read(ck)
	block, _ := aes.NewCipher(ck)
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	_, _ = rand.Read(nonce)
	ct := gcm.Seal(nonce, nonce, []byte("message secret"), nil)
	pt, _ := gcm.Open(nil, ct[:gcm.NonceSize()], ct[gcm.NonceSize():], nil)
	tampered := bytes.Clone(ct)
	tampered[len(tampered)-1] ^= 0xFF
	_, errTamper := gcm.Open(nil, tampered[:gcm.NonceSize()], tampered[gcm.NonceSize():], nil)
	fmt.Printf("aes-gcm     : déchiffré=%q · altération rejetée=%t\n", pt, errTamper != nil)

	// §1.7 Ed25519 : signature. La clé EST la graine lue dans le reader (cf. §1.1).
	seed := bytes.Repeat([]byte{7}, 32)
	_, priv1, _ := ed25519.GenerateKey(bytes.NewReader(seed))
	_, priv2, _ := ed25519.GenerateKey(bytes.NewReader(seed))
	sig := ed25519.Sign(priv1, []byte("m"))
	fmt.Printf("ed25519     : reader déterministe→clés identiques=%t · vérif=%t\n",
		bytes.Equal(priv1, priv2),
		ed25519.Verify(priv1.Public().(ed25519.PublicKey), []byte("m"), sig))
}
