import ObjC from "frida-objc-bridge";

class CoverageCollector {
    public DEBUG: boolean = false;

    private events: any[] = [];
    private gcConter: number = 0;
    private funcAddr: NativePointer;

    private _lastNewBlocks: number = 0;
    private globalCoverage: Set<string> = new Set();

    constructor(funcAddr: NativePointer, debug: boolean = false) {
        this.DEBUG = debug;

        this.funcAddr = funcAddr;
        Stalker.trustThreshold = 3;
        Stalker.queueCapacity = 0x8000;
        Stalker.queueDrainInterval = 1000 * 1000;
    }

    start() {
        const self = this;

        try {
            Interceptor.attach(this.funcAddr, {
                onEnter: function (args) {
                    this._args = args;

                    self.debug("[*] Interceptor ENTER");
                    Stalker.follow(Process.getCurrentThreadId(), {
                        events: {
                            call: false,
                            ret: false,
                            exec: false,
                            block: false,
                            compile: true
                        },
                        onReceive: (events) => {
                            const parsed = Stalker.parse(events, { stringify: false, annotate: false });
                            self.events.push(...parsed);
                        }
                    });
                },
                onLeave: function () {
                    self.debug("[*] Interceptor LEAVE");

                    Stalker.unfollow();
                    Stalker.flush();
                    if (self.gcConter > 300) {
                        Stalker.garbageCollect();
                        self.gcConter = 0;
                    }
                    self.gcConter++;

                    let newBlocks = 0;
                    for (const event of self.events) {
                        const addr = event[1]?.toString?.() ?? String(event[1]);
                        if (!self.globalCoverage.has(addr)) {
                            self.globalCoverage.add(addr);
                            newBlocks++;
                        }
                    }

                    if (newBlocks > 0) {
                        self.debug(`[+] New blocks found: ${newBlocks}, Total unique blocks: ${self.globalCoverage.size}`);
                    }
                    self._lastNewBlocks = newBlocks;

                    self.events = [];
                }
            });
        }
        catch (e) {
            self.debug(`[-] Error starting Stalker: ${e instanceof Error ? e.message : JSON.stringify(e)}`);
        }

    }

    hasNewBlocks(): boolean {
        const hasNew = this._lastNewBlocks > 0;
        this._lastNewBlocks = 0;
        return hasNew;
    }

    debug(msg: string) {
        if (this.DEBUG) { console.log("[+ (" + Process.id + ")] " + msg) }
    }
}

const {
    AppDelegate,
    UIApplication,
    NSURL,
    NSMutableDictionary,
    UIWindowScene,
    NSUserActivity,
    UIOpenURLContext,
    UISceneOpenURLOptions,
    NSSet,
} = ObjC.classes;

const NSUserActivityTypeBrowsingWeb: string = "NSUserActivityTypeBrowsingWeb";

class Fuzzer {
    private app!: ObjC.Object;
    private delegate!: ObjC.Object;
    private scene!: ObjC.Object;
    private sceneDelegate!: ObjC.Object;
    private activity!: ObjC.Object;
    private ctx!: ObjC.Object;
    private ctxOpts!: ObjC.Object;
    private opts: ObjC.Object = NSMutableDictionary.alloc().init();

    public coverer: CoverageCollector | null = null;

    setup(method: string, appName: string, delegateName: string, sceneName: string, debug: boolean) {
        let addr: NativePointer | null = null;

        switch (method) {
            case "delegate":
                if (delegateName !== "") {
                    this.delegate = ObjC.chooseSync(ObjC.classes[delegateName])[0];
                } else {
                    this.delegate = ObjC.chooseSync(AppDelegate)[0];
                }
                if (appName !== "") {
                    this.app = ObjC.chooseSync(ObjC.classes[appName])[0];
                } else {
                    this.app = ObjC.chooseSync(UIApplication)[0];
                }

                if (this.delegate)
                    addr = this.delegate["- application:openURL:options:"].implementation;
                break;
            case "application":
                if (appName === "") {
                    this.app = ObjC.chooseSync(UIApplication)[0];
                } else {
                    this.app = ObjC.chooseSync(ObjC.classes[appName])[0];
                }

                if (this.app)
                    addr = this.app["- openURL:"].implementation;
                break;
            case "scene_activity":
                this.activity = NSUserActivity.alloc().initWithActivityType_(NSUserActivityTypeBrowsingWeb);
                this.sceneDelegate = ObjC.chooseSync(ObjC.classes[delegateName])[0];
                if (sceneName !== "") {
                    this.scene = ObjC.chooseSync(ObjC.classes[sceneName])[0];
                } else {
                    this.scene = ObjC.chooseSync(UIWindowScene)[0];
                }

                if (this.sceneDelegate)
                    addr = this.sceneDelegate["- scene:continueUserActivity:"].implementation;
                break;
            case "scene_context":
                this.sceneDelegate = ObjC.chooseSync(ObjC.classes[delegateName])[0];
                if (sceneName != "") {
                    this.scene = ObjC.chooseSync(ObjC.classes[sceneName])[0];
                } else {
                    this.scene = ObjC.chooseSync(UIWindowScene)[0];
                }
                this.ctx = UIOpenURLContext.alloc().init();
                this.ctxOpts = UISceneOpenURLOptions.alloc().init();

                if (this.sceneDelegate)
                    addr = this.sceneDelegate["- scene:openURLContexts:"].implementation;
                break;
            case "delegate_activity":
                this.activity = NSUserActivity.alloc().initWithActivityType_(NSUserActivityTypeBrowsingWeb);
                this.delegate = ObjC.chooseSync(ObjC.classes[delegateName])[0];
                if (appName === "") {
                    this.app = ObjC.chooseSync(UIApplication)[0];
                } else {
                    this.app = ObjC.chooseSync(ObjC.classes[appName])[0];
                }

                if (this.delegate)
                    addr = this.delegate["- application:continueUserActivity:restorationHandler:"].implementation;
                break;
            default:
                return "method not implemented";
        }

        if (!addr) return "function not found";
        this.coverer = new CoverageCollector(addr, debug);
        this.coverer.start();
    }

    fuzz(method: string, url: string) {
        let ur = NSURL.URLWithString_(url);
        switch (method) {
            case "delegate":
                this.opts.setValue_forKey_(0, "UIApplicationOpenURLOptionsOpenInPlaceKey");
                this.delegate.application_openURL_options_(this.app, ur, this.opts);
                break;
            case "application":
                ObjC.schedule(ObjC.mainQueue, () => {
                    this.app.openURL_(ur);
                });
                break;
            case "scene_activity":
                if (this.activity != null && this.sceneDelegate != null) {
                    this.activity.setWebPageURL_(ur);

                    ObjC.schedule(ObjC.mainQueue, () => {
                        this.sceneDelegate.scene_continueUserActivity_(this.scene, this.activity);
                    })
                }
                break;
            case "scene_context":
                if (this.sceneDelegate != null) {
                    this.ctx.$ivars._URL = ur;
                    this.ctx.$ivars._options = this.ctxOpts;
                    let setCtx = NSSet.setWithObject_(this.ctx);
                    ObjC.schedule(ObjC.mainQueue, () => {
                        this.sceneDelegate.scene_openURLContexts_(this.scene, setCtx);
                    });
                }
                break;
            case "delegate_activity":
                this.activity.setWebPageURL_(ur);

                ObjC.schedule(ObjC.mainQueue, () => {
                    this.delegate.application_continueUserActivity_restorationHandler_(this.app, this.activity, null);
                })
                break;
            default:
                return "method not implemented";
        }
    }
}

const fuzzer = new Fuzzer();
rpc.exports = {
    setup_fuzz: function (method: string, appName: string, delegateName: string, sceneName: string, debug: boolean) {
        fuzzer.setup(method, appName, delegateName, sceneName, debug);
    },
    fuzz: function (method: string, url: string) {
        fuzzer.fuzz(method, url);
    },
    has_new_blocks: function (): boolean {
        if (!fuzzer.coverer) {
            return false;
        }

        return fuzzer.coverer.hasNewBlocks();
    },
};
