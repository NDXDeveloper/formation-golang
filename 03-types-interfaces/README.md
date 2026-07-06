🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 3. Types, méthodes et interfaces

Les [fondamentaux du langage](../02-fondamentaux-langage/README.md) nous ont donné les briques : types de base, fonctions, slices, maps, pointeurs, gestion des erreurs. Ce module passe à l'étage au-dessus — **modéliser des données et des comportements**, c'est-à-dire concevoir ses propres types et les faire coopérer.

C'est ici que Go révèle sa personnalité. Le langage n'a **ni classes, ni héritage, ni hiérarchie de types**. À la place, il propose trois mécanismes volontairement minimalistes :

- des **méthodes** que l'on attache à n'importe quel type nommé — pas seulement aux structs ;
- la **composition** par embedding, à la place de l'héritage ;
- des **interfaces implicites** : un type satisfait une interface du seul fait qu'il en possède les méthodes, sans mot-clé `implements` ni déclaration explicite.

Cette combinaison produit un style de conception particulier — faiblement couplé, testable, lisible — qui distingue le code Go idiomatique du simple « code qui compile ». L'objectif du module est précisément de faire ce pas : passer de la syntaxe à la **conception**.

## 🎯 Objectifs du module

À l'issue de ce module, vous serez capable de :

- définir des **structs** (littéraux nommés, tags de champs) et leur attacher des **méthodes**, en choisissant le receveur — valeur ou pointeur — en connaissance de cause ;
- composer des types par **embedding**, en sachant pourquoi ce n'est **pas** de l'héritage (pas de dispatch virtuel) ;
- concevoir avec des **interfaces implicites** : petites interfaces, définition côté consommateur, « accepter des interfaces, retourner des structs » ;
- déjouer les pièges du module : **method sets**, **nil typé**, valeurs non adressables, sur-abstraction ;
- écrire des **génériques** avec contraintes (`comparable`, unions, `~`, `cmp.Ordered`) — et surtout savoir quand s'en passer ;
- organiser un projet : **packages**, **`internal/`**, layout minimal, **workspaces** (`go work`).

## 🧭 Le fil conducteur : le découplage par les interfaces

Une idée résume l'esprit du module. En Go, on définit généralement une interface **du côté qui la consomme**, pas du côté qui l'implémente. Le producteur d'un type n'a pas besoin de connaître à l'avance les interfaces qui existeront un jour : il expose des méthodes, et n'importe quel consommateur décrit le contrat minimal dont il a besoin.

```go
// Le consommateur définit le contrat minimal dont il a besoin…
type Notifier interface {
	Notify(msg string) error
}

// … et ce type le satisfait sans jamais mentionner Notifier.
type EmailSender struct{}

func (e EmailSender) Notify(msg string) error {
	// envoi de l'e-mail
	return nil
}
```

`EmailSender` ne « déclare » rien : il devient un `Notifier` parce qu'il en a la forme. C'est le typage structurel (*structural typing*) de Go, et c'est le socle de tout ce qui suit — des petites interfaces de la bibliothèque standard (`io.Reader`, `io.Writer`) jusqu'aux frontières d'architecture des modules ultérieurs.

De là découle un adage que l'on retrouvera tout au long de la formation : **« accepter des interfaces, retourner des structs »** — prendre le contrat le plus général possible en entrée, et renvoyer un type concret exploitable en sortie.

## 📋 Prérequis

- Le [module 2 — Fondamentaux du langage](../02-fondamentaux-langage/README.md), et en particulier la [sémantique valeur vs référence et les pointeurs](../02-fondamentaux-langage/08-pointeurs.md) : le choix entre receveur valeur et receveur pointeur ([§ 3.1](01-structs-methodes.md)) en découle directement.
- L'[idiome de gestion des erreurs](../02-fondamentaux-langage/09-gestion-erreurs.md) — `error` est lui-même une interface, un premier exemple concret de ce que ce module généralise.

## 🗺️ Plan du module

| Section | Sujet | En bref |
|---------|-------|---------|
| 3.1 | [Structs, méthodes, receveurs valeur vs pointeur](01-structs-methodes.md) | Définir des types, y attacher des méthodes, choisir le bon receveur. |
| 3.2 | [Composition plutôt qu'héritage (embedding)](02-composition-embedding.md) | Réutiliser et assembler du comportement sans hiérarchie de classes. |
| 3.3 ⭐ | [Interfaces implicites — le cœur du design Go](03-interfaces.md) | Satisfaction implicite, petites interfaces (`io.Reader`, `io.Writer`), découplage. |
| 3.4 | [Génériques (contraintes, `any`, `comparable`)](04-generiques.md) | Paramétrer par le type — et surtout, quand *ne pas* le faire. |
| 3.5 | [Organisation du code : packages, `internal/`, layout, workspaces](05-organisation-code.md) | Structurer un projet, isoler l'implémentation, gérer plusieurs modules avec `go work`. |

La section **3.3** est le pivot du module : les interfaces conditionnent la testabilité (mocks du [module 13](../13-tests-qualite/README.md)), la structure des services ([module 10](../10-architecture-services/README.md)) et une bonne partie de ce qui rend un code Go « idiomatique ». Les **génériques** ([§ 3.4](04-generiques.md)), disponibles depuis Go 1.18 et encore affinés jusqu'à Go 1.26, sont un outil puissant mais à dégainer avec parcimonie : bien souvent, une interface bien pensée reste la meilleure réponse.

## 💡 Une mise en garde utile

Si vous venez de Java, C# ou Python, la tentation est grande de reconstruire des hiérarchies d'héritage, des interfaces massives ou des abstractions génériques « au cas où ». Go pousse dans l'autre sens : **commencer concret, extraire l'abstraction seulement quand un besoin réel apparaît**. Les interfaces trop larges et les génériques prématurés comptent parmi les anti-patterns les plus fréquents en Go — on les retrouve recensés en [annexe B](../annexes/go-idiomatique/README.md). Gardez ce réflexe à l'esprit en parcourant les sections qui suivent.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [3.1 — Structs, méthodes, receveurs valeur vs pointeur](01-structs-methodes.md)

⏭ [Structs, méthodes, receveurs valeur vs pointeur](/03-types-interfaces/01-structs-methodes.md)
