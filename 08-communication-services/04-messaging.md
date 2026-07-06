🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 8.4 Messaging : NATS, Kafka (notions)

Après la communication synchrone ([§ 8.1](01-consommer-api.md), [§ 8.2](02-grpc.md)) et la poussée temps réel ([§ 8.3](03-websockets-sse.md)), voici le mode **asynchrone et découplé**. Un **intermédiaire** (le *broker*) s'intercale entre producteur et consommateur : le producteur émet un message sans attendre, le broker le route et, souvent, le met en tampon. On y gagne du **découplage** (producteurs et consommateurs s'ignorent), de la **résilience** (le tampon absorbe les pics et les pannes momentanées), la **diffusion** (un message pour plusieurs consommateurs) et les **architectures événementielles**. Cette section, au niveau des notions, situe deux systèmes représentatifs : **NATS** (léger, rapide, Go-natif) et **Kafka** (journal d'événements durable).

## Pourquoi du messaging (et quand)

On passe à l'asynchrone quand **une réponse immédiate n'est pas nécessaire** et que l'on veut découpler les services : traitement en arrière-plan, diffusion d'événements (« commande créée »), lissage de charge, communication *un-vers-plusieurs*. Quatre motifs reviennent : **publication/souscription**, **file de travail** (répartir des tâches entre ouvriers), **flux d'événements** (un journal ordonné et rejouable) et **requête-réponse** asynchrone.

## Les garanties de livraison

Un point structurant, commun à tous les brokers :

- **au plus une fois** : « on émet et on oublie » — un message peut être perdu, jamais dupliqué ;
- **au moins une fois** : livré une fois **ou plus** — le plus courant, qui **impose des consommateurs idempotents** (rappel de l'idempotence du [§ 8.1](01-consommer-api.md)) ;
- **exactement une fois** : difficile et d'usage limité (déduplication, transactions) ; en pratique, on vise souvent « effectivement une fois » via l'idempotence.

## NATS : léger, rapide, Go-natif

Le serveur NATS est écrit en Go, et son client `github.com/nats-io/nats.go` est de première qualité. Deux couches se distinguent.

**NATS Core** offre la publication/souscription, la requête-réponse et les groupes de file, en *au plus une fois*, sans persistance — extrêmement rapide et léger.

```go
nc, err := nats.Connect(nats.DefaultURL) // nats://127.0.0.1:4222
if err != nil {
	return err
}
defer nc.Close()

nc.Subscribe("orders.created", func(m *nats.Msg) { // souscription
	log.Printf("reçu : %s", m.Data)
})
nc.Publish("orders.created", []byte(`{"id":42}`)) // publication

reply, err := nc.Request("users.get", []byte(`{"id":42}`), time.Second) // requête-réponse

nc.QueueSubscribe("tasks", "workers", handler) // groupe de file : un seul membre reçoit chaque message
```

Les *subjects* sont hiérarchiques et acceptent des jokers (`*` pour un niveau, `>` pour le reste). **JetStream** ajoute la couche de persistance : streaming, garanties *au moins une fois* et *exactement une fois*, rejeu historique et magasin clé/valeur, par-dessus le NATS Core en *au plus une fois*. Son API vit désormais dans le paquet `jetstream` (fondé sur le `context`, orienté consommateurs « pull »), qui remplace l'implémentation JetStream historique du paquet `nats`.

```go
js, err := jetstream.New(nc) // nouvelle API JetStream
stream, _ := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
	Name:     "ORDERS",
	Subjects: []string{"orders.>"},
	Storage:  jetstream.FileStorage, // persistance sur disque
})
cons, _ := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{Durable: "processor"})
cons.Consume(func(msg jetstream.Msg) {
	// ... traitement ...
	msg.Ack() // accuser réception APRÈS traitement (au moins une fois)
})
```

## Kafka : le journal d'événements durable

Kafka est un **journal distribué, ordonné et rejouable**. Les messages vont dans des **topics**, découpés en **partitions** (parallélisme, et ordre garanti *au sein* d'une partition) ; chaque consommateur suit sa position par un **offset**, et un **groupe de consommateurs** se répartit les partitions. Grâce à la **rétention** (les messages sont conservés un certain temps), on peut **rejouer** l'historique — d'où sa place de choix pour l'*event sourcing* et les pipelines de données.

Côté Go, plusieurs clients, avec un arbitrage familier autour de cgo ([§ 6.4](../06-cli-outillage/04-distribution.md), [§ 7.2](../07-acces-donnees/02-drivers.md)) :

- **`github.com/segmentio/kafka-go`** — pur Go, avec des API haut niveau (`Reader`/`Writer`) qui reprennent les conventions de la bibliothèque standard et gèrent le `context` ; un bon défaut quand on veut éviter cgo ;
- **`github.com/IBM/sarama`** (ex-Shopify) — de loin le plus utilisé historiquement, mais d'un abord difficile : API de bas niveau exposant le protocole Kafka, sans support des `context` ;
- **`github.com/twmb/franz-go`** — un client pur Go moderne, performant et complet ;
- **`github.com/confluentinc/confluent-kafka-go`** — officiel, mais enveloppe cgo autour de la bibliothèque C `librdkafka`, donc une dépendance C (cross-compilation et binaires statiques compliqués, [§ 6.4](../06-cli-outillage/04-distribution.md)).

```go
import "github.com/segmentio/kafka-go"

w := &kafka.Writer{Addr: kafka.TCP("localhost:9092"), Topic: "orders"} // producteur
_ = w.WriteMessages(ctx, kafka.Message{Key: []byte("42"), Value: []byte(`{"id":42}`)})

r := kafka.NewReader(kafka.ReaderConfig{ // consommateur, dans un groupe
	Brokers: []string{"localhost:9092"},
	Topic:   "orders",
	GroupID: "order-processors",
})
for {
	msg, err := r.ReadMessage(ctx) // avance l'offset après lecture
	if err != nil {
		break
	}
	// ... traiter msg.Value ...
}
```

Pour un contrôle fin (validation *au moins une fois*), on sépare la lecture de la validation de l'offset (`FetchMessage` puis `CommitMessages`) afin de ne valider qu'**après** traitement réussi.

## NATS ou Kafka ?

| Critère | NATS | Kafka |
|---------|------|-------|
| Nature | messagerie légère, temps réel | journal d'événements durable |
| Latence / débit | très faible latence | haut débit, forte rétention |
| Persistance | optionnelle (JetStream) | native (cœur du modèle) |
| Rejeu de l'historique | via JetStream | natif (offsets, rétention) |
| Requête-réponse | intégrée | non (hors du modèle) |
| Exploitation | légère (binaire ~15 Mo) | plus lourde |

En pratique : **NATS** pour une messagerie légère et à faible latence entre microservices, la requête-réponse et une exploitation simple (JetStream ajoutant la durabilité au besoin) ; **Kafka** pour un journal d'événements durable et **rejouable**, une forte rétention et des pipelines à grande échelle. Ce ne sont pas des concurrents stricts, mais des outils de profils différents.

## Bonnes pratiques (notions)

Quelques réflexes transverses : des **consommateurs idempotents** (déduplication par identifiant de message), **accuser réception / valider l'offset après** traitement réussi, prévoir une **file de rebut** (*dead-letter*) pour les messages toxiques, **versionner le schéma** des messages (protobuf ou JSON, dans l'esprit du contrat du [§ 8.2](02-grpc.md)), utiliser le `context` pour un arrêt propre des consommateurs, et propager le **contexte de traçage** dans les en-têtes des messages — l'observabilité à travers une frontière asynchrone étant plus délicate (OpenTelemetry, [§ 12.4](../12-erreurs-debogage/04-observabilite.md)).

## Côté IDE : GoLand et VS Code

Ce sujet relève surtout de l'infrastructure et de la ligne de commande. Pour sonder les brokers, on utilise le **CLI `nats`** (publier, souscrire, gérer streams et consommateurs) et, pour Kafka, **`kcat`** (ex-`kafkacat`) ou les outils console de Kafka. Les deux IDE n'offrent pas d'intégration profonde ; l'essentiel se joue au terminal et via les **Testcontainers**, qui démarrent un broker NATS ou Kafka jetable pour les tests ([§ 13.3](../13-tests-qualite/03-tests-integration.md)). Point de vigilance déjà rencontré : `confluent-kafka-go` étant en **cgo**, l'environnement de build doit fournir un compilateur C, ce qui complique la compilation statique et croisée ([§ 6.4](../06-cli-outillage/04-distribution.md)). Le débogage passe par Delve.

## En résumé

- Le messaging **découple** producteur et consommateur (asynchrone, tampon, diffusion, événementiel) ; motifs : pub/sub, file de travail, flux d'événements, requête-réponse.
- Garanties : *au plus une fois*, ***au moins une fois*** (le cas courant → **consommateurs idempotents**), *exactement une fois* (difficile).
- **NATS** (Go-natif) : Core (pub/sub, requête-réponse, groupes de file, *au plus une fois*) + **JetStream** (persistance, *au moins une fois*, nouvelle API `jetstream`).
- **Kafka** : journal durable (topics, partitions, offsets, groupes, rétention/rejeu) ; clients Go **segmentio/kafka-go** (pur Go, `context`) et **franz-go** (moderne), **sarama** (mûr mais bas niveau), **confluent** (cgo).
- Choix : NATS pour la messagerie légère/temps réel et la requête-réponse ; Kafka pour le journal d'événements durable et rejouable.
- Bonnes pratiques : idempotence, validation après traitement, *dead-letter*, versionnement des messages, traçage propagé ([§ 12.4](../12-erreurs-debogage/04-observabilite.md)).

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [9 — Conteneurs et déploiement cloud](../09-conteneurs-cloud/README.md)

⏭ [Conteneurs et déploiement cloud](/09-conteneurs-cloud/README.md)
