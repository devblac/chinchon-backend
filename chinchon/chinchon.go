package chinchon

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
)

// DefaultMaxPoints is the points a player must reach to win the game.
// It is set as a const in case support for different point limits are needed in the future.
const DefaultMaxPoints = 100

// Action names for Chinchón
const (
	DRAW_FROM_DRAW_PILE    = "draw_from_draw_pile"
	DRAW_FROM_DISCARD_PILE = "draw_from_discard_pile"
	DISCARD_CARD           = "discard_card"
	MELD_CARDS             = "meld_cards"
	KNOCK                  = "knock"
	CONFIRM_ROUND_FINISHED = "confirm_round_finished"
)

// Pile represents a pile of cards (like draw pile or discard pile).
type Pile struct {
	Cards []Card `json:"cards"`
}

// TopCard returns the top card of the pile without removing it.
// Returns an error if the pile is empty.
func (p *Pile) TopCard() (Card, error) {
	if len(p.Cards) == 0 {
		return Card{}, errors.New("pile is empty")
	}
	return p.Cards[len(p.Cards)-1], nil
}

// DrawCard removes and returns the top card from the pile.
// Returns an error if the pile is empty.
func (p *Pile) DrawCard() (Card, error) {
	if len(p.Cards) == 0 {
		return Card{}, errors.New("pile is empty")
	}
	card := p.Cards[len(p.Cards)-1]
	p.Cards = p.Cards[:len(p.Cards)-1]
	return card, nil
}

// AddCard adds a card to the top of the pile.
func (p *Pile) AddCard(card Card) {
	p.Cards = append(p.Cards, card)
}

// IsEmpty returns true if the pile has no cards.
func (p *Pile) IsEmpty() bool {
	return len(p.Cards) == 0
}

// MeldType represents the type of a meld.
type MeldType string

const (
	MeldTypeSet MeldType = "set" // Three or more cards of the same rank
	MeldTypeRun MeldType = "run" // Three or more cards in sequence of the same suit
)

// Meld represents a melded combination of cards.
type Meld struct {
	Type  MeldType `json:"type"`
	Cards []Card   `json:"cards"`
}

// IsValid checks if this meld is valid according to Chinchón rules
func (m *Meld) IsValid() bool {
	switch m.Type {
	case MeldTypeSet:
		if len(m.Cards) < 3 {
			return false
		}
		// All cards must have the same number
		number := m.Cards[0].Number
		for _, card := range m.Cards[1:] {
			if card.Number != number {
				return false
			}
		}
		return true
	case MeldTypeRun:
		if len(m.Cards) < 3 {
			return false
		}
		// All cards must have the same suit and be consecutive numbers
		suit := m.Cards[0].Suit
		numbers := make([]int, len(m.Cards))
		for i, card := range m.Cards {
			if card.Suit != suit {
				return false
			}
			numbers[i] = card.Number
		}
		sort.Ints(numbers)
		for i := 1; i < len(numbers); i++ {
			if numbers[i] != numbers[i-1]+1 {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// calculateDeadwoodPoints calculates the deadwood points for a player's hand.
// Cards in melds are not counted. Deadwood values: 1-7 = face value, 8-K = 10 points.
func calculateDeadwoodPoints(hand []Card, melds []*Meld) int {
	// Create a set of melded cards for quick lookup
	meldedCards := make(map[Card]bool)
	for _, meld := range melds {
		for _, card := range meld.Cards {
			meldedCards[card] = true
		}
	}

	points := 0
	for _, card := range hand {
		if !meldedCards[card] {
			if card.Number >= 1 && card.Number <= 7 {
				points += card.Number
			} else {
				points += 10
			}
		}
	}
	return points
}

// GameState represents the state of a Chinchón game. It is the central struct to this package.
//
// If you want to implement a client, you should look at ClientGameState instead.
type GameState struct {
	// RoundNumber is the number of the current round, starting from 1.
	RoundNumber int `json:"roundNumber"`

	// TurnPlayerID is the player ID of the player whose turn it is to play an action.
	TurnPlayerID int `json:"turnPlayerID"`

	// TurnOpponentPlayerID is the player ID of the opponent of the player whose turn it is.
	TurnOpponentPlayerID int `json:"turnOpponentPlayerID"`

	// Players is a map of player IDs to their respective hands, melds, and scores.
	// There are 2 players in a game. Use TurnPlayerID and TurnOpponentPlayerID to index
	// into this map, or iterate over it to discover player ids.
	Players map[int]*Player `json:"players"`

	// PossibleActions is a list of possible actions that the current player can take.
	// Possible actions are calculated based on game state and updated after each action.
	PossibleActions []json.RawMessage `json:"possibleActions"`

	// DrawPile contains the cards remaining to be drawn.
	DrawPile *Pile `json:"drawPile"`

	// DiscardPile contains the cards that have been discarded. The top card is visible.
	DiscardPile *Pile `json:"discardPile"`

	// HasDrawnThisTurn tracks whether the current player has drawn a card this turn.
	HasDrawnThisTurn bool `json:"hasDrawnThisTurn"`

	// HasDiscardedThisTurn tracks whether the current player has discarded a card this turn.
	HasDiscardedThisTurn bool `json:"hasDiscardedThisTurn"`

	// KnockedPlayerID is the player ID of the player who knocked (went out), or -1 if no one has knocked.
	KnockedPlayerID int `json:"knockedPlayerID"`

	// IsRoundFinished is true if the current round is finished. Each action's `Run()` method is responsible
	// for setting this. During `GameState.RunAction()`, If the action's `Run()` method sets this to true,
	// then `GameState.startNewRound()` will be called.
	IsRoundFinished bool `json:"isRoundFinished"`

	// IsGameEnded is true if the whole game is ended, rather than an individual round. This happens when
	// a player reaches MaxPoints points.
	IsGameEnded bool `json:"isGameEnded"`

	// WinnerPlayerID is the player ID of the player who won the game. This is only set when `IsGameEnded` is
	// `true`. Otherwise, it's -1.
	WinnerPlayerID int `json:"winnerPlayerID"`

	// RoundsLog is the ordered list of logs of each round that was played in the game.
	//
	// Use GameState.RoundNumber to index into this list (note thus that it's 1-indexed).
	// This means that there is an empty round at the beginning of the list.
	//
	// Note that there is a "live entry" for the current round.
	RoundsLog []*RoundLog `json:"roundsLog"`

	RoundFinishedConfirmedPlayerIDs map[int]bool `json:"roundFinishedConfirmedPlayerIDs"`

	RuleMaxPoints int `json:"ruleMaxPoints"`

	deck *deck `json:"-"`
}

type Player struct {
	// Hand contains the cards in the player's hand (not yet melded).
	Hand *Hand `json:"hand"`

	// Melds contains the melded combinations (sets and runs) laid down by the player.
	Melds []*Meld `json:"melds"`

	// Score is the player's total score (from 0 to MaxPoints).
	Score int `json:"score"`
}

// RoundLog is a log of a round that was played in the game
type RoundLog struct {
	// HandsDealt is a map from PlayerID to their initial hand during this round.
	HandsDealt map[int]*Hand `json:"handsDealt"`

	// MeldsDealt is a map from PlayerID to their melds at the end of the round.
	MeldsDealt map[int][]*Meld `json:"meldsDealt"`

	// KnockedPlayerID is the player who knocked to end the round, or -1 if no one knocked.
	KnockedPlayerID int `json:"knockedPlayerID"`

	// WinnerPlayerID is the player who won this round (the one with lower deadwood points).
	WinnerPlayerID int `json:"winnerPlayerID"`

	// LoserPlayerID is the player who lost this round.
	LoserPlayerID int `json:"loserPlayerID"`

	// WinnerDeadwoodPoints is the deadwood point total of the winner.
	WinnerDeadwoodPoints int `json:"winnerDeadwoodPoints"`

	// LoserDeadwoodPoints is the deadwood point total of the loser.
	LoserDeadwoodPoints int `json:"loserDeadwoodPoints"`

	// PointsAwarded is the number of points awarded to the winner.
	PointsAwarded int `json:"pointsAwarded"`

	// ActionsLog is the ordered list of actions of this round.
	ActionsLog []ActionLog `json:"actionsLog"`
}

// ActionLog is a log of an action that was run in a round.
type ActionLog struct {
	// PlayerID is the player ID of the player who ran the action.
	PlayerID int `json:"playerID"`

	// Action is a JSON-serialized action. This is because `Action` is an interface, and we can't
	// serialize it directly otherwise. Clients should use `chinchon.DeserializeAction`.`
	Action json.RawMessage `json:"action"`
}

// WithMaxPoints sets the maximum points required to win the game.
func WithMaxPoints(maxPoints int) func(*GameState) {
	return func(gs *GameState) {
		gs.RuleMaxPoints = maxPoints
	}
}

func New(opts ...func(*GameState)) *GameState {
	gs := &GameState{
		RoundNumber:          0,
		TurnPlayerID:         0, // Player 0 starts first
		TurnOpponentPlayerID: 1,
		Players: map[int]*Player{
			0: {Hand: nil, Melds: nil, Score: 0},
			1: {Hand: nil, Melds: nil, Score: 0},
		},
		IsGameEnded:          false,
		WinnerPlayerID:       -1,
		RoundsLog:            []*RoundLog{{}}, // initialised with an empty round to be 1-indexed
		KnockedPlayerID:      -1,
		HasDrawnThisTurn:     false,
		HasDiscardedThisTurn: false,
		deck:                 newDeck(),
		RuleMaxPoints:        DefaultMaxPoints,
	}

	for _, opt := range opts {
		opt(gs)
	}

	gs.startNewRound()

	return gs
}

func (g *GameState) startNewRound() {
	g.deck.shuffle()
	g.RoundNumber++

	// Alternate who starts the round
	g.TurnPlayerID = g.OpponentOf(g.TurnPlayerID)
	g.TurnOpponentPlayerID = g.OpponentOf(g.TurnPlayerID)

	// Deal 7 cards to each player
	player0Hand := &Hand{}
	player1Hand := &Hand{}
	for i := 0; i < 7; i++ {
		player0Hand.Revealed = append(player0Hand.Revealed, g.deck.cards[0])
		g.deck.cards = g.deck.cards[1:]
		player1Hand.Revealed = append(player1Hand.Revealed, g.deck.cards[0])
		g.deck.cards = g.deck.cards[1:]
	}

	g.Players[0].Hand = player0Hand
	g.Players[1].Hand = player1Hand
	g.Players[0].Melds = []*Meld{}
	g.Players[1].Melds = []*Meld{}

	// Create draw pile with remaining cards
	g.DrawPile = &Pile{Cards: make([]Card, len(g.deck.cards))}
	copy(g.DrawPile.Cards, g.deck.cards)

	// Create discard pile with one card from draw pile
	g.DiscardPile = &Pile{}
	if !g.DrawPile.IsEmpty() {
		if card, err := g.DrawPile.DrawCard(); err == nil {
			g.DiscardPile.AddCard(card)
		}
	}

	// Reset round state
	g.KnockedPlayerID = -1
	g.HasDrawnThisTurn = false
	g.HasDiscardedThisTurn = false
	g.IsRoundFinished = false
	g.RoundFinishedConfirmedPlayerIDs = map[int]bool{}

	g.RoundsLog = append(g.RoundsLog, &RoundLog{
		HandsDealt: map[int]*Hand{
			0: g.Players[0].Hand,
			1: g.Players[1].Hand,
		},
		MeldsDealt: map[int][]*Meld{
			0: g.Players[0].Melds,
			1: g.Players[1].Melds,
		},
		KnockedPlayerID:      -1,
		WinnerPlayerID:       -1,
		LoserPlayerID:        -1,
		WinnerDeadwoodPoints: 0,
		LoserDeadwoodPoints:  0,
		PointsAwarded:        0,
		ActionsLog:           []ActionLog{},
	})

	g.PossibleActions = _serializeActions(g.CalculatePossibleActions())
}

func (g *GameState) RunAction(action Action) error {
	if action == nil {
		return nil
	}

	if g.IsGameEnded {
		return fmt.Errorf("%w trying to run [%v]", errGameIsEnded, action)
	}

	if !g.IsRoundFinished && action.GetPlayerID() != g.TurnPlayerID {
		return errNotYourTurn
	}

	if !action.IsPossible(*g) {
		return fmt.Errorf("%w trying to run [%v]", errActionNotPossible, action)
	}
	err := action.Run(g)
	if err != nil {
		return fmt.Errorf("%w trying to run [%v] after checking it was possible", err, action)
	}

	if action.GetName() != CONFIRM_ROUND_FINISHED {
		g.RoundsLog[g.RoundNumber].ActionsLog = append(g.RoundsLog[g.RoundNumber].ActionsLog, ActionLog{
			PlayerID: g.TurnPlayerID,
			Action:   SerializeAction(action),
		})
	}

	// Start new round if current round is finished
	if !g.IsGameEnded && g.IsRoundFinished && len(g.RoundFinishedConfirmedPlayerIDs) == 2 {
		// fmt.Println("Starting new round...")
		g.startNewRound()
		return nil
	}

	// Switch player turn within current round (unless current action doesn't yield turn)
	if !g.IsGameEnded && !g.IsRoundFinished && action.YieldsTurn(*g) {
		g.TurnPlayerID, g.TurnOpponentPlayerID = g.TurnOpponentPlayerID, g.TurnPlayerID
		// Reset turn state for the new player
		g.HasDrawnThisTurn = false
		g.HasDiscardedThisTurn = false
	}

	if !g.IsGameEnded && g.IsRoundFinished && len(g.RoundFinishedConfirmedPlayerIDs) == 1 {
		if g.RoundFinishedConfirmedPlayerIDs[g.TurnPlayerID] {
			g.changeTurn()
		}
	}

	// Handle end of game due to score
	for playerID := range g.Players {
		if g.Players[playerID].Score >= g.RuleMaxPoints {
			g.Players[playerID].Score = g.RuleMaxPoints
			g.IsGameEnded = true
			g.WinnerPlayerID = playerID
		}
	}

	possibleActions := g.CalculatePossibleActions()
	if g.countActionsOfTurnPlayer() == 0 {
		// If the current player has no actions left, it's the opponent's turn.
		g.changeTurn()
		possibleActions = g.CalculatePossibleActions()
	}

	g.PossibleActions = _serializeActions(possibleActions)

	// log.Printf("Possible actions: %v\n", possibleActions)

	return nil
}

func (g *GameState) changeTurn() {
	g.TurnPlayerID, g.TurnOpponentPlayerID = g.TurnOpponentPlayerID, g.TurnPlayerID
}

func (g GameState) countActionsOfTurnPlayer() int {
	count := 0
	for _, a := range g.CalculatePossibleActions() {
		if a.GetPlayerID() == g.TurnPlayerID {
			count++
		}
	}
	return count
}

func (g GameState) OpponentOf(playerID int) int {
	for id := range g.Players {
		if id != playerID {
			return id
		}
	}
	return -1 // Unreachable
}

func (g GameState) Serialize() ([]byte, error) {
	return json.Marshal(g)
}

func (g *GameState) PrettyPrint() (string, error) {
	var prettyJSON []byte
	prettyJSON, err := json.MarshalIndent(g, "", "    ")
	if err != nil {
		return "", err
	}
	return string(prettyJSON), nil
}

// generatePossibleMeldActions generates all possible valid meld actions for a player
func (g *GameState) generatePossibleMeldActions(playerID int) []Action {
	actions := []Action{}
	hand := g.Players[playerID].Hand.Revealed

	// Generate all possible sets (3+ cards of same rank)
	actions = append(actions, g.generateSetMeldActions(hand, playerID)...)

	// Generate all possible runs (3+ consecutive cards of same suit)
	actions = append(actions, g.generateRunMeldActions(hand, playerID)...)

	return actions
}

// generateSetMeldActions generates all possible set melds (same rank, different suits)
func (g *GameState) generateSetMeldActions(hand []Card, playerID int) []Action {
	actions := []Action{}

	// Group cards by rank
	rankGroups := make(map[int][]Card)
	for _, card := range hand {
		rankGroups[card.Number] = append(rankGroups[card.Number], card)
	}

	// For each rank with 3+ cards, generate all possible combinations of 3 cards
	for _, cards := range rankGroups {
		if len(cards) >= 3 {
			// Generate combinations of 3 cards from the available cards
			combinations := g.generateCombinations(cards, 3)
			for _, combo := range combinations {
				// Check if it's a valid set (different suits)
				if g.isValidSet(combo) {
					action := NewActionMeldCards(combo, MeldTypeSet, playerID)
					actions = append(actions, action)
				}
			}
		}
	}

	return actions
}

// generateRunMeldActions generates all possible run melds (consecutive ranks, same suit)
func (g *GameState) generateRunMeldActions(hand []Card, playerID int) []Action {
	actions := []Action{}

	// Group cards by suit
	suitGroups := make(map[string][]Card)
	for _, card := range hand {
		suitGroups[card.Suit] = append(suitGroups[card.Suit], card)
	}

	// For each suit, find all possible runs
	for _, cards := range suitGroups {
		if len(cards) >= 3 {
			// Sort cards by number
			sortedCards := make([]Card, len(cards))
			copy(sortedCards, cards)
			for i := 0; i < len(sortedCards)-1; i++ {
				for j := i + 1; j < len(sortedCards); j++ {
					if sortedCards[i].Number > sortedCards[j].Number {
						sortedCards[i], sortedCards[j] = sortedCards[j], sortedCards[i]
					}
				}
			}

			// Find all consecutive sequences of 3+ cards
			runs := g.findConsecutiveRuns(sortedCards)
			for _, run := range runs {
				action := NewActionMeldCards(run, MeldTypeRun, playerID)
				actions = append(actions, action)
			}
		}
	}

	return actions
}

// generateCombinations generates all combinations of size k from the given cards
func (g *GameState) generateCombinations(cards []Card, k int) [][]Card {
	if k == 0 {
		return [][]Card{{}}
	}
	if len(cards) < k {
		return [][]Card{}
	}

	var combinations [][]Card

	// Include first card
	first := cards[0]
	remaining := cards[1:]
	subCombos := g.generateCombinations(remaining, k-1)
	for _, combo := range subCombos {
		newCombo := append([]Card{first}, combo...)
		combinations = append(combinations, newCombo)
	}

	// Exclude first card
	combinations = append(combinations, g.generateCombinations(remaining, k)...)

	return combinations
}

// findConsecutiveRuns finds all maximal consecutive runs in sorted cards
func (g *GameState) findConsecutiveRuns(sortedCards []Card) [][]Card {
	var runs [][]Card

	i := 0
	for i < len(sortedCards) {
		start := i
		// Find end of current run
		for i < len(sortedCards)-1 && sortedCards[i+1].Number == sortedCards[i].Number+1 {
			i++
		}

		// If run is 3+ cards, add all possible sub-runs of length 3+
		runLength := i - start + 1
		if runLength >= 3 {
			runCards := sortedCards[start : i+1]
			// Add all possible runs of length 3+ from this sequence
			for length := 3; length <= runLength; length++ {
				for startIdx := 0; startIdx <= runLength-length; startIdx++ {
					run := runCards[startIdx : startIdx+length]
					runs = append(runs, run)
				}
			}
		}

		i++
	}

	return runs
}

// isValidSet checks if the given cards form a valid set (same rank, different suits)
func (g *GameState) isValidSet(cards []Card) bool {
	if len(cards) < 3 {
		return false
	}

	rank := cards[0].Number
	suits := make(map[string]bool)

	for _, card := range cards {
		if card.Number != rank {
			return false
		}
		if suits[card.Suit] {
			return false // Duplicate suit
		}
		suits[card.Suit] = true
	}

	return true
}

// calculateRoundScore calculates the scores for both players at the end of a round
func (g *GameState) calculateRoundScore() {
	roundLog := g.RoundsLog[g.RoundNumber]

	// Calculate deadwood for both players
	player0Deadwood := calculateDeadwoodPoints(g.Players[0].Hand.Revealed, g.Players[0].Melds)
	player1Deadwood := calculateDeadwoodPoints(g.Players[1].Hand.Revealed, g.Players[1].Melds)

	roundLog.WinnerDeadwoodPoints = player0Deadwood
	roundLog.LoserDeadwoodPoints = player1Deadwood
	roundLog.WinnerPlayerID = 0
	roundLog.LoserPlayerID = 1

	// Determine winner (lower deadwood wins)
	if player1Deadwood < player0Deadwood {
		roundLog.WinnerDeadwoodPoints = player1Deadwood
		roundLog.LoserDeadwoodPoints = player0Deadwood
		roundLog.WinnerPlayerID = 1
		roundLog.LoserPlayerID = 0
	} else if player1Deadwood == player0Deadwood {
		// Tie goes to the player who didn't knock, or if both knocked, to the non-knocker
		// For simplicity, if it's a tie, the non-knocker wins
		if roundLog.KnockedPlayerID == 0 {
			roundLog.WinnerPlayerID = 1
			roundLog.LoserPlayerID = 0
		} else {
			roundLog.WinnerPlayerID = 0
			roundLog.LoserPlayerID = 1
		}
	}

	// Calculate points awarded
	winnerDeadwood := roundLog.WinnerDeadwoodPoints
	loserDeadwood := roundLog.LoserDeadwoodPoints
	points := loserDeadwood - winnerDeadwood

	// Bonus for going gin (0 deadwood)
	if winnerDeadwood == 0 {
		points += 25
	}

	// Bonus for undercutting (opponent has higher deadwood when you knock)
	if roundLog.KnockedPlayerID != -1 && roundLog.KnockedPlayerID != roundLog.WinnerPlayerID {
		points += 10
	}

	roundLog.PointsAwarded = points
	g.Players[roundLog.WinnerPlayerID].Score += points
}

type Action interface {
	IsPossible(g GameState) bool
	Run(g *GameState) error
	GetName() string
	GetPlayerID() int
	YieldsTurn(g GameState) bool
	// Some actions need to be enriched with additional information.
	// e.g. a knock action might be enriched with additional game state.
	// GameState.CalculatePossibleActions() must call this method on all actions.
	Enrich(g GameState)

	// GetPriority is used by GameState to calculate which actions are possible.
	// By default, all actions have priority 0. In principle, all actions that are
	// possible will be collected. If an action with higher priority is found,
	// all possible actions are removed, and only actions with this higher priority
	// will be collected. And so on.
	//
	// For example, if Flor is possible, then it should be higher priority.
	GetPriority() int

	AllowLowerPriority() bool

	fmt.Stringer
}

var (
	errActionNotPossible = errors.New("action not possible")
	errGameIsEnded       = errors.New("game is ended")
	errNotYourTurn       = errors.New("not your turn")
)

func (g GameState) CalculatePossibleActions() []Action {
	allActions := []Action{}

	// If round is finished, both players can confirm
	if g.IsRoundFinished {
		allActions = append(allActions,
			NewActionConfirmRoundFinished(g.TurnPlayerID),
			NewActionConfirmRoundFinished(g.TurnOpponentPlayerID),
		)
	} else {
		// Normal turn actions
		if !g.HasDrawnThisTurn {
			// Player must draw first
			allActions = append(allActions,
				NewActionDrawFromDrawPile(g.TurnPlayerID),
				NewActionDrawFromDiscardPile(g.TurnPlayerID),
			)
		} else if !g.HasDiscardedThisTurn {
			// Player must discard after drawing
			for _, card := range g.Players[g.TurnPlayerID].Hand.Revealed {
				allActions = append(allActions, NewActionDiscardCard(card, g.TurnPlayerID))
			}
		} else {
			// Player has drawn and discarded, can now meld or knock
			allActions = append(allActions, NewActionKnock(g.TurnPlayerID))
			// Add all possible meld actions
			meldActions := g.generatePossibleMeldActions(g.TurnPlayerID)
			allActions = append(allActions, meldActions...)
		}
	}

	possibleActions := []Action{}
	for _, action := range allActions {
		action.Enrich(g)
		if action.IsPossible(g) {
			possibleActions = append(possibleActions, action)
		}
	}
	return possibleActions
}

func SerializeAction(action Action) []byte {
	bs, _ := json.Marshal(action)
	return bs
}

func DeserializeAction(bs []byte) (Action, error) {
	var actionName struct {
		Name string `json:"name"`
	}

	err := json.Unmarshal(bs, &actionName)
	if err != nil {
		return nil, err
	}

	var action Action
	switch actionName.Name {
	case DRAW_FROM_DRAW_PILE:
		action = &ActionDrawFromDrawPile{}
	case DRAW_FROM_DISCARD_PILE:
		action = &ActionDrawFromDiscardPile{}
	case DISCARD_CARD:
		action = &ActionDiscardCard{}
	case MELD_CARDS:
		action = &ActionMeldCards{}
	case KNOCK:
		action = &ActionKnock{}
	case CONFIRM_ROUND_FINISHED:
		action = &ActionConfirmRoundFinished{}
	default:
		return nil, fmt.Errorf("unknown action: [%v]", string(bs))
	}

	err = json.Unmarshal(bs, action)
	if err != nil {
		return nil, err
	}

	return action, nil
}

func _serializeActions(as []Action) []json.RawMessage {
	_as := []json.RawMessage{}
	for _, a := range as {
		_as = append(_as, json.RawMessage(SerializeAction(a)))
	}
	return _as
}

func _deserializeCurrentRoundLastAction(g GameState) Action {
	lastAction := g.RoundsLog[g.RoundNumber].ActionsLog[len(g.RoundsLog[g.RoundNumber].ActionsLog)-1].Action
	a, _ := DeserializeAction(lastAction)
	return a
}

func _deserializeCurrentRoundActions(g GameState) []Action {
	curRoundActions := g.RoundsLog[g.RoundNumber].ActionsLog
	actions := make([]Action, len(curRoundActions))
	for i, actionLog := range curRoundActions {
		action, _ := DeserializeAction(actionLog.Action)
		actions[i] = action
	}
	return actions
}

func _deserializeCurrentRoundActionsByPlayerID(playerID int, g GameState) []Action {
	actions := _deserializeCurrentRoundActions(g)
	filteredActions := []Action{}
	for _, a := range actions {
		if a.GetPlayerID() == playerID {
			filteredActions = append(filteredActions, a)
		}
	}
	return filteredActions
}

func (g *GameState) ToClientGameState(youPlayerID int) ClientGameState {
	themPlayerID := g.OpponentOf(youPlayerID)

	// GameState may have possible game actions that this player can't take.
	filteredPossibleActions := []Action{}
	for _, a := range g.CalculatePossibleActions() {
		if a.GetPlayerID() == youPlayerID {
			filteredPossibleActions = append(filteredPossibleActions, a)
		}
	}

	cgs := ClientGameState{
		RoundNumber:         g.RoundNumber,
		TurnPlayerID:        g.TurnPlayerID,
		YouPlayerID:         youPlayerID,
		ThemPlayerID:        themPlayerID,
		YourScore:           g.Players[youPlayerID].Score,
		TheirScore:          g.Players[themPlayerID].Score,
		YourHandCards:       g.Players[youPlayerID].Hand.Revealed,
		TheirHandCards:      g.Players[themPlayerID].Hand.Revealed,
		YourMelds:           g.Players[youPlayerID].Melds,
		TheirMelds:          g.Players[themPlayerID].Melds,
		DiscardPileTopCard:  func() Card { card, _ := g.DiscardPile.TopCard(); return card }(),
		PossibleActions:     _serializeActions(filteredPossibleActions),
		IsGameEnded:         g.IsGameEnded,
		IsRoundFinished:     g.IsRoundFinished,
		WinnerPlayerID:      g.WinnerPlayerID,
		KnockedPlayerID:     g.KnockedPlayerID,
		YourDeadwoodPoints:  calculateDeadwoodPoints(g.Players[youPlayerID].Hand.Revealed, g.Players[youPlayerID].Melds),
		TheirDeadwoodPoints: calculateDeadwoodPoints(g.Players[themPlayerID].Hand.Revealed, g.Players[themPlayerID].Melds),
		RuleMaxPoints:       g.RuleMaxPoints,
	}

	if len(g.RoundsLog[g.RoundNumber].ActionsLog) > 0 {
		actionsLog := g.RoundsLog[g.RoundNumber].ActionsLog
		cgs.LastActionLog = &actionsLog[len(actionsLog)-1]
	}

	return cgs
}

// ClientGameState represents the state of a Chinchón game as available to a client.
//
// It is returned by the server on every single call, so if you want to implement a client,
// you need to be very familiar with this struct.
type ClientGameState struct {
	// RoundNumber is the number of the current round, starting from 1.
	RoundNumber int `json:"roundNumber"`

	// TurnPlayerID is the player ID of the player whose turn it is to play an action.
	TurnPlayerID int `json:"turnPlayerID"`

	YouPlayerID        int     `json:"you"`
	ThemPlayerID       int     `json:"them"`
	YourScore          int     `json:"yourScore"`
	TheirScore         int     `json:"theirScore"`
	YourHandCards      []Card  `json:"yourHandCards"`
	TheirHandCards     []Card  `json:"theirHandCards"`
	YourMelds          []*Meld `json:"yourMelds"`
	TheirMelds         []*Meld `json:"theirMelds"`
	DiscardPileTopCard Card    `json:"discardPileTopCard"`

	// PossibleActions is a list of possible actions that the current player can take.
	PossibleActions []json.RawMessage `json:"possibleActions"`

	// IsGameEnded is true if the whole game is ended, rather than an individual round. This happens when
	// a player reaches MaxPoints points.
	IsGameEnded bool `json:"isGameEnded"`

	IsRoundFinished bool `json:"isRoundFinished"`

	// WinnerPlayerID is the player ID of the player who won the game. This is only set when `IsGameEnded` is
	// `true`. Otherwise, it's -1.
	WinnerPlayerID int `json:"winnerPlayerID"`

	// KnockedPlayerID is the player who knocked to end the round, or -1 if no one has knocked.
	KnockedPlayerID int `json:"knockedPlayerID"`

	// Deadwood points for each player (calculated from unmelded cards)
	YourDeadwoodPoints  int `json:"yourDeadwoodPoints"`
	TheirDeadwoodPoints int `json:"theirDeadwoodPoints"`

	// LastActionLog is the log of the last action that was run in the current round. If the round has
	// just started, this will be nil. Clients typically want to use this to show the current player
	// what the opponent just did.
	LastActionLog *ActionLog `json:"lastActionLog"`

	RuleMaxPoints int `json:"ruleMaxPoints"`
}

type Bot interface {
	ChooseAction(ClientGameState) Action
}
