# Social Network - Working README

## 1) But du repo
Ce repo sert a construire le projet social-network pas a pas, de facon propre, testable et explicable devant le prof.

Le scope principal actuel est le mandatory.
Les bonus ne sont pas pris.
Le projet optionnel app-image (Electron) est trace dans un document separe.

## 2) Etat actuel
- Sujet mandatory recu: oui
- Audit mandatory recu: oui
- Questionnaire mandatory recu: oui
- Bonus: hors scope
- App image optionnel: recu, planifie apres le mandatory

## 3) Documents de suivi
- docs/00_plan_global.md
- docs/01_audit_questions.md
- docs/02_progress_log.md
- docs/03_decisions.md
- docs/04_git_strategy.md
- docs/05_app_image_optional.md

## 4) Methode de travail (pas a pas)
1. Transformer chaque exigence en tache testable.
2. Lier chaque tache a une ou plusieurs questions d audit.
3. Construire une milestone a la fois.
4. Verifier en local avec scenarios navigateur multiples.
5. Mettre a jour la preuve dans la checklist audit.
6. Faire un commit propre par grosse partie.
7. Push et journaliser la session.

## 5) Definition de termine
Une partie est terminee si:
- le code fonctionne
- la logique est comprise
- l explication orale est prete
- les questions d audit liees sont repondues
- un commit clair est fait

## 6) Routine de debut/fin de session
Debut:
- Lire docs/02_progress_log.md (derniere session)
- Lire docs/01_audit_questions.md (questions ouvertes)
- Choisir un objectif concret de session

Fin:
- Mettre a jour docs/02_progress_log.md
- Cocher ce qui est valide dans docs/01_audit_questions.md
- Noter les decisions dans docs/03_decisions.md
- Faire commit + push

## 7) Prochaines actions concretes
1. Fixer la stack frontend JS framework.
2. Poser l architecture backend (server/app/db + migrations sqlite).
3. Initialiser docker backend/frontend.
4. Implementer auth/session/cookies avant le reste.
