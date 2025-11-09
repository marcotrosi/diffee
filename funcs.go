package main

// imports <<<
import (
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

// Enums <<<
type SizeDiffState int
const (
    SameSize SizeDiffState = iota
    Bigger
    Smaller
)

type TimeDiffState int
const (
	SameTime TimeDiffState = iota
	Newer
	Older
)
// >>>

// Entry struct <<<
type Entry struct {
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
	SizeDiff   map[string]SizeDiffState
	TimeDiff   map[string]TimeDiffState
}

func (self Entry) String() string {
	return fmt.Sprintf("%s", self.Name)
}
// >>>

// Variables <<<
var (
	NameRegEx *regexp.Regexp = regexp.MustCompile("[^/]+/?$")
	StyleRoot    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	StyleMissing = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	StyleOrphan  = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	StyleBigger  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	StyleSmaller = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	StyleNewer   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	StyleOlder   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	StyleDiff    = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))

	SizeStyles = map[SizeDiffState]lipgloss.Style{
		SameSize: lipgloss.NewStyle(),
		Bigger:   StyleBigger,
		Smaller:  StyleSmaller,
	}

	TimeStyles = map[TimeDiffState]lipgloss.Style{
		SameTime: lipgloss.NewStyle(),
		Newer:    StyleNewer,
		Older:    StyleOlder,
	}
)
// >>>

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
		var LeftSizeState  SizeDiffState = SameSize
		var LeftTimeState  TimeDiffState = SameTime

		var RightSize      int64     = 0
		var RightModTime   time.Time = time.Time{}
		var RightChecksum  string    = ""
		var RightIsOrphan  bool      = false
		var RightIsMissing bool      = false
		var RightSizeState  SizeDiffState = SameSize
		var RightTimeState  TimeDiffState = SameTime

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
				if LeftSize > RightSize {
					LeftSizeState = Bigger
				} else if LeftSize < RightSize {
					LeftSizeState = Smaller
				}

				if LeftModTime.After(RightModTime) {
					LeftTimeState = Newer
				} else if LeftModTime.Before(RightModTime) {
					LeftTimeState = Older
				}
			}
		}

		if RightIsMissing {
			LeftIsOrphan = true
		} else {
			if IsDir == false {
				if RightSize > LeftSize {
					RightSizeState = Bigger
				} else if RightSize < LeftSize {
					RightSizeState = Smaller
				}

				if RightModTime.After(LeftModTime) {
					RightTimeState = Newer
				} else if RightModTime.Before(LeftModTime) {
					RightTimeState = Older
				}
			}
		}

		NewEntry := Entry {

			NormPath  : NormPath,
			Name      : Name,
			IsDir     : IsDir,
			IsDotfile : IsDotfile,
			IsDiff    : (LeftChecksum != RightChecksum),

			Path     : map[string]string        { "left": LeftPath     , "right": RightPath      },
			Size     : map[string]int64         { "left": LeftSize     , "right": RightSize      },
			ModTime  : map[string]time.Time     { "left": LeftModTime  , "right": RightModTime   },
			Checksum : map[string]string        { "left": LeftChecksum , "right": RightChecksum  },
			IsMissing: map[string]bool          { "left": LeftIsMissing, "right": RightIsMissing },
			IsOrphan : map[string]bool          { "left": LeftIsOrphan , "right": RightIsOrphan  },
			SizeDiff : map[string]SizeDiffState { "left": LeftSizeState, "right": RightSizeState },
			TimeDiff : map[string]TimeDiffState { "left": LeftTimeState, "right": RightTimeState }}

			*content  = append(*content, NewEntry)
	}

}// >>>

func decorateText(entry *Entry, side string) string {// <<<

	if (*entry).IsMissing[side] {
		return StyleMissing.Render(strings.Repeat("░", len((*entry).Name)))
	}

	var Style lipgloss.Style = lipgloss.NewStyle()
	var Info string = ""

	if (*entry).IsOrphan[side] {
		Style = StyleOrphan

	} else {

		if Arg_Size {
			State := (*entry).SizeDiff[side]
			Style = SizeStyles[State]

			if Arg_Info && State != SameSize {
				Info = " (" + strconv.FormatInt(int64((*entry).Size[side]), 10) + " bytes)"
			}

		} else if Arg_Time {
			State := (*entry).TimeDiff[side]
			Style = TimeStyles[State]

			if Arg_Info && State != SameTime {
				Info = " (" + (*entry).ModTime[side].Format(time.RFC3339) + ")"
			}

		} else if Arg_CRC32 {
			if (*entry).IsDiff {
				Style = StyleDiff
				if Arg_Info {
					Info = " (" + (*entry).Checksum[side] + ")"
				}
			}
		}
	}

	return Style.Render((*entry).Name) + Info
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
			// 1. Determine the "isSame" status based on the active comparison mode
			var isSame bool = false // Default to "different"

			if Arg_Size {
				// It's "same" if BOTH sides show SameSize.
				// (Note: For orphans, one side won't be SameSize, so this is safe)
				isSame = (E.SizeDiff["left"] == SameSize && E.SizeDiff["right"] == SameSize)

			} else if Arg_Time {
				// It's "same" if BOTH sides show SameTime.
				isSame = (E.TimeDiff["left"] == SameTime && E.TimeDiff["right"] == SameTime)

			} else {
				// Default to CRC32 comparison (or if no mode is selected)
				// You can just use E.IsDiff here.
				isSame = !E.IsDiff 
			}

			// 2. Apply the filter logic
			if Arg_Diff && isSame {
				// "Show only diff" -> hide if same
				shouldHide = true
			} else if Arg_Same && !isSame {
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

	if Arg_LeftAlias != "" {
		LeftTree.Node.SetText(StyleRoot.Render(Arg_LeftAlias))
	}
	if Arg_RightAlias != "" {
		RightTree.Node.SetText(StyleRoot.Render(Arg_RightAlias))
	}

	filterTrees(&LeftTree.Node, &RightTree.Node)

	if Arg_Swap {
		LeftTree, RightTree = RightTree, LeftTree
	}

	Output = lipgloss.JoinHorizontal(lipgloss.Top, strings.Join(LeftTree.RenderTree(), "\n"), strings.Join(RightTree.SetRenderOffset(10).RenderTree(), "\n"))
	fmt.Println(Output)
}// >>>

func shouldHideEntry(E *Entry) bool {// <<<
// THIS FUNCTION WAS GENERATED USING AI BASED ON THE FILTERTREES() FUNCTION.
// THE CODE SEEMS TO MAKE SENSE AND SEEMS TO WORK.
	// --- 1. Universal Filters (Orphans) ---
	if Arg_Orphans && (!E.IsOrphan["left"] && !E.IsOrphan["right"]) {
		return true
	} else if Arg_NoOrphans && (E.IsOrphan["left"] || E.IsOrphan["right"]) {
		return true
	} else if Arg_LeftOrphans && !E.IsOrphan["left"] {
		return true
	} else if Arg_RightOrphans && !E.IsOrphan["right"] {
		return true
	}

	// --- 2. Type-Specific Filters ---
	if E.IsDir {
		if Arg_Files {
			return true
		}
		// Note: Arg_NoEmpty cannot be checked here easily for flat view 
		// without knowing about children. For flat view, we might just 
		// ignore --no-empty, or we'd need a pre-pass. 
		// For now, let's assume it doesn't apply to flat view or we accept it won't work there yet.
	} else {
		if Arg_Folders {
			return true
		}

		// --- 3. Diff/Same Logic ---
		var isSame bool = false
		if Arg_Size {
			isSame = (E.SizeDiff["left"] == SameSize && E.SizeDiff["right"] == SameSize)
		} else if Arg_Time {
			isSame = (E.TimeDiff["left"] == SameTime && E.TimeDiff["right"] == SameTime)
		} else {
			isSame = !E.IsDiff
		}

		if Arg_Diff && isSame {
			return true
		} else if Arg_Same && !isSame {
			return true
		}
	}

	return false
}// >>>

func printPlain(contents *[]Entry, QuoteChar string) {// <<<
	var Left  string = "left"
	var Right string = "right"

	if Arg_Swap {
		Left, Right= Right, Left
	}

	for i:=1; i < len(*contents); i++ {
      if !shouldHideEntry(&(*contents)[i]) {
         fmt.Printf("%s%s%s %s%s%s\n", QuoteChar, (*contents)[i].Path[Left], QuoteChar, QuoteChar, (*contents)[i].Path[Right], QuoteChar)
      }
	}
}// >>>

// vim: fdm=marker fmr=<<<,>>>
