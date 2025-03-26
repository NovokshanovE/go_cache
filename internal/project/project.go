package project

import (
	"encoding/json"
	"fmt"
	mycache "go-cache/internal/cache"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func AnalyzeProject(projectPath string) {
	var cache mycache.Snapshot

	fset := token.NewFileSet()
	pkgFiles := make(map[string][]*ast.File)

	fileCount := 0

	walkFunc := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			base := filepath.Base(path)
			if base == "vendor" || strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, parser.AllErrors)

		pkgName := file.Name.Name
		pkgFiles[pkgName] = append(pkgFiles[pkgName], file)
		fileCount++

		return nil
	}

	err := filepath.WalkDir(projectPath, walkFunc)

	if err != nil {
		log.Fatalf("Ошибка при обходе директории: %v", err)
	}

	if fileCount == 0 {
		log.Fatalf("Не найдено Go файлов в %s", projectPath)
	}
	cache.Packages = make(map[string]mycache.Package)

	for pkgName, files := range pkgFiles {
		info := &types.Info{
			Types:      make(map[ast.Expr]types.TypeAndValue),
			Defs:       make(map[*ast.Ident]types.Object),
			Uses:       make(map[*ast.Ident]types.Object),
			Implicits:  make(map[ast.Node]types.Object),
			Selections: make(map[*ast.SelectorExpr]*types.Selection),
			Scopes:     make(map[ast.Node]*types.Scope),
		}

		conf := types.Config{
			Importer: importer.Default(),
		}

		_, err := conf.Check(pkgName, fset, files, info)
		if err != nil {
			fmt.Printf("Предупреждение при проверке типов: %v\n", err)
		}
		typeInfo := mycache.ConvertTypesInfoToSerializable(info, fset)

		if pkg, ok := cache.Packages[pkgName]; ok {
			pkg.TypeInfos = append(pkg.TypeInfos, *typeInfo)
			cache.Packages[pkgName] = pkg
		} else {
			cache.Packages[pkgName] = mycache.Package{
				Name:      pkgName,
				TypeInfos: []mycache.TypeInfo{*typeInfo},
			}

		}

	}

	// err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
	// 	if err != nil {
	// 		return err
	// 	}

	// 	if info.IsDir() || filepath.Ext(path) != ".go" {
	// 		return nil
	// 	}

	// 	fset := token.NewFileSet()
	// 	file, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	// 	if err != nil {
	// 		return nil
	// 	}

	// 	visitor := &mycache.IdentifierVisitor{
	// 		PackageName: file.Name.Name,
	// 		FileName:    path,
	// 		FSet:        fset,
	// 		Identifiers: []mycache.Identifier{},
	// 	}

	// 	ast.Walk(visitor, file)
	// 	// cache.Identifiers = append(cache.Identifiers, visitor.Identifiers...)

	// 	return nil
	// })

	// if err != nil {
	// 	log.Printf("Error processing project %s: %v", projectPath, err)

	// }

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		log.Printf("Error marshaling cache: %v", err)

	}

	outputFile := fmt.Sprintf(".cache/new_cache.json")
	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		log.Printf("Error writing cache file: %v", err)

	}

	fmt.Printf("Created cache for prokect %s\n", projectPath)
}
