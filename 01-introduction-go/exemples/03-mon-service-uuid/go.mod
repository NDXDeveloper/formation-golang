// ============================================================================
// Section 1.3 : L'écosystème Go (toolchain, modules, cycle de release)
// Description : Le go.mod d'exemple de la section, tel quel — chemin de
//               module, directive go, dépendance épinglée (require)
// Fichier source : 03-ecosysteme-go.md
// ============================================================================

module github.com/exemple/mon-service

go 1.26.0

require github.com/google/uuid v1.6.0
