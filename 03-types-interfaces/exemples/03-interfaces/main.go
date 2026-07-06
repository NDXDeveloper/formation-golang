/* ============================================================================
   Section 3.3 : Interfaces implicites — le cœur du design Go
   Description : Satisfaction implicite (Notifier/EmailSender), affectation
                 muette, petites interfaces universelles (firstLine/io.Reader),
                 « accepter des interfaces, retourner des structs », le piège du
                 nil typé, assertions virgule-ok, type switch, capacité
                 optionnelle (io.Closer)
   Fichier source : 03-interfaces.md
   ============================================================================ */

package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Un contrat de comportement : « tout ce qui sait notifier ».
type Notifier interface {
	Notify(msg string) error
}

// EmailSender ne mentionne JAMAIS Notifier — et pourtant il le satisfait,
// parce qu'il en a la forme (typage STRUCTUREL, pas de mot-clé implements).
type EmailSender struct{ from string }

func (e EmailSender) Notify(msg string) error { return nil }

// L'affectation muette : si EmailSender cesse d'honorer le contrat,
// cette ligne ne compile plus — garantie gratuite, dès la compilation.
var _ Notifier = (*EmailSender)(nil)

// Une PETITE interface (io.Reader) rend la fonction universelle : elle
// marche avec une chaîne, un fichier, une connexion réseau… sans les connaître.
func firstLine(r io.Reader) (string, error) {
	sc := bufio.NewScanner(r)
	sc.Scan()
	return sc.Text(), sc.Err()
}

// « Accepter des interfaces (Save prend le contrat minimal)…
func Save(w io.Writer, data []byte) error {
	_, err := w.Write(data)
	return err
}

// … retourner des structs » : l'appelant garde TOUTE l'API du type concret.
func NewBuffer() *bytes.Buffer {
	return &bytes.Buffer{}
}

// LE piège du nil typé : une interface est une paire (type, valeur).
// Elle ne vaut nil que si LES DEUX sont nils.
type MyError struct{ Code int }

func (e *MyError) Error() string { return "erreur " + strconv.Itoa(e.Code) }

// Renvoie (*MyError)(nil) rangé dans une error : le TYPE est renseigné,
// donc l'interface n'est PAS nil — err != nil sera vrai !
func typedNil() error {
	var e *MyError = nil
	return e
}

// La forme correcte : en l'absence d'erreur, renvoyer le nil LITTÉRAL.
func literalNil() error {
	return nil
}

// Le type switch : aiguiller selon le type dynamique caché dans l'interface.
func describe(i any) string {
	switch v := i.(type) {
	case nil:
		return "nil"
	case int:
		return fmt.Sprintf("entier %d", v)
	case error: // on peut aiguiller vers UNE AUTRE interface
		return "erreur : " + v.Error()
	default:
		return fmt.Sprintf("type %T", v)
	}
}

// Capacité OPTIONNELLE : asserter vers une autre interface pour découvrir
// un talent supplémentaire — on ne ferme la source que si elle sait se fermer.
type closableReader struct {
	io.Reader
	closed bool
}

func (c *closableReader) Close() error { c.closed = true; return nil }

func process(r io.Reader) error {
	if c, ok := r.(io.Closer); ok { // ce Reader est-il AUSSI un Closer ?
		defer c.Close()
	}
	return nil
}

func main() {
	fmt.Println("=== Satisfaction implicite ===")
	var n Notifier = EmailSender{from: "no-reply@x.io"} // la forme suffit
	fmt.Println("EmailSender est un Notifier :", n != nil)

	fmt.Println("=== Petites interfaces universelles ===")
	line, _ := firstLine(strings.NewReader("bonjour\nmonde")) // une chaîne…
	fmt.Println("firstLine(chaîne) →", line)                  // …ou un fichier, une socket

	fmt.Println("=== Accepter des interfaces, retourner des structs ===")
	buf := NewBuffer()
	_ = Save(buf, []byte("données")) // Save écrit dans N'IMPORTE quel Writer
	fmt.Println("Save → Buffer →", buf.String())

	fmt.Println("=== Le piège du nil typé ===")
	fmt.Println("typedNil()  == nil →", typedNil() == nil, " ← LE PIÈGE : pointeur nil typé ≠ interface nil")
	fmt.Println("literalNil() == nil →", literalNil() == nil)

	fmt.Println("=== Assertions virgule-ok ===")
	var i any = "bonjour"
	s, ok := i.(string) // forme sûre : ok dit si l'assertion a réussi
	fmt.Println("i.(string) →", s, ok)
	nb, ok2 := i.(int) // mauvais type : zéro-value + false, PAS de panique
	fmt.Println("i.(int)    →", nb, ok2, " (pas de panique grâce à ok)")
	// Sans le ok — i.(int) — l'assertion PANIQUERAIT sur un mauvais type.

	fmt.Println("=== Type switch ===")
	for _, x := range []any{nil, 42, errors.New("échec"), 3.14} {
		fmt.Println("describe →", describe(x))
	}

	fmt.Println("=== Capacité optionnelle (io.Closer) ===")
	cr := &closableReader{Reader: strings.NewReader("x")}
	_ = process(cr) // ce Reader sait se fermer → Close est appelé
	fmt.Println("Close appelé sur le fermable :", cr.closed)
	fmt.Println("Reader nu accepté sans souci :", process(strings.NewReader("y")) == nil)
}
