🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 2.6 Chaînes, runes, UTF-8

En Go, une **chaîne (`string`) est une suite immuable d'octets** — le plus souvent du texte encodé en UTF-8. La distinction à intérioriser est celle entre **octets, runes et caractères** : l'indexation et `len()` travaillent sur des octets, tandis que `range` travaille sur des runes. Trois packages standard complètent le traitement du texte : `strings` (manipulation), `strconv` (conversions vers et depuis d'autres types) et `fmt` (formatage).

## Les chaînes : des octets immuables

Une chaîne est une séquence d'octets, conventionnellement de l'UTF-8. Indexer une chaîne (`s[i]`) renvoie un **octet**, et `len(s)` compte des **octets** — pas des caractères.

```go
s := "héllo"        // chaîne interprétée (UTF-8)
chemin := `C:\temp` // chaîne brute (backticks) : pas d'échappement, \n littéral

fmt.Println(s[0])   // 104 — un OCTET (le « h »), pas un caractère
fmt.Println(len(s)) // 6 — nombre d'OCTETS (« é » en compte 2)
// s[0] = 'H'       // ← erreur : une chaîne est immuable
```

Les chaînes étant immuables, on n'en modifie pas le contenu en place : on en construit de nouvelles. Les littéraux « interprétés » (guillemets) gèrent les échappements (`\n`, `\t`…) ; les littéraux « bruts » (backticks) prennent le texte tel quel, sauts de ligne compris.

## Octets, runes et caractères

Une **rune** est un **point de code Unicode** (alias de `int32`). En UTF-8, un caractère occupe de **1 à 4 octets** — d'où l'écart entre nombre d'octets et nombre de runes.

```go
s := "héllo"
fmt.Println(len(s))                    // 6 octets
fmt.Println(utf8.RuneCountInString(s)) // 5 runes

for i, r := range s { // i = indice d'OCTET, r = rune décodée
	fmt.Printf("%d:%c ", i, r) // 0:h 1:é 3:l 4:l 5:o
}
```

Les indices d'octet « sautent » (0, 1, **3**…) parce que le « é » occupe deux octets. Pour manipuler le texte rune par rune, on convertit en slice de runes :

```go
runes := []rune(s)    // découpe la chaîne en runes
fmt.Println(runes[1]) // 233 — le point de code de « é »
```

Le package `unicode/utf8` fournit les outils de bas niveau (`RuneCountInString`, `DecodeRuneInString`, `ValidString`…).

**Deux pièges à connaître :**

- La conversion `string(unEntier)` interprète l'entier comme une **rune**, pas comme du texte. `string(65)` vaut `"A"`, et non `"65"` — pour obtenir `"65"`, on passe par `strconv` (voir plus bas). `go vet` signale d'ailleurs cette conversion suspecte.
- Une rune reste un **point de code**, pas un « caractère perçu ». Certains symboles (emojis avec modificateurs, par exemple) se composent de plusieurs runes ; ce niveau (les *grapheme clusters*) dépasse le cadre des types intégrés.

Les conversions courantes se résument à : `[]byte(s)` et `string(b)` entre chaîne et octets, `[]rune(s)` et `string(runes)` entre chaîne et runes.

## Le package `strings`

`strings` regroupe les manipulations usuelles :

```go
strings.Contains("gopher", "go")      // true
strings.ToUpper("go")                 // "GO"
strings.Split("a,b,c", ",")           // ["a" "b" "c"]
strings.Join([]string{"a", "b"}, "-") // "a-b"
strings.TrimSpace("  salut  ")        // "salut"
```

S'ajoutent `HasPrefix`, `HasSuffix`, `Index`, `ReplaceAll`, `Fields`, `Repeat`, etc. Pour **assembler beaucoup de fragments** (dans une boucle, notamment), on privilégie `strings.Builder`, qui évite les réallocations d'une concaténation répétée par `+` :

```go
var b strings.Builder
for i := range 3 {
	fmt.Fprintf(&b, "ligne %d\n", i)
}
resultat := b.String()
```

## Le package `strconv`

`strconv` convertit entre chaînes et autres types de base. Les fonctions d'analyse renvoient une **erreur** en cas d'échec :

```go
n, err := strconv.Atoi("42")           // 42, nil
f, err := strconv.ParseFloat("3.14", 64)
ok, err := strconv.ParseBool("true")

s := strconv.Itoa(42)                  // "42"
```

`Atoi`/`Itoa` traitent le cas courant des entiers ; `ParseInt`/`ParseFloat`/`ParseBool` et leurs pendants `FormatInt`/`FormatFloat`/`FormatBool` couvrent le reste, avec le contrôle de la base et de la précision.

## Le package `fmt`

`fmt` gère l'affichage et le formatage. Les fonctions les plus fréquentes sont `Println`, `Printf`, `Sprintf` (qui **construit** une chaîne) et `Errorf` (qui construit une erreur).

```go
fmt.Printf("%d %s %v\n", 42, "go", true) // 42 go true
fmt.Printf("%q\n", "salut")              // "salut" (entre guillemets)
fmt.Printf("%T\n", 3.14)                 // float64
msg := fmt.Sprintf("id=%d", 7)           // construit la chaîne "id=7"
```

Les principaux **verbes** :

| Verbe | Usage |
|-------|-------|
| `%v` | valeur, format par défaut |
| `%+v` | structure avec noms de champs |
| `%#v` | représentation Go de la valeur |
| `%T` | type de la valeur |
| `%d` | entier (décimal) |
| `%s` | chaîne |
| `%q` | chaîne entre guillemets, échappée |
| `%f` | flottant |
| `%t` | booléen |
| `%c` | caractère (à partir d'une rune) |
| `%x` | hexadécimal |

Deux liens vers la suite : `fmt.Errorf` sert à **envelopper** une erreur avec le verbe `%w` (voir [section 2.9](09-gestion-erreurs.md)), et un type qui implémente `String() string` (l'interface `fmt.Stringer`) contrôle sa propre représentation avec `%v` et `%s` (les interfaces sont détaillées en [section 3.3](../03-types-interfaces/03-interfaces.md)).

## En résumé

Une chaîne Go est une **suite immuable d'octets** en UTF-8 ; toute la subtilité tient à distinguer **octets** (indexation, `len`), **runes** (`range`, `[]rune`, `unicode/utf8`) et **caractères perçus**. Autour de ce socle, `strings` manipule, `strconv` convertit et `fmt` formate. Poursuivons avec les structures de données qui portent le plus d'idiomes — et de pièges — du langage : **slices et maps**, en [section suivante](07-slices-maps.md).

---

[🔝 Sommaire](../SOMMAIRE.md) · [⏭️ Section suivante : 2.7 Tableaux, slices et maps](07-slices-maps.md)

⏭ [Tableaux, slices (capacité, `append`, pièges) et maps](/02-fondamentaux-langage/07-slices-maps.md)
