package domain

import (
	"fmt"
	"sort"
)

type Homepage struct {
	Title      string
	Date       string
	Categories []string
	Sources    []Source
	Articles   []Article
}

type Layout struct {
	Size        int
	Width       int
	TitleSize   int
	MaxChars    int
	MaxElements int
}

var (
	Layout1   = Layout{1, 2, 6, 200, 32}
	Layout2   = Layout{2, 2, 6, 500, 32}
	Layout3   = Layout{3, 2, 6, 2250, 32}
	Layout4   = Layout{4, 4, 4, 2800, 45}
	Layout5   = Layout{5, 4, 4, 4000, 45}
	Layout6   = Layout{6, 6, 2, 3300, 60}
	LayoutBar = Layout{0, 12, 0, 0, 0}
)

var layoutSequence = []Layout{
	Layout3,
	Layout6,
	Layout5,

	LayoutBar,

	Layout1,
	Layout1,
	Layout1,
	Layout1,
	Layout1,
	Layout1,

	Layout4,
	Layout4,
	Layout4,

	LayoutBar,

	Layout3,
	Layout3,
	Layout3,
	Layout6,

	LayoutBar,

	Layout6,
	Layout3,
	Layout3,
	Layout3,

	Layout3,
	Layout4,
	Layout4,
	Layout3,

	LayoutBar,

	Layout4,
	Layout3,
	Layout4,
	Layout3,

	Layout4,
	Layout3,
	Layout6,

	LayoutBar,

	Layout1,
	Layout1,
	Layout1,
	Layout1,
	Layout1,
	Layout1,

	Layout3,
	Layout6,
	Layout5,
}

func LayoutArticles(aa []Article) []Article {
	// first let's group articles by size
	bySize := make([][]Article, 6)
	for _, a := range aa {
		s := 0
		switch {
		case s < 200:
			bySize[0] = append(bySize[0], a)
		case 200 <= s && s < 500:
			bySize[1] = append(bySize[1], a)
		case 500 <= s && s < 3000:
			bySize[2] = append(bySize[2], a)
		case 3000 <= s && s < 5000:
			bySize[3] = append(bySize[3], a)
		case 5000 <= s && s < 7000:
			bySize[4] = append(bySize[4], a)
		default:
			bySize[5] = append(bySize[5], a)
		}
	}
	var out []Article
	// now iterate over our layout sequence
	for {
		for _, l := range layoutSequence {
			// full width? that's a horizontal bar
			if l.Width == 12 {
				out = append(out, Article{ID: "bar", Layout: l})
				continue
			}
			var sizes []int
			var picked *Article
			// which sizes of article fit in this layout?
			switch l.Size {
			case 1:
				sizes = []int{5, 4, 3, 2, 1, 0}
			case 2:
				sizes = []int{5, 4, 3, 2, 0, 1}
			case 3:
				sizes = []int{5, 4, 0, 1, 2, 3}
			case 4:
				sizes = []int{0, 1, 2, 3, 4, 5}
			case 5:
				sizes = []int{0, 1, 2, 3, 4, 5}
			case 6:
				sizes = []int{0, 1, 2, 3, 4, 5}
			}
			// iterate over compatible sizes
			for i := len(sizes) - 1; i >= 0; i-- {
				sized := bySize[sizes[i]]
				if len(sized) == 0 {
					continue
				}
				// organise articles of this size by source
				// this way we can mix up the sources for each
				// size rather than have them all from one source
				bySource := make(map[string][]Article)
				for _, a := range sized {
					bySource[a.Source.Name] = append(bySource[a.Source.Name], a)
				}
				o := [][]Article{}
				// sort the sourced slices by time, so
				// we prioritise more recent articles
				for _, s := range bySource {
					sort.Slice(s, func(i, j int) bool {
						return s[i].Timestamp.After(s[j].Timestamp)
					})
					o = append(o, s)
				}
				// make sure we iterate over the sources in the same order
				sort.Slice(o, func(i, j int) bool {
					return o[i][0].Source.Name < o[j][0].Source.Name
				})
				sorted := []Article{}
				for len(o) > 0 {
					for i := range o {
						// pop the first item off each
						// source onto the output slice
						sorted = append(sorted, o[i][0])
						o[i] = o[i][1:]
					}
					// have we used all the articles from a source?
					// we now delete all empty source slices
					for i, rlen := 0, len(o); i < rlen; i++ {
						j := i - (rlen - len(o))
						if len(o[j]) == 0 {
							o = append(o[:j], o[j+1:]...)
						}
					}
				}
				// assign the sorted slice back to the bySize array
				// so we can pop this article off the list and keep
				// track of it, whilst avoiding re-sorting
				bySize[sizes[i]] = sorted

				picked = &sorted[0]
				// pop the picked article off the list
				bySize[sizes[i]] = sorted[1:]
				break
			}
			if picked == nil {
				fmt.Println(len(out))
				return out
			}
			// set the layout on the picked article
			picked.Layout = l

			out = append(out, *picked)
		}
	}
}

func repeat(n int, l []Layout) []Layout {
	out := []Layout{}
	for i := 0; i < n; i++ {
		out = append(out, l...)
	}
	return out
}
