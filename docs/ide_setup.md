# IDE Setup 2025-07-06

As far as I know, the best IDE (as of now) for Go is **[GoLand](https://www.jetbrains.com/go/)**. It's a paid IDE, but students and teachers can get a free license.

I'm a big fan of Visual Studio Code, I will use it as much as I can (knowing that it's not the best choice for Go development).

Here, I will describe how to set up VS Code for Go development. As a starting point, read the [this](https://code.visualstudio.com/docs/languages/go) tutorial.

To learn more about your options, read [this](https://go.dev/wiki/IDEsAndTextEditorPlugins) wiki.

## Table of Contents

- [IDE Setup 2025-07-06](#ide-setup-2025-07-06)
  - [Table of Contents](#table-of-contents)
  - [Install SDK](#install-sdk)
  - [Install Extensions](#install-extensions)
  - [More customizations](#more-customizations)
    - [Code formatting](#code-formatting)

## Install SDK

To help you set up quickly, you can download and install a binary release from [here](https://go.dev/dl/).

If you're using Windows, you can use [WinGet](https://learn.microsoft.com/en-us/windows/package-manager/winget/), like so:

```powershell
winget install --id=GoLang.Go  -e
```

## Install Extensions

To work with Go in Visual Studio Code, you will need to install some extensions.

I've chosen the following ones:

- **[Go](https://marketplace.visualstudio.com/items?itemName=golang.go)** - after installing this extension, please remember to install all the tools (`Ctrl+Shift+P` -> `Go: Install/Update Tools`),
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
