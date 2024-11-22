
tool folder             -> compare to current dir
tool folder1 folder2    -> compare folders

--side-by-side default so not needed
-f/--flat or -u/--unified
-i/--interactive
   - copy
   - delete
   - exclude
   - orphans
   - non-orphans
   - open file diff view
   - compare to
   - set as root
   - navigate tree
   - open/close folders
-c/--color
-d/--depth
-t/--time
-s/--size
--dirs-first
--md5 or diff/cmp
--exclude
--ignore
--respect-vcs-ignore
-a/--all
--filediffcmd="icdiff {} {}"  or env var
-v/--version
-h/--help

"io/ioutil"
"unicode/utf8"

func determineIndents(s string) int {// <<<
	var S = strings.Split(s, "/")
	var L = len(S)
	if S[L-1] == "" {
		return L-2
	} else {
		return L-1
	}
}// >>>

func getDirContent(dirpath string) []Entry {// <<<
	var Result []Entry
	 files, err := ioutil.ReadDir(dirpath)
    if err != nil {
        printError("Could not read path")
    }

    for _, file := range files {
		 var entry = Entry{Path: dirpath + "/" + file.Name(), Name: file.Name(), IsDir: file.IsDir(), Size: file.Size(), ModTime: file.ModTime()}
		 if entry.IsDir {
			 entry.Content = getDirContent(entry.Path)
		 }
       Result = append(Result, entry)
    }
	return Result
}// >>>

func printDirectory(content []Entry, indent int) {// <<<
	for _,e:= range content {
        fmt.Println(strings.Repeat("  ", indent), e.Path)
		  if e.IsDir {
			  printDirectory(e.Content, indent + 1)
		  }
    }
}// >>>

func getDirContent2(dirpath string) []Entry2 {// <<<
	var Result []Entry2

	WalkerFunc := func(fpath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		var entry = Entry2{Path: fpath, Name: info.Name(), IsDir: info.IsDir(), Size: info.Size(), ModTime: info.ModTime()}
		Result = append(Result, entry)

		return nil
	}

	var err = filepath.Walk(dirpath, WalkerFunc)
	if err != nil {
		fmt.Println(err)
	}

	return Result
}// >>>

func printSideBySide(union []string, leftroot string, leftcontent []Entry2, rightroot string, rightcontent []Entry2) {// <<<

	var LeftOutput []string
	var RightOutput []string
	// var SplitPath []string
	var Longest int = 0
	var Indents int = 0
	var Line string
	var FPath string
	var LastElement_s string
	var LastElement_r *regexp.Regexp = regexp.MustCompile("[^/]+/?$")

	// loop left side
	for i := range union {

		if i == 0 {
			// Line = union[i]
			Line = leftroot
			LeftOutput = append(LeftOutput, Line)
			Longest = max(utf8.RuneCountInString(Line), Longest)
			continue
		}

		FPath = leftroot + union[i]
		Indents = determineIndents(union[i])
		// getting last element
			// variant 1 - regex based
		LastElement_s = LastElement_r.FindString(union[i])
			// variant 2 - split based
			// t.b.d
		if isPath(FPath) {
			Line = strings.Repeat("    ", Indents) + "├── " + LastElement_s
		} else {
			Line = strings.Repeat("    ", Indents) + "├── "
		}
		Longest = max(utf8.RuneCountInString(Line), Longest)
		LeftOutput = append(LeftOutput, Line)
	}

	var Offset = Longest + 10

	// loop right side
	for i := range union {

		if i == 0 {
			// Line = union[i]
			Line = rightroot
			RightOutput = append(RightOutput, strings.Repeat(" ", Offset - utf8.RuneCountInString(LeftOutput[i])) + Line)
			continue
		}

		FPath = rightroot + union[i]
		Indents = determineIndents(union[i])
		LastElement_s = LastElement_r.FindString(union[i])
		if isPath(FPath) {
			Line = strings.Repeat("    ", Indents) + "├── " + LastElement_s
		} else {
			Line = strings.Repeat("    ", Indents) + "├── "
		}
		RightOutput = append(RightOutput,  strings.Repeat(" ", Offset - utf8.RuneCountInString(LeftOutput[i])) + Line)
	}

	for i := range LeftOutput {
		fmt.Println(LeftOutput[i] + RightOutput[i])
	}
}// >>>

// structs <<<
type Entry struct {
	Path string
	Name string
	IsDir bool
	Size int64
	ModTime time.Time
	Content []Entry
}

type Entry2 struct {
	Path string
	Name string
	IsDir bool
	Size int64
	ModTime time.Time
}
// >>>

