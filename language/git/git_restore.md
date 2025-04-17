`git restore` 是 Git 2.23 版本引入的一个新命令，用于恢复工作区文件和暂存区的内容。它的设计目的是将一些原本复杂的操作拆分成更清晰、更易于理解的命令，从而替代了部分 `git checkout` 和 `git reset` 的功能。以下是关于 `git restore` 的详细介绍：

### 恢复工作区文件
如果你对工作区的文件进行了修改，但还没有将这些修改添加到暂存区，你可以使用 `git restore` 命令将文件恢复到上一次提交时的状态。
```bash
git restore <file>
```
这里的 `<file>` 是你想要恢复的文件的路径。如果你想恢复多个文件，可以在命令后面列出这些文件的路径，用空格分隔。例如：
```bash
git restore file1.txt file2.txt
```
如果你想恢复当前目录下的所有文件，可以使用 `.` 来表示当前目录：
```bash
git restore .
```

### 恢复暂存区文件
如果你已经将文件添加到了暂存区，但还没有提交，你可以使用 `--staged` 选项将文件从暂存区移除，恢复到工作区未修改的状态：
```bash
git restore --staged <file>
```
例如，如果你想将 `file1.txt` 从暂存区移除，可以运行：
```bash
git restore --staged file1.txt
```

### 从指定提交恢复文件
你还可以从指定的提交中恢复文件。使用 `-s` 或 `--source` 选项指定提交的引用（如提交哈希、分支名、标签名等）：
```bash
git restore -s <commit> <file>
```
例如，如果你想从 `HEAD~1`（即上一次提交）中恢复 `file1.txt`，可以运行：
```bash
git restore -s HEAD~1 file1.txt
```

### 结合 `--worktree` 和 `--staged`
你可以同时使用 `--worktree` 和 `--staged` 选项，将文件从指定提交恢复到暂存区和工作区：
```bash
git restore --worktree --staged -s <commit> <file>
```
例如，将 `file1.txt` 从 `HEAD~1` 恢复到暂存区和工作区：
```bash
git restore --worktree --staged -s HEAD~1 file1.txt
```

### 总结
`git restore` 命令为恢复工作区和暂存区的文件提供了清晰、简洁的方式。通过不同的选项，你可以灵活地控制文件的恢复操作，避免了使用 `git checkout` 和 `git reset` 时可能出现的混淆。