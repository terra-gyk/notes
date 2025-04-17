`git checkout` 是 Git 里一个极为常用且功能丰富的命令，能在不同分支间切换、恢复文件、查看旧版本内容等。下面为你详细介绍其主要用法：

### 分支切换
借助 `git checkout` 可在不同分支间切换。命令格式如下：
```bash
git checkout <branch_name>
```
这里的 `<branch_name>` 指的是你要切换到的目标分支名。例如，要从当前分支切换到 `feature` 分支，可执行：
```bash
git checkout feature
```
要是目标分支不存在，你可以使用 `-b` 选项创建并切换到该分支：
```bash
git checkout -b <new_branch_name>
```
例如，创建并切换到名为 `new-feature` 的新分支：
```bash
git checkout -b new-feature
```

### 恢复文件
若要把工作区的文件恢复到上一次提交时的状态，也能使用 `git checkout` 命令。命令格式如下：
```bash
git checkout -- <file_path>
```
这里的 `<file_path>` 是你想要恢复的文件路径。例如，恢复 `src/main.py` 文件：
```bash
git checkout -- src/main.py
```
要注意，此操作会让工作区的修改丢失，且无法恢复。

### 查看旧版本内容
你还能使用 `git checkout` 查看某个提交版本下文件的内容。命令格式如下：
```bash
git checkout <commit_hash> -- <file_path>
```
这里的 `<commit_hash>` 是提交的哈希值，`<file_path>` 是文件路径。例如，查看 `abc123` 这个提交版本下 `README.md` 文件的内容：
```bash
git checkout abc123 -- README.md
```
执行该命令后，工作区的 `README.md` 文件会变成 `abc123` 提交时的内容。若要回到最新版本，可切换到当前分支：
```bash
git checkout <current_branch>
```

### 切换到远程分支
当要切换到远程分支时，可使用以下命令：
```bash
git checkout -b <local_branch_name> origin/<remote_branch_name>
```
例如，从远程仓库的 `origin/feature` 分支创建并切换到本地的 `feature` 分支：
```bash
git checkout -b feature origin/feature
```

### 总结
`git checkout` 命令用途广泛，不过从 Git 2.23 版本开始，部分功能已被 `git switch`（用于分支切换）和 `git restore`（用于文件恢复）替代。但由于兼容性和习惯问题，`git checkout` 仍被广泛使用。 