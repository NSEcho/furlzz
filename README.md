# furlzz

furlzz is a small fuzzer written to test out iOS URL schemes. 
It does so by attaching to the application using Frida and based on the input/seed it mutates the data 
and tries to open the mutated URL. furlzz works in-process, meaning you aren't actually opening 
the URL using apps such as SpringBoard.

# Installation

Download prebuilt binaries from [here](https://github.com/NSEcho/furlzz/releases) or do it manually.

To manually install furlzz, do:
* Follow the instructions for devkit documented [here](https://github.com/frida/frida-go)
* Run `go install github.com/nsecho/furlzz@latest`

# Usage

```bash
$ furlzz --help
Fuzz iOS URL schemes

Usage:
  furlzz [command]

Available Commands:
  crash       Run the application with crash
  fuzz        Fuzz URL scheme
  help        Help about any command

Flags:
  -h, --help   help for furlzz

Use "furlzz [command] --help" for more information about a command.
```

There are basically two ways you can go with fuzzing using `furlzz`:

* give base URL (`--base`) with `FUZZ` keyword in it along with `--input` directory containing inputs
* just give base URL without `FUZZ` keyword which would fuzz the raw base url passed (less efficient)

furlzz supports two post-process methods right now; url and base64. The first one does URL 
encode on the mutated input while the second one generates base64 from it.

# Fuzzing

1. Figure out the method of opening URLs inside the application (with `frida-trace` for example)
2. Find out base url
3. Create some inputs
4. Pass the flags to `furlzz fuzz`
5. Most of the time, values have to be URL encoded, so use `--function url`
6. Adjust timeout if you would like to go with slower fuzzing
7. If the crash happen, replay it with `furlzz crash` passing created session and crash files

![Running against Telegram](telegram.png)

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

# URL open methods

Right now furlzz supports two methods of opening URLs:
* `delegate` when the application uses `-[AppDelegate application:openURL:options:]`
* `app` when the application is using `-[UIApplication openURL:]`

PRs are more than welcome to extend any functionality inside the furlzz