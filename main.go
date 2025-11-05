package main

import ( // <<<
	"os"
	"fmt"
	"flag"
	"path"
	"regexp"
) // >>>

// global variables, constants and types <<<
const Tool_s    string = "diffee"
const Version_s string = "0.1.0"
const (
	OK int = iota
	INTERNAL
	TOO_MANY_ARGS
	NOT_A_DIR
	EXCLUSIVE_OPTS
)
type RegExes []*regexp.Regexp
var (
	Arg_Version      bool
	Arg_Help         bool
	Arg_Flat         bool
	Arg_All          bool
	Arg_Size         bool
	Arg_Time         bool
	Arg_CRC32        bool
	Arg_Info         bool
	Arg_Swap         bool
	Arg_Depth        int
	Arg_NoColor      bool
	Arg_Orphans      bool
	Arg_NoOrphans    bool
	Arg_LeftOrphans  bool
	Arg_RightOrphans bool
	Arg_Files        bool
	Arg_Folders      bool
	Arg_HideEmpty    bool
	Arg_Diff         bool
	Arg_Same         bool
	Arg_Exclude      RegExes
	Arg_Include      RegExes
	Arg_Test         bool
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
	var LeftDir               string
	var RightDir              string
	var XORDiffType           int = 0
	var XOROrphanType         int = 0
	var UnionSetOfDirContents []string
	var DirContentInformation []Entry
	// >>>

	// parse cli args <<<
	flag.BoolVar(&Arg_Version       , "version"         , false, "print version"                                 )
	flag.BoolVar(&Arg_Help          , "help"            , false, "print help"                                    )
	flag.BoolVar(&Arg_Flat          , "flat"            , false, "print differences flat"                        )
	flag.BoolVar(&Arg_All           , "all"             , false, "don't ignore dotfiles"                         )
	flag.BoolVar(&Arg_Size          , "size"            , false, "compare file size"                             )
	flag.BoolVar(&Arg_Time          , "time"            , false, "compare modification time"                     )
	flag.BoolVar(&Arg_CRC32         , "crc32"           , false, "compare CRC32 checksum"                        )
	flag.BoolVar(&Arg_Info          , "info"            , false, "print file diff info"                          )
	flag.BoolVar(&Arg_Swap          , "swap"            , false, "swap sides"                                    )
	flag.IntVar (&Arg_Depth         , "depth"           , 0    , "limit depth, 0 is no limit"                    )
	flag.BoolVar(&Arg_NoColor       , "no-color"        , false, "turn colored output off"                       )
	flag.BoolVar(&Arg_Orphans       , "orphans"         , false, "show only orphans"                             )
	flag.BoolVar(&Arg_NoOrphans     , "no-orphans"      , false, "do not show orphans"                           )
	flag.BoolVar(&Arg_LeftOrphans   , "left-orphans"    , false, "show only left orphans, same as -right-missing")
	flag.BoolVar(&Arg_LeftOrphans   , "right-missing"   , false, "show only right missing, same as -left-orphans")
	flag.BoolVar(&Arg_RightOrphans  , "right-orphans"   , false, "show only right orphans, same as -left-missing")
	flag.BoolVar(&Arg_RightOrphans  , "left-missing"    , false, "show only left missing, same as -right-orphans")
	flag.BoolVar(&Arg_Files         , "files"           , false, "show only files, no empty folders"             )
	flag.BoolVar(&Arg_Folders       , "folders"         , false, "show only folders"                             )
	flag.BoolVar(&Arg_HideEmpty     , "hide-empty"      , false, "hide empty folders"                            )
	flag.BoolVar(&Arg_Diff          , "diff"            , false, "show only files that differ"                   )
	flag.BoolVar(&Arg_Same          , "same"            , false, "show only files that are the same"             )
	flag.Var    (&Arg_Exclude       , "exclude"         ,        "exclude matching paths from diff"              )
	flag.Var    (&Arg_Include       , "include"         ,        "exclude non-matching paths from diff"          )
	flag.BoolVar(&Arg_Test          , "test"            , false, "testing Hide function"                         )
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

	if Arg_Size  { XORDiffType += 1 }
	if Arg_Time  { XORDiffType += 1 }
	if Arg_CRC32 { XORDiffType += 1 }
	if XORDiffType > 1 {
		printError("-size, -time and -crc32 are mutual exclusive, use only one")
		os.Exit(EXCLUSIVE_OPTS)
	}

	if Arg_Orphans      { XOROrphanType += 1 }
	if Arg_NoOrphans    { XOROrphanType += 1 }
	if Arg_LeftOrphans  { XOROrphanType += 1 }
	if Arg_RightOrphans { XOROrphanType += 1 }
	if XOROrphanType > 1 {
		printError("-orphans, -no-orphans, -left-orphans/-right-missing and -right-orphans/-left-missing can not be used together, use only one")
		os.Exit(EXCLUSIVE_OPTS)
	}

	if Arg_Diff && Arg_Same {
		printError("-diff and -same can not be used together, use only one")
		os.Exit(EXCLUSIVE_OPTS)
	}

	if Arg_Files && Arg_Folders {
		printError("-files and -folders can not be used together, use only one")
		os.Exit(EXCLUSIVE_OPTS)
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
		LeftDir  = "./"
		RightDir = path.Clean(flag.Arg(0)) + "/"
	} else { // 2 args given
		LeftDir  = path.Clean(flag.Arg(0)) + "/"
		RightDir = path.Clean(flag.Arg(1)) + "/"
	}
	// >>>

	// check if dirs exists <<<
	if isDirectory(LeftDir) == false {
		printError("left is not a directory")
		os.Exit(NOT_A_DIR)
	}

	if isDirectory(RightDir) == false {
		printError("right is not a directory")
		os.Exit(NOT_A_DIR)
	}
	// >>>

	// get dir contents <<<
	getUnionSetOfDirContents(LeftDir, RightDir, &UnionSetOfDirContents)
	getDirContentInformation(LeftDir, RightDir, &UnionSetOfDirContents, &DirContentInformation)
	// >>>

	// print flat comparison <<<
	if Arg_Flat {
		printFlat(&DirContentInformation)
		os.Exit(OK)
	}// >>>
	
	// start interactive comparison <<<
	// if Interactive {
		// runInteractive(&DirContentInformation)
		// os.Exit(OK)
	// } // >>>

	// print side by side comparison <<<
	printSideBySide(&DirContentInformation)
	os.Exit(OK)
	// >>>

}

// vim: fdm=marker fmr=<<<,>>>
