#!/bin/sh
ps -ef | grep '\./master' | grep -v grep | cut -c 9-15 | xargs kill
sleep 10s
ps -ef | grep '\./loginmgr' | grep -v grep | cut -c 9-15 | xargs kill -9
ps -ef | grep '\./login' | grep -v grep | cut -c 9-15 | xargs kill -9
ps -ef | grep '\./basemgr' | grep -v grep | cut -c 9-15 | xargs kill -9
ps -ef | grep '\./base' | grep -v grep | cut -c 9-15 | xargs kill -9
ps -ef | grep '\./db' | grep -v grep | cut -c 9-15 | xargs kill -9

