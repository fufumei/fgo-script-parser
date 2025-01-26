A simple script to get the actual dialogue line and character count for FGO.

## Usage & Output

**NOTE:** This is a command-line script, and it will not wait for user input of any kind. In other words, even when using the `.exe` release, it needs to be run through the CLI as shown below.

You can either use this to fetch scripts directly from Atlas, or parse files stored locally on your device.

```
fgo-script-parser atlas [--war <ID>] [--quest <ID>] [--script <ID>] [--no-file]

fgo-script-parser local <path> [--no-file]
```

By default the result will output to `script-length.csv` in the same location as the script. If `--no-file` is present, the result will only be output to the command line.

Regardless of output destination, the format is a tab-separated list with the format:  
`name    total lines    total characters`.

When `atlas` is used, `--war`, `--quest`, and `--script` are mutually exclusive.  
Note also that, for the time being, it's not possible to enter multiple IDs at once.

When `local` is used, if the given path is a file (FGO story scripts are in `.txt` format by default), it will simply count the lines and characters and always output the result to the command-line.  
If the given path is a folder, the script will traverse every underlying path until it finds a file to open. It will then count the total lines and characters in the current folder, write the result to the output, and repeat for any remaining folders (**the script will likely not work if you have files and folders mixed on the same level**).

## How it works

### Regex matching

The script regex-matches dialogue lines with the following pattern: `(＠([A-Z][：:])?(.*)\n)(.*?\n(?:.*?\n)?)?(.*?)\n\[k\]|(？.+?：.+)`

`(＠([A-Z][：:])?(.*)\n)(.*?\n(?:.*?\n)?)?(.*?)\n\[k\]` should match any regular dialogue line, whether standard format, narration (no nametag), or interspersed with function tags.

`(？.+?：.+)` should match any player choice, including ones with special interactions, such as in LB6.

### Character counting

To clean the lines for counting, a pattern of `(\[[^#&]+?\]|[\[\]#&:]|？.+?：|^＠.+|\n)` is used.

This should clear out any regular function tags in the text (`[r]`, `[line 3]`, `[image *]` etc), while only removing the square brackets, hashtag, ampersand, and colon for any ruby tags (`[#計画:コ　ト]` becomes `計画コ　ト`) or gender tags (`[&ああ:うん]` becomes `ああうん`).  
The same goes for any emphasis tags, which are structured like ruby tags, but without the colon.

Where this presents an issue is where they sometimes will use image tags to insert text with a different font, but there is no possible way to count that using the source script itself, so this count is _as close as we can get_.

**I have no real way of really confirming the character count, as opposed to line count, but it looks correct compared to line count and previous data.**

### Special cases

For OC2, the appendix quest is not considered part of the quest list for war `403`, so it won't be fetched automatically when using Atlas. Because of this, there's a special clause when using `--war 403` which makes it also fetch the script from quest `4000327`, so that it's included. This quest should be part of whatever quest list the Bleached Earth has, so take note of that.

For the Ordeal Call Prologue, the two scripts `0400010110` and `0400019910` are actually the exact same script (one is just a redirect to the other, for some reason). The counting script doesn't take this into account, so you either need to divide the total ouput for this by 2, or simply remove one of the files from wherever you are counting (the total lines for this should be 95, for reference).
