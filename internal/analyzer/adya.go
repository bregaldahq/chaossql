package analyzer

import (
	"fmt"
	"regexp"
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

func (g *AdyaGraph) AddEdge(from, to string, depType DependencyType, item string, isAbortedWriter bool) {
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
	parts := strings.Split(item, ":")
	return parts[0]
}

func BuildGraph(trace domain.ExecutionTrace) *AdyaGraph {
	g := NewAdyaGraph()
	lastWriter := make(map[string]string)
	readers := make(map[string][]string)
	abortedTx := make(map[string]bool)

	for _, event := range trace {
		txID := fmt.Sprintf("T%d-%d", event.WorkerID, event.OpIndex)
		if event.Type == domain.EventRollback {
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
				g.AddEdge(lw, txID, DepWW, item, abortedTx[lw])
			}
			// Read-Write conflict on exact item and table scan
			for _, r := range readers[item] {
				if r != txID {
					g.AddEdge(r, txID, DepRW, item, abortedTx[r])
				}
			}
			if item != table {
				for _, r := range readers[table] {
					if r != txID {
						g.AddEdge(r, txID, DepRW, item, abortedTx[r])
					}
				}
			}
			lastWriter[item] = txID
			lastWriter[table] = txID
			readers[item] = nil
		} else {
			// Write-Read conflict on exact item or table
			if lw, ok := lastWriter[item]; ok && lw != txID {
				g.AddEdge(lw, txID, DepWR, item, abortedTx[lw])
			}
			if item == table {
				// Table scan reads any previous writes to table
				for k, lw := range lastWriter {
					if getTable(k) == table && lw != txID {
						g.AddEdge(lw, txID, DepWR, k, abortedTx[lw])
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
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var path []Edge

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
		return domain.AnomalyWriteSkew
	}

	return domain.AnomalyUnknown
}
