# diffee

`diffee` is a commandline folder comparison tool that offers a variety of outputs.


## Background

I'm very much used to `Beyond Compare 4` at work, but they didn't provide a proper version for macOS and I also need
something for the commandline. In the meantime `Beyond Compare 5` was released. I wasn't able to find a commandline tool
in my package manager to diff directories in a similar fashion. Then I searched on Github and there are many repos with
names like `cmpdir`, `dircmp` or any other combination with the words `[cmp|diff]+[dir(s)|folder(s)|tree(s)]`, but none
of them provides screenshots or a good description, and from what I can tell they are not doing what I need.
Therefore I created this tool with tools like
[`Beyond Compare`](https://www.scootersoftware.com),
[`icdiff`](https://github.com/jeffkaufman/icdiff),
[`diff -y`](https://www.gnu.org/software/diffutils/) and
[`tree`](https://github.com/Old-Man-Programmer/tree) in mind, hence the name `diffee`.


## Disclaimer

This is the first and probably last time I used the _Go_ language and I have no clue if I used it properly. The reason why I used it
instead of _C_ is because I wanted to use the _charmbracelet_ libs for colored output and rendering the directory trees.
Unfortunately _liploss/tree_ [has/had some bugs](https://github.com/charmbracelet/lipgloss/discussions/452),
so I decided to write my own tree renderer.


## Overview

By default `diffee` does a static side-by-side comparison with a colored tree view. But it also offers a flat view
(`-flat`) which is useful to pipe the text into the next command.


## Usage

    diffee [options] [left_dir] right_dir

Compare `left_dir` to `right_dir`. If `left_dir` is omitted, the current working directory is used as `left_dir`.


## Options

| Option                       | Description                                                                                                                                                              | 
|------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-
| `--version/-v`               | Print version.                                                                                                                                                           | 
| `--help/-h`                  | Print help.                                                                                                                                                              | 
| `--flat/-f`                  | Print flat diff, without tree view.                                                                                                                                      | 
| `--all/-a`                   | By default hidden folders/files (dotfiles) are ignored, this option turns this behavior off.                                                                             | 
| `--depth <value>/-d <value>` | By default directory trees are traversed recursively all the way down, which is the same as `-depth 0`. But the depth can also be limited by providing a non-zero value. | 
| `--info`                     | Print information on differences (checksum, size, modtime).                                                                                                              | 
| `--swap/-w`                  | Swap sides.                                                                                                                                                              | 
| `--no-color/-n`              | Print without colors.                                                                                                                                                    | 
| `--include <regex>`          | Include paths that match the regex pattern. Can be used multiple times.                                                                                                  | 
| `--exclude <regex>`          | Exclude paths that match the regex pattern. If `--include` is used `--exclude` is applied on paths matching the _include regex_. Can be used multiple times.             | 
| `--files`                    | Only show files, don't care about empty dirs.                                                                                                                            | 
| `--crc32`                    | Use crc32 checksum to detect if files are different.                                                                                                                     | 
| `--size/-s`                  | Use file size to detect if files are different.                                                                                                                          | 
| `--time/-t`                  | Use modification time to detect if files are different.                                                                                                                  | 
| `--no-orphans/-O`            | Do not show orphans.                                                                                                                                                     | 
| `--orphans/-o`               | Only show orphans.                                                                                                                                                       | 
| `--diff`                     | Show only differences, hide same files.                                                                                                                                  | 
| `--same`                     | Show only same files, hide files with differences.                                                                                                                       | 


## Ideas

- `--interactive/-i` interactive mode to bring it much closer to `Beyond Compare`.
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
	- swap sides
	- --filediffcmd="icdiff {} {}"  or env var
- `-u/--unified`
- `--respect-vcs-ignore`
- second `--all` or `-A` to also not skip .git folders?
- use better args parser? https://pkg.go.dev/github.com/akamensky/argparse
- `--ignore <regex>` Do not check for differences on paths that match the regex pattern. do I really need this?
- auto-depth - to not descent into folders that don't have differences, to reduce print output
- Maybe the performance can be improved by using multi-threading?
- default diff detection by size followed by checksum if size is same
- combine diff states and use enums
- use no color hint given at Github
- should I replace the booleans with bits?

