核武器级选项：filter-branch
有另一个历史改写的选项，如果想要通过脚本的方式改写大量提交的话可以使用它 - 例如，全局修改你的邮箱地址或从每一个提交中移除一个文件。 
这个命令是 filter-branch，它可以改写历史中大量的提交，除非你的项目还没有公开并且其他人没有基于要改写的工作的提交做的工作，你不应当使用它。 
然而，它可以很有用。 你将会学习到几个常用的用途，这样就得到了它适合使用地方的想法。

从每一个提交移除一个文件
这经常发生。 有人粗心地通过 git add . 提交了一个巨大的二进制文件，你想要从所有地方删除它。 可能偶然地提交了一个包括一个密码的文件，然而你想要开源项目。 
filter-branch 是一个可能会用来擦洗整个提交历史的工具。 为了从整个提交历史中移除一个叫做 passwords.txt 的文件，可以使用 --tree-filter 选项给 filter-branch：

git filter-branch --tree-filter 'rm -f passwords.txt' HEAD
Rewrite 6b9b3cf04e7c5686a9cb838c3f36a8cb6a0fc2bd (21/21)
Ref 'refs/heads/master' was rewritten

--tree-filter 选项在检出项目的每一个提交后运行指定的命令然后重新提交结果。 在本例中，你从每一个快照中移除了一个叫作 passwords.txt 的文件，无论它是否存在。 
如果想要移除所有偶然提交的编辑器备份文件，可以运行类似 git filter-branch --tree-filter 'rm -f *~' HEAD 的命令。

最后将可以看到 Git 重写树与提交然后移动分支指针。 通常一个好的想法是在一个测试分支中做这件事，然后当你决定最终结果是真正想要的，可以硬重置 master 分支。 为了让 filter-branch 在所有分支上运行，可以给命令传递 --all 选项。








---
`git filter-branch` 是一个强大但危险的 Git 命令，用于批量修改历史提交。它可以重写提交历史，例如删除敏感文件、修改作者信息或调整提交内容。不过，由于直接修改历史，可能导致协作问题，因此通常建议作为最后的手段使用。


### **基本用法**
```bash
git filter-branch [选项] -- <命令> [范围]
```
- **选项**：如 `--commit-filter`、`--tree-filter`、`--index-filter` 等。
- **命令**：指定要执行的修改操作。
- **范围**：指定要修改的提交范围（如 `HEAD`、`master`、`branch-name` 等）。


### **常见场景**
#### 1. **删除敏感文件**
假设需要从所有提交中彻底删除 `passwords.txt`：
```bash
git filter-branch --force --index-filter \
  'git rm --cached --ignore-unmatch passwords.txt' \
  --prune-empty --tag-name-filter cat -- --all
```
- `--index-filter`：高效修改索引（比 `--tree-filter` 更快）。
- `git rm --cached`：删除文件但保留工作区。
- `--prune-empty`：移除空提交。
- `-- --all`：处理所有分支和标签。

#### 2. **修改作者/提交者信息**
修改所有提交的作者邮箱：
```bash
git filter-branch --env-filter '
  if [ "$GIT_AUTHOR_EMAIL" = "old@example.com" ]; then
    GIT_AUTHOR_NAME="New Name";
    GIT_AUTHOR_EMAIL="new@example.com";
    GIT_COMMITTER_NAME="$GIT_AUTHOR_NAME";
    GIT_COMMITTER_EMAIL="$GIT_AUTHOR_EMAIL";
  fi
  export GIT_AUTHOR_NAME GIT_AUTHOR_EMAIL
  export GIT_COMMITTER_NAME GIT_COMMITTER_EMAIL
' --tag-name-filter cat -- --all
```

#### 3. **修改文件路径**
将所有提交中的 `src/` 目录移动到 `lib/`：
```bash
git filter-branch --index-filter \
  'git ls-files -s | sed "s-\t\"*src/-\t\"lib/-" |
   GIT_INDEX_FILE=$GIT_INDEX_FILE.new git update-index --index-info &&
   mv "$GIT_INDEX_FILE.new" "$GIT_INDEX_FILE"' \
  -- --all
```


### **安全替代方案**
由于 `git filter-branch` 会重写整个历史，现代 Git 推荐使用以下命令：

#### 1. **git filter-repo**（推荐）
需先安装 `git-filter-repo`（独立工具）：
```bash
# 删除敏感文件
git filter-repo --path passwords.txt --invert-paths

# 修改邮箱
git filter-repo --replace-refs delete-no-add \
  --email-callback '
    return "new@example.com" if email == "old@example.com" else email
  '
```
- **优势**：更快、更安全，直接操作底层数据。

#### 2. **git rebase -i**（小规模修改）
仅修改最近几个提交：
```bash
git rebase -i HEAD~3  # 修改最近3个提交
# 在编辑器中选择 `edit`，然后：
git reset HEAD~1      # 撤销上一个提交
# 修改文件后重新提交
git add .
git commit -m "New message"
git rebase --continue
```


### **注意事项**
1. **不要修改已发布的历史**  
   修改已推送的提交会导致与远程仓库的历史冲突，需强制推送（`git push -f`），可能影响协作者。

2. **备份仓库**  
   操作前创建备份：
   ```bash
   git clone --mirror <原仓库> <备份仓库>
   ```

3. **清理残留对象**
   修改后清理未引用的对象并压缩仓库：
   ```bash
   git for-each-ref --format="delete %(refname)" refs/original | git update-ref --stdin
   git reflog expire --expire=now --all
   git gc --prune=now --aggressive
   ```


### **总结**
- **使用场景**：当需要彻底修改历史（如删除敏感数据）且没有其他选择时。
- **推荐顺序**：优先使用 `git rebase -i` 或 `git filter-repo`，最后考虑 `git filter-branch`。
- **协作风险**：修改历史后需通知所有协作者拉取新历史并丢弃旧分支。

掌握 `git filter-branch` 能解决复杂的历史修改需求，但需谨慎操作，避免不必要的风险。