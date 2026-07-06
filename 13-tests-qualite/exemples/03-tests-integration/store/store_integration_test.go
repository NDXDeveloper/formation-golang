//go:build integration

/* ============================================================================
   Section 13.3 : Tests d'intégration avec Testcontainers
   Description : Le patron complet de la section. La BALISE DE BUILD
                 « //go:build integration » isole ces tests de la boucle
                 unitaire rapide (ils ne tournent qu'avec -tags=integration).
                 Le module dédié postgres.Run démarre PostgreSQL ; ORDRE
                 CRUCIAL du cours : testcontainers.CleanupContainer(t, c) est
                 posé AVANT le require.NoError, pour nettoyer même si le
                 démarrage échoue. Chaque sous-test s'isole via Snapshot /
                 Restore — et le piège vérifié du cours : Snapshot ET Restore
                 exigent ZÉRO session ouverte (ils font CREATE/DROP TEMPLATE),
                 d'où le pool fermé/rouvert autour de Restore.
   Fichier source : 03-tests-integration.md
   Lancer : go test -tags=integration ./...   (Docker requis)
   ============================================================================ */

package store

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // enregistre le driver database/sql « pgx »
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupPostgres(t *testing.T) (*pgxpool.Pool, *postgres.PostgresContainer, string) {
	t.Helper()
	ctx := context.Background()

	pg, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("app"),
		postgres.WithUsername("u"),
		postgres.WithPassword("p"),
		// Snapshot/Restore via une vraie connexion pgx (pas le fallback docker exec).
		postgres.WithSQLDriver("pgx"),
		// Postgres redémarre au premier boot : attendre « ready » DEUX fois.
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(60*time.Second)),
	)
	testcontainers.CleanupContainer(t, pg) // AVANT le check : nil-safe
	require.NoError(t, err)

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	require.NoError(t, New(pool).Migrate(ctx))
	// prendre l'instantané de référence : AUCUNE session ne doit être ouverte
	pool.Close()
	require.NoError(t, pg.Snapshot(ctx))
	pool, err = pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return pool, pg, dsn
}

func TestStore_CreateGet(t *testing.T) {
	ctx := context.Background()
	pool, _, _ := setupPostgres(t)
	st := New(pool)

	created, err := st.Create(ctx, "alice@exemple.fr")
	require.NoError(t, err)
	require.NotZero(t, created.ID)

	got, err := st.Get(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, "alice@exemple.fr", got.Email)

	// utilisateur absent → ErrNotFound (l'erreur typée du store)
	_, err = st.Get(ctx, 99999)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestStore_IsolationParRestore(t *testing.T) {
	ctx := context.Background()
	pool, pg, dsn := setupPostgres(t)

	// insérer, puis restaurer l'instantané : la table doit redevenir vide
	_, err := New(pool).Create(ctx, "bob@exemple.fr")
	require.NoError(t, err)

	pool.Close() // Restore exige lui aussi zéro session ouverte
	require.NoError(t, pg.Restore(ctx))
	pool, err = pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	var n int
	require.NoError(t, pool.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&n))
	require.Equal(t, 0, n, "Restore doit ramener la base à l'instantané vide")
}
