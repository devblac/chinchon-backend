package chinchon

// ActionDrawFromDrawPile represents drawing a card from the draw pile.
type ActionDrawFromDrawPile struct {
	act
}

// IsPossible returns true if the player can draw from the draw pile.
// This is possible at the start of their turn if they haven't drawn yet.
func (a *ActionDrawFromDrawPile) IsPossible(g GameState) bool {
	return g.TurnPlayerID == a.PlayerID &&
		!g.HasDrawnThisTurn &&
		!g.DrawPile.IsEmpty() &&
		!g.IsRoundFinished
}

// Run executes the action of drawing from the draw pile.
func (a *ActionDrawFromDrawPile) Run(g *GameState) error {
	if !a.IsPossible(*g) {
		return errActionNotPossible
	}

	// Draw the top card from the draw pile
	if card, err := g.DrawPile.DrawCard(); err == nil {
		// Add the card to the player's hand
		g.Players[a.PlayerID].Hand.Revealed = append(g.Players[a.PlayerID].Hand.Revealed, card)
		g.HasDrawnThisTurn = true
	}

	return nil
}

func (a *ActionDrawFromDrawPile) YieldsTurn(g GameState) bool {
	return false // Drawing doesn't end the turn
}

// ActionDrawFromDiscardPile represents drawing the top card from the discard pile.
type ActionDrawFromDiscardPile struct {
	act
}

// IsPossible returns true if the player can draw from the discard pile.
// This is possible at the start of their turn if they haven't drawn yet.
func (a *ActionDrawFromDiscardPile) IsPossible(g GameState) bool {
	return g.TurnPlayerID == a.PlayerID &&
		!g.HasDrawnThisTurn &&
		!g.DiscardPile.IsEmpty() &&
		!g.IsRoundFinished
}

// Run executes the action of drawing from the discard pile.
func (a *ActionDrawFromDiscardPile) Run(g *GameState) error {
	if !a.IsPossible(*g) {
		return errActionNotPossible
	}

	// Draw the top card from the discard pile
	if card, err := g.DiscardPile.DrawCard(); err == nil {
		// Add the card to the player's hand
		g.Players[a.PlayerID].Hand.Revealed = append(g.Players[a.PlayerID].Hand.Revealed, card)
		g.HasDrawnThisTurn = true
	}

	return nil
}

func (a *ActionDrawFromDiscardPile) YieldsTurn(g GameState) bool {
	return false // Drawing doesn't end the turn
}
