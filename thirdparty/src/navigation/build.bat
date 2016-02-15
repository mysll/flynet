del .\build\* /s/q
del .\lib\* /s/q
FOR /f "delims=" %%i in ('dir /b/a ".\DebugUtils\Source\*.cpp"') DO gcc -I ./DebugUtils/Include/ -I ./Detour/Include/ -I ./DetourCrowd/Include/ -I ./DetourTileCache/Include/ -I ./Recast/Include/ -o ./build/%%~ni.o -c ./DebugUtils/Source/%%i
FOR /f "delims=" %%i in ('dir /b/a ".\Detour\Source\*.cpp"') DO gcc -I ./DebugUtils/Include/ -I ./Detour/Include/ -I ./DetourCrowd/Include/ -I ./DetourTileCache/Include/ -I ./Recast/Include/ -o ./build/%%~ni.o -c ./Detour/Source/%%i
FOR /f "delims=" %%i in ('dir /b/a ".\DetourCrowd\Source\*.cpp"') DO gcc -I ./DebugUtils/Include/ -I ./Detour/Include/ -I ./DetourCrowd/Include/ -I ./DetourTileCache/Include/ -I ./Recast/Include/ -o ./build/%%~ni.o -c ./DetourCrowd/Source/%%i
FOR /f "delims=" %%i in ('dir /b/a ".\DetourTileCache\Source\*.cpp"') DO gcc -I ./DebugUtils/Include/ -I ./Detour/Include/ -I ./DetourCrowd/Include/ -I ./DetourTileCache/Include/ -I ./Recast/Include/ -o ./build/%%~ni.o -c ./DetourTileCache/Source/%%i
FOR /f "delims=" %%i in ('dir /b/a ".\Recast\Source\*.cpp"') DO gcc -I ./DebugUtils/Include/ -I ./Detour/Include/ -I ./DetourCrowd/Include/ -I ./DetourTileCache/Include/ -I ./Recast/Include/ -o ./build/%%~ni.o -c ./Recast/Source/%%i
FOR /f "delims=" %%i in ('dir /b/a ".\Navigation\Source\*.cpp"') DO gcc -I ./Navigation/Include/ -I ./DebugUtils/Include/ -I ./Detour/Include/ -I ./DetourCrowd/Include/ -I ./DetourTileCache/Include/ -I ./Recast/Include/ -o ./build/%%~ni.o -c ./Navigation/Source/%%i
FOR /f "delims=" %%i in ('dir /b/a ".\*.cpp"') DO gcc -I ./ -I ./Navigation/Include/ -I ./DebugUtils/Include/ -I ./Detour/Include/ -I ./DetourCrowd/Include/ -I ./DetourTileCache/Include/ -I ./Recast/Include/ -o ./build/%%~ni.o -c ./%%i

ar rcs ./lib/libnavi.a ./build/*.o