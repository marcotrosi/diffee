package tree

import (// <<<
	"fmt"
	"strings"
)// >>>

type Node struct {// <<<
	text         string
	hidenode     bool
	hidechildren bool
	parent       *Node
	children   []*Node
}// >>>

func (self *Node) GetText() string {// <<<
	return self.text
}// >>>

func (self *Node) SetText(txt string) *Node {// <<<
	self.text = txt
	return self
}// >>>

func (self *Node) IsHidden() bool {// <<<
	return self.hidenode
}// >>>

func (self *Node) HideNode(h bool) *Node {// <<<
	self.hidenode = h
	return self
}// >>>

func (self *Node) HideChildren(h bool) *Node {// <<<
	self.hidechildren = h
	return self
}// >>>

func (self *Node) GetParent() *Node {// <<<
	return self.parent
}// >>>

func (self *Node) GetChildren() []*Node {// <<<
	return self.children
}// >>>

func (self *Node) GetChild(n int) *Node {// <<<
	// child n= 1 is at index 0
	// child n=-1 is the last child, len()-1
	if n == 0 {
		return nil
	}

	var Index = 0
	var NumOfChildren = len(self.children)

	if n < 0 { // counting from end
		Index = NumOfChildren + n
	} else { // n > 0, n == 0 already handled before
		Index = n - 1
	}

	if Index > (NumOfChildren-1) { // out of range
		return nil
	}

	return self.children[Index]
}// >>>

func (self *Node) CountChildren(visible bool) int {// <<<
	if visible == false { // count all children
		return len(self.children)
	} else { // count only not hidden children
		var Counter = 0
		for _,c := range self.children {
			if c.hidenode == false {
				Counter = Counter + 1
			}
		}
		return Counter
	}
}// >>>

func (self *Node) AddChild(txt string) *Node {// <<<
	Newborn := &Node{parent: self, text: txt, hidenode: false, hidechildren: false}
	self.children = append(self.children, Newborn)
	return Newborn
}// >>>

func (self *Node) AddSibling(txt string) *Node {// <<<
	Newborn := &Node{parent: self.parent, text: txt, hidenode: false, hidechildren: false}
	self.parent.children = append(self.parent.children, Newborn)
	return Newborn
}// >>>

type Tree struct {// <<<
	Node
	offset int
	renderline func(self *Tree, lvl *[]Level, node *Node) string
}// >>>

type Level struct {// <<<
	currentchild int
	lastchild    int
}// >>>

func RenderTabsStyle(self *Tree, lvl *[]Level, node *Node) string {// <<<
// Tree
// 	Foo
// 	Bar
// 		Baz
// 	Qux
	return fmt.Sprintf("%s%s", strings.Repeat("\t", len(*lvl)), node.text)
}// >>>

func RenderNumberedStyle(self *Tree, lvl *[]Level, node *Node) string {// <<<
// Tree
// 1. Foo
// 2. Bar
//    2.1 Baz
// 3. Qux
	var indent string = ""
	var number string = ""
	var curlvl int    = len(*lvl)

	for i,l := range *lvl {
		if (i+1) < curlvl { // handle prior indent levels
			number = fmt.Sprintf("%s%d.", number, l.currentchild)
		} else {            // handle current indent level
			indent = strings.Repeat("   ", curlvl)
			number = fmt.Sprintf("%s%d. ", number, l.currentchild)
		}
	}

	return fmt.Sprintf("%s%s%s", indent, number, node.text)
}// >>>

func RenderTreeStyle(self *Tree, lvl *[]Level, node *Node) string {// <<<
// Tree
// ├── Foo
// ├── Bar
// │   ├── Bar
// │   └── Baz
// └── Qux
	var treechars string = ""
	var curlvl    int    = len(*lvl)

	for i,l := range *lvl {
		if (i+1) < curlvl { // handle prior indent levels
			if (l.currentchild < l.lastchild){
				treechars = fmt.Sprintf("%s%s", treechars, "│   ")
			} else {
				treechars = fmt.Sprintf("%s%s", treechars, "    ")
			}
		} else {            // handle current indent level
			if (l.currentchild < l.lastchild){
				treechars = fmt.Sprintf("%s%s", treechars, "├── ")
			} else {
				treechars = fmt.Sprintf("%s%s", treechars, "└── ")
			}
		}
	}

	return fmt.Sprintf("%s%s%s", strings.Repeat(" ", self.offset), treechars, node.text)
}// >>>

func RenderFolderStyle(self *Tree, lvl *[]Level, node *Node) string {// <<<
// ▼ Tree
//	  ▼ Foo
//	    ▼ Bar
//	        Baz
//	  ▶ Qux
   var foldersign string
	if node.CountChildren(true) == 0 {
		foldersign = "  "
	} else {
		if node.hidechildren == false {
			foldersign = "▼ "
		} else {
			foldersign = "▶ "
		}
	}

	return fmt.Sprintf("%s%s%s", strings.Repeat("   ", len(*lvl)), foldersign, node.text)
}// >>>

func NewTree(txt string) *Tree {// <<<
	t := Tree{}
	t.SetText(txt)
	t.parent = nil
	t.offset = 0
	t.renderline = RenderTreeStyle
	return &t
}// >>>

func (self *Tree) SetRenderStyle(f func(self *Tree, lvl *[]Level, node *Node) string) *Tree {// <<<
	self.renderline = f
	return self
}// >>>

func (self *Tree) SetRenderOffset(o int) *Tree {// <<<
	self.offset = o
	return self
}// >>>

func (self *Tree) RenderTree() []string {// <<<
	var Lines  []string
	var Levels []Level
	var RenderTreeHelper func(n *Node)

	RenderTreeHelper = func(n *Node) {

		Lines = append(Lines, self.renderline(self, &Levels, n))

		if n.hidechildren == false {

			Levels = append(Levels, Level{currentchild: 0, lastchild: n.CountChildren(true)}) // add level
			for _,c := range n.children {
				if c.hidenode == false {
					Levels[len(Levels)-1].currentchild++ // set current child number
					RenderTreeHelper(c)
				}
			}
			Levels = Levels[:len(Levels)-1] // remove level
		}
	}

	RenderTreeHelper(&self.Node)

	return Lines
}// >>>

// vim: fdm=marker fmr=<<<,>>>
