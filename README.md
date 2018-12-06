# Ram 🐏 — golang opiniated continuous testing tool

This is a very opiniated « continuous testing » tool for =Go=.
In a nutshell it does : watch a folder (gopath or not…) and execute
tests when file changes.

It supports:
- changing code in a package will only re-run tests on this package
- changing a test code, it will only re-run *that* test
