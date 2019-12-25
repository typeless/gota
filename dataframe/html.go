package dataframe

import (
	"io"
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type remainder struct {
	index int
	text  string
	nrows int
}

func readRows(trs []*html.Node) [][]string {
	rems := []remainder{}
	rows := [][]string{}
	for _, tr := range trs {
		xrems := []remainder{}
		row := []string{}
		index := 0
		text := ""
		for j, td := 0, tr.FirstChild; td != nil; j, td = j+1, td.NextSibling {
			if td.Type == html.ElementNode && td.DataAtom == atom.Td {

				for len(rems) > 0 {
					v := rems[0]
					if v.index > index {
						break
					}
					v, rems = rems[0], rems[1:]
					row = append(row, v.text)
					if v.nrows > 1 {
						xrems = append(xrems, remainder{v.index, v.text, v.nrows - 1})
					}
					index++
				}

				rowspan, colspan := 1, 1
				for _, attr := range td.Attr {
					switch attr.Key {
					case "rowspan":
						if k, err := strconv.Atoi(attr.Val); err == nil {
							rowspan = k
						}
					case "colspan":
						if k, err := strconv.Atoi(attr.Val); err == nil {
							colspan = k
						}
					}
				}
				for c := td.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.TextNode {
						text = strings.TrimSpace(c.Data)
					}
				}

				for k := 0; k < colspan; k++ {
					row = append(row, text)
					if rowspan > 1 {
						xrems = append(xrems, remainder{index, text, rowspan - 1})
					}
					index++
				}
			}
		}
		for j := 0; j < len(rems); j++ {
			v := rems[j]
			row = append(row, v.text)
			if v.nrows > 1 {
				xrems = append(xrems, remainder{v.index, v.text, v.nrows - 1})
			}
		}
		rows = append(rows, row)
		rems = xrems
	}
	for len(rems) > 0 {
		xrems := []remainder{}
		row := []string{}
		for i := 0; i < len(rems); i++ {
			v := rems[i]
			row = append(row, v.text)
			if v.nrows > 1 {
				xrems = append(xrems, remainder{v.index, v.text, v.nrows - 1})
			}
		}
		rows = append(rows, row)
		rems = xrems
	}
	return rows
}

func ReadHTML(r io.Reader, options ...LoadOption) []DataFrame {
	var err error
	var dfs []DataFrame
	var doc *html.Node
	var f func(*html.Node)

	doc, err = html.Parse(r)
	if err != nil {
		return []DataFrame{DataFrame{Err: err}}
	}

	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.DataAtom == atom.Table {
			trs := []*html.Node{}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && c.DataAtom == atom.Tbody {
					for cc := c.FirstChild; cc != nil; cc = cc.NextSibling {
						if cc.Type == html.ElementNode && (cc.DataAtom == atom.Th || cc.DataAtom == atom.Tr) {
							trs = append(trs, cc)
						}
					}
				}
			}

			df := LoadRecords(readRows(trs), options...)
			if df.Err == nil {
				dfs = append(dfs, df)
			}
			return
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)
	return dfs
}
