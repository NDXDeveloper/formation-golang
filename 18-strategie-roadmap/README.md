🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 18. Stratégie, feuille de route et ressources

Dix-sept modules durant, cette formation a enseigné *comment* écrire du Go. Ce dernier module prend de la hauteur pour répondre aux questions qu'une équipe finit toujours par se poser : **Go est-il un pari sûr sur la durée ? Comment le langage évolue-t-il ? Et comment rester à jour ?**

La réponse courte tient dans le trait qui distingue Go de la plupart des langages : il évolue **avec prudence**, et **il ne casse pas votre code**. Ce module ne parle pas de nouvelle syntaxe, mais de *durabilité* et de *continuité*.

## Le point d'ancrage : la promesse de compatibilité

L'atout stratégique majeur de Go est sa **promesse de compatibilité Go 1.x** : du code écrit pour Go 1.0 en 2012 compile encore aujourd'hui. Cette stabilité assumée est l'une des raisons pour lesquelles Go est un choix à faible risque pour des systèmes appelés à durer — c'est elle qui permet de bâtir sur Go sans redouter la prochaine version. Elle complète la grille « quand choisir Go » du [§ 1.6](../01-introduction-go/06-positionnement-2026.md) : à la question « ce langage sera-t-il encore là, inchangé, dans dix ans ? », Go répond oui.

Deux mécanismes soutiennent cette confiance : une **cadence semestrielle prévisible** (une version en février, une en août — cf. [§ 1.3](../01-introduction-go/03-ecosysteme-go.md)) et une **gouvernance publique et transparente** (un processus de *proposals* ouvert). La prévisibilité est, en soi, une valeur.

Enfin, Go évolue **lentement et prudemment** : les génériques, puis des ajouts mesurés, jamais la nouveauté pour la nouveauté — souvent introduits derrière des garde-fous (expériences `GOEXPERIMENT`, réglages `GODEBUG`) qui préservent l'existant. Toutes les fonctionnalités 🆕 croisées au fil de la formation — `log/slog`, `os.Root`, le PGO, le GC « Green Tea », le TLS post-quantique, les modernizers de `go fix` — ont été introduites **sans rien casser**.

## Ce que ce module couvre — et ce qu'il ne duplique pas

Trois angles, du plus institutionnel au plus pratique : la gouvernance et la promesse, la trajectoire, puis les ressources pour continuer seul.

Un point de cohérence : la [section 18.1](01-gouvernance-compatibilite.md) explique la *gouvernance et la promesse de compatibilité* — le pourquoi et le comment ; les **tableaux de référence des versions** vivent, eux, dans l'[annexe H](../annexes/versions-reference/README.md) — le quoi. Et l'argument « pourquoi choisir Go » est en [§ 1.6](../01-introduction-go/06-positionnement-2026.md) : ce module en apporte la moitié « durabilité ».

## 🎯 Objectifs du module

À l'issue de ce module, vous saurez :

- Comprendre comment Go est gouverné et pourquoi la promesse de compatibilité en fait un pari sûr sur la durée.
- Situer les évolutions récentes du langage et lire sa feuille de route.
- Savoir où regarder pour rester à jour et continuer à progresser.

## 📋 Prérequis

Aucun prérequis technique particulier : ce module de clôture prend le reste de la formation comme toile de fond, en particulier l'[écosystème](../01-introduction-go/03-ecosysteme-go.md) (1.3) et le [positionnement de Go](../01-introduction-go/06-positionnement-2026.md) (1.6).

## 🗺️ Contenu du module

### 18.1 · [Gouvernance du langage, proposals, promesse de compatibilité Go 1.x](01-gouvernance-compatibilite.md)
Qui décide, comment une évolution est proposée et acceptée, et ce que garantit — exactement — la promesse de compatibilité.

### 18.2 · [Roadmap et évolutions récentes du langage](02-roadmap.md)
La trajectoire : d'où vient Go, où il va, et comment lire les évolutions récentes sans se laisser distraire par le bruit.

### 18.3 · [Communauté, veille et ressources pour continuer](03-communaute-ressources.md)
Où suivre les annonces, à quelles sources se fier, et comment continuer à apprendre une fois la formation refermée.

## Pour finir

C'est ici que la formation s'achève — mais la stabilité de Go fait que ce que vous avez appris reste valable. La promesse de compatibilité est aussi une promesse qui vous est faite : votre code et vos connaissances vieillissent bien. Les [ressources de la section 18.3](03-communaute-ressources.md) et les annexes sont là pour la suite.

---

🔝 [Sommaire](../SOMMAIRE.md) · ⏭️ [18.1 Gouvernance, proposals, compatibilité](01-gouvernance-compatibilite.md)

⏭ [Gouvernance du langage, proposals, promesse de compatibilité Go 1.x](/18-strategie-roadmap/01-gouvernance-compatibilite.md)
