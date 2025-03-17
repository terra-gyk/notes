- Ubuntu 安装zookeeper c api
  - sudo apt install libzookeeper-mt-dev

- centos 安装 zookeeper c api
  - sudo yum install epel-release
  - sudo yum install zookeeper-devel

- 如果需要源码安装，流程：
  - 下载源码之后，在 zookeeper-jute 目录执行 mvn compile
  - 如果很慢，就取消之后重新执行
  - 然后到 zookeeper-c-client 目录，执行  autoreconf -if
  - 注意，目录下的 cmakelist.txt 直接执行了没用，要用autoreconf -if 生成 configure 文件
  - 这里缺少装啥
  - 然后执行 configure
  - 可能会有找不到 cppunit 的情况，执行如下，或者按提示解决问题
  - export PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:$PKG_CONFIG_PATH
  - 最后 make && make install
