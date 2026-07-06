🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 13.6 Couverture de code ; génération de tests par IA

Ce module se clôt sur deux questions complémentaires. La **couverture** répond à « quelles lignes mes tests ont-ils exécutées ? » — un signal utile mais facile à mal lire. La **génération de tests par IA** répond à « un modèle peut-il écrire ces tests à ma place ? » — un accélérateur réel qui exige d'être relu, précisément à cause de ce que la couverture ne dit pas. Les deux se rejoignent sur un point : **exécuter une ligne n'est pas la tester**.

---

## Mesurer la couverture

La couverture est native. `go test -cover` affiche un pourcentage par package ; pour l'exploiter, on produit un profil et on l'inspecte :

```sh
go test -covermode=atomic -coverprofile=cover.out ./...
go tool cover -func=cover.out   # détail par fonction + total
go tool cover -html=cover.out   # rapport visuel : couvert (vert) / non couvert (rouge)
```

Le mode se choisit avec `-covermode` : `set` (par défaut : la ligne a-t-elle été exécutée ?), `count` (combien de fois) et `atomic` (comme `count` mais sûr en concurrence — **requis avec `-race`**). Le drapeau `-coverpkg=./...` étend la mesure au-delà du package testé : utile quand un test d'un package en exerce un autre.

### Couverture au-delà des tests unitaires

Depuis Go 1.20, on peut instrumenter un **binaire entier** plutôt que la seule suite `go test` : `go build -cover -o app` produit un exécutable instrumenté ; lancé avec la variable `GOCOVERDIR` pointant vers un dossier, il y écrit la couverture obtenue lors d'exécutions **réelles** du programme. On agrège et convertit ensuite avec `go tool covdata` (`percent`, `textfmt`, `merge`). C'est le pendant naturel des tests d'intégration ([§ 13.3](03-tests-integration.md)) : mesurer ce que des scénarios de bout en bout couvrent vraiment, et pas seulement les tests unitaires.

---

## Lire la couverture sans en faire une religion ⭐

La couverture mesure **quelles lignes se sont exécutées**, jamais si le comportement est **correct**. Une couverture de 100 % avec des assertions faibles ou absentes ne prouve rien. C'est l'illustration de la loi de Goodhart : dès que la couverture devient une cible, elle cesse d'être un bon indicateur — on se met à écrire des tests sans assertion pour atteindre un chiffre.

On l'utilise donc comme un **radar de zones aveugles** : repérer le rouge — chemins d'erreur non testés, cas limites oubliés — et décider ce qui mérite un test. Inutile de courir après les derniers pour-cent sur du câblage trivial, du code généré ou des branches défensives inatteignables. En CI, un **seuil de non-régression** (« ne pas descendre sous X ») ou, mieux, la **couverture du diff** (les lignes modifiées dans une *pull request*) est plus utile qu'un 100 % absolu imposé.

---

## Génération de tests par IA 🤖

Les assistants (Copilot, Cursor, Claude…) rédigent des tests Go en quelques secondes. Bien employé, c'est un vrai gain de temps ; le piège est précis, et c'est exactement la même zone aveugle que celle de la couverture.

### Ce que l'IA fait bien

Échafauder des tests table-driven, énumérer les cas évidents, générer des *fakes* ou des mocks à partir d'une interface, produire le passe-partout `httptest`, remplir des assertions répétitives, suggérer des cas limites qu'on aurait manqués.

### Ce qu'il faut surveiller

- **Des tests qui n'affirment rien d'utile**, ou qui affirment l'implémentation (tautologiques) : ils gonflent la couverture sans valeur. C'est le lien direct avec la « couverture spectacle » ci-dessus.
- **Le sur-ajustement à la sortie actuelle** : le modèle fixe `want` à ce que le code renvoie *aujourd'hui*, figeant un éventuel bug comme résultat attendu.
- **Du Go non idiomatique** : pas de table, erreurs ignorées (`_ =`), sur-*mocking*, helpers d'assertion maison, `interface{}` au lieu de `any` (→ `go fix`, [§ 13.5](05-linters.md)), vérifications d'erreur faibles (`err != nil` plutôt que `errors.Is`/`errors.As`, cf. [§ 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md)).
- **Des API hallucinées** ou de mauvaises versions : des fonctions plausibles mais inexistantes.
- **Les cas intéressants manqués** : dense sur le chemin nominal, mince sur le vide, le nil, le débordement, la frontière, le concurrent — précisément ce que débusquent le fuzzing ([§ 13.4](04-fuzzing-benchmarks.md)) et le mutation testing.

### Un flux de travail sain

On traite la sortie de l'IA comme un **premier jet à relire** : le développeur reste responsable de la justesse. Concrètement, on fournit au modèle la fonction **et son contrat** (sa doc, pas seulement son corps), on demande un test table-driven, puis on relit chaque cas — `want` reflète-t-il la **spécification**, ou seulement la sortie courante ? On *prompte* pour du Go idiomatique ([§ 17.1](../17-developpement-ia/01-prompting-go.md)) : réclamer explicitement des tables et des sous-tests `t.Run`, `errors.Is` pour les erreurs, pas de bibliothèque d'assertion sauf demande, `any` plutôt que `interface{}` ; fournir un test existant comme ancre de style aide beaucoup.

Sur la sortie, on applique les garde-fous du module : l'exécuter, la passer sous `-race`, la soumettre aux linters ([§ 13.5](05-linters.md)) et à `go fix` (le Go généré par IA tend vers d'anciens idiomes), et regarder le delta de couverture — mais en **vérifiant que les nouveaux tests affirment**, au lieu de seulement s'exécuter. Les agents intégrés à l'éditeur peuvent consommer les diagnostics de gopls comme garde-fous en direct ([§ 13.5](05-linters.md)).

### Le meilleur garde-fou : le mutation testing

La couverture dit qu'une ligne a été exécutée ; le **mutation testing** dit si vos tests **s'apercevraient qu'elle est fausse**. Il modifie le code (`>=` → `>`, `+` → `-`, suppression d'une instruction…), relance la suite, et rapporte les mutants **survivants** — un survivant signale qu'aucun test n'a détecté le changement. C'est l'antidote le plus tranchant à la « couverture spectacle » de l'IA :

```go
func EligibleAuRabais(qté int) bool { return qté >= 10 } // seuil : 10 inclus

// Test « couvrant » généré : 100 % de couverture, CI verte…
func TestEligibleAuRabais(t *testing.T) {
	if !EligibleAuRabais(15) {
		t.Error("15 devrait être éligible")
	}
	if EligibleAuRabais(3) {
		t.Error("3 ne devrait pas l'être")
	}
}
// …mais si >= 10 devenait > 10, ces deux cas passeraient encore : la frontière
// (qté == 10) n'est jamais épinglée. Le mutant « survivant » révèle ce trou.
```

Le score (MSI, *Mutation Score Indicator*) est le pourcentage de mutants tués ; les mutations d'opérateurs relationnels (frontières, égalités) visent les bugs les plus fréquents en production. Côté outillage, **gremlins** est l'outil autonome bien maintenu, simple d'emploi et adapté à des modules de taille modeste (des forks de `go-mutesting` ajoutent seuils MSI, *baseline* et mode git-diff pour la CI). Contrepartie : c'est lent (recompilation et ré-exécution par mutant) — on le **cible** (packages modifiés, logique clé) plutôt que de muter tout le dépôt à chaque passe.

### Assistants intégrés

GoLand propose **Generate Unit Tests** via son AI Assistant ; VS Code, la génération inline de Copilot et la commande `/tests`. Pratique, mais la sortie passe par la même relecture, les mêmes linters, et — idéalement — un contrôle par mutation plutôt qu'un simple chiffre de couverture. Ces assistants sont détaillés en [§ 17.4](../17-developpement-ia/04-assistants-ide.md), et les bonnes pratiques d'usage de l'IA sont consolidées en annexe C.

---

## Côté IDE : GoLand et VS Code

**Couverture.** GoLand offre **Run with Coverage** : coloration dans la gouttière, fenêtre dédiée et pourcentage par package. VS Code (extension Go) fournit **Go: Toggle Test Coverage**, qui met en surbrillance les lignes couvertes ou non (en lançant `go test -cover` en arrière-plan).

**Génération de tests par IA.** Sous GoLand, l'AI Assistant génère un test pour la fonction sous le curseur ; sous VS Code, Copilot le propose en ligne ou via `/tests`. Dans les deux cas, la règle est la même : relire, passer les linters ([§ 13.5](05-linters.md)), et préférer un contrôle par mutation à un pourcentage de couverture.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [14 — Performance et gestion de la mémoire](../14-performance/README.md)

⏭ [Performance et gestion de la mémoire](/14-performance/README.md)
