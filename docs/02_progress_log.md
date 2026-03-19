# Progress Log

## Session Template
- Date:
- Objectif de session:
- Travail fait:
- Blocages:
- Prochaine etape:
- Questions pour le prof:

---

## Session 2026-03-18
- Date: 2026-03-18
- Objectif de session: Initialiser le cadre de pilotage du projet.
- Travail fait: Creation README + plan global + tracker audit + journal + strategie git.
- Blocages: Sujet/audit/questionnaire pas encore colles dans le repo.
- Prochaine etape: Integrer les vrais enonces et prioriser les milestones.
- Questions pour le prof: A renseigner apres lecture du sujet.

## Session 2026-03-18 (cadrage mandatory complet)
- Date: 2026-03-18
- Objectif de session: Integrer le sujet social-network, l audit mandatory et les contraintes reelles.
- Travail fait: README mis a jour, plan global concret cree, checklist audit mandatory complete avec IDs, scope bonus exclu confirme.
- Blocages: Aucun blocage technique pour le cadrage. Les choix de stack et architecture executable restent a verrouiller.
- Prochaine etape: Demarrer M1 (architecture backend/frontend + migrations sqlite + docker skeleton).
- Questions pour le prof: Confirmer qu app-image reste strictement optionnel et hors notation mandatory.

## Session 2026-03-18 (scope lock final)
- Date: 2026-03-18
- Objectif de session: Verrouiller le scope a 100 pourcent mandatory.
- Travail fait: Suppression des references optionnelles, confirmation du mode mandatory strict, alignment des docs de pilotage.
- Blocages: Aucun.
- Prochaine etape: Demarrer implementation M1 sans deviation de scope.
- Questions pour le prof: Aucune sur le scope, il est fixe.

## Session 2026-03-18 (M1 architecture done)
- Date: 2026-03-18
- Objectif de session: Poser la base technique mandatory backend/frontend/db/docker.
- Travail fait: Structure backend server/app/db creee, migrations SQLite ajoutees, frontend Vue+Vite initialise, Dockerfiles + docker-compose ajoutes, build backend/frontend valide.
- Blocages: docker compose plugin absent localement, mais docker-compose v1 disponible.
- Prochaine etape: Demarrer M2 avec register/login/logout + sessions/cookies.
- Questions pour le prof: Aucune.

## Session 2026-03-18 (M2 auth baseline)
- Date: 2026-03-18
- Objectif de session: Implementer authentification mandatory et profil de base.
- Travail fait: Endpoints register/login/logout/me implementes, sessions+cookies actifs, upload avatar jpeg/png/gif ajoute, endpoint profile me + toggle public/private ajoute, UI frontend login/register/profil connectee.
- Blocages: Aucun blocage fonctionnel. Validation docker runtime complete reste en M6.
- Prochaine etape: Implementer followers puis privacy des posts (M3).
- Questions pour le prof: Aucune.

## Session 2026-03-18 (M3 followers + profile access)
- Date: 2026-03-18
- Objectif de session: Implementer et valider toute la logique followers et acces profil public/prive.
- Travail fait: Endpoints users/follows/follow-requests ajoutes (request, accept, decline, unfollow), endpoint profile view avec controle de visibilite, UI frontend pour discovery users + demandes entrantes + actions follow/unfollow.
- Blocages: Aucun blocage technique majeur.
- Prochaine etape: Implementer posts/commentaires + privacy public/almost/private.
- Questions pour le prof: Aucune.

## Session 2026-03-18 (M3 posts + privacy done)
- Date: 2026-03-18
- Objectif de session: Implementer posts/commentaires avec media + privacy mandatory et valider les lignes AQ-POST.
- Travail fait: Endpoints /api/posts, /api/posts/feed, /api/posts/comments ajoutes avec controle privacy (public/almost_private/private), selection followers autorises pour private, upload JPEG/PNG/GIF pour posts/commentaires, affichage feed + commentaires + formulaire de post/comment dans frontend, profil enrichi avec posts visibles.
- Blocages: Ajustement du script de test pour extraire correctement les IDs JSON (la logique applicative etait correcte).
- Prochaine etape: Demarrer M4 (groupes + invitations + demandes + events).
- Questions pour le prof: Aucune.

## Session 2026-03-19 (M4 groupes core)
- Date: 2026-03-19
- Objectif de session: Implementer le coeur groupes mandatory (creation, invitations, demandes d entree) et valider AQ-GRP-001 a AQ-GRP-006.
- Travail fait: Backend groupes ajoute (create/list groups, invites incoming/respond, join requests incoming/respond), regle invitation limitee aux followers de l inviteur, UI frontend ajoutee pour creation groupe + invitation follower + join request + accept/decline.
- Blocages: Aucun blocage applicatif. Scenario e2e passe sur DB temporaire avec preuves chiffrables.
- Prochaine etape: Implementer posts/commentaires de groupe puis events + vote (AQ-GRP-007/008/009).
- Questions pour le prof: Aucune.

## Session 2026-03-19 (M4 group activity done)
- Date: 2026-03-19
- Objectif de session: Finaliser M4 avec posts/commentaires de groupe et events avec vote.
- Travail fait: Endpoints backend ajoutes pour group posts/comments/events/vote, controles membership appliques, UI frontend ajoutee pour gerer activite du groupe selectionne, et scenarios e2e validates pour AQ-GRP-007/008/009.
- Blocages: Aucun blocage fonctionnel. Une modification locale inattendue a ete verifiee puis l etat git a ete nettoye avant finalisation.
- Prochaine etape: Demarrer M5 (chat + notifications temps reel).
- Questions pour le prof: Aucune.

## Session 2026-03-19 (M5 chat + notifications done)
- Date: 2026-03-19
- Objectif de session: Implementer et valider chat prive/groupe temps reel + centre notifications.
- Travail fait: Ajout backend websocket auth (/api/ws), endpoints chat prive/groupe, endpoints notifications (list/read), emission de notifications sur follow request, group invite, group join request et event creation, puis integration frontend (sections Notifications, Private chat, Group chat + ecoute WS realtime).
- Blocages: Aucun blocage fonctionnel. Ajustement proxy WS cote Vite/Nginx pour le mode dev/docker.
- Prochaine etape: Demarrer M6 (docker final + dernier passage audit).
- Questions pour le prof: Aucune.

## Session 2026-03-19 (M2 evidence closure)
- Date: 2026-03-19
- Objectif de session: Fermer les 2 lignes M2 encore en READY (AQ-AUTH-002, AQ-PROF-001).
- Travail fait: Verification scriptable de register/profile pour prouver presence des champs register et absence de password dans /api/profile/me.
- Blocages: Aucun.
- Prochaine etape: Continuer M6.
- Questions pour le prof: Aucune.

## Session 2026-03-19 (M6 docker + audit final done)
- Date: 2026-03-19
- Objectif de session: Valider les lignes Docker AQ-DOCK-001 et AQ-DOCK-002, puis clore le mandatory.
- Travail fait: Execution docker-compose up --build -d, verification des conteneurs backend/frontend avec tailles virtuelles non nulles, verification HTTP frontend/backend (3000 et 8080), correction du Dockerfile frontend (suppression de directives backend erronees), puis validation finale des 2 lignes Docker.
- Blocages: Erreur compose v1 lors de recreation frontend (ContainerConfig) contournee par nettoyage du conteneur stale puis recreation.
- Prochaine etape: Stabilisation finale pour soutenance (gel version + checks de routine).
- Questions pour le prof: Aucune.
