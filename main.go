package main

import ( // <<<
	"os"
	"fmt"
	"flag"
	"path"
) // >>>

// global variables and constants <<<
const Tool    string = "diffdir"
const Version string = "0.1.0"
const (
	OK int = iota
	INTERNAL
	TOO_MANY_ARGS
	NOT_A_DIR
	EXLUSIVE_OPTS
)
var (
	version bool
	help bool
	size bool
	date bool
	crc32 bool
)
// >>>

func main() {

	// variables <<<
	var Left  string
	var Right string
	var XORCnt int = 0
	// >>>

	// parse cli args <<<
	flag.BoolVar(&version, "version", false, "print version"            )
	flag.BoolVar(&help   , "help"   , false, "print help"               )
	flag.BoolVar(&size   , "size"   , false, "compare file size"        )
	flag.BoolVar(&date   , "date"   , false, "compare modification date")
	flag.BoolVar(&crc32  , "crc32"  , false, "compare CRC32 checksum"   )
	flag.Parse()
	// >>>

	// check cli args <<<
	if flag.NArg() > 2 {
		printError("too many args")
		os.Exit(TOO_MANY_ARGS)
	}

	if size {
		XORCnt += 1
	}
	if date {
		XORCnt += 1
	}
	if crc32 {
		XORCnt += 1
	}
	if XORCnt > 1 {
		printError("-size, -date and -crc32 are mutual exclusive, use only one")
		os.Exit(EXLUSIVE_OPTS)
	}
	// >>>

	// print help <<<
	if flag.NArg() == 0 || help {
		printHelp()
		os.Exit(OK)
	}
	// >>>

	// print version <<<
	if version {
		fmt.Println(Version)
		os.Exit(OK)
	}
	// >>>

	// get directory paths from args <<<
	if flag.NArg() == 1 {
		Left  = "."
		Right = path.Clean(flag.Arg(0))
	} else { // 2 args given
		Left  = path.Clean(flag.Arg(0))
		Right = path.Clean(flag.Arg(1))
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
	var UnionSetOfDirContents = getDirContents(Left, Right)
	// >>>

	// print flat comparison
	
	// start interactive comparison

	// print side by side comparison
	printSideBySide(UnionSetOfDirContents, Left + "/", Right + "/")
}
