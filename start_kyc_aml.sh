#!/bin/sh

cd kyc-aml-data
go build
./kyc-aml-data&
cd ../kyc-aml-fuzzy
go build
./kyc-aml-fuzzy&
cd ../kyc-aml-metaphone
go build
./kyc-aml-metaphone&
cd ../kyc-aml-doublemetaphone/build
cmake ..
make
./kyc-aml-doublemetaphone&
cd ../..
