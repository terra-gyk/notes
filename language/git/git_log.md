// 以树形的方式展示git提交记录，包含提交人提交时间，提交日志，最多20条
git log --graph --decorate --oneline --pretty=format:"%h %d %an: %s %ad" --max-count=20 

// 显示每次提交的内容差异
git log -p -2 