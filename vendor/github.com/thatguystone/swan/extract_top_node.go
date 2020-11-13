package swan

import (
	"math"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type extractTopNode struct {
	a *Article
}

const (
	maxStepsFromNode = 3
	minStopwordCount = 5
)

var (
	linkTags     = cascadia.MustCompile("a")
	nodesToCheck = cascadia.MustCompile("p, pre, td")
)

func (e extractTopNode) run(a *Article) error {
	var ccs []*contentCache

	e.a = a

	for _, n := range a.Doc.FindMatcher(nodesToCheck).Nodes {
		cc := e.a.getCCache(n)
		if cc.stopwords > 2 && !cc.highLinkDensity {
			ccs = append(ccs, cc)
		}
	}

	startingBoost := 1.0
	bottomNegativeScore := int(float32(len(ccs)) * 0.25)

	for i, cc := range ccs {
		boostScore := 0.0
		if i > 0 && e.isBoostable(cc) {
			boostScore = (1.0 / startingBoost) * 50
			startingBoost++
		}

		if len(ccs) > 15 {
			if (len(ccs) - i) <= bottomNegativeScore {
				booster := float64(bottomNegativeScore - (len(ccs) - i))
				boostScore = -math.Pow(booster, 2.0)
				if math.Abs(boostScore) > 40 {
					boostScore = 5.0
				}
			}
		}

		upscore := int(cc.stopwords) + int(boostScore)

		p := cc.s.Nodes[0].Parent
		if p == nil {
			continue
		}

		score := a.scores[p]
		a.scores[p] = score + upscore

		p = p.Parent
		if p != nil {
			pscore, _ := a.scores[p]
			a.scores[p] = pscore + (upscore / 2)
		}
	}

	var topNode *html.Node
	topScore := 0
	for n, score := range a.scores {
		if score > topScore {
			topNode = n
			topScore = score
		}

		if topNode == nil {
			topNode = n
		}
	}

	if topNode != nil {
		a.TopNode = goquery.NewDocumentFromNode(topNode).Selection
	}

	return nil
}

func (e *extractTopNode) isBoostable(cc *contentCache) bool {
	stepsAway := 0
	for sib := cc.s.Nodes[0].PrevSibling; sib != nil; sib = sib.PrevSibling {
		if sib.Type == html.ElementNode && sib.DataAtom == atom.P {
			if stepsAway > maxStepsFromNode {
				return false
			}

			scc := e.a.getCCache(sib)
			if scc.stopwords > minStopwordCount {
				return true
			}
		}

		stepsAway++
	}

	return false
}
