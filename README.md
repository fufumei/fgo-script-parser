A simple script to get the *actual* dialogue line count for any FGO story script.

### Usage & Output

The script takes a filepath as its only argument.  
If the given path is a file (FGO story scripts are by default in `.txt` format), it will simply count the lines and output the result to the command line.  
If the given path is a folder, the script will traverse every underlying path until it finds a file to open. At that point, it will count the total lines in the current folder, write the result to a file, then repeat for any remaining folders (**note that the script will likely not work if you have files and folders mixed on the same level**).

The output file is a tab-separated file named `script-length.csv`, outputted on the same level as the script.  
The output format is `folder name    total lines`. 

Note that, if for example all LB6 scripts are separated into different folders for the respective "parts", the output will reflect this.   

### Regex matching
The script regex-matches dialogue lines with the following pattern: `(＠([A-Z][：:])?(.*)\n)(.*?\n(?:.*?\n)?)?(.*?)\n\[k\]|(？.+?：)`  

`(＠([A-Z][：:])?(.*)\n)(.*?\n(?:.*?\n)?)?(.*?)\n\[k\]` should match any regular dialogue line, whether standard format, narration (no nametag), or interspersed with function tags.  

`(？.+?：)` should match any player choice, including ones with special interactions, such as in LB6.

### Special cases

In one of the scripts for OC2 (I forget which one specifically), there exists a bug wherein one of the dialogue lines is missing a `＠` character at the start. This bug makes is technically apparent in the game itself, and as a result the script can't match this line. Thus, when counting OC2 scripts an additional match will automatically be counted by default.  
If this is ever fixed in-game, I'll hopefully remember to remove this from the script.

For the Ordeal Call Prologue, the two scripts `0400010110` and `0400019910` are actually the exact same script (one is just a redirect to the other, for some reason). The counting script doesn't take this into account, so you either need to divide the total ouput for this by 2, or simply remove one of the files from wherever you are counting (the total lines for this should be 95, for reference).
