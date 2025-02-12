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