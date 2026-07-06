/* ============================================================================
   Section 16.1 : OWASP appliqué à Go (injection, XSS, validation d'entrées)
   Description : Les défenses de la section, en un programme autonome (stdlib
                 pure). Toute injection se règle en SÉPARANT LE CODE DE LA
                 DONNÉE : os/exec sans shell (le « ; » reste un argument),
                 os.Root qui confine les accès fichiers, html/template à
                 l'échappement CONTEXTUEL (une URL javascript: devient
                 #ZgotmplZ), MaxBytesReader contre le DoS, la validation par
                 « parser plutôt que valider » (net/mail), la liste blanche, et
                 la défense SSRF (validation de la destination). L'injection SQL
                 est démontrée à part, sur une vraie base : voir ./sqli.
   Fichier source : 16.1 (01-owasp-go.md)
   Lancer : go run .
   ============================================================================ */

package main

import (
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

func main() {
	demoExecSansShell()
	demoOsRoot()
	demoHTMLTemplate()
	demoMaxBytes()
	demoValidation()
	demoSSRF()
}

// §1.2 : os/exec n'invoque PAS de shell — les métacaractères sont inertes.
func demoExecSansShell() {
	out, _ := exec.Command("echo", "x.png; rm -rf /").Output()
	fmt.Printf("exec sans shell : %q — le « ; » n'est pas interprété\n", strings.TrimSpace(string(out)))
}

// §1.3 : os.Root (Go 1.24) confine les accès sous une racine et rejette « .. ».
func demoOsRoot() {
	_ = os.MkdirAll("srv", 0o755)
	_ = os.WriteFile("srv/allowed.txt", []byte("ok"), 0o644)
	_ = os.WriteFile("secret_hors.txt", []byte("secret"), 0o644)
	root, err := os.OpenRoot("srv")
	if err != nil {
		return
	}
	defer root.Close()
	_, inErr := root.Open("allowed.txt")
	_, outErr := root.Open("../secret_hors.txt")
	fmt.Printf("os.Root : allowed.txt→ok=%t · ../secret_hors.txt→rejeté=%t\n", inErr == nil, outErr != nil)
}

// §2.1/2.2 : html/template échappe, et le fait de façon CONTEXTUELLE.
func demoHTMLTemplate() {
	var body strings.Builder
	template.Must(template.New("b").Parse(`<p>{{.C}}</p>`)).
		Execute(&body, struct{ C string }{C: `<script>alert(1)</script>`})
	fmt.Println("html/template (corps) :", body.String())

	var href strings.Builder
	template.Must(template.New("h").Parse(`<a href="{{.U}}">x</a>`)).
		Execute(&href, struct{ U string }{U: "javascript:alert(1)"})
	fmt.Println("html/template (URL)   :", href.String(), "— javascript: neutralisé en #ZgotmplZ")
}

// §3.2 : http.MaxBytesReader plafonne le corps → 413 au-delà.
func demoMaxBytes() {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 10)
		if _, err := r.Body.Read(make([]byte, 100)); err != nil {
			http.Error(w, "trop gros", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(h)
	defer srv.Close()
	resp, err := http.Post(srv.URL, "text/plain", strings.NewReader(strings.Repeat("A", 50)))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	fmt.Printf("MaxBytesReader (50o > 10o) → HTTP %d\n", resp.StatusCode)
}

// §3.3/3.4 : « parser plutôt que valider » (net/mail) + liste blanche.
var allowedRoles = map[string]bool{"reader": true, "editor": true, "admin": true}

func demoValidation() {
	_, badErr := mail.ParseAddress("pas-un-email")
	addr, _ := mail.ParseAddress("alice@exemple.fr")
	fmt.Printf("NewEmail : invalide rejeté=%t · valide → %s\n", badErr != nil, addr.Address)
	fmt.Printf("liste blanche rôles : admin=%t · root=%t (refusé)\n", allowedRoles["admin"], allowedRoles["root"])
}

// §3.5 : SSRF — valider la DESTINATION (schéma + IP non interne).
func safeTarget(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if u.Scheme != "https" {
		return errors.New("schéma non autorisé")
	}
	ips, err := net.LookupIP(u.Hostname())
	if err != nil {
		return err
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
			return errors.New("cible interne interdite")
		}
	}
	return nil
}

func demoSSRF() {
	// IP publique littérale pour le cas « autorisé » : déterministe, sans DNS.
	fmt.Printf("SSRF : https://127.0.0.1→rejeté=%t · http://8.8.8.8→rejeté=%t · https://8.8.8.8→autorisé=%t\n",
		safeTarget("https://127.0.0.1/admin") != nil, // loopback interne
		safeTarget("http://8.8.8.8") != nil,          // schéma non-https
		safeTarget("https://8.8.8.8") == nil)         // IP publique, https
}
