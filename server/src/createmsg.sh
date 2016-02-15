#!/bin/sh
cd proto/s2c
protoc --go_out=../../pb/s2c *.proto
cd ../c2s
protoc --go_out=../../pb/c2s *.proto