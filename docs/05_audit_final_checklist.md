# Audit Final Checklist (1 Page)

Objectif: verifier rapidement que tout le mandatory est pret avant soutenance.

Date:
Auditeur:
Version / commit:

## 1) Boot & Docker
- [ ] ./server.sh stop puis ./server.sh start sans blocage
- [ ] Frontend accessible: http://localhost:3000
- [ ] Backend health OK: http://localhost:8080/api/health (status ok)
- [ ] docker ps montre backend + frontend up

## 2) Auth & Session
- [ ] Register OK
- [ ] Login OK
- [ ] /api/auth/me OK apres refresh
- [ ] Logout OK
- [ ] Permissions sans session: routes protegees refusent l acces

## 3) Profile & Follow
- [ ] Profil public visible sans follow
- [ ] Profil prive refuse sans follow
- [ ] Follow request private: pending + accept/decline OK
- [ ] Follow public: auto-follow OK
- [ ] Notification follow visible (temps reel + liste)

## 4) Posts
- [ ] Create post public OK
- [ ] Create post almost_private OK
- [ ] Create post private avec users autorises OK
- [ ] Comment post OK
- [ ] Upload media (jpg/png/gif) post/comment OK

## 5) Groups
- [ ] Create group OK
- [ ] Invite follower OK
- [ ] Accept/decline invite OK
- [ ] Join request OK
- [ ] Accept/decline join request OK
- [ ] Group posts/comments/events/votes OK

## 6) Realtime (WS)
- [ ] Private chat realtime OK
- [ ] Group chat realtime OK
- [ ] Notifications realtime OK
- [ ] Updates posts sans F5 OK
- [ ] Updates groupes sans F5 OK

## 7) Access Rules
- [ ] Private chat interdit sans mutual follow
- [ ] Group chat interdit hors membre
- [ ] Actions groupe interdites hors membre/role

## 8) UI & Demo Flow
- [ ] Navigation pages OK: #/profile #/posts #/groups #/chat
- [ ] Aucune action critique ne demande refresh manuel
- [ ] Erreurs utilisateur lisibles (403/401/validation)

## 9) Quick Evidence to Keep Ready
- [ ] 1 capture docker up + health
- [ ] 1 capture follow request + notif
- [ ] 1 capture private chat realtime
- [ ] 1 capture group flow (invite/join/message)

## 10) Final Go/No-Go
- [ ] Tout coche: GO soutenance
- [ ] Si NOK: corriger puis refaire sections 1, 6, 7 minimum
