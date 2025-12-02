# Go — origin and selling points

## Origins (very short)
- Designed at Google in 2007 by Robert Griesemer, Rob Pike and Ken Thompson to address engineering pain points in large systems: long build times, complex toolchains (C++), and concurrency needs.
- Open-sourced in 2009; stable, widely adopted releases and ecosystem matured over the 2010s.
- Influences: simplicity and small language surface (Pike/Thompson), CSP-style concurrency ideas (channels/goroutines).

## Core design goals
- Simplicity and clarity over language complexity.
- Fast edit-compile-test cycle.
- Good concurrency support for multi-core and networked servers.
- Strong standard tooling and batteries-included standard library.
- Predictable performance and easy deployment (static binaries/cross-compilation).

## Big selling points
- Simplicity and readability — small, opinionated language surface that’s easy to learn.
- Fast compilation — very quick incremental builds for large codebases.
- Concurrency primitives — goroutines (lightweight threads) + channels, scheduler built into runtime.
- Batteries-included toolchain — go fmt, go vet, go build, go test, go mod, etc.
- Robust standard library — networking, HTTP, crypto, JSON, etc., ready for production.
- Static linking & easy cross-compilation — simple deployment of single binaries.
- Memory safety + garbage collection — safe managed memory with focus on low-latency GC.
- Interfaces & duck typing — simple, composable polymorphism without explicit declarations.
- Performance close to C for many workloads, with much higher developer productivity.
- Ecosystem & adoption — used to build Docker, Kubernetes, many cloud services; strong community and module system.
- Generics (since Go 1.18) — parametric polymorphism added without compromising Go’s simplicity.

## When to pick Go
- Networked/back-end services, microservices, CLI tools, distributed systems, tooling where fast builds, simple deployment, and concurrency matter.

(Add more detail as needed.)

# --------------------------------------------------------- #

New Topics:

1. ServeMux   
- a.k.a. router :)  
- read: https://www.alexedwards.net/blog/an-introduction-to-handlers-and-servemuxes-in-go  
2. checksum database for module dependencies
3. nil
