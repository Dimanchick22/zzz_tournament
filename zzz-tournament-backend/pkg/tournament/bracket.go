// pkg/tournament/bracket.go
package tournament

import (
	"errors"
	"math"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Player struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Rating   int    `json:"rating"`
}

type Match struct {
	ID           int     `json:"id"`
	TournamentID int     `json:"tournament_id"`
	Round        int     `json:"round"`
	Player1ID    int     `json:"player1_id"`
	Player2ID    int     `json:"player2_id"`
	Player1      *Player `json:"player1,omitempty"`
	Player2      *Player `json:"player2,omitempty"`
	WinnerID     *int    `json:"winner_id"`
	Winner       *Player `json:"winner,omitempty"`
	Status       string  `json:"status"` // pending, in_progress, finished
}

type Bracket struct {
	TournamentID int     `json:"tournament_id"`
	Rounds       int     `json:"rounds"`
	Matches      []Match `json:"matches"`
	Players      []Player `json:"players"`
}

// GenerateBracket создает турнирную сетку на выбывание
func GenerateBracket(players []Player) (*Bracket, error) {
	if len(players) < 2 {
		return nil, errors.New("need at least 2 players for tournament")
	}

	if len(players) > 64 {
		return nil, errors.New("maximum 64 players allowed")
	}

	// Перемешиваем участников для рандомности
	shuffled := make([]Player, len(players))
	copy(shuffled, players)
	
	for i := range shuffled {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	// Рассчитываем количество раундов
	rounds := int(math.Ceil(math.Log2(float64(len(shuffled)))))
	
	bracket := &Bracket{
		Rounds:  rounds,
		Players: shuffled,
		Matches: []Match{},
	}

	// Генерируем первый раунд
	round := 1
	currentPlayers := shuffled
	
	for len(currentPlayers) > 1 {
		var nextRoundPlayers []Player
		
		for i := 0; i < len(currentPlayers); i += 2 {
			if i+1 < len(currentPlayers) {
				// Обычный матч между двумя игроками
				match := Match{
					Round:     round,
					Player1ID: currentPlayers[i].ID,
					Player2ID: currentPlayers[i+1].ID,
					Player1:   &currentPlayers[i],
					Player2:   &currentPlayers[i+1],
					Status:    "pending",
				}
				bracket.Matches = append(bracket.Matches, match)
				
				// Placeholder для следующего раунда
				nextRoundPlayers = append(nextRoundPlayers, Player{ID: -1})
			} else {
				// Нечетное количество - игрок проходит дальше автоматически (bye)
				nextRoundPlayers = append(nextRoundPlayers, currentPlayers[i])
			}
		}
		
		currentPlayers = nextRoundPlayers
		round++
	}

	return bracket, nil
}

// GetNextMatches возвращает матчи, готовые к проведению
func (b *Bracket) GetNextMatches() []Match {
	var availableMatches []Match
	
	for _, match := range b.Matches {
		if match.Status == "pending" && match.Player1ID > 0 && match.Player2ID > 0 {
			availableMatches = append(availableMatches, match)
		}
	}
	
	return availableMatches
}

// AdvanceMatch продвигает турнир после завершения матча
func (b *Bracket) AdvanceMatch(matchIndex int, winnerID int) error {
	if matchIndex >= len(b.Matches) {
		return errors.New("invalid match index")
	}

	match := &b.Matches[matchIndex]
	
	// Проверяем, что победитель - один из участников матча
	if winnerID != match.Player1ID && winnerID != match.Player2ID {
		return errors.New("winner must be one of the match participants")
	}

	// Обновляем матч
	match.WinnerID = &winnerID
	match.Status = "finished"
	
	// Устанавливаем информацию о победителе
	if winnerID == match.Player1ID {
		match.Winner = match.Player1
	} else {
		match.Winner = match.Player2
	}

	// Находим следующий матч в следующем раунде
	nextRound := match.Round + 1
	
	for i := range b.Matches {
		if b.Matches[i].Round == nextRound && b.Matches[i].Status == "pending" {
			// Ищем пустое место в следующем матче
			if b.Matches[i].Player1ID == -1 {
				b.Matches[i].Player1ID = winnerID
				b.Matches[i].Player1 = match.Winner
				break
			} else if b.Matches[i].Player2ID == -1 {
				b.Matches[i].Player2ID = winnerID
				b.Matches[i].Player2 = match.Winner
				break
			}
		}
	}
	
	return nil
}

// IsTournamentFinished проверяет, завершен ли турнир
func (b *Bracket) IsTournamentFinished() (bool, *Player) {
	if len(b.Matches) == 0 {
		return false, nil
	}
	
	// Находим финальный матч (максимальный раунд)
	maxRound := 0
	var finalMatch *Match
	
	for i := range b.Matches {
		if b.Matches[i].Round > maxRound {
			maxRound = b.Matches[i].Round
			finalMatch = &b.Matches[i]
		}
	}
	
	// Проверяем, завершен ли финальный матч
	if finalMatch != nil && finalMatch.Status == "finished" && finalMatch.WinnerID != nil {
		return true, finalMatch.Winner
	}
	
	return false, nil
}

// GetMatchesByRound возвращает матчи определенного раунда
func (b *Bracket) GetMatchesByRound(round int) []Match {
	var matches []Match
	
	for _, match := range b.Matches {
		if match.Round == round {
			matches = append(matches, match)
		}
	}
	
	return matches
}

// GetProgress возвращает прогресс турнира
func (b *Bracket) GetProgress() map[string]interface{} {
	totalMatches := len(b.Matches)
	finishedMatches := 0
	
	for _, match := range b.Matches {
		if match.Status == "finished" {
			finishedMatches++
		}
	}
	
	progress := float64(finishedMatches) / float64(totalMatches) * 100
	
	return map[string]interface{}{
		"total_matches":    totalMatches,
		"finished_matches": finishedMatches,
		"progress":         progress,
		"current_round":    b.getCurrentRound(),
		"total_rounds":     b.Rounds,
	}
}

// getCurrentRound возвращает текущий активный раунд
func (b *Bracket) getCurrentRound() int {
	for round := 1; round <= b.Rounds; round++ {
		roundMatches := b.GetMatchesByRound(round)
		for _, match := range roundMatches {
			if match.Status == "pending" && match.Player1ID > 0 && match.Player2ID > 0 {
				return round
			}
		}
	}
	
	return b.Rounds // Турнир завершен
}

// ValidateBracket проверяет корректность турнирной сетки
func (b *Bracket) ValidateBracket() error {
	if len(b.Players) < 2 {
		return errors.New("not enough players")
	}
	
	if len(b.Matches) == 0 {
		return errors.New("no matches generated")
	}
	
	// Проверяем, что все игроки уникальны
	playerIDs := make(map[int]bool)
	for _, player := range b.Players {
		if playerIDs[player.ID] {
			return errors.New("duplicate player found")
		}
		playerIDs[player.ID] = true
	}
	
	return nil
}

// GetSeededBracket создает сетку с учетом рейтинга (сильные игроки не встречаются в начале)
func GenerateSeededBracket(players []Player) (*Bracket, error) {
	if len(players) < 2 {
		return nil, errors.New("need at least 2 players for tournament")
	}

	// Сортируем игроков по рейтингу (по убыванию)
	sorted := make([]Player, len(players))
	copy(sorted, players)
	
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Rating < sorted[j].Rating {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Создаем посевную сетку (1 vs последний, 2 vs предпоследний и т.д.)
	seeded := make([]Player, len(sorted))
	for i := 0; i < len(sorted); i += 2 {
		seeded[i] = sorted[i/2]
		if i+1 < len(sorted) {
			seeded[i+1] = sorted[len(sorted)-1-i/2]
		}
	}

	return GenerateBracket(seeded)
}