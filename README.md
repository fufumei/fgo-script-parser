A very simple script to get the *actual* dialogue line count for any FGO story script.

### Regex matching
The script regex-matches dialogue lines with the following pattern: `＠.*\n(.|\n)+?\n\[k\]|(？.+?：)`  
`＠.*\n(.|\n)+?\n\[k\]` should match any regular dialogue line, whether standard format, narration (no nametag), or interspersed with function tags.  
`(？.+?：)` should match any player choice, including ones with special interactions, such as in LB6.

### Usage & Output

The `go script` takes a single argument, which is the filepath to any given text file (preferably just the raw script), and will output the number of dialogue lines found in that file.

To count in multiple files at once, the included `powershell script` takes a single argument which is a directory, which should contain another directory for **each chapter's files**. For example with the following file structure:
```
-/Scripts
    -/Avalon
        -300080010.txt
        -300080110.txt
        -300080120.txt
    -/Tunguska
        -9406490210.txt
```
you would call the script with `.\parse-all-script-directories.ps1 Scripts` to parse each respective chapter.  
This script will output a markdown file containing a table, where the first column is the `chapter folder name` in the above structure, and the second column is the total line count within that folder.  

**Note:** the script is recursive and *within a chapter folder* it will go deeper until a file is found, meaning chapters can have multiple folders within their structure (in the case of LB6 and LB7, for example) but the output will combine all those files into one total. 

**The powershell script is not very efficient whatsoever and doesn't have the best output format. Write your own script for actually useful mass output.**
