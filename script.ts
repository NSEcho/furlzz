import ObjC from "frida-objc-bridge";

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

let app : ObjC.Object;
let delegate : ObjC.Object;
let shared : ObjC.Object | null = null;
let scene : ObjC.Object | null = null;
let sceneDelegate : ObjC.Object;
let opts : ObjC.Object = NSMutableDictionary.alloc().init();

let NSUserActivityTypeBrowsingWeb = null
let activity: ObjC.Object;

let ctx: ObjC.Object;
let ctxOpts: ObjC.Object;

rpc.exports = {
    setup_fuzz(method, appName, delegateName, sceneName) {
        NSUserActivityTypeBrowsingWeb = Process.getModuleByName("CoreFoundation").getExportByName("NSUserActivityTypeBrowsingWeb");
        switch (method) {
            case "delegate":
                    if (delegateName !== "") {
                        delegate = ObjC.chooseSync(ObjC.classes[delegateName])[0];
                    } else {
                        delegate = ObjC.chooseSync(AppDelegate)[0];
                    }
                    if (appName !== "") {
                        app = ObjC.chooseSync(ObjC.classes[appName])[0];
                    } else {
                        app = ObjC.chooseSync(UIApplication)[0];
                    }
                break;
            case "app":
                if (appName === "") {
                    app = ObjC.chooseSync(UIApplication)[0];
                } else {
                    app = ObjC.chooseSync(ObjC.classes[appName])[0];
                }
                break;
            case "scene_activity":
                activity = NSUserActivity.alloc().initWithActivityType_(NSUserActivityTypeBrowsingWeb);
                sceneDelegate = ObjC.chooseSync(ObjC.classes[delegateName])[0];
                shared = UIApplication.sharedApplication();
                if (sceneName !== "") {
                    scene = ObjC.chooseSync(ObjC.classes[sceneName])[0];
                } else {
                    scene = ObjC.chooseSync(UIWindowScene)[0];
                }
                break;
            case "scene_context":
                sceneDelegate = ObjC.chooseSync(ObjC.classes[delegateName])[0];
                if (sceneName != "") {
                    scene = ObjC.chooseSync(ObjC.classes[sceneName])[0];
                } else {
                    scene = ObjC.chooseSync(UIWindowScene)[0];
                }
                ctx = UIOpenURLContext.alloc().init();
                ctxOpts = UISceneOpenURLOptions.alloc().init();
                break;
            case "delegate_activity":
                activity = NSUserActivity.alloc().initWithActivityType_(NSUserActivityTypeBrowsingWeb);
                delegate = ObjC.chooseSync(ObjC.classes[delegateName])[0];
                shared = UIApplication.sharedApplication();
                if (appName === "") {
                    app = ObjC.chooseSync(UIApplication)[0];
                } else {
                    app = ObjC.chooseSync(ObjC.classes[appName])[0];
                }
                break;
            default:
                return "method not implemented";
        }
    },
    fuzz(method, url) {
        let ur = NSURL.URLWithString_(url);
        switch (method) {
            case "delegate":
                opts.setValue_forKey_(0, "UIApplicationOpenURLOptionsOpenInPlaceKey");
                delegate.application_openURL_options_(app, ur, opts);
                break;
            case "app":
                ObjC.schedule(ObjC.mainQueue, () => {
                    app.openURL_(ur);
                });
                break;
            case "scene_activity":
                if (activity != null && sceneDelegate != null) {
                    activity.setWebPageURL_(ur);

                    ObjC.schedule(ObjC.mainQueue, () => {
                        sceneDelegate.scene_continueUserActivity_(scene, activity);
                    })
                }
                break;
            case "scene_context":
                if (sceneDelegate != null) {
                    ctx.$ivars._URL = ur;
                    ctx.$ivars._options = ctxOpts;
                    let setCtx = NSSet.setWithObject_(ctx);
                    ObjC.schedule(ObjC.mainQueue, () => {
                        sceneDelegate.scene_openURLContexts_(scene, setCtx);
                    });
                }
                break;
            case "delegate_activity":
                activity.setWebPageURL_(ur);

                ObjC.schedule(ObjC.mainQueue, () => {
                    delegate.application_continueUserActivity_restorationHandler_(app,activity,activity);
                })
                break;
            default:
                return "method not implemented";
        }
    }
};
