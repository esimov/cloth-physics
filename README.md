# cloth-physics
[![Build](https://github.com/esimov/cloth-physics/actions/workflows/build.yml/badge.svg)](https://github.com/esimov/cloth-physics/actions/workflows/build.yml)
[![License](https://img.shields.io/github/license/esimov/cloth-physics)](./LICENSE)
[![Release](https://img.shields.io/badge/release-v1.0.4-blue.svg)](https://github.com/esimov/cloth-physics/releases/tag/v1.0.4)

**cloth-physics** is a native desktop application for 2D cloth physics simulation implementing [Verlet integration](https://en.wikipedia.org/wiki/Verlet_integration). It's written in [Gio](https://gioui.org), a GUI framework for [Go](https://golang.org/). 

It has the following characteristics:
- [x] Possibility to tear up the cloth by applying a mouse pressure on the cloth structure. You can increase the mouse dragging force by pressing and holding the left mouse button. The mouse focus area will change its color depending on the applied force.
- [x] Possibility to make up a hole in the cloth structure by pressing the right mouse button.
- [x] You can change the mouse cloth interaction area by using the scroll button.
- [x] With <kbd>CTRL-left</kbd> click you can pin up the cloth stick under the mouse position.

**Note:** In case you want to learn more about the implementation details, here is a detailed article I wrote: https://medium.com/@esimov/2d-cloth-simulation-in-go-using-gio-gui-b3dfe00b7223.

<p align="center"><img src="./cloth-sim.gif"/></p>

## How to run
Before running the application check the Gio [documentation](https://gioui.org/doc/install) for the system requirements. Install the required dependencies then type the following commands.

```bash
$ git clone https://github.com/esimov/cloth-physics
$ go run ./...
```

Another way is to build the executable yourself then simply run it. 

```bash
$ go build ./...
$ cloth-physics
```

If you don't have Go installed on your machine you can run the prebuild binary files from the project [packages](https://github.com/esimov/cloth-physics/tree/master/packages) page.

## Supported key bindings:
* <kbd>F1</kbd> - Show/hide the quick help panel
* <kbd>SPACE</kbd> - Redraw the cloth
* <kbd>RIGHT CLICK</kbd> - Tear the cloth at the mouse position
* <kbd>SCROLL Up/Down</kbd> - Increase/decrease the mouse focus area
* <kbd>CTRL+CLICK</kbd> - Pin up the cloth on the mouse position
* <kbd>LEFT CLICK+HOLD</kbd> - Increase the mouse pressure

## Author
* Endre Simo ([@simo_endre](https://twitter.com/simo_endre))

## License
Copyright © 2023 Endre Simo

This software is distributed under the MIT license. See the [LICENSE](https://github.com/esimov/cloth-physics/blob/master/LICENSE) file for the full license text.
