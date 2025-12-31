package handlers

import (
	"context"
	"sort"
	"strings"
)

type TagGraphDTO struct {
	Tag     string       `json:"tag"`
	Related []TagEdgeDTO `json:"related"`
}

type TagEdgeDTO struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

func (h *Handlers) buildTagGraph(ctx context.Context) ([]TagGraphDTO, error) {
	rows, err := h.App.Queries.ListJournalTags(ctx)
	if err != nil {
		return nil, err
	}

	type edgeKey struct {
		from string
		to   string
	}
	edges := make(map[edgeKey]int)

	for _, raw := range rows {
		tags := uniqueTags(tagsFromJSONString(raw))
		if len(tags) < 2 {
			continue
		}
		for i := 0; i < len(tags); i++ {
			for j := i + 1; j < len(tags); j++ {
				a := tags[i]
				b := tags[j]
				edges[edgeKey{from: a, to: b}]++
				edges[edgeKey{from: b, to: a}]++
			}
		}
	}

	adj := make(map[string]map[string]int)
	for key, count := range edges {
		if _, ok := adj[key.from]; !ok {
			adj[key.from] = make(map[string]int)
		}
		adj[key.from][key.to] = count
	}

	nodes := make([]TagGraphDTO, 0, len(adj))
	for tag, related := range adj {
		edges := make([]TagEdgeDTO, 0, len(related))
		for other, count := range related {
			edges = append(edges, TagEdgeDTO{Tag: other, Count: count})
		}
		sort.Slice(edges, func(i, j int) bool {
			if edges[i].Count == edges[j].Count {
				return edges[i].Tag < edges[j].Tag
			}
			return edges[i].Count > edges[j].Count
		})
		nodes = append(nodes, TagGraphDTO{Tag: tag, Related: edges})
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Tag < nodes[j].Tag
	})

	return nodes, nil
}

func uniqueTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		normalized := strings.TrimSpace(tag)
		if normalized == "" {
			continue
		}
		normalized = strings.ToLower(normalized)
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	sort.Strings(out)
	return out
}
