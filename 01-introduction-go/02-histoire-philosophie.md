🔝 Retour au [Sommaire](/SOMMAIRE.md)

# 1.2 Histoire et philosophie du langage

Comprendre *pourquoi* Go a été créé éclaire l'essentiel de ses choix de conception. Son histoire est courte, mais sa philosophie est d'une remarquable constance : **la simplicité et la lisibilité y priment sur presque tout le reste.**

Là où la [section 1.1](01-quest-ce-que-go.md) répondait au « quoi » et au « pour quoi faire », cette section s'attache au « pourquoi » — l'origine du langage et les principes qui le gouvernent encore aujourd'hui.

## Aux origines : les problèmes de Google

Go naît en **septembre 2007** chez Google, de l'initiative de trois ingénieurs : **Robert Griesemer, Rob Pike et Ken Thompson**. L'anecdote fondatrice est restée célèbre : lassés d'attendre de longues minutes la compilation d'un gros programme C++, ils commencent à esquisser un langage qui règlerait, entre autres, ce problème-là.

Car le vrai moteur du projet, c'est la difficulté de **développer du logiciel à l'échelle de Google** : des bases de code énormes, des temps de compilation qui s'allongent, des dépendances qui s'enchevêtrent, et une concurrence (au sens du parallélisme) difficile à écrire correctement dans les langages de l'époque. Go est d'abord une réponse pragmatique à ces frustrations d'ingénierie.

Le pedigree de ses auteurs explique beaucoup. Pike et Thompson viennent de l'univers Unix et Plan 9 (Thompson est l'un des pères d'Unix ; Pike et lui ont conçu l'encodage UTF-8), et ont exploré avant Go plusieurs langages concurrents — Newsqueak, Alef, Limbo — dont l'héritage se retrouve directement dans le modèle de concurrence de Go. Griesemer, lui, avait travaillé sur des moteurs et compilateurs (V8, HotSpot, le langage Sawzall).

L'ambition tenait en une synthèse : réunir **l'efficacité et la sûreté d'un langage compilé et typé statiquement** (à la manière de C++) avec **la simplicité et la productivité d'un langage dynamique** (à la manière de Python), en y ajoutant une **concurrence de première classe**.

## Les grandes étapes

| Année | Étape |
|-------|-------|
| 2007 | Conception du langage chez Google (Griesemer, Pike, Thompson). |
| 2009 | Annonce publique et passage en open source (novembre). |
| 2012 | Sortie de **Go 1.0**, assortie de la **promesse de compatibilité Go 1**. |
| 2022 | Arrivée des **génériques** (Go 1.18), après des années de maturation prudente. |
| Depuis | Évolution régulière, au rythme d'une version tous les six mois environ. |

L'emblème du langage, le **gopher** (dessiné par Renée French), est devenu l'une des mascottes les plus reconnaissables de la programmation.

Le rythme de publication et les évolutions récentes sont détaillés en [section 1.3](03-ecosysteme-go.md) et au [module 18](../18-strategie-roadmap/README.md).

## Une philosophie : la simplicité comme principe directeur

Le trait le plus frappant de Go est ce qu'il **choisit de ne pas avoir**. La spécification du langage est courte, on y compte 25 mots-clés, et bon nombre de constructions habituelles ont été volontairement écartées :

- pas d'héritage de classes — mais de la **composition** (*embedding*) ;
- pas d'exceptions — mais des **erreurs traitées comme des valeurs** ;
- pas de surcharge d'opérateurs, pas de conversions implicites, pas d'opérateur ternaire ;
- des génériques ajoutés tardivement (2022) et à dessein limités, pour ne pas alourdir le langage.

Ce minimalisme a un coût — le code Go est parfois plus verbeux — mais il offre en retour de l'**uniformité** et de la **prévisibilité** : il y a généralement une manière évidente de faire les choses, et le code d'un projet ressemble à celui d'un autre.

> **Proverbe Go** — *« Clear is better than clever »* : un code clair vaut mieux qu'un code astucieux.

## La lisibilité avant tout

Un principe irrigue toute la conception : **on lit le code bien plus souvent qu'on ne l'écrit.** Go optimise donc pour la personne qui lit, et pour les grandes équipes qui maintiennent un code sur de longues années. L'équipe Go aime d'ailleurs distinguer la *programmation* du *génie logiciel* : ce dernier, c'est la programmation à laquelle on ajoute le temps qui passe et d'autres développeurs.

L'outil `gofmt` incarne cette priorité : il impose un format unique à tout le code Go. Personne ne choisit ce style, mais tout le monde y gagne — les débats d'indentation et d'accolades disparaissent, et n'importe quelle base de code Go paraît immédiatement familière.

> **Proverbe Go** — *« gofmt's style is no one's favorite, yet gofmt is everyone's favorite. »*

## Les partis pris de conception

La philosophie de Go se décline en quelques choix structurants, que vous retrouverez tout au long de la formation :

- **Explicite plutôt qu'implicite.** Les erreurs sont des valeurs que l'on traite ouvertement, et il n'y a pas de flot de contrôle caché (pas d'exceptions surgissant à distance). → [section 2.9](../02-fondamentaux-langage/09-gestion-erreurs.md).
  > **Proverbe Go** — *« Errors are values. »*
- **Composition plutôt qu'héritage.** L'*embedding* et les **interfaces satisfaites implicitement** remplacent les hiérarchies de classes. → [module 3](../03-types-interfaces/README.md).
- **Concurrence par la communication.** Inspiré des *Communicating Sequential Processes* de Tony Hoare, Go privilégie l'échange par *channels* au partage de mémoire. → [module 4](../04-concurrence/README.md).
  > **Proverbe Go** — *« Don't communicate by sharing memory; share memory by communicating. »*
- **La compilation rapide comme objectif.** Le modèle de dépendances a été pensé pour que les builds restent rapides, même à grande échelle.
- **Un outillage et un style uniformes.** Format, tests, analyse statique : fournis avec le langage, identiques partout.
- **La stabilité comme engagement.** La promesse de compatibilité Go 1 garantit que le code d'aujourd'hui compilera demain. La gouvernance qui la sous-tend est détaillée en [section 18.1](../18-strategie-roadmap/01-gouvernance-compatibilite.md).

## Un compromis assumé

Venant d'un langage riche en fonctionnalités, on peut d'abord trouver Go **trop simple**, voire répétitif. C'est un choix, pas une lacune. Le pari de ses concepteurs est que la simplicité paie sur la durée : prise en main rapide, revues de code plus faciles, moins de « magie » à déchiffrer, et un code qui reste lisible des années plus tard, y compris par quelqu'un d'autre.

Ce parti pris pragmatique — préférer la clarté à l'exhaustivité, l'ingénierie de long terme à l'élégance immédiate — est la clé de lecture de tout ce qui suit. Le condensé d'*Effective Go* et les anti-patterns de l'[annexe B](../annexes/go-idiomatique/README.md) prolongent utilement ce chapitre.

## En résumé

Go est né chez Google (2007-2012) d'une volonté pragmatique : rendre le développement à grande échelle plus simple, plus rapide et plus fiable. Sa philosophie — **simplicité, lisibilité, choix explicites, composition, concurrence par communication et stabilité dans le temps** — n'a guère varié depuis, et elle sous-tend chacune des décisions techniques que vous rencontrerez. Reste à voir comment cet état d'esprit se traduit concrètement dans l'outillage : c'est l'objet de la [section suivante](03-ecosysteme-go.md).

---

[🔝 Sommaire](../SOMMAIRE.md) · [⏭️ Section suivante : 1.3 L'écosystème Go](03-ecosysteme-go.md)

⏭ [L'écosystème Go (toolchain, modules, cycle de release semestriel)](/01-introduction-go/03-ecosysteme-go.md)
