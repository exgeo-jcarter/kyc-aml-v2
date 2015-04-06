#ifndef EXGEO_KYCAMLDOUBLEMETAPHONE_HPP
#define EXGEO_KYCAMLDOUBLEMETAPHONE_HPP

#include <memory>
#include <boost/asio.hpp>
#include <vector>
#include <string>
#include <map>

using boost::asio::ip::tcp;

namespace exGeoKycAml {

class KycAmlDoubleMetaphone;

struct KycAmlDoubleMetaphoneConfS {

	std::string Host;
	std::string Port;
	std::string Protocol;
};

class Session : public std::enable_shared_from_this<Session>
{
	void doRead();
	void doWrite(std::string msg);

	KycAmlDoubleMetaphone* parent;
	tcp::socket socket;
	boost::asio::streambuf streambuf;

public:
	Session(KycAmlDoubleMetaphone* parent, tcp::socket socket);
	void Start();
};

class Server
{
	void doAccept();

	KycAmlDoubleMetaphone* parent;
	tcp::acceptor acceptor;
	tcp::socket socket;

public:
	Server(KycAmlDoubleMetaphone* parent, boost::asio::io_service& io_service, short port);
};

class KycAmlDoubleMetaphone {

public:
	KycAmlDoubleMetaphone(const std::string& conf_filename);
	~KycAmlDoubleMetaphone();

	bool LoadConf(const std::string& filename);
	bool Listen();
	bool TrainSdn(const std::string& sdn_list);

	KycAmlDoubleMetaphoneConfS* Conf;

	std::map<std::string, std::string> SdnListMapNames;
	std::map<std::string, std::string> SdnListMapRevNames;

	std::map<std::string, std::string> SdnListMapAkas;
	std::map<std::string, std::string> SdnListMapRevAkas;

	std::map<std::string, std::string> SdnListMapAddresses;
	std::map<std::string, std::string> SdnListMapPostalCodes;
};

}

#endif
