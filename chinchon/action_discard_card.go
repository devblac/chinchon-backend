package chinchon

import "fmt"

// ActionDiscardCard represents discarding a card from the player's hand.
type ActionDiscardCard struct {
	act
	Card Card `json:"card"`
}

// IsPossible returns true if the player can discard the specified card.
// This is possible after drawing and if the card is in their hand.
func (a *ActionDiscardCard) IsPossible(g GameState) bool {
	if g.TurnPlayerID != a.PlayerID || !g.HasDrawnThisTurn || g.HasDiscardedThisTurn || g.IsRoundFinished {
		return false
	}

	// Check if the card is in the player's hand
	for _, card := range g.Players[a.PlayerID].Hand.Revealed {
		if card == a.Card {
			return true
		}
	}
	return false
}

// Run executes the action of discarding the card.
func (a *ActionDiscardCard) Run(g *GameState) error {
	if !a.IsPossible(*g) {
		return errActionNotPossible
	}

	// Remove the card from the player's hand
	newHand := []Card{}
	for _, card := range g.Players[a.PlayerID].Hand.Revealed {
		if card != a.Card {
			newHand = append(newHand, card)
		}
	}
	g.Players[a.PlayerID].Hand.Revealed = newHand

	// Add the card to the discard pile
	g.DiscardPile.AddCard(a.Card)
	g.HasDiscardedThisTurn = true

	return nil
}

func (a *ActionDiscardCard) YieldsTurn(g GameState) bool {
	return true // Discarding ends the turn
}

func (a *ActionDiscardCard) String() string {
	return fmt.Sprintf("Player %v discards %v", a.PlayerID, a.Card)
}
