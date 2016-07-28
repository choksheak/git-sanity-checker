# git-sanity-checker

## Introduction

Git sanity checker is a small utility that can be embedded into the Git workflow seamlessly to perform quick checks of files before checking in code. Its main purpose is to catch common issues of code that should never be checked in. However, it should not be used as a blocker for commits because there are too many exceptions that people would like to accept.

## Key Points
1. The rules are not to be enforced to stop commits because they are somewhat granular.
2. Users can choose to integrate it as part of their regular Git workflow but they don't have to.
3. This tool can be run standalone without Git.

## List of Git Commands Used

(REMOVED) Get list of files that are modified and cached in Git.

    git diff --name-only --cached

Get list of files that are modified but not cached in Git.

    git diff --name-only

Get list of files that are new to the Git repo and not checked in yet.

    git ls-files --others --exclude-standard

## User Workflow
1. Run "git check" or execute git-sanity-checker.exe directly.
2. When no arguments are given, use the git command to get list of files to check.
3. When one or more arguments are given, use the arguments as file names or directory names to check for.
4. git-sanity-checker will try to read a config file from the same directory as where the executable is.
5. If the config file does not exist, then exit with an error message.
6. If the config file is empty or has any unrecognized rules, then exit with an error message.
7. The config file will list the rules that we want to run to check the files.
8. For each file, for each rule, we will say whether the file passes or fails, with one or more lines of error messages.

## List of Check Rules
Please see the config file [git-sanity-checker.cfg](https://github.com/choksheak/git-sanity-checker/blob/master/git-sanity-checker.cfg) for details as this list is dynamic.

## Setup / Installation
1. Copy git-sanity-checker.exe and git-sanity-checker.cfg to some folder on your computer.
2. Add this folder to the search path (PATH environment variable).
3. Configure git to run this script. Under .gitconfig "[alias]" section, add the following lines:

    check = !git-sanity-checker
    
    c = !git-sanity-checker
    
    s = "!f() { git status; git-sanity-checker; }; f"

4. Anywhere within any git repo, run "git check" or "git c" or "git s"

## Checking Any File
If you want to check any file, just pass the filenames or directory names as the command line argument:

    git-sanity-checker BadFile.cs
    git-sanity-checker .
    git-sanity-checker *

