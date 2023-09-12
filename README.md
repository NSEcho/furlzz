# furlzz

furlzz is a small fuzzer written to test out iOS URL schemes. 
It does so by attaching to the application using Frida and based on the input/seed it mutates the data 
and try to open the URL. This works in process, meaning you aren't actually opening the URL using apps 
such as SpringBoard.

# Installation

Download one of the prebuilt binaries for macOS(x86_64 or arm64) from [here](#) or do it manually as described below.

* Follow the instructions for devkit documented [here](https://github.com/frida/frida-go)
* Run `go install github.com/nsecho/furlzz@latest`

# Usage

```bash
$ furlzz --help
Fuzz iOS URL schemes

Usage:
  furlzz [flags]

Flags:
  -a, --app string        Application name to attach to (default "Gadget")
  -b, --base string       base URL to fuzz
  -f, --function string   apply the function to mutated input (url, base64)
  -h, --help              help for furlzz
  -i, --input string      path to input directory
  -r, --runs uint         number of runs
  -t, --timeout uint      sleep X seconds between each case (default 1)
```

There are basically two ways you can go with fuzzing using `furlzz`:

* give base URL (`--base`) with `FUZZ` keyword in it along with `--seed` directory containing inputs
* just give base URL without `FUZZ` keyword which would fuzz the raw base url passed

Let's say that we would like to fuzz `tg://bg?color=` inside Telegram application. This accepts hex color bytes, 
for example `bbff00`.

1. Create `seed` directory and give some sample inputs

```bash
$ mkdir seeds
$ echo -n bbff00 > seeds/bbff00
$ echo -n '00ffab' > seeds/00ffab
$ echo -n 'ffffff' > seeds/ffffff
```

2. Run furlzz

```bash
$ furlzz -a Telegram -b 'tg://bg?color=FUZZ' -s seeds/ -r 100 -fn url
```

furlzz supports two post-process methods right now; url and base64. The first one does URL 
encode on the mutated input while the second one generates base64 from it.

# Mutations

* `insert` - inserts random byte at random location inside the input
* `del` - deletes random byte
* `substitute` - substitute byte at random position with random byte
* `byteOp` - takes random byte and random position inside the string and do arithmetic operation on them (+, -, *, /)
* `duplicateRange` - duplicates random range inside the original string random number of times
* `bitFlip` - flips the bit at random position inside random location inside input
* `bitmask` - applies random bitmask on random location inside the string
* `duplicate` - duplicates original string random number of times (2 < 10)
* `another` - run other mutations random number of times
