/* ============================================================================
   Section 11.2 : WebAssembly (WASI)
   Description : Le module INVITÉ — du Go compilé vers wasip1 qui EXPORTE une
                 fonction à l'hôte via go:wasmexport (Go 1.24+). Compilé en
                 mode « réacteur » (bibliothèque) avec -buildmode=c-shared :
                 l'hôte appelle _initialize puis les exports, au lieu de
                 lancer _start et de voir le module se terminer. C'est le
                 wasm « produit par n'importe quel langage » du cours — ici,
                 par Go lui-même.
   Fichier source : 02-webassembly.md
   Construire : GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o ../host/add.wasm .
   ============================================================================ */

package main

//go:wasmexport add
func add(a, b int32) int32 { return a + b }

func main() {} // requis, mais l'invité sert de bibliothèque (réacteur)
