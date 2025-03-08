package main

import ( // <<<
	"os"
	"fmt"
	"flag"
	"path"
	"regexp"
) // >>>

// global variables, constants and types <<<
const Tool_s    string = "diffdir"
const Version_s string = "0.1.0"
const (
	OK int = iota
	INTERNAL
	TOO_MANY_ARGS
	NOT_A_DIR
	EXLUSIVE_OPTS
)
type RegExes []*regexp.Regexp
var (
	Arg_Version   bool
	Arg_Help      bool
	Arg_Flat      bool
	Arg_All       bool
	Arg_Size      bool
	Arg_Time      bool
	Arg_CRC32     bool
	Arg_Info      bool
	Arg_Swap      bool
	Arg_Depth     int
	Arg_NoColor   bool
	Arg_Orphans   bool
	Arg_NoOrphans bool
	Arg_Files     bool
	Arg_Diff      bool
	Arg_Same      bool
	Arg_Exclude   RegExes
	Arg_Include   RegExes
	Arg_Test      bool
)
// >>>

// local functions <<<
func (n *RegExes) String() string {
    return fmt.Sprintf("%s", *n)
}

func (n *RegExes) Set(value string) error {
    *n = append(*n, regexp.MustCompile(value))
    return nil
}
// >>>

func main() {

	// variables <<<
	var Left  string
	var Right string
	var XORCnt int = 0
	var UnionSetOfDirContents []string
	var DirContents []Entry
	// >>>

	// parse cli args <<<
	flag.BoolVar(&Arg_Version  , "version"   , false, "print version"                         )
	flag.BoolVar(&Arg_Help     , "help"      , false, "print help"                            )
	flag.BoolVar(&Arg_Flat     , "flat"      , false, "print differences flat"                )
	flag.BoolVar(&Arg_All      , "all"       , false, "don't ignore dotfiles"                 )
	flag.BoolVar(&Arg_Size     , "size"      , false, "compare file size"                     )
	flag.BoolVar(&Arg_Time     , "time"      , false, "compare modification time"             )
	flag.BoolVar(&Arg_CRC32    , "crc32"     , false, "compare CRC32 checksum"                )
	flag.BoolVar(&Arg_Info     , "info"      , false, "print file diff info"                  )
	flag.BoolVar(&Arg_Swap     , "swap"      , false, "swap sides"                            )
	flag.IntVar (&Arg_Depth    , "depth"     , 0    , "limit depth, 0 is no limit"            )
	flag.BoolVar(&Arg_NoColor  , "no-color"  , false, "turn colored output off"               )
	flag.BoolVar(&Arg_Orphans  , "orphans"   , false, "show only orphans"                     )
	flag.BoolVar(&Arg_NoOrphans, "no-orphans", false, "do not show orphans"                   )
	flag.BoolVar(&Arg_Files    , "files"     , false, "show only files, no empty dirs"        )
	flag.BoolVar(&Arg_Diff     , "diff"      , false, "show only files that differ"           )
	flag.BoolVar(&Arg_Same     , "same"      , false, "show only files that are the same"     )
	flag.Var    (&Arg_Exclude  , "exclude"   ,        "exclude matching paths from diff"      )
	flag.Var    (&Arg_Include  , "include"   ,        "exclude non-matching paths from diff"  )
	flag.BoolVar(&Arg_Test     , "test"      , false, "testing Hide function"                 )
	flag.Parse()
	// >>>

	// check cli args <<<
	if flag.NArg() > 2 {
		printError("too many args")
		os.Exit(TOO_MANY_ARGS)
	}

	if Arg_Test {
		Testing()
		os.Exit(OK)
	}

	if Arg_Size {
		XORCnt += 1
	}
	if Arg_Time {
		XORCnt += 1
	}
	if Arg_CRC32 {
		XORCnt += 1
	}
	if XORCnt > 1 {
		printError("-size, -time and -crc32 are mutual exclusive, use only one")
		os.Exit(EXLUSIVE_OPTS)
	}

	if Arg_Orphans && Arg_NoOrphans {
		printError("-orphans and -noorphans can not be used together, use only one")
		os.Exit(EXLUSIVE_OPTS)
	}

	if Arg_Diff && Arg_Same {
		printError("-diff and -same can not be used together, use only one")
		os.Exit(EXLUSIVE_OPTS)
	}
	// >>>

	// print version <<<
	if Arg_Version {
		fmt.Println(Version_s)
		os.Exit(OK)
	}
	// >>>

	// print help <<<
	if flag.NArg() == 0 || Arg_Help {
		printHelp()
		os.Exit(OK)
	}
	// >>>

	// no color <<<
	if Arg_NoColor {
		setNoColor()
	}
	// >>>

	// get directory paths from args <<<
	if flag.NArg() == 1 {
		Left  = "./"
		Right = path.Clean(flag.Arg(0)) + "/"
	} else { // 2 args given
		Left  = path.Clean(flag.Arg(0)) + "/"
		Right = path.Clean(flag.Arg(1)) + "/"
	}
	// >>>

	// check if dirs exists <<<
	if isDirectory(Left) == false {
		printError("left is not a directory")
		os.Exit(NOT_A_DIR)
	}

	if isDirectory(Right) == false {
		printError("right is not a directory")
		os.Exit(NOT_A_DIR)
	}
	// >>>

	// get dir contents <<<
	getUnionSetOfDirContents(Left, Right, &UnionSetOfDirContents)
	getDirContents(Left, Right, &UnionSetOfDirContents, &DirContents)
	// >>>

	// print flat comparison <<<
	if Arg_Flat {
		printFlat(&DirContents)
		os.Exit(OK)
	}// >>>
	
	// start interactive comparison <<<
	// if Interactive {
		// runInteractive(&DirContents)
		// os.Exit(OK)
	// } // >>>

	// print side by side comparison <<<
	printSideBySide(&DirContents)
	os.Exit(OK)
	// >>>

}

// vim: fdm=marker fmr=<<<,>>>
