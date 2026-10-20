# Lot B, étape 1 : la proposition au propriétaire

Forme retenue : **un commentaire sur la discussion #66**, pas une nouvelle discussion. Le point 2 de #66 annonce déjà la proposition et demande l'avis du propriétaire. Une seconde discussion ferait doublon. Si le propriétaire préfère un fil séparé, le même texte sert tel quel.

État : **à publier par Lauro, après lecture**. Commande, depuis la racine du projet :

```
gh api graphql -f query='mutation($d:ID!, $body:String!) { addDiscussionComment(input:{discussionId:$d, body:$body}) { comment { url } } }' -f d=D_kwDOPZvEVs4AqM6k -f body="$(cat docs/pr/lot_B/00_proposal_comment.body.md)"
```

Attente : 7 jours au plus. Puis GC-03 part dans tous les cas. GC-04 part en PR brouillon si aucune réponse, en PR normale sur accord de principe, et reste dans le fork (patch temporaire) sur refus.
