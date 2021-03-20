# grab
A friendly cross-platform package manager.

## Todos

* Ensure a well-known "grab" directory is present on the user's machine, e.g:
    * darwin:  `/usr/local/bin/grab` + `/pkg` for packages
    * windows: `%ALLUSERSPROFILE%\grab` + `\pkg` for packages
    * linux:   `/usr/local/bin/grab` + `/pkg` for packages

* Ensure the well-known "grab" directory is in the user's path.

* Create a lock file and put it in the "grab" directory:

```
{
    "codingconcepts": {
        "pa55": "1.0.1"
    }
}
```

* Process multiple pages if a version was specified until that version is found.