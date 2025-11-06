package main

import ( // <<<
	"os"
	"fmt"
	"path"
	"regexp"
	"github.com/spf13/cobra"
) // >>>

// global variables, constants and types <<<
const Tool_s    string = "diffee"
const Version_s string = "0.1.0"
const (
	OK int = iota
	INTERNAL
	CMDLINE
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
	Arg_NoEmpty      bool
	Arg_Diff         bool
	Arg_Same         bool
	Arg_Bash         bool
	Arg_Exclude      RegExes
	Arg_Include      RegExes
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

func (n *RegExes) Type() string {
    return "regex"
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
	rootCmd := &cobra.Command{
		Use:   Tool_s,
		Short: "Diff directories",
		Run: func(cmd *cobra.Command, args []string) {

			// check cli args <<<
			if Arg_Version {
				fmt.Println(Version_s)
				return
			}

			if Arg_Help {
				cmd.Help()
				return
			}

			if Arg_Bash {
				return
			}

			if len(args) > 2 {
				printError("too many arguments")
				os.Exit(TOO_MANY_ARGS)
			}

			if Arg_Size  { XORDiffType += 1 }
			if Arg_Time  { XORDiffType += 1 }
			if Arg_CRC32 { XORDiffType += 1 }
			if XORDiffType > 1 {
				printError("--size, --time and --crc32 are mutual exclusive, use only one")
				os.Exit(EXCLUSIVE_OPTS)
			}

			if Arg_Orphans      { XOROrphanType += 1 }
			if Arg_NoOrphans    { XOROrphanType += 1 }
			if Arg_LeftOrphans  { XOROrphanType += 1 }
			if Arg_RightOrphans { XOROrphanType += 1 }
			if XOROrphanType > 1 {
				printError("--orphans, --no-orphans, --left-orphans/--right-missing and --right-orphans/--left-missing can not be used together, use only one")
				os.Exit(EXCLUSIVE_OPTS)
			}

			if Arg_Diff && Arg_Same {
				printError("--diff and --same can not be used together, use only one")
				os.Exit(EXCLUSIVE_OPTS)
			}

			if Arg_Files && Arg_Folders {
				printError("--files and --folders can not be used together, use only one")
				os.Exit(EXCLUSIVE_OPTS)
			}
			// >>>

			// no color <<<
			if Arg_NoColor {
				setNoColor()
			}
			// >>>

			// get directory paths from args <<<
			switch len(args) {
			case 1:
				LeftDir = "./"
				RightDir = path.Clean(args[0]) + "/"
			case 2:
				LeftDir = path.Clean(args[0]) + "/"
				RightDir = path.Clean(args[1]) + "/"
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

		},
	}
	// >>>

	// commandline parameter definition <<<
	// flags (bools)
	rootCmd.Flags().BoolVarP(&Arg_Version      , "version"      , "v", false , "print version")
	rootCmd.Flags().BoolVarP(&Arg_Flat         , "flat"         , "T", false , "print differences flat")
	rootCmd.Flags().BoolVarP(&Arg_All          , "all"          , "a", false , "don't ignore dotfiles")
	rootCmd.Flags().BoolVarP(&Arg_Size         , "size"         , "s", false , "compare file size")
	rootCmd.Flags().BoolVarP(&Arg_Time         , "time"         , "t", false , "compare modification time")
	rootCmd.Flags().BoolVarP(&Arg_CRC32        , "crc32"        , "c", false , "compare CRC32 checksum")
	rootCmd.Flags().BoolVarP(&Arg_Info         , "info"         , "n", false , "print file diff info")
	rootCmd.Flags().BoolVarP(&Arg_Swap         , "swap"         , "x", false , "swap sides")
	rootCmd.Flags().IntVarP(&Arg_Depth         , "depth"        , "p", 0     , "limit depth, 0 is no limit")
	rootCmd.Flags().BoolVarP(&Arg_NoColor      , "no-color"     , "C", false , "turn colored output off")
	rootCmd.Flags().BoolVarP(&Arg_Orphans      , "orphans"      , "o", false , "show only orphans")
	rootCmd.Flags().BoolVarP(&Arg_NoOrphans    , "no-orphans"   , "O", false , "do not show orphans")
	rootCmd.Flags().BoolVarP(&Arg_LeftOrphans  , "left-orphans" , "l", false , "show only left orphans, same as --right-missing")
	rootCmd.Flags().BoolVarP(&Arg_LeftOrphans  , "right-missing", "R", false , "show only right missing, same as --left-orphans")
	rootCmd.Flags().BoolVarP(&Arg_RightOrphans , "right-orphans", "r", false , "show only right orphans, same as --left-missing")
	rootCmd.Flags().BoolVarP(&Arg_RightOrphans , "left-missing" , "L", false , "show only left missing, same as --right-orphans")
	rootCmd.Flags().BoolVarP(&Arg_Files        , "files"        , "f", false , "show only files")
	rootCmd.Flags().BoolVarP(&Arg_Folders      , "folders"      , "F", false , "show only folders")
	rootCmd.Flags().BoolVarP(&Arg_NoEmpty      , "no-empty"     , "E", false , "do not show empty folders")
	rootCmd.Flags().BoolVarP(&Arg_Diff         , "diff"         , "d", false , "show only files that differ")
	rootCmd.Flags().BoolVarP(&Arg_Same         , "same"         , "m", false , "show only files that are the same")
	rootCmd.Flags().BoolVarP(&Arg_Bash         , "bash"         , "b", false , "generate bash-completion script")

	// multi-value flags
	rootCmd.Flags().Var(&Arg_Exclude, "exclude", "exclude matching paths from diff")
	rootCmd.Flags().Var(&Arg_Include, "include", "exclude non-matching paths from diff")
	// >>>

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(CMDLINE)
	}

	if Arg_Bash {
		err := rootCmd.GenBashCompletionFile("diffee.bash")
		if err != nil {
			fmt.Println("Error generating bash completion:", err)
			os.Exit(INTERNAL)
		}
		fmt.Println("Generated diffee.bash")
		os.Exit(OK)
	}
}

// vim: fdm=marker fmr=<<<,>>>
