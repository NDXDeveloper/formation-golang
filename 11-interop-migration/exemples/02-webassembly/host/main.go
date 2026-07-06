/* ============================================================================
   Section 11.2 : WebAssembly (WASI)
   Description : L'HÔTE — le bloc wazero de la section : un moteur WebAssembly
                 EN GO PUR, sans cgo (la cross-compilation reste intacte,
                 CGO_ENABLED=0 fonctionne). Il charge add.wasm (l'invité
                 ci-contre), lui accorde les imports WASI (la sécurité par
                 capacités : sans cet octroi explicite, le module ne peut
                 RIEN), l'instancie en mode réacteur (_initialize) puis
                 appelle la fonction exportée add(2, 3). Écho direct du
                 « préférer le Go pur » du § 11.1.
   Fichier source : 02-webassembly.md
   Lancer : go run .   (après avoir construit l'invité — voir le README ;
            le .md montre //go:embed : ici on lit le fichier pour ne pas
            committer de binaire dans le dépôt)
   ============================================================================ */

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

func main() {
	ctx := context.Background()

	// L'invité, produit par n'importe quel langage compilé en wasm (ici : Go).
	// (Le cours utilise //go:embed add.wasm — même logique, binaire embarqué.)
	addWasm, err := os.ReadFile("add.wasm")
	if err != nil {
		log.Fatal("construisez d'abord l'invité (voir README) : ", err)
	}

	// Un runtime en Go pur : aucune dépendance C, la cross-compilation reste intacte.
	r := wazero.NewRuntime(ctx)
	defer r.Close(ctx)

	// La sécurité par capacités : l'hôte ACCORDE les imports WASI à l'invité.
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	// Module RÉACTEUR : on appelle _initialize (pas _start, qui terminerait).
	compiled, err := r.CompileModule(ctx, addWasm)
	if err != nil {
		log.Fatal(err)
	}
	cfg := wazero.NewModuleConfig().WithStartFunctions("_initialize")
	mod, err := r.InstantiateModule(ctx, compiled, cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Appel d'une fonction exportée par le module « invité ».
	res, err := mod.ExportedFunction("add").Call(ctx, 2, 3)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res[0]) // 5
}
