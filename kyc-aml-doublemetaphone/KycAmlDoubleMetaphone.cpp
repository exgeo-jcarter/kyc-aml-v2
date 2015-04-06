#include <iostream>
#include <vector>
#include <string>
#include <string.h>
#include <stdio.h>
#include <ctype.h>
#include <stdlib.h>
#include <string.h>
#include <stdarg.h>
#include <assert.h>
#include "double-metaphone/double_metaphone.h"
#include <boost/asio.hpp>
#include <memory>
#include "KycAmlDoubleMetaphone.hpp"
#include <iostream>
#include <fstream>
#include "json/json.h"
#include <cstring>
#include <locale>

using boost::asio::ip::tcp;

namespace exGeoKycAml {

Session::Session(KycAmlDoubleMetaphone* parent, tcp::socket socket)
: parent(parent), socket(std::move(socket))
{
}

void Session::Start() {
	this->doRead();
}

void Session::doRead() {

	try {

		auto self(shared_from_this());
		//this->socket.async_read_some(boost::asio::buffer(this->data, this->max_length),
		boost::asio::async_read_until(this->socket, this->streambuf, '\n',
			[this, self](boost::system::error_code ec, std::size_t length) {
				if (!ec) {

					std::ostringstream ss;
					ss << &this->streambuf;
					std::string data = ss.str();

					Json::Value msg;
					Json::Reader reader;

					bool parsed = reader.parse(data, msg, false);
					if (!parsed) {
						std::cerr << "Error: " << reader.getFormatedErrorMessages() << "\n";
						return;
					}

					std::string res;

					if (msg["action"].asString() == "train_sdn") {

						// TODO: check if already trained.

						if (parent->SdnListMapNames.size() != 0) {
							res = "{\"result\": \"Already trained.\"}\n";
							this->doWrite(res);
							return;
						}

						bool train_res = parent->TrainSdn(msg["value"].asString());
						if (!train_res) {
							res = "{\"result\": \"Training SDN list failed.\"}\n";
							this->doWrite(res);
							return;
						}

						res = "{\"result\": \"Training SDN list complete.\"}\n";
						this->doWrite(res);
					}

					else {

						Json::Value res;

						std::string doublemetaphone_query;
						std::locale loc;

						for (auto& ch: msg["value"].asString()) {
							doublemetaphone_query += std::tolower(ch, loc);
						}

						std::vector<std::string> doublemetaphone_encoded_query;
						DoubleMetaphone(doublemetaphone_query, &doublemetaphone_encoded_query);

						std::string doublemetaphone_encoded_query1;
						for (auto& ch: doublemetaphone_encoded_query[0]) {
							doublemetaphone_encoded_query1 += std::tolower(ch, loc);
						}

						std::string doublemetaphone_encoded_query2;
						for (auto& ch: doublemetaphone_encoded_query[1]) {
							doublemetaphone_encoded_query2 += std::tolower(ch, loc);
						}

						res["query"] = doublemetaphone_query;
						res["encoded_query1"] = doublemetaphone_encoded_query1;

						if (doublemetaphone_encoded_query1 != doublemetaphone_encoded_query2) {
							res["encoded_query2"] = doublemetaphone_encoded_query2;
						}

						if (msg["action"].asString() == "query_name") {

							// If encoded name is found
							if (parent->SdnListMapNames.find(doublemetaphone_encoded_query1) != parent->SdnListMapNames.end()) {
								res["name_result1"] = Json::arrayValue;
								res["name_result1"].append(parent->SdnListMapNames[doublemetaphone_encoded_query1]);
							}

							if (doublemetaphone_encoded_query1 != doublemetaphone_encoded_query2) {
								if (parent->SdnListMapNames.find(doublemetaphone_encoded_query2) != parent->SdnListMapNames.end()) {
									res["name_result2"] = Json::arrayValue;
									res["name_result2"].append(parent->SdnListMapNames[doublemetaphone_encoded_query2]);
								}
							}

							if (parent->SdnListMapRevNames.find(doublemetaphone_encoded_query1) != parent->SdnListMapRevNames.end()) {
								res["revname_result1"] = Json::arrayValue;
								res["revname_result1"].append(parent->SdnListMapRevNames[doublemetaphone_encoded_query1]);
							}

							if (doublemetaphone_encoded_query1 != doublemetaphone_encoded_query2) {
								if (parent->SdnListMapRevNames.find(doublemetaphone_encoded_query2) != parent->SdnListMapRevNames.end()) {
									res["revname_result2"] = Json::arrayValue;
									res["revname_result2"].append(parent->SdnListMapRevNames[doublemetaphone_encoded_query2]);
								}
							}

							if (parent->SdnListMapAkas.find(doublemetaphone_encoded_query1) != parent->SdnListMapAkas.end()) {
								res["aka_result1"] = Json::arrayValue;
								res["aka_result1"].append(parent->SdnListMapAkas[doublemetaphone_encoded_query1]);
							}

							if (doublemetaphone_encoded_query1 != doublemetaphone_encoded_query2) {
								if (parent->SdnListMapAkas.find(doublemetaphone_encoded_query2) != parent->SdnListMapAkas.end()) {
									res["aka_result2"] = Json::arrayValue;
									res["aka_result2"].append(parent->SdnListMapAkas[doublemetaphone_encoded_query2]);
								}
							}

							if (parent->SdnListMapRevAkas.find(doublemetaphone_encoded_query1) != parent->SdnListMapRevAkas.end()) {
								res["revaka_result1"] = Json::arrayValue;
								res["revaka_result1"].append(parent->SdnListMapRevAkas[doublemetaphone_encoded_query1]);
							}

							if (doublemetaphone_encoded_query1 != doublemetaphone_encoded_query2) {
								if (parent->SdnListMapRevAkas.find(doublemetaphone_encoded_query2) != parent->SdnListMapRevAkas.end()) {
									res["revaka_result2"] = Json::arrayValue;
									res["revaka_result2"].append(parent->SdnListMapRevAkas[doublemetaphone_encoded_query2]);
								}
							}
						}

						if (msg["action"].asString() == "query_address") {

							if (parent->SdnListMapAddresses.find(doublemetaphone_encoded_query1) != parent->SdnListMapAddresses.end()) {
								res["address_result1"] = Json::arrayValue;
								res["address_result1"].append(parent->SdnListMapAddresses[doublemetaphone_encoded_query1]);
							}

							if (doublemetaphone_encoded_query1 != doublemetaphone_encoded_query2) {
								if (parent->SdnListMapAddresses.find(doublemetaphone_encoded_query2) != parent->SdnListMapAddresses.end()) {
									res["address_result2"] = Json::arrayValue;
									res["address_result2"].append(parent->SdnListMapAddresses[doublemetaphone_encoded_query2]);
								}
							}

							if (parent->SdnListMapPostalCodes.find(doublemetaphone_encoded_query1) != parent->SdnListMapPostalCodes.end()) {
								res["postal_code_result1"] = Json::arrayValue;
								res["postal_code_result1"].append(parent->SdnListMapPostalCodes[doublemetaphone_encoded_query1]);
							}

							if (doublemetaphone_encoded_query1 != doublemetaphone_encoded_query2) {
								if (parent->SdnListMapPostalCodes.find(doublemetaphone_encoded_query2) != parent->SdnListMapPostalCodes.end()) {
									res["postal_code_result2"] = Json::arrayValue;
									res["postal_code_result2"].append(parent->SdnListMapPostalCodes[doublemetaphone_encoded_query2]);
								}
							}
						}

						Json::StreamWriterBuilder builder;
						builder.settings_["indentation"] = "";
						std::string res_str = Json::writeString(builder, res);

						this->doWrite(res_str+"\n");
					}
				}
			});
	}
	catch (std::exception& e) {
		std::cerr << "Exception: " << e.what() << "\n";
		return;
	}

}

void Session::doWrite(std::string msg) {
	auto self(shared_from_this());
	boost::asio::async_write(this->socket, boost::asio::buffer(msg.c_str(), msg.length()),
		[this, self](boost::system::error_code ec, std::size_t /* length */) {
			if (!ec) {
				this->doRead();
			}
		});
}

Server::Server(KycAmlDoubleMetaphone* parent, boost::asio::io_service& io_service, short port)
: parent(parent), acceptor(io_service, tcp::endpoint(tcp::v4(), port)),
  socket(io_service)
{
	this->doAccept();
}

void Server::doAccept() {
	this->acceptor.async_accept(this->socket,
		[this](boost::system::error_code ec) {
			if (!ec) {
				std::make_shared<Session>(this->parent, std::move(this->socket))->Start();
			}

			this->doAccept();
		});
}

KycAmlDoubleMetaphone::KycAmlDoubleMetaphone(const std::string& conf_filename)
: Conf(new KycAmlDoubleMetaphoneConfS())
{
	bool conf_loaded = this->LoadConf(conf_filename);
	if (!conf_loaded) {
		return;
	}
}

KycAmlDoubleMetaphone::~KycAmlDoubleMetaphone() {
	delete this->Conf;
}

bool KycAmlDoubleMetaphone::LoadConf(const std::string& filename) {

	try {

		std::ifstream filestream(filename);
		if (!filestream) {
			std::cerr << "Error: File couldn't be opened: " << filename << "\n";
			return false;
		}

		Json::Value conf_json;
		filestream >> conf_json;

		this->Conf->Host = conf_json["host"].asString();
		this->Conf->Port = conf_json["port"].asString();
		this->Conf->Protocol = conf_json["protocol"].asString();

		std::cout << "DoubleMetaphone server config file loaded." << std::endl;
	}
	catch (std::exception& e) {
		std::cerr << "Exception: " << e.what() << "\n";
		return false;
	}

	return true;
}

bool KycAmlDoubleMetaphone::Listen() {

	try {
		boost::asio::io_service io_service;
		Server s(this, io_service, std::stoi(this->Conf->Port));
		std::cout << "DoubleMetaphone server listening." << std::endl;
		io_service.run();
	}

	catch (std::exception& e) {
		std::cerr << "Exception: " << e.what() << "\n";
		return false;
	}

	return true;
}

bool KycAmlDoubleMetaphone::TrainSdn(const std::string& sdn_list) {

	try {
		std::cout << "Parsing JSON SDN list." << std::endl;

		Json::Value sdn_list_json;
		Json::Reader reader;

		bool parsed = reader.parse(sdn_list, sdn_list_json, false);
		if (!parsed) {
			std::cerr << "Error: " << reader.getFormatedErrorMessages() << "\n";
			return false;
		}

		std::cout << "Training DoubleMetaphone SDN list." << std::endl;

		// Add names to map.
		for (auto& sdn_entry: sdn_list_json["sdn_entry"]) {

			std::string name = sdn_entry["first_name"].asString()+" "+sdn_entry["last_name"].asString();
			std::vector<std::string> name_codes;
			DoubleMetaphone(name, &name_codes);

			std::locale loc;

			std::string name_lower1;
			for (auto& ch: name_codes[0]) {
				name_lower1 += std::tolower(ch, loc);
			}
			std::string name_lower2;
			for (auto& ch: name_codes[1]) {
				name_lower2 += std::tolower(ch, loc);
			}

			this->SdnListMapNames.insert(std::make_pair(name_lower1, name));
			this->SdnListMapNames.insert(std::make_pair(name_lower2, name));

			std::string revname = sdn_entry["last_name"].asString()+" "+sdn_entry["first_name"].asString();
			std::vector<std::string> revname_codes;
			DoubleMetaphone(revname, &revname_codes);

			std::string revname_lower1;
			for (auto& ch: revname_codes[0]) {
				revname_lower1 += std::tolower(ch, loc);
			}
			std::string revname_lower2;
			for (auto& ch: revname_codes[1]) {
				revname_lower2 += std::tolower(ch, loc);
			}

			this->SdnListMapRevNames.insert(std::make_pair(revname_lower1, revname));
			this->SdnListMapRevNames.insert(std::make_pair(revname_lower2, revname));

			for (auto& akas: sdn_entry["aka_list"]["aka"]) {

				std::string aka = akas["first_name"].asString()+" "+akas["last_name"].asString();
				std::vector<std::string> aka_codes;
				DoubleMetaphone(aka, &aka_codes);

				std::string aka_lower1;
				for (auto& ch: aka_codes[0]) {
					aka_lower1 += std::tolower(ch, loc);
				}
				std::string aka_lower2;
				for (auto& ch: aka_codes[1]) {
					aka_lower2 += std::tolower(ch, loc);
				}

				this->SdnListMapAkas.insert(std::make_pair(aka_lower1, aka));
				this->SdnListMapAkas.insert(std::make_pair(aka_lower2, aka));

				std::string revaka = akas["last_name"].asString()+" "+akas["first_name"].asString();
				std::vector<std::string> revaka_codes;
				DoubleMetaphone(revaka, &revaka_codes);

				std::string revaka_lower1;
				for (auto& ch: revaka_codes[0]) {
					revaka_lower1 += std::tolower(ch, loc);
				}
				std::string revaka_lower2;
				for (auto& ch: revaka_codes[1]) {
					revaka_lower2 += std::tolower(ch, loc);
				}

				this->SdnListMapRevAkas.insert(std::make_pair(revaka_lower1, revaka));
				this->SdnListMapRevAkas.insert(std::make_pair(revaka_lower2, revaka));
			}

			for (auto& address: sdn_entry["address_list"]["addresses"]) {

				std::string address1 = address["address1"].asString();
				std::vector<std::string> address_codes;
				DoubleMetaphone(address1, &address_codes);

				std::string address_lower1;
				for (auto& ch: address_codes[0]) {
					address_lower1 += std::tolower(ch, loc);
				}
				std::string address_lower2;
				for (auto& ch: address_codes[1]) {
					address_lower2 += std::tolower(ch, loc);
				}

				this->SdnListMapAddresses.insert(std::make_pair(address_lower1, address1));
				this->SdnListMapAddresses.insert(std::make_pair(address_lower2, address1));

				std::string postal_code = address["postal_code"].asString();
				std::vector<std::string> postal_code_codes;
				DoubleMetaphone(postal_code, &postal_code_codes);

				std::string postal_code_lower1;
				for (auto& ch: postal_code_codes[0]) {
					postal_code_lower1 += std::tolower(ch, loc);
				}
				std::string postal_code_lower2;
				for (auto& ch: postal_code_codes[1]) {
					postal_code_lower2 += std::tolower(ch, loc);
				}

				this->SdnListMapPostalCodes.insert(std::make_pair(postal_code_lower1, postal_code));
				this->SdnListMapPostalCodes.insert(std::make_pair(postal_code_lower2, postal_code));
			}
		}

		std::cout << "DoubleMetaphone search training complete. You can perform queries now." << std::endl;
	}
	catch (std::exception& e) {
		std::cerr << "Exception: " << e.what() << "\n";
		return false;
	}

	return true;
}

}
