# MirServer-Go
传奇服务器Go语言实现


### windowns下编译
1. 安装cygwin

2. 安装gcc
```bash
apt-cyg install mingw64-x86_64-gcc-g++
```

3. 使用如下命令编译
```bash
env CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC="x86_64-w64-mingw32-gcc" \
    go install -ldflags=-linkmode=internal
```

### idea 项目
设置'Run/Debug Configuration'

* Environment 项添加:
```
CGO_ENABLED=1 GOOS=windows GOARCH=amd64  CC="x86_64-w64-mingw32-gcc"
```

* Go tool arguments 项
```
-ldflags=-linkmode=internal
```