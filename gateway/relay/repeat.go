package relay

import (
	"bytes"
	"encoding/json"
	"unicode"
	"unicode/utf8"
)

// repeatedKey reports the first key given twice in one object of raw, at
// any depth. Two keys are the same when they are equal under the case
// folding of encoding/json, which decodes both into one struct field.
// Two JSON readers may keep different values of a repeated key,
// so the gateway relays only arguments in which every reader sees the same
// value. Raw that is not JSON has no repeat: the brain refuses it.
func repeatedKey(raw []byte) (string, bool) {
	type frame struct {
		keys map[string]bool // nil in an array
		key  bool            // the next token of an object is a key
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	var stack []*frame
	for {
		tok, err := dec.Token()
		if err != nil {
			return "", false
		}
		var top *frame
		if len(stack) > 0 {
			top = stack[len(stack)-1]
		}
		switch tok {
		case json.Delim('{'):
			stack = append(stack, &frame{keys: map[string]bool{}, key: true})
			continue
		case json.Delim('['):
			stack = append(stack, &frame{})
			continue
		case json.Delim('}'), json.Delim(']'):
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				return "", false
			}
			top = stack[len(stack)-1]
		default:
			if top != nil && top.keys != nil && top.key {
				k := tok.(string)
				folded := foldKey(k)
				if top.keys[folded] {
					return k, true
				}
				top.keys[folded] = true
				top.key = false
				continue
			}
		}
		// A value ended: in an object, a key comes next.
		if top == nil {
			return "", false
		}
		top.key = true
	}
}

// foldKey folds a key as encoding/json does to match a struct field:
// foldKey(x) == foldKey(y) exactly when bytes.EqualFold(x, y). ASCII
// letters go to upper case; any other rune goes to the smallest rune of its
// Unicode fold set.
func foldKey(k string) string {
	out := make([]byte, 0, len(k))
	for _, r := range k {
		if r < utf8.RuneSelf {
			if 'a' <= r && r <= 'z' {
				r -= 'a' - 'A'
			}
			out = append(out, byte(r))
			continue
		}
		for {
			next := unicode.SimpleFold(r)
			if next <= r {
				r = next
				break
			}
			r = next
		}
		out = utf8.AppendRune(out, r)
	}
	return string(out)
}
