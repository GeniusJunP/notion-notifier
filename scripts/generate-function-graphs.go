package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type FuncNode struct {
	Full     string
	Lang     string
	Group    string
	File     string
	Name     string
	BaseName string
	Pkg      string
	IsMethod bool
}

type Edge struct {
	From string
	To   string
}

type goRec struct {
	node    FuncNode
	body    *ast.BlockStmt
	imports map[string]string // alias -> package name
}

type feImport struct {
	TargetFile string
	OrigName   string
	IsNS       bool
}

type feRec struct {
	node    FuncNode
	body    string
	imports map[string]feImport // local binding -> import metadata
}

type shRec struct {
	node FuncNode
	body string
}

func main() {
	root, err := os.Getwd()
	must(err)

	goNodes, goEdges, err := scanGo(root)
	must(err)
	feNodes, feEdges, err := scanFrontend(root)
	must(err)
	shNodes, shEdges, err := scanShell(root)
	must(err)

	nodes := append(append(goNodes, feNodes...), shNodes...)
	edges := dedupEdges(append(append(goEdges, feEdges...), shEdges...))

	must(writeDetailed(filepath.Join(root, "docs", "function-graph-detailed.md"), root, nodes, edges))
	must(writeIntegrated(filepath.Join(root, "docs", "function-graph-integrated.md"), root, nodes, edges, len(goNodes), len(feNodes), len(shNodes)))

	fmt.Printf("generated docs/function-graph-detailed.md and docs/function-graph-integrated.md\n")
	fmt.Printf("functions=%d (go=%d fe=%d script=%d) edges=%d\n", len(nodes), len(goNodes), len(feNodes), len(shNodes), len(edges))
}

func scanGo(root string) ([]FuncNode, []Edge, error) {
	files := []string{}
	for _, dir := range []string{"cmd", "internal"} {
		base := filepath.Join(root, dir)
		files = append(files, walkFiles(base, ".go")...)
	}
	sort.Strings(files)

	recs := []goRec{}
	byPkgBaseAll := map[string][]FuncNode{}
	byPkgBaseFuncs := map[string][]FuncNode{}

	for _, file := range files {
		fset := token.NewFileSet()
		af, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			return nil, nil, fmt.Errorf("parse go %s: %w", file, err)
		}

		imports := map[string]string{}
		for _, imp := range af.Imports {
			p, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			pkg := filepath.Base(p)
			alias := pkg
			if imp.Name != nil {
				if imp.Name.Name == "_" || imp.Name.Name == "." {
					continue
				}
				alias = imp.Name.Name
			}
			imports[alias] = pkg
		}

		for _, d := range af.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok {
				continue
			}
			base := fn.Name.Name
			name := base
			isMethod := false
			if fn.Recv != nil && len(fn.Recv.List) > 0 {
				isMethod = true
				recv := recvName(fn.Recv.List[0].Type)
				if recv != "" {
					name = "(" + recv + ")." + base
				}
			}
			rel := filepath.ToSlash(strings.TrimPrefix(file, root+string(filepath.Separator)))
			n := FuncNode{
				Full:     fmt.Sprintf("go|%s|%s", rel, name),
				Lang:     "go",
				Group:    "go:" + af.Name.Name,
				File:     file,
				Name:     name,
				BaseName: base,
				Pkg:      af.Name.Name,
				IsMethod: isMethod,
			}
			recs = append(recs, goRec{node: n, body: fn.Body, imports: imports})
			key := af.Name.Name + "::" + base
			byPkgBaseAll[key] = append(byPkgBaseAll[key], n)
			if !isMethod {
				byPkgBaseFuncs[key] = append(byPkgBaseFuncs[key], n)
			}
		}
	}

	nodes := make([]FuncNode, 0, len(recs))
	for _, r := range recs {
		nodes = append(nodes, r.node)
	}

	edges := []Edge{}
	seen := map[string]bool{}
	for _, r := range recs {
		if r.body == nil {
			continue
		}
		idCalls, selCalls := collectGoCalls(r.body)
		for _, name := range idCalls {
			for _, t := range byPkgBaseFuncs[r.node.Pkg+"::"+name] {
				k := r.node.Full + "->" + t.Full
				if seen[k] {
					continue
				}
				seen[k] = true
				edges = append(edges, Edge{From: r.node.Full, To: t.Full})
			}
		}
		for _, sc := range selCalls {
			pkg, ok := r.imports[sc.left]
			if !ok {
				continue
			}
			for _, t := range byPkgBaseFuncs[pkg+"::"+sc.right] {
				k := r.node.Full + "->" + t.Full
				if seen[k] {
					continue
				}
				seen[k] = true
				edges = append(edges, Edge{From: r.node.Full, To: t.Full})
			}
		}

		// If a method is called via function value style (Type.Method), ident mapping can miss it.
		// Conservative fallback: selector target name inside same package may reference helpers.
		for _, sc := range selCalls {
			for _, t := range byPkgBaseAll[r.node.Pkg+"::"+sc.right] {
				if t.IsMethod {
					continue
				}
				k := r.node.Full + "->" + t.Full
				if seen[k] {
					continue
				}
				seen[k] = true
				edges = append(edges, Edge{From: r.node.Full, To: t.Full})
			}
		}
	}

	return nodes, edges, nil
}

type goSelCall struct {
	left  string
	right string
}

func collectGoCalls(body *ast.BlockStmt) ([]string, []goSelCall) {
	ids := map[string]bool{}
	selKeys := map[string]goSelCall{}
	ast.Inspect(body, func(n ast.Node) bool {
		c, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fn := c.Fun.(type) {
		case *ast.Ident:
			ids[fn.Name] = true
		case *ast.SelectorExpr:
			left, ok := fn.X.(*ast.Ident)
			if !ok {
				return true
			}
			k := left.Name + "::" + fn.Sel.Name
			selKeys[k] = goSelCall{left: left.Name, right: fn.Sel.Name}
		}
		return true
	})
	idList := make([]string, 0, len(ids))
	for k := range ids {
		idList = append(idList, k)
	}
	sort.Strings(idList)
	selList := make([]goSelCall, 0, len(selKeys))
	for _, v := range selKeys {
		selList = append(selList, v)
	}
	sort.Slice(selList, func(i, j int) bool {
		if selList[i].left == selList[j].left {
			return selList[i].right < selList[j].right
		}
		return selList[i].left < selList[j].left
	})
	return idList, selList
}

func scanFrontend(root string) ([]FuncNode, []Edge, error) {
	base := filepath.Join(root, "web", "src")
	files := []string{}
	files = append(files, walkFiles(base, ".ts", ".svelte")...)
	sort.Strings(files)

	declRe := regexp.MustCompile(`(?m)^\s*(?:export\s+)?(?:async\s+)?function\s+([A-Za-z_][A-Za-z0-9_]*)\s*(?:<[^\n\r>]*>)?\s*\(`)
	callRe := regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	memberCallRe := regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\.([A-Za-z_][A-Za-z0-9_]*)\s*\(`)

	recs := []feRec{}
	byFileFunc := map[string][]FuncNode{}
	byFileBase := map[string]map[string][]FuncNode{}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, nil, err
		}
		imports := parseFEImports(file, data)
		matches := declRe.FindAllSubmatchIndex(data, -1)
		for _, m := range matches {
			name := string(data[m[2]:m[3]])
			braceOpen := indexNextByte(data, '{', m[1])
			if braceOpen < 0 {
				continue
			}
			braceClose := matchBrace(data, braceOpen)
			if braceClose < 0 || braceClose <= braceOpen {
				continue
			}
			body := string(data[braceOpen+1 : braceClose])
			rel := filepath.ToSlash(strings.TrimPrefix(file, root+string(filepath.Separator)))
			n := FuncNode{
				Full:     fmt.Sprintf("fe|%s|%s", rel, name),
				Lang:     "fe",
				Group:    "fe:" + rel,
				File:     file,
				Name:     name,
				BaseName: name,
			}
			recs = append(recs, feRec{node: n, body: body, imports: imports})
			byFileFunc[file] = append(byFileFunc[file], n)
			if byFileBase[file] == nil {
				byFileBase[file] = map[string][]FuncNode{}
			}
			byFileBase[file][name] = append(byFileBase[file][name], n)
		}
	}

	nodes := make([]FuncNode, 0, len(recs))
	for _, r := range recs {
		nodes = append(nodes, r.node)
	}

	edges := []Edge{}
	seen := map[string]bool{}
	reserved := map[string]bool{
		"if": true, "for": true, "while": true, "switch": true, "catch": true,
		"function": true, "return": true, "typeof": true, "await": true, "new": true,
	}

	for _, r := range recs {
		localCalls := map[string]bool{}
		for _, m := range callRe.FindAllStringSubmatch(r.body, -1) {
			if len(m) < 2 {
				continue
			}
			name := m[1]
			if reserved[name] {
				continue
			}
			localCalls[name] = true
		}

		for name := range localCalls {
			for _, t := range byFileBase[r.node.File][name] {
				k := r.node.Full + "->" + t.Full
				if seen[k] {
					continue
				}
				seen[k] = true
				edges = append(edges, Edge{From: r.node.Full, To: t.Full})
			}

			if imp, ok := r.imports[name]; ok && imp.TargetFile != "" {
				targetName := imp.OrigName
				if targetName == "" || imp.IsNS {
					continue
				}
				for _, t := range byFileBase[imp.TargetFile][targetName] {
					k := r.node.Full + "->" + t.Full
					if seen[k] {
						continue
					}
					seen[k] = true
					edges = append(edges, Edge{From: r.node.Full, To: t.Full})
				}
			}
		}

		for _, m := range memberCallRe.FindAllStringSubmatch(r.body, -1) {
			if len(m) < 3 {
				continue
			}
			obj := m[1]
			method := m[2]
			imp, ok := r.imports[obj]
			if !ok || imp.TargetFile == "" {
				continue
			}
			for _, t := range byFileBase[imp.TargetFile][method] {
				k := r.node.Full + "->" + t.Full
				if seen[k] {
					continue
				}
				seen[k] = true
				edges = append(edges, Edge{From: r.node.Full, To: t.Full})
			}
		}
	}

	return nodes, edges, nil
}

func parseFEImports(file string, data []byte) map[string]feImport {
	imports := map[string]feImport{}
	s := string(data)

	namedRe := regexp.MustCompile(`(?m)^\s*import\s*\{([^}]*)\}\s*from\s*["']([^"']+)["'];?`)
	namespaceRe := regexp.MustCompile(`(?m)^\s*import\s+\*\s+as\s+([A-Za-z_][A-Za-z0-9_]*)\s+from\s+["']([^"']+)["'];?`)
	defaultRe := regexp.MustCompile(`(?m)^\s*import\s+([A-Za-z_][A-Za-z0-9_]*)\s+from\s+["']([^"']+)["'];?`)

	for _, m := range namespaceRe.FindAllStringSubmatch(s, -1) {
		if len(m) < 3 {
			continue
		}
		local := strings.TrimSpace(m[1])
		target := resolveRelativeImport(file, strings.TrimSpace(m[2]))
		imports[local] = feImport{TargetFile: target, IsNS: true}
	}

	for _, m := range namedRe.FindAllStringSubmatch(s, -1) {
		if len(m) < 3 {
			continue
		}
		chunk := m[1]
		target := resolveRelativeImport(file, strings.TrimSpace(m[2]))
		parts := strings.Split(chunk, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			p = strings.TrimPrefix(p, "type ")
			if p == "" {
				continue
			}
			orig := p
			local := p
			if strings.Contains(p, " as ") {
				sp := strings.SplitN(p, " as ", 2)
				orig = strings.TrimSpace(sp[0])
				local = strings.TrimSpace(sp[1])
			}
			imports[local] = feImport{TargetFile: target, OrigName: orig}
		}
	}

	for _, m := range defaultRe.FindAllStringSubmatch(s, -1) {
		if len(m) < 3 {
			continue
		}
		local := strings.TrimSpace(m[1])
		target := resolveRelativeImport(file, strings.TrimSpace(m[2]))
		if _, exists := imports[local]; exists {
			continue
		}
		imports[local] = feImport{TargetFile: target, OrigName: "default"}
	}

	return imports
}

func resolveRelativeImport(currentFile, spec string) string {
	if !strings.HasPrefix(spec, ".") {
		return ""
	}
	dir := filepath.Dir(currentFile)
	base := filepath.Clean(filepath.Join(dir, spec))
	candidates := []string{base, base + ".ts", base + ".svelte", filepath.Join(base, "index.ts"), filepath.Join(base, "index.svelte")}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && !info.IsDir() {
			return c
		}
	}
	return ""
}

func scanShell(root string) ([]FuncNode, []Edge, error) {
	files := walkFiles(filepath.Join(root, "scripts"), ".sh", ".ps1")
	sort.Strings(files)
	declRe := regexp.MustCompile(`(?m)^\s*([A-Za-z_][A-Za-z0-9_]*)\s*\(\)\s*\{`)
	callRe := regexp.MustCompile(`\b([A-Za-z_][A-Za-z0-9_]*)\b`)

	recs := []shRec{}
	byFile := map[string]map[string][]FuncNode{}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, nil, err
		}
		for _, m := range declRe.FindAllSubmatchIndex(data, -1) {
			name := string(data[m[2]:m[3]])
			braceOpen := indexNextByte(data, '{', m[1]-1)
			if braceOpen < 0 {
				continue
			}
			braceClose := matchBrace(data, braceOpen)
			if braceClose < 0 || braceClose <= braceOpen {
				continue
			}
			body := string(data[braceOpen+1 : braceClose])
			rel := filepath.ToSlash(strings.TrimPrefix(file, root+string(filepath.Separator)))
			n := FuncNode{Full: fmt.Sprintf("script|%s|%s", rel, name), Lang: "script", Group: "script:" + rel, File: file, Name: name, BaseName: name}
			recs = append(recs, shRec{node: n, body: body})
			if byFile[file] == nil {
				byFile[file] = map[string][]FuncNode{}
			}
			byFile[file][name] = append(byFile[file][name], n)
		}
	}

	nodes := make([]FuncNode, 0, len(recs))
	for _, r := range recs {
		nodes = append(nodes, r.node)
	}

	edges := []Edge{}
	seen := map[string]bool{}
	for _, r := range recs {
		for _, m := range callRe.FindAllStringSubmatch(r.body, -1) {
			if len(m) < 2 {
				continue
			}
			name := m[1]
			for _, t := range byFile[r.node.File][name] {
				k := r.node.Full + "->" + t.Full
				if seen[k] {
					continue
				}
				seen[k] = true
				edges = append(edges, Edge{From: r.node.Full, To: t.Full})
			}
		}
	}
	return nodes, edges, nil
}

func writeDetailed(outPath, root string, nodes []FuncNode, edges []Edge) error {
	nodesByGroup := map[string][]FuncNode{}
	for _, n := range nodes {
		nodesByGroup[n.Group] = append(nodesByGroup[n.Group], n)
	}
	groups := sortedGroups(nodesByGroup)

	edgesByGroup := map[string][]Edge{}
	nodeByFull := map[string]FuncNode{}
	for _, n := range nodes {
		nodeByFull[n.Full] = n
	}
	for _, e := range edges {
		from, ok1 := nodeByFull[e.From]
		to, ok2 := nodeByFull[e.To]
		if !ok1 || !ok2 || from.Group != to.Group {
			continue
		}
		edgesByGroup[from.Group] = append(edgesByGroup[from.Group], e)
	}

	for g := range nodesByGroup {
		sortNodes(nodesByGroup[g])
		sortEdges(edgesByGroup[g])
	}

	var b strings.Builder
	b.WriteString("# Function-Level Detailed Graph (All Functions)\n\n")
	b.WriteString("Generated: 2026-02-17 JST\n\n")
	goCount, feCount, shCount := countByLang(nodes)
	b.WriteString("## Coverage\n")
	b.WriteString(fmt.Sprintf("- total functions: %d\n", len(nodes)))
	b.WriteString(fmt.Sprintf("- go functions (including tests): %d\n", goCount))
	b.WriteString(fmt.Sprintf("- frontend functions (ts+svelte): %d\n", feCount))
	b.WriteString(fmt.Sprintf("- script functions: %d\n", shCount))
	b.WriteString("- edge rule: static same-group call references (cross-group links are shown in the integrated graph).\n\n")

	b.WriteString("## Group Index\n")
	b.WriteString("| Group | Functions | Edges |\n")
	b.WriteString("|---|---:|---:|\n")
	for _, g := range groups {
		b.WriteString(fmt.Sprintf("| `%s` | %d | %d |\n", g, len(nodesByGroup[g]), len(edgesByGroup[g])))
	}
	b.WriteString("\n")

	for _, g := range groups {
		b.WriteString(fmt.Sprintf("## %s\n\n", g))
		b.WriteString("```mermaid\n")
		b.WriteString("flowchart TD\n")
		idByFull := map[string]string{}
		for i, n := range nodesByGroup[g] {
			id := fmt.Sprintf("n%d", i+1)
			idByFull[n.Full] = id
			rel := filepath.ToSlash(strings.TrimPrefix(n.File, root+string(filepath.Separator)))
			b.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", id, mermaidEscape(rel+":"+n.Name)))
		}
		for _, e := range edgesByGroup[g] {
			from, ok1 := idByFull[e.From]
			to, ok2 := idByFull[e.To]
			if !ok1 || !ok2 {
				continue
			}
			b.WriteString(fmt.Sprintf("  %s --> %s\n", from, to))
		}
		b.WriteString("```\n\n")
	}

	return os.WriteFile(outPath, []byte(b.String()), 0644)
}

func writeIntegrated(outPath, root string, nodes []FuncNode, edges []Edge, goCount, feCount, shCount int) error {
	nodesByGroup := map[string][]FuncNode{}
	for _, n := range nodes {
		nodesByGroup[n.Group] = append(nodesByGroup[n.Group], n)
	}
	groups := sortedGroups(nodesByGroup)
	for g := range nodesByGroup {
		sortNodes(nodesByGroup[g])
	}
	sortEdges(edges)

	nodeByFull := map[string]FuncNode{}
	for _, n := range nodes {
		nodeByFull[n.Full] = n
	}

	edgeMatrix := map[string]int{}
	for _, e := range edges {
		from, ok1 := nodeByFull[e.From]
		to, ok2 := nodeByFull[e.To]
		if !ok1 || !ok2 {
			continue
		}
		k := from.Group + " -> " + to.Group
		edgeMatrix[k]++
	}
	matrixKeys := make([]string, 0, len(edgeMatrix))
	for k := range edgeMatrix {
		matrixKeys = append(matrixKeys, k)
	}
	sort.Strings(matrixKeys)

	idByFull := map[string]string{}
	allNodes := append([]FuncNode(nil), nodes...)
	sortNodes(allNodes)
	for i, n := range allNodes {
		idByFull[n.Full] = fmt.Sprintf("f%d", i+1)
	}

	var b strings.Builder
	b.WriteString("# Integrated Function Graph (Unified AST View)\n\n")
	b.WriteString("Generated: 2026-02-17 JST\n\n")
	b.WriteString("## Coverage\n")
	b.WriteString(fmt.Sprintf("- total functions: %d\n", len(nodes)))
	b.WriteString(fmt.Sprintf("- go functions (including tests): %d\n", goCount))
	b.WriteString(fmt.Sprintf("- frontend functions (ts+svelte): %d\n", feCount))
	b.WriteString(fmt.Sprintf("- script functions: %d\n", shCount))
	b.WriteString(fmt.Sprintf("- total inferred edges: %d\n", len(edges)))
	b.WriteString("\n")

	b.WriteString("## Edge Matrix (Group to Group)\n")
	b.WriteString("| From -> To | Edges |\n")
	b.WriteString("|---|---:|\n")
	for _, k := range matrixKeys {
		b.WriteString(fmt.Sprintf("| `%s` | %d |\n", k, edgeMatrix[k]))
	}
	b.WriteString("\n")

	b.WriteString("## Unified Graph\n\n")
	b.WriteString("```mermaid\n")
	b.WriteString("flowchart LR\n")
	for gi, g := range groups {
		sg := fmt.Sprintf("sg%d", gi+1)
		b.WriteString(fmt.Sprintf("  subgraph %s[\"%s\"]\n", sg, mermaidEscape(g)))
		for _, n := range nodesByGroup[g] {
			id := idByFull[n.Full]
			rel := filepath.ToSlash(strings.TrimPrefix(n.File, root+string(filepath.Separator)))
			b.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", id, mermaidEscape(rel+":"+n.Name)))
		}
		b.WriteString("  end\n")
	}
	for _, e := range edges {
		from, ok1 := idByFull[e.From]
		to, ok2 := idByFull[e.To]
		if !ok1 || !ok2 {
			continue
		}
		b.WriteString(fmt.Sprintf("  %s --> %s\n", from, to))
	}
	b.WriteString("```\n")

	return os.WriteFile(outPath, []byte(b.String()), 0644)
}

func walkFiles(base string, exts ...string) []string {
	out := []string{}
	_ = filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		for _, ext := range exts {
			if strings.HasSuffix(path, ext) {
				out = append(out, path)
				break
			}
		}
		return nil
	})
	return out
}

func sortedGroups(m map[string][]FuncNode) []string {
	out := make([]string, 0, len(m))
	for g := range m {
		out = append(out, g)
	}
	sort.Strings(out)
	return out
}

func sortNodes(nodes []FuncNode) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Group == nodes[j].Group {
			if nodes[i].File == nodes[j].File {
				return nodes[i].Name < nodes[j].Name
			}
			return nodes[i].File < nodes[j].File
		}
		return nodes[i].Group < nodes[j].Group
	})
}

func sortEdges(edges []Edge) {
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From == edges[j].From {
			return edges[i].To < edges[j].To
		}
		return edges[i].From < edges[j].From
	})
}

func dedupEdges(edges []Edge) []Edge {
	seen := map[string]bool{}
	out := make([]Edge, 0, len(edges))
	for _, e := range edges {
		k := e.From + "->" + e.To
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, e)
	}
	return out
}

func mermaidEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, `"`, `\\"`)
	return s
}

func recvName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return recvName(t.X)
	case *ast.SelectorExpr:
		left := recvName(t.X)
		if left == "" {
			return t.Sel.Name
		}
		return left + "." + t.Sel.Name
	default:
		return ""
	}
}

func indexNextByte(data []byte, target byte, from int) int {
	if from < 0 {
		from = 0
	}
	for i := from; i < len(data); i++ {
		if data[i] == target {
			return i
		}
	}
	return -1
}

func matchBrace(data []byte, open int) int {
	if open < 0 || open >= len(data) || data[open] != '{' {
		return -1
	}
	depth := 0
	inS := false
	inD := false
	inB := false
	inLineComment := false
	inBlockComment := false
	escape := false

	for i := open; i < len(data); i++ {
		ch := data[i]
		next := byte(0)
		if i+1 < len(data) {
			next = data[i+1]
		}

		if inLineComment {
			if ch == '\n' {
				inLineComment = false
			}
			continue
		}
		if inBlockComment {
			if ch == '*' && next == '/' {
				inBlockComment = false
				i++
			}
			continue
		}

		if escape {
			escape = false
			continue
		}
		if inS {
			if ch == '\\' {
				escape = true
			} else if ch == '\'' {
				inS = false
			}
			continue
		}
		if inD {
			if ch == '\\' {
				escape = true
			} else if ch == '"' {
				inD = false
			}
			continue
		}
		if inB {
			if ch == '`' {
				inB = false
			}
			continue
		}

		if ch == '/' && next == '/' {
			inLineComment = true
			i++
			continue
		}
		if ch == '/' && next == '*' {
			inBlockComment = true
			i++
			continue
		}

		if ch == '\'' {
			inS = true
			continue
		}
		if ch == '"' {
			inD = true
			continue
		}
		if ch == '`' {
			inB = true
			continue
		}

		if ch == '{' {
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func countByLang(nodes []FuncNode) (goCount, feCount, shCount int) {
	for _, n := range nodes {
		switch n.Lang {
		case "go":
			goCount++
		case "fe":
			feCount++
		case "script":
			shCount++
		}
	}
	return
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
