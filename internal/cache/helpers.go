package cache

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
)

// ConvertTypesInfoToSerializable преобразует стандартную структуру Go types.Info
// в сериализуемую структуру TypeInfo
func ConvertTypesInfoToSerializable(info *types.Info, fset *token.FileSet) *TypeInfo {
	result := &TypeInfo{
		TypesMap:      make(map[string]TypeMetadata),
		DefsMap:       make(map[string]ObjectMetadata),
		UsesMap:       make(map[string]ObjectMetadata),
		ImplicitsMap:  make(map[string]ObjectMetadata),
		SelectionsMap: make(map[string]SelectionMetadata),
		ScopesMap:     make(map[string]ScopeMetadata),
	}

	// Преобразование карты Types
	for expr, tv := range info.Types {
		key := exprToString(expr)
		result.TypesMap[key] = TypeMetadata{
			TypeString: tv.Type.String(),
			Value:      fmt.Sprintf("%v", tv.Value),
			IsValue:    tv.IsValue(),
		}
	}

	// Преобразование карты Defs
	for ident, obj := range info.Defs {
		if obj != nil {
			key := identToString(ident)
			result.DefsMap[key] = objectToMetadata(obj, fset)
		}
	}

	// Преобразование карты Uses
	for ident, obj := range info.Uses {
		if obj != nil {
			key := identToString(ident)
			result.UsesMap[key] = objectToMetadata(obj, fset)
		}
	}

	// Преобразование карты Implicits
	for node, obj := range info.Implicits {
		if obj != nil {
			key := nodeToString(node)
			result.ImplicitsMap[key] = objectToMetadata(obj, fset)
		}
	}

	// Преобразование карты Selections
	for selExpr, sel := range info.Selections {
		key := selectorExprToString(selExpr)
		result.SelectionsMap[key] = SelectionMetadata{
			Recv:     sel.Recv().String(),
			Expr:     selExpr.Sel.Name,
			Kind:     selectionKindToString(sel.Kind()),
			Indirect: sel.Indirect(),
		}
	}

	// Преобразование карты Scopes
	for node, scope := range info.Scopes {
		key := nodeToString(node)
		var parent string
		if scope.Parent() != nil {
			parent = fmt.Sprintf("%p", scope.Parent()) // Используем адрес указателя как уникальный идентификатор
		}

		names := []string{}
		// for _, name := range scope.Names() {
		names = append(names, scope.Names()...)
		// }

		result.ScopesMap[key] = ScopeMetadata{
			Parent: parent,
			Names:  names,
		}
	}

	return result
}

// exprToString преобразует выражение AST в строковое представление
func exprToString(expr ast.Expr) string {
	var buf bytes.Buffer
	printer.Fprint(&buf, token.NewFileSet(), expr)
	return fmt.Sprintf("%p:%s", expr, buf.String())
}

// identToString преобразует идентификатор AST в строковое представление
func identToString(ident *ast.Ident) string {
	return fmt.Sprintf("%p:%s", ident, ident.Name)
}

// nodeToString преобразует узел AST в строковое представление
func nodeToString(node ast.Node) string {
	return fmt.Sprintf("%p:%T", node, node)
}

// selectorExprToString преобразует выражение селектора в строку
func selectorExprToString(sel *ast.SelectorExpr) string {
	var buf bytes.Buffer
	printer.Fprint(&buf, token.NewFileSet(), sel)
	return fmt.Sprintf("%p:%s", sel, buf.String())
}

// objectToMetadata преобразует types.Object в ObjectMetadata
func objectToMetadata(obj types.Object, fset *token.FileSet) ObjectMetadata {
	metadata := ObjectMetadata{
		Name: obj.Name(),
	}

	if obj.Type() != nil {
		metadata.Type = obj.Type().String()
	}

	if obj.Pos().IsValid() {
		position := fset.Position(obj.Pos())
		metadata.Position = fmt.Sprintf("%s:%d:%d", position.Filename, position.Line, position.Column)
	}

	return metadata
}

// selectionKindToString преобразует types.SelectionKind в строку
func selectionKindToString(kind types.SelectionKind) string {
	switch kind {
	case types.FieldVal:
		return "FieldVal"
	case types.MethodVal:
		return "MethodVal"
	case types.MethodExpr:
		return "MethodExpr"
	default:
		return "Unknown"
	}
}
