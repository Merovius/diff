[Documentation](https://godoc.org/github.com/Merovius/diff)

# Simple, yet flexible diffing package for Go

This is an implementation of the algorithm from ["An Algorithm for Differential
File Comparison" by Hunt and McIlroy](https://www.cs.dartmouth.edu/~doug/diff.pdf).
I wrote it because I needed something more flexible than the common line-based
diff and it seemed to cumbersome to do it as a transformation on the input (and
the reverse on the output). As such, I attempted to make it flexible enough for
whatever semantics you might need. It's intended to be mostly finished and
frozen, but feel free to file issues if you have a problem.

The implementation is intentionally kept close to the paper, with just a little
bit of extra convenience provided by Go. As such, the performance is probably
"meh" (though I haven't perceived any issues yet) and I don't intend to improve
it considerably.

# License

```
Copyright 2020 Axel Wagner

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
