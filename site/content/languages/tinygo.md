+++
title = "TinyGo"
+++

## Introduction

[TinyGo][1] is an alternative compiler for Go source code. It can generate
`%.wasm` files instead of architecture-specific binaries through its `wasi`
target. The resulting wasm depends on a subset of features in the [WebAssembly
2.0 Core specification][2], as well [WASI][3] host imports. TinyGo also
supports importing custom host functions and exporting functions back to the
host.

## Example

Here's a basic example of source in TinyGo:

```go
package main

//export add
func add(x, y uint32) uint32 {
	return x + y
}
```

The following flags will result in the most compact (smallest) wasm file.
```bash
tinygo build -o main.wasm -scheduler=none --no-debug -target=wasi main.go
```

The resulting wasm exports the `add` function so that the embedding host can
call it, regardless of if the host is written in Go or not.

## Disclaimer

This document includes notes contributed by the wazero community. While wazero
includes TinyGo examples, and maintainers often contribute to TinyGo, this
isn't a TinyGo official document. For more help, consider the [TinyGo Using
WebAssembly Guide][4] or joining the [#TinyGo channel on the Gophers Slack][5].

Meanwhile, please help us [maintain][6] this document and [star our GitHub
repository][7], if it is helpful. Together, we can help make WebAssembly easier
on the next person.

## Constraints

Like other compilers that can target wasm, there are constraints using TinyGo.
These constraints affect the library design and dependency choices in your Go
source.

The first constraint people notice is that `encoding/json` usage compiles, but
panics at runtime. This is due to limited support for reflection.

## Memory

When TinyGo compiles go into wasm, it treats a portion of the WebAssembly
linear memory as heap. The embedding host can allocate and free memory using
TinyGo's allocator via WebAssembly exported functions: `malloc` and `free`.

Here is what the signatures look like when inspecting wasm generated by TinyGo:
```webassembly
(func (export "malloc") (param $size i32) (result (;$ptr;) i32))
(func (export "free") (param $ptr i32))
```
Note that TinyGo compiles a `unsafe.Pointer` as a linear memory offset.

The general flow is that the host allocates memory by calling `malloc`, then
using the result as the memory offset to write data. Once the host is finished,
it calls `free` with that same memory offset. wazero includes an [example
project][8] that shows allocation in wasm generated by TinyGo.

These are not documented, though widely used. See the following issues for
clarifications:
* [WebAssembly exports for allocation][9]
* [Memory ownership of TinyGo allocated pointers][10]

## Frequently Asked Questions

### How do I use json?
TinyGo doesn't yet implement [reflection APIs][11] needed by `encoding/json`.
Meanwhile, most users resort to non-reflective parsers, such as [gjson][12].

### Why does my wasm import WASI functions even when I don't use it?
TinyGo does not have a standalone wasm target, rather only `wasi`. Some users
are surprised to see [WASI][3] imports even when there is no main function and
the compiled function uses no memory. Most notably, `fd_write` is used to
implement panics.

### Why is my wasm so big?
TinyGo minimally needs to implement garbage collection and `panic`, and the
wasm to implement that is often not considered big (~4KB). What's often
surprising to users are APIs that seem simple, but require a lot of supporting
functions, such as `fmt.Println`, which can require 100KB of wasm.

[1]: https://tinygo.org/
[2]: https://www.w3.org/TR/2022/WD-wasm-core-2-20220419/
[3]: https://github.com/WebAssembly/WASI
[4]: https://tinygo.org/docs/guides/webassembly/
[5]: https://github.com/tinygo-org/tinygo#getting-help
[6]: https://github.com/tetratelabs/wazero/tree/main/site/content/languages/tinygo.md
[7]: https://github.com/tetratelabs/wazero/stargazers
[8]: https://github.com/tetratelabs/wazero/tree/main/examples/allocation/tinygo
[9]: https://github.com/tinygo-org/tinygo/issues/2788
[10]: https://github.com/tinygo-org/tinygo/issues/2787
[11]: https://github.com/tinygo-org/tinygo/issues/2660
[12]: https://github.com/tidwall/gjson