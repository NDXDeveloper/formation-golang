🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 17.3 Génération de tests, migration assistée, revue de code par IA

Le [README du module](README.md) l'annonçait : ces trois tâches sont le point fort de l'IA *précisément parce qu'elles sont vérifiables*. Un test s'exécute et réussit ou échoue ; une migration est contrôlée par le compilateur et par les tests ; une revue est recoupée par les linters et par la réalité. Dans les trois cas, l'outillage (et les tests existants) fait l'arbitre — ce qui rend la délégation ici bien plus sûre que déléguer la *conception*.

Une même discipline traverse les trois, en trois règles d'or que cette section décline :

- **tests** : vérifier le comportement *correct*, pas le comportement *actuel* ;
- **migration** : viser du Go *idiomatique*, pas une traduction *littérale* ;
- **revue** : traiter chaque remarque comme une *hypothèse à vérifier*, pas un verdict.

## 1. Génération de tests

### 1.1 Un bon terrain

Écrire des tests, c'est du travail répétitif (fastidieux) *et* vérifiable (ils s'exécutent). L'IA absorbe la corvée ; l'exécution est le contrôle. À une condition : que les tests testent vraiment quelque chose.

### 1.2 Demandez du table-driven

Le test idiomatique en Go est *table-driven* avec des sous-tests `t.Run` ([§ 13.1](../13-tests-qualite/01-tests-unitaires.md)). Sans consigne, l'IA écrit parfois une série de tests répétitifs ; demandez la forme tabulaire.

```go
func TestDiscount(t *testing.T) {
    tests := []struct {
        name    string
        price   int
        pct     int
        want    int
        wantErr bool
    }{
        {"remise simple", 100, 10, 90, false},
        {"remise nulle", 100, 0, 100, false},
        {"pourcentage invalide", 100, 150, 0, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Discount(tt.price, tt.pct)
            if (err != nil) != tt.wantErr {
                t.Fatalf("Discount() err = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("Discount() = %d, want %d", got, tt.want)
            }
        })
    }
}
```

L'IA excelle aussi à *énumérer les cas limites* — entrée vide, `nil`, valeurs aux bornes, chemins d'erreur. C'est un vrai atout, à condition de garder la main sur ce que chaque cas *affirme*.

### 1.3 Le piège n°1 : tester le comportement *actuel*, pas *correct*

Générer des tests à partir du code existant fige ce que le code fait — **bugs compris**. Les tests « passent » parce qu'ils affirment le comportement observé, pas le comportement attendu.

```go
// ❌ Test tautologique : il refait le calcul de l'implémentation,
//    donc il « passe » même si Total() est faux.
want := 0
for _, it := range items {
    want += it.Price
}
if Total(items) != want {
    t.Errorf("Total() = %d, want %d", Total(items), want)
}
```

La parade : relire chaque assertion en se demandant « est-ce ce qui *devrait* arriver ? » et non « est-ce ce que le code *fait* ? ». Un bon test **échoue si l'on casse volontairement le code** (état d'esprit du *mutation testing*). Un test qui ne peut pas échouer ne teste rien.

### 1.4 Couverture ≠ correction

Laissée libre, l'IA court après les 100 % de couverture avec des tests superficiels. La couverture ([§ 13.6](../13-tests-qualite/06-couverture-tests-ia.md)) dit ce qui est *exécuté*, pas ce qui est *vérifié*. C'est une mesure utile, jamais une preuve.

### 1.5 Au-delà : mocks, fuzzing, concurrence

- **Mocks par interfaces** ([§ 13.2](../13-tests-qualite/02-mocks-testify.md)) : l'IA les génère volontiers, mais attention au *sur-mockage*. Préférez de vraies dépendances quand c'est praticable (`httptest`, Testcontainers en [§ 13.3](../13-tests-qualite/03-tests-integration.md)).
- **Fuzzing natif** ([§ 13.4](../13-tests-qualite/04-fuzzing-benchmarks.md)) : l'IA aide à écrire des cibles de fuzz, qui explorent des entrées auxquelles on n'aurait pas pensé.
- **Tests concurrents** : `testing/synctest`, stable depuis Go 1.25 ([§ 4.6](../04-concurrence/06-tester-code-concurrent.md)), rend les tests de goroutines déterministes. L'IA, entraînée avant sa stabilisation, l'ignore souvent — orientez-la explicitement vers ce paquet.

## 2. Migration assistée

Traduire un service Python, Java ou Node vers Go ([§ 11.3](../11-interop-migration/03-migrer-vers-go.md)) est un bon terrain pour l'IA — la traduction est du *pattern-matching*, sa force — mais sous conditions, car la source impose sa forme.

### 2.1 Idiomatique, pas littéral

C'est ici que les travers de la [section 17.2](02-pieges-ia.md) se concentrent : demandez « traduis ce code » et vous obtenez du Python-en-Go ou du Java-en-Go (classes, exceptions transformées en `panic`, héritage plaqué sur l'embedding). Demandez plutôt un **équivalent Go idiomatique préservant le comportement**, en nommant les correspondances : exceptions → `(T, error)` ; classes → structs + méthodes ; héritage → composition ; `async`/Promise → goroutines/channels ou simplement du synchrone.

```python
# Python (source) : l'absence lève une exception.
def load_user(uid):
    row = db.query("SELECT id, name FROM users WHERE id = %s", uid)
    if row is None:
        raise NotFoundError(uid)
    return User(row)
```

```go
// Go idiomatique (cible) : l'exception devient une erreur retournée.
func LoadUser(ctx context.Context, uid int) (*User, error) {
    var u User
    row := db.QueryRowContext(ctx, "SELECT id, name FROM users WHERE id = $1", uid)
    if err := row.Scan(&u.ID, &u.Name); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, fmt.Errorf("utilisateur %d : %w", uid, ErrNotFound)
        }
        return nil, fmt.Errorf("chargement utilisateur %d : %w", uid, err)
    }
    return &u, nil
}
```

Migrer d'un langage dynamique vers Go apporte un bonus : le typage statique fait remonter des bugs que la source masquait. Encore faut-il que la traduction rende les types *explicites* — ce que le réflexe `any`-partout de l'IA (17.2) sabote.

### 2.2 Les tests sont le filet de la migration

La règle d'or : on migre un **comportement**, prouvé par des tests. En pratique, on caractérise d'abord le comportement de l'ancien système par des tests en boîte noire, on les porte en Go, puis on porte l'implémentation jusqu'à les faire passer. Le rappel de 17.2 vaut plus que jamais : **« ça compile » n'est pas « ça se comporte pareil ».** Les différences sémantiques sournoises — division entière, encodage des chaînes, sémantique des erreurs, modèle de concurrence — ne se voient qu'à l'exécution.

### 2.3 Par petits pas

Migrez module par module (approche *strangler*), jamais d'un bloc — la stratégie d'ensemble est en [§ 11.3](../11-interop-migration/03-migrer-vers-go.md) et [§ 10.1](../10-architecture-services/01-monolithe-vs-microservices.md). L'IA travaille mieux sur de petits périmètres vérifiables, comme vu en [§ 17.1](01-prompting-go.md).

## 3. Revue de code par IA

### 3.1 Un premier relecteur, pas le dernier

La revue par IA prend sa place dans une superposition :

1. **linters déterministes** (`go vet`, `staticcheck`, `gosec`) — la référence pour ce qu'ils vérifient ;
2. **revue par IA** — large, heuristique, elle *explique* et attrape des odeurs sémantiques que les linters manquent ;
3. **revue humaine** — conception, architecture, contexte métier, et responsabilité.

L'IA se situe entre le mécanique et l'humain ; elle ne remplace ni l'un ni l'autre.

### 3.2 Ce qu'elle repère bien

Erreurs ignorées, contexte d'erreur manquant, déréférencements `nil` potentiels, motifs non idiomatiques, cas limites oubliés, noms obscurs, simplifications possibles — et elle *dit pourquoi*, sur l'ensemble d'un diff, là où un linter se contente d'un code d'avertissement.

### 3.3 Ce qu'elle rate ou invente

Elle **hallucine des problèmes** (faux positifs affirmés avec aplomb), ignore les défauts d'architecture, n'a pas le contexte métier, ne peut pas *exécuter* le code (elle devine le comportement), et peut « corriger » vers des idiomes périmés (17.2). D'où la règle d'or de la revue : ses remarques sont des **hypothèses à vérifier**, pas des verdicts.

### 3.4 Ne pas déléguer le jugement

Celui qui *fusionne* est responsable, pas l'IA. Le mode d'échec classique est un humain pressé qui fusionne du code écrit par IA et relu par IA. La revue humaine garde la main sur ce qui engage : conception et sécurité.

## 4. Le point commun : vérifiable, donc délégable

Ces trois tâches partagent un trait : leur résultat est **contrôlable**. C'est ce qui en fait le terrain de prédilection de l'IA — et la thèse du module, rendue concrète :

- **tests** → `go test`, `go test -cover`, `-race`, `go test -fuzz` ;
- **migration** → le compilateur + la suite de tests portée + `-race` ;
- **revue** → `go vet` / `staticcheck` / `gosec` en garde-fou déterministe sous la passe heuristique de l'IA.

Côté IDE, des deux côtés du double environnement de la formation :

- **VS Code** — Copilot génère des tests depuis le chat (« des tests table-driven pour cette fonction ») et relit les *pull requests* en honorant `.github/copilot-instructions.md` ; l'extension Go lance `go test` et la couverture.
- **GoLand** — l'AI Assistant génère des tests pour une sélection et explique/relit un diff ; le lanceur de tests et la couverture sont intégrés.

La configuration détaillée de ces assistants est en [§ 17.4](04-assistants-ide.md).

La frontière, récurrente, reste la même : l'outillage vérifie la *mécanique* (les tests passent, ça compile, le linter est vert). Savoir si ce sont les *bons* tests, si le comportement porté est le *bon*, si la conception tient — cela demeure humain.

## En résumé

- Trois tâches où l'IA rend le plus de service **parce que le résultat se vérifie** : tests, migration, revue.
- **Tests** : exiger du table-driven ; le piège n°1 est de tester le comportement *actuel* (bugs inclus) plutôt que *correct* — un bon test échoue si l'on casse le code ; la couverture n'est pas une preuve.
- **Migration** : viser l'idiomatique (exceptions → `error`, classes → structs, héritage → composition), pas le littéral ; les tests portés sont le filet ; par petits pas.
- **Revue** : un premier relecteur entre les linters et l'humain — remarques = hypothèses à vérifier ; le jugement (conception, sécurité) et la responsabilité restent humains.
- Partout : l'outillage arbitre la mécanique ; la justesse d'intention reste au relecteur.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [17.4 Assistants intégrés : GoLand AI Assistant, Copilot dans VS Code](04-assistants-ide.md)

⏭ [Assistants intégrés : GoLand AI Assistant, Copilot dans VS Code](/17-developpement-ia/04-assistants-ide.md)
