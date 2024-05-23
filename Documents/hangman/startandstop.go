package lettre

import (
	"encoding/json"
	"flag"
	"fmt"
	"hangman/jose"
	"math/rand"
	"os"
	"time"
)

type Stop struct { //création de la struct qu'on utilisera pour l'encodage et le decodage via json, elle prend comme champ le nombres d'essaies restants, les lettres déja trouvées et le mot à trouver
	Attemptleft int
	LettresFind []rune
	Mot         []rune
}

func Lettre(motrandom string) {
	Lettrestrouves := []rune{}
	Motatrouver := []rune{}
	lettrerandom_rune := []rune(motrandom)
	var reveal []rune
	b := 0
	var position []int
	// Choisit le nombre de lettre à révéler
	n := len(lettrerandom_rune)/2 - 1
	// Choisit des lettres à révéler au hasard et fais en sorte de ne pas tomber sur la meme
	for i := 0; i < n; i++ {
		b = 0
		a := rand.Intn(len(lettrerandom_rune))
		for j := 0; j < len(reveal); j++ {
			if lettrerandom_rune[a] == reveal[j] {
				b = b + 1
			}
		}
		if b > 0 {
			i--
		} else {
			reveal = append(reveal, lettrerandom_rune[a])
			position = append(position, a)
		}
	}
	var total []rune
	var count int

	//Révèle les lettres en plusieurs exemplaires
	// si l'une de celle choisie au début est contenue plusieurs fois dans le mot

	for i := 0; i < len(reveal); i++ {
		for j := 0; j < len(lettrerandom_rune); j++ {
			if lettrerandom_rune[j] == reveal[i] {
				count = count + 1
			}
		}
		for k := 0; k < count; k++ {
			total = append(total, reveal[i])
		}

		count = 0
	}

	//Placer les lettres aux bons endroits
	var motjeu []rune
	underscore := '_'
	space := ' '
	var count2 int
	for k := 0; k < len(lettrerandom_rune); k++ {
		count2 = 0
		for m := 0; m < len(total); m++ {
			if lettrerandom_rune[k] == total[m] {
				count2 = count2 + 1

			}
		}
		if count2 == 0 {
			motjeu = append(motjeu, underscore)
			motjeu = append(motjeu, space)
		} else {
			motjeu = append(motjeu, lettrerandom_rune[k])
			motjeu = append(motjeu, space)
		}
	}

	attempts := 10
	var lettre string
	var count3 int
	var count4 bool

	// Si on lance le jeu en mode sauvegarde

	startWith := flag.String("startWith", "", "spécifier le fichier avec lequel démarrer") //on crée un flag de type string et de valeur "startWith"
	flag.Parse()                                                                           //analyse des arguments de la ligne de commande fournis au programme
	_, erreur2 := os.ReadFile(*startWith)                                                  // lecture du fichier lors de l'utilisation de startWith
	if erreur2 == nil {                                                                    //Si la commande a été utilisé
		Stop1 := Stop{
			Attemptleft: attempts,
			LettresFind: motjeu,
			Mot:         lettrerandom_rune,
		}
		flag.Parse()
		Lecture, erreur2 := os.ReadFile(*startWith) //ouverture et lecture du fichier save.txt
		if erreur2 == nil {
			erreur3 := json.Unmarshal(Lecture, &Stop1) //decodage des elements sauvegardés
			if erreur3 != nil {
				fmt.Println("erreur:", erreur3)
			}
			attempts = Stop1.Attemptleft //on recupere le nombre d'essaies restants qui étaient dans save.txt
			fmt.Println("Welcome back il vous reste", attempts, "essaies")
			for i := 0; i < len(Stop1.LettresFind); i++ { //conversion de la string en tableau de rune
				Lettrestrouves = append(Lettrestrouves, rune(Stop1.LettresFind[i]))
			}
			motjeu = Lettrestrouves               //on recupere les lettres déja trouvé de la partie sauvegardée
			for i := 0; i < len(Stop1.Mot); i++ { //conversion de la string en tableau de rune
				Motatrouver = append(Motatrouver, rune(Stop1.Mot[i]))
			}
			lettrerandom_rune = Motatrouver //on recupere le mot a trouver de la partie sauvegardée
			fmt.Println(string(motjeu))
		}

	} else {
		fmt.Println("Bonne chance, vous avez 10 essais")
		fmt.Println(string(motjeu))
	}
	for {

		// Scan l'input du joueur

		count4 = false
		fmt.Println("Veuillez taper une lettre", "il vous reste", attempts, "essais")
		fmt.Println("Tapez Stop pour sauvegarder la partie et quitter")
		fmt.Scanf("%s", &lettre)
		lettre2 := []rune(lettre)

		// Si le joueur écrit plus d'un caractère

		if len(lettre2) > 1 {

			// Si le joueur écrit "stop" cela lance le processus de sauvegarde de sa partie

			if string(lettre2) == "stop" || string(lettre2) == "Stop" {
				fmt.Println("Sauvegarde en cours...")
				time.Sleep(3 * time.Second)
				Stop1 := Stop{ //on assigne au valeur de la struct les valeurs de la partie actuelle afin des les sauvegarder dans save.txt
					Attemptleft: attempts,
					LettresFind: motjeu,
					Mot:         lettrerandom_rune,
				}
				Save, erreur := json.Marshal(Stop1) //encodage des élements de Stop1
				if erreur != nil {
					fmt.Println("Erreur lors de l'encodage", erreur)
					return
				}

				fichier, err := os.Create("save.txt") //création du fichier save.txt
				if err != nil {
					fmt.Println("Erreur lors de la creation du fichier", err)
				}
				_, errr := fichier.Write(Save) //ecriture des données dans le fichier save.txt
				if errr != nil {
					fmt.Println("Erreur lors de l'ecriture des donnees", errr)
				}
				fichier.Close() //fermeture du fichier
				fmt.Println("La partie a bien été sauvegardé dans save.txt")
				os.Exit(0) //fermeture du jeu

				// Si le joueur écrit autre chose que stop et plus d'un caractère on renvoie une erreur

			} else {
				fmt.Println("Erreur vous avez écrit trop de caractères")
				count4 = true
			}

			// Si le joueur ecrit un caractère mais qu'il n'est pas
			// Compris entre a et z

		} else if lettre2[0] > 'z' || lettre2[0] < 'a' {
			fmt.Println("Veuillez écrire une lettre minuscule")
			count4 = true

			// Si le joueur écrit une lettre comprise entre a et z

		} else {
			fmt.Println(string(lettre2))
			a := lettre2[0]
			// Vérifie si la lettre choisie est comprise dans le mot
			for i := 0; i < len(lettrerandom_rune); i++ {
				if a == lettrerandom_rune[i] {
					motjeu[i*2] = a
					count4 = true
				}
			}
		}

		// Affiche le mot avec les lettres qui ont été découvertes

		fmt.Println(string(motjeu))

		// Si la lettre n'est pas comprise dans le mot, un dessin du pendu est renvoyé
		if !count4 {
			jose.Jose(attempts)
			attempts = attempts - 1
		}

		// Vérifie s'il nous reste des essais.

		if attempts == 0 {
			fmt.Println("Vous n'avez plus d'essais.  Game Over")
			fmt.Println("Le mot était : ", string(lettrerandom_rune))
			fmt.Println("le jeu va se fermer dans 3 secondes")
			time.Sleep(3 * time.Second)
			os.Exit(0)
		}

		// Vérifier si le mot a été trouver
		count3 = 0
		for j := 0; j < len(motjeu); j++ {
			if motjeu[j] == '_' {
				count3 = count3 + 1
			}
		}

		// Vérifie si on a trouvé le mot

		if count3 == 0 {
			fmt.Println("Vous avez gagné.")
			fmt.Println("le jeu va se fermer dans 3 secondes")
			time.Sleep(3 * time.Second)
			os.Exit(0)
		}
	}
}
