// internal/handlers/helpers.go - исправленная версия с защитой от SQL инъекций
package handlers

import (
	"strconv"
	"strings"
)

// joinStrings объединяет строки с разделителем (только для не-SQL целей)
func joinStrings(strings []string, separator string) string {
	if len(strings) == 0 {
		return ""
	}
	if len(strings) == 1 {
		return strings[0]
	}

	result := strings[0]
	for i := 1; i < len(strings); i++ {
		result += separator + strings[i]
	}
	return result
}

// ❌ ОПАСНО: Эта функция НЕ должна использоваться для SQL запросов!
// Используйте buildSQLConditions или buildParameterizedQuery вместо неё

// buildSQLConditions безопасно строит WHERE условия с placeholder'ами
func buildSQLConditions(conditions []SQLCondition) (string, []interface{}) {
	if len(conditions) == 0 {
		return "", nil
	}

	var whereParts []string
	var args []interface{}
	argIndex := 1

	for _, condition := range conditions {
		whereParts = append(whereParts, condition.Field+" "+condition.Operator+" $"+strconv.Itoa(argIndex))
		args = append(args, condition.Value)
		argIndex++
	}

	return strings.Join(whereParts, " AND "), args
}

// SQLCondition представляет условие для SQL запроса
type SQLCondition struct {
	Field    string      // имя поля (должно быть валидным)
	Operator string      // оператор (=, >, <, LIKE и т.д.)
	Value    interface{} // значение
}

// buildParameterizedQuery безопасно строит SQL запрос с параметрами
func buildParameterizedQuery(baseQuery string, conditions []SQLCondition, orderBy, limit string) (string, []interface{}) {
	query := baseQuery
	var args []interface{}

	if len(conditions) > 0 {
		whereClause, whereArgs := buildSQLConditions(conditions)
		query += " WHERE " + whereClause
		args = append(args, whereArgs...)
	}

	if orderBy != "" {
		// Валидируем ORDER BY (только разрешенные поля)
		if isValidOrderBy(orderBy) {
			query += " ORDER BY " + orderBy
		}
	}

	if limit != "" {
		// Валидируем LIMIT
		if isValidLimit(limit) {
			query += " LIMIT " + limit
		}
	}

	return query, args
}

// isValidOrderBy проверяет, что ORDER BY безопасен
func isValidOrderBy(orderBy string) bool {
	// Список разрешенных полей для сортировки
	allowedFields := map[string]bool{
		"id":         true,
		"name":       true,
		"username":   true,
		"rating":     true,
		"created_at": true,
		"updated_at": true,
		"wins":       true,
		"losses":     true,
		"element":    true,
		"rarity":     true,
		"role":       true,
		"status":     true,
		"round":      true,
	}

	// Разбираем ORDER BY на компоненты
	parts := strings.Fields(strings.ToLower(orderBy))
	if len(parts) == 0 || len(parts) > 2 {
		return false
	}

	// Проверяем поле
	field := parts[0]
	if !allowedFields[field] {
		return false
	}

	// Проверяем направление сортировки
	if len(parts) == 2 {
		direction := parts[1]
		if direction != "asc" && direction != "desc" {
			return false
		}
	}

	return true
}

// isValidLimit проверяет, что LIMIT безопасен
func isValidLimit(limit string) bool {
	// Проверяем, что это число
	if _, err := strconv.Atoi(limit); err != nil {
		return false
	}

	// Проверяем разумные границы
	limitNum, _ := strconv.Atoi(limit)
	return limitNum > 0 && limitNum <= 1000
}

// buildUpdateQuery безопасно строит UPDATE запрос
func buildUpdateQuery(table string, updates map[string]interface{}, whereConditions []SQLCondition) (string, []interface{}) {
	if len(updates) == 0 {
		return "", nil
	}

	var setParts []string
	var args []interface{}
	argIndex := 1

	// Строим SET часть
	for field, value := range updates {
		if isValidFieldName(field) {
			setParts = append(setParts, field+" = $"+strconv.Itoa(argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	if len(setParts) == 0 {
		return "", nil
	}

	query := "UPDATE " + table + " SET " + strings.Join(setParts, ", ")

	// Добавляем WHERE условия
	if len(whereConditions) > 0 {
		whereClause, whereArgs := buildSQLConditionsWithIndex(whereConditions, argIndex)
		query += " WHERE " + whereClause
		args = append(args, whereArgs...)
	}

	return query, args
}

// buildSQLConditionsWithIndex строит условия начиная с определенного индекса
func buildSQLConditionsWithIndex(conditions []SQLCondition, startIndex int) (string, []interface{}) {
	if len(conditions) == 0 {
		return "", nil
	}

	var whereParts []string
	var args []interface{}
	argIndex := startIndex

	for _, condition := range conditions {
		whereParts = append(whereParts, condition.Field+" "+condition.Operator+" $"+strconv.Itoa(argIndex))
		args = append(args, condition.Value)
		argIndex++
	}

	return strings.Join(whereParts, " AND "), args
}

// isValidFieldName проверяет, что имя поля безопасно
func isValidFieldName(field string) bool {
	// Список разрешенных полей для UPDATE
	allowedFields := map[string]bool{
		"name":        true,
		"description": true,
		"max_players": true,
		"is_private":  true,
		"password":    true,
		"updated_at":  true,
		"username":    true,
		"email":       true,
		"rating":      true,
		"wins":        true,
		"losses":      true,
		"element":     true,
		"rarity":      true,
		"role":        true,
		"is_active":   true,
		"image_url":   true,
		"status":      true,
		"winner_id":   true,
		"player1_id":  true,
		"player2_id":  true,
		"content":     true,
		"last_seen":   true,
	}

	return allowedFields[field]
}

// QueryBuilder безопасный построитель запросов
type QueryBuilder struct {
	baseQuery  string
	conditions []SQLCondition
	orderBy    string
	limit      string
	offset     string
}

// NewQueryBuilder создает новый построитель запросов
func NewQueryBuilder(baseQuery string) *QueryBuilder {
	return &QueryBuilder{
		baseQuery: baseQuery,
	}
}

// Where добавляет WHERE условие
func (qb *QueryBuilder) Where(field, operator string, value interface{}) *QueryBuilder {
	if isValidFieldName(field) && isValidOperator(operator) {
		qb.conditions = append(qb.conditions, SQLCondition{
			Field:    field,
			Operator: operator,
			Value:    value,
		})
	}
	return qb
}

// OrderBy добавляет сортировку
func (qb *QueryBuilder) OrderBy(orderBy string) *QueryBuilder {
	if isValidOrderBy(orderBy) {
		qb.orderBy = orderBy
	}
	return qb
}

// Limit добавляет ограничение
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	if limit > 0 && limit <= 1000 {
		qb.limit = strconv.Itoa(limit)
	}
	return qb
}

// Offset добавляет смещение
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	if offset >= 0 {
		qb.offset = strconv.Itoa(offset)
	}
	return qb
}

// Build строит финальный запрос
func (qb *QueryBuilder) Build() (string, []interface{}) {
	query := qb.baseQuery
	var args []interface{}
	argIndex := 1

	// Добавляем WHERE условия
	if len(qb.conditions) > 0 {
		var whereParts []string
		for _, condition := range qb.conditions {
			whereParts = append(whereParts, condition.Field+" "+condition.Operator+" $"+strconv.Itoa(argIndex))
			args = append(args, condition.Value)
			argIndex++
		}
		query += " WHERE " + strings.Join(whereParts, " AND ")
	}

	// Добавляем ORDER BY
	if qb.orderBy != "" {
		query += " ORDER BY " + qb.orderBy
	}

	// Добавляем LIMIT
	if qb.limit != "" {
		query += " LIMIT " + qb.limit
	}

	// Добавляем OFFSET
	if qb.offset != "" {
		query += " OFFSET " + qb.offset
	}

	return query, args
}

// isValidOperator проверяет оператор
func isValidOperator(operator string) bool {
	allowedOperators := map[string]bool{
		"=":      true,
		"!=":     true,
		">":      true,
		"<":      true,
		">=":     true,
		"<=":     true,
		"LIKE":   true,
		"ILIKE":  true,
		"IN":     true,
		"NOT IN": true,
	}
	return allowedOperators[strings.ToUpper(operator)]
}

// contains проверяет, содержит ли строка подстроку
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// trimSpace удаляет пробелы в начале и конце строки
func trimSpace(s string) string {
	return strings.TrimSpace(s)
}

// validateStringLength проверяет длину строки
func validateStringLength(s string, min, max int) bool {
	length := len(s)
	return length >= min && length <= max
}

// isEmptyString проверяет, является ли строка пустой
func isEmptyString(s string) bool {
	return strings.TrimSpace(s) == ""
}

// sanitizeSQL очищает строку от потенциально опасных символов для SQL
func sanitizeSQL(input string) string {
	// Удаляем потенциально опасные символы
	dangerous := []string{
		";", "--", "/*", "*/", "xp_", "sp_",
		"DROP", "DELETE", "INSERT", "UPDATE",
		"CREATE", "ALTER", "EXEC", "EXECUTE",
	}

	result := input
	for _, danger := range dangerous {
		result = strings.ReplaceAll(result, danger, "")
		result = strings.ReplaceAll(result, strings.ToLower(danger), "")
		result = strings.ReplaceAll(result, strings.ToUpper(danger), "")
	}

	return result
}
