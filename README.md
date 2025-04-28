# gloop Project

## Overview
The gloop project is a Go package that provides a simple web server implementation through the `SiteServer` type. This package allows developers to create and manage a web server with customizable options.

## Files

- **site/site.go**: Contains the definitions for `SiteServer` and `SiteServerOptions`, along with a constructor function `NewSiteServer` and a method `Start` for the `SiteServer`.

- **site/site_test.go**: This file is used for testing the functionality of the `SiteServer` and its methods to ensure they behave as expected.

- **go.mod**: The module definition file that specifies the module name and its dependencies.

## Installation

To use the gloop package, you need to have Go installed on your machine. You can download it from the official Go website.

Once Go is installed, you can clone this repository and navigate to the project directory:

```bash
git clone <repository-url>
cd gloop
```

## Usage

To create a new `SiteServer`, you can use the `NewSiteServer` function. Here is an example:

```go
package main

import (
    "gloop/site"
)

func main() {
    server := site.NewSiteServer("config.yaml")
    server.Start()
}
```

## Running Tests

To run the tests for the `SiteServer`, navigate to the `site` directory and use the following command:

```bash
go test
```

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue for any enhancements or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.