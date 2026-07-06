/* ============================================================================
   Section 13.2 : Mocks par interfaces, testify et httptest
   Description : La directive go:generate du cours — mockgen (épinglé dans
                 go.mod par « go get -tool », Go 1.24) génère un mock typé de
                 l'interface Mailer. On lance « go generate ./... » avant de
                 committer ; le mock est commité avec le reste.
   Fichier source : 02-mocks-testify.md
   ============================================================================ */

package notify

//go:generate go tool mockgen -source=service.go -destination=mock_mailer_test.go -package=notify
