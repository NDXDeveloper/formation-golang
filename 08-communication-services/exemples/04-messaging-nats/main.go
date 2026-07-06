/* ============================================================================
   Section 8.4 : Messaging — NATS, Kafka (notions)
   Description : NATS en action contre le container du README — Core :
                 publication/souscription (le message ARRIVE), requête-réponse
                 (un répondeur + nc.Request), groupe de file (5 messages,
                 2 workers : chaque message n'est reçu que par UN membre),
                 sujets à jokers (« logs.> ») ; puis JetStream (nouvelle API
                 du paquet jetstream) : stream persisté sur disque
                 (FileStorage), consommateur durable, msg.Ack() APRÈS
                 traitement — la garantie « au moins une fois ». (Kafka, le
                 journal durable, reste illustré dans le .md : validé lors
                 des tests du chapitre.)
   Fichier source : 04-messaging.md
   Prérequis : le container NATS du README (JetStream activé), puis : go run .
   ============================================================================ */

package main

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	ctx := context.Background()

	nc, err := nats.Connect(nats.DefaultURL) // nats://127.0.0.1:4222
	if err != nil {
		fmt.Fprintln(os.Stderr, "erreur :", err, "— lancez le container du README d'abord")
		os.Exit(1)
	}
	defer nc.Close()

	fmt.Println("=== Core : publication / souscription ===")
	recu := make(chan string, 1)
	_, _ = nc.Subscribe("orders.created", func(m *nats.Msg) { // souscription
		recu <- string(m.Data)
	})
	_ = nc.Publish("orders.created", []byte(`{"id":42}`)) // publication
	fmt.Println("message reçu :", <-recu)

	fmt.Println("=== Requête-réponse (intégrée à NATS) ===")
	_, _ = nc.Subscribe("users.get", func(m *nats.Msg) {
		_ = m.Respond([]byte(`{"name":"Alice"}`)) // le répondeur
	})
	reply, err := nc.Request("users.get", []byte(`{"id":42}`), time.Second)
	fmt.Println("réponse :", string(reply.Data), "· err =", err)

	fmt.Println("=== Groupe de file : un seul membre reçoit chaque message ===")
	var w1, w2 atomic.Int32
	_, _ = nc.QueueSubscribe("tasks", "workers", func(m *nats.Msg) { w1.Add(1) })
	_, _ = nc.QueueSubscribe("tasks", "workers", func(m *nats.Msg) { w2.Add(1) })
	_ = nc.Flush()
	for range 5 {
		_ = nc.Publish("tasks", []byte("tâche"))
	}
	time.Sleep(300 * time.Millisecond)
	fmt.Printf("5 tâches réparties : worker1=%d worker2=%d (total %d, zéro doublon)\n",
		w1.Load(), w2.Load(), w1.Load()+w2.Load())

	fmt.Println("=== Sujets hiérarchiques et jokers ===")
	wild := make(chan string, 1)
	_, _ = nc.Subscribe("capteurs.>", func(m *nats.Msg) { wild <- m.Subject }) // > : tout le reste
	_ = nc.Flush()
	_ = nc.Publish("capteurs.eu.temperature", []byte("x"))
	fmt.Println("« capteurs.> » a capté :", <-wild)

	fmt.Println("=== JetStream : persistance + « au moins une fois » ===")
	js, err := jetstream.New(nc) // la nouvelle API (paquet jetstream)
	if err != nil {
		panic(err)
	}
	stream, err := js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     "ORDERS",
		Subjects: []string{"orders.>"},
		Storage:  jetstream.FileStorage, // persistance sur disque
	})
	if err != nil {
		panic(err)
	}
	cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{Durable: "processor"})
	if err != nil {
		panic(err)
	}
	done := make(chan string, 2)
	cc, err := cons.Consume(func(msg jetstream.Msg) {
		done <- msg.Subject()
		_ = msg.Ack() // accuser réception APRÈS traitement (au moins une fois)
	})
	if err != nil {
		panic(err)
	}
	defer cc.Stop()

	_, _ = js.Publish(ctx, "orders.eu.created", []byte(`{"id":1}`))
	_, _ = js.Publish(ctx, "orders.us.created", []byte(`{"id":2}`))
	for range 2 {
		fmt.Println("consommé + Ack :", <-done)
	}
	info, _ := stream.Info(ctx)
	fmt.Println("messages persistés dans le stream (FileStorage) :", info.State.Msgs)
}
