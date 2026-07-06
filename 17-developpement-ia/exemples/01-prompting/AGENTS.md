# Instructions Go pour les assistants

<!-- Section 17.1 (01-prompting-go.md) : les instructions PERSISTANTES du projet.
     Convention transverse AGENTS.md, lue par plusieurs assistants (Copilot,
     Claude Code, Cursor, Gemini…). Principe cardinal : ne rien gaspiller sur ce
     que gofmt/go vet/staticcheck imposent déjà — ne coder que les idiomes
     NON mécanisables, et justifier chaque règle. -->

## Idiomes (non couverts par l'outillage)

- Erreurs explicites : retourner `error`, jamais `panic` pour le contrôle de flux.
  Wrapper avec `%w` et un contexte : `fmt.Errorf("lecture %q : %w", nom, err)`.
- Composition, pas héritage : embedding et petites interfaces définies côté
  consommateur — pas une grande interface par type.
- stdlib d'abord : justifier toute dépendance ; pas de framework HTTP ni d'ORM
  sans raison explicite.
- Pas de getters/setters à la Java ; exporter les champs quand c'est légitime.
- Concurrence : `context.Context` en premier paramètre ; pas de goroutine dont
  personne n'attend la fin.

## Tests

- Table-driven avec sous-tests `t.Run` ; pas de framework d'assertion sauf demande.

## Vérification

- Le code doit passer `gofmt`, `go vet` et `staticcheck`.
