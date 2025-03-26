package cache

import (
	"go/ast"
	"go/token"
)

func (v *IdentifierVisitor) getPosition(pos token.Pos) Position {
	if pos == token.NoPos {
		return Position{}
	}
	p := v.FSet.Position(pos)
	return Position{
		File:   v.FileName,
		Line:   p.Line,
		Column: p.Column,
	}
}

func (v *IdentifierVisitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}

	switch node := n.(type) {
	case *ast.FuncDecl:
		ident := Identifier{
			Name:     node.Name.Name,
			Type:     "function",
			Position: v.getPosition(node.Pos()),
			Package:  v.PackageName,
		}
		if node.Recv != nil {
			ident.Type = "method"
		}
		v.Identifiers = append(v.Identifiers, ident)

	case *ast.GenDecl:
		for _, spec := range node.Specs {
			switch spec := spec.(type) {
			case *ast.TypeSpec:
				ident := Identifier{
					Name:     spec.Name.Name,
					Type:     "type",
					Position: v.getPosition(spec.Pos()),
					Package:  v.PackageName,
				}

				switch spec.Type.(type) {
				case *ast.StructType:
					ident.Type = "struct"
				case *ast.InterfaceType:
					ident.Type = "interface"
				}
				v.Identifiers = append(v.Identifiers, ident)

			case *ast.ValueSpec:
				var identType string
				switch node.Tok {
				case token.VAR:
					identType = "variable"
				case token.CONST:
					identType = "constant"
				default:
					identType = "unknown"
				}

				for _, id := range spec.Names {
					v.Identifiers = append(v.Identifiers, Identifier{
						Name:     id.Name,
						Type:     identType,
						Position: v.getPosition(id.Pos()),
						Package:  v.PackageName,
					})
				}
			}
		}
	}

	return v
}
