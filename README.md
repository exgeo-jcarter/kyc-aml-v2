Checks for matches against a blacklist.  

To set up and use this program:  

$ go get github.com/sajari/fuzzy  
$ go get github.com/dotcypress/phonetics  
$ git clone https://github.com/exgeo-jcarter/kyc-aml-v2.git  
$ cd kyc-aml-v2/kyc-aml-server  
$ go build  
$ ./kyc-aml-server&  

... wait for server to load (about 1 minute)  

$ cd ../kyc-aml-client  
$ go build  
$ ./kyc-aml-client "ali akbar mohummad"  

Note: The server currently requires about 3.2 GB of RAM to run.  

To run the test suite:  

$ cd kyc-aml-v2/kyc-aml-server/KycAmlServer  
$ go test -v  

