package main

import (
	"fmt"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/analysis"
	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/console"
)

func settingsMenu() {
	for {
		menuIndex, _ := console.PromptIndexedMenu("\nSETTINGS MENU:",
			[]string{"Set address",
				"Change ordering",
				"Show fungible tokens",
				"Show nonfungible tokens",
				"Go back"})

		switch menuIndex {
		case 1:
			fmt.Println("Current address: ", appOpts.Address)
			appOpts.Address = console.PromptStringInput("Enter address: ")
		case 2:
			fmt.Println("")
			if appOpts.ordering == analysis.Asc {
				if console.PromptYNChoice("Change ordering? [Current ordering: Asc]") {
					appOpts.ordering = analysis.Desc
				}
			} else {
				if console.PromptYNChoice("Change ordering? [Current ordering: Desc]") {
					appOpts.ordering = analysis.Asc
				}
			}
		case 3:
			appOpts.ShowFungible = console.PromptYNChoice("Show fungible tokens in tx descriptions?")
		case 4:
			appOpts.ShowNonfungible = console.PromptYNChoice("Show nonfungible tokens in tx descriptions?")
		case 5:
			return
		}
	}
}

func interactiveMainMenu() {
	logout := false
	for !logout {
		menuIndex, _ := console.PromptIndexedMenu("\nPHANTASMA ACCOUNT ANALYZER. MENU:",
			[]string{"Settings",
				"List address transactions with balance tracking",
				"Logout"})

		switch menuIndex {
		case 1:
			settingsMenu()
		case 2:
			printTransactions(appOpts.Address, appOpts.TokenSymbol, appOpts.ordering, pagination, appOpts.ShowFungible, appOpts.ShowNonfungible)
		case 3:
			logout = true
		}
	}
}
