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
	"github.com/charmbracelet/lipgloss/v2/tree"
	"github.com/charmbracelet/lipgloss"
) // >>>

type Entry struct {// <<<
	Path      string
	Name      string
	IsDir     bool
	Size      int64
	ModTime   time.Time
	IsMissing bool
	IsOrphan  bool
	IsBigger  bool
	IsSmaller bool
	IsNewer   bool
	IsOlder   bool
	IsDiff    bool
	IsDotfile bool
	Checksum  string
}// >>>

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

func printHelp() {// <<<
	fmt.Println(" Usage of diffdir:")
	fmt.Println("-help")
	fmt.Println("  	print help")
	fmt.Println("-version")
	fmt.Println("  	print version")
}// >>>

func printError(msg string) {// <<<
	fmt.Fprintln(os.Stderr, Tool_s + " error: " + msg)
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

func removeDuplicates(sortedSlice *[]string) {// <<<

	if len(*sortedSlice) == 0 {
		return
	}

	var WrIdx int = 0

	for RdIdx := 1; RdIdx < len(*sortedSlice); RdIdx++ {

		if (*sortedSlice)[RdIdx] != (*sortedSlice)[WrIdx] {
			WrIdx = WrIdx + 1
		}

		(*sortedSlice)[WrIdx] = (*sortedSlice)[RdIdx]
	}

	 (*sortedSlice) = (*sortedSlice)[:WrIdx+1]
}// >>>

func getUnionSetOfDirContents(left string, right string, ListOfPaths *[]string) {// <<<

	var Root string

	WalkerFunc := func(fpath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fpath == Root {
			*ListOfPaths = append(*ListOfPaths, ".")
			return nil
		}

		fpath = path.Clean(strings.Replace(fpath, Root, "", 1))

		if Depth != 0 {
			if strings.Count(fpath, string(os.PathSeparator)) >= Depth {
				return filepath.SkipDir
			}
		}

		if All == false {
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

		if len(Include) > 0 {
			MatchFound := false
			for in:=0 ; in < len(Include) ; in++ {
				Match := Include[in].FindString(fpath)
				if Match != "" {
					MatchFound = true
				}
			}
			if MatchFound == false {
				return nil
			}
		}

		if len(Exclude) > 0 {
			for ex:=0 ; ex < len(Exclude) ; ex++ {
				Match := Exclude[ex].FindString(fpath)
				if Match != "" {
					if info.IsDir() {
						return filepath.SkipDir
					} else {
						return nil
					}
				}
			}
		}

		if Files {
			if info.IsDir() {
				return nil
			}
		}

		if Files || (len(Include) > 0) {
			// TODO maybe find another way as it produces lots of data, it also relies on the fact that we sort-unique at the end
			SplitPath := strings.SplitAfter(fpath, "/")
			CombinedPath := ""
			for i:=0; i < len(SplitPath); i++ {
				CombinedPath = CombinedPath + SplitPath[i]
				*ListOfPaths = append(*ListOfPaths, CombinedPath)
			}
			return nil
		}

		*ListOfPaths = append(*ListOfPaths, fpath)

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

	sort.Strings(*ListOfPaths)
	removeDuplicates(ListOfPaths)

}// >>>

func getDirContents(leftroot string, rightroot string, unionset *[]string, leftcontent *[]Entry, rightcontent *[]Entry) {// <<<

	*leftcontent  = append(*leftcontent , Entry{Path : leftroot })
	*rightcontent = append(*rightcontent, Entry{Path : rightroot})

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

		var LeftIsDir      bool      = isDir(Name)
		var LeftSize       int64     = 0
		var LeftModTime    time.Time = time.Time{}
		var LeftChecksum   string    = ""
		var LeftIsOrphan   bool      = false
		var LeftIsMissing  bool      = false
		var LeftIsBigger   bool      = false
		var LeftIsSmaller  bool      = false
		var LeftIsNewer    bool      = false
		var LeftIsOlder    bool      = false

		var RightIsDir     bool      = isDir(Name)
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
		} else if LeftIsDir != LeftFileInfo.IsDir() {
			LeftIsMissing = true
		} else {
			if LeftIsDir == false {
				LeftSize       = LeftFileInfo.Size()
				LeftModTime    = LeftFileInfo.ModTime()
				LeftChecksum,_ = checksum.CRC32(LeftPath)
			}
		}

		if RErr != nil {
			RightIsMissing = true
		} else if RightIsDir != RightFileInfo.IsDir() {
			RightIsMissing = true
		} else {
			if RightIsDir == false {
				RightSize       = RightFileInfo.Size()
				RightModTime    = RightFileInfo.ModTime()
				RightChecksum,_ = checksum.CRC32(RightPath)
			}
		}

		if LeftIsMissing {
			RightIsOrphan = true
		} else {
			if LeftIsDir == false {
				LeftIsBigger  = (LeftSize > RightSize)
				LeftIsSmaller = (LeftSize < RightSize)
				LeftIsNewer   = LeftModTime.After(RightModTime)
				LeftIsOlder   = LeftModTime.Before(RightModTime)
			}
		}

		if RightIsMissing {
			LeftIsOrphan = true
		} else {
			if RightIsDir == false {
				RightIsBigger  = (RightSize > LeftSize)
				RightIsSmaller = (RightSize < LeftSize)
				RightIsNewer   = RightModTime.After(LeftModTime)
				RightIsOlder   = RightModTime.Before(LeftModTime)
			}
		}

		LeftEntry := Entry{
			Path      : NormPath,
			Name      : Name,
			IsMissing : LeftIsMissing,
			IsOrphan  : LeftIsOrphan,
			IsDir     : LeftIsDir,
			Size      : LeftSize,
			ModTime   : LeftModTime,
			IsDotfile : IsDotfile,
			IsBigger  : LeftIsBigger,
			IsSmaller : LeftIsSmaller,
			IsNewer   : LeftIsNewer,
			IsOlder   : LeftIsOlder,
			IsDiff    : (LeftChecksum != RightChecksum),
			Checksum  : LeftChecksum }

		RightEntry := Entry{
			Path      : NormPath,
			Name      : Name,
			IsMissing : RightIsMissing,
			IsOrphan  : RightIsOrphan,
			IsDir     : RightIsDir,
			Size      : RightSize,
			ModTime   : RightModTime,
			IsDotfile : IsDotfile,
			IsBigger  : RightIsBigger,
			IsSmaller : RightIsSmaller,
			IsNewer   : RightIsNewer,
			IsOlder   : RightIsOlder,
			IsDiff    : (RightChecksum != LeftChecksum),
			Checksum  : RightChecksum }

		if Orphans {
			if (!(LeftIsMissing || LeftIsOrphan) && (LeftIsDir == false)) { // TODO write better and fix path issue to avoid check for files
				continue
			}
		}

		if NoOrphans {
			if (LeftIsMissing || LeftIsOrphan) { // TODO write better
				continue
			}
		}

		if Diff && (LeftEntry.IsDir == false) && (RightChecksum == LeftChecksum) { // TODO write better
			continue
		} 

		if Same && (LeftEntry.IsDir == false) && (RightChecksum != LeftChecksum) { // TODO write better
			continue
		}

		*leftcontent  = append(*leftcontent, LeftEntry)
		*rightcontent = append(*rightcontent, RightEntry)
	}

}// >>>

func decorateText(entry *Entry) string {// <<<

	if (*entry).IsMissing {
		return StyleMissing.Render(strings.Repeat("░", len((*entry).Name)))
	}

	if (*entry).IsOrphan {
		return StyleOrphan.Render((*entry).Name)
	}

	if Size {

		if Info {

			if (*entry).IsBigger {
				return StyleBigger.Render((*entry).Name) + " (" + strconv.FormatInt(int64((*entry).Size), 10) + " bytes)"
			}

			if (*entry).IsSmaller {
				return StyleSmaller.Render((*entry).Name) + " (" + strconv.FormatInt(int64((*entry).Size), 10) + " bytes)"
			}
		} else {

			if (*entry).IsBigger {
				return StyleBigger.Render((*entry).Name)
			}

			if (*entry).IsSmaller {
				return StyleSmaller.Render((*entry).Name)
			}
		}
	}

	if Time {

		if Info {

			if (*entry).IsNewer {
				return StyleNewer.Render((*entry).Name) + " (" + (*entry).ModTime.Format(time.RFC3339) + ")"
			}

			if (*entry).IsOlder {
				return StyleOlder.Render((*entry).Name) + " (" + (*entry).ModTime.Format(time.RFC3339) + ")"
			}
		} else {

			if (*entry).IsNewer {
				return StyleNewer.Render((*entry).Name)
			}

			if (*entry).IsOlder {
				return StyleOlder.Render((*entry).Name)
			}
		}
	}

	if CRC32 {

		if Info {

			if (*entry).IsDiff {
				return StyleDiff.Render((*entry).Name) + " (" + (*entry).Checksum + ")"
			}
		} else {

			if (*entry).IsDiff {
				return StyleDiff.Render((*entry).Name)
			}
		}
	}

	return (*entry).Name
}// >>>

func convertSliceToTree(content *[]Entry) *tree.Tree {// <<<

	var Result *tree.Tree = tree.Root(StyleRoot.Render((*content)[0].Path))
	var Stack []*tree.Tree
	var LastDepth    int = 0
	var CurrentDepth int = 0
	var DecoratedText string

	Stack = append(Stack, Result)

	for i := 1; i < len(*content); i++ {

		Entry := (*content)[i]

		CurrentDepth = strings.Count(Entry.Path, "/")
		if Entry.IsDir {
			CurrentDepth = CurrentDepth - 1
		}

		if CurrentDepth > LastDepth {
			LastChildren := Stack[LastDepth].Children()
			Stack = append(Stack, LastChildren.At(LastChildren.Length()-1).(*tree.Tree))
			LastDepth = CurrentDepth
		} else if CurrentDepth < LastDepth {
			Stack = Stack[:len(Stack)-(LastDepth-CurrentDepth)]
			LastDepth = CurrentDepth
		}

		DecoratedText = decorateText(&Entry)

		if Entry.IsDir {
			Stack[LastDepth].Child(tree.Root(DecoratedText))
		} else {
			Stack[LastDepth].Child(DecoratedText)
		}

	}
	return Result
}// >>>

func printSideBySide(left *[]Entry, right *[]Entry) {// <<<

	var LeftTree  = convertSliceToTree(left)
	var RightTree = convertSliceToTree(right)
	var Whitespace string = strings.Repeat(" ", 10)
	var Offset []string
	var Output string
	for range *left {
		Offset = append(Offset, Whitespace)
	}

	if Swap {
		Output = lipgloss.JoinHorizontal(lipgloss.Top, RightTree.String(), strings.Join(Offset[:], "\n"), LeftTree.String())
	} else {
		Output = lipgloss.JoinHorizontal(lipgloss.Top, LeftTree.String(), strings.Join(Offset[:], "\n"), RightTree.String())
	}
	fmt.Println(Output)
}// >>>

func printFlat(left *[]Entry, right *[]Entry) {// <<<
	leftroot  := (*left)[0].Path
	rightroot := (*right)[0].Path
	for i:=1; i < len(*left); i++ {
		fmt.Printf("%q %q\n", leftroot+(*left)[i].Path, rightroot+(*right)[i].Path)
	}
}// >>>
