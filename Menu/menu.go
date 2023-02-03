package menu

import (
	"fmt"
)

func PrintMenu() {
	fmt.Println()
	fmt.Println("-------------------- MENU --------------------")
	fmt.Println("1 - Dodovanje elementa(kljuc : vrednost)")
	fmt.Println("2 - Brisanje elementa")
	fmt.Println("3 - Pretraga")
	fmt.Println("4 - Uradi kompakciju fajlova u bazi")
	fmt.Println("5 - Ostale funkcionalnosti")
	fmt.Println("X - Izlazak iz programa")
}

func PrintSearchMenu() {
	fmt.Println()
	fmt.Println("-------------------- Vrste pretrage --------------------")
	fmt.Println("1 - Obicna pretraga")
	fmt.Println("2 - RANGE SCAN")
	fmt.Println("3 - LIST")
	fmt.Println("4 - Nazad")
	fmt.Println("X - Izlazak iz programa")
}

func PrintOtherMenu() {
	fmt.Println()
	fmt.Println("-------------------- Ostale funkcionalnosti --------------------")
	fmt.Println("1 - BloomFilter funkcionalnosti")
	fmt.Println("2 - CMS funkcionalnosti")
	fmt.Println("3 - HLL funkcionalnosti")
	fmt.Println("4 - SimHash funkcionalnosti")
	fmt.Println("5 - Nazad")
	fmt.Println("X - Izlazak iz programa")
}

func PrintBloomFilterMenu() {
	fmt.Println()
	fmt.Println("-------------------- BloomFilter funkcionalnosti --------------------")
	fmt.Println("1 - Dodavanje novog BloomFilter-a u bazu")
	fmt.Println("2 - Brisanje BloomFilter-a iz baze")
	fmt.Println("3 - Dodavanje elementa")
	fmt.Println("4 - Provera elementa")
	fmt.Println("5 - Nazad")
	fmt.Println("X - Izlazak iz programa")

}

func PrintCMSMenu() {
	fmt.Println()
	fmt.Println("-------------------- CMS funkcionalnosti --------------------")
	fmt.Println("1 - Dodavanje novog CMS-a u bazu")
	fmt.Println("2 - Brisanje CMS-a iz baze")
	fmt.Println("3 - Dodavanje elementa")
	fmt.Println("4 - Provera elementa")
	fmt.Println("5 - Nazad")
	fmt.Println("X - Izlazak iz programa")
}

func PrintHLLMenu() {
	fmt.Println()
	fmt.Println("-------------------- HLL funkcionalnosti --------------------")
	fmt.Println("1 - Dodavanje novog HLL-a u bazu")
	fmt.Println("2 - Brisanje HLL-a iz baze")
	fmt.Println("3 - Dodavanje elementa")
	fmt.Println("4 - Provera kardinalnosti")
	fmt.Println("5 - Nazad")
	fmt.Println("X - Izlazak iz programa")
}

func PrintSimHashMenu() {
	fmt.Println()
	fmt.Println("-------------------- SimHash funkcionalnosti --------------------")
	fmt.Println("1 - Dodavanje novog SimHash-a u bazu")
	fmt.Println("2 - Brisanje SimHash-a iz baze")
	fmt.Println("3 - Provera sa drugim tekstom")
	fmt.Println("4 - Nazad")
	fmt.Println("X - Izlazak iz programa")
}

func ReadValue(text string) string {
	var input string
	for {
		fmt.Print(text)
		n, err := fmt.Scanln(&input)
		if err != nil {
			fmt.Println("Lose ste uneli komandu. Probajte ponovo.")
			continue
		}

		if n == 0 || input == "" {
			fmt.Println("Lose ste uneli komandu. Probajte ponovo.")
			continue
		}
		break
	}
	return input
}
