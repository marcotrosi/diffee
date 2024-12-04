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
	Version bool
	Help    bool
	Flat    bool
	All     bool
	Size    bool
	Time    bool
	CRC32   bool
	Info    bool
	Swap    bool
	Depth   int
	NoColor bool
	Exclude RegExes
	Include RegExes
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
	var LeftDirContents  []Entry
	var RightDirContents []Entry
	// >>>

	// parse cli args <<<
	flag.BoolVar(&Version  , "version", false, "print version"                         )
	flag.BoolVar(&Help     , "help"   , false, "print help"                            )
	flag.BoolVar(&Flat     , "flat"   , false, "print differences flat"                )
	flag.BoolVar(&All      , "all"    , false, "don't ignore dotfiles"                 )
	flag.BoolVar(&Size     , "size"   , false, "compare file size"                     )
	flag.BoolVar(&Time     , "time"   , false, "compare modification time"             )
	flag.BoolVar(&CRC32    , "crc32"  , false, "compare CRC32 checksum"                )
	flag.BoolVar(&Info     , "info"   , false, "print file diff info"                  )
	flag.BoolVar(&Swap     , "swap"   , false, "swap sides"                            )
	flag.IntVar(&Depth     , "depth"  , 0    , "limit depth, 0 is no limit"            )
	flag.BoolVar(&NoColor  , "nocolor", false, "turn colored output off"               )
	flag.Var(&Exclude      , "exclude",        "exclude matching paths from diff"      )
	flag.Var(&Include      , "include",        "exclude non-matching paths from diff"  )
	flag.Parse()
	// >>>

	// check cli args <<<
	if flag.NArg() > 2 {
		printError("too many args")
		os.Exit(TOO_MANY_ARGS)
	}

	if Size {
		XORCnt += 1
	}
	if Time {
		XORCnt += 1
	}
	if CRC32 {
		XORCnt += 1
	}
	if XORCnt > 1 {
		printError("-size, -time and -crc32 are mutual exclusive, use only one")
		os.Exit(EXLUSIVE_OPTS)
	}
	// >>>

	// print help <<<
	if flag.NArg() == 0 || Help {
		printHelp()
		os.Exit(OK)
	}
	// >>>

	// print version <<<
	if Version {
		fmt.Println(Version_s)
		os.Exit(OK)
	}
	// >>>

	// no color <<<
	if NoColor {
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
		printError("left not a directory")
		os.Exit(NOT_A_DIR)
	}

	if isDirectory(Right) == false {
		printError("right not a directory")
		os.Exit(NOT_A_DIR)
	}
	// >>>

	// get dir contents <<<
	getUnionSetOfDirContents(Left, Right, &UnionSetOfDirContents)
	getDirContents(Left, Right, &UnionSetOfDirContents, &LeftDirContents, &RightDirContents)
	// >>>

	// print flat comparison
	if Flat {
		printFlat(&LeftDirContents, &RightDirContents)
		os.Exit(OK)
	}
	
	// start interactive comparison
	// if Interactive {
		// runInteractive(&LeftDirContents, &RightDirContents)
		// os.Exit(OK)
	// }

	// print side by side comparison
	printSideBySide(&LeftDirContents, &RightDirContents)
	os.Exit(OK)
}
