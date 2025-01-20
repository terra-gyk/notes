#include <memory>

class father {};

// 多重继承时，并不共享 public 关键字，每个类前都需要加关键字，否则，默认 private
class son : public father, public std::enable_shared_from_this<son>{};