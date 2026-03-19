# Plan Global (Mandatory First)

## 1) Vision
Livrer un reseau social type Facebook conforme a l audit mandatory.

Contraintes:
- JS framework obligatoire pour le frontend
- backend en Go avec SQLite
- sessions et cookies obligatoires
- websocket obligatoire pour chat temps reel
- migrations obligatoires et bien organisees
- 2 conteneurs Docker (backend + frontend)
- bonus non inclus

## 2) Milestones
| Milestone | Objectif | Livrables | Critere de validation | Statut |
|---|---|---|---|---|
| M0 - Cadrage | Verrouiller le scope mandatory et l audit | Checklist audit complete, plan concret | Plus aucune ambiguite sur le scope | DONE |
| M1 - Architecture | Poser la base technique backend/frontend/db/docker | Arborescence, scripts, migrations sqlite, run local | Le projet demarre et migre la DB | DONE |
| M2 - Auth + Profiles | Implementer inscription, login, session, profil public/prive | Formulaires, middleware session, pages profils | Scenarios auth/profiles valides | DONE |
| M3 - Followers + Posts | Implementer suivi + posts/commentaires avec privacy | Follow requests, feed, permissions de visibilite | Scenarios followers/posts valides | DONE |
| M4 - Groups + Events | Implementer groupes, invitations, demandes, events | Groupes, demandes entree, vote event | Scenarios groupes/evenements valides | DONE |
| M5 - Chat + Notifications | Implementer chat prive/groupe temps reel + notifications | WS hub, canaux prive/groupe, centre notif global | Scenarios chat/notifs valides | TODO |
| M6 - Docker + Audit final | Stabiliser et valider tout l audit mandatory | 2 images/containers, checklist finalisee | Passage checklist mandatory complet | TODO |

## 3) Ordre de travail recommande
1. M1 d abord pour eviter les reworks structuraux.
2. M2 ensuite car toutes les features dependent de l authentification.
3. M3 puis M4 pour la logique sociale et communautaire.
4. M5 apres, quand les permissions de base existent deja.
5. M6 en fin de mandatory.

## 4) Decoupage sprint (indicatif)
| Sprint | Focus | Resultat attendu |
|---|---|---|
| S1 | M1 | Base technique stable |
| S2 | M2 | Auth + profils fonctionnels |
| S3 | M3 | Followers + posts operationnels |
| S4 | M4 | Groupes + events operationnels |
| S5 | M5 | Chat + notifications en temps reel |
| S6 | M6 | Docker + audit mandatory valide |

## 5) Regles de pilotage
- Travailler en blocs courts et testables.
- Un bloc = un objectif, une verification, un commit.
- Toute fonctionnalite doit pointer vers des IDs de docs/01_audit_questions.md.
- Si un point n est pas testable, il n est pas termine.
