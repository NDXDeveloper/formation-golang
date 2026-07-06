🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 7.6 Fichiers et E/S (`io`, `bufio`, `os`, `embed`) ⭐

Les fichiers sont, eux aussi, une source de données. Go y répond par une trousse d'entrées/sorties de la bibliothèque standard, structurée autour de **deux interfaces** — `io.Reader` et `io.Writer` — dont la petitesse fait toute la puissance ([§ 3.3](../03-types-interfaces/03-interfaces.md)) : fichiers, connexions réseau, tampons mémoire, corps de requêtes HTTP… tout les implémente, donc tout se compose. Autour d'elles : `os` pour les fichiers et le système, `bufio` pour le tamponnage et la lecture ligne à ligne, `io/fs` pour abstraire le système de fichiers, et `embed` pour **loger des fichiers dans le binaire**. C'est aussi ici que se placent deux apports récents : `os.Root` (Go 1.24) et un `io.ReadAll` nettement plus rapide (Go 1.26).

## Les interfaces `io` : `Reader` et `Writer`

Tout part de deux contrats minimalistes :

```go
type Reader interface {
	Read(p []byte) (n int, err error)
}
type Writer interface {
	Write(p []byte) (n int, err error)
}
```

La fonction phare est **`io.Copy`**, qui transfère une source vers une destination **en flux**, sans tout charger en mémoire :

```go
n, err := io.Copy(dst, src) // dst est un io.Writer, src un io.Reader
```

Parce que `*os.File`, `net.Conn`, `bytes.Buffer`, `strings.Reader` ou un corps de réponse HTTP satisfont tous ces interfaces, le même code fonctionne pour toutes ces sources. La fin d'un flux est signalée par la sentinelle `io.EOF`.

Deux fonctions à distinguer. **`io.ReadAll`** lit *tout* en mémoire et n'est justifié que pour un contenu **borné** :

```go
data, err := io.ReadAll(r) // à réserver aux contenus dont la taille est connue et petite
```

🆕 Bonne nouvelle côté performances : depuis Go 1.26, `io.ReadAll` alloue moins de mémoire intermédiaire et renvoie une tranche de taille minimale ; il est souvent environ deux fois plus rapide, pour environ deux fois moins de mémoire, avec un gain croissant sur les grandes entrées — et cela, sans modifier une ligne. Quelques assistants complètent l'ensemble : `io.LimitReader` (borner le nombre d'octets lus — utile pour plafonner un corps de requête), `io.MultiReader`/`io.MultiWriter` (concaténer/diffuser) et `io.TeeReader` (lire tout en recopiant).

## `os` : fichiers et système

### Ouvrir, lire, écrire

```go
f, err := os.Open("data.txt") // lecture seule ; f est un io.Reader
if err != nil {
	return err
}
defer f.Close()

out, err := os.Create("out.txt")                                               // écrit, tronque, crée
app, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644) // contrôle fin
```

Pour un fichier entier et de taille raisonnable, les raccourcis suffisent : `os.ReadFile(name)` et `os.WriteFile(name, data, perm)`.

### Fermer proprement — surtout en écriture

En lecture, un simple `defer f.Close()` suffit. **En écriture, l'erreur de `Close` compte** : des données peuvent n'être écrites (vidées) qu'à la fermeture. L'idiome utilise un retour nommé pour la remonter :

```go
func writeConfig(path string, data []byte) (err error) {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := f.Close(); err == nil {
			err = cerr // Close peut révéler une erreur de vidage
		}
	}()
	_, err = f.Write(data)
	return err
}
```

### Parcourir, tester, manipuler

```go
entries, err := os.ReadDir("mydir") // []os.DirEntry (léger, depuis Go 1.16)
for _, e := range entries {
	_ = e.Name()
	_ = e.IsDir()
}

os.MkdirAll("a/b/c", 0o755)
os.Rename("ancien", "nouveau")

if _, err := os.Stat("x"); errors.Is(err, fs.ErrNotExist) {
	// le fichier n'existe pas — erreur typée (§ 2.9)
}
```

`os.CreateTemp` et `os.MkdirTemp` créent fichiers et répertoires temporaires.

### 🆕 `os.Root` : accès fichiers confiné (Go 1.24)

Dès qu'un chemin provient d'une **source non fiable** (nom de fichier fourni par l'utilisateur, extraction d'archive), la traversée de répertoire (`../../etc/passwd`) est un risque de sécurité. Le type `os.Root`, obtenu via `os.OpenRoot`, confine toutes les opérations à un répertoire donné : ses méthodes n'autorisent aucun chemin pointant hors du répertoire, y compris via des liens symboliques qui en sortiraient. Elles reflètent la plupart des opérations du paquet `os` — par exemple `os.Root.Open`, `os.Root.Create`, `os.Root.Mkdir` et `os.Root.Stat`.

```go
root, err := os.OpenRoot("/srv/uploads")
if err != nil {
	return err
}
defer root.Close()

f, err := root.Open(nomFourniParUtilisateur) // un « ../ » ou un symlink d'évasion est refusé
```

C'est la parade idiomatique à la traversée de chemin, sujet relié à la sécurité applicative ([§ 16.1](../16-securite/01-owasp-go.md)).

## `bufio` : E/S tamponnées et lecture ligne à ligne

`bufio` réduit les appels système en tamponnant. Deux usages dominent. La **lecture ligne à ligne** d'un gros fichier, sans tout charger, via `bufio.Scanner` :

```go
sc := bufio.NewScanner(f)
for sc.Scan() {
	line := sc.Text()
	_ = line
}
if err := sc.Err(); err != nil { // ne pas oublier : distingue fin et erreur
	return err
}
```

Piège classique : la taille maximale d'un jeton vaut 64 Kio par défaut ; pour des lignes plus longues, agrandir le tampon avec `sc.Buffer(...)`. Côté **écriture tamponnée**, un `bufio.Writer` *doit* être vidé :

```go
w := bufio.NewWriter(out)
defer w.Flush() // impératif, sinon des données restent en tampon
// ... de nombreux petits Write
```

## `io/fs` : l'abstraction du système de fichiers

L'interface **`fs.FS`** (Go 1.16) découple le code du système de fichiers réel. Une fonction qui prend un `fs.FS` fonctionne aussi bien sur le disque, sur un `embed.FS` ou sur une archive — et devient **testable** en lui passant un système en mémoire.

```go
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
```

`os.DirFS("racine")` présente un répertoire du disque comme un `fs.FS` ; `fs.ReadFile`, `fs.Glob` et `fs.Sub` complètent la panoplie.

## `embed` : embarquer des fichiers dans le binaire

La directive **`//go:embed`** (Go 1.16) incorpore des fichiers dans l'exécutable **à la compilation**. Deux formes :

```go
import "embed"

//go:embed schema.sql
var schemaSQL string // un fichier → string (ou []byte)

//go:embed templates/*.html static/*
var assets embed.FS // un ou plusieurs fichiers → embed.FS (qui implémente fs.FS)
```

Comme `embed.FS` satisfait `fs.FS`, il se branche partout : gabarits (`template.ParseFS`), service de fichiers statiques (`http.FS`) et, on l'a vu, migrations embarquées ([§ 7.4](04-migrations.md)). C'est l'aboutissement du **binaire unique** ([§ 6.4](../06-cli-outillage/04-distribution.md)) : schéma, gabarits, ressources et configuration par défaut voyagent dans le seul fichier livré. Quelques règles : les chemins sont relatifs au dossier du fichier `.go`, on ne peut pas remonter avec `..` ni sortir du paquet, et les fichiers cachés (préfixe `.` ou `_`) exigent le préfixe `all:` (`//go:embed all:templates`).

## Streaming ou tout charger ? Le bon réflexe

Le défaut idiomatique est de **traiter les données en flux** — `io.Copy`, `bufio.Scanner`, lectures par blocs — car la mémoire reste **bornée** quelle que soit la taille de l'entrée. On ne réserve `os.ReadFile`/`io.ReadAll` qu'aux contenus dont on **sait** qu'ils sont petits (un fichier de configuration, une courte réponse). Pour un corps HTTP ou un téléversement, on borne (`io.LimitReader`) ou on diffuse. L'accélération de `ReadAll` en 1.26 est bienvenue, mais elle ne change pas ce principe de maîtrise de la mémoire.

## Côté IDE : GoLand et VS Code

Deux points concrets pour ce sujet.

Le **répertoire de travail** décide de la résolution des chemins relatifs (`os.Open("data.txt")`) : c'est la cause la plus fréquente d'un « fichier introuvable » lorsqu'on exécute depuis l'IDE plutôt que depuis le terminal. On le fixe dans le champ *Working directory* de **GoLand** et via `cwd` dans le `launch.json` de **VS Code** (comme au [§ 6.1](../06-cli-outillage/01-flag-args-env.md)).

La directive **`//go:embed`** est comprise des deux environnements : `gopls` (VS Code) comme GoLand résolvent les chemins embarqués, les complètent, et **signalent dès l'édition** un motif qui ne correspond à aucun fichier — avant même la compilation. Enfin, avec Delve, on inspecte une tranche d'octets ou une erreur de fichier ; attention toutefois : un `io.Reader` est consommé à la lecture et ne se réinspecte pas après un `ReadAll`.

## En résumé

- `io.Reader`/`io.Writer` sont la colonne vertébrale : tout les implémente, donc tout se compose ; `io.Copy` diffuse sans tout charger, `io.EOF` marque la fin.
- `os` : `Open`/`Create`/`OpenFile` (+ raccourcis `ReadFile`/`WriteFile`) ; en écriture, **ne pas ignorer l'erreur de `Close`** ; erreurs typées via `errors.Is(err, fs.ErrNotExist)`.
- 🆕 `os.Root`/`os.OpenRoot` (Go 1.24) confine les accès à un répertoire et refuse toute évasion (`..`, symlinks) — la parade à la traversée de chemin ([§ 16.1](../16-securite/01-owasp-go.md)).
- `bufio` : `Scanner` pour lire ligne à ligne (vérifier `Err()`, attention aux 64 Kio) ; `Writer` à **vider** (`Flush`).
- `io/fs` abstrait le système de fichiers (code testable) ; **`//go:embed`** loge fichiers, gabarits et migrations dans le binaire unique ([§ 6.4](../06-cli-outillage/04-distribution.md), [§ 7.4](04-migrations.md)).
- Réflexe mémoire : **diffuser** par défaut ; `ReadAll`/`ReadFile` seulement pour du contenu borné — 🆕 `io.ReadAll` est ~2× plus rapide en 1.26, mais le principe demeure.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [8 — Communication entre services](../08-communication-services/README.md)

⏭ [Communication entre services](/08-communication-services/README.md)
