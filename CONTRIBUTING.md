# Contribuer à Télé-VERT-Ailleur

Merci de passer par là ! 🟢 Ce projet est un petit utilitaire fait sur un coin
de table : toute contribution est la bienvenue, et pas besoin d'être un·e pro de
Go pour mettre les mains dedans.

## Par où commencer

- 💡 **Une idée, une amélioration ?** Fonce : code-la et **soumets directement
  une _pull request_**, pas besoin d'en discuter avant.
- 🎨 **Envie de bricoler l'interface ?** C'est le terrain de jeu le plus ouvert —
  design, ergonomie, thèmes, nouveaux réglages… Lance-toi, ça me ferait plaisir !
- 🐛 **Un bug ?** Une [issue](../../issues) suffit, ou — encore mieux — une PR qui
  le corrige.

## Proposer une pull request

1. *Fork* le dépôt, puis crée une branche : `git checkout -b ma-super-idee`.
2. Code en gardant le style existant (commentaires en français, noms clairs).
3. Vérifie que tout passe : `./scripts/test.ps1` (go vet + tests).
4. Compile pour tester en vrai : `./scripts/build.ps1`.
5. Ouvre la *pull request* en expliquant le **quoi** et le **pourquoi**.

Pré-requis de build (Go + un compilateur C pour Fyne) : voir la section
[Compiler depuis les sources](README.md#option-2--compiler-depuis-les-sources)
du README.

## Bon à savoir

- Le projet est **Windows uniquement** (APIs Win32 : `SendInput`,
  `SetThreadExecutionState`, registre…). Les fichiers `*_windows.go` portent la
  contrainte de build correspondante.
- Toute la logique vit dans le package `internal/app` ; le point d'entrée est
  dans `cmd/fuckteamsstatus`.

Pas de règle rigide ici : sois sympa, teste ton truc, et amuse-toi. Merci ! 🟢
