# App Image (Optional) - Electron Track

## 1) Objectif
Construire une app desktop messenger type Facebook/Discord, multi plateforme (Windows, Linux, macOS), connectee au backend social-network.

## 2) Scope fonctionnel attendu
- Presence en ligne (online/offline) en temps reel
- Notification lors de reception de message
- Chat temps reel entre utilisateurs
- Envoi d emojis
- Mode offline: lecture des messages possible, envoi/reception bloques avec message explicite
- Recherche interactive de messages (resultats pendant la saisie)

## 3) Contraintes
- Authentification obligatoire dans l app
- Formulaire login: email + password
- Si utilisateur non inscrit: redirection vers le site social-network pour inscription
- Session persistante entre redemarrages de l app jusqu a expiration/logout
- Websocket obligatoire
- Bonne pratiques de code

## 4) Points websocket obligatoires
- Message envoye depuis le site visible en temps reel dans l app et inversement
- Changement online/offline propage en temps reel aux followers

## 5) Mode offline (criteres)
- Detection perte internet
- Message clair indiquant l etat offline
- Historique visible
- Tentative d envoi refusee avec message d erreur clair

## 6) Checklist optionnelle (a valider si M7 lance)
| ID | Exigence | Reponse | Preuve | Statut |
|---|---|---|---|---|
| APP-001 | Login email/password fonctionnel | TODO | TODO | TODO |
| APP-002 | Redirection vers website si non inscrit | TODO | TODO | TODO |
| APP-003 | Session persistante apres redemarrage app | TODO | TODO | TODO |
| APP-004 | Presence online/offline temps reel | TODO | TODO | TODO |
| APP-005 | Notification reception message | TODO | TODO | TODO |
| APP-006 | Chat temps reel app <-> website | TODO | TODO | TODO |
| APP-007 | Emojis envoyables et affichables | TODO | TODO | TODO |
| APP-008 | Mode offline detecte et affiche | TODO | TODO | TODO |
| APP-009 | Envoi bloque en offline avec message erreur | TODO | TODO | TODO |
| APP-010 | Recherche interactive sans bouton | TODO | TODO | TODO |

## 7) Plan de realisation (si active)
1. Initialiser shell Electron + auth flow.
2. Connecter websocket et presence.
3. Integrer chat prive + notifications.
4. Ajouter offline mode + persistance locale.
5. Ajouter recherche interactive.
6. Passer checklist APP-001 a APP-010.

## 8) Decision actuelle
M7 optionnel uniquement, apres passage complet du mandatory.
