#include "test.h"

#include <leveldb/db.h>
#include <spdlog/spdlog.h>


#include <iostream>
#include <string>


void node::print(){
  interface::print();
}

void node::test(){
  SPDLOG_INFO("node test func");
}

void interface::print(){
  SPDLOG_INFO("this is a test");
}

int main1(int argc, char** argv){
  if(argc < 3) {
    SPDLOG_INFO("argc < 3");
  }
  
  printf("argc: %d\n",argc);

  return 0;
}

int main() {
    // 定义数据库路径
    std::string dbpath = "/tmp/testdb";

    // 打开数据库
    leveldb::DB* db;
    leveldb::Options options;
    options.create_if_missing = true;
    leveldb::Status status = leveldb::DB::Open(options, dbpath, &db);
    if (!status.ok()) {
        std::cerr << "无法打开数据库: " << status.ToString() << std::endl;
        return 1;
    }

    // 插入数据
    std::string key = "example_key";
    std::string value = "example_value";
    status = db->Put(leveldb::WriteOptions(), key, value);
    if (!status.ok()) {
        std::cerr << "插入数据失败: " << status.ToString() << std::endl;
        delete db;
        return 1;
    }

    // 读取数据
    std::string read_value;
    status = db->Get(leveldb::ReadOptions(), key, &read_value);
    if (status.ok()) {
        std::cout << "读取的数据: " << read_value << std::endl;
    } else if (status.IsNotFound()) {
        std::cout << "未找到键: " << key << std::endl;
    } else {
        std::cerr << "读取数据失败: " << status.ToString() << std::endl;
    }

    // 删除数据
    status = db->Delete(leveldb::WriteOptions(), key);
    if (!status.ok()) {
        std::cerr << "删除数据失败: " << status.ToString() << std::endl;
    }

    // 关闭数据库
    delete db;

    return 0;
}