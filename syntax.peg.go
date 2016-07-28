package yasp

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	rulestart
	ruleexpr
	ruleopenBrace
	rulecloseBrace
	ruleID
	ruleNUMBER
	ruleSTRING
	ruleWS
	ruleESCAPE
	ruleAction0
	ruleAction1
	rulePegText
	ruleAction2
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9
	ruleAction10
	ruleAction11
	ruleAction12
	ruleAction13
	ruleAction14
	ruleAction15
	ruleAction16
	ruleAction17
	ruleAction18
	ruleAction19
	ruleAction20
	ruleAction21
	ruleAction22
	ruleAction23

	rulePre
	ruleIn
	ruleSuf
)

var rul3s = [...]string{
	"Unknown",
	"start",
	"expr",
	"openBrace",
	"closeBrace",
	"ID",
	"NUMBER",
	"STRING",
	"WS",
	"ESCAPE",
	"Action0",
	"Action1",
	"PegText",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"Action10",
	"Action11",
	"Action12",
	"Action13",
	"Action14",
	"Action15",
	"Action16",
	"Action17",
	"Action18",
	"Action19",
	"Action20",
	"Action21",
	"Action22",
	"Action23",

	"Pre_",
	"_In_",
	"_Suf",
}

type tokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule pegRule, begin, end, next uint32, depth int)
	Expand(index int) tokenTree
	Tokens() <-chan token32
	AST() *node32
	Error() []token32
	trim(length int)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(string(([]rune(buffer)[node.begin:node.end]))))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (node *node32) Print(buffer string) {
	node.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegRule
	begin, end, next uint32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: uint32(t.begin), end: uint32(t.end), next: uint32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = uint32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, uint32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{pegRule: ruleIn, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: ruleSuf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth uint32, index int) {
	t.tree[index] = token32{pegRule: rule, begin: uint32(begin), end: uint32(end), next: uint32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

/*func (t *tokens16) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2 * len(tree))
		for i, v := range tree {
			expanded[i] = v.getToken32()
		}
		return &tokens32{tree: expanded}
	}
	return nil
}*/

func (t *tokens32) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	return nil
}

type YaspPEG struct {
	Parsing

	Buffer string
	buffer []rune
	rules  [35]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	Pretty bool
	tokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *YaspPEG
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *YaspPEG) PrintSyntaxTree() {
	p.tokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *YaspPEG) Highlighter() {
	p.tokenTree.PrintSyntax()
}

func (p *YaspPEG) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for token := range p.tokenTree.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:
			p.OpenBrace()
		case ruleAction1:
			p.CloseBrace()
		case ruleAction2:
			p.AddID(buffer[begin:end])
		case ruleAction3:
			p.AddNumber(buffer[begin:end])
		case ruleAction4:
			p.StartString()
		case ruleAction5:
			p.AddCharacter(buffer[begin:end])
		case ruleAction6:
			p.EndString()
		case ruleAction7:
			p.AddCharacter("\a")
		case ruleAction8:
			p.AddCharacter("\b")
		case ruleAction9:
			p.AddCharacter("\x1B")
		case ruleAction10:
			p.AddCharacter("\f")
		case ruleAction11:
			p.AddCharacter("\n")
		case ruleAction12:
			p.AddCharacter("\r")
		case ruleAction13:
			p.AddCharacter("\t")
		case ruleAction14:
			p.AddCharacter("\v")
		case ruleAction15:
			p.AddCharacter("'")
		case ruleAction16:
			p.AddCharacter("\"")
		case ruleAction17:
			p.AddCharacter("[")
		case ruleAction18:
			p.AddCharacter("]")
		case ruleAction19:
			p.AddCharacter("-")
		case ruleAction20:

			hexa, _ := strconv.ParseInt(text, 16, 32)
			p.AddCharacter(string(hexa))
		case ruleAction21:

			octal, _ := strconv.ParseInt(text, 8, 8)
			p.AddCharacter(string(octal))
		case ruleAction22:

			octal, _ := strconv.ParseInt(text, 8, 8)
			p.AddCharacter(string(octal))
		case ruleAction23:
			p.AddCharacter("\\")

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *YaspPEG) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
		p.buffer = append(p.buffer, endSymbol)
	}

	var tree tokenTree = &tokens32{tree: make([]token32, math.MaxInt16)}
	var max token32
	position, depth, tokenIndex, buffer, _rules := uint32(0), uint32(0), 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokenTree = tree
		if matches {
			p.tokenTree.trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin uint32) {
		if t := tree.Expand(tokenIndex); t != nil {
			tree = t
		}
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position, depth}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 start <- <(WS? expr (WS expr)* WS? !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				{
					position2, tokenIndex2, depth2 := position, tokenIndex, depth
					if !_rules[ruleWS]() {
						goto l2
					}
					goto l3
				l2:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
				}
			l3:
				if !_rules[ruleexpr]() {
					goto l0
				}
			l4:
				{
					position5, tokenIndex5, depth5 := position, tokenIndex, depth
					if !_rules[ruleWS]() {
						goto l5
					}
					if !_rules[ruleexpr]() {
						goto l5
					}
					goto l4
				l5:
					position, tokenIndex, depth = position5, tokenIndex5, depth5
				}
				{
					position6, tokenIndex6, depth6 := position, tokenIndex, depth
					if !_rules[ruleWS]() {
						goto l6
					}
					goto l7
				l6:
					position, tokenIndex, depth = position6, tokenIndex6, depth6
				}
			l7:
				{
					position8, tokenIndex8, depth8 := position, tokenIndex, depth
					if !matchDot() {
						goto l8
					}
					goto l0
				l8:
					position, tokenIndex, depth = position8, tokenIndex8, depth8
				}
				depth--
				add(rulestart, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 expr <- <((&('(') (openBrace WS? expr? (WS expr)* WS? closeBrace)) | (&('\'') STRING) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') NUMBER) | (&('!' | '#' | '$' | '%' | '&' | '*' | '+' | '-' | '/' | '<' | '=' | '>' | '?' | '@' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '^' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') ID))> */
		func() bool {
			position9, tokenIndex9, depth9 := position, tokenIndex, depth
			{
				position10 := position
				depth++
				{
					switch buffer[position] {
					case '(':
						{
							position12 := position
							depth++
							if buffer[position] != rune('(') {
								goto l9
							}
							position++
							{
								add(ruleAction0, position)
							}
							depth--
							add(ruleopenBrace, position12)
						}
						{
							position14, tokenIndex14, depth14 := position, tokenIndex, depth
							if !_rules[ruleWS]() {
								goto l14
							}
							goto l15
						l14:
							position, tokenIndex, depth = position14, tokenIndex14, depth14
						}
					l15:
						{
							position16, tokenIndex16, depth16 := position, tokenIndex, depth
							if !_rules[ruleexpr]() {
								goto l16
							}
							goto l17
						l16:
							position, tokenIndex, depth = position16, tokenIndex16, depth16
						}
					l17:
					l18:
						{
							position19, tokenIndex19, depth19 := position, tokenIndex, depth
							if !_rules[ruleWS]() {
								goto l19
							}
							if !_rules[ruleexpr]() {
								goto l19
							}
							goto l18
						l19:
							position, tokenIndex, depth = position19, tokenIndex19, depth19
						}
						{
							position20, tokenIndex20, depth20 := position, tokenIndex, depth
							if !_rules[ruleWS]() {
								goto l20
							}
							goto l21
						l20:
							position, tokenIndex, depth = position20, tokenIndex20, depth20
						}
					l21:
						{
							position22 := position
							depth++
							if buffer[position] != rune(')') {
								goto l9
							}
							position++
							{
								add(ruleAction1, position)
							}
							depth--
							add(rulecloseBrace, position22)
						}
						break
					case '\'':
						{
							position24 := position
							depth++
							if buffer[position] != rune('\'') {
								goto l9
							}
							position++
							{
								add(ruleAction4, position)
							}
						l26:
							{
								position27, tokenIndex27, depth27 := position, tokenIndex, depth
								{
									position28, tokenIndex28, depth28 := position, tokenIndex, depth
									{
										position30 := position
										depth++
										{
											position31, tokenIndex31, depth31 := position, tokenIndex, depth
											if buffer[position] != rune('\\') {
												goto l32
											}
											position++
											{
												position33, tokenIndex33, depth33 := position, tokenIndex, depth
												if buffer[position] != rune('a') {
													goto l34
												}
												position++
												goto l33
											l34:
												position, tokenIndex, depth = position33, tokenIndex33, depth33
												if buffer[position] != rune('A') {
													goto l32
												}
												position++
											}
										l33:
											{
												add(ruleAction7, position)
											}
											goto l31
										l32:
											position, tokenIndex, depth = position31, tokenIndex31, depth31
											if buffer[position] != rune('\\') {
												goto l36
											}
											position++
											{
												position37, tokenIndex37, depth37 := position, tokenIndex, depth
												if buffer[position] != rune('b') {
													goto l38
												}
												position++
												goto l37
											l38:
												position, tokenIndex, depth = position37, tokenIndex37, depth37
												if buffer[position] != rune('B') {
													goto l36
												}
												position++
											}
										l37:
											{
												add(ruleAction8, position)
											}
											goto l31
										l36:
											position, tokenIndex, depth = position31, tokenIndex31, depth31
											if buffer[position] != rune('\\') {
												goto l40
											}
											position++
											{
												position41, tokenIndex41, depth41 := position, tokenIndex, depth
												if buffer[position] != rune('e') {
													goto l42
												}
												position++
												goto l41
											l42:
												position, tokenIndex, depth = position41, tokenIndex41, depth41
												if buffer[position] != rune('E') {
													goto l40
												}
												position++
											}
										l41:
											{
												add(ruleAction9, position)
											}
											goto l31
										l40:
											position, tokenIndex, depth = position31, tokenIndex31, depth31
											if buffer[position] != rune('\\') {
												goto l44
											}
											position++
											{
												position45, tokenIndex45, depth45 := position, tokenIndex, depth
												if buffer[position] != rune('f') {
													goto l46
												}
												position++
												goto l45
											l46:
												position, tokenIndex, depth = position45, tokenIndex45, depth45
												if buffer[position] != rune('F') {
													goto l44
												}
												position++
											}
										l45:
											{
												add(ruleAction10, position)
											}
											goto l31
										l44:
											position, tokenIndex, depth = position31, tokenIndex31, depth31
											if buffer[position] != rune('\\') {
												goto l48
											}
											position++
											{
												position49, tokenIndex49, depth49 := position, tokenIndex, depth
												if buffer[position] != rune('n') {
													goto l50
												}
												position++
												goto l49
											l50:
												position, tokenIndex, depth = position49, tokenIndex49, depth49
												if buffer[position] != rune('N') {
													goto l48
												}
												position++
											}
										l49:
											{
												add(ruleAction11, position)
											}
											goto l31
										l48:
											position, tokenIndex, depth = position31, tokenIndex31, depth31
											if buffer[position] != rune('\\') {
												goto l52
											}
											position++
											{
												position53, tokenIndex53, depth53 := position, tokenIndex, depth
												if buffer[position] != rune('r') {
													goto l54
												}
												position++
												goto l53
											l54:
												position, tokenIndex, depth = position53, tokenIndex53, depth53
												if buffer[position] != rune('R') {
													goto l52
												}
												position++
											}
										l53:
											{
												add(ruleAction12, position)
											}
											goto l31
										l52:
											position, tokenIndex, depth = position31, tokenIndex31, depth31
											if buffer[position] != rune('\\') {
												goto l56
											}
											position++
											{
												position57, tokenIndex57, depth57 := position, tokenIndex, depth
												if buffer[position] != rune('t') {
													goto l58
												}
												position++
												goto l57
											l58:
												position, tokenIndex, depth = position57, tokenIndex57, depth57
												if buffer[position] != rune('T') {
													goto l56
												}
												position++
											}
										l57:
											{
												add(ruleAction13, position)
											}
											goto l31
										l56:
											position, tokenIndex, depth = position31, tokenIndex31, depth31
											if buffer[position] != rune('\\') {
												goto l60
											}
											position++
											{
												position61, tokenIndex61, depth61 := position, tokenIndex, depth
												if buffer[position] != rune('v') {
													goto l62
												}
												position++
												goto l61
											l62:
												position, tokenIndex, depth = position61, tokenIndex61, depth61
												if buffer[position] != rune('V') {
													goto l60
												}
												position++
											}
										l61:
											{
												add(ruleAction14, position)
											}
											goto l31
										l60:
											position, tokenIndex, depth = position31, tokenIndex31, depth31
											if buffer[position] != rune('\\') {
												goto l64
											}
											position++
											if buffer[position] != rune('\'') {
												goto l64
											}
											position++
											{
												add(ruleAction15, position)
											}
											goto l31
										l64:
											position, tokenIndex, depth = position31, tokenIndex31, depth31
											if buffer[position] != rune('\\') {
												goto l66
											}
											position++
											if buffer[position] != rune('"') {
												goto l66
											}
											position++
											{
												add(ruleAction16, position)
											}
											goto l31
										l66:
											position, tokenIndex, depth = position31, tokenIndex31, depth31
											if buffer[position] != rune('\\') {
												goto l68
											}
											position++
											if buffer[position] != rune('[') {
												goto l68
											}
											position++
											{
												add(ruleAction17, position)
											}
											goto l31
										l68:
											position, tokenIndex, depth = position31, tokenIndex31, depth31
											if buffer[position] != rune('\\') {
												goto l70
											}
											position++
											if buffer[position] != rune(']') {
												goto l70
											}
											position++
											{
												add(ruleAction18, position)
											}
											goto l31
										l70:
											position, tokenIndex, depth = position31, tokenIndex31, depth31
											if buffer[position] != rune('\\') {
												goto l72
											}
											position++
											if buffer[position] != rune('-') {
												goto l72
											}
											position++
											{
												add(ruleAction19, position)
											}
											goto l31
										l72:
											position, tokenIndex, depth = position31, tokenIndex31, depth31
											if buffer[position] != rune('\\') {
												goto l74
											}
											position++
											if buffer[position] != rune('0') {
												goto l74
											}
											position++
											{
												position75, tokenIndex75, depth75 := position, tokenIndex, depth
												if buffer[position] != rune('x') {
													goto l76
												}
												position++
												goto l75
											l76:
												position, tokenIndex, depth = position75, tokenIndex75, depth75
												if buffer[position] != rune('X') {
													goto l74
												}
												position++
											}
										l75:
											{
												position77 := position
												depth++
												{
													switch buffer[position] {
													case 'A', 'B', 'C', 'D', 'E', 'F':
														if c := buffer[position]; c < rune('A') || c > rune('F') {
															goto l74
														}
														position++
														break
													case 'a', 'b', 'c', 'd', 'e', 'f':
														if c := buffer[position]; c < rune('a') || c > rune('f') {
															goto l74
														}
														position++
														break
													default:
														if c := buffer[position]; c < rune('0') || c > rune('9') {
															goto l74
														}
														position++
														break
													}
												}

											l78:
												{
													position79, tokenIndex79, depth79 := position, tokenIndex, depth
													{
														switch buffer[position] {
														case 'A', 'B', 'C', 'D', 'E', 'F':
															if c := buffer[position]; c < rune('A') || c > rune('F') {
																goto l79
															}
															position++
															break
														case 'a', 'b', 'c', 'd', 'e', 'f':
															if c := buffer[position]; c < rune('a') || c > rune('f') {
																goto l79
															}
															position++
															break
														default:
															if c := buffer[position]; c < rune('0') || c > rune('9') {
																goto l79
															}
															position++
															break
														}
													}

													goto l78
												l79:
													position, tokenIndex, depth = position79, tokenIndex79, depth79
												}
												depth--
												add(rulePegText, position77)
											}
											{
												add(ruleAction20, position)
											}
											goto l31
										l74:
											position, tokenIndex, depth = position31, tokenIndex31, depth31
											if buffer[position] != rune('\\') {
												goto l83
											}
											position++
											{
												position84 := position
												depth++
												if c := buffer[position]; c < rune('0') || c > rune('3') {
													goto l83
												}
												position++
												if c := buffer[position]; c < rune('0') || c > rune('7') {
													goto l83
												}
												position++
												if c := buffer[position]; c < rune('0') || c > rune('7') {
													goto l83
												}
												position++
												depth--
												add(rulePegText, position84)
											}
											{
												add(ruleAction21, position)
											}
											goto l31
										l83:
											position, tokenIndex, depth = position31, tokenIndex31, depth31
											if buffer[position] != rune('\\') {
												goto l86
											}
											position++
											{
												position87 := position
												depth++
												if c := buffer[position]; c < rune('0') || c > rune('7') {
													goto l86
												}
												position++
												{
													position88, tokenIndex88, depth88 := position, tokenIndex, depth
													if c := buffer[position]; c < rune('0') || c > rune('7') {
														goto l88
													}
													position++
													goto l89
												l88:
													position, tokenIndex, depth = position88, tokenIndex88, depth88
												}
											l89:
												depth--
												add(rulePegText, position87)
											}
											{
												add(ruleAction22, position)
											}
											goto l31
										l86:
											position, tokenIndex, depth = position31, tokenIndex31, depth31
											if buffer[position] != rune('\\') {
												goto l29
											}
											position++
											if buffer[position] != rune('\\') {
												goto l29
											}
											position++
											{
												add(ruleAction23, position)
											}
										}
									l31:
										depth--
										add(ruleESCAPE, position30)
									}
									goto l28
								l29:
									position, tokenIndex, depth = position28, tokenIndex28, depth28
									{
										position92 := position
										depth++
										{
											position95, tokenIndex95, depth95 := position, tokenIndex, depth
											{
												position96, tokenIndex96, depth96 := position, tokenIndex, depth
												if buffer[position] != rune('\'') {
													goto l97
												}
												position++
												goto l96
											l97:
												position, tokenIndex, depth = position96, tokenIndex96, depth96
												if buffer[position] != rune('\\') {
													goto l95
												}
												position++
											}
										l96:
											goto l27
										l95:
											position, tokenIndex, depth = position95, tokenIndex95, depth95
										}
										if !matchDot() {
											goto l27
										}
									l93:
										{
											position94, tokenIndex94, depth94 := position, tokenIndex, depth
											{
												position98, tokenIndex98, depth98 := position, tokenIndex, depth
												{
													position99, tokenIndex99, depth99 := position, tokenIndex, depth
													if buffer[position] != rune('\'') {
														goto l100
													}
													position++
													goto l99
												l100:
													position, tokenIndex, depth = position99, tokenIndex99, depth99
													if buffer[position] != rune('\\') {
														goto l98
													}
													position++
												}
											l99:
												goto l94
											l98:
												position, tokenIndex, depth = position98, tokenIndex98, depth98
											}
											if !matchDot() {
												goto l94
											}
											goto l93
										l94:
											position, tokenIndex, depth = position94, tokenIndex94, depth94
										}
										depth--
										add(rulePegText, position92)
									}
									{
										add(ruleAction5, position)
									}
								}
							l28:
								goto l26
							l27:
								position, tokenIndex, depth = position27, tokenIndex27, depth27
							}
							if buffer[position] != rune('\'') {
								goto l9
							}
							position++
							{
								add(ruleAction6, position)
							}
							depth--
							add(ruleSTRING, position24)
						}
						break
					case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
						{
							position103 := position
							depth++
							{
								position104 := position
								depth++
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l9
								}
								position++
							l105:
								{
									position106, tokenIndex106, depth106 := position, tokenIndex, depth
									if c := buffer[position]; c < rune('0') || c > rune('9') {
										goto l106
									}
									position++
									goto l105
								l106:
									position, tokenIndex, depth = position106, tokenIndex106, depth106
								}
								depth--
								add(rulePegText, position104)
							}
							{
								add(ruleAction3, position)
							}
							depth--
							add(ruleNUMBER, position103)
						}
						break
					default:
						{
							position108 := position
							depth++
							{
								position109 := position
								depth++
								{
									switch buffer[position] {
									case '?':
										if buffer[position] != rune('?') {
											goto l9
										}
										position++
										break
									case '=':
										if buffer[position] != rune('=') {
											goto l9
										}
										position++
										break
									case '>':
										if buffer[position] != rune('>') {
											goto l9
										}
										position++
										break
									case '<':
										if buffer[position] != rune('<') {
											goto l9
										}
										position++
										break
									case '&':
										if buffer[position] != rune('&') {
											goto l9
										}
										position++
										break
									case '^':
										if buffer[position] != rune('^') {
											goto l9
										}
										position++
										break
									case '%':
										if buffer[position] != rune('%') {
											goto l9
										}
										position++
										break
									case '$':
										if buffer[position] != rune('$') {
											goto l9
										}
										position++
										break
									case '#':
										if buffer[position] != rune('#') {
											goto l9
										}
										position++
										break
									case '@':
										if buffer[position] != rune('@') {
											goto l9
										}
										position++
										break
									case '!':
										if buffer[position] != rune('!') {
											goto l9
										}
										position++
										break
									case '/':
										if buffer[position] != rune('/') {
											goto l9
										}
										position++
										break
									case '*':
										if buffer[position] != rune('*') {
											goto l9
										}
										position++
										break
									case '+':
										if buffer[position] != rune('+') {
											goto l9
										}
										position++
										break
									case '-':
										if buffer[position] != rune('-') {
											goto l9
										}
										position++
										break
									case '_':
										if buffer[position] != rune('_') {
											goto l9
										}
										position++
										break
									case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
										if c := buffer[position]; c < rune('A') || c > rune('Z') {
											goto l9
										}
										position++
										break
									default:
										if c := buffer[position]; c < rune('a') || c > rune('z') {
											goto l9
										}
										position++
										break
									}
								}

							l111:
								{
									position112, tokenIndex112, depth112 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
											{
												position114, tokenIndex114, depth114 := position, tokenIndex, depth
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l115
												}
												position++
												goto l114
											l115:
												position, tokenIndex, depth = position114, tokenIndex114, depth114
												if c := buffer[position]; c < rune('0') || c > rune('9') {
													goto l112
												}
												position++
											}
										l114:
											break
										case '?':
											if buffer[position] != rune('?') {
												goto l112
											}
											position++
											break
										case '=':
											if buffer[position] != rune('=') {
												goto l112
											}
											position++
											break
										case '>':
											if buffer[position] != rune('>') {
												goto l112
											}
											position++
											break
										case '<':
											if buffer[position] != rune('<') {
												goto l112
											}
											position++
											break
										case '"':
											if buffer[position] != rune('"') {
												goto l112
											}
											position++
											break
										case '\'':
											if buffer[position] != rune('\'') {
												goto l112
											}
											position++
											break
										case '&':
											if buffer[position] != rune('&') {
												goto l112
											}
											position++
											break
										case '^':
											if buffer[position] != rune('^') {
												goto l112
											}
											position++
											break
										case '%':
											if buffer[position] != rune('%') {
												goto l112
											}
											position++
											break
										case '$':
											if buffer[position] != rune('$') {
												goto l112
											}
											position++
											break
										case '#':
											if buffer[position] != rune('#') {
												goto l112
											}
											position++
											break
										case '@':
											if buffer[position] != rune('@') {
												goto l112
											}
											position++
											break
										case '!':
											if buffer[position] != rune('!') {
												goto l112
											}
											position++
											break
										case '/':
											if buffer[position] != rune('/') {
												goto l112
											}
											position++
											break
										case '*':
											if buffer[position] != rune('*') {
												goto l112
											}
											position++
											break
										case '+':
											if buffer[position] != rune('+') {
												goto l112
											}
											position++
											break
										case '-':
											if buffer[position] != rune('-') {
												goto l112
											}
											position++
											break
										case '_':
											if buffer[position] != rune('_') {
												goto l112
											}
											position++
											break
										case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
											if c := buffer[position]; c < rune('A') || c > rune('Z') {
												goto l112
											}
											position++
											break
										default:
											if c := buffer[position]; c < rune('a') || c > rune('z') {
												goto l112
											}
											position++
											break
										}
									}

									goto l111
								l112:
									position, tokenIndex, depth = position112, tokenIndex112, depth112
								}
								depth--
								add(rulePegText, position109)
							}
							{
								add(ruleAction2, position)
							}
							depth--
							add(ruleID, position108)
						}
						break
					}
				}

				depth--
				add(ruleexpr, position10)
			}
			return true
		l9:
			position, tokenIndex, depth = position9, tokenIndex9, depth9
			return false
		},
		/* 2 openBrace <- <('(' Action0)> */
		nil,
		/* 3 closeBrace <- <(')' Action1)> */
		nil,
		/* 4 ID <- <(<(((&('?') '?') | (&('=') '=') | (&('>') '>') | (&('<') '<') | (&('&') '&') | (&('^') '^') | (&('%') '%') | (&('$') '$') | (&('#') '#') | (&('@') '@') | (&('!') '!') | (&('/') '/') | (&('*') '*') | (&('+') '+') | (&('-') '-') | (&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z])) ((&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') ([0-9] / [0-9])) | (&('?') '?') | (&('=') '=') | (&('>') '>') | (&('<') '<') | (&('"') '"') | (&('\'') '\'') | (&('&') '&') | (&('^') '^') | (&('%') '%') | (&('$') '$') | (&('#') '#') | (&('@') '@') | (&('!') '!') | (&('/') '/') | (&('*') '*') | (&('+') '+') | (&('-') '-') | (&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))*)> Action2)> */
		nil,
		/* 5 NUMBER <- <(<[0-9]+> Action3)> */
		nil,
		/* 6 STRING <- <('\'' Action4 (ESCAPE / (<(!('\'' / '\\') .)+> Action5))* '\'' Action6)> */
		nil,
		/* 7 WS <- <((&('\n') '\n') | (&('\r') '\r') | (&('\t') '\t') | (&(' ') ' '))+> */
		func() bool {
			position122, tokenIndex122, depth122 := position, tokenIndex, depth
			{
				position123 := position
				depth++
				{
					switch buffer[position] {
					case '\n':
						if buffer[position] != rune('\n') {
							goto l122
						}
						position++
						break
					case '\r':
						if buffer[position] != rune('\r') {
							goto l122
						}
						position++
						break
					case '\t':
						if buffer[position] != rune('\t') {
							goto l122
						}
						position++
						break
					default:
						if buffer[position] != rune(' ') {
							goto l122
						}
						position++
						break
					}
				}

			l124:
				{
					position125, tokenIndex125, depth125 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '\n':
							if buffer[position] != rune('\n') {
								goto l125
							}
							position++
							break
						case '\r':
							if buffer[position] != rune('\r') {
								goto l125
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l125
							}
							position++
							break
						default:
							if buffer[position] != rune(' ') {
								goto l125
							}
							position++
							break
						}
					}

					goto l124
				l125:
					position, tokenIndex, depth = position125, tokenIndex125, depth125
				}
				depth--
				add(ruleWS, position123)
			}
			return true
		l122:
			position, tokenIndex, depth = position122, tokenIndex122, depth122
			return false
		},
		/* 8 ESCAPE <- <(('\\' ('a' / 'A') Action7) / ('\\' ('b' / 'B') Action8) / ('\\' ('e' / 'E') Action9) / ('\\' ('f' / 'F') Action10) / ('\\' ('n' / 'N') Action11) / ('\\' ('r' / 'R') Action12) / ('\\' ('t' / 'T') Action13) / ('\\' ('v' / 'V') Action14) / ('\\' '\'' Action15) / ('\\' '"' Action16) / ('\\' '[' Action17) / ('\\' ']' Action18) / ('\\' '-' Action19) / ('\\' ('0' ('x' / 'X')) <((&('A' | 'B' | 'C' | 'D' | 'E' | 'F') [A-F]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f') [a-f]) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]))+> Action20) / ('\\' <([0-3] [0-7] [0-7])> Action21) / ('\\' <([0-7] [0-7]?)> Action22) / ('\\' '\\' Action23))> */
		nil,
		/* 10 Action0 <- <{ p.OpenBrace() }> */
		nil,
		/* 11 Action1 <- <{ p.CloseBrace() }> */
		nil,
		nil,
		/* 13 Action2 <- <{ p.AddID(buffer[begin:end]) }> */
		nil,
		/* 14 Action3 <- <{ p.AddNumber(buffer[begin:end]) }> */
		nil,
		/* 15 Action4 <- <{ p.StartString() }> */
		nil,
		/* 16 Action5 <- <{ p.AddCharacter(buffer[begin:end]) }> */
		nil,
		/* 17 Action6 <- <{ p.EndString() }> */
		nil,
		/* 18 Action7 <- <{ p.AddCharacter("\a") }> */
		nil,
		/* 19 Action8 <- <{ p.AddCharacter("\b") }> */
		nil,
		/* 20 Action9 <- <{ p.AddCharacter("\x1B") }> */
		nil,
		/* 21 Action10 <- <{ p.AddCharacter("\f") }> */
		nil,
		/* 22 Action11 <- <{ p.AddCharacter("\n") }> */
		nil,
		/* 23 Action12 <- <{ p.AddCharacter("\r") }> */
		nil,
		/* 24 Action13 <- <{ p.AddCharacter("\t") }> */
		nil,
		/* 25 Action14 <- <{ p.AddCharacter("\v") }> */
		nil,
		/* 26 Action15 <- <{ p.AddCharacter("'") }> */
		nil,
		/* 27 Action16 <- <{ p.AddCharacter("\"") }> */
		nil,
		/* 28 Action17 <- <{ p.AddCharacter("[") }> */
		nil,
		/* 29 Action18 <- <{ p.AddCharacter("]") }> */
		nil,
		/* 30 Action19 <- <{ p.AddCharacter("-") }> */
		nil,
		/* 31 Action20 <- <{
		   hexa, _ := strconv.ParseInt(text, 16, 32)
		   p.AddCharacter(string(hexa)) }> */
		nil,
		/* 32 Action21 <- <{
		   octal, _ := strconv.ParseInt(text, 8, 8)
		   p.AddCharacter(string(octal)) }> */
		nil,
		/* 33 Action22 <- <{
		   octal, _ := strconv.ParseInt(text, 8, 8)
		   p.AddCharacter(string(octal)) }> */
		nil,
		/* 34 Action23 <- <{ p.AddCharacter("\\") }> */
		nil,
	}
	p.rules = _rules
}
