package analyzer

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bregaldahq/chaossql/internal/domain"
)

type DependencyType string

const (
	DepWR DependencyType = "WR"
	DepWW DependencyType = "WW"
	DepRW DependencyType = "RW"
)

type Edge struct {
	From            string
	To              string
	Type            DependencyType
	Item            string
	IsAbortedWriter bool
}

type Cycle []Edge

type AdyaGraph struct {
	Nodes map[string]bool
	Edges map[string][]Edge
}

func NewAdyaGraph() *AdyaGraph {
	return &AdyaGraph{
		Nodes: make(map[string]bool),
		Edges: make(map[string][]Edge),
	}
}

func (g *AdyaGraph) AddNode(node string) {
	g.Nodes[node] = true
}

func (g *AdyaGraph) AddEdge(from, to string, depType DependencyType, item string) {
	g.AddEdgeWithAbort(from, to, depType, item, false)
}

func (g *AdyaGraph) AddEdgeWithAbort(from, to string, depType DependencyType, item string, isAbortedWriter bool) {
	g.AddNode(from)
	g.AddNode(to)
	for _, e := range g.Edges[from] {
		if e.To == to && e.Type == depType && e.Item == item {
			return
		}
	}
	g.Edges[from] = append(g.Edges[from], Edge{
		From:            from,
		To:              to,
		Type:            depType,
		Item:            item,
		IsAbortedWriter: isAbortedWriter,
	})
}

var (
	reSelect = regexp.MustCompile(`(?i)SELECT\s+.*?\s+FROM\s+([a-zA-Z0-9_]+)(?:\s+WHERE\s+id\s*=\s*(\d+))?`)
	reWrite  = regexp.MustCompile(`(?i)(?:UPDATE|INSERT\s+INTO|DELETE\s+FROM)\s+([a-zA-Z0-9_]+)(?:\s+.*?\s+WHERE\s+id\s*=\s*(\d+))?`)
)

func extractItem(sql string) (isWrite bool, item string) {
	upper := strings.ToUpper(sql)
	var table string
	var id string

	// Fast-path for UPDATE <table> SET ... WHERE id = <id>
	if strings.HasPrefix(upper, "UPDATE ") {
		isWrite = true
		rest := strings.TrimSpace(sql[7:])
		fields := strings.Fields(rest)
		if len(fields) > 0 {
			table = fields[0]
		}
		if idIdx := strings.Index(upper, "WHERE ID = "); idIdx != -1 {
			idPart := strings.TrimSpace(sql[idIdx+11:])
			idFields := strings.Fields(idPart)
			if len(idFields) > 0 {
				id = strings.TrimRight(idFields[0], ";,")
			}
		}
	} else if strings.HasPrefix(upper, "INSERT INTO ") {
		isWrite = true
		rest := strings.TrimSpace(sql[12:])
		fields := strings.Fields(rest)
		if len(fields) > 0 {
			table = strings.TrimRight(fields[0], " (")
		}
	} else if strings.HasPrefix(upper, "DELETE FROM ") {
		isWrite = true
		rest := strings.TrimSpace(sql[12:])
		fields := strings.Fields(rest)
		if len(fields) > 0 {
			table = fields[0]
		}
		if idIdx := strings.Index(upper, "WHERE ID = "); idIdx != -1 {
			idPart := strings.TrimSpace(sql[idIdx+11:])
			idFields := strings.Fields(idPart)
			if len(idFields) > 0 {
				id = strings.TrimRight(idFields[0], ";,")
			}
		}
	} else if strings.HasPrefix(upper, "SELECT ") {
		isWrite = false
		if fromIdx := strings.Index(upper, " FROM "); fromIdx != -1 {
			rest := strings.TrimSpace(sql[fromIdx+6:])
			fields := strings.Fields(rest)
			if len(fields) > 0 {
				table = strings.TrimRight(fields[0], ";,")
			}
			if idIdx := strings.Index(upper, "WHERE ID = "); idIdx != -1 {
				idPart := strings.TrimSpace(sql[idIdx+11:])
				idFields := strings.Fields(idPart)
				if len(idFields) > 0 {
					id = strings.TrimRight(idFields[0], ";,")
				}
			}
		}
	}

	if table != "" {
		if id != "" {
			if _, err := strconv.Atoi(id); err == nil {
				return isWrite, table + ":" + id
			}
		}
		return isWrite, table
	}

	// Regex fallback for non-standard queries
	if m := reWrite.FindStringSubmatch(sql); m != nil {
		item := m[1]
		if len(m) > 2 && m[2] != "" {
			item += ":" + m[2]
		}
		return true, item
	}
	if m := reSelect.FindStringSubmatch(sql); m != nil {
		item := m[1]
		if len(m) > 2 && m[2] != "" {
			item += ":" + m[2]
		}
		return false, item
	}
	return false, ""
}

func getTable(item string) string {
	if idx := strings.IndexByte(item, ':'); idx != -1 {
		return item[:idx]
	}
	return item
}

func BuildGraph(trace domain.ExecutionTrace) *AdyaGraph {
	g := NewAdyaGraph()
	lastWriter := make(map[string]string)
	readers := make(map[string][]string)
	abortedTx := make(map[string]bool)

	for _, event := range trace {
		if event.Type == domain.EventRollback {
			txID := fmt.Sprintf("T%d-%d", event.WorkerID, event.OpIndex)
			abortedTx[txID] = true
		}
	}

	for _, event := range trace {
		if event.Type != domain.EventExec {
			continue
		}

		txID := fmt.Sprintf("T%d-%d", event.WorkerID, event.OpIndex)
		isWrite, item := extractItem(event.SQL)
		if item == "" {
			continue
		}

		table := getTable(item)
		g.AddNode(txID)

		if isWrite {
			// Write-Write conflict on exact item
			if lw, ok := lastWriter[item]; ok && lw != txID {
				g.AddEdgeWithAbort(lw, txID, DepWW, item, abortedTx[lw])
			}
			// Read-Write conflict on exact item and table scan
			for _, r := range readers[item] {
				if r != txID {
					g.AddEdgeWithAbort(r, txID, DepRW, item, abortedTx[r])
				}
			}
			if item != table {
				for _, r := range readers[table] {
					if r != txID {
						g.AddEdgeWithAbort(r, txID, DepRW, item, abortedTx[r])
					}
				}
			}
			lastWriter[item] = txID
			lastWriter[table] = txID
			readers[item] = nil
		} else {
			// Write-Read conflict on exact item or table
			if lw, ok := lastWriter[item]; ok && lw != txID {
				g.AddEdgeWithAbort(lw, txID, DepWR, item, abortedTx[lw])
			}
			if item == table {
				// Table scan reads any previous writes to table
				for k, lw := range lastWriter {
					if getTable(k) == table && lw != txID {
						g.AddEdgeWithAbort(lw, txID, DepWR, k, abortedTx[lw])
					}
				}
			}
			found := false
			for _, r := range readers[item] {
				if r == txID {
					found = true
					break
				}
			}
			if !found {
				readers[item] = append(readers[item], txID)
			}
		}
	}
	return g
}

func FindCycles(g *AdyaGraph) []Cycle {
	var cycles []Cycle
	visited := make(map[string]bool, len(g.Nodes))
	recStack := make(map[string]bool, len(g.Nodes))
	path := make([]Edge, 0, 16)

	var dfs func(u string)
	dfs = func(u string) {
		visited[u] = true
		recStack[u] = true

		for _, e := range g.Edges[u] {
			v := e.To
			if !visited[v] {
				path = append(path, e)
				dfs(v)
				path = path[:len(path)-1]
			} else if recStack[v] {
				var c Cycle
				cycleStart := -1
				for i, pe := range path {
					if pe.From == v {
						cycleStart = i
						break
					}
				}
				if cycleStart != -1 {
					for i := cycleStart; i < len(path); i++ {
						c = append(c, path[i])
					}
					c = append(c, e)
					cycles = append(cycles, c)
				} else {
					if len(path) == 0 {
						c = append(c, e)
						cycles = append(cycles, c)
					} else {
						if len(path) > 0 && path[0].From == v {
							for i := 0; i < len(path); i++ {
								c = append(c, path[i])
							}
							c = append(c, e)
							cycles = append(cycles, c)
						}
					}
				}
			}
		}
		recStack[u] = false
	}

	for node := range g.Nodes {
		if !visited[node] {
			dfs(node)
		}
	}
	return cycles
}

func ClassifyCycle(c Cycle) domain.AnomalyType {
	if len(c) == 0 {
		return domain.AnomalyUnknown
	}

	hasWW := false
	hasRW := false
	hasWR := false
	hasAbortedWR := false

	for _, e := range c {
		switch e.Type {
		case DepWW:
			hasWW = true
		case DepRW:
			hasRW = true
		case DepWR:
			hasWR = true
			if e.IsAbortedWriter {
				hasAbortedWR = true
			}
		}
	}

	if hasAbortedWR {
		return domain.AnomalyG1aDirtyRead
	}
	if hasWW && !hasRW && !hasWR {
		return domain.AnomalyG0DirtyWrite
	}
	if hasWR && !hasWW && !hasRW {
		return domain.AnomalyG1cCircularInfo
	}
	if hasRW && hasWR {
		return domain.AnomalyA5AReadSkew
	}
	if hasWW && hasRW && !hasWR {
		return domain.AnomalyLostUpdate
	}
	if hasRW && !hasWW && !hasWR {
		if len(c) > 2 {
			return domain.AnomalyG2AntiDependency
		}
		return domain.AnomalyWriteSkew
	}

	return domain.AnomalyUnknown
}
