package main

import ( // <<<
	"os"
	"io/fs"
	"fmt"
	"sort"
	"time"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"strconv"
	"github.com/codingsince1985/checksum"
	"github.com/charmbracelet/lipgloss"
	"diffee/tree"
) // >>>

type Entry struct {// <<<
	// same for left and right
	NormPath   string
	Name       string
	IsDir      bool
	IsDotfile  bool
	IsDiff     bool

	// different for left and right
	Path       map[string]string
	Size       map[string]int64
	ModTime    map[string]time.Time
	Checksum   map[string]string

	IsMissing  map[string]bool
	IsOrphan   map[string]bool
	IsBigger   map[string]bool
	IsSmaller  map[string]bool
	IsNewer    map[string]bool
	IsOlder    map[string]bool
}

func (self Entry) String() string {
	return fmt.Sprintf("%s", self.Name)
}
// >>>

var ( // <<<
	NameRegEx *regexp.Regexp = regexp.MustCompile("[^/]+/?$")
	StyleRoot    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	StyleMissing = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	StyleOrphan  = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	StyleBigger  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	StyleSmaller = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	StyleNewer   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	StyleOlder   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	StyleDiff    = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
) // >>>

func printError(msg string) {// <<<
	fmt.Fprintln(os.Stderr, "diffee error: " + msg)
}// >>>

func setNoColor() {// <<<
	StyleRoot    = lipgloss.NewStyle()
	StyleMissing = lipgloss.NewStyle()
	StyleOrphan  = lipgloss.NewStyle()
	StyleBigger  = lipgloss.NewStyle()
	StyleSmaller = lipgloss.NewStyle()
	StyleNewer   = lipgloss.NewStyle()
	StyleOlder   = lipgloss.NewStyle()
	StyleDiff    = lipgloss.NewStyle()
}// >>>

func isDir(dirpath string) bool {// <<<
	return dirpath[len(dirpath)-1:] == "/"
}// >>>

func isDirectory(dirpath string) bool {// <<<
	fileInfo, err := os.Stat(dirpath)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}// >>>

func getUnionSetOfDirContents(left string, right string, ListOfPaths *[]string) {// <<<

	var Root string
	var SetOfPaths = make(map[string]struct{})

	WalkerFunc := func(fpath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fpath == Root {
			SetOfPaths["."] = struct{}{}
			return nil
		}

		fpath = path.Clean(strings.Replace(fpath, Root, "", 1))

		if Arg_Depth != 0 {
			if strings.Count(fpath, string(os.PathSeparator)) >= Arg_Depth {
				return filepath.SkipDir
			}
		}

		if Arg_All == false {
			NameChunk := NameRegEx.FindString(fpath)
			if NameChunk[:1] == "." {
				if info.IsDir() {
					return filepath.SkipDir
				} else {
					return nil
				}
			}
		}

		if info.IsDir() {
			fpath = fpath + "/"
		}

		if len(Arg_Include) > 0 {
			MatchFound := false
			for in:=0 ; in < len(Arg_Include) ; in++ {
				Match := Arg_Include[in].FindString(fpath)
				if Match != "" {
					MatchFound = true
				}
			}
			if MatchFound == false {
				return nil
			}
		}

		if len(Arg_Exclude) > 0 {
			for ex:=0 ; ex < len(Arg_Exclude) ; ex++ {
				Match := Arg_Exclude[ex].FindString(fpath)
				if Match != "" {
					if info.IsDir() {
						return filepath.SkipDir
					} else {
						return nil
					}
				}
			}
		}

		if Arg_Files {
			if info.IsDir() {
				return nil
			}
		}

		if Arg_Folders {
			if info.IsDir() == false {
				return nil
			}
		}

		if Arg_Files || (len(Arg_Include) > 0) {
			SplitPath := strings.SplitAfter(fpath, "/")
			CombinedPath := ""
			for i:=0; i < len(SplitPath); i++ {
				CombinedPath = CombinedPath + SplitPath[i]
				SetOfPaths[CombinedPath] = struct{}{}
			}
			return nil
		}

		SetOfPaths[fpath] = struct{}{}

		return nil
	}

	Root = left
	err := filepath.Walk(left, WalkerFunc)
	if err != nil {
		fmt.Println(err)
	}
	Root = right
	err = filepath.Walk(right, WalkerFunc)
	if err != nil {
		fmt.Println(err)
	}

	for p := range SetOfPaths {
		*ListOfPaths = append(*ListOfPaths, p)
	}
	sort.Strings(*ListOfPaths)

}// >>>

func getDirContentInformation(leftroot string, rightroot string, unionset *[]string, content *[]Entry) {// <<<

	*content = append(*content, Entry{ Path: map[string]string{ "left":  leftroot, "right": rightroot } } )

	for i:=1 ; i < len(*unionset) ; i++ {

		LeftPath  := leftroot  + (*unionset)[i]
		RightPath := rightroot + (*unionset)[i]
		NormPath  := (*unionset)[i]
		IsDotfile := false

		Name := NameRegEx.FindString(NormPath)
		if Name[:1] == "." {
			IsDotfile = true
		}

		LeftFileInfo , LErr := os.Stat(LeftPath)
		RightFileInfo, RErr := os.Stat(RightPath)

		var IsDir          bool      = isDir(Name)

		var LeftSize       int64     = 0
		var LeftModTime    time.Time = time.Time{}
		var LeftChecksum   string    = ""
		var LeftIsOrphan   bool      = false
		var LeftIsMissing  bool      = false
		var LeftIsBigger   bool      = false
		var LeftIsSmaller  bool      = false
		var LeftIsNewer    bool      = false
		var LeftIsOlder    bool      = false

		var RightSize      int64     = 0
		var RightModTime   time.Time = time.Time{}
		var RightChecksum  string    = ""
		var RightIsOrphan  bool      = false
		var RightIsMissing bool      = false
		var RightIsBigger  bool      = false
		var RightIsSmaller bool      = false
		var RightIsNewer   bool      = false
		var RightIsOlder   bool      = false

		if LErr != nil {
			LeftIsMissing = true
		} else if IsDir != LeftFileInfo.IsDir() {
			LeftIsMissing = true
		} else {
			// println(LeftPath, NormPath, Name, LeftFileInfo.Name())
			if IsDir == false {
				LeftSize       = LeftFileInfo.Size()
				LeftModTime    = LeftFileInfo.ModTime()
				LeftChecksum,_ = checksum.CRC32(LeftPath)
			}
		}

		if RErr != nil {
			RightIsMissing = true
		} else if IsDir != RightFileInfo.IsDir() {
			RightIsMissing = true
		} else {
			// println(RightPath, NormPath, Name, RightFileInfo.Name())
			if IsDir == false {
				RightSize       = RightFileInfo.Size()
				RightModTime    = RightFileInfo.ModTime()
				RightChecksum,_ = checksum.CRC32(RightPath)
			}
		}

		if LeftIsMissing {
			RightIsOrphan = true
		} else {
			if IsDir == false {
				LeftIsBigger  = (LeftSize > RightSize)
				LeftIsSmaller = (LeftSize < RightSize)
				LeftIsNewer   = LeftModTime.After(RightModTime)
				LeftIsOlder   = LeftModTime.Before(RightModTime)
			}
		}

		if RightIsMissing {
			LeftIsOrphan = true
		} else {
			if IsDir == false {
				RightIsBigger  = (RightSize > LeftSize)
				RightIsSmaller = (RightSize < LeftSize)
				RightIsNewer   = RightModTime.After(LeftModTime)
				RightIsOlder   = RightModTime.Before(LeftModTime)
			}
		}

		NewEntry := Entry{
			NormPath  : NormPath,
			Name      : Name,
			IsDir     : IsDir,
			IsDotfile : IsDotfile,
			IsDiff    : (LeftChecksum != RightChecksum),
                        
			Path     : map[string]string   { "left": LeftPath     , "right": RightPath      },
			Size     : map[string]int64    { "left": LeftSize     , "right": RightSize      },
			ModTime  : map[string]time.Time{ "left": LeftModTime  , "right": RightModTime   },
			Checksum : map[string]string   { "left": LeftChecksum , "right": RightChecksum  },
			IsMissing: map[string]bool     { "left": LeftIsMissing, "right": RightIsMissing },
			IsOrphan : map[string]bool     { "left": LeftIsOrphan , "right": RightIsOrphan  },
			IsBigger : map[string]bool     { "left": LeftIsBigger , "right": RightIsBigger  },
			IsSmaller: map[string]bool     { "left": LeftIsSmaller, "right": RightIsSmaller },
			IsNewer  : map[string]bool     { "left": LeftIsNewer  , "right": RightIsNewer   },
			IsOlder  : map[string]bool     { "left": LeftIsOlder  , "right": RightIsOlder   }}

		*content  = append(*content, NewEntry)
	}

}// >>>

func decorateText(entry *Entry, side string) string {// <<<

	if (*entry).IsMissing[side] {
		return StyleMissing.Render(strings.Repeat("░", len((*entry).Name)))
	}

	if (*entry).IsOrphan[side] {
		return StyleOrphan.Render((*entry).Name)
	}

	if Arg_Size {

		if Arg_Info {

			if (*entry).IsBigger[side] {
				return StyleBigger.Render((*entry).Name) + " (" + strconv.FormatInt(int64((*entry).Size[side]), 10) + " bytes)"
			}

			if (*entry).IsSmaller[side] {
				return StyleSmaller.Render((*entry).Name) + " (" + strconv.FormatInt(int64((*entry).Size[side]), 10) + " bytes)"
			}
		} else {

			if (*entry).IsBigger[side] {
				return StyleBigger.Render((*entry).Name)
			}

			if (*entry).IsSmaller[side] {
				return StyleSmaller.Render((*entry).Name)
			}
		}
	}

	if Arg_Time {

		if Arg_Info {

			if (*entry).IsNewer[side] {
				return StyleNewer.Render((*entry).Name) + " (" + (*entry).ModTime[side].Format(time.RFC3339) + ")"
			}

			if (*entry).IsOlder[side] {
				return StyleOlder.Render((*entry).Name) + " (" + (*entry).ModTime[side].Format(time.RFC3339) + ")"
			}
		} else {

			if (*entry).IsNewer[side] {
				return StyleNewer.Render((*entry).Name)
			}

			if (*entry).IsOlder[side] {
				return StyleOlder.Render((*entry).Name)
			}
		}
	}

	if Arg_CRC32 {

		if Arg_Info {

			if (*entry).IsDiff {
				return StyleDiff.Render((*entry).Name) + " (" + (*entry).Checksum[side] + ")"
			}
		} else {

			if (*entry).IsDiff {
				return StyleDiff.Render((*entry).Name)
			}
		}
	}

	return (*entry).Name
}// >>>

func convertSliceToTree(content *[]Entry, side string) *tree.Tree { // <<<

	var Result = tree.NewTree(StyleRoot.Render((*content)[0].Path[side]))
	var Stack []*tree.Node
	var LastDepth     int = 0
	var CurrentDepth  int = 0

	Stack = append(Stack, &(Result.Node))

	for i := 1; i < len(*content); i++ {
		
		E := (*content)[i]

		CurrentDepth = strings.Count(E.NormPath, "/") // count slashes
		if E.IsDir {
			CurrentDepth = CurrentDepth - 1
		}

		if CurrentDepth > LastDepth { // push new child onto stack
			Stack = append(Stack, Stack[LastDepth].GetChild(-1))
			LastDepth = CurrentDepth
		} else if CurrentDepth < LastDepth { // // pop from stack as many as we go directories upwards
			Stack = Stack[:len(Stack)-(LastDepth-CurrentDepth)]
			LastDepth = CurrentDepth
		}

		Stack[LastDepth].AddChild(&E).SetText(decorateText(&E, side))
	}

	return Result
}// >>>

func filterTrees(leftNode *tree.Node, rightNode *tree.Node) {// <<<
// THIS FUNCTION WAS GENERATED USING AI BASED ON A PREVIOUS FUNCTION.
// THE CODE SEEMS TO MAKE SENSE AND SEEMS TO WORK.

// filterTrees recursively filters *both* trees simultaneously to keep them aligned.
// It works "bottom-up" (post-order traversal).

	// --- 1. Recurse First (Bottom-Up) ---
	// We assume the trees have an identical structure, as they were built
	// from the same slice. We must iterate them together.
	if len(leftNode.GetChildren()) != len(rightNode.GetChildren()) {
		// This should never happen if build logic is correct, but it's a safe check.
		return
	}

	for i := 0; i < len(leftNode.GetChildren()); i++ {
		// GetChild() is 1-based, so we use i+1
		filterTrees(leftNode.GetChild(i+1), rightNode.GetChild(i+1))
	}

	// --- 2. Get Data & Handle Root ---
	// Root nodes (the paths) are never hidden.
	if leftNode.GetParent() == nil {
		return
	}

	// Both nodes point to the *same* Entry struct,
	// so we only need to get the data from one.
	data, ok := leftNode.GetData().(*Entry)
	if !ok {
		// This shouldn't happen, but it's safe to skip if it does.
		return
	}
	E := data // E for Entry

	var shouldHide bool = false // A single decision for both nodes

	// --- 3. Universal Filters (Orphans) ---
	// This logic is unchanged, as it's already based on the combined Entry.
	if Arg_Orphans && (!E.IsOrphan["left"] && !E.IsOrphan["right"]) {
		shouldHide = true
	} else if Arg_NoOrphans && (E.IsOrphan["left"] || E.IsOrphan["right"]) {
		shouldHide = true
	} else if Arg_LeftOrphans && !E.IsOrphan["left"] {
		// "Show only left orphans" -> hide if NOT a left orphan
		shouldHide = true
	} else if Arg_RightOrphans && !E.IsOrphan["right"] {
		// "Show only right orphans" -> hide if NOT a right orphan
		shouldHide = true
	}

	// --- 4. Type-Specific Filters ---
	if E.IsDir {
		// --- Directory-specific filters ---
		if Arg_Files { // Hide all folders if -files is set
			shouldHide = true
		}

		// --- Hide Empty Dirs ---
		// A directory is only hidden if it's considered empty on *both* sides.
		// Since child nodes are already filtered, CountChildren(true) is accurate.
		if !shouldHide && Arg_NoEmpty {
			if leftNode.CountChildren(true) == 0 && rightNode.CountChildren(true) == 0 {
				shouldHide = true
			}
		}

	} else {
		// --- File-specific filters ---
		if Arg_Folders { // Hide all files if -folders is set
			shouldHide = true
		}

		// Only check diff/same if the file isn't already hidden
		if !shouldHide {
			if Arg_Diff && (E.Checksum["left"] == E.Checksum["right"]) {
				// "Show only diff" -> hide if same
				shouldHide = true
			} else if Arg_Same && (E.Checksum["left"] != E.Checksum["right"]) {
				// "Show only same" -> hide if diff
				shouldHide = true
			}
		}
	}

	// --- 5. Apply the Filter (Synchronized) ---
	// Apply the *same* decision to both nodes.
	leftNode.HideNode(shouldHide)
	rightNode.HideNode(shouldHide)
}// >>>

// func resetTreeVisibility(node *tree.Node) {<<<
// 	node.HideNode(false)
// 	for _, child := range node.GetChildren() {
// 		resetTreeVisibility(child)
// 	}
// }>>>

func printSideBySide(contents *[]Entry) {// <<<

	var LeftTree  = convertSliceToTree(contents, "left")
	var RightTree = convertSliceToTree(contents, "right")
	var Output string

	filterTrees(&LeftTree.Node, &RightTree.Node)

	if Arg_Swap {
		LeftTree, RightTree = RightTree, LeftTree
	}

	Output = lipgloss.JoinHorizontal(lipgloss.Top, strings.Join(LeftTree.RenderTree(), "\n"), strings.Join(RightTree.SetRenderOffset(10).RenderTree(), "\n"))
	fmt.Println(Output)
}// >>>

func printPlain(contents *[]Entry, QuoteChar string) {// <<<
	leftroot  := (*contents)[0].Path["left"]
	rightroot := (*contents)[0].Path["right"]
	for i:=1; i < len(*contents); i++ {
		fmt.Printf("%s%s%s %s%s%s\n", QuoteChar, leftroot+(*contents)[i].Path["left"], QuoteChar, QuoteChar, rightroot+(*contents)[i].Path["right"], QuoteChar)
	}
}// >>>

// vim: fdm=marker fmr=<<<,>>>
