# gclone

**Note: Very early development stage.**

A tool for keeping track of *git clones*.
Register a new clone or add an already existing one to the *glcone* storage.
Now it is possible to print the current status for all registered repositories.
It also allows to update all repositories with a `git fetch` by a single
command.

## Installation

```console
go get github.com/towoe/gclone
```

## Usage

### Register a repository

Cloning a new repository adds it directly to the register.

```console
$ gclone clone https://github.com/user/repo [folder]
```

Add a repository which is already locally available.

```console
$ gclone add path/to/folder
```

### Print the status

Invoking `gclone` without any arguments will print the status for each
registered repository. The command `status` can also be used.

```console
$ gclone
$ gclone status
```

### Updating all repositories

Adding the argument `fetch` will perform a normal `git fetch` for each
repository. No `pull` is performed as this includes a merge, which might lead
to an unintended state.

```console
$ glcone fetch
```

### Index file

The default storage location for the index file is
`$XDG_DATA_HOME/gclone/register.json`.

The argument `-i, --index` can be supplied to use a different file.
It **must** be given prior to additional arguments (path or URL).
This can be useful in order to group several repositories together.
