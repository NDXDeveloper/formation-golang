/* ============================================================================
   Section 8.2 : gRPC (protobuf, streaming, interceptors)
   Description : Le service gRPC de la section, DE BOUT EN BOUT et sans
                 ouvrir de port — bufconn, le listener en mémoire recommandé
                 pour les tests. Contrat users/v1/users.proto → code généré
                 commité dans gen/ (régénérable : go generate ./..., via buf).
                 Serveur : Unimplemented…Server embarqué PAR VALEUR, erreur
                 codée status/codes. Client : grpc.NewClient (Dial est
                 déprécié) + stub généré. Démo : appel unaire, NotFound
                 extrait par status.Code, STREAMING serveur lu jusqu'à
                 io.EOF, intercepteur unaire chaîné (le middleware de gRPC).
   Fichier source : 02-grpc.md
   Lancer : go run .            (aucun port, aucun service requis)
   Régénérer : go generate ./...  (nécessite buf — cf. README)
   ============================================================================ */

//go:generate buf generate

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	usersv1 "github.com/acme/app/gen/users/v1"
)

// userServer implémente le service : on EMBARQUE Unimplemented…Server
// (par valeur) pour la compatibilité ascendante — un RPC ajouté au contrat
// ne casse pas la compilation des serveurs existants.
type userServer struct {
	usersv1.UnimplementedUserServiceServer
}

func (s *userServer) GetUser(ctx context.Context, req *usersv1.GetUserRequest) (*usersv1.User, error) {
	if req.Id != 42 {
		// L'erreur CODÉE : le client lira le code avec status.Code(err).
		return nil, status.Errorf(codes.NotFound, "utilisateur %d introuvable", req.Id)
	}
	return &usersv1.User{Id: req.Id, Name: "Alice"}, nil
}

// ListUsers : STREAMING serveur — une requête, un flux de réponses ;
// le retour nil clôt le flux (io.EOF côté client).
func (s *userServer) ListUsers(req *usersv1.ListUsersRequest, stream grpc.ServerStreamingServer[usersv1.User]) error {
	for _, u := range []*usersv1.User{{Id: 1, Name: "Ada"}, {Id: 2, Name: "Alan"}, {Id: 3, Name: "Grace"}} {
		if err := stream.Send(u); err != nil {
			return err
		}
	}
	return nil // fin du flux
}

// loggingUnary : l'intercepteur de la section — il enveloppe chaque RPC
// unaire (journalisation, auth, métriques… : les préoccupations transverses).
func loggingUnary(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req) // appelle la suite de la chaîne
	fmt.Printf("   [intercepteur] %s (%s) → %s\n", info.FullMethod, time.Since(start).Round(time.Microsecond), status.Code(err))
	return resp, err
}

func main() {
	// bufconn : un listener EN MÉMOIRE — le serveur « écoute » sans port.
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(loggingUnary)) // chaînable
	usersv1.RegisterUserServiceServer(srv, &userServer{})
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	// Client : grpc.NewClient (mode « idle », connexion paresseuse) branché
	// sur bufconn par un dialer ; en production : credentials.NewTLS(...).
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := usersv1.NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // comme partout : le context gouverne chaque RPC

	fmt.Println("=== Appel unaire ===")
	u, err := client.GetUser(ctx, &usersv1.GetUserRequest{Id: 42})
	fmt.Println("GetUser(42) →", u.GetName(), "· err =", err)

	fmt.Println("=== Erreur codée : status.Code côté client ===")
	_, err = client.GetUser(ctx, &usersv1.GetUserRequest{Id: 7})
	fmt.Println("GetUser(7)  → code =", status.Code(err), "· message =", status.Convert(err).Message())

	fmt.Println("=== Streaming serveur : Recv jusqu'à io.EOF ===")
	stream, err := client.ListUsers(ctx, &usersv1.ListUsersRequest{})
	if err != nil {
		panic(err)
	}
	for {
		u, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Println("fin du flux (io.EOF)")
			break
		}
		if err != nil {
			panic(err)
		}
		fmt.Println("reçu →", u.GetName())
	}
}
