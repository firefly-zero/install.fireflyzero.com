# install.fireflyzero.com

A web service for installing [Firefly Zero] apps using [firefly-installer](https://github.com/firefly-zero/firefly-installer).

Try it: [install.fireflyzero.com](https://install.fireflyzero.com/)

Written in [Go](https://go.dev/) using no dependencies except [postcard](https://github.com/orsinium-labs/postcard) decoder. Requires no auth and uses no tracking and no telemetry.

## Running locally

1. In the device FS, create `data/sys/installer/etc/addr` file with the service address (IP:port). For example, for [Firefly Emulator](https://github.com/firefly-zero/firefly-emulator) running on the same computer, it will be `127.0.0.1:19742`.
1. Run the server: `go run .` .
1. Open the web interface in your browser: [127.0.0.1:19742](http://127.0.0.1:19742/).
1. Launch the installer and follow the regular app installation workflow.

## License

MIT License. Feel free to modify the installer server as you see fit or write your own. Keep in mind, though, that there might be breaking changes to the installer app which might require you to update your custom implementation.
