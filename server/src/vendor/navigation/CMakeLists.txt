CMAKE_MINIMUM_REQUIRED(VERSION 2.8) 
SET(CMAKE_C_COMPILER "gcc.exe")
SET(CMAKE_CXX_COMPILER "g++.exe")

PROJECT(navigation)

set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY  ${PROJECT_BINARY_DIR}/lib)

INCLUDE_DIRECTORIES(
./Navigation/Include/ 
./DebugUtils/Include/ 
./Detour/Include/ 
./DetourCrowd/Include/ 
./DetourTileCache/Include/ 
./Recast/Include/ 
./tmxparser/base64/ 
./tmxparser/tinyxml/ 
./tmxparser/zlib/ 
./tmxparser/
)
 
AUX_SOURCE_DIRECTORY(./DebugUtils/Source DIR_DebugUtils)
AUX_SOURCE_DIRECTORY(./Detour/Source DIR_Detour)
AUX_SOURCE_DIRECTORY(./DetourCrowd/Source DIR_DetourCrowd)
AUX_SOURCE_DIRECTORY(./DetourTileCache/Source DIR_DetourTileCache)
AUX_SOURCE_DIRECTORY(./Recast/Source DIR_Recast)
AUX_SOURCE_DIRECTORY(./tmxparser/base64 DIR_base64)
AUX_SOURCE_DIRECTORY(./tmxparser/tinyxml DIR_tinyxml)
AUX_SOURCE_DIRECTORY(./tmxparser/zlib DIR_zlib)
AUX_SOURCE_DIRECTORY(./tmxparser/ DIR_tmxparser)
AUX_SOURCE_DIRECTORY(./Navigation/Source DIR_Navigation)

ADD_LIBRARY(navi navigation.cpp ${DIR_DebugUtils} ${DIR_Detour} ${DIR_DetourCrowd} ${DIR_DetourTileCache} ${DIR_Recast} ${DIR_base64} ${DIR_tinyxml} ${DIR_zlib} ${DIR_tmxparser} ${DIR_DebugUtils} ${DIR_Navigation})
