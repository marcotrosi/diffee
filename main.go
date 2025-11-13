package main

import ( // <<<
	"os"
	"fmt"
	"path"
	"regexp"
	"github.com/spf13/cobra"
) // >>>

// global variables, constants and types <<<
const Version string = "0.3.0"

const (
	OK int = iota
	INTERNAL
	CMDLINE
	TOO_MANY_ARGS
	NOT_A_DIR
	EXCLUSIVE_OPTS
)

var QuoteChar string = ""

type RegExes []*regexp.Regexp
var (
	Arg_Version      bool
	Arg_Help         bool
	Arg_Plain        bool
	Arg_SingleQuotes bool
	Arg_DoubleQuotes bool
	Arg_All          bool
	Arg_Size         bool
	Arg_Time         bool
	Arg_CRC32        bool
	Arg_Info         bool
	Arg_Swap         bool
	// Arg_ShortenRoot  bool
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
	Arg_LeftAlias    string
	Arg_RightAlias   string
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
		Use:   "diffee [left_dir] <right_dir>",
		Short: "Diff directories",
		Run: func(cmd *cobra.Command, args []string) {

			// check cli args <<<
			if Arg_Version {
				fmt.Println(Version)
				return
			}

			if Arg_Bash {
				return
			}

			if Arg_Help || (len(args) == 0) {
				cmd.Help()
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
				printError("--orphans, --no-orphans, --left-orphans and --right-orphans can not be used together, use only one")
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

			if Arg_SingleQuotes && Arg_DoubleQuotes {
				printError("--single-quotes and --double-quotes can not be used together, use only one")
				os.Exit(EXCLUSIVE_OPTS)
			}
			// >>>

			// no color <<<
			if Arg_NoColor {
				setNoColor()
			} else {
				if Value, Exists := os.LookupEnv("NO_COLOR"); Exists && Value != "" {
					setNoColor()
				}
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
				printError(fmt.Sprintf("given path '%s' is not a directory", LeftDir))
				os.Exit(NOT_A_DIR)
			}

			if isDirectory(RightDir) == false {
				printError(fmt.Sprintf("given path '%s' is not a directory", RightDir))
				os.Exit(NOT_A_DIR)
			}
			// >>>

			// get dir contents <<<
			getUnionSetOfDirContents(LeftDir, RightDir, &UnionSetOfDirContents)
			getDirContentInformation(LeftDir, RightDir, &UnionSetOfDirContents, &DirContentInformation)
			// >>>

			// print plain output <<<
			if Arg_Plain {
				if Arg_SingleQuotes {
					QuoteChar = "'"
				}

				if Arg_DoubleQuotes {
					QuoteChar = "\""
				}

				printPlain(&DirContentInformation, QuoteChar)
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
	// general
	rootCmd.Flags().BoolVarP(&Arg_Version      , "version"      , "v", false , "print version")
	rootCmd.Flags().BoolVarP(&Arg_Bash         , "bash"         , "b", false , "generate bash-completion script")
	// control input
	rootCmd.Flags().BoolVarP(&Arg_All          , "all"          , "a", false , "don't ignore dotfiles")
	rootCmd.Flags().IntVarP(&Arg_Depth         , "depth"        , "D", 0     , "limit depth, 0 is no limit and the default")
	rootCmd.Flags().VarP(&Arg_Include          , "include"      , "I",         "include matching paths into diff, can be used multiple times, if --include and --exclude are used together then --include is applied first")
	rootCmd.Flags().VarP(&Arg_Exclude          , "exclude"      , "E",         "exclude matching paths from diff, can be used multiple times, if --include and --exclude are used together then --include is applied first")
	// control output
	rootCmd.Flags().BoolVarP(&Arg_Diff         , "diff"         , "d", false , "show only files that differ")
	rootCmd.Flags().BoolVarP(&Arg_Same         , "same"         , "m", false , "show only files that are the same")
	rootCmd.Flags().BoolVarP(&Arg_Files        , "files"        , "f", false , "show only files")
	rootCmd.Flags().BoolVarP(&Arg_Folders      , "folders"      , "F", false , "show only folders")
	rootCmd.Flags().BoolVarP(&Arg_NoEmpty      , "no-empty"     , "e", false , "do not show empty folders")
	rootCmd.Flags().BoolVarP(&Arg_Orphans      , "orphans"      , "o", false , "show only orphans")
	rootCmd.Flags().BoolVarP(&Arg_NoOrphans    , "no-orphans"   , "O", false , "do not show orphans")
	rootCmd.Flags().BoolVarP(&Arg_LeftOrphans  , "left-orphans" , "L", false , "show only left orphans")
	rootCmd.Flags().BoolVarP(&Arg_RightOrphans , "right-orphans", "R", false , "show only right orphans")
	rootCmd.Flags().BoolVarP(&Arg_Plain        , "plain"        , "p", false , "print differences in plain format, use --single-quotes/-q or --double-quotes/-Q to wrap in quotes, useful in combination with xargs")
	rootCmd.Flags().BoolVarP(&Arg_SingleQuotes , "single-quotes", "q", false , "wrap plain output in single quotes")
	rootCmd.Flags().BoolVarP(&Arg_DoubleQuotes , "double-quotes", "Q", false , "wrap plain output in double quotes")
	// control comparison
	rootCmd.Flags().BoolVarP(&Arg_Size         , "size"         , "s", false , "compare file size")
	rootCmd.Flags().BoolVarP(&Arg_Time         , "time"         , "t", false , "compare modification time")
	rootCmd.Flags().BoolVarP(&Arg_CRC32        , "crc32"        , "c", false , "compare CRC32 checksum")
	// control display
	rootCmd.Flags().BoolVarP(&Arg_Swap         , "swap"         , "x", false , "swap sides")
	rootCmd.Flags().BoolVarP(&Arg_Info         , "info"         , "n", false , "print file diff info")
	rootCmd.Flags().BoolVarP(&Arg_NoColor      , "no-color"     , "C", false , "turn colored output off, overwrites NO_COLOR")
	rootCmd.Flags().StringVarP(&Arg_LeftAlias  , "left-alias"   , "l", ""    , "display the given string as left root folder name")
	rootCmd.Flags().StringVarP(&Arg_RightAlias , "right-alias"  , "r", ""    , "display the given string as right root folder name")
	// rootCmd.Flags().BoolVarP(&Arg_ShortenRoot  , "shorten-root" , "S", false , "shorten the root path if possible")
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
