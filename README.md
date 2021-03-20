# grab
A friendly cross-platform package manager.

## Todos

* Single link installs:
    * darwin: `/bin/bash -c "$(curl -fsSL https://github.com/codingconcepts/grab/install/darwin.sh)"`
    * linux: `/bin/bash -c "$(curl -fsSL https://github.com/codingconcepts/grab/install/linux.sh)"`
    * windows: `TODO`

* Ensure a well-known "grab" directory is present on the user's machine, e.g:
    * darwin:  `/usr/local/bin/grab` + `/pkg` for packages
    * windows: `%ALLUSERSPROFILE%\grab` + `\pkg` for packages
    * linux:   `/usr/local/bin/grab` + `/pkg` for packages

* Ensure the well-known "grab" directory is in the user's path.


* Process multiple pages if a version was specified until that version is found.