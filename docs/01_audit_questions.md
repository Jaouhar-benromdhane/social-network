# Audit Questions Tracker (Mandatory)

Ce fichier centralise toutes les questions de l audit mandatory.

Scope:
- Inclus: mandatory
- Exclu: bonus
- Exclu: optionnel app-image

Statuts recommandes:
- TODO: pas commence
- WIP: en cours
- READY: reponse prete mais a relire
- VALIDATED: verifiee en pratique

Conseil de remplissage:
- Reponse: courte et factuelle
- Preuve: chemin fichier, commande, capture, ou scenario test

## Functional
| ID | Question (copie audit) | Reponse | Preuve | Statut |
|---|---|---|---|---|
| AQ-GEN-001 | Has the requirement for the allowed packages been respected? | Oui pour le scope implemente: stdlib + golang-migrate + sqlite3 + bcrypt + google/uuid. | backend/go.mod | READY |
| AQ-GEN-002 | Open the project. | Le projet ouvre et build localement (backend et frontend). | go test ./... ; npm run build | VALIDATED |
| AQ-GEN-003 | When examining the file system of the backend, did you find a well-organized structure, similar to the example provided in the subject, with a clear separation of packages and migrations folders? | Oui, separation server/app/db/migrations en place. | backend/cmd/server/main.go ; backend/internal/app/app.go ; backend/pkg/db/migrations/sqlite | READY |
| AQ-GEN-004 | Is the file system for the frontend well organized? | Oui, frontend structure framework + src + docker claire. | frontend/package.json ; frontend/src/App.vue ; frontend/Dockerfile | READY |

## Backend
| ID | Question (copie audit) | Reponse | Preuve | Statut |
|---|---|---|---|---|
| AQ-BE-001 | Does the backend include a clear separation of responsibilities among its three major parts - Server, App, and Database? | Oui, separation explicite implementee. | backend/cmd/server/main.go ; backend/internal/app/app.go ; backend/pkg/db/sqlite/sqlite.go | READY |
| AQ-BE-002 | Is there a server that effectively receives incoming requests and serves as the entry point for all requests to the application? | Oui, serveur HTTP Go actif sur APP_PORT avec entree /api/health. | backend/cmd/server/main.go ; curl /api/health | READY |
| AQ-BE-003 | Does the application (App) running on the server effectively listen for requests, retrieve information from the database, and send responses? | Oui, endpoints auth/profile actifs avec lecture/ecriture DB et reponses JSON. | backend/internal/app/auth.go ; backend/internal/app/profile.go | READY |
| AQ-BE-004 | Is the core logic of the social network implemented within the App component, including the logic for handling various types of requests based on HTTP or other protocols? | Base posee, logique sociale complete encore a implementer en M2-M5. | backend/internal/app/app.go | WIP |

## Database
| ID | Question (copie audit) | Reponse | Preuve | Statut |
|---|---|---|---|---|
| AQ-DB-001 | Is SQLite being used in the project as the database? | Oui, SQLite utilise comme DB principale. | backend/pkg/db/sqlite/sqlite.go ; backend/data/social_network.db | READY |
| AQ-DB-002 | Are clients able to request information stored in the database, and can they submit data to be added to it without encountering errors or issues? | Oui pour les flux auth/profil deja testes (insert users, sessions, select profile). | /api/auth/register ; /api/auth/me ; /api/profile/me | VALIDATED |
| AQ-DB-003 | Does the app implement a migration system? | Oui, migrations appliquees automatiquement au demarrage backend. | backend/pkg/db/sqlite/sqlite.go | READY |
| AQ-DB-004 | Is that migration file system well organized? (like the example from the subject) | Oui, dossiers et fichiers up/down organises sous pkg/db/migrations/sqlite. | backend/pkg/db/migrations/sqlite | READY |
| AQ-DB-005 | Start the social network application, then enter the database using the command sqlite3 <database_name.db>. Are the migrations being applied by the migration system? | Oui, verification effectuee avec sqlite3 et tables creees. | sqlite3 backend/data/social_network.db '.tables' | VALIDATED |

## Authentication
| ID | Question (copie audit) | Reponse | Preuve | Statut |
|---|---|---|---|---|
| AQ-AUTH-001 | Does the app implement sessions for the authentication of the users? | Oui, cookie session_token + table sessions + expiration. | backend/internal/app/auth.go ; table sessions | VALIDATED |
| AQ-AUTH-002 | Are the correct form elements being used in the registration? (Email, Password, First Name, Last Name, Date of Birth, Avatar/Image (Optional), Nickname (Optional), About Me (Optional)) | Oui, tous les champs sont presents dans le formulaire frontend. | frontend/src/App.vue | READY |
| AQ-AUTH-003 | Try to register a user. During registration, when attempting to register a user, did the application correctly save the registered user to the database without any errors? | Oui, inscription OK (201) et user recuperable via session. | test curl REGISTER_U1/U2 sur /api/auth/register | VALIDATED |
| AQ-AUTH-004 | Try to log in with the user you just registered. When attempting to log in with the user you just registered, did the login process work without any problems? | Oui, login OK renvoie 200 et session cookie active. | test LOGIN_OK sur /api/auth/login | VALIDATED |
| AQ-AUTH-005 | Try to log in with the user you created, but with a wrong password or email. Did the application correctly detect and respond to the incorrect login details? | Oui, login invalide renvoie 401. | test LOGIN_WRONG_PASSWORD sur /api/auth/login | VALIDATED |
| AQ-AUTH-006 | Try to register the same user you already registered. Did the app detect if the email/user is already present in the database? | Oui, email duplique renvoie 409. | test REGISTER_DUPLICATE_EMAIL sur /api/auth/register | VALIDATED |
| AQ-AUTH-007 | Open two browsers (ex: Chrome and Firefox), log in into one and refresh the other browsers. Can you confirm that the browser non logged remains unregistered? | Oui, sans cookie actif /api/auth/me renvoie 401. | test ME_NO_COOKIE sur /api/auth/me | VALIDATED |
| AQ-AUTH-008 | Using the two browsers, log in with different users in each one. Then refresh both the browsers. Can you confirm that both browsers continue with the right users? | Oui, deux cookie jars distincts gardent deux sessions utilisateur differentes. | tests ME_U1 et ME_U2 sur /api/auth/me | VALIDATED |

## Followers
| ID | Question (copie audit) | Reponse | Preuve | Statut |
|---|---|---|---|---|
| AQ-FOL-001 | Try to follow a private user. Are you able to send a following request to the private user? | TODO | TODO | TODO |
| AQ-FOL-002 | Try to follow a public user. Are you able to follow the public user without the need of sending a following request? | TODO | TODO | TODO |
| AQ-FOL-003 | Open two browsers (ex: Chrome and Firefox), log in as two different private users and with one of them try to follow the other. Is the user who received the request able to accept or decline the following request? | TODO | TODO | TODO |
| AQ-FOL-004 | After following another user successfully try to unfollow him. Were you able to do so? | TODO | TODO | TODO |

## Profile
| ID | Question (copie audit) | Reponse | Preuve | Statut |
|---|---|---|---|---|
| AQ-PROF-001 | Try opening your own profile. Does the profile displays every information requested in the register form, apart from the password? | Oui pour les donnees profil exposees (sans mot de passe). | /api/profile/me ; frontend/src/App.vue | READY |
| AQ-PROF-002 | Try opening your own profile. Does the profile displays every post created by the user? | TODO | TODO | TODO |
| AQ-PROF-003 | Try opening your own profile. Does the profile displays the users that you follow and the ones who are following you? | TODO | TODO | TODO |
| AQ-PROF-004 | Try opening your own profile. Are you able to change between private profile and public profile? | Oui, endpoint de mise a jour visibilite operationnel. | test PATCH_VISIBILITY sur /api/profile/me/visibility | VALIDATED |
| AQ-PROF-005 | Open two browsers and log in with different users on them, with one of the users having a private profile and successfully follow that user. Are you able to see a followed user private profile? | TODO | TODO | TODO |
| AQ-PROF-006 | Using the two browsers with the same users, with one of the users having a private profile and be sure not to follow him. Are you prevented from seeing a non-followed user private profile? | TODO | TODO | TODO |
| AQ-PROF-007 | Using the two browsers with the users, with one of the users having a public profile and be sure not to follow him. Are you able to see a non-followed user public profile? | TODO | TODO | TODO |
| AQ-PROF-008 | Using the two browsers with the users, with one of the users having a public profile and successfully follow that user. Are you able to see a followed user public profile? | TODO | TODO | TODO |

## Posts
| ID | Question (copie audit) | Reponse | Preuve | Statut |
|---|---|---|---|---|
| AQ-POST-001 | Are you able to create a post and comment on already existing posts after logging in? | TODO | TODO | TODO |
| AQ-POST-002 | Try creating a post. Are you able to include an image (JPG or PNG) or a GIF on it? | TODO | TODO | TODO |
| AQ-POST-003 | Try creating a comment. Are you able to include an image (JPG or PNG) or a GIF on it? | TODO | TODO | TODO |
| AQ-POST-004 | Try creating a post. Can you specify the type of privacy of the post (public, almost private, private)? | TODO | TODO | TODO |
| AQ-POST-005 | If you choose the private privacy option, can you specify the users that are allowed to see the post? | TODO | TODO | TODO |

## Groups
| ID | Question (copie audit) | Reponse | Preuve | Statut |
|---|---|---|---|---|
| AQ-GRP-001 | Try creating a group. | TODO | TODO | TODO |
| AQ-GRP-002 | Were you able to invite one of your followers to join the group? | TODO | TODO | TODO |
| AQ-GRP-003 | Open two browsers, log in with different users on each browser, follow each other and with one of the users create a group and invite the other user. Did the other user received a group invitation that he/she can refuse/accept? | TODO | TODO | TODO |
| AQ-GRP-004 | Using the same browsers and the same users, with one of the users create a group and with the other try to make a group entering request. Did the owner of the group received a request that he/she can refuse/accept? | TODO | TODO | TODO |
| AQ-GRP-005 | Can a user make group invitations, after being part of the group (being the user different from the creator of the group)? | TODO | TODO | TODO |
| AQ-GRP-006 | Can a user make a group entering request (a request to enter a group)? | TODO | TODO | TODO |
| AQ-GRP-007 | After being part of a group, can the user create posts and comment already created posts? | TODO | TODO | TODO |
| AQ-GRP-008 | Try to create an event in a group. Were you asked for a title, a description, a day/time and at least two options (going, not going)? | TODO | TODO | TODO |
| AQ-GRP-009 | Using the same browsers and the same users, after both of them becoming part of the same group, create an event with one of them. Is the other user able to see the event and vote in which option he wants? | TODO | TODO | TODO |

## Chat
| ID | Question (copie audit) | Reponse | Preuve | Statut |
|---|---|---|---|---|
| AQ-CHAT-001 | Try and open two browsers (ex: Chrome and Firefox), log in with different users in each one. Then with one of the users try to send a private message to the other user. Did the other user received the message in realtime? | TODO | TODO | TODO |
| AQ-CHAT-002 | Try and open two browsers (ex: Chrome and Firefox), log in with different users that are not following each other at all. Then with one of the users try to send a private message to the other user. Can you confirm that it was not possible to create a chat between these two users? | TODO | TODO | TODO |
| AQ-CHAT-003 | Using the two browsers with the users start a chat between the two of them. Did the chat between the users went well? (did not crash the server) | TODO | TODO | TODO |
| AQ-CHAT-004 | Try and open three browsers (ex: Chrome and Firefox or a private browser), log in with different users in each one. Then with one of the users try to send a private message to one of the other users. Did only the targeted user received the message? | TODO | TODO | TODO |
| AQ-CHAT-005 | Using the three browsers with the users, enter with each user a common group. Then start sending messages to the common chat room using one of the users. Did all the users that are common to the group receive the message in realtime? | TODO | TODO | TODO |
| AQ-CHAT-006 | Using the three browsers with the users, continue chatting between the users in the group. Did the chat between the users went well? (did not crash the server) | TODO | TODO | TODO |
| AQ-CHAT-007 | Can you confirm that it is possible to send emojis via chat to other users? | TODO | TODO | TODO |

## Notifications
| ID | Question (copie audit) | Reponse | Preuve | Statut |
|---|---|---|---|---|
| AQ-NOTIF-001 | Can you check the notifications on every page of the project? | TODO | TODO | TODO |
| AQ-NOTIF-002 | Open two browsers, log in as two different private users and with one of them try to follow the other. Did the other user received a notification regarding the following request? | TODO | TODO | TODO |
| AQ-NOTIF-003 | Open two browsers, log in with different users on each browser, follow each other and with one of the users create a group and invite the other user. Did the invited user received a notification regarding the group invitation request? | TODO | TODO | TODO |
| AQ-NOTIF-004 | Open two browsers, log in with different users on each browser, create a group with one of them and with the other send a group entering request. Did the other user received a notification regarding the group entering request? | TODO | TODO | TODO |
| AQ-NOTIF-005 | Open two browsers, log in with different users on each browser, become part of the same group with both users and with one of the users create an event. Did the other user received a notification regarding the creation of the event? | TODO | TODO | TODO |

## Docker
| ID | Question (copie audit) | Reponse | Preuve | Statut |
|---|---|---|---|---|
| AQ-DOCK-001 | Try to run the application and use the docker command docker ps -a. Can you confirm that there are two containers (backend and frontend), and both containers have non-zero sizes indicating that they are not empty? | Configuration docker prete, execution containers a valider pendant M6. | docker-compose.yml ; backend/Dockerfile ; frontend/Dockerfile | WIP |
| AQ-DOCK-002 | Try to access the social network application through your web browser. Were you able to access the social network application through your web browser after running the docker containers, confirming that the containers are running and serving the application as expected? | Pas encore execute en local, prevu en M6. | docker-compose.yml | TODO |

## Check final avant soutenance
- Toutes les lignes ont un statut READY ou VALIDATED.
- Chaque ligne a une preuve concrete.
- Les scenarios multi navigateurs ont ete repasses.
- Les points faibles connus sont notes clairement.
