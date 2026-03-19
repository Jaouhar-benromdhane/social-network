# Social Network - Working README

## 1) But du repo
Ce repo sert a construire le projet social-network pas a pas, de facon propre, testable et explicable devant le prof.

Le scope principal actuel est le mandatory.
Les bonus ne sont pas pris.
Tout ce qui est optionnel est exclu.

## 2) Etat actuel
- Sujet mandatory recu: oui
- Audit mandatory recu: oui
- Questionnaire mandatory recu: oui
- Bonus: hors scope
- Optionnel app-image: hors scope
- M1 Architecture: done
- M2 Auth + Profiles: done
- M3 Followers + Posts: done
- M4 Groups + Events: done

## 3) Documents de suivi
- docs/00_plan_global.md
- docs/01_audit_questions.md
- docs/02_progress_log.md
- docs/03_decisions.md
- docs/04_git_strategy.md

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
1. Demarrer M5: chat prive en temps reel.
2. Etendre au chat de groupe temps reel.
3. Ajouter le centre de notifications global.
4. Valider les lignes AQ-CHAT et AQ-NOTIF avec scenarios multi navigateurs.
5. Finaliser M6 docker + verification audit complete.

## 8) Run local (M1)
Backend:
1. cd backend
2. go run ./cmd/server

Frontend:
1. cd frontend
2. npm install
3. npm run dev

Verification rapide:
1. Ouvrir http://localhost:5173
2. Verifier que le statut API devient ok
3. Verifier http://localhost:8080/api/health

## 9) Run Docker (M1)
Si docker compose plugin est indisponible localement, utiliser docker-compose.

1. docker-compose -f docker-compose.yml up --build
2. Frontend: http://localhost:3000
3. Backend health: http://localhost:8080/api/health
