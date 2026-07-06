🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 16.1 OWASP appliqué à Go (injection, XSS et `html/template`, validation d'entrées)

L'[OWASP Top 10](https://owasp.org/Top10/) est le document de sensibilisation de référence sur les risques applicatifs web — un panorama des dangers, pas un cadre complet. Sa révision **2025** (annoncée en novembre 2025, finalisée en janvier 2026) réordonne les catégories : **Injection** recule de la 3ᵉ à la **5ᵉ place (A05:2025)**, précisément parce que les langages et frameworks ont fait des requêtes paramétrées et de l'échappement automatique la norme. Go est un bon élève sur ce point : sa bibliothèque standard rend plusieurs défenses **actives par défaut**.

Mais « par défaut » ne veut pas dire « automatique quoi qu'on écrive ». Cette section couvre les trois familles où le code Go décide, ligne à ligne, s'il est vulnérable : **les injections** (A05, qui inclut aussi le XSS), l'**échappement XSS via `html/template`**, et la **validation d'entrées** — la première couche de la défense en profondeur posée dans le [README du module](README.md).

## Panorama : le Top 10 2025 et où il vit dans cette formation

La sécurité est transversale. Ce tableau situe chaque risque 2025 ; cette section traite la ligne **A05** (et la part « validation de cible » de A01) :

| Risque 2025 | Traité principalement dans |
|-------------|----------------------------|
| A01 · Broken Access Control (SSRF inclus) | [§ 5.6](../05-backend-http/06-authentification.md) · SSRF ici en [§ 3.5](#35-cas-particulier--ssrf-désormais-dans-a01) |
| A02 · Security Misconfiguration | [§ 16.3](03-durcissement-http.md) · [§ 9.2](../09-conteneurs-cloud/02-kubernetes.md) |
| A03 · Software Supply Chain Failures | [§ 15.3](../15-deploiement-devops/03-supply-chain.md) |
| A04 · Cryptographic Failures | [§ 16.2](02-cryptographie-tls.md) |
| **A05 · Injection (SQLi, XSS…)** | **cette section** |
| A06 · Insecure Design | [§ 10.2](../10-architecture-services/02-clean-architecture.md) · [§ 5.5](../05-backend-http/05-api-rest-complete.md) |
| A07 · Authentication Failures | [§ 5.6](../05-backend-http/06-authentification.md) |
| A08 · Software and Data Integrity Failures | [§ 15.1](../15-deploiement-devops/01-build-versioning.md) · [§ 15.3](../15-deploiement-devops/03-supply-chain.md) |
| A09 · Security Logging & Alerting Failures | [§ 12.3](../12-erreurs-debogage/03-slog.md) · [§ 12.4](../12-erreurs-debogage/04-observabilite.md) |
| A10 · Mishandling of Exceptional Conditions | [§ 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md) · [§ 12.1](../12-erreurs-debogage/01-strategies-erreurs.md) · [§ 16.3](03-durcissement-http.md) |

## 1. La famille des injections (A05:2025)

Toutes les injections ont la même cause racine : **des données non fiables mélangées à du code interprété** par un moteur (SQL, shell, système de fichiers, moteur de template, en-têtes HTTP). Le remède est lui aussi unique : **séparer le code de la donnée**. On ne concatène jamais une entrée utilisateur dans une commande ; on la passe comme un *paramètre* que le moteur traite comme une valeur inerte.

### 1.1 Injection SQL

C'est le cas d'école. Le danger n'est pas SQL, c'est la concaténation :

```go
// ❌ Vulnérable : la donnée est interprétée comme du SQL.
query := fmt.Sprintf("SELECT id, email FROM users WHERE name = '%s'", name)
rows, err := db.Query(query) // name = "x' OR '1'='1" contourne le filtre ; "'; DROP TABLE users; --" le détruit
```

La parade tient en un mot : **placeholders**. La donnée voyage hors de la chaîne SQL, jamais réanalysée :

```go
// ✅ Sûr : la valeur passe par un paramètre lié.
rows, err := db.QueryContext(ctx,
    "SELECT id, email FROM users WHERE name = $1", name)
if err != nil {
    return nil, fmt.Errorf("recherche utilisateur : %w", err)
}
defer rows.Close()
```

Le symbole du placeholder dépend du driver : `$1`, `$2`… pour PostgreSQL (`pgx`, `lib/pq`), `?` pour MySQL et SQLite. Le détail est traité en [§ 7.1](../07-acces-donnees/01-database-sql.md) ; à retenir ici : c'est le driver, pas vous, qui assemble la requête finale.

**Nuance importante** — les *identifiants* (noms de table, de colonne, sens de tri) ne peuvent **pas** être des paramètres. Si l'un d'eux vient du client, on le valide contre une **liste blanche** :

```go
// L'ordre de tri vient de l'URL : on le mappe vers un ensemble fixe.
var orderBy string
switch r.URL.Query().Get("sort") {
case "name":
    orderBy = "name"
case "created_at":
    orderBy = "created_at"
default:
    orderBy = "id"
}
query := "SELECT id, email FROM users ORDER BY " + orderBy // sûr : orderBy ∈ ensemble contrôlé
```

Pour éliminer le risque à la racine, générez vos requêtes typées avec **sqlc** (SQL vérifié à la compilation) plutôt que de les écrire à la main — comparé aux ORM en [§ 7.3](../07-acces-donnees/03-sqlc-vs-orm.md).

### 1.2 Injection de commandes

Ici Go part avec un **avantage** : `os/exec` n'invoque **pas** de shell. `exec.Command("prog", a, b)` transmet les arguments directement au programme, sans interprétation de `;`, `|`, `$()`, etc. On réintroduit la faille uniquement en rappelant un shell soi-même :

```go
// ❌ Le shell est de retour → injection de commande.
out, err := exec.Command("sh", "-c", "convert "+userFile+" out.png").Output()
// userFile = "x.png; rm -rf /" exécute deux commandes.
```

```go
// ✅ Pas de shell : userFile est un argument, pas du code.
out, err := exec.CommandContext(ctx, "convert", userFile, "out.png").Output()
if err != nil {
    return fmt.Errorf("conversion : %w", err)
}
```

Deux réflexes complémentaires : ne laissez **jamais** l'utilisateur choisir le *nom du programme* (liste blanche de binaires), et méfiez-vous de l'**injection d'arguments** — une valeur commençant par `-` peut être interprétée comme une option. Quand l'outil le permet, isolez les opérandes avec `--`.

### 1.3 Traversée de chemin (path traversal)

Ouvrir un fichier dont le nom vient du client est une injection : la « commande » est un chemin, et `../` en est la charge utile. Le piège classique est de croire que `filepath.Join` protège :

```go
// ❌ filepath.Join nettoie le chemin mais ne CONFINE pas.
p := filepath.Join("/srv/files", userPath) // userPath = "../../etc/passwd" → /etc/passwd
data, err := os.ReadFile(p)
```

Depuis **Go 1.24**, la stdlib offre une primitive dédiée : **`os.Root`** confine toutes les opérations sous un répertoire racine et **rejette toute sortie** — via `..` comme via un lien symbolique qui s'échapperait (voir le [billet officiel](https://go.dev/blog/osroot)) :

```go
// ✅ os.Root : accès fichiers confiné.
root, err := os.OpenRoot("/srv/files")
if err != nil {
    return nil, err
}
defer root.Close()

f, err := root.Open(userPath) // toute cible hors racine → erreur *PathError
if err != nil {
    return nil, fmt.Errorf("ouverture fichier : %w", err)
}
defer f.Close()
```

Pour un accès ponctuel, `os.OpenInRoot("/srv/files", userPath)` fait la même chose en une ligne. La couverture des opérations s'est élargie en Go 1.25. Ce sujet est ancré côté E/S en [§ 7.6](../07-acces-donnees/06-fichiers-io.md).

> **Confiner n'est pas autoriser.** `os.Root` empêche de *sortir* de la racine, mais ne dit rien sur les fichiers *à l'intérieur* : un utilisateur pourrait atteindre `user2/prive.txt` alors que seul `user1/` lui est permis. Le confinement est une couche ; il ne remplace pas le contrôle d'accès (A01). Des traversées de chemin ont mené à des RCE activement exploitées (p. ex. CVE-2024-3400) — d'où l'intérêt de superposer les deux.

### 1.4 Autres injections, même principe

- **Injection de template** — ne construisez **jamais** le *texte* d'un template à partir d'une entrée. La donnée utilisateur est une *donnée* de template, pas du *code* :

  ```go
  // ❌ template.New(...).Parse(userInput) exécute un template arbitraire.
  // ✅ La donnée passe par le contexte d'exécution :
  t := template.Must(template.New("greet").Parse("Bonjour {{.Name}}"))
  err := t.Execute(w, struct{ Name string }{Name: userInput})
  ```

- **Injection de logs** — la journalisation structurée [`log/slog`](../12-erreurs-debogage/03-slog.md) encode chaque attribut séparément : une valeur ne peut pas forger une fausse ligne de log. Un `log.Printf("user=" + userInput)` naïf, lui, laisse un `"\n[ADMIN] ..."` fabriquer une entrée.

  ```go
  slog.Info("connexion", "user", userInput, "ip", ip) // valeurs encodées, pas concaténées
  ```

- **Injection d'en-têtes / response splitting** — `net/http` valide les valeurs d'en-tête et refuse les caractères de contrôle (CR/LF), neutralisant le response splitting sans effort de votre part. `w.Header().Set("X-Trace", userValue)` ne peut pas fabriquer un en-tête supplémentaire.

- **Injection NoSQL** — même règle : utilisez les filtres typés du driver (p. ex. MongoDB en [§ 7.5](../07-acces-donnees/05-nosql-redis.md)), pas des documents de requête assemblés par concaténation de chaînes.

## 2. XSS et `html/template`

Le XSS consiste à injecter du HTML ou du JavaScript dans une page rendue à **un autre utilisateur**. Cette formation est recentrée sur les API et le cloud-native, sans SSR généralisé — mais **dès que vous produisez du HTML**, une seule règle s'applique : passez par `html/template`, jamais par `text/template` ni par de la concaténation de chaînes.

### 2.1 `text/template` vs `html/template` — la distinction cruciale

Les deux paquets partagent la **même API**. La différence est invisible dans le code mais décisive à l'exécution : `text/template` n'échappe **rien**, `html/template` échappe **automatiquement**.

```go
data := struct{ Comment string }{Comment: `<script>alert(1)</script>`}

// ❌ text/template : aucun échappement → le <script> est injecté tel quel (XSS).
texttmpl.Must(texttmpl.New("t").Parse(`<p>{{.Comment}}</p>`)).Execute(w, data)

// ✅ html/template : échappement automatique.
htmltmpl.Must(htmltmpl.New("t").Parse(`<p>{{.Comment}}</p>`)).Execute(w, data)
// rendu : <p>&lt;script&gt;alert(1)&lt;/script&gt;</p>
```

### 2.2 L'échappement est **contextuel**

`html/template` ne fait pas un bête remplacement d'entités : il **analyse où la valeur atterrit** et échappe en conséquence. La même donnée est traitée différemment selon qu'elle tombe dans le corps HTML, un attribut, une URL, un bloc `<script>` ou du CSS :

```go
const page = `
<a href="{{.URL}}">{{.Label}}</a>
<script>var user = {{.Name}};</script>`
//  {{.URL}}   → échappement d'URL dans l'attribut href
//  {{.Label}} → échappement HTML dans le corps du lien
//  {{.Name}}  → échappement JavaScript (chaîne JS valide) dans le <script>
```

C'est cette conscience du contexte qui rend l'auto-échappement de Go fiable là où un `strings.ReplaceAll` maison échouerait.

### 2.3 Les types « échappatoires » — à manier avec précaution

Parfois on veut injecter du HTML de confiance sans qu'il soit échappé. Les types `template.HTML`, `template.JS`, `template.URL`, `template.CSS`, `template.HTMLAttr` disent au moteur : « fais-moi confiance, n'échappe pas ». Ce sont des **portes de sortie de la sécurité** :

```go
var trusted template.HTML = "<b>Mention légale approuvée</b>" // OK : contenu que VOUS contrôlez
// template.HTML(userInput) // ⚠️ réintroduit le XSS — à ne JAMAIS faire sur une entrée
```

Règle simple : ces types ne s'appliquent qu'à des constantes ou du contenu de confiance, jamais à une donnée venue de l'extérieur.

### 2.4 Au-delà du template

Pour une **API JSON**, la surface XSS se déplace : le vrai risque est de refléter une entrée dans une *page d'erreur HTML* ou une *redirection*. Servez un `Content-Type: application/json` correct et ne renvoyez pas d'entrée utilisateur dans du HTML d'erreur. En défense en profondeur, une en-tête **Content-Security-Policy** limite la casse même en cas d'oubli — voir le durcissement HTTP en [§ 16.3](03-durcissement-http.md).

## 3. Validation d'entrées

La validation est la couche qui précède tout le reste. Trois principes portent tout le sujet : **valider à la frontière**, préférer une **liste blanche** à une liste noire, et *« parser plutôt que valider »* — transformer une entrée douteuse en un **type** qui ne peut pas représenter un état invalide.

### 3.1 Valider à la frontière (DTO)

On décode la requête dans une structure dédiée (*Data Transfer Object*), on la valide, puis on la traduit vers les types du domaine. La donnée non fiable ne circule jamais en profondeur sous forme de `map` ou de `string` brute — le patron est détaillé en [§ 5.3](../05-backend-http/03-json.md).

### 3.2 Borner la taille (déni de service)

Une entrée sans limite est un vecteur de DoS. Deux gestes de la stdlib suffisent : plafonner le corps de la requête et rejeter les champs inconnus.

```go
func handleCreate(w http.ResponseWriter, r *http.Request) {
    r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // plafond dur : 1 Mio

    var req createUserDTO
    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields() // un champ non prévu → erreur
    if err := dec.Decode(&req); err != nil {
        http.Error(w, "requête invalide", http.StatusBadRequest)
        return
    }
    if err := req.validate(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    // ... req est désormais digne de confiance
}
```

Complétez côté serveur par des *timeouts* (lecture, écriture, en-têtes) — voir [§ 16.3](03-durcissement-http.md).

### 3.3 « Parser plutôt que valider »

Plutôt que de trimballer une `string` et de la revérifier à chaque étage, on la **parse une fois** en un type qui garantit sa validité pour toute la suite. La stdlib fournit les briques : `strconv` pour les nombres, `net/url` pour les URL, `net/mail` pour les adresses e-mail, `time.Parse` pour les dates.

```go
type Email string

func NewEmail(s string) (Email, error) {
    addr, err := mail.ParseAddress(s)
    if err != nil {
        return "", fmt.Errorf("email invalide : %w", err)
    }
    return Email(addr.Address), nil
}
```

Une fois `NewEmail` franchi, tout le code en aval reçoit un `Email` — impossible d'y glisser une chaîne non validée. C'est de la sécurité portée par le système de types, cohérente avec le style Go idiomatique (annexe [B](../annexes/go-idiomatique/README.md)).

### 3.4 Liste blanche plutôt que liste noire

Bloquer les « mauvais » caractères est fragile : l'attaquant trouve toujours un encodage que vous n'aviez pas prévu. **N'accepter que le connu-bon** est robuste :

```go
var allowedRoles = map[string]bool{"reader": true, "editor": true, "admin": true}

func parseRole(s string) (string, error) {
    if !allowedRoles[s] {
        return "", fmt.Errorf("rôle non autorisé : %q", s)
    }
    return s, nil
}
```

Pour des règles nombreuses (tags de struct, contraintes croisées), des bibliothèques comme `go-playground/validator` existent — mais commencez par des validations explicites en stdlib, et n'ajoutez une dépendance que lorsque la répétition le justifie.

### 3.5 Cas particulier : SSRF (désormais dans A01)

Un service Go est souvent un **client HTTP** ([§ 8.1](../08-communication-services/01-consommer-api.md)). Si l'URL cible vient de l'utilisateur, il peut forcer votre serveur à appeler une ressource **interne** (boucle locale, réseau privé, endpoint de métadonnées cloud `169.254.169.254`) : c'est le SSRF, fondu dans **A01:2025**. La défense est une validation de la *destination* :

```go
func safeTarget(raw string) (*url.URL, error) {
    u, err := url.Parse(raw)
    if err != nil {
        return nil, err
    }
    if u.Scheme != "https" { // liste blanche de schémas
        return nil, errors.New("schéma non autorisé")
    }
    ips, err := net.LookupIP(u.Hostname())
    if err != nil {
        return nil, err
    }
    for _, ip := range ips {
        if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
            return nil, errors.New("cible interne interdite") // 127/8, 10/8, 169.254/16, …
        }
    }
    return u, nil
}
```

> Attention au **TOCTOU** (DNS rebinding) : l'IP résolue peut changer entre la vérification et l'appel. Une défense rigoureuse épingle l'IP validée dans le `DialContext` du client, ou revérifie au moment de la connexion — et complète cela par un contrôle d'accès réseau (egress) côté infrastructure.

## 4. Outillage : détecter les injections statiquement

L'analyse statique (SAST) attrape des *motifs* connus — concaténation dans une requête, `exec` avec une variable, crypto faible. Elle ne remplace pas la revue (elle ignore la logique métier), mais elle rattrape l'inattention. L'outil de référence en Go est **`gosec`**, généralement piloté via **golangci-lint** (voir [§ 13.5](../13-tests-qualite/05-linters.md)). Il est complété par **`govulncheck`** pour les vulnérabilités *connues* des dépendances ([§ 15.3](../15-deploiement-devops/03-supply-chain.md)).

- **GoLand** — de nombreuses inspections signalent déjà ces motifs à la volée. Pour `gosec`/`golangci-lint`, ajoutez une configuration d'exécution *Go Tool* (ou utilisez le plugin golangci-lint), et un *File Watcher* pour un retour à chaque sauvegarde.
- **VS Code** — l'extension officielle Go délègue au linter configuré : réglez `"go.lintTool": "golangci-lint"` et `"go.lintOnSave": "package"` dans les *settings*. `gosec` se pilote alors via la configuration golangci-lint du dépôt, ou en tant que *task* dédiée.

Dans les deux IDE, l'objectif est le même : rendre l'alerte **immédiate**, au moment où l'on écrit la ligne à risque, plutôt qu'en CI seulement.

## En résumé

- Toute injection se règle en **séparant le code de la donnée** : placeholders SQL, arguments `os/exec` sans shell, `os.Root` pour les chemins, données passées au *contexte* d'un template.
- Le XSS se prévient avec **`html/template`** (échappement **contextuel** automatique) — jamais `text/template` pour du HTML, et les types `template.HTML`/`JS`/… uniquement sur du contenu de confiance.
- La validation repose sur trois réflexes : **frontière** (DTO + taille bornée), **liste blanche**, et **« parser plutôt que valider »** vers des types sûrs.
- Go fournit beaucoup de défenses **par défaut** (pas de shell, en-têtes validés, `os.Root`, échappement HTML) ; votre travail est de ne pas les contourner — et `gosec`/`govulncheck` vous préviennent quand vous le faites.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [16.2 Cryptographie, TLS, secrets](02-cryptographie-tls.md)

⏭ [Cryptographie (`crypto/*`), TLS, gestion des secrets](/16-securite/02-cryptographie-tls.md)
