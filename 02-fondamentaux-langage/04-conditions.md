🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 2.4 Structures conditionnelles

Go garde un contrôle de flot minimal et lisible. Les conditions se déclinent en trois formes : le **`if`** (souvent doté d'une instruction d'initialisation), le **`switch`** — plus souple qu'en C —, et le **`type switch`** pour les valeurs d'interface. Les boucles font l'objet de la [section 2.5](05-boucles.md).

Deux conventions valent pour toutes ces structures : **pas de parenthèses** autour de la condition, mais des **accolades toujours obligatoires** (même pour une seule instruction), l'accolade ouvrante restant sur la même ligne.

## `if` et `if` avec initialisation

La forme habituelle enchaîne `if` / `else if` / `else` :

```go
if score >= 90 {
	grade = "A"
} else if score >= 80 {
	grade = "B"
} else {
	grade = "C"
}
```

Go **n'a pas d'opérateur ternaire** : le `if` fait le travail.

Sa particularité la plus utile est l'**instruction d'initialisation**, exécutée juste avant la condition, dont la variable est **limitée à la portée du `if`/`else`** :

```go
if n := len(s); n > maxLen {
	return fmt.Errorf("trop long : %d caractères", n)
}
// n n'existe plus ici
```

C'est l'idiome central de la gestion d'erreurs, où l'on teste immédiatement l'erreur renvoyée par un appel :

```go
if err := save(user); err != nil {
	return err // err n'est visible que dans ce bloc if/else
}
```

Ce style — confiner la variable au bloc qui l'utilise — garde le reste de la fonction propre. L'idiome complet est développé en [section 2.9](09-gestion-erreurs.md).

## `switch`

Le `switch` de Go est plus expressif que son équivalent en C, sur plusieurs points.

**Pas de `fallthrough` implicite** : chaque `case` s'arrête de lui-même. Le mot-clé `fallthrough` force le passage au cas suivant, mais son usage reste rare. Un `case` peut regrouper **plusieurs valeurs** :

```go
switch day {
case "samedi", "dimanche":
	fmt.Println("week-end")
case "vendredi":
	fmt.Println("presque !")
default:
	fmt.Println("au travail")
}
```

Comme le `if`, le `switch` accepte une **instruction d'initialisation** :

```go
switch os := runtime.GOOS; os {
case "linux":
	// …
case "darwin":
	// …
}
```

Enfin, un `switch` **sans expression** équivaut à une chaîne `if`/`else if` lisible : chaque `case` porte alors une condition booléenne.

```go
switch {
case score >= 90:
	grade = "A"
case score >= 80:
	grade = "B"
default:
	grade = "C"
}
```

Les `case` peuvent contenir des expressions quelconques (pas seulement des constantes), ce qui rend cette forme particulièrement souple.

## `type switch`

Le **`type switch`** aiguille selon le **type dynamique** d'une valeur d'interface. Dans chaque `case`, la variable liée prend le type concret correspondant :

```go
func describe(x any) string {
	switch v := x.(type) {
	case int:
		return fmt.Sprintf("entier : %d", v)
	case string:
		return fmt.Sprintf("chaîne de %d caractères", len(v))
	case bool:
		return fmt.Sprintf("booléen : %t", v)
	default:
		return fmt.Sprintf("type inconnu : %T", v)
	}
}
```

Ici `any` est un alias de `interface{}` (« n'importe quel type »). Dans le cas `int`, `v` est un `int` ; dans le cas `string`, un `string` ; et ainsi de suite.

La forme à un seul type — l'assertion `x.(T)` — ainsi que les interfaces elles-mêmes sont détaillées en [section 3.3](../03-types-interfaces/03-interfaces.md).

## En résumé

Trois structures suffisent au contrôle conditionnel de Go : le **`if`**, dont l'**instruction d'initialisation** confine proprement les variables (au cœur de l'idiome d'erreur), le **`switch`** — sans `fallthrough` implicite, avec cas multiples, initialisation et forme sans expression —, et le **`type switch`** pour discriminer le type dynamique d'une interface. Reste l'autre pan du contrôle de flot : les **boucles**, en [section suivante](05-boucles.md).

---

[🔝 Sommaire](../SOMMAIRE.md) · [⏭️ Section suivante : 2.5 Boucles et itérateurs](05-boucles.md)

⏭ [Boucles (`for` unique, `range`, itérateurs `range-over-func`)](/02-fondamentaux-langage/05-boucles.md)
