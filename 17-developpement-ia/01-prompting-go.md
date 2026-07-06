🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 17.1 Copilot, Claude, ChatGPT : prompting efficace pour du Go idiomatique

Le [README du module](README.md) l'a posé : laissée à elle-même, l'IA produit du Go *plausible* mais rarement *idiomatique*. Prompter efficacement, c'est fournir le contrepoids — orienter le modèle vers les conventions de Go plutôt que vers un plus petit dénominateur commun valable dans n'importe quel langage. Cette section est la mécanique de cette orientation, largement indépendante de l'outil, avec les quelques endroits où Copilot, Claude et ChatGPT diffèrent.

Trois leviers, du plus ponctuel au plus durable : le **prompt immédiat**, les **instructions persistantes** du projet, et la façon dont vous **alimentez le contexte**.

## Pourquoi l'IA dérive

Un modèle est entraîné sur tous les langages à la fois. Sa réponse « moyenne » importe donc des réflexes venus de Java ou de Python : getters/setters, hiérarchies d'abstraction, exceptions. En Go, cette moyenne tombe à côté — le langage valorise l'inverse (erreurs explicites, composition, petites interfaces). Votre prompt existe pour tirer le modèle hors de cette moyenne. Les travers eux-mêmes sont catalogués en [§ 17.2](02-pieges-ia.md) ; ici, on cherche à les *prévenir*.

## 1. Le prompt immédiat : cadrer explicitement

### 1.1 Donnez le contexte et les contraintes

Une demande vague produit du code générique. Comparez :

> ❌ « Écris une fonction qui lit un fichier de config. »

Le modèle comble les blancs à sa guise — souvent un `panic`, une erreur avalée, ou une dépendance de configuration inutile.

> ✅ « Écris une fonction Go idiomatique qui lit un fichier de config JSON dans une struct. Signature `func LoadConfig(path string) (*Config, error)`. stdlib uniquement (`os`, `encoding/json`). Erreurs explicites, wrappées avec `%w` et un contexte utile ; jamais de `panic`. Aucun état global. »

De la seconde, on obtient du Go qu'on n'a pas à réécrire :

```go
type Config struct {
    Addr    string `json:"addr"`
    Timeout int    `json:"timeout"`
}

func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("lecture config %q : %w", path, err)
    }
    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parsing config %q : %w", path, err)
    }
    return &cfg, nil
}
```

### 1.2 Nommez les idiomes attendus

Le modèle *connaît* les idiomes de Go ; il ne les applique simplement pas par défaut. Les nommer suffit souvent à changer la sortie :

- « retourne une `error`, ne panique pas » ([§ 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md)) ;
- « wrappe les erreurs avec `%w` » ;
- « définis une petite interface, côté consommateur » ([§ 3.3](../03-types-interfaces/03-interfaces.md)) ;
- « pas de getters/setters » ;
- « stdlib avant toute bibliothèque tierce ».

### 1.3 Montrez un exemple (few-shot)

Le levier le plus fiable : coller un extrait de votre code existant et demander « dans le même style ». Le modèle recopie la structure, la gestion d'erreurs, le nommage. Un court échantillon idiomatique pèse plus qu'un paragraphe de consignes abstraites.

### 1.4 Demandez le pourquoi et les alternatives

« Y a-t-il plus idiomatique ? », « explique les compromis » : ces questions font remonter les choix discutables — une interface prématurée, une abstraction de trop, un `any` de facilité — que vous pouvez alors retirer. Utile en particulier autour des génériques ([§ 3.4](../03-types-interfaces/04-generiques.md)), que l'IA a tendance à sur-employer.

### 1.5 Découpez

Une tâche petite et bien définie produit du meilleur Go qu'un « construis-moi tout le service », qui dérive vers du générique tentaculaire. Découper, c'est aussi rendre chaque sortie vérifiable.

## 2. Les instructions persistantes : régler une fois pour toutes

Répéter les mêmes idiomes à chaque prompt est fastidieux. On les encode une fois dans un fichier d'instructions du projet, lu automatiquement à chaque requête.

### 2.1 `AGENTS.md` et ses variantes

Une convention transverse s'est imposée en 2026 : **`AGENTS.md`**, à la racine du dépôt, lu par plusieurs assistants (Copilot, Claude Code, Cursor, Gemini…). Chaque outil accepte aussi son fichier propre — `.github/copilot-instructions.md` (Copilot), `CLAUDE.md` (Claude Code), `GEMINI.md` — mais le contenu, lui, est le même : vos règles Go. Le standard transverse évite de dupliquer ces règles par outil.

### 2.2 Écrire de bonnes règles Go

Le principe cardinal, recommandé par la documentation des outils eux-mêmes : **ne gaspillez pas d'instructions sur ce que `gofmt`, `go vet` et `staticcheck` imposent déjà.** Le formatage, l'ordre des imports, les fautes que `vet` détecte : la chaîne s'en charge. Réservez vos règles aux idiomes *non mécanisables* — et justifiez-les, car le « pourquoi » aide le modèle dans les cas limites :

```markdown
# Instructions Go pour les assistants

## Idiomes (non couverts par l'outillage)
- Erreurs explicites : retourner `error`, jamais `panic` pour le contrôle de flux.
  Wrapper avec `%w` et un contexte : `fmt.Errorf("lecture %q : %w", nom, err)`.
- Composition, pas héritage : embedding et petites interfaces définies côté
  consommateur — pas une grande interface par type.
- stdlib d'abord : justifier toute dépendance ; pas de framework HTTP ni d'ORM
  sans raison explicite.
- Pas de getters/setters à la Java ; exporter les champs quand c'est légitime.
- Concurrence : `context.Context` en premier paramètre ; pas de goroutine dont
  personne n'attend la fin.

## Tests
- Table-driven avec sous-tests `t.Run` ; pas de framework d'assertion sauf demande.

## Vérification
- Le code doit passer `gofmt`, `go vet` et `staticcheck`.
```

Ces règles prolongent l'[annexe C](../annexes/bonnes-pratiques/README.md) (bonnes pratiques avec l'IA) et, pour le fond, l'[annexe B](../annexes/go-idiomatique/README.md). Un mot de prudence : gardez le fichier court et chaque règle autonome — une liste trop longue se dilue.

## 3. Complétion inline vs chat/agent : deux façons de piloter

Distinction souvent ignorée : les fichiers d'instructions ci-dessus s'appliquent au **chat et à l'agent**, mais **pas à la complétion inline** (le texte fantôme affiché pendant la frappe). Celle-ci se pilote autrement — par le **contexte** :

- ouvrez les fichiers pertinents (le modèle lit les onglets ouverts) ;
- écrivez d'abord une **signature précise et un commentaire de doc** : ils *sont* votre prompt pour la complétion.

```go
// LoadConfig lit un fichier JSON et le désérialise dans *Config.
// Retourne une erreur wrappée si le fichier est absent ou mal formé.
func LoadConfig(path string) (*Config, error) {
    // la complétion suit la signature et ce commentaire
}
```

Côté outillage, le pilotage vaut des deux côtés du double environnement de cette formation :

- **VS Code** — le plugin Copilot lit `.github/copilot-instructions.md` et `AGENTS.md` pour le chat et l'agent ; pour la complétion, tout se joue dans le contexte ouvert et la signature.
- **GoLand** — le plugin Copilot y honore les mêmes fichiers ; l'AI Assistant natif dispose de son propre mécanisme d'instructions de projet, et la complétion s'y pilote de la même façon (signature + doc).

La configuration détaillée de ces assistants est le sujet de [§ 17.4](04-assistants-ide.md) ; on s'en tient ici au *pilotage*.

## 4. Le prompt ne suffit pas

Même excellemment prompté, le résultat reste un **brouillon**. C'est la thèse du module : l'IA propose, l'outillage Go dispose. Un bon prompt améliore la *première passe* ; ce sont `gofmt`, `go vet`, `staticcheck`, les tests et le détecteur de races qui *garantissent* la correction ([§ 13.5](../13-tests-qualite/05-linters.md)). Prompter mieux réduit le nombre d'allers-retours — cela ne dispense jamais de vérifier, ce que détaillent les [pièges](02-pieges-ia.md) et la [délégation vérifiable](03-tests-migration-ia.md) des sections suivantes.

## En résumé

- L'IA dérive vers un Go générique ; le prompt est le contrepoids qui la ramène aux idiomes.
- **Prompt immédiat** : contexte et contraintes explicites, idiomes nommés, exemple few-shot, demande du « pourquoi », tâches découpées.
- **Instructions persistantes** : encodez vos règles Go une fois (`AGENTS.md`, ou `.github/copilot-instructions.md` / `CLAUDE.md`), en ciblant les idiomes *non couverts* par `gofmt` / `vet` / `staticcheck`, et en justifiant chaque règle.
- **Complétion inline ≠ chat** : elle se pilote par le contexte ouvert et une signature accompagnée d'un commentaire de doc soigné.
- Le meilleur prompt ne donne qu'un brouillon ; l'outillage Go reste le juge.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [17.2 Pièges de l'IA en Go](02-pieges-ia.md)

⏭ [Pièges de l'IA en Go (code non idiomatique, erreurs ignorées, sur-abstraction)](/17-developpement-ia/02-pieges-ia.md)
