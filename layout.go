package main

import "sort"

type Layout struct {
	Size      int
	Width     int
	TitleSize int
	Height    int
}

var (
	Layout1   = Layout{1, 2, 6, 220}
	Layout2   = Layout{2, 2, 6, 300}
	Layout3   = Layout{3, 2, 6, 800}
	Layout4   = Layout{4, 4, 4, 800}
	Layout5   = Layout{5, 4, 4, 800}
	Layout6   = Layout{6, 6, 2, 800}
	LayoutBar = Layout{0, 12, 0, 0}
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
}

func LayoutArticles(aa []Article) []Article {
	bySize := make([][]Article, 6)
	for _, a := range aa {
		s := len(a.Content)
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
	for _, l := range layoutSequence {
		if l.Width == 12 {
			out = append(out, Article{Layout: l})
			continue
		}
		var sizes []int
		var picked *Article
		switch l.Size {
		case 1:
			sizes = []int{0}
		case 2:
			sizes = []int{0, 1}
		case 3:
			sizes = []int{1, 2, 3}
		case 4:
			sizes = []int{3, 4, 5}
		case 5:
			sizes = []int{4, 5}
		case 6:
			sizes = []int{4, 5}
		}
		for i := len(sizes) - 1; i >= 0; i-- {
			sized := bySize[sizes[i]]
			if len(sized) == 0 {
				continue
			}
			sort.Slice(sized, func(i, j int) bool {
				return sized[i].Timestamp.After(sized[j].Timestamp)
			})
			picked = &sized[0]
			// pop the picked article off the list
			bySize[sizes[i]] = sized[1:]
		}
		if picked == nil {
			break
		}
		picked.Layout = l
		out = append(out, *picked)
	}
	return out
}
