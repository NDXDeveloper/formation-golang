/* ============================================================================
   Section 13.2 : Mocks par interfaces, testify et httptest
   Description : Le levier de la section — une PETITE interface définie côté
                 CONSOMMATEUR (Mailer : une seule méthode), injectée dans
                 Service. Parce qu'elle est minuscule, la doubler pour les
                 tests ne coûte presque rien : c'est ce qui permet à Go de se
                 passer le plus souvent de framework de mocking.
   Fichier source : 02-mocks-testify.md
   ============================================================================ */

package notify

import "context"

// Mailer est l'unique dépendance dont Service a besoin — interface étroite,
// déclarée ICI (côté consommateur), implémentée ailleurs.
type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

type Service struct {
	mailer Mailer
}

func New(m Mailer) *Service { return &Service{mailer: m} }

func (s *Service) Welcome(ctx context.Context, email string) error {
	return s.mailer.Send(ctx, email, "Bienvenue", "Merci de votre inscription.")
}
