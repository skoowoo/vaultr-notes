package search

import (
	"strings"
	"sync"

	"github.com/blevesearch/bleve/v2/analysis"
	"github.com/blevesearch/bleve/v2/registry"
	"github.com/yanyiwu/gojieba"
)

const jiebaTokenizerName = "jieba"

// jiebaHandle is a package-level singleton. Loading the jieba dictionaries
// is expensive (~1s, ~50MB), so we do it once and protect concurrent
// Tokenize calls with jiebaMu (gojieba is not goroutine-safe).
var (
	jiebaOnce   sync.Once
	jiebaHandle *gojieba.Jieba
	jiebaMu     sync.Mutex
)

func initJieba() {
	jiebaOnce.Do(func() {
		jiebaHandle = gojieba.NewJieba()
	})
}

// JiebaTokenizer is a bleve analysis.Tokenizer backed by gojieba.
// It uses SearchMode, which emits both full compound words and their
// sub-components (e.g. "中华人民" also produces "中华" and "人民"),
// giving better search recall than DefaultMode.
type JiebaTokenizer struct{}

func newJiebaTokenizer(_ map[string]interface{}, _ *registry.Cache) (analysis.Tokenizer, error) {
	initJieba()
	return &JiebaTokenizer{}, nil
}

// Tokenize implements analysis.Tokenizer.
// gojieba.Word.Start and .End are already byte offsets into the input string,
// so no rune→byte conversion is needed.
func (t *JiebaTokenizer) Tokenize(input []byte) analysis.TokenStream {
	text := string(input)

	jiebaMu.Lock()
	words := jiebaHandle.Tokenize(text, gojieba.SearchMode, true)
	jiebaMu.Unlock()

	tokens := make(analysis.TokenStream, 0, len(words))
	for i, w := range words {
		if strings.TrimSpace(w.Str) == "" {
			continue
		}
		tokens = append(tokens, &analysis.Token{
			Term:     []byte(w.Str),
			Start:    w.Start,
			End:      w.End,
			Position: i + 1,
			Type:     analysis.AlphaNumeric,
		})
	}
	return tokens
}

func init() {
	if err := registry.RegisterTokenizer(jiebaTokenizerName, newJiebaTokenizer); err != nil {
		panic("search: register jieba tokenizer: " + err.Error())
	}
}
