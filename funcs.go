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

var (
	Dirname_r *regexp.Regexp = regexp.MustCompile("[^/]+/?$") // TODO rename variable
	StyleRoot    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	StyleMissing = lipgloss.NewStyle().Background(lipgloss.Color("4"))
	StyleOrphan  = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	StyleBigger  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	StyleSmaller = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	StyleNewer   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	StyleOlder   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	StyleDiff    = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
)

func printHelp() {// <<<
  fmt.Println(" Usage of diffdir:")
  fmt.Println("-help")
  fmt.Println("  	print help")
  fmt.Println("-version")
  fmt.Println("  	print version")
}// >>>

func printError(msg string) {// <<<
	fmt.Fprintln(os.Stderr, Tool + " error: " + msg)
}// >>>

func isPath(fpath string) bool {// <<<
	_, err := os.Stat(fpath)
	if err != nil {
		return false
	} else {
		return true
	}
}// >>>

func isDirectory(dirpath string) bool {// <<<
	fileInfo, err := os.Stat(dirpath)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}// >>>

func isFile(filepath string) bool {// <<<
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return false
	}
	return !fileInfo.IsDir()
}// >>>

func getSize(filepath string) int64 {// <<<
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return 0
	}
	return fileInfo.Size()
}// >>>

func getModTime(filepath string) time.Time {// <<<
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return time.Now()
	}
	return fileInfo.ModTime()
}// >>>

func removeDuplicates(sortedSlice []string) []string {// <<<
    if len(sortedSlice) == 0 {
        return sortedSlice
    }

    // Initialize a new slice to store unique values
    result := []string{sortedSlice[0]} // Start with the first element

    for i := 1; i < len(sortedSlice); i++ {
        // If current element is different from the last element in the result
        if sortedSlice[i] != result[len(result)-1] {
            result = append(result, sortedSlice[i])
        }
    }
    return result
}// >>>

func getDirContents(left string, right string) []string {// <<<

	var ListOfPaths []string
	var Root string

	WalkerFunc := func(fpath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fpath == Root {
			ListOfPaths = append(ListOfPaths, ".")
			return nil
		}

		fpath = path.Clean(strings.Replace(fpath, Root + "/", "", 1))
		if info.IsDir() {
			ListOfPaths = append(ListOfPaths, fpath + "/")
		} else {
			ListOfPaths = append(ListOfPaths, fpath)
		}

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

	sort.Strings(ListOfPaths)
	ListOfPaths = removeDuplicates(ListOfPaths)

	return ListOfPaths
}// >>>

func colorizeElement(element string, root string, compareroot string) string {// <<<
	var Result string
	var Dirname_s string = Dirname_r.FindString(element) // TODO rename variable

	if isPath(root + element) && !isPath(compareroot + element) {
		Result = StyleOrphan.Render(Dirname_s)
	} else if !isPath(root + element) && isPath(compareroot + element) {
		Result = StyleMissing.Render("      ")
	} else {

		if isFile(root + element) {

			if size {
				filesize    := getSize(root + element)
				comparesize := getSize(compareroot + element)

				if filesize > comparesize {
					Result = StyleBigger.Render(Dirname_s) + " (" + strconv.FormatInt(int64(filesize), 10) + " bytes)"
				} else if filesize < comparesize {
					Result = StyleSmaller.Render(Dirname_s) + " (" + strconv.FormatInt(int64(filesize), 10) + " bytes)"
				} else {
					Result = Dirname_s
				}
			} else if date {
				filedate    := getModTime(root + element)
				comparedate := getModTime(compareroot + element)

				if filedate.After(comparedate) {
					Result = StyleNewer.Render(Dirname_s) + " (" + filedate.Format(time.RFC3339) + ")"
				} else if filedate.Before(comparedate) {
					Result = StyleOlder.Render(Dirname_s) + " (" + filedate.Format(time.RFC3339) + ")"
				} else {
					Result = Dirname_s
				}
			} else if crc32 {
				filechecksum   ,_ := checksum.CRC32(root + element)        // TODO eval error
				comparechecksum,_ := checksum.CRC32(compareroot + element) // TODO eval error
				if filechecksum != comparechecksum {
					Result = StyleDiff.Render(Dirname_s)
				} else {
					Result = Dirname_s
				}
			} else {
				Result = Dirname_s
			}

		} else {
			Result = Dirname_s
		}
	}
	
	return Result
}// >>>

func convertSliceToTree(union []string, root string, compareroot string) *tree.Tree {// <<<

	var Result *tree.Tree = tree.Root(StyleRoot.Render(root))
	var Stack []*tree.Tree
	var LastDepth    int = 0
	var CurrentDepth int = 0
	var ColoredElement_s string

	Stack = append(Stack, Result)

	for i := 1; i < len(union); i++ {

		element := union[i]

		if element[len(element)-1:] == "/" {
			CurrentDepth = len(strings.Split(element[:len(element)-1], "/")) - 1
		} else {
			CurrentDepth = len(strings.Split(element, "/")) - 1
		}

		if CurrentDepth > LastDepth {
			LastChildren := Stack[LastDepth].Children()
			Stack = append(Stack, LastChildren.At(LastChildren.Length()-1).(*tree.Tree))
			LastDepth = CurrentDepth
		}

		if CurrentDepth < LastDepth {
			Stack = Stack[:len(Stack)-1]
			LastDepth = CurrentDepth
		}

		ColoredElement_s = colorizeElement(element, root, compareroot)

		if element[len(element)-1:] == "/" {
			Stack[LastDepth].Child(tree.Root(ColoredElement_s))
		} else {
			Stack[LastDepth].Child(ColoredElement_s)
		}

	}
	return Result
}// >>>

func printSideBySide(union []string, leftroot string, rightroot string) {// <<<
	var lefttree  = convertSliceToTree(union, leftroot, rightroot)
	var righttree = convertSliceToTree(union, rightroot, leftroot)
	var whitespace string = "          "
	var offset []string
	for range union {
		offset = append(offset, whitespace)
	}
	var str string = lipgloss.JoinHorizontal(lipgloss.Top, lefttree.String(), strings.Join(offset[:], "\n"), righttree.String())
	fmt.Println(str)
}// >>>

