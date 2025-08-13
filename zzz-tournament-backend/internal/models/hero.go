// internal/models/hero.go
package models

// Hero модель героя
type Hero struct {
	ID          int    `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Element     string `json:"element" db:"element"`
	Rarity      string `json:"rarity" db:"rarity"`
	Role        string `json:"role" db:"role"`
	Description string `json:"description" db:"description"`
	ImageURL    string `json:"image_url" db:"image_url"`
	IsActive    bool   `json:"is_active" db:"is_active"`
}

// HeroElement константы для элементов героев
const (
	ElementFire     = "fire"
	ElementIce      = "ice"
	ElementElectric = "electric"
	ElementPhysical = "physical"
	ElementEther    = "ether"
)

// HeroRarity константы для редкости героев
const (
	RarityA = "A"
	RarityS = "S"
)

// HeroRole константы для ролей героев
const (
	RoleAttacker = "attacker"
	RoleDefender = "defender"
	RoleStunner  = "stunner"
	RoleSupport  = "support"
	RoleAnomaly  = "anomaly"
)

// IsValidElement проверяет валидность элемента
func IsValidElement(element string) bool {
	switch element {
	case ElementFire, ElementIce, ElementElectric, ElementPhysical, ElementEther:
		return true
	default:
		return false
	}
}

// IsValidRarity проверяет валидность редкости
func IsValidRarity(rarity string) bool {
	switch rarity {
	case RarityA, RarityS:
		return true
	default:
		return false
	}
}

// IsValidRole проверяет валидность роли
func IsValidRole(role string) bool {
	switch role {
	case RoleAttacker, RoleDefender, RoleStunner, RoleSupport, RoleAnomaly:
		return true
	default:
		return false
	}
}
