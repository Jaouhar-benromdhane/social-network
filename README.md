# Social Network - Working README

## 1) But du repo
Ce repo sert a construire le projet social-network pas a pas, de facon propre .

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
- M5 Chat + Notifications: done
- M6 Docker + Audit final: done

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
1. Geler la version de soutenance (tag/release local).
2. Repasser un smoke test complet avant presentation.
3. Garder les scripts e2e de preuve a disposition.
4. Verifier les ports libres et l environnement demo le jour J.
5. Presenter les preuves audit par milestone.

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

## 10) Recap oral soutenance (10 points)
1. Le projet respecte le scope mandatory a 100 pourcent (bonus et optionnel exclus).
2. Architecture claire: backend Go, frontend Vue, base SQLite, migrations versionnees.
3. Authentification par sessions/cookies (register, login, logout, me) avec controle d acces serveur.
4. Profils publics/prives avec regles de visibilite et gestion followers/follow requests.
5. Posts/commentaires avec media (jpeg/png/gif) et privacy public/almost_private/private.
6. Groupes complets: creation, invitations, demandes d entree, accept/decline, membership checks.
7. Activity de groupe: posts/commentaires internes, events avec options et vote unique par user.
8. Chat temps reel via websocket: prive (mutual follow) + groupe (membres seulement) + support emojis.
9. Notifications centralisees (follow request, group invite, join request, event created) avec mark read.
10. Livraison docker validee: 2 conteneurs up, app accessible via navigateur, preuves audit renseignees.
