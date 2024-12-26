#include <spdlog/spdlog.h>
#include <gtest/gtest.h>

#include <future>

int sum(int num1, int num2){
  return num1 + num2;
}

TEST(sum_test,sum){
  ASSERT_EQ(sum(1,2),3);
}

int main(){
  testing::InitGoogleTest();
  return RUN_ALL_TESTS();
  // return 0;
}
 