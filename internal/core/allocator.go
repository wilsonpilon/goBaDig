package core

import (
	"errors"
	"fmt"
)

// ErrAliasesExhausted is returned by VariableAllocator.Allocate when a
// category has no more short aliases available. Running out is always
// a fatal error, never a silent alias - regra explícita da seção 4.1
// do OUTLINE.md para o MSX, generalizada aqui para qualquer categoria.
var ErrAliasesExhausted = errors.New("no more aliases available for this category")

// VariableCategory describes one class of variable a platform
// allocates short aliases for (e.g. scalar, string, array, for-next
// control) - each with its own alias length and valid character set.
// Generalizes MSX (everything 2 letters), ZX81 (everything 1 letter)
// and Spectrum (native names + 1 letter per category) under one
// mechanism (decisão já tomada em OUTLINE.md, seção 6, item 5).
type VariableCategory struct {
	Name       string
	AliasChars string
	AliasLen   int
}

// VariableAllocator assigns short aliases to long variable names
// within a category, returning ErrAliasesExhausted once a category's
// alias space is used up.
type VariableAllocator interface {
	Allocate(category VariableCategory, longName string) (alias string, err error)
}

// descendingAllocator hands out aliases in descending order within
// each category - matching MSX's own documented scheme (aa..zz,
// ordem descendente), generalized to any alphabet/length. Starting
// from the top of the alias space leaves its low end free for
// variables a user explicitly aliases (`declare nome:xy`) without
// colliding with auto-allocated ones.
type descendingAllocator struct {
	assigned map[string]map[string]string // nome da categoria -> nome longo -> alias
	next     map[string]int               // nome da categoria -> próximo índice (decrescente)
}

// NewDescendingAllocator returns a VariableAllocator that hands out
// aliases in descending order within each category.
func NewDescendingAllocator() VariableAllocator {
	return &descendingAllocator{
		assigned: make(map[string]map[string]string),
		next:     make(map[string]int),
	}
}

func (a *descendingAllocator) Allocate(category VariableCategory, longName string) (string, error) {
	assigned, ok := a.assigned[category.Name]
	if !ok {
		assigned = make(map[string]string)
		a.assigned[category.Name] = assigned
		a.next[category.Name] = aliasSpace(category) - 1
	}
	if alias, ok := assigned[longName]; ok {
		return alias, nil
	}

	index := a.next[category.Name]
	if index < 0 {
		return "", fmt.Errorf("category %q: %w", category.Name, ErrAliasesExhausted)
	}
	alias := aliasAt(category, index)
	assigned[longName] = alias
	a.next[category.Name] = index - 1
	return alias, nil
}

// aliasSpace returns how many distinct aliases a category can produce.
func aliasSpace(c VariableCategory) int {
	n := 1
	for i := 0; i < c.AliasLen; i++ {
		n *= len(c.AliasChars)
	}
	return n
}

// aliasAt renders index (0 <= index < aliasSpace(c)) as a fixed-length
// alias over c.AliasChars, treating index as a mixed-radix number.
func aliasAt(c VariableCategory, index int) string {
	base := len(c.AliasChars)
	buf := make([]byte, c.AliasLen)
	for i := c.AliasLen - 1; i >= 0; i-- {
		buf[i] = c.AliasChars[index%base]
		index /= base
	}
	return string(buf)
}
