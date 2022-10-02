# Updater

building resources (one time)
```rsrc -manifest updater.manifest -ico=updater.ico -o rsrc.syso```

build
building for x86!
```go version
go version go1.19 windows/386```

build console version
```go build -ldflags "-s -w"
ren appsitory.exe appsitory_console.exe ```

build gui-only version
```go build -ldflags "-s -w -H=windowsgui"  ```

then open appsitory.exe (Open File) in new solution in Visual Studio (not VS Code), and choose in context menu "Add resource", "Version". Set CompanyName, FileDescription, InternalName, OriginalFilename, ProductName to "Appsitory", LegalCopyright to "Copyright (C) 2022 Appsitory". Then Save appsitory.exe (in main menu)