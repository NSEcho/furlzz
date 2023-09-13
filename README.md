# furlzz

furlzz is a small fuzzer written to test out iOS URL schemes. 
It does so by attaching to the application using Frida and based on the input/seed it mutates the data 
and try to open the URL. This works in process, meaning you aren't actually opening the URL using apps 
such as SpringBoard.

# Installation

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
  -m, --method string     method of opening url (delegate, app) (default "delegate")
  -r, --runs uint         number of runs
  -t, --timeout uint      sleep X seconds between each case (default 1)
```

There are basically two ways you can go with fuzzing using `furlzz`:

* give base URL (`--base`) with `FUZZ` keyword in it along with `--seed` directory containing inputs
* just give base URL without `FUZZ` keyword which would fuzz the raw base url passed

Let's say that we would like to fuzz `tg://bg?color=` inside Telegram application. This accepts hex color bytes, 
for example `bbff00`.

1. Decide the method of opening URLs

Run `frida-trace` to trace for `openURL` to determine how application opens URLs. If we see `application:openURL:options:` being called 
we need to pass `-m delegate` to furlzz, and if we see `_openURL:` we will pass `-m app`. There are more methods, but 
these two are supported right now and of course PR are welcome.

2. Create `seed` directory and give some sample inputs

```bash
$ mkdir seeds
$ echo -n bbff00 > seeds/bbff00
$ echo -n '00ffab' > seeds/00ffab
$ echo -n 'ffffff' > seeds/ffffff
```

3. Run furlzz

```bash
$ ./furlzz -b "tg://bg?color=FUZZ" -f url -i seeds/ -t 1 -a Telegram -m delegate
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
