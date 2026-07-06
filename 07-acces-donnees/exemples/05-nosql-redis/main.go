/* ============================================================================
   Section 7.5 : NoSQL et cache — MongoDB, Redis
   Description : Les deux volets de la section, exécutables contre les
                 containers du README. MongoDB avec le driver v2 (Connect
                 SANS contexte — le contexte va sur Ping —, bson.ObjectID,
                 FindOne/Decode + ErrNoDocuments, curseur cur.All, UpdateOne
                 $set) ; puis Redis/Valkey avec go-redis v9 : Set avec TTL,
                 redis.Nil = cache miss, et le patron CACHE-ASIDE complet
                 (miss → load → hit, invalidation sur écriture). L'exemple
                 pointe Valkey : le code est identique — c'est la claim de
                 compatibilité de la section, démontrée.
   Fichier source : 05-nosql-redis.md
   Prérequis : les containers MongoDB + Valkey du README, puis : go run .
   ============================================================================ */

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// Le document : tags bson pour le mappage (v2 : bson.ObjectID, ex-primitive).
type Product struct {
	ID    bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name  string        `bson:"name" json:"name"`
	Price float64       `bson:"price" json:"price"`
}

var ErrNotFound = errors.New("produit introuvable")

func main() {
	ctx := context.Background()

	fmt.Println("=== MongoDB (driver v2) ===")
	// v2 : Connect ne prend PLUS de contexte — il va sur Ping.
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://127.0.0.1:27018"))
	must(err)
	must(client.Ping(ctx, readpref.Primary()))
	defer client.Disconnect(ctx)

	coll := client.Database("shop").Collection("products")
	_ = coll.Drop(ctx) // démo relançable

	// Insertion : l'ID généré revient en bson.ObjectID.
	res, err := coll.InsertOne(ctx, Product{Name: "Pi", Price: 3.14})
	must(err)
	id := res.InsertedID.(bson.ObjectID)
	_, _ = coll.InsertOne(ctx, Product{Name: "Tau", Price: 6.28})
	fmt.Println("inséré, ObjectID :", id.Hex()[:8], "…")

	// FindOne + Decode ; l'absence est la sentinelle ErrNoDocuments.
	var p Product
	must(coll.FindOne(ctx, bson.M{"name": "Pi"}).Decode(&p))
	fmt.Println("FindOne :", p.Name, p.Price)
	err = coll.FindOne(ctx, bson.M{"name": "Absent"}).Decode(&p)
	fmt.Println("absent → ErrNoDocuments :", errors.Is(err, mongo.ErrNoDocuments))

	// Plusieurs documents : curseur décodé d'un coup (cur.All).
	cur, err := coll.Find(ctx, bson.M{"price": bson.M{"$lt": 10}})
	must(err)
	var cheap []Product
	must(cur.All(ctx, &cheap))
	fmt.Println("prix < 10 :", len(cheap), "produits")

	// Mise à jour ciblée.
	_, err = coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"price": 4.20}})
	must(err)
	_ = coll.FindOne(ctx, bson.M{"_id": id}).Decode(&p)
	fmt.Println("après $set :", p.Price)

	fmt.Println()
	fmt.Println("=== Redis/Valkey (go-redis v9) — ici : VALKEY, code identique ===")
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6380"})
	defer rdb.Close()
	must(rdb.Ping(ctx).Err())

	// Set avec TTL ; la clé ABSENTE n'est pas une erreur : redis.Nil.
	must(rdb.Set(ctx, "user:42:name", "Alice", 10*time.Minute).Err())
	val, _ := rdb.Get(ctx, "user:42:name").Result()
	fmt.Println("Get :", val)
	_, err = rdb.Get(ctx, "absente").Result()
	fmt.Println("clé absente → redis.Nil :", errors.Is(err, redis.Nil))

	// Le patron cache-aside : miss → load → hit (load appelé UNE fois).
	loads := 0
	load := func(ctx context.Context, id int64) (Product, error) {
		loads++ // en vrai : la requête SQL/Mongo coûteuse
		return Product{Name: "Chargé", Price: float64(id)}, nil
	}
	_ = rdb.Del(ctx, "product:7").Err()
	p1, _ := GetProduct(ctx, rdb, load, 7) // miss : va au load
	p2, _ := GetProduct(ctx, rdb, load, 7) // hit : servi du cache
	fmt.Printf("cache-aside : %s puis %s · load appelé %d fois (le cache a servi le 2e)\n",
		p1.Name, p2.Name, loads)

	// Invalidation sur écriture : Del, puis le prochain accès recharge.
	_ = rdb.Del(ctx, "product:7").Err()
	_, _ = GetProduct(ctx, rdb, load, 7)
	fmt.Println("après invalidation (Del) : rechargé, load =", loads)
}

// GetProduct : le cache-aside de la section — les erreurs de cache ne sont
// jamais bloquantes, on dégrade vers la source durable.
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

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "erreur :", err)
		os.Exit(1)
	}
}
