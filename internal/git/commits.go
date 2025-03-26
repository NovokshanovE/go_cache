package git

import (
	"encoding/json"
	"fmt"
	mycache "go-cache/internal/cache"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func AnalizeCommit(repoPath string) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		log.Fatal(err)
	}

	commitIter, err := repo.Log(&git.LogOptions{All: true})
	if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(".cache", 0755); err != nil {
		log.Fatalf("Не удалось создать директорию кеша: %v", err)
	}

	for {
		commit, err := commitIter.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Анализ коммита %s (%s)...\n", commit.Hash.String()[:7], commit.Message)

		tree, err := commit.Tree()
		if err != nil {
			log.Printf("Ошибка при получении дерева коммита %s: %v", commit.Hash, err)
			continue
		}

		var cache mycache.Snapshot
		cache.Packages = make(map[string]mycache.Package)

		pkgFiles := make(map[string][]*ast.File)
		fset := token.NewFileSet()
		fileCount := 0

		err = tree.Files().ForEach(func(f *object.File) error {
			if !strings.HasSuffix(f.Name, ".go") || strings.HasSuffix(f.Name, "_test.go") {
				return nil
			}

			parts := strings.Split(f.Name, string(filepath.Separator))
			for _, part := range parts {
				if part == "vendor" || strings.HasPrefix(part, ".") {
					return nil
				}
			}

			content, err := f.Contents()
			if err != nil {
				return err
			}

			file, err := parser.ParseFile(fset, f.Name, content, parser.AllErrors)
			if err != nil {
				log.Printf("Предупреждение при парсинге файла %s: %v", f.Name, err)
				return nil
			}

			pkgName := file.Name.Name
			pkgFiles[pkgName] = append(pkgFiles[pkgName], file)
			fileCount++

			return nil
		})

		if err != nil {
			log.Printf("Ошибка при обработке коммита %s: %v", commit.Hash, err)
			continue
		}

		if fileCount == 0 {
			log.Printf("В коммите %s не найдено Go файлов", commit.Hash)
			continue
		}

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
				Error:    func(err error) {},
			}

			_, err := conf.Check(pkgName, fset, files, info)
			if err != nil {
				log.Printf("Предупреждение при проверке типов для пакета %s: %v", pkgName, err)
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

		data, err := json.MarshalIndent(cache, "", "  ")
		if err != nil {
			log.Printf("Ошибка при сериализации кеша: %v", err)
			continue
		}

		outputFile := fmt.Sprintf(".cache/%s.json", commit.Hash)
		if err := os.WriteFile(outputFile, data, 0644); err != nil {
			log.Printf("Ошибка при записи файла кеша: %v", err)
			continue
		}

		fmt.Printf("Создан кеш для коммита %s\n", commit.Hash.String()[:7])
	}
}
