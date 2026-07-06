🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 7.5 NoSQL et cache : MongoDB, Redis

Au-delà du relationnel, deux stockages spécialisés reviennent constamment. **MongoDB** est une base **documentaire** : des documents BSON au schéma souple, adaptés aux données imbriquées ou évolutives. **Redis** est un magasin **clé-valeur en mémoire**, employé avant tout comme **cache**, mais aussi pour la publication/souscription, les files, les verrous ou la limitation de débit. Tous deux disposent de clients Go mûrs et attentifs au `context.Context`. Le principe reste : **le bon outil**, pas un remplacement systématique du relationnel.

## MongoDB : base documentaire

### Le driver officiel — état 2026 (v1 → v2)

Point de vigilance immédiat : **le driver v2 est désormais la référence** (`go.mongodb.org/mongo-driver/v2`), et le v1 a été formellement déprécié début 2026, ne recevant plus que des correctifs de sécurité. La migration est notable car le paquet `primitive` a été fusionné dans `bson` (par exemple `primitive.ObjectID` devient `bson.ObjectID`), `mongo.Connect` n'accepte plus de contexte — on le passe désormais à `Ping` — et `SessionContext` cède la place à un simple `context.Context`. Les extraits ci-dessous utilisent le v2.

### Se connecter et obtenir une collection

```go
import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func Open(ctx context.Context, uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri)) // v2 : pas de contexte ici
	if err != nil {
		return nil, fmt.Errorf("connexion : %w", err)
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil { // le contexte va sur Ping
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("ping : %w", err)
	}
	return client, nil
}
```

Comme `*sql.DB` ([§ 7.1](01-database-sql.md)), un `*mongo.Client` est **sûr en concurrence** et se partage sur toute la durée de l'application. On obtient ensuite une collection : `coll := client.Database("shop").Collection("products")`.

### BSON et mappage des structures

MongoDB stocke des documents BSON. Le driver expose `bson.D` (document ordonné, suite de paires clé/valeur), `bson.M` (document non ordonné, une map), `bson.A` (tableau) et `bson.E` (un élément) ; les structures se mappent par des tags `bson` :

```go
type Product struct {
	ID    bson.ObjectID `bson:"_id,omitempty"` // v2 : bson.ObjectID (ex-primitive.ObjectID)
	Name  string        `bson:"name"`
	Price float64       `bson:"price"`
	Tags  []string      `bson:"tags,omitempty"`
}
```

`bson.D` sert quand l'ordre compte (pipelines d'agrégation) ; `bson.M` convient aux filtres simples.

### Opérations courantes

L'absence de résultat se teste comme au [§ 7.1](01-database-sql.md), avec la sentinelle `mongo.ErrNoDocuments` :

```go
// Insertion
res, err := coll.InsertOne(ctx, Product{Name: "Pi", Price: 3.14})
id := res.InsertedID.(bson.ObjectID)

// Lecture d'un document
var p Product
err = coll.FindOne(ctx, bson.M{"name": "Pi"}).Decode(&p)
switch {
case errors.Is(err, mongo.ErrNoDocuments): // cas métier, pas une panne
	return ErrNotFound
case err != nil:
	return err
}

// Lecture de plusieurs documents : curseur, décodé d'un coup
cur, err := coll.Find(ctx, bson.M{"price": bson.M{"$lt": 10}})
if err != nil {
	return err
}
defer cur.Close(ctx)
var cheap []Product
if err := cur.All(ctx, &cheap); err != nil {
	return err
}

// Mise à jour
_, err = coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"price": 4.20}})
```

### Quand MongoDB

MongoDB s'impose quand le **modèle documentaire** correspond réellement au domaine : schéma souple ou évolutif, données naturellement imbriquées, agrégats dénormalisés que l'on lit d'un bloc. Ce n'est pas un défaut à opposer au relationnel par réflexe : pour des données fortement structurées et transactionnelles, une base relationnelle ([§ 7.1](01-database-sql.md)–7.4) reste souvent le meilleur choix.

## Redis : clé-valeur en mémoire — surtout un cache

### Le paysage 2026 : Redis, licence, Valkey

Le contexte a changé depuis 2024. Redis, historiquement sous licence BSD, est passé en mars 2024 à un double modèle SSPL/RSALv2 (non approuvé par l'OSI), qui restreint notamment l'offre de Redis en service managé. En réaction, la Linux Foundation a lancé Valkey, un fork sous licence BSD, soutenu par de grands fournisseurs cloud et compatible au niveau du protocole et des commandes. Puis Redis 8 a ajouté en 2025 l'AGPLv3, approuvée par l'OSI, un retour vers l'open source. Conséquence pratique pour le développeur Go : ces évolutions pèsent sur les **décisions de déploiement et de licence**, mais **le code reste identique** — les clients ci-dessous parlent aussi bien à Redis qu'à Valkey.

### Le client go-redis (et les options haute performance)

`github.com/redis/go-redis/v9` est le client Redis officiel et le plus répandu pour Go, visant le support des trois dernières versions de Redis.

```go
import "github.com/redis/go-redis/v9"

rdb := redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})
defer rdb.Close()

if err := rdb.Ping(ctx).Err(); err != nil {
	return err
}
```

Pour des besoins de performance élevée, **rueidis** (`github.com/redis/rueidis`) et son pendant **valkey-go** (`github.com/valkey-io/valkey-go`) offrent l'auto-pipelining et un cache côté client assisté par le serveur, avec un débit nettement supérieur à go-redis, au prix d'une API à base de constructeur de commandes. go-redis reste le choix par défaut pour sa simplicité.

### Opérations de base : `Set`/`Get`, TTL, `redis.Nil`

Une clé absente n'est pas une erreur : c'est la sentinelle `redis.Nil`, à traiter comme un défaut de cache.

```go
// Écrire avec expiration (10 min ; 0 = pas d'expiration)
if err := rdb.Set(ctx, "user:42:name", "Alice", 10*time.Minute).Err(); err != nil {
	return err
}

val, err := rdb.Get(ctx, "user:42:name").Result()
switch {
case errors.Is(err, redis.Nil): // clé absente ou expirée : cache miss
	// recalculer / recharger
case err != nil:
	return err
default:
	_ = val
}
```

### Le patron *cache-aside*

Le motif de cache le plus courant : on tente le cache, et en cas d'absence on charge depuis la source durable puis on peuple le cache avec un TTL. Les erreurs de cache ne sont **pas bloquantes** — on doit pouvoir dégrader vers la base.

```go
func GetProduct(ctx context.Context, rdb *redis.Client, load func(context.Context, int64) (Product, error), id int64) (Product, error) {
	key := fmt.Sprintf("product:%d", id)

	if data, err := rdb.Get(ctx, key).Bytes(); err == nil { // 1) cache
		var p Product
		if json.Unmarshal(data, &p) == nil {
			return p, nil // cache hit
		}
	}

	p, err := load(ctx, id) // 2) source durable (SQL, Mongo…)
	if err != nil {
		return Product{}, err
	}

	if data, err := json.Marshal(p); err == nil { // 3) réécrire avec TTL
		_ = rdb.Set(ctx, key, data, 5*time.Minute).Err()
	}
	return p, nil
}
```

Sur une écriture, on **invalide** la clé (`rdb.Del(ctx, key)`) pour éviter de servir une valeur périmée.

### Au-delà du cache (notions)

Redis dépasse le simple cache : publication/souscription (`Subscribe`/`Publish`), files d'attente (listes avec `LPUSH`/`BRPOP`), **verrous distribués** (`SET NX PX`), limitation de débit (`INCR` + expiration) ou stockage de sessions. Ces usages relèvent d'autres modules pour leur mise en œuvre complète.

### Pièges

- **Ce n'est pas la source de vérité.** Traitez le cache comme volatil ; les données durables vivent dans la base relationnelle ou documentaire.
- **TTL et invalidation.** Un TTL trop long sert des données périmées, trop court réduit le taux de succès ; sur écriture, invalidez.
- **Ruée sur le cache (*stampede*).** À l'expiration d'une clé très sollicitée, de nombreuses requêtes frappent la base simultanément : on atténue par un verrou, un recalcul anticipé ou des TTL avec gigue.

## Choisir : relationnel, documentaire, cache

Un repère simple. Le **relationnel** ([§ 7.1](01-database-sql.md)–7.4) est le défaut pour des données structurées et transactionnelles. **MongoDB** entre en jeu quand le modèle documentaire correspond vraiment au domaine. **Redis/Valkey** est une couche **éphémère** — cache, sessions, files, verrous — que l'on ajoute *à côté* d'un stockage durable, jamais *à la place*.

## Côté IDE : GoLand et VS Code

**GoLand** gère MongoDB et Redis comme sources de données dans sa fenêtre *Database* (moteur DataGrip) : on explore collections et documents, on parcourt les clés, on exécute requêtes et commandes sans quitter l'éditeur — pratique pour éprouver une requête avant de l'écrire en Go.

**VS Code** s'appuie sur des extensions dédiées : *MongoDB for VS Code* (exploration et *playground* de requêtes) et une extension Redis pour parcourir les clés et lancer des commandes.

Dans les deux cas, `mongosh` et `redis-cli` (ou `valkey-cli`) restent disponibles au terminal, les tests s'exécutent contre une instance jetable via les Testcontainers ([§ 13.3](../13-tests-qualite/03-tests-integration.md)), et le débogage passe par Delve.

## En résumé

- **MongoDB** (documentaire) : driver **v2** (`go.mongodb.org/mongo-driver/v2`), v1 déprécié ; attention aux changements (`mongo.Connect` sans contexte, `bson.ObjectID` ex-`primitive`) ; BSON (`bson.D`/`M`/`A`), `FindOne().Decode()` + `mongo.ErrNoDocuments`, curseurs (`cur.All`).
- **Redis/Valkey** (clé-valeur, surtout cache) : client **go-redis** (`v9`), `redis.Nil` pour la clé absente ; rueidis/valkey-go pour la haute performance. Depuis 2024, licence Redis mouvante (SSPL/RSAL puis AGPLv3 en Redis 8) et fork **Valkey** (BSD) compatible — **le code Go ne change pas**.
- Motif de cache : **cache-aside** avec TTL, erreurs de cache non bloquantes, invalidation sur écriture ; attention à la ruée sur le cache.
- Choix : relationnel par défaut ; MongoDB si le modèle documentaire s'impose ; Redis en couche éphémère **à côté** d'un stockage durable, jamais à sa place.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [7.6 — Fichiers et E/S (`io`, `bufio`, `os`, `embed`)](06-fichiers-io.md)

⏭ [Fichiers et E/S (`io`, `bufio`, `os`, `embed`)](/07-acces-donnees/06-fichiers-io.md)
