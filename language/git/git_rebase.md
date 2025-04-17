// git 从 master checkout出分支 f1， f1上有多次提交，怎么把 f1 的多次提交合并成一个提交合并到master，最终在master上只有一个提交记录
// 注意不要push之后再进行rebase，否则其余成员都需要使用 git pull --rebase 来拉取分支
// 如果单独的开发分支，只有自己在用，那么push之后需要合并，可以 rebase -i 之后 push -f，强制推送再合并到主分支

使用 git rebase -i 结合合并操作
先在 f1 分支上使用交互式变基将多次提交合并成一个，再将 f1 分支合并到 master 分支。
具体步骤
 
切换到 f1 分支
git checkout f1

进行交互式变基
使用 git rebase -i 命令将 f1 分支上的多次提交合并成一个。你需要指定要合并的提交范围，例如，如果要合并最近的 3 次提交，可以执行：
git rebase -i HEAD~3

执行该命令后，Git 会打开一个文本编辑器，列出你指定范围内的提交，类似如下：
plaintext
pick 123abc Some commit message 1
pick 456def Some commit message 2
pick 789ghi Some commit message 3

将除了第一个 pick 之外的其他 pick 改为 squash 或者 s，表示将这些提交合并到前一个提交中，修改后如下：
plaintext
pick 123abc Some commit message 1
s 456def Some commit message 2
s 789ghi Some commit message 3
保存并关闭文件后，Git 会再次打开编辑器让你编辑合并后的提交信息，编辑完成后保存并关闭。

切换到 master 分支
git checkout master
拉取 master 分支的最新代码

git pull origin master
合并 f1 分支到 master 分支

git merge f1
如果合并过程中没有冲突，Git 会自动完成合并。若有冲突，需要手动解决冲突，然后使用 git add 和 git commit 完成合并。

6. 将合并后的 master 分支推送到远程仓库
git push origin master

