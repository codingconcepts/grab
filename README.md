# grab
A friendly cross-platform package manager.

## Todos

* Create grab_lock to contain the change to make (if it exists, there's another process running).

The lock file is the same as the state file but just for the current operation. Once the operation is finished, the file is deleted. That way, if anything fails, we can rollback.

```
{
    "codingconcepts": {
        "pa55": "1.0.1"
    }
}
```

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