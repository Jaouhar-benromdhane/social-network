# Decisions Log

Utilise ce journal pour garder la trace des choix importants.

## Decision Template
- Date:
- Contexte:
- Options considerees:
- Decision retenue:
- Pourquoi:
- Impact:

---

## D-001 - Mise en place d un suivi documentaire
- Date: 2026-03-18
- Contexte: Besoin de progresser pas a pas et de pouvoir reprendre facilement.
- Options considerees: 1) notes libres 2) docs structures
- Decision retenue: Utiliser README + docs de suivi dedies.
- Pourquoi: Plus simple pour suivre la progression et preparer la soutenance.
- Impact: Chaque session doit mettre a jour les docs.

## D-002 - Scope principal = mandatory sans bonus
- Date: 2026-03-18
- Contexte: Audit mandatory volumineux, besoin de securiser la note avant toute extension.
- Options considerees: 1) mandatory + bonus 2) mandatory strict
- Decision retenue: Faire mandatory strict, bonus exclus.
- Pourquoi: Minimiser le risque de dispersion et garantir une couverture complete des points evalues.
- Impact: Toute priorisation de travail suit la checklist mandatory.

## D-003 - App image en piste optionnelle separee
- Date: 2026-03-18
- Contexte: Nouveau sujet optionnel app-image (Electron) recu.
- Options considerees: 1) melanger avec mandatory 2) piste separee apres mandatory
- Decision retenue: Traiter app-image comme Milestone M7, apres validation mandatory.
- Pourquoi: Eviter les retards sur le perimetre principal note.
- Impact: Le suivi app-image est documente a part et n impacte pas les commits mandatory.

## D-004 - Scope final verrouille: mandatory uniquement
- Date: 2026-03-18
- Contexte: Priorite absolue au passage de l audit mandatory et au fonctionnement stable.
- Options considerees: 1) mandatory + optionnel 2) mandatory strict
- Decision retenue: Mandatory strict uniquement. Aucun bonus, aucun app-image.
- Pourquoi: Concentration totale sur la validation de l audit et la robustesse du projet principal.
- Impact: Toute tache non mandatory est retiree du plan et des commits.

## D-005 - Stack technique M1
- Date: 2026-03-18
- Contexte: Besoin de lancer rapidement une base propre et testable pour l audit.
- Options considerees: plusieurs frameworks frontend et structures backend.
- Decision retenue: Frontend Vue + Vite, backend Go, SQLite + golang-migrate, docker-compose 2 services.
- Pourquoi: setup rapide, lisible, compatible avec les contraintes mandatory.
- Impact: Les prochaines features seront implementees sur cette base.

## D-006 - Auth mandatory via sessions et cookies
- Date: 2026-03-18
- Contexte: Besoin de valider rapidement tout le bloc Authentication de l audit.
- Options considerees: token stateless vs sessions cote serveur.
- Decision retenue: Sessions server-side en base SQLite + cookie HttpOnly session_token.
- Pourquoi: Conforme au sujet (sessions/cookies) et simple a verifier en audit multi navigateur.
- Impact: Tous les endpoints prives passent par lecture de cookie de session active.

## D-007 - Avatar au register
- Date: 2026-03-18
- Contexte: Le formulaire impose Avatar/Image optionnel.
- Options considerees: stockage binaire DB vs stockage fichier + chemin DB.
- Decision retenue: Stockage fichier dans UPLOAD_DIR et chemin public en base.
- Pourquoi: Plus simple, performant et conforme au sujet.
- Impact: Endpoint /uploads expose les avatars, validation MIME JPEG/PNG/GIF.

## D-008 - Strategie followers et profils prives
- Date: 2026-03-18
- Contexte: Besoin de passer les scenarios audit followers + visibilite profil.
- Options considerees: logique permissive simple vs workflow request/accept/decline strict.
- Decision retenue: Profil public => follow direct. Profil prive => request pending puis accept/decline.
- Pourquoi: Repond exactement aux exigences du sujet et facilite les tests multi navigateurs.
- Impact: Endpoints dedies /api/follows/* et /api/profile/view avec controle d acces.

## D-009 - Privacy des posts et pieces jointes media
- Date: 2026-03-18
- Contexte: Les lignes AQ-POST imposent creation de posts/commentaires, media image/gif et regles de visibilite fines.
- Options considerees: 1) filtrage uniquement SQL 2) controle applicatif centralise par regle de privacy.
- Decision retenue: Controle applicatif centralise avec 3 modes (public, almost_private, private) + table post_allowed_users pour le mode private.
- Pourquoi: Plus lisible, plus facile a auditer et a faire evoluer avec les regles followers deja en place.
- Impact: Nouveaux endpoints /api/posts* ; UI feed/commentaires ; private restreint aux followers selectionnes.
