# furlzz

![Running against Telegram](telegram.png)

furlzz is a small fuzzer written to test out iOS URL schemes.
It does so by attaching to the application using Frida and based on the input/seed it mutates the data 
and tries to open the mutated URL. furlzz works in-process, meaning you aren't actually opening 
the URL using apps such as SpringBoard. furlzz supports universal links which are being used with 
`scene:continueUserActivity` and `application:continueUserActivity`. On some applications it is worth trying to use `app` as method for custom links, because that 
can work as well.

# Installation

Download prebuilt binaries from [here](https://github.com/NSEcho/furlzz/releases) or do it manually.

To manually install furlzz, do:
* Follow the instructions for devkit documented [here](https://github.com/frida/frida-go)
* Run `go install github.com/nsecho/furlzz@latest`

# Usage

## Binary

Simply run the binary with corresponding flags with either attaching over USB or on over the network with `-n` flag.

```bash
$ furlzz fuzz --help
Fuzz URL scheme

Usage:
  furlzz fuzz [flags]

Flags:
  -a, --app string        Application name to attach to (default "Gadget")
  -b, --base string       base URL to fuzz
  -c, --crash             ignore previous crashes
  -d, --delegate string   if the method is scene_activity, you need to specify UISceneDelegate class
  -f, --function string   apply the function to mutated input (url, base64)
  -h, --help              help for fuzz
  -i, --input string      path to input directory
  -m, --method string     method of opening url (delegate, app) (default "delegate")
  -n, --network string    Connect to remote network device (default is "USB")
  -r, --runs uint         number of runs
  -s, --scene string      scene class name
  -t, --timeout uint      sleep X seconds between each case (default 1)
  -u, --uiapp string      UIApplication name
```

## Docker
Starting from `2.5.0`, furlzz now can be run inside of Docker container, for full details visit [Dockerfile.md](./Dockerfile.md) 
for documentation.

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



# Mutations

* `insert` - inserts random byte at random location inside the input
* `del` - deletes random byte
* `substitute` - substitute byte at random position with random byte
* `byteOp` - takes random byte and random position inside the string and do arithmetic operation on them (+, -, *, /)
* `duplicateRange` - duplicates random range inside the original string random number of times
* `bitFlip` - flips the bit at random position inside random location inside input
* `bitmask` - applies random bitmask on random location inside the string
* `duplicate` - duplicates original string random number of times (2 < 10)
* `multiple` - run other mutations random number of times

# URL open methods

Right now furlzz supports a couple of methods of opening URLs:
* `delegate` when the application uses `-[AppDelegate application:openURL:options:]`
* `app` when the application is using `-[UIApplication openURL:]`
* `scene_activity` - when the application is using `-[UISceneDelegate scene:continueUserActivity]` - Universal Links
* `scene_context` when the application is using `-[UISceneDelegate scene:openURLContexts:]`
* `delegate_activity` when the application is using `-[AppDelegate application:continueUserActivity:restorationHandler]` - Universal Links

# Additional flags

* For the method of `scene_activity` you need to pass the `UISceneDelegate` class name
* For the method of `delegate` you need to pass the `AppDelegate` class name
* For the method of `scene_context` you need to pass `UISceneDelegate` class name
* For the method of `delegate_activity` you need to pass `AppDelegate` class name

PRs are more than welcome to extend any functionality inside the furlzz

# Crashes found

* [Bear 2.0.10](https://www.ns-echo.com/posts/furlzz_fuzzing_bear.html)
