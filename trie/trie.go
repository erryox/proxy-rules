package trie

type Trie struct {
	children map[rune]*Trie
	end      bool
}

func NewTrie() *Trie {
	return &Trie{children: make(map[rune]*Trie)}
}

func (t *Trie) Insert(s string) {
	node := t
	for _, char := range s {
		if _, exists := node.children[char]; !exists {
			node.children[char] = &Trie{children: make(map[rune]*Trie)}
		}
		node = node.children[char]
	}
	node.end = true
}

func (t *Trie) GetAllFirstLevelPrefixes() []string {
	var prefixes []string
	var dfs func(node *Trie, prefix string)
	dfs = func(node *Trie, prefix string) {
		if node.end {
			prefixes = append(prefixes, prefix)
			return
		}
		for char, child := range node.children {
			dfs(child, prefix+string(char))
		}
	}
	dfs(t, "")
	return prefixes
}
