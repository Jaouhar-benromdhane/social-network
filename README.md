# Social Network

Projet de reseau social (mandatory) avec backend Go, frontend Vue 3 et base SQLite.

Le projet couvre:
- authentification par session
- profils publics/prives + systeme de follow
- posts/commentaires avec medias
- groupes (invites, demandes d acces, activite)
- chat prive/groupe en temps reel
- notifications en temps reel

## Stack technique
- Backend: Go
- Frontend: Vue 3 + Vite
- Base de donnees: SQLite
- Realtime: WebSocket
- Conteneurisation: Docker + docker compose

## Lancer le projet (recommande)
Depuis la racine du repo:

```bash
./server.sh start
```

Arret:

```bash
./server.sh stop
```

Redemarrage:

```bash
./server.sh restart
```

Etat:

```bash
./server.sh status
```

Acces:
- Frontend: http://localhost:3000
- Backend health: http://localhost:8080/api/health

## Lancer manuellement (sans script)
```bash
docker compose -f docker-compose.yml up --build
```

Si `docker compose` n est pas disponible:
```bash
docker-compose -f docker-compose.yml up --build
```

## Lancer en dev local (hors Docker)
Backend:
```bash
cd backend
go run ./cmd/server
```

Frontend:
```bash
cd frontend
npm install
npm run dev
```

## Fonctionnalites principales
- Register / login / logout / session persistante
- Profils publics/prives avec controle d acces
- Follow direct (profil public) ou follow request (profil prive)
- Feed avec posts `public`, `almost_private`, `private`
- Commentaires et upload media (jpg/png/gif)
- Groupes: creation, invitation, demande de join, moderation
- Activite groupe: posts, commentaires, events, votes
- Chat prive (mutual follow requis)
- Chat groupe (membre requis)
- Notifications et mises a jour temps reel

## Notes utiles pour la soutenance
- Checklist audit rapide: [docs/05_audit_final_checklist.md](docs/05_audit_final_checklist.md)
- Script de pilotage serveur: [server.sh](server.sh)

## Structure du repo
- `backend/`: API Go, logique metier, DB
- `frontend/`: application Vue
- `docs/`: documentation de suivi et checklist
- `docker-compose.yml`: orchestration backend/frontend
