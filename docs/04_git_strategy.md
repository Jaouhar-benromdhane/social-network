# Git Strategy

## 1) Principe
Un commit par grosse partie coherente. Eviter les commits melanges.

## 2) Convention de messages (simple)
Format:
<type>: <resume court>

Types conseilles:
- feat: nouvelle fonctionnalite
- fix: correction bug
- refactor: reorganisation sans changer le comportement
- docs: documentation
- test: ajout/modif de tests
- chore: tache technique annexe

Exemples:
- docs: add initial project tracking documents
- feat: implement user registration flow
- fix: handle empty login payload

## 3) Checklist avant commit
- Le code passe localement.
- Les fichiers inutiles sont exclus.
- Le message explique clairement le changement.
- Le tracker audit est mis a jour si besoin.

## 4) Cycle recommande
1. Pull/rebase de securite.
2. Developpement d un bloc.
3. Verification locale.
4. Mise a jour docs.
5. Commit.
6. Push.

## 5) Commandes utiles
- git status
- git add -A
- git commit -m "type: resume"
- git push
