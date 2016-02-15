#!/bin/sh
./kill.sh
cd ..
echo "update svn..."
svn up
cd server/src
echo "now building..."
./build.sh
case $1 in
        sync)
                ./sync.sh;;
esac
echo "build complete"
cd ../../bin
echo "run app"
./master &
