🔝 Retour au [Sommaire](/SOMMAIRE.md)

# Annexe D — Raccourcis et astuces GoLand & VS Code

Cette annexe est le **compagnon du quotidien** : raccourcis clavier et astuces de productivité pour développer en Go, côté GoLand **et** côté VS Code. Pour *choisir* entre les deux (forces, limites, licence), voir la grille de décision en [§1.6.2](../../01-introduction-go/06.2-goland-vs-vscode.md).

> **Avant tout : un seul raccourci à connaître.** Les keymaps varient selon le système et se personnalisent entièrement — inutile de tout mémoriser. Apprenez la **commande de découverte**, d'où l'on invoque et retrouve tout le reste (avec son raccourci affiché) :
> - **GoLand — Find Action** : ⌘⇧A (macOS) · Ctrl+Shift+A (Windows/Linux)
> - **VS Code — Palette de commandes** : ⌘⇧P (macOS) · Ctrl+Shift+P (Windows/Linux)
>
> Les tableaux ci-dessous donnent les **keymaps par défaut**. Pour les voir ou les changer : GoLand → *Settings → Keymap* ; VS Code → *Keyboard Shortcuts* (⌘K ⌘S / Ctrl+K Ctrl+S). Chaque IDE fournit aussi une fiche imprimable (menu *Help → Keymap Reference* / *Keyboard Shortcuts Reference*).

Le formatage et l'organisation des imports **à l'enregistrement** sont supposés configurés (le *pourquoi* est en [annexe C](../bonnes-pratiques/README.md), le *où cliquer* est [en fin d'annexe](#réglages-recommandés)).

---

## GoLand

Basé sur IntelliJ. `gofmt`/`goimports`, Delve et les outils Go sont intégrés.

### Navigation

| Action | macOS | Windows/Linux |
|---|---|---|
| Rechercher partout (*Search Everywhere*) | ⇧⇧ (double Maj) | Shift Shift |
| Aller au fichier | ⌘⇧O | Ctrl+Shift+N |
| Aller au type | ⌘O | Ctrl+N |
| Aller au symbole | ⌘⌥O | Ctrl+Alt+Shift+N |
| Aller à la déclaration/définition | ⌘B ou ⌘+clic | Ctrl+B ou Ctrl+clic |
| Aller à l'implémentation | ⌘⌥B | Ctrl+Alt+B |
| Aller au test / au code testé | ⌘⇧T | Ctrl+Shift+T |
| Trouver les usages | ⌥F7 | Alt+F7 |
| Structure du fichier | ⌘F12 | Ctrl+F12 |
| Fichiers récents | ⌘E | Ctrl+E |
| Emplacements récents | ⌘⇧E | Ctrl+Shift+E |
| Reculer / avancer | ⌘[ / ⌘] | Ctrl+Alt+← / → |
| Aller à la ligne | ⌘L | Ctrl+G |

### Édition

| Action | macOS | Windows/Linux |
|---|---|---|
| Complétion | ⌃Espace | Ctrl+Space |
| **Actions rapides / intentions** | ⌥⏎ | Alt+Enter |
| Info paramètres | ⌘P | Ctrl+P |
| Doc rapide | F1 | Ctrl+Q |
| Étendre / réduire la sélection | ⌥↑ / ⌥↓ | Ctrl+W / Ctrl+Shift+W |
| Dupliquer la ligne | ⌘D | Ctrl+D |
| Supprimer la ligne | ⌘⌫ | Ctrl+Y |
| Déplacer la ligne | ⌥⇧↑ / ⌥⇧↓ | Alt+Shift+↑ / ↓ |
| Commenter la ligne | ⌘/ | Ctrl+/ |
| Reformater le code | ⌘⌥L | Ctrl+Alt+L |
| Optimiser les imports | ⌃⌥O | Ctrl+Alt+O |
| Curseur sur l'occurrence suivante | ⌃G | Alt+J |
| Sélectionner toutes les occurrences | ⌃⌘G | Ctrl+Alt+Shift+J |

Le raccourci le plus rentable est **⌥⏎ (Alt+Enter)** : il ouvre les *intentions* et *quick-fixes* contextuels — traiter une erreur, remplir une struct, créer une méthode manquante, ajouter un import, etc.

### Refactoring et génération

| Action | macOS | Windows/Linux |
|---|---|---|
| Renommer (tout le projet) | ⇧F6 | Shift+F6 |
| Extraire une variable | ⌘⌥V | Ctrl+Alt+V |
| Extraire une fonction | ⌘⌥M | Ctrl+Alt+M |
| Extraire une constante | ⌘⌥C | Ctrl+Alt+C |
| Modifier la signature | ⌘F6 | Ctrl+F6 |
| Intégrer (*inline*) | ⌘⌥N | Ctrl+Alt+N |
| Menu « Refactoriser » | ⌃T | Ctrl+Alt+Shift+T |
| Générer… (constructeur, test…) | ⌘N | Alt+Insert |
| Implémenter les méthodes (interface) | ⌃I | Ctrl+I |

### Exécution et débogage

| Action | macOS | Windows/Linux |
|---|---|---|
| Exécuter | ⌃R | Shift+F10 |
| Déboguer | ⌃D | Shift+F9 |
| Exécuter la config au curseur | ⌃⇧R | Ctrl+Shift+F10 |
| Point d'arrêt | ⌘F8 | Ctrl+F8 |
| Pas par-dessus / dans / sortir | F8 / F7 / ⇧F8 | F8 / F7 / Shift+F8 |
| Reprendre | ⌥⌘R | F9 |
| Évaluer l'expression | ⌥F8 | Alt+F8 |

### Git et outils

| Action | macOS | Windows/Linux |
|---|---|---|
| Commit | ⌘K | Ctrl+K |
| Push | ⌘⇧K | Ctrl+Shift+K |
| Update (pull) | ⌘T | Ctrl+T |
| Terminal | ⌥F12 | Alt+F12 |
| Trouver une action | ⌘⇧A | Ctrl+Shift+A |
| Réglages | ⌘, | Ctrl+Alt+S |
| Fichier *scratch* | ⌘⇧N | Ctrl+Alt+Shift+Insert |

### Fonctions maison à connaître

- **Intentions (⌥⏎)** : *Handle error* (génère le `if err != nil`), *Fill fields* (remplit une struct), *Implement interface*, *Add import*.
- **Live templates** : `forr` (boucle `range`), `err` (bloc de gestion d'erreur), `ff` (`fmt.Printf`), `meth`, `t__` (fonction de test)… et la **complétion postfixe** : `.if`, `.for`, `.nil`, `.err`, `.var`, `.print` après une expression.
- **Test avec couverture** : bouton ▶ → *Run with Coverage* (surlignage dans la gouttière).
- **Test avec `-race`** : *Run/Debug Configurations* → cocher **« Enable data race detection »**.
- **Client HTTP intégré** : créer un fichier `.http` et envoyer des requêtes sans quitter l'IDE — pratique pour tester une API REST (cf. [§5.5](../../05-backend-http/05-api-rest-complete.md)).
- **Outils bases de données** intégrés (console SQL, schéma), utiles avec [§7.1](../../07-acces-donnees/01-database-sql.md).
- **Profileur** : ouverture des profils `pprof` avec *flame graphs* (cf. [§14.1](../../14-performance/01-pprof.md)).
- **Signets** (*Bookmarks*) et **historique local** (récupération de versions non commitées).

---

## VS Code

Léger et extensible. La puissance Go vient de l'**extension Go officielle** + du serveur de langage **`gopls`** ; le débogage passe par **Delve**.

### Navigation

| Action | macOS | Windows/Linux |
|---|---|---|
| Palette de commandes | ⌘⇧P | Ctrl+Shift+P |
| Ouverture rapide (fichier) | ⌘P | Ctrl+P |
| Symbole dans le fichier | ⌘⇧O | Ctrl+Shift+O |
| Symbole dans le projet | ⌘T | Ctrl+T |
| Aller à la définition | F12 | F12 |
| Aperçu de la définition (*peek*) | ⌥F12 | Alt+F12 |
| Aller aux implémentations | ⌘F12 | Ctrl+F12 |
| Toutes les références | ⇧F12 | Shift+F12 |
| Aller à la ligne | ⌃G | Ctrl+G |
| Reculer / avancer | ⌃- / ⌃⇧- | Alt+← / Alt+→ |

### Édition

| Action | macOS | Windows/Linux |
|---|---|---|
| IntelliSense | ⌃Espace | Ctrl+Space |
| **Action de code / correction rapide** | ⌘. | Ctrl+. |
| Renommer le symbole | F2 | F2 |
| Formater le document | ⇧⌥F | Shift+Alt+F |
| Organiser les imports | ⇧⌥O | Shift+Alt+O |
| Commenter la ligne | ⌘/ | Ctrl+/ |
| Copier la ligne au-dessus/dessous | ⇧⌥↑ / ⇧⌥↓ | Shift+Alt+↑ / ↓ |
| Déplacer la ligne | ⌥↑ / ⌥↓ | Alt+↑ / ↓ |
| Supprimer la ligne | ⌘⇧K | Ctrl+Shift+K |
| Curseur sur l'occurrence suivante | ⌘D | Ctrl+D |
| Sélectionner toutes les occurrences | ⌘⇧L | Ctrl+Shift+L |
| Ajouter un curseur au-dessus/dessous | ⌥⌘↑ / ⌥⌘↓ | Ctrl+Alt+↑ / ↓ |

L'équivalent du ⌥⏎ de GoLand est **⌘. (Ctrl+.)** : le menu *Code Action* de `gopls` — remplir une struct, extraire, implémenter une interface, ajouter les imports manquants.

### Interface et terminal

| Action | macOS | Windows/Linux |
|---|---|---|
| Basculer la barre latérale | ⌘B | Ctrl+B |
| Terminal intégré | ⌃ + accent grave | Ctrl + accent grave |
| Panneau des problèmes | ⌘⇧M | Ctrl+Shift+M |
| Diviser l'éditeur | ⌘\ | Ctrl+\ |

### Exécution et débogage

| Action | macOS | Windows/Linux |
|---|---|---|
| Démarrer le débogage | F5 | F5 |
| Exécuter sans déboguer | ⌃F5 | Ctrl+F5 |
| Point d'arrêt | F9 | F9 |
| Pas par-dessus / dans / sortir | F10 / F11 / ⇧F11 | F10 / F11 / Shift+F11 |
| Arrêter | ⇧F5 | Shift+F5 |

### Commandes Go à connaître (via la palette)

Ces commandes n'ont pas de raccourci par défaut mais forment le cœur de la productivité Go ; elles s'invoquent depuis la palette (⌘⇧P).

- **Go: Test Function At Cursor** · **Go: Test Package** · **Go: Test File** — lancer précisément ce qu'on vise.
- **Go: Toggle Test File** — basculer entre `foo.go` et `foo_test.go`.
- **Go: Generate Unit Tests For Function / File / Package** — squelettes de tests (outil `gotests`).
- **Go: Add Tags To Struct Fields** — ajoute les tags `json:"…"` (outil `gomodifytags`).
- **Go: Fill Struct** — remplit une struct (outil `fillstruct`), aussi via ⌘.
- **Go: Toggle Test Coverage In Current Package** — surligne la couverture.
- **Go: Install/Update Tools** · **Go: Restart Language Server** — maintenance de l'outillage.

Astuces complémentaires : les **CodeLens** *run test | debug test* apparaissent au-dessus des fonctions de test (et *run | debug* au-dessus de `main`) ; les **inlay hints** (noms de paramètres, types) s'activent dans les réglages `gopls`.

---

## Astuces Go communes aux deux IDE

| Tâche | GoLand | VS Code |
|---|---|---|
| Implémenter une interface | ⌃I / Ctrl+I (ou ⌥⏎ sur le type) | ⌘. → *stub methods* (`gopls`) |
| Remplir une struct | ⌥⏎ → *Fill fields* | ⌘. → *Fill struct* (ou cmd *Go: Fill Struct*) |
| Basculer / générer le test | ⌘⇧T ; ⌘N → *Test…* | *Go: Toggle Test File* ; *Go: Generate Unit Tests…* |
| Ajouter les tags JSON | ⌥⏎ (intentions) | *Go: Add Tags To Struct Fields* |
| Extraire fonction / variable | ⌘⌥M / ⌘⌥V | Sélection → ⌘. → *Extract…* |
| Renommer partout | ⇧F6 | F2 |
| Un test / avec `-race` / couverture | ▶ gouttière ; config *Enable data race detection* ; *Run with Coverage* | CodeLens *run test* ; `"go.testFlags": ["-race"]` ; *Toggle Test Coverage* |
| Organiser format + imports à l'enregistrement | *Actions on Save* | `formatOnSave` + `source.organizeImports` |
| Déboguer avec Delve | Intégré (⌃D / config) | `launch.json` (*Debug Package/Test*) |
| Tester une API REST | Client HTTP intégré (`.http`) | Extension *REST Client* (`.http`) |

Le débogage avec Delve, des deux côtés, est traité en détail en [§12.2](../../12-erreurs-debogage/02-debogage-delve.md).

---

## Réglages recommandés

Le *pourquoi* (fmt, vet, lint, `-race`, `govulncheck`) est développé en [annexe C](../bonnes-pratiques/README.md) ; voici le *où*.

**VS Code** — `settings.json` :

```json
{
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": { "source.organizeImports": "explicit" },
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.testFlags": ["-race"],
  "gopls": {
    "ui.semanticTokens": true,
    "ui.completion.usePlaceholders": true
  }
}
```

**GoLand** — via *Settings* :

- **Format + imports à l'enregistrement** : *Tools → Actions on Save* → cocher *Reformat code* et *Optimize imports* (`goimports` intégré).
- **Linter** : *Go → Linters* (intégration `golangci-lint`), inspections activées par défaut.
- **`-race`** : *Run/Debug Configurations* → *Enable data race detection*.
- **Couverture** : bouton ▶ → *Run with Coverage*.

> **Pour ceux qui passent d'un IDE à l'autre** : VS Code propose l'extension *IntelliJ IDEA Keybindings*, et GoLand un *keymap* « VS Code ». On garde ses réflexes le temps de la transition.

---

## Pour aller plus loin

- **Choisir entre GoLand et VS Code** (forces, limites, licence) : [§1.6.2](../../01-introduction-go/06.2-goland-vs-vscode.md).
- **Installation et outils** (gopls, dlv, extensions) : [§1.4](../../01-introduction-go/04-installation-outils.md).
- **Débogage avec Delve**, en détail : [§12.2](../../12-erreurs-debogage/02-debogage-delve.md).
- **Linters et qualité** : [§13.5](../../13-tests-qualite/05-linters.md) · **Bonnes pratiques** (dont la config qualité) : [annexe C](../bonnes-pratiques/README.md).
- **Idiomes du langage** : [annexe B](../go-idiomatique/README.md).

---

🔝 [Retour au sommaire](../../SOMMAIRE.md) · ⏭️ [Annexe E — Layout de projet standard commenté](../layout-projet/README.md)


⏭ [Layout de projet standard commenté](/annexes/layout-projet/README.md)
