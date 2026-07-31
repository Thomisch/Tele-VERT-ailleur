# Télé-VERT-ailleur

> Garde ta pastille Microsoft Teams au **vert** sans effort. 🟢

Petit utilitaire **Windows** (Go + [Fyne](https://fyne.io)) qui maintient ta
présence Teams active en simulant une micro-activité clavier/souris inoffensive.
Fenêtre à onglets soignée, icône dans la barre des tâches, état persistant —
avec plages horaires, anti-veille et randomisation « humaine ».

`Windows` · `Go 1.25+` · `Licence MIT`

## Comment ça marche

Teams te passe en « Absent » après quelques minutes **d'inactivité d'entrée**
(souris/clavier). L'appli combine **deux techniques** toutes les ~60 s (ou à
intervalle randomisé si le mode « comportement humain » est actif) :

1. **Touche F15** — une touche qui existe mais n'a aucun effet visible (ne tape
   rien, ne change pas de fenêtre).
2. **Micro-mouvement souris** — un déplacement minime aller-retour, invisible
   (ou plusieurs petits mouvements variés en mode « comportement humain »).

Les deux cumulées rendent l'astuce quasi infaillible : si une appli ignore le
clavier, le mouvement souris la rattrape, et inversement. Par défaut, **aucun
Alt+Tab** (pas de vol de focus pendant que tu travailles) — un mode Alt+Tab
optionnel existe dans l'onglet « Avancé » si tu le veux.

## Prise en main

### Option 1 — Télécharger le binaire (recommandé)

1. Va dans l'onglet **[Releases](../../releases)** du dépôt et télécharge
   `TeleVERTailleur.exe` (dernière version).
2. Double-clique dessus. Windows SmartScreen peut afficher un avertissement
   (binaire non signé) : **Informations complémentaires → Exécuter quand même**.
3. Une **icône verte** apparaît près de l'horloge : c'est actif. 🟢

Aucune installation requise (ni Go, ni .NET) — c'est un exécutable autonome.

### Option 2 — Compiler depuis les sources

**Prérequis** : Go 1.25+ **et un compilateur C** (Fyne lie OpenGL/GLFW en cgo).
Sous Windows, installer MinGW, par exemple via winget :

```powershell
winget install BrechtSanders.WinLibs.POSIX.UCRT
```

Puis compiler avec le script fourni (localise gcc, active cgo, supprime la
console) :

```powershell
./scripts/build.ps1
```

> Si l'exécution du script est bloquée par la politique PowerShell :
> `powershell -ExecutionPolicy Bypass -File scripts/build.ps1`

Le binaire est produit dans `bin/TeleVERTailleur.exe`. Pour lancer les tests :

```powershell
./scripts/test.ps1
```

Build manuel équivalent, si tu préfères :

```powershell
$env:CGO_ENABLED = "1"
go build -ldflags "-H=windowsgui -s -w" -o bin/TeleVERTailleur.exe ./cmd/fuckteamsstatus
```

## Interface

Une fenêtre à onglets s'ouvre, avec un **bandeau d'état** en haut (vert =
« Actif », gris = « En pause »).

### Onglet « État »

- **Maintien actif** — interrupteur principal du keep-alive.
- **Lancer au démarrage de Windows** — ajoute/retire l'appli du démarrage de
  session (clé de registre `HKCU\…\Run`, chemin auto-corrigé si l'exe est
  déplacé).
- **Empêcher la mise en veille de l'ordinateur** — tant que le maintien tourne,
  l'ordinateur ne se met pas en veille (API `SetThreadExecutionState`). L'écran
  peut toujours s'éteindre. L'option suit l'état du maintien : en pause,
  l'anti-veille est relâchée.

### Onglet « Planning » (mode planifié)

Déclare des **plages horaires** pendant lesquelles le maintien s'active
automatiquement.

- **Activer le mode planifié** — quand il est ON, l'horaire pilote tout : le
  maintien tourne pendant les plages et reste coupé en dehors. Le toggle manuel
  de l'onglet « État » devient alors indicatif (grisé).
- Chaque plage : heure **De / à** (`HH:MM`) + jours cochables
  (**L M M J V S D**), et un **switch d'activation** par plage. Bornes en
  intervalle `[début, fin[`.
- **Ajouter / supprimer** des plages librement.
- Mode planifié désactivé → retour au contrôle manuel pur.

### Onglet « Avancé » (furtivité)

- **Imiter un comportement humain** — randomise l'intervalle de réveil
  (≈ 40–95 s au lieu de 60 s pile) et les mouvements de souris (plusieurs
  micro-déplacements d'amplitude et direction variables, sans dériver le
  curseur). Rend l'activité simulée moins régulière, donc moins repérable.
- **Mode Alt+Tab** — en plus du maintien (cumulable), bascule de fenêtre toutes
  les X secondes. Intervalle **réglable** avec variation aléatoire, plancher de
  sécurité **5 s** (pour toujours pouvoir reprendre la main). Volontairement
  visible — à n'activer que quand tu n'utilises pas la machine.

### Icône de la barre des tâches

L'icône (près de l'horloge) reflète l'état (verte/grise). Clic droit :

- **Ouvrir** — réaffiche la fenêtre
- **Activer / Désactiver** — bascule le keep-alive
- **Quitter** — ferme réellement l'application

### Comportements

- **Fermer la fenêtre (croix)** la réduit dans le tray ; l'appli continue de
  tourner. Pour quitter pour de bon : menu tray → **Quitter**.
- **Instance unique** : relancer l'exe alors qu'une instance tourne déjà ne crée
  pas de doublon — ça ramène la fenêtre existante au premier plan.
- **État persistant** : tout est sauvegardé et rechargé au lancement suivant.

## Configuration

L'état est stocké dans :

```
%APPDATA%\FuckTeamsStatus\config.json
```

(le dossier conserve le nom de code interne du projet)

```json
{
  "keepAliveEnabled": true,
  "launchAtStartup": false,
  "preventSleep": false,
  "scheduleEnabled": true,
  "schedules": [
    { "start": "09:00", "end": "12:00", "days": [1, 2, 3, 4, 5] },
    { "start": "13:00", "end": "17:30", "days": [1, 2, 3, 4, 5], "disabled": true }
  ],
  "humanize": false,
  "altTabEnabled": false,
  "altTabSeconds": 10
}
```

- `days` suit la convention `time.Weekday` : 0 = dimanche, 1 = lundi … 6 = samedi.
- `disabled` (optionnel) désactive une plage sans la supprimer ; absent = active.
- `humanize`, `altTabEnabled`, `altTabSeconds` correspondent à l'onglet Avancé.

Le fichier tolère un BOM UTF-8 en tête (ajouté par certains éditeurs).

## Réglages fins

L'intervalle de réveil (mode non randomisé) est `keepAliveInterval` dans
[`internal/app/engine.go`](internal/app/engine.go) — 60 s par défaut, largement
sous le seuil d'inactivité de Teams. Les bornes de randomisation et le plancher
Alt+Tab sont dans [`internal/app/randomize.go`](internal/app/randomize.go).

## Structure du projet

```
.
├── cmd/fuckteamsstatus/     Point d'entrée (main → app.Run)
├── internal/app/            Toute la logique applicative (package app)
├── scripts/                 build.ps1, test.ps1, test-f15.ps1
├── go.mod / go.sum
└── README.md
```

Le code applicatif est un package unique `app` dans `internal/app/`. Les
fichiers suffixés `_windows.go` ne compilent que sous Windows (contraintes de
build). Principaux fichiers :

| Fichier | Rôle |
|---|---|
| `run_windows.go` | Câblage config/engine/scheduler/UI, exporte `Run()` |
| `ui.go` | Fenêtre Fyne à onglets (bandeau, onglet État) + icône tray |
| `ui_schedule_windows.go` | Onglet Planning : éditeur de plages horaires |
| `ui_advanced_windows.go` | Onglet Avancé : randomisation + Alt+Tab |
| `toggle_windows.go` | Widget switch on/off réutilisable |
| `engine.go` | Moteur keep-alive thread-safe + notifications |
| `scheduler.go` | Mode planifié : pilote l'engine selon les plages |
| `schedule.go` | Logique des plages (parsing HH:MM, bornes, jours) |
| `randomize.go` | Intervalles aléatoires + mouvements souris humanisés |
| `input_windows.go` | F15 + mouvements souris via `SendInput` |
| `alttab_windows.go` | Boucle Alt+Tab périodique |
| `sleep_windows.go` | Anti-veille via `SetThreadExecutionState` |
| `config.go` | Lecture/écriture du `config.json` (tolérant au BOM) |
| `autostart_windows.go` | Démarrage Windows (registre Run, auto-correction) |
| `single_windows.go` | Instance unique (mutex + event nommés) |
| `icons_windows.go` | Icônes tray générées en mémoire (PNG) |
| `*_test.go` | Tests : config, moteur, plages, scheduler, anti-veille, randomisation |

## Avertissement

Outil fourni **tel quel**, à des fins éducatives et personnelles. Contourner une
mesure de présence peut aller à l'encontre du règlement de ton employeur —
utilise-le en connaissance de cause, tu en es seul responsable.

## Licence

Distribué sous licence **MIT** — voir le fichier [LICENSE](LICENSE).
