# Targets



## Asterisc

### v1.1.2

Despite having an handler for the following instructions, the Asteric v1.1.2 target implements them as NO-OP. They can't be considered as supported.
- "07": "FLW/FLD"
- "27": "FSW/FSD"
- "53": "FADD"


Multiple syscalls are implemented. The Asterisc JSON file defines only 3 of them: read, write and exit.


