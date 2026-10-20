package loader

import (
	"go/ast"
	"go/constant"
	"go/scanner"
	"go/token"
	"go/types"
	"strings"
)

func (p *Package) checkNoRuntimeInitializers() []error {
	var errs []error
	for _, file := range p.Files {
		for _, decl := range file.Decls {
			decl, ok := decl.(*ast.GenDecl)
			if !ok || decl.Tok != token.VAR {
				continue
			}
			for _, spec := range decl.Specs {
				spec := spec.(*ast.ValueSpec)
				marked := false
				for _, doc := range []*ast.CommentGroup{decl.Doc, spec.Doc} {
					if doc == nil {
						continue
					}
					for _, comment := range doc.List {
						parts := strings.Fields(comment.Text)
						if len(parts) == 0 || parts[0] != "//go:noruntime" {
							continue
						}
						msg := ""
						if len(parts) != 1 {
							msg = "//go:noruntime does not accept parameters"
						} else if len(spec.Values) == 0 {
							msg = "//go:noruntime variable requires an initializer"
						}
						if msg != "" {
							errs = append(errs, scanner.Error{Pos: p.program.fset.Position(comment.Pos()), Msg: msg})
						}
						marked = true
					}
				}
				if marked {
					if p.noRuntimeVars == nil {
						p.noRuntimeVars = make(map[*types.Var]bool)
					}
					for _, name := range spec.Names {
						p.noRuntimeVars[p.info.Defs[name].(*types.Var)] = true
					}
				}
			}
		}
	}
	return errs
}

func (p *Package) lowerNoRuntimeInitializers() {
	if len(p.noRuntimeVars) == 0 || p.NoRuntimeInitializers != nil {
		return
	}
	p.NoRuntimeInitializers = make(map[*ast.FuncLit]string)
	replacements := make(map[ast.Expr]ast.Expr)
	for i, initializer := range p.info.InitOrder {
		marked := false
		for _, variable := range initializer.Lhs {
			marked = marked || p.noRuntimeVars[variable]
		}
		if !marked {
			continue
		}
		pos := initializer.Lhs[0].Pos()
		lhs := make([]ast.Expr, len(initializer.Lhs))
		names := make([]string, len(initializer.Lhs))
		for i, variable := range initializer.Lhs {
			id := &ast.Ident{NamePos: variable.Pos(), Name: variable.Name()}
			lhs[i] = id
			p.info.Uses[id] = variable
			p.info.Types[id] = types.TypeAndValue{Type: variable.Type()}
			names[i] = p.Pkg.Path() + "." + variable.Name()
		}
		falseValue := &ast.Ident{NamePos: pos, Name: "false"}
		p.info.Uses[falseValue] = types.Universe.Lookup("false")
		p.info.Types[falseValue] = types.TypeAndValue{Type: types.Typ[types.Bool], Value: constant.MakeBool(false)}
		boolType := &ast.Ident{NamePos: pos, Name: "bool"}
		p.info.Uses[boolType] = types.Universe.Lookup("bool")
		p.info.Types[boolType] = types.TypeAndValue{Type: types.Typ[types.Bool]}
		fnType := &ast.FuncType{
			Func:    pos,
			Params:  &ast.FieldList{Opening: pos, Closing: pos},
			Results: &ast.FieldList{List: []*ast.Field{{Type: boolType}}},
		}
		signature := types.NewSignatureType(nil, nil, nil, types.NewTuple(), types.NewTuple(types.NewVar(pos, p.Pkg, "", types.Typ[types.Bool])), false)
		p.info.Types[fnType] = types.TypeAndValue{Type: signature}
		fn := &ast.FuncLit{
			Type: fnType,
			Body: &ast.BlockStmt{Lbrace: pos, Rbrace: initializer.Rhs.End(), List: []ast.Stmt{
				&ast.AssignStmt{Lhs: lhs, TokPos: pos, Tok: token.ASSIGN, Rhs: []ast.Expr{initializer.Rhs}},
				&ast.ReturnStmt{Return: pos, Results: []ast.Expr{falseValue}},
			}},
		}
		p.info.Types[fn] = types.TypeAndValue{Type: signature}
		p.NoRuntimeInitializers[fn] = strings.Join(names, ", ")
		call := &ast.CallExpr{Fun: fn, Lparen: pos, Rparen: initializer.Rhs.End()}
		p.info.Types[call] = types.TypeAndValue{Type: types.Typ[types.Bool]}

		// Keep the type checker's initialization order. The blank result lets
		// SSA emit a call containing the complete assignment.
		p.info.InitOrder[i] = &types.Initializer{
			Lhs: []*types.Var{types.NewVar(pos, p.Pkg, "_", types.Typ[types.Bool])},
			Rhs: call,
		}
		replacements[initializer.Rhs] = call
	}
	for _, file := range p.Files {
		for _, decl := range file.Decls {
			decl, ok := decl.(*ast.GenDecl)
			if !ok || decl.Tok != token.VAR {
				continue
			}
			for _, spec := range decl.Specs {
				for i, expr := range spec.(*ast.ValueSpec).Values {
					if replacement := replacements[expr]; replacement != nil {
						spec.(*ast.ValueSpec).Values[i] = replacement
					}
				}
			}
		}
	}
}
