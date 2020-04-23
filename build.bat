SET APP_DIR=build
SET GOARCH=386
buildmingw32 go build -o %APP_DIR%\sensel.exe github.com/fpawel/sensel/cmd
