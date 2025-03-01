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
	"github.com/charmbracelet/lipgloss/tree"
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
	var HelpText []string = []string {
		"Usage of diffee:",
		"  -all",
		"      don't ignore dotfiles",
		"  -crc32",
		"      compare CRC32 checksum",
		"  -depth int",
		"      limit depth, 0 is no limit",
		"  -diff",
		"      show only files that differ",
		"  -exclude value",
		"      exclude matching paths from diff",
		"  -files",
		"      show only files, no empty dirs",
		"  -flat",
		"      print differences flat",
		"  -help",
		"      print help",
		"  -include value",
		"      exclude non-matching paths from diff",
		"  -info",
		"      print file diff info",
		"  -no-color",
		"      turn colored output off",
		"  -no-orphans",
		"      do not show orphans",
		"  -orphans",
		"      show only orphans",
		"  -same",
		"      show only files that are the same",
		"  -size",
		"      compare file size",
		"  -swap",
		"      swap sides",
		"  -time",
		"      compare modification time",
		"  -version",
		"      print version" }

	for i := range(HelpText) {
		fmt.Println(HelpText[i])
	}
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

func removeDuplicates(sortedSlice *[]string) {// TODO delete <<<

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

func getDirContents(leftroot string, rightroot string, unionset *[]string, content *[]Entry) {// <<<

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
			NormPath       : NormPath,
			Name           : Name,
			IsDir          : IsDir,
			IsDotfile      : IsDotfile,
			IsDiff         : (LeftChecksum != RightChecksum),
                        
			Path: map[string]string{ "left":  LeftPath, "right": RightPath },

			Size: map[string]int64{ "left":  LeftSize, "right": RightSize },

			ModTime: map[string]time.Time{ "left":  LeftModTime, "right": RightModTime },

			Checksum: map[string]string{ "left":  LeftChecksum, "right": RightChecksum },

			IsMissing: map[string]bool{ "left":  LeftIsMissing, "right": RightIsMissing },

			IsOrphan: map[string]bool{ "left":  LeftIsOrphan, "right": RightIsOrphan },

			IsBigger: map[string]bool{ "left":  LeftIsBigger, "right": RightIsBigger },

			IsSmaller: map[string]bool{ "left":  LeftIsSmaller, "right": RightIsSmaller },

			IsNewer: map[string]bool{ "left":  LeftIsNewer, "right": RightIsNewer },

			IsOlder: map[string]bool{ "left":  LeftIsOlder, "right": RightIsOlder }}

		// if Arg_Orphans {
		// 	if (!(LeftIsMissing || LeftIsOrphan) && (LeftIsDir == false)) { // TODO write better and fix path issue to avoid check for files
		// 		continue
		// 	}
		// }
      //
		// if Arg_NoOrphans {
		// 	if (LeftIsMissing || LeftIsOrphan) { // TODO write better
		// 		continue
		// 	}
		// }
      //
		// if Arg_Diff && (LeftEntry.IsDir == false) && (RightChecksum == LeftChecksum) { // TODO write better
		// 	continue
		// } 
      //
		// if Arg_Same && (LeftEntry.IsDir == false) && (RightChecksum != LeftChecksum) { // TODO write better
		// 	continue
		// }

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

func convertSliceToTree(content *[]Entry, side string) *tree.Tree {// <<<

	var Result *tree.Tree = tree.Root(StyleRoot.Render((*content)[0].Path[side]))
	var Stack []*tree.Tree
	var LastDepth    int = 0
	var CurrentDepth int = 0
	var DecoratedText string

	Stack = append(Stack, Result)

	for i := 1; i < len(*content); i++ {

		Entry := (*content)[i]

		CurrentDepth = strings.Count(Entry.NormPath, "/")
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

		DecoratedText = decorateText(&Entry, side)

		if Entry.IsDir {
			Stack[LastDepth].Child(tree.Root(DecoratedText))
		} else {
			Stack[LastDepth].Child(DecoratedText)
		}

	}
	return Result
}// >>>

func printSideBySide(contents *[]Entry) {// <<<

	var LeftTree  = convertSliceToTree(contents, "left")
	var RightTree = convertSliceToTree(contents, "right")
	var Whitespace string = strings.Repeat(" ", 10)
	var Offset []string
	var Output string
	for range *contents {
		Offset = append(Offset, Whitespace)
	}

	if Arg_Swap {
		Output = lipgloss.JoinHorizontal(lipgloss.Top, RightTree.String(), strings.Join(Offset[:], "\n"), LeftTree.String())
	} else {
		Output = lipgloss.JoinHorizontal(lipgloss.Top, LeftTree.String(), strings.Join(Offset[:], "\n"), RightTree.String())
	}
	fmt.Println(Output)
}// >>>

func printFlat(contents *[]Entry) {// <<<
	leftroot  := (*contents)[0].Path["left"]
	rightroot := (*contents)[0].Path["right"]
	for i:=1; i < len(*contents); i++ {
		fmt.Printf("%q %q\n", leftroot+(*contents)[i].Path["left"], rightroot+(*contents)[i].Path["right"])
	}
}// >>>

func Testing() {// <<<
	var T *tree.Tree = tree.Root("root")
	T.Child("first")
	T.Child("second")
	T.Child(tree.NewLeaf("third", false))
	fmt.Println(T)
}// >>>
