# blueprinter

`blueprinter` is a Golang library for auto-generating source code for DI (Dependency Injection) containers. It analyzes the AST of `.go` files and generates source code for DI containers, simplifying the structuring of projects and the management of dependencies.

## Features

- Parses the AST of `.go` files and generates source code for DI containers.
- Recursively searches through specified directories to resolve dependencies.
- Provides a simple and intuitive API for easy integration.


## Installation

To install blueprinter, run the following command:

```
go install github.com/yuemori/blueprinter@latest
```

## Usage

TODO

## Commands

Run `blueprinter --help` to more details.

```
Usage:
  blueprinter [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  generate    Generate DI container code
  help        Help about any command

Flags:
  -h, --help   help for blueprinter
```

### generate

```
Usage:
  blueprinter generate <path/to/package> <container struct name> [flags]

Flags:
  -h, --help              help for generate
  -i, --ignore string     Glob pattern for ignoring files
  -o, --out string        Output file for generated code. If not specified, output to stdout
  -t, --template string   Template file for generating code. If not speicied, use default template
  -v, --verbose           Verbose mode
  -w, --workdir string    Workdir for generating code. If not specified, use current directory (default ".")
```

## Key Features and Benefits

Unlike traditional DI libraries, blueprinter takes a unique approach by generating source code, rather than relying on runtime resolution with reflection or implicit resolution at the build time. This approach brings several key benefits:

### Improved Performance
blueprinter resolves dependencies at compile time without using reflection, reducing runtime overhead. This leads to faster application startup times and improved runtime performance.

### Enhanced Type Safety
Since blueprinter generates source code at compile time, issues such as type mismatches or unresolved dependencies are caught early. This significantly reduces the risk of runtime errors and bugs, enhancing overall type safety.

### Clear Visualization of Dependencies
Dependencies are explicitly defined in the source code, making it easier for developers to track and understand them. This clarity improves the maintainability and scalability of the project.

### Compatibility with Go Tools
The generated source code is standard Go code, ensuring compatibility with existing Go tools and IDE features like auto-completion and static analysis tools.

### Ease of Debugging

Debugging runtime errors caused by reflection can be challenging. In contrast, the source code generated by blueprinter can be easily debugged using standard tools and debuggers, simplifying the debugging process.

These benefits make blueprinter an effective DI solution, especially for large-scale applications where performance and maintainability are crucial.

## Contributing

Contributions are welcome! For bug reports, feature requests, or pull requests, please use the GitHub issue tracker.

## License

This project is released under the MIT LICENSE. For more information, see the [LICENSE](./LICENSE) file.
