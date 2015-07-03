// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General
// Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

package markdown

import (
	"io"
	"strconv"
	"strings"

	"github.com/opennota/html"
)

type Renderer struct {
	w *monadicWriter
}

func NewRenderer(w io.Writer) *Renderer {
	return &Renderer{newMonadicWriter(w)}
}

func (r *Renderer) Render(tokens []Token, options RenderOptions) error {
	for i, tok := range tokens {
		if tok, ok := tok.(*Inline); ok {
			r.renderInline(tok.Children, options)
		} else {
			r.renderToken(tokens, i, options)
		}
	}

	r.w.Flush()

	return r.w.err
}

func (r *Renderer) renderInline(tokens []Token, o RenderOptions) {
	for i := range tokens {
		r.renderToken(tokens, i, o)
	}
}

func (r *Renderer) renderInlineAsText(tokens []Token) {
	for _, tok := range tokens {
		if text, ok := tok.(*Text); ok {
			html.WriteEscapedString(r.w, text.Content)
		} else if img, ok := tok.(*Image); ok {
			r.renderInlineAsText(img.Tokens)
		}
	}
}

func (r *Renderer) renderToken(tokens []Token, idx int, options RenderOptions) {
	tok := tokens[idx]

	switch tok := tok.(type) {
	case *BlockquoteClose:
		r.w.WriteString("</blockquote>")

	case *BlockquoteOpen:
		r.w.WriteString("<blockquote>")

	case *BulletListClose:
		r.w.WriteString("</ul>")

	case *BulletListOpen:
		r.w.WriteString("<ul>")

	case *CodeBlock:
		r.w.WriteString("<pre><code>")
		html.WriteEscapedString(r.w, tok.Content)
		r.w.WriteString("</code></pre>")

	case *CodeInline:
		r.w.WriteString("<code>")
		html.WriteEscapedString(r.w, tok.Content)
		r.w.WriteString("</code>")

	case *EmphasisClose:
		r.w.WriteString("</em>")

	case *EmphasisOpen:
		r.w.WriteString("<em>")

	case *Fence:
		r.w.WriteString("<pre><code")
		if tok.Params != "" {
			langName := strings.SplitN(unescapeAll(tok.Params), " ", 2)[0]
			r.w.WriteString(` class="`)
			r.w.WriteString(options.LangPrefix)
			html.WriteEscapedString(r.w, langName)
			r.w.WriteByte('"')
		}
		r.w.WriteByte('>')
		html.WriteEscapedString(r.w, tok.Content)
		r.w.WriteString("</code></pre>")

	case *Hardbreak:
		if options.XHTML {
			r.w.WriteString("<br />\n")
		} else {
			r.w.WriteString("<br>\n")
		}

	case *HeadingClose:
		r.w.WriteString("</h")
		r.w.WriteByte("0123456789"[tok.HLevel])
		r.w.WriteString(">")

	case *HeadingOpen:
		r.w.WriteString("<h")
		r.w.WriteByte("0123456789"[tok.HLevel])
		r.w.WriteByte('>')

	case *Hr:
		if options.XHTML {
			r.w.WriteString("<hr />")
		} else {
			r.w.WriteString("<hr>")
		}

	case *HTMLBlock:
		r.w.WriteString(tok.Content)
		return // no newline

	case *HTMLInline:
		r.w.WriteString(tok.Content)

	case *Image:
		r.w.WriteString(`<img src="`)
		html.WriteEscapedString(r.w, tok.Src)
		r.w.WriteString(`" alt="`)
		r.renderInlineAsText(tok.Tokens)
		r.w.WriteByte('"')

		if tok.Title != "" {
			r.w.WriteString(` title="`)
			html.WriteEscapedString(r.w, html.ReplaceEntities(tok.Title))
			r.w.WriteByte('"')
		}
		if options.XHTML {
			r.w.WriteString(" />")
		} else {
			r.w.WriteByte('>')
		}

	case *LinkClose:
		r.w.WriteString("</a>")

	case *LinkOpen:
		r.w.WriteString(`<a href="`)
		html.WriteEscapedString(r.w, tok.Href)
		r.w.WriteByte('"')
		if tok.Title != "" {
			r.w.WriteString(` title="`)
			html.WriteEscapedString(r.w, html.ReplaceEntities(tok.Title))
			r.w.WriteByte('"')
		}
		if tok.Target != "" {
			r.w.WriteString(` target="`)
			html.WriteEscapedString(r.w, tok.Target)
			r.w.WriteByte('"')
		}
		if options.Nofollow {
			r.w.WriteString(` rel="nofollow"`)
		}
		r.w.WriteByte('>')

	case *ListItemClose:
		r.w.WriteString("</li>")

	case *ListItemOpen:
		r.w.WriteString("<li>")

	case *OrderedListClose:
		r.w.WriteString("</ol>")

	case *OrderedListOpen:
		if tok.Order > 1 {
			r.w.WriteString(`<ol start="`)
			r.w.WriteString(strconv.Itoa(tok.Order))
			r.w.WriteString(`">`)
		} else {
			r.w.WriteString("<ol>")
		}

	case *ParagraphClose:
		if !tok.Tight {
			r.w.WriteString("</p>")
		} else if tokens[idx+1].Closing() {
			return // no newline
		}

	case *ParagraphOpen:
		if !tok.Tight {
			r.w.WriteString("<p>")
		}

	case *Softbreak:
		if options.Breaks {
			if options.XHTML {
				r.w.WriteString("<br />\n")
			} else {
				r.w.WriteString("<br>\n")
			}
		} else {
			r.w.WriteByte('\n')
		}
		return

	case *StrongClose:
		r.w.WriteString("</strong>")

	case *StrongOpen:
		r.w.WriteString("<strong>")

	case *StrikethroughClose:
		r.w.WriteString("</s>")

	case *StrikethroughOpen:
		r.w.WriteString("<s>")

	case *TableClose:
		r.w.WriteString("</table>")

	case *TableOpen:
		r.w.WriteString("<table>")

	case *TbodyClose:
		r.w.WriteString("</tbody>")

	case *TbodyOpen:
		r.w.WriteString("<tbody>")

	case *TdClose:
		r.w.WriteString("</td>")

	case *TdOpen:
		if tok.Align != AlignNone {
			r.w.WriteString(`<td style="text-align:`)
			r.w.WriteString(tok.Align.String())
			r.w.WriteString(`">`)
		} else {
			r.w.WriteString("<td>")
		}

	case *Text:
		html.WriteEscapedString(r.w, tok.Content)

	case *TheadClose:
		r.w.WriteString("</thead>")

	case *TheadOpen:
		r.w.WriteString("<thead>")

	case *ThClose:
		r.w.WriteString("</th>")

	case *ThOpen:
		if align := tok.Align; align != AlignNone {
			r.w.WriteString(`<th style="text-align:`)
			r.w.WriteString(align.String())
			r.w.WriteString(`">`)
		} else {
			r.w.WriteString("<th>")
		}

	case *TrClose:
		r.w.WriteString("</tr>")

	case *TrOpen:
		r.w.WriteString("<tr>")

	default:
		panic("unknown token type")
	}

	needLf := false
	if tok.Block() {
		needLf = true

		if tok.Opening() {
			nextTok := tokens[idx+1]
			switch nextTok := nextTok.(type) {
			case *Inline:
				needLf = false
			case *ParagraphOpen:
				if nextTok.Tight {
					needLf = false
				}
			case *ParagraphClose:
				if nextTok.Tight {
					needLf = false
				}
			}
			if needLf && nextTok.Closing() && nextTok.Tag() == tok.Tag() {
				needLf = false
			}
		}
	}

	if needLf {
		r.w.WriteByte('\n')
	}
}
