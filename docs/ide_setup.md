# IDE Setup 2024-03-29

As far as I know, the best IDE (as of now) for Go is **[GoLand](https://www.jetbrains.com/go/)**. It's a paid IDE, but students and teachers can get a free license.

I'm a big fan of Visual Studio Code, I will use it as much as I can (knowing that it's not the best choice for Go development).

Here, I will describe how to set up VS Code for Go development; as a start, read [this](https://code.visualstudio.com/docs/languages/go) tutorial and also [this](https://levene.me/boost-your-golang-development-with-these-top-vscode-extensions) one.

To learn more about your options, read [this](https://go.dev/wiki/IDEsAndTextEditorPlugins) wiki.

## Table of Contents

- [IDE Setup 2024-03-29](#ide-setup-2024-03-29)
  - [Table of Contents](#table-of-contents)
  - [Install SDK](#install-sdk)
  - [Install Extensions](#install-extensions)
  - [More customizations](#more-customizations)
    - [Code formatting](#code-formatting)

## Install SDK

To help you set up quickly, you can download and install a binary release from [here](https://go.dev/dl/).

If you're using Windows and using [Chocolatey](https://chocolatey.org/), you can install Go by running the following command:

```powershell
choco install golang
```

Alternatively, you can use [WinGet](https://learn.microsoft.com/en-us/windows/package-manager/winget/), like so:

```powershell
winget install --id=GoLang.Go  -e
```

## Install Extensions

To work with Go in Visual Studio Code, you will need to install some extensions.

I've chosen the following ones:

- **[Go](https://marketplace.visualstudio.com/items?itemName=golang.go)** - after installing this extension, please remember to install all the tools (`Ctrl+Shift+P` -> `Go: Install/Update Tools`),
- **[Go Test Explorer](https://marketplace.visualstudio.com/items?itemName=premparihar.gotestexplorer)**,
- **[Go Auto Struct Tag](https://marketplace.visualstudio.com/items?itemName=vivaldy22.go-auto-struct-tag)**,
- **[Go Outliner](https://marketplace.visualstudio.com/items?itemName=766b.go-outliner)**,
- **[Go Doc](https://marketplace.visualstudio.com/items?itemName=msyrus.go-doc)**,
- [Protobuf (Protocol Buffers)](https://marketplace.visualstudio.com/items?itemName=pbkit.vscode-pbkit).

## More customizations

You'll find more customizations I've made to my VS Code setup here.

### Code formatting

You should perform the following steps to format your code automatically (on save).

Add the following settings to your `settings.json`:

```json
"[go]": {
    "editor.defaultFormatter": "golang.go",
    "editor.formatOnSave": true
}
```
