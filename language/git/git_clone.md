(1) 克隆主项目及子模块
git clone --recurse-submodules https://github.com/example/main-repo.git

(2) 如果已经克隆了主项目
cd main-repo
git submodule update --init --recursive