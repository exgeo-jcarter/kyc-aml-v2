#include <string>
#include "KycAmlDoubleMetaphone.hpp"
#include <iostream>

int main(int argc, char** argv) {

	/*
	if (argc != 2) {
		std::cout << "usage: kyc-aml-doublemetaphone \"phrase to be encoded\"" << std::endl;
		return 1;
	}

	vector<string> codes;
	DoubleMetaphone(argv[1], &codes);

	std::cout << argv[1] << " : ";

	bool what = false;
	for (auto& val: codes) {

		std::cout << val << " ";
		what = !what;

		if (!what) {
			std::cout << std::endl;
		}
	}
	*/
	std::string conf_file = "../config.json";

	if (argc >= 2) {
		conf_file = argv[1];
	}

	exGeoKycAml::KycAmlDoubleMetaphone server(conf_file);
	bool listen_res = server.Listen();
	if (!listen_res) {
		std::cerr << "Error: Listening failed" << std::endl;
		return 1;
	}

	return 0;
}
