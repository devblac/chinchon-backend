package chinchon

import (
	"fmt"
	"strings"
)

type act struct {
	Name     string `json:"name"`
	PlayerID int    `json:"playerID"`

	fmt.Stringer `json:"-"`
}

func (a act) GetName() string {
	return a.Name
}

func (a act) GetPlayerID() int {
	return a.PlayerID
}

func (a act) GetPriority() int {
	return 0
}

func (a act) AllowLowerPriority() bool {
	return false
}

// By default, actions don't need to be enriched.
func (a act) Enrich(g GameState) {}

func (a act) String() string {
	name := strings.ReplaceAll(a.Name, "_", " ")
	return fmt.Sprintf("Player %v %v", a.PlayerID, name)
}

func (a act) YieldsTurn(g GameState) bool {
	return true
}

func NewActionDrawFromDrawPile(playerID int) Action {
	return &ActionDrawFromDrawPile{act: act{Name: DRAW_FROM_DRAW_PILE, PlayerID: playerID}}
}

func NewActionDrawFromDiscardPile(playerID int) Action {
	return &ActionDrawFromDiscardPile{act: act{Name: DRAW_FROM_DISCARD_PILE, PlayerID: playerID}}
}

func NewActionDiscardCard(card Card, playerID int) Action {
	return &ActionDiscardCard{act: act{Name: DISCARD_CARD, PlayerID: playerID}, Card: card}
}

func NewActionMeldCards(cards []Card, meldType MeldType, playerID int) Action {
	return &ActionMeldCards{act: act{Name: MELD_CARDS, PlayerID: playerID}, Cards: cards, MeldType: meldType}
}

func NewActionKnock(playerID int) Action {
	return &ActionKnock{act: act{Name: KNOCK, PlayerID: playerID}}
}

func NewActionConfirmRoundFinished(playerID int) Action {
	return &ActionConfirmRoundFinished{act: act{Name: CONFIRM_ROUND_FINISHED, PlayerID: playerID}}
}

