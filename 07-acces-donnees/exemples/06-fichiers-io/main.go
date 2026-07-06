/* ============================================================================
   Section 7.6 : Fichiers et E/S (io, bufio, os, embed)
   Description : La section entière, auto-démonstrative et SANS dépendance —
                 io.Copy et les assistants (LimitReader, MultiWriter,
                 TeeReader), os (WriteFile/ReadFile/ReadDir/Stat +
                 fs.ErrNotExist), le pattern « erreur de Close en écriture »,
                 os.Root (1.24) qui REFUSE l'évasion « ../ », bufio.Scanner
                 (ligne à ligne, le mur des 64 Kio et sa parade sc.Buffer),
                 bufio.Writer et le Flush oublié (données perdues, démontré),
                 io/fs (le même WalkDir sur le disque et sur embed.FS) et
                 //go:embed (string, FS, préfixe all: pour le fichier caché)
   Fichier source : 06-fichiers-io.md
   Lancer : go run .          (aucun service requis ; négatifs inclus)
   ============================================================================ */

package main

import (
	"bufio"
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed schema.sql
var schemaSQL string // un fichier → string

//go:embed all:templates
var assets embed.FS // dossier avec fichier caché → all: requis

var bad bool

func eq(label string, got, want any) {
	if fmt.Sprint(got) != fmt.Sprint(want) {
		fmt.Println("ÉCART", label, "got:", got, "want:", want)
		bad = true
	}
}

func main() {
	dir, _ := os.MkdirTemp("", "ch76-*")
	defer os.RemoveAll(dir)

	// ===== io.Copy : transfert en flux, io.EOF géré en interne
	src := strings.NewReader("contenu en flux")
	var dst bytes.Buffer
	n, err := io.Copy(&dst, src)
	eq("io.Copy octets", n, 15)
	eq("io.Copy err", err, nil)

	// ===== helpers : LimitReader, MultiWriter, TeeReader
	lr := io.LimitReader(strings.NewReader("0123456789"), 4)
	b, _ := io.ReadAll(lr)
	eq("LimitReader borne", string(b), "0123")
	var w1, w2 bytes.Buffer
	_, _ = io.MultiWriter(&w1, &w2).Write([]byte("x"))
	eq("MultiWriter diffuse", w1.String()+w2.String(), "xx")
	var copie bytes.Buffer
	tee := io.TeeReader(strings.NewReader("abc"), &copie)
	_, _ = io.ReadAll(tee)
	eq("TeeReader recopie", copie.String(), "abc")

	// ===== os : écrire, relire, ReadDir, Stat/ErrNotExist
	p := filepath.Join(dir, "data.txt")
	eq("WriteFile", os.WriteFile(p, []byte("ligne1\nligne2\n"), 0o644), nil)
	data, _ := os.ReadFile(p)
	eq("ReadFile", string(data), "ligne1\nligne2\n")
	entries, _ := os.ReadDir(dir)
	eq("ReadDir count", len(entries), 1)
	eq("DirEntry.Name", entries[0].Name(), "data.txt")
	_, err = os.Stat(filepath.Join(dir, "absent"))
	eq("Stat → fs.ErrNotExist (errors.Is)", errors.Is(err, fs.ErrNotExist), true)

	// ===== fermeture en écriture : le pattern retour nommé remonte l'erreur
	err = writeConfig(filepath.Join(dir, "conf"), []byte("ok"))
	eq("writeConfig (pattern Close)", err, nil)

	// ===== os.Root (1.24) : l'évasion est REFUSÉE
	sub := filepath.Join(dir, "uploads")
	_ = os.Mkdir(sub, 0o755)
	_ = os.WriteFile(filepath.Join(sub, "ok.txt"), []byte("autorisé"), 0o644)
	root, err := os.OpenRoot(sub)
	eq("OpenRoot", err, nil)
	defer root.Close()
	f, err := root.Open("ok.txt") // chemin sain : autorisé
	eq("Root.Open légitime", err, nil)
	if f != nil {
		f.Close()
	}
	_, err = root.Open("../data.txt") // traversée : REFUSÉE
	eq("Root refuse ../ (négatif)", err != nil, true)
	fmt.Println("   message Root :", err)

	// ===== bufio.Scanner : ligne à ligne, puis le mur des 64 Kio
	fl, _ := os.Open(p)
	sc := bufio.NewScanner(fl)
	var lignes []string
	for sc.Scan() {
		lignes = append(lignes, sc.Text())
	}
	eq("Scanner lignes", fmt.Sprint(lignes), "[ligne1 ligne2]")
	eq("Scanner.Err", sc.Err(), nil)
	fl.Close()

	longue := strings.Repeat("a", 70*1024) // > 64 Kio par défaut
	sc2 := bufio.NewScanner(strings.NewReader(longue))
	eq("ligne > 64 Kio : Scan false (négatif)", sc2.Scan(), false)
	eq("l'erreur est ErrTooLong", errors.Is(sc2.Err(), bufio.ErrTooLong), true)
	sc3 := bufio.NewScanner(strings.NewReader(longue))
	sc3.Buffer(make([]byte, 0, 128*1024), 128*1024) // la parade du .md
	eq("avec sc.Buffer : Scan true", sc3.Scan(), true)

	// ===== bufio.Writer : Flush oublié = données perdues (démonstration)
	var sink bytes.Buffer
	bw := bufio.NewWriter(&sink)
	_, _ = bw.WriteString("perdu sans Flush")
	eq("AVANT Flush : rien n'est écrit", sink.Len(), 0)
	_ = bw.Flush()
	eq("APRÈS Flush : écrit", sink.String(), "perdu sans Flush")

	// ===== io/fs : WalkDir sur DirFS et sur embed.FS (même code !)
	nDisk, _ := countFiles(os.DirFS(dir))
	eq("countFiles(DirFS)", nDisk >= 3, true)
	nEmb, _ := countFiles(assets)
	eq("countFiles(embed.FS) — caché inclus via all:", nEmb, 2)

	// ===== embed
	eq("embed → string", strings.TrimSpace(schemaSQL), "CREATE TABLE demo (id INTEGER);")
	hidden, err := assets.ReadFile("templates/.cache")
	eq("all: embarque le fichier caché", err == nil && string(hidden) == "caché", true)

	if bad {
		os.Exit(1)
	}
	fmt.Println("✔ 7.6 : io/bufio/os/fs/embed — tous les comportements (et négatifs) conformes")
}

func writeConfig(path string, data []byte) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}()
	_, err = f.Write(data)
	return err
}

func countFiles(fsys fs.FS) (int, error) {
	var n int
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			n++
		}
		return nil
	})
	return n, err
}
