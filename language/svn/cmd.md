在 macOS 系统中，SVN（Subversion）是一个常用的版本控制系统。虽然 macOS 默认不再预装 SVN（从 macOS Catalina 开始），但你可以通过 Homebrew 安装它。安装完成后，就可以在终端中使用各种 SVN 命令。

---

### 一、安装 SVN（如未安装）

```bash
# 安装 Homebrew（如未安装）
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 安装 SVN
brew install svn
```

验证安装：

```bash
svn --version
```

---

### 二、常用 SVN 命令

#### 1. 检出（Checkout）代码仓库

```bash
svn checkout https://svn.example.com/repo/project_name
# 或简写
svn co https://svn.example.com/repo/project_name
```

#### 2. 更新本地代码到最新版本

```bash
svn update
# 或
svn up
```

#### 3. 查看状态（哪些文件被修改、新增、删除等）

```bash
svn status
# 或
svn st
```

#### 4. 添加新文件或目录到版本控制

```bash
svn add filename
svn add dir_name --depth=infinity  # 添加整个目录
```

#### 5. 提交更改

```bash
svn commit -m "提交说明信息"
# 或
svn ci -m "提交说明信息"
```

#### 6. 查看日志

```bash
svn log
svn log -l 10          # 查看最近10条日志
svn log -v             # 显示详细信息（包括文件变更）
```

#### 7. 查看差异（本地修改 vs 仓库版本）

```bash
svn diff
svn diff filename      # 查看某个文件的差异
```

#### 8. 删除文件（并加入删除操作到版本控制）

```bash
svn delete filename
# 或
svn del filename
# 或
svn rm filename
```

> 注意：不要直接用 `rm` 删除，否则 SVN 不知道你删除了文件。

#### 9. 恢复误删或误改的文件

```bash
svn revert filename
svn revert . --depth=infinity   # 恢复当前目录及子目录所有更改（慎用！）
```

#### 10. 查看当前工作副本的信息

```bash
svn info
```

#### 11. 切换分支或标签（switch）

```bash
svn switch https://svn.example.com/repo/branches/feature-branch
```

#### 12. 解决冲突后标记为已解决

```bash
svn resolve --accept=working filename
```

---

### 三、常见问题

#### Q：提示 `svn: command not found`？

A：说明未安装 SVN，请按上面方法用 Homebrew 安装。

#### Q：如何忽略某些文件（如 .DS_Store、build/ 目录）？

A：设置 `svn:ignore` 属性：

```bash
svn propset svn:ignore ".DS_Store
build/
*.log" .
svn commit -m "设置忽略文件"
```

或对单个文件：

```bash
svn propset svn:ignore ".DS_Store" .
```

---


---

## ✅ 一、SVN 加锁常用命令（命令行）

### 1. **给文件加锁**

```bash
svn lock 文件名
```

示例：
```bash
svn lock docs/design.psd
svn lock report.xlsx
```

> 🔐 加锁后，只有你（或管理员）能提交对该文件的修改，其他人无法提交，直到你解锁。

---

### 2. **查看文件是否被锁定 / 谁锁了**

```bash
svn info 文件名
```

或查看整个工作副本的锁状态：
```bash
svn status --show-updates
# 或简写
svn st -u
```

输出中如果有 `K` 或 `O` 表示被锁：
- `K`：你拥有这个锁（Locked by you）
- `O`：别人锁了（Locked by others）

也可以用：
```bash
svn info 文件名 | grep "Lock"
```

---

### 3. **提交修改（必须先加锁）**

对已加锁的文件修改后，正常提交即可：
```bash
svn commit -m "更新设计稿" docs/design.psd
```

> ⚠️ 如果你没加锁就修改了别人锁住的文件，提交时会失败。

---

### 4. **解锁文件**

```bash
svn unlock 文件名
```

示例：
```bash
svn unlock docs/design.psd
```

> ✅ 提交后建议及时解锁，以便他人可以修改。

---

### 5. **强制解锁（管理员或文件所有者）**

如果别人忘记解锁，你有权限的话可以强制解锁：

```bash
svn unlock --force 文件名
```

> 🔒 注意：`--force` 只能由 **锁的拥有者** 或 **有 svn:owner 权限的管理员** 使用。

---

## ✅ 二、加锁注意事项

### 1. **哪些文件需要加锁？**
- 二进制文件（图片、Office 文档、音视频、压缩包等）
- 无法文本合并的文件

> ✅ 源代码（`.c`, `.java`, `.py` 等）**通常不需要加锁**，SVN 支持自动合并。

### 2. **加锁是“建议性”的**
SVN 的锁机制是 **协作式（advisory）**，不是强制性的：
- 别人仍可 `svn update` 获取你修改前的版本
- 但 **无法提交** 对该文件的修改（除非强制抢锁）

### 3. **服务器必须支持锁**
确保你的 SVN 服务器配置允许锁操作（大多数都默认支持）。

---

## ✅ 三、图形化工具中的加锁（如 macOS 上的工具）

如果你用的是 GUI 工具，操作更直观：

| 工具 | 加锁方式 |
|------|--------|
| **SnailSVN**（macOS Finder 集成） | 右键文件 → SVN → Lock |
| **Cornerstone** | 右键 → Lock |
| **TortoiseSVN**（Windows） | 右键 → TortoiseSVN → Lock |

---

## ✅ 四、常见问题

### Q：加锁后别人还能更新（update）吗？
✅ **可以**。别人仍能 `svn update` 获取最新版本，但不能提交修改。

### Q：提交后锁会自动释放吗？
❌ **不会！** SVN 默认 **提交后保留锁**（防止你继续修改）。  
如果你希望提交后自动解锁，加参数：
```bash
svn commit --no-unlock -m "提交" file  # 默认行为（保留锁）
svn commit --no-lock -m "提交" file    # ❌ 不存在这个参数
```

> 实际上，SVN **没有自动解锁的提交参数**。你必须手动 `svn unlock`。

但你可以提交并解锁分两步：
```bash
svn commit -m "更新"
svn unlock 文件名
```

或者写个脚本一键完成。

---

## ✅ 五、查看仓库中所有锁（管理员用）

```bash
svnadmin lslocks /path/to/svn/repo
```

或通过 HTTP(S) 查看（如果服务器支持）：
```bash
svn info --show-item lock-owner URL
```

---

