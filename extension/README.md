# Extension Chrome de Thalamus

Un bouton « proposer » là où l'employé lit. Le popup envoie la page, ou la sélection, dans un cerveau où l'employé peut écrire, par `POST /v1/add` du gateway, sous sa propre clé. Le document attend un curateur ; un curateur voit le sien servi aussitôt. Manifest V3, sans bibliothèque.

## Fichiers

- `manifest.json` : permissions `storage`, `activeTab`, `scripting`, `tabs`, `identity` (la connexion), hôtes `http` et `https` (le gateway et la page), icônes de l'extension et du bouton.
- `icons/` : `icon16.png`, `icon32.png`, `icon48.png`, `icon128.png` (deux lobes clairs sur fond indigo), dessinées par `make_icons.py` sans bibliothèque (`python3 -I extension/icons/make_icons.py`). `manifest.test.js` vérifie que chaque icône nommée existe, est un PNG et a sa taille.
- `lib.js` : les règles, sans DOM ni `chrome.*`, couvertes par `node --test` (`lib.test.js`) : `writableBrains` (cerveaux `write` ou `curate` avec une clé, triés), `documentID` (identifiant stable par page, fragment ignoré : reproposer la même page met à jour le même document), `buildAdd` (la sélection sinon la page, espaces repliés, titre à 200 caractères, contenu à 100 000, métadonnées `via: chrome-extension`, `kind: selection|page`, et pour une page `extract: landmark|blocks|body`), `mainText` (le contenu principal de la page : le plus long `article` d'au moins 200 caractères, sinon le plus long `main` ; sinon le groupe de blocs de texte qui a le plus de texte hors liens, une classe ou un id de contenu comptant moitié plus, avec ses groupes frères, sans les groupes dont la classe, l'id ou la balise dit navigation, pied de page, commentaires, partage ; sinon tout le texte du `body`), `statusText` (une phrase par statut HTTP), `settingsOf` (adresse sans barre finale, clé, les deux requis), `authorizeURL` (adresse de connexion du fournisseur, flux implicite `id_token` en fragment), `tokenOf` (lit l'`id_token` de la redirection, vérifie `state` et `nonce`, dit un refus), `signInText`, `randomString`.
- `options.html`, `options.js` : adresse du gateway et clé, dans `chrome.storage.local`. Bouton « Sign in » : `GET /v1/identity` du gateway, discovery du fournisseur, `chrome.identity.launchWebAuthFlow`, `POST /v1/login`, la clé rendue est enregistrée (`#signed`, `data-status` : `no-gateway`, `refused`, `failed`, `signed`). Sans fournisseur sur le gateway, la clé se colle à la main.
- `popup.html`, `popup.js` : liste des cerveaux (`GET /v1/brains` filtré), titre de l'onglet, texte de la sélection ou contenu principal de la page (`chrome.scripting.executeScript` collecte seulement : repères, blocs avec leur parent, leur grand-parent, leurs indices et la longueur de leurs liens ; `mainText` choisit ; `#content` porte `data-extract`), bouton « Propose », phrase d'état dans `#status` (`data-status` : `no-settings`, `no-page`, `no-text`, `brains-refused`, `no-gateway`, `no-brain`, `invalid`, `sent`, `refused`).
- `background.js` : vide ; il donne un identifiant à l'extension pour le banc.
- `e2e_driver.mjs` : le pilote Playwright du banc, ici pour que `node` trouve `playwright` dans `node_modules`.

## Installer

1. `chrome://extensions`, mode développeur, « Charger l'extension non empaquetée », ce dossier.
2. Options : l'adresse du gateway, puis « Sign in » avec le compte de l'organisation. Sans fournisseur d'identité, coller la clé que `th-gateway key issue` (ou `key sync`) a donnée à l'employé.
3. Sur une page : cliquer l'icône, choisir le cerveau, corriger le titre, « Propose ».

## Vérifier

```sh
cd extension && npm install && node --test     # 14 tests : les règles (lib.test.js) et les icônes (manifest.test.js)
cd gateway && python3 -I mutate.py            # les mutations de lib.js en font partie (paquet « extension »)
sh tests/e2e_extension.sh                     # Chromium (Playwright 1.58.2) charge l'extension devant un vrai gateway et un faux fournisseur (tests/fake_idp.mjs), 13 contrôles, environ 30 s
```

Le banc : sans réglages le popup renvoie aux options ; une mauvaise clé est dite ; la connexion par le fournisseur enregistre la clé que le gateway avait émise ; la clé liste le seul cerveau où lfoletti écrit ; la page est proposée sans sa navigation, son encart de partage ni son pied de page (`extract: blocks`), puis la sélection seule ; l'audit porte deux `th_add` pour le même identifiant ; `th_get` rend la sélection, proposée, par lfoletti, avec la page comme source.

## Limites

- Pas de lecture ni de rappel depuis l'extension : elle propose, rien d'autre.
- L'extraction est une heuristique sans bibliothèque : une page sans `article` ni `main`, dont le contenu est éclaté en blocs sous des parents sans lien de parenté, ne garde que le plus gros groupe. Le popup montre le texte avant l'envoi, et l'employé le corrige ou choisit une sélection.
- Le contenu reste coupé à 100 000 caractères, espaces repliés : les sauts de paragraphe ne sont pas gardés.
