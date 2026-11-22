package game

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
)

const MaxDice = 100

func RollDie(sides int) int {
	return rand.IntN(sides) + 1
}

func parseDice(formula string) (int, int, error) {
	before, after, ok := strings.Cut(formula, "d")
	if !ok {
		return 0, 0, fmt.Errorf("invalid dice formula %q", formula)
	}

	dices, err := strconv.Atoi(before)
	if err != nil {
		return 0, 0, fmt.Errorf("not number of dice found: %w", err)
	}

	faces, err := strconv.Atoi(after)
	if err != nil {
		return 0, 0, fmt.Errorf("not number of face on dice(s) found: %w", err)
	}

	if faces <= 0 || dices <= 0 {
		return 0, 0, errors.New("the number of dices and faces has to be > 0")
	}
	if faces > MaxDice || dices > MaxDice {
		return 0, 0, fmt.Errorf("dice formula %q is too big (max %dd%d)", formula, MaxDice, MaxDice)
	}
	return dices, faces, nil
}

func ValidDice(formula string) error {
	_, _, err := parseDice(formula)
	return err
}

func Roll(formula string) (int, error) {
	dices, faces, err := parseDice(formula)
	if err != nil {
		return 0, err
	}

	var total int
	for i := 0; i < dices; i++ {
		total += RollDie(faces)
	}

	return total, nil
}
