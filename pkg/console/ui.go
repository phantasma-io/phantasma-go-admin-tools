package console

import (
	"bufio"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
)

func PromptIndexedMenu(title string, items []string) (int, string) {
	if title != "" {
		fmt.Println(title)
	}

	for menuIndex, each := range items {
		if len(items) > 9 {
			fmt.Printf("%02d - %s\n", menuIndex+1, each)
		} else {
			fmt.Printf("%d - %s\n", menuIndex+1, each)
		}
	}

	reader := bufio.NewReader(os.Stdin)

	menuIndex := 0
	for {
		fmt.Print("Enter menu index: ")
		menuIndexStr, _ := reader.ReadString('\n')
		menuIndexStr = strings.TrimSuffix(menuIndexStr, "\n")
		menuIndex, _ = strconv.Atoi(menuIndexStr)

		if menuIndex >= 1 && menuIndex <= len(items) {
			return menuIndex, items[menuIndex-1]
		}
		fmt.Printf("Please enter menu index in the range [%d-%d]\n", 1, len(items))
	}
}

func PromptYNChoice(message string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(message, " Please enter 'y' or 'n': ")
		choiceYN, _ := reader.ReadString('\n')
		choiceYN = strings.TrimSuffix(choiceYN, "\n")
		if strings.ToLower(choiceYN) == "n" {
			return false
		}
		if strings.ToLower(choiceYN) == "y" {
			return true
		}
		fmt.Println("Please enter 'y' or 'n'")
	}
}

func PromptIntInput(message string, minValue int, maxValue int) int {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [%d-%d]: ", message, minValue, maxValue)
		inputStr, _ := reader.ReadString('\n')
		inputStr = strings.TrimSuffix(inputStr, "\n")
		input, _ := strconv.Atoi(inputStr)

		if input < minValue || input > maxValue {
			fmt.Printf("Entered value '%d' is out of range [%d-%d]\n", input, minValue, maxValue)
		} else {
			return input
		}
	}
}

func PromptBigFloatInput(message string, minValue *big.Float, maxValue *big.Float) (*big.Float, string) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [%s-%s]: ", message, minValue.String(), maxValue.String())
		inputStr, _ := reader.ReadString('\n')
		inputStr = strings.TrimSuffix(inputStr, "\n")
		input, _ := big.NewFloat(0).SetString(inputStr)

		if input.Cmp(minValue) == -1 || input.Cmp(maxValue) == 1 {
			fmt.Printf("Entered value '%s' is out of range [%s-%s]\n", input.String(), minValue.String(), maxValue.String())
		} else {
			return input, inputStr
		}
	}
}

func PromptStringInput(message string) string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print(message)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSuffix(input, "\n")

	return input
}
