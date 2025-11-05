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

|             Option             | Description                                                                                                                                                              |
|--------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--version`/`-v`               | Print version.                                                                                                                                                           |
| `--help`/`-h`                  | Print help.                                                                                                                                                              |
| `--flat`/`-I`                  | Print flat diff, without tree view.                                                                                                                                      |
| `--all`/`-a`                   | By default hidden folders/files (dotfiles) are ignored, this option turns this behavior off.                                                                             |
| `--depth <value>`/`-p <value>` | By default directory trees are traversed recursively all the way down, which is the same as `-depth 0`. But the depth can also be limited by providing a non-zero value. |
| `--info`/`-i`                  | Print information on differences (checksum, size, modtime).                                                                                                              |
| `--swap`/`-x`                  | Swap sides.                                                                                                                                                              |
| `--no-color`/`-C`              | Print without colors.                                                                                                                                                    |
| `--include <regex>`            | Include paths that match the regex pattern. Can be used multiple times.                                                                                                  |
| `--exclude <regex>`            | Exclude paths that match the regex pattern. If `--include` is used `--exclude` is applied on paths matching the _include regex_. Can be used multiple times.             |
| `--files`/`-f`                 | Only show files.                                                                                                                                                         |
| `--folders`/`-F`               | Only show folders.                                                                                                                                                       |
| `--no-empty`/`-E`              | Do not show empty folders.                                                                                                                                               |
| `--crc32`/`-c`                 | Use crc32 checksum to detect if files are different.                                                                                                                     |
| `--size`/`-s`                  | Use file size to detect if files are different.                                                                                                                          |
| `--time`/`-t`                  | Use modification time to detect if files are different.                                                                                                                  |
| `--orphans`/`-o`               | Only show orphans.                                                                                                                                                       |
| `--no-orphans`/`-O`            | Do not show orphans.                                                                                                                                                     |
| `--left-orphans`/`-l`          | Show only left orphans, same as --right-missing.                                                                                                                         |
| `--right-missing`/`-R`         | Show only right missing, same as --left-orphans.                                                                                                                         |
| `--right-orphans`/`-r`         | Show only right orphans, same as --left-missing.                                                                                                                         |
| `--left-missing`/`-L`          | Show only left missing, same as --right-orphans.                                                                                                                         |
| `--diff`/`-d`                  | Show only differences, hide same files.                                                                                                                                  |
| `--same`/`-m`                  | Show only same files, hide files with differences.                                                                                                                       |
| `--bash`/`-b`                  | Generate bash-completion script.                                                                                                                                         |


## ToDo

- default diff detection by size followed by checksum if size is same
- use no color hint given at Github
- config file for default flags and colors
- finish -flat, -q/-Q to wrap in ' and " quote
- combine diff states and use enums
- second `--all` or `--All/-A` to also not skip .git folders?
- `-I` like `--respect-vcs-ignore`
- how to handle big depths that don't fit on screen?
- ignore casing on Windows if ever supported, highlight orange if different
- ignore casing `-i/-c` can it be useful for unixoids?
- Maybe the performance can be improved by using multi-threading?

- document code/workflow/strategy
- write system and unit tests

- create diff report (html, pdf)
- `--interactive/-i` interactive mode to bring it much closer to `Beyond Compare`.
    - has to use some kind of view/window in case the tree doesn't fit on its half side
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
	- --filediffcmd="icdiff {} {}"  and/or config/env var

