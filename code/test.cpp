#include <iostream>
#include <set>

int main() {

    std::set<std::string> set_mm;
    set_mm.insert("DISTRIBUTE_LOCK:HttpAgentNEV:tokenthread");

    std::cout << set_mm.size() << "\n";

    auto iter = set_mm.find("DISTRIBUTE_LOCK:HttpAgentNEV:tokenthread");
    if(iter != set_mm.end())
    {
        std::cout << "earse \n";
        set_mm.erase(iter);
    }

    std::cout << set_mm.size() << "\n";



    return 0;
}