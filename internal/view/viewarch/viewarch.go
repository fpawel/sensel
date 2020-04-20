package viewarch

import (
	"fmt"
	"github.com/lxn/walk"
	"sort"
	"time"
)

func NewTreeViewModel(arch []time.Time) walk.TreeModel {
	x := new(treeModel)
	for _, y := range archYears(arch) {
		ndY := &treeNode{
			parent: nil,
			txt:    fmt.Sprintf("%d", y),
			Year:   y,
		}
		for _, mo := range archMonths(arch, y) {
			ndMo := &treeNode{
				parent: ndY,
				txt:    fmt.Sprintf("%02d", mo),
				Year:   y,
				Month:  mo,
			}
			for _, d := range archDays(arch, y, mo) {
				ndMo.children = append(ndMo.children, &treeNode{
					parent: ndMo,
					txt:    fmt.Sprintf("%02d - %d", d, archDayCount(arch, y, mo, d)),
					Year:   y,
					Month:  mo,
					Day:    d,
				})
			}
			ndY.children = append(ndY.children, ndMo)
		}
		x.children = append(x.children, ndY)
	}
	return x
}

func GetItemDate(x walk.TreeItem) (int, time.Month, int, bool) {
	y, ok := x.(*treeNode)
	if !ok {
		return 0, 0, 0, false
	}
	return y.Year, time.Month(y.Month), y.Day, y.Level() == 2
}

type treeModel struct {
	walk.TreeModelBase
	children []walk.TreeItem
}

type treeNode struct {
	parent walk.TreeItem
	txt    string
	Year,
	Month,
	Day int
	children []walk.TreeItem
}

var _ walk.TreeModel = new(treeModel)
var _ walk.TreeItem = new(treeNode)

func (m *treeModel) RootCount() int {
	return len(m.children)
}

func (m *treeModel) RootAt(i int) walk.TreeItem {
	return m.children[i]
}

func (x *treeNode) Text() string {
	return x.txt
}

func (x *treeNode) Parent() walk.TreeItem {
	return x.parent
}

func (x *treeNode) ChildCount() int {
	return len(x.children)
}

func (x *treeNode) ChildAt(i int) walk.TreeItem {
	return x.children[i]
}

func (x *treeNode) Level() (l int) {
	var y walk.TreeItem = x
	for ; y.Parent() != nil; y = y.Parent() {
		l++
	}
	return
}

func archYears(arch []time.Time) (xs []int) {
	mp := make(map[int]struct{})
	for _, t := range arch {
		mp[t.Year()] = struct{}{}
	}
	for y := range mp {
		xs = append(xs, y)
	}
	sort.Ints(xs)
	return
}

func archMonths(arch []time.Time, y int) (xs []int) {
	mp := make(map[int]struct{})
	for _, t := range arch {
		if t.Year() == y {
			mp[int(t.Month())] = struct{}{}
		}
	}
	for m := range mp {
		xs = append(xs, m)
	}
	sort.Ints(xs)
	return
}

func archDays(arch []time.Time, y, mo int) (xs []int) {
	mp := make(map[int]struct{})
	for _, t := range arch {
		if t.Year() == y && int(t.Month()) == mo {
			mp[t.Day()] = struct{}{}
		}
	}
	for m := range mp {
		xs = append(xs, m)
	}
	sort.Ints(xs)
	return
}

func archDayCount(arch []time.Time, y, mo, d int) (count int) {
	for _, t := range arch {
		if t.Year() == y && int(t.Month()) == mo && t.Day() == d {
			count++
		}
	}
	return
}
