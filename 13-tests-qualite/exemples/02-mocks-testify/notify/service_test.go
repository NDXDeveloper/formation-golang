/* ============================================================================
   Section 13.2 : Mocks par interfaces, testify et httptest
   Description : Le spectre des doublures de test de la section, du plus
                 simple au plus outillé, sur la MÊME petite interface :
                 (1) un STUB à champ-fonction (injecte un comportement) ;
                 (2) un SPY qui enregistre les appels (vérifie les
                 interactions) ; (3) testify/mock (attentes programmables
                 On/Return, AssertExpectations) ; (4) mockgen + gomock
                 (EXPECT typé). Un dernier test montre le NÉGATIF : une
                 attente non honorée fait échouer AssertExpectations.
   Fichier source : 02-mocks-testify.md
   Lancer : go generate ./... && go test ./...
   ============================================================================ */

package notify

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// (1) STUB à champ-fonction : léger, on injecte le comportement voulu.
type mailerStub struct {
	sendFunc func(ctx context.Context, to, subject, body string) error
}

func (m mailerStub) Send(ctx context.Context, to, subject, body string) error {
	return m.sendFunc(ctx, to, subject, body)
}

func TestWelcome_transmetLErreur(t *testing.T) {
	panne := errors.New("SMTP indisponible")
	svc := New(mailerStub{
		sendFunc: func(context.Context, string, string, string) error { return panne },
	})
	if err := svc.Welcome(context.Background(), "a@exemple.fr"); !errors.Is(err, panne) {
		t.Fatalf("Welcome err = %v, want %v", err, panne)
	}
}

// (2) SPY : enregistre les appels pour vérifier les interactions.
type mailerSpy struct{ destinataires []string }

func (m *mailerSpy) Send(_ context.Context, to, _, _ string) error {
	m.destinataires = append(m.destinataires, to)
	return nil
}

func TestWelcome_envoieAuBonDestinataire(t *testing.T) {
	spy := &mailerSpy{}
	if err := New(spy).Welcome(context.Background(), "b@exemple.fr"); err != nil {
		t.Fatalf("Welcome: %v", err)
	}
	if want := []string{"b@exemple.fr"}; !slices.Equal(spy.destinataires, want) {
		t.Errorf("destinataires = %v, want %v", spy.destinataires, want)
	}
}

// (3) testify/mock : attentes programmables.
type MailerMock struct{ mock.Mock }

func (m *MailerMock) Send(ctx context.Context, to, subject, body string) error {
	return m.Called(ctx, to, subject, body).Error(0)
}

func TestWelcome_avecTestifyMock(t *testing.T) {
	m := new(MailerMock)
	m.On("Send", mock.Anything, "c@exemple.fr", mock.Anything, mock.Anything).
		Return(nil).Once()
	require.NoError(t, New(m).Welcome(context.Background(), "c@exemple.fr"))
	m.AssertExpectations(t) // échoue si Send n'a pas été appelé comme prévu
}

// (4) mockgen + gomock : EXPECT typé (le mock est généré par go:generate).
func TestWelcome_avecMockgen(t *testing.T) {
	ctrl := gomock.NewController(t) // enregistre un Cleanup : pas de defer Finish
	m := NewMockMailer(ctrl)
	m.EXPECT().Send(gomock.Any(), "d@exemple.fr", gomock.Any(), gomock.Any()).Return(nil)
	require.NoError(t, New(m).Welcome(context.Background(), "d@exemple.fr"))
}

// NÉGATIF : une attente non honorée fait échouer AssertExpectations.
func TestMock_attenteNonHonoree(t *testing.T) {
	m := new(MailerMock)
	m.On("Send", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()
	// on n'appelle PAS Welcome : l'attente reste insatisfaite
	fake := new(testing.T)
	if m.AssertExpectations(fake); !fake.Failed() {
		t.Fatal("AssertExpectations aurait dû échouer (attente non honorée)")
	}
}
