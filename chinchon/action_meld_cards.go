package chinchon

import "fmt"

// ActionMeldCards represents melding cards into a valid combination.
type ActionMeldCards struct {
	act
	Cards    []Card   `json:"cards"`
	MeldType MeldType `json:"meldType"`
}

// IsPossible returns true if the player can meld the specified cards.
// This is possible if the cards form a valid set or run and are in the player's hand.
func (a *ActionMeldCards) IsPossible(g GameState) bool {
	if g.TurnPlayerID != a.PlayerID || g.IsRoundFinished {
		return false
	}

	// Check if all cards are in the player's hand
	hand := make(map[Card]bool)
	for _, card := range g.Players[a.PlayerID].Hand.Revealed {
		hand[card] = true
	}

	for _, card := range a.Cards {
		if !hand[card] {
			return false
		}
	}

	// Check if the cards form a valid meld
	return a.isValidMeld()
}

// isValidMeld checks if the cards form a valid meld (set or run).
func (a *ActionMeldCards) isValidMeld() bool {
	if len(a.Cards) < 3 {
		return false
	}

	if a.MeldType == MeldTypeSet {
		return a.isValidSet()
	} else if a.MeldType == MeldTypeRun {
		return a.isValidRun()
	}

	return false
}

// isValidSet checks if the cards form a valid set (same rank, different suits).
func (a *ActionMeldCards) isValidSet() bool {
	if len(a.Cards) < 3 {
		return false
	}

	rank := a.Cards[0].Number
	for _, card := range a.Cards {
		if card.Number != rank {
			return false
		}
	}

	// Check for duplicate suits
	suits := make(map[string]bool)
	for _, card := range a.Cards {
		if suits[card.Suit] {
			return false
		}
		suits[card.Suit] = true
	}

	return true
}

// isValidRun checks if the cards form a valid run (consecutive ranks, same suit).
func (a *ActionMeldCards) isValidRun() bool {
	if len(a.Cards) < 3 {
		return false
	}

	suit := a.Cards[0].Suit
	for _, card := range a.Cards {
		if card.Suit != suit {
			return false
		}
	}

	// Sort cards by number
	sortedCards := make([]Card, len(a.Cards))
	copy(sortedCards, a.Cards)
	for i := 0; i < len(sortedCards)-1; i++ {
		for j := i + 1; j < len(sortedCards); j++ {
			if sortedCards[i].Number > sortedCards[j].Number {
				sortedCards[i], sortedCards[j] = sortedCards[j], sortedCards[i]
			}
		}
	}

	// Check for consecutive numbers
	for i := 1; i < len(sortedCards); i++ {
		if sortedCards[i].Number != sortedCards[i-1].Number+1 {
			return false
		}
	}

	return true
}

// Run executes the action of melding the cards.
func (a *ActionMeldCards) Run(g *GameState) error {
	if !a.IsPossible(*g) {
		return errActionNotPossible
	}

	// Remove the cards from the player's hand
	newHand := []Card{}
	hand := g.Players[a.PlayerID].Hand.Revealed
	for _, card := range hand {
		found := false
		for _, meldCard := range a.Cards {
			if card == meldCard {
				found = true
				break
			}
		}
		if !found {
			newHand = append(newHand, card)
		}
	}
	g.Players[a.PlayerID].Hand.Revealed = newHand

	// Add the meld to the player's melds
	meld := &Meld{
		Type:  a.MeldType,
		Cards: a.Cards,
	}
	g.Players[a.PlayerID].Melds = append(g.Players[a.PlayerID].Melds, meld)

	return nil
}

func (a *ActionMeldCards) String() string {
	return fmt.Sprintf("Player %v melds %d cards as %s", a.PlayerID, len(a.Cards), a.MeldType)
}
