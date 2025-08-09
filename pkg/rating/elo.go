// pkg/rating/elo.go
package rating

import "math"

const (
	// Стандартные K-факторы
	KFactorNovice     = 40.0 // Новички (< 30 игр)
	KFactorRegular    = 25.0 // Обычные игроки (< 2100 рейтинга)
	KFactorStrong     = 15.0 // Сильные игроки (< 2400 рейтинга)
	KFactorPro        = 10.0 // Профессионалы (2400+ рейтинга)
	
	// Рейтинговые пороги
	RatingThresholdStrong = 2100
	RatingThresholdPro    = 2400
	
	// Пороги по количеству игр
	GamesThresholdNovice = 30
	
	// Минимальный и максимальный рейтинг
	MinRating = 0
	MaxRating = 4000
)

type Player struct {
	ID          int `json:"id"`
	Rating      int `json:"rating"`
	GamesPlayed int `json:"games_played"`
	Wins        int `json:"wins"`
	Losses      int `json:"losses"`
}

type MatchResult struct {
	WinnerID int `json:"winner_id"`
	LoserID  int `json:"loser_id"`
	Type     string `json:"type"` // "tournament", "ranked", "casual"
}

// CalculateRatingChange рассчитывает изменение рейтинга по системе Elo
func CalculateRatingChange(winnerRating, loserRating int, kFactor float64) (int, int) {
	if kFactor == 0 {
		kFactor = KFactorRegular
	}
	
	// Ожидаемый результат для победителя
	expectedWinner := 1.0 / (1.0 + math.Pow(10, float64(loserRating-winnerRating)/400.0))
	
	// Ожидаемый результат для проигравшего
	expectedLoser := 1.0 / (1.0 + math.Pow(10, float64(winnerRating-loserRating)/400.0))
	
	// Новые рейтинги
	newWinnerRating := winnerRating + int(kFactor*(1.0-expectedWinner))
	newLoserRating := loserRating + int(kFactor*(0.0-expectedLoser))
	
	// Ограничиваем рейтинги
	if newWinnerRating > MaxRating {
		newWinnerRating = MaxRating
	}
	if newLoserRating < MinRating {
		newLoserRating = MinRating
	}
	
	return newWinnerRating, newLoserRating
}

// GetKFactor возвращает коэффициент K в зависимости от рейтинга и количества игр
func GetKFactor(rating, gamesPlayed int) float64 {
	if gamesPlayed < GamesThresholdNovice {
		return KFactorNovice // Новые игроки
	}
	
	if rating < RatingThresholdStrong {
		return KFactorRegular // Обычные игроки
	}
	
	if rating < RatingThresholdPro {
		return KFactorStrong // Сильные игроки
	}
	
	return KFactorPro // Профи
}

// CalculateExpectedScore рассчитывает ожидаемый результат матча
func CalculateExpectedScore(playerRating, opponentRating int) float64 {
	return 1.0 / (1.0 + math.Pow(10, float64(opponentRating-playerRating)/400.0))
}

// GetRatingDifference возвращает разность рейтингов между игроками
func GetRatingDifference(rating1, rating2 int) int {
	if rating1 > rating2 {
		return rating1 - rating2
	}
	return rating2 - rating1
}

// GetMatchProbability возвращает вероятность победы первого игрока
func GetMatchProbability(rating1, rating2 int) float64 {
	return CalculateExpectedScore(rating1, rating2)
}

// UpdatePlayerRatings обновляет рейтинги обоих игроков после матча
func UpdatePlayerRatings(winner, loser *Player, matchType string) {
	// Получаем K-факторы для обоих игроков
	winnerKFactor := GetKFactor(winner.Rating, winner.GamesPlayed)
	loserKFactor := GetKFactor(loser.Rating, loser.GamesPlayed)
	
	// Применяем модификаторы в зависимости от типа матча
	switch matchType {
	case "tournament":
		// Турнирные матчи более важны
		winnerKFactor *= 1.2
		loserKFactor *= 1.2
	case "casual":
		// Казуальные матчи менее важны
		winnerKFactor *= 0.8
		loserKFactor *= 0.8
	}
	
	// Рассчитываем новые рейтинги
	newWinnerRating, newLoserRating := CalculateRatingChange(
		winner.Rating, 
		loser.Rating, 
		winnerKFactor,
	)
	
	// Обновляем статистику
	winner.Rating = newWinnerRating
	winner.Wins++
	winner.GamesPlayed++
	
	loser.Rating = newLoserRating
	loser.Losses++
	loser.GamesPlayed++
}

// GetRatingTier возвращает тир игрока на основе рейтинга
func GetRatingTier(rating int) string {
	switch {
	case rating < 800:
		return "Bronze"
	case rating < 1200:
		return "Silver"
	case rating < 1600:
		return "Gold"
	case rating < 2000:
		return "Platinum"
	case rating < 2400:
		return "Diamond"
	case rating < 2800:
		return "Master"
	default:
		return "Grandmaster"
	}
}

// GetRatingColor возвращает цвет для отображения рейтинга
func GetRatingColor(rating int) string {
	switch {
	case rating < 800:
		return "#CD7F32" // Bronze
	case rating < 1200:
		return "#C0C0C0" // Silver
	case rating < 1600:
		return "#FFD700" // Gold
	case rating < 2000:
		return "#E5E4E2" // Platinum
	case rating < 2400:
		return "#B9F2FF" // Diamond
	case rating < 2800:
		return "#FF6347" // Master
	default:
		return "#9400D3" // Grandmaster
	}
}

// CalculateRatingRequiredForTier возвращает рейтинг, необходимый для следующего тира
func CalculateRatingRequiredForTier(currentRating int) (string, int) {
	switch {
	case currentRating < 800:
		return "Silver", 800 - currentRating
	case currentRating < 1200:
		return "Gold", 1200 - currentRating
	case currentRating < 1600:
		return "Platinum", 1600 - currentRating
	case currentRating < 2000:
		return "Diamond", 2000 - currentRating
	case currentRating < 2400:
		return "Master", 2400 - currentRating
	case currentRating < 2800:
		return "Grandmaster", 2800 - currentRating
	default:
		return "Max Tier", 0
	}
}

// GetWinRate возвращает процент побед игрока
func GetWinRate(wins, losses int) float64 {
	totalGames := wins + losses
	if totalGames == 0 {
		return 0.0
	}
	return float64(wins) / float64(totalGames) * 100.0
}

// IsRatingGainReasonable проверяет, не слишком ли большое изменение рейтинга
func IsRatingGainReasonable(oldRating, newRating int) bool {
	change := newRating - oldRating
	if change < 0 {
		change = -change
	}
	
	// Максимальное изменение за один матч не должно превышать 50 пунктов
	return change <= 50
}

// CalculatePerformanceRating рассчитывает performance rating для турнира
func CalculatePerformanceRating(playerRating int, results []MatchResult, opponents []Player) int {
	if len(results) == 0 {
		return playerRating
	}
	
	totalScore := 0.0
	totalExpected := 0.0
	
	for i, result := range results {
		if i >= len(opponents) {
			break
		}
		
		var score float64
		if result.WinnerID != 0 { // Если это игрок (не 0 - значит он выиграл)
			score = 1.0
		} else {
			score = 0.0
		}
		
		expected := CalculateExpectedScore(playerRating, opponents[i].Rating)
		
		totalScore += score
		totalExpected += expected
	}
	
	if totalExpected == 0 {
		return playerRating
	}
	
	performance := totalScore / totalExpected
	performanceRating := int(float64(playerRating) * performance)
	
	return performanceRating
}