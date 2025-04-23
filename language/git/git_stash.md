// git 在分支 f1 上做了修改，但是未存储， 现在需要切换到 f2 分支上进行开发，我希望 f1 的被保存，且不提交，同时 f2 无法看到f1 的修改，应该怎么做
// 注意 使用stash 命令之后，修改会被暂存，也就是文件会被恢复

步骤 1：暂存 f1 分支上的修改
在 f1 分支下，使用 git stash 命令将未提交的修改暂存起来。执行该命令时，Git 会把工作目录和暂存区的修改保存到一个栈中，同时将工作目录恢复到上一次提交时的状态。
git stash

如果你想为这次暂存操作添加一个描述信息，方便后续识别，可以使用以下命令：
git stash save "描述信息，例如：f1 分支未完成的功能修改"

步骤 2：切换到 f2 分支
使用 git checkout 命令切换到 f2 分支。由于上一步已经将 f1 分支的修改暂存，切换分支时不会出现冲突，并且 f2 分支也不会受到 f1 分支修改的影响。
git checkout f2

步骤 3：在 f2 分支进行开发
在 f2 分支上进行你需要的开发工作，完成相应的修改和提交操作。例如：

# 进行一些文件的修改
# ...
# 将修改添加到暂存区
git add .
# 提交修改
git commit -m "在 f2 分支上完成的新功能开发"

步骤 4：（可选）返回 f1 分支并恢复修改
当你在 f2 分支的工作完成后，如果需要回到 f1 分支继续之前的工作，可以先切换回 f1 分支，然后使用 git stash pop 或 git stash apply 命令恢复之前暂存的修改。
使用 git stash pop
该命令会将栈顶的暂存内容恢复到工作目录，并且将该暂存记录从栈中移除。

# 切换回 f1 分支
git checkout f1
# 恢复暂存的修改
git stash pop
使用 git stash apply
该命令同样会将栈顶的暂存内容恢复到工作目录，但不会将该暂存记录从栈中移除，这样你之后还可以再次使用该暂存记录。

# 切换回 f1 分支
git checkout f1
# 恢复暂存的修改
git stash apply
或者 git stash apply stash@{0}

步骤 5：（可选）查看暂存栈和清理暂存记录
如果你多次使用了 git stash 命令，可能会有多个暂存记录。可以使用以下命令查看暂存栈中的记录：
git stash list

如果你想删除某个暂存记录，可以使用 git stash drop 命令，例如删除栈顶的记录：
git stash drop stash@{0}

如果你想清空整个暂存栈，可以使用 git stash clear 命令：
git stash clear
通过以上步骤，你可以在不提交 f1 分支修改的情况下安全地切换到 f2 分支进行开发，并且后续还能方便地恢复 f1 分支的修改继续工作。





---
`git stash` 是 Git 中用于临时保存工作目录和暂存区修改的命令，尤其适合在需要切换分支但不想提交当前未完成工作的场景。它将修改保存到一个栈中，之后可以随时恢复。


### **基本用法**
#### 1. **保存当前修改**
```bash
git stash [save "描述信息"]
```
- 会保存工作目录（未暂存）和暂存区（已 `add`）的修改。
- **示例**：
  ```bash
  git stash save "WIP: 临时保存未完成的功能"
  ```

#### 2. **查看 stash 列表**
```bash
git stash list
```
- 输出类似：
  ```
  stash@{0}: WIP on feature: 123456 Add new feature
  stash@{1}: On main: Fix typo
  ```

#### 3. **恢复 stash**
```bash
git stash apply [stash@{n}]  # 恢复但保留 stash
git stash pop [stash@{n}]   # 恢复并删除 stash
```
- **示例**：恢复最近的 stash：
  ```bash
  git stash pop
  ```

#### 4. **删除 stash**
```bash
git stash drop [stash@{n}]  # 删除指定 stash
git stash clear             # 删除所有 stash
```


### **进阶用法**
#### 1. **仅保存工作目录（不保存暂存区）**
```bash
git stash --keep-index
```
- 暂存区的修改会保留，仅保存未暂存的内容。

#### 2. **保存特定文件**
```bash
git stash push -m "只保存部分文件" <文件路径>
```
- **示例**：只保存 `src/module/` 目录的修改：
  ```bash
  git stash push -m "部分保存" src/module/
  ```

#### 3. **恢复到指定分支**
```bash
git stash branch <新分支名> [stash@{n}]
```
- 创建新分支并应用 stash，然后删除该 stash。


### **常见场景**
#### 1. **切换分支前临时保存**
```bash
# 当前在 feature 分支工作，未完成但需切换到 main
git stash
git checkout main
# 在 main 分支完成任务后，回到 feature 恢复工作
git checkout feature
git stash pop
```

#### 2. **解决 stash 冲突**
- 恢复 stash 时可能出现冲突，需手动解决：
  ```bash
  git stash pop  # 出现冲突
  # 手动编辑冲突文件
  git add <冲突文件>
  git stash drop  # 冲突解决后删除 stash
  ```


### **相关命令**
| 命令               | 描述                                  |
|--------------------|---------------------------------------|
| `git stash show`   | 查看 stash 的修改摘要                  |
| `git stash show -p` | 查看 stash 的完整补丁                 |
| `git stash create` | 创建 stash 但不保存到栈（返回 commit ID） |
| `git stash store`  | 将 commit ID 保存到 stash 栈           |


### **注意事项**
1. **stash 不保存未跟踪文件**  
   若需保存未跟踪文件，使用：
   ```bash
   git stash -u  # 包含未跟踪文件
   git stash -a  # 包含所有文件（包括忽略文件）
   ```

2. **stash 是本地操作**  
   stash 不会同步到远程仓库，仅存在于本地。

3. **定期清理 stash**  
   不再需要的 stash 应及时删除，避免栈堆积过多。


### **总结**
`git stash` 是处理临时工作的高效工具，尤其适合：
- 切换分支前保存未完成的修改。
- 尝试新想法但不想创建提交。
- 保存部分文件的修改。

合理使用 `git stash` 可以保持提交历史的整洁，提高开发效率。