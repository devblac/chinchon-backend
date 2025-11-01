package chinchon

import "fmt"

// ActionKnock represents a player knocking (going out) to end the round.
type ActionKnock struct {
	act
}

// IsPossible returns true if the player can knock.
// This is possible after drawing and discarding, and if the player has valid melds.
func (a *ActionKnock) IsPossible(g GameState) bool {
	if g.TurnPlayerID != a.PlayerID || !g.HasDrawnThisTurn || !g.HasDiscardedThisTurn || g.IsRoundFinished {
		return false
	}

	// Check if the player has valid melds that leave minimal deadwood
	return a.hasValidMelds(g)
}

// hasValidMelds checks if the player has 10 or fewer deadwood points (can knock).
func (a *ActionKnock) hasValidMelds(g GameState) bool {
	deadwood := calculateDeadwoodPoints(g.Players[a.PlayerID].Hand.Revealed, g.Players[a.PlayerID].Melds)
	return deadwood <= 10
}

// Run executes the action of knocking.
func (a *ActionKnock) Run(g *GameState) error {
	if !a.IsPossible(*g) {
		return errActionNotPossible
	}

	g.KnockedPlayerID = a.PlayerID

	// Calculate round scores
	g.calculateRoundScore()

	// Update round log with melds
	roundLog := g.RoundsLog[g.RoundNumber]
	roundLog.KnockedPlayerID = a.PlayerID
	roundLog.MeldsDealt = map[int][]*Meld{
		0: append([]*Meld(nil), g.Players[0].Melds...),
		1: append([]*Meld(nil), g.Players[1].Melds...),
	}

	g.IsRoundFinished = true

	return nil
}

func (a *ActionKnock) YieldsTurn(g GameState) bool {
	return true // Knocking ends the round
}

func (a *ActionKnock) String() string {
	return fmt.Sprintf("Player %v knocks", a.PlayerID)
}
