// Commande fuckteamsstatus : point d'entrée de l'application. Toute la logique
// vit dans le package internal/app ; ce main se contente de la démarrer.
package main

import "fuckteamsstatus/internal/app"

func main() {
	app.Run()
}
