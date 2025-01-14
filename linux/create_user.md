## 创建用户 -s 是默认的终端，不加，无法打开shell。 -m 是创建用户目录
sudo useradd -r -m -s /bin/bash terra
## 修改密码
sudo passwd terra 
## 这个 l 很重要
## 表示登录 shell。这会使得切换后的用户拥有与该用户登录时相同的 shell 环境
## 包括环境变量、用户主目录、登录 shell 等。
su -l terra 

## 更改拥有者
chown -R terra ./work_space
## 更改拥有组
chgrp -R terra ./work_space