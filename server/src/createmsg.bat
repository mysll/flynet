del .\pb\c2s\* /s/q
del .\pb\s2c\* /s/q
cd proto
FOR /f "delims=" %%i in ('dir /b/a ".\s2c\*.proto"') DO C:\home\work\goserver\tools\protoc.exe  --go_out=..\pb .\s2c\%%i
FOR /f "delims=" %%i in ('dir /b/a ".\c2s\*.proto"') DO C:\home\work\goserver\tools\protoc.exe  --go_out=..\pb .\c2s\%%i