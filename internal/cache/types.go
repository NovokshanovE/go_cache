package cache

import (
	"go/token"
	"time"
)

// Snapshot - основная структура для хранения всего кэша
type Snapshot struct {
	LastUpdated time.Time          `json:"lastUpdated"`
	ProjectPath string             `json:"projectPath"`
	Packages    map[string]Package `json:"packages"`
}

// Package - структура для хранения информации о пакете
type Package struct {
	Name        string       `json:"name"`
	Identifiers []Identifier `json:"identifiers"`
	TypeInfos   []TypeInfo   `json:"typeInfo"`
}

// Identifier - информация об идентификаторе
type Identifier struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Position Position `json:"position"`
	Package  string   `json:"package"`
}

// Position - позиция идентификатора в файле
type Position struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Column int    `json:"column"`
}

// TypeInfo - сериализуемая структура для хранения информации о типах
type TypeInfo struct {
	TypesMap      map[string]TypeMetadata      `json:"typesMap"`
	DefsMap       map[string]ObjectMetadata    `json:"defsMap"`
	UsesMap       map[string]ObjectMetadata    `json:"usesMap"`
	ImplicitsMap  map[string]ObjectMetadata    `json:"implicitsMap"`
	SelectionsMap map[string]SelectionMetadata `json:"selectionsMap"`
	ScopesMap     map[string]ScopeMetadata     `json:"scopesMap"`
}

// TypeMetadata - метаданные о типе
type TypeMetadata struct {
	TypeString string `json:"typeString"`
	Value      string `json:"value,omitempty"`
	IsValue    bool   `json:"isValue"`
}

// ObjectMetadata - метаданные об объекте
type ObjectMetadata struct {
	Name     string `json:"name"`
	Type     string `json:"type,omitempty"`
	Position string `json:"position,omitempty"`
}

// SelectionMetadata - метаданные о выборе
type SelectionMetadata struct {
	Recv     string `json:"recv"`
	Expr     string `json:"expr"`
	Kind     string `json:"kind"`
	Indirect bool   `json:"indirect"`
}

// ScopeMetadata - метаданные о области видимости
type ScopeMetadata struct {
	Parent string   `json:"parent,omitempty"`
	Names  []string `json:"names"`
}

// IdentifierVisitor - посетитель для обхода AST
type IdentifierVisitor struct {
	Identifiers []Identifier
	PackageName string
	FileName    string
	FSet        *token.FileSet
}
