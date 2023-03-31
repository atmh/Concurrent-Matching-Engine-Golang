#include <fstream>
#include <string>
#include <vector>



int main(int argc, char** argv) {
    std::vector<std::string> ticker_symbols;
    ticker_symbols.push_back("AAPL");
    ticker_symbols.push_back("MSFT");
    // ticker_symbols.push_back("GOOG");
    // ticker_symbols.push_back("AMZN");
    // ticker_symbols.push_back("FB");

    std::string COMMENT = argv[1];
    std::string FILE_NAME = argv[2];
    
    // Open a file for writing
    std::ofstream out_file(FILE_NAME);

    int num_of_threads = 40;
    // comment line
    out_file << "#" << COMMENT << std::endl;
    // first line
    out_file << num_of_threads << std::endl;
    out_file << 0 <<  "-" << num_of_threads - 1 << " o" << std::endl;
    int order_id = 1;
    for (int i = 0; i < num_of_threads; i++) {
    //    for (std::string symbol : ticker_symbols) {
           srand(time(0));
        out_file << i << " B " << order_id++ << " " << ticker_symbols[0] << " " << 2705 << " " << 30 << std::endl;
        out_file << i << " S " << order_id++ << " " << ticker_symbols[1] << " " << 3260 << " " << 1 << std::endl;
        out_file << i << " C " << order_id - 2 << std::endl;
        out_file << i << " S " << order_id++ << " " << ticker_symbols[0] << " " << 2701 << " " << 25 << std::endl;
        out_file << i << " B " << order_id++ << " " << ticker_symbols[1] << " " << 3290 << " " << 10 << std::endl;
        out_file << i << " C " << order_id - 3 << std::endl;
        
    //    }
    }


    //last 2 lines
    out_file << "." << std::endl;
    out_file << 0 <<  "-" << num_of_threads - 1 << " x" << std::endl;
    // Close the file
    out_file.close();

    return 0;
}